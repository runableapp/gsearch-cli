package db

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

// CreateTestDatabase creates a test database file with sample data
func CreateTestDatabase(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create test database file: %w", err)
	}
	defer file.Close()

	writer := &testDBWriter{file: file, seeker: file}

	// Create test data
	testData := createTestData()

	// Write header
	if err := writer.writeHeader(); err != nil {
		return err
	}

	// Write metadata
	indexFlags := IndexFlagName | IndexFlagSize | IndexFlagModificationTime
	if err := writer.writeMetadata(indexFlags, uint32(len(testData.folders)), uint32(len(testData.files))); err != nil {
		return err
	}

	// Write placeholder block sizes (will update later)
	folderBlockSizeOffset, err := writer.seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if err := writer.writeUint64(0); err != nil { // folder block size placeholder
		return err
	}
	fileBlockSizeOffset, err := writer.seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if err := writer.writeUint64(0); err != nil { // file block size placeholder
		return err
	}

	// Write indexes (0)
	if err := writer.writeUint32(0); err != nil {
		return err
	}

	// Write excludes (0)
	if err := writer.writeUint32(0); err != nil {
		return err
	}

	// Write exclude pattern (0 bytes - nothing to write)

	// Write folders
	folderBlockStart, err := writer.seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if err := writer.writeFolders(testData.folders, indexFlags); err != nil {
		return err
	}
	folderBlockEnd, err := writer.seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	folderBlockSize := folderBlockEnd - folderBlockStart

	// Write files
	fileBlockStart, err := writer.seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if err := writer.writeFiles(testData.files, indexFlags); err != nil {
		return err
	}
	fileBlockEnd, err := writer.seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	fileBlockSize := fileBlockEnd - fileBlockStart

	// Write sorted arrays (empty for test)
	if err := writer.writeUint32(0); err != nil { // num sorted arrays
		return err
	}

	// Update block sizes
	if _, err := writer.seeker.Seek(folderBlockSizeOffset, io.SeekStart); err != nil {
		return err
	}
	if err := writer.writeUint64(uint64(folderBlockSize)); err != nil {
		return err
	}

	if _, err := writer.seeker.Seek(fileBlockSizeOffset, io.SeekStart); err != nil {
		return err
	}
	if err := writer.writeUint64(uint64(fileBlockSize)); err != nil {
		return err
	}

	return nil
}

type testDBWriter struct {
	file   io.Writer
	seeker io.Seeker
}

func (w *testDBWriter) writeHeader() error {
	// Magic number
	if _, err := w.file.Write([]byte(MagicNumber)); err != nil {
		return err
	}
	// Major version
	if err := w.writeUint8(MajorVersion); err != nil {
		return err
	}
	// Minor version
	if err := w.writeUint8(MinorVersion); err != nil {
		return err
	}
	return nil
}

func (w *testDBWriter) writeMetadata(indexFlags IndexFlags, numFolders, numFiles uint32) error {
	if err := w.writeUint64(uint64(indexFlags)); err != nil {
		return err
	}
	if err := w.writeUint32(numFolders); err != nil {
		return err
	}
	if err := w.writeUint32(numFiles); err != nil {
		return err
	}
	return nil
}

func (w *testDBWriter) writeFolders(folders []testFolder, indexFlags IndexFlags) error {
	previousName := ""
	for i, folder := range folders {
		// db_index (2 bytes) - currently unused, set to 0
		if err := w.writeUint16(0); err != nil {
			return err
		}

		// Write delta-compressed name
		if err := w.writeDeltaName(folder.name, previousName); err != nil {
			return err
		}
		previousName = folder.name

		// Write size if indexed
		if indexFlags&IndexFlagSize != 0 {
			if err := w.writeInt64(folder.size); err != nil {
				return err
			}
		}

		// Write mtime if indexed
		if indexFlags&IndexFlagModificationTime != 0 {
			if err := w.writeInt64(folder.mtime.Unix()); err != nil {
				return err
			}
		}

		// Write parent index
		parentIdx := folder.parentIdx
		if parentIdx == uint32(i) {
			// Self-reference means no parent (root)
			parentIdx = uint32(i)
		}
		if err := w.writeUint32(parentIdx); err != nil {
			return err
		}
	}
	return nil
}

func (w *testDBWriter) writeFiles(files []testFile, indexFlags IndexFlags) error {
	previousName := ""
	for _, file := range files {
		// Write delta-compressed name
		if err := w.writeDeltaName(file.name, previousName); err != nil {
			return err
		}
		previousName = file.name

		// Write size if indexed
		if indexFlags&IndexFlagSize != 0 {
			if err := w.writeInt64(file.size); err != nil {
				return err
			}
		}

		// Write mtime if indexed
		if indexFlags&IndexFlagModificationTime != 0 {
			if err := w.writeInt64(file.mtime.Unix()); err != nil {
				return err
			}
		}

		// Write parent index
		if err := w.writeUint32(file.parentIdx); err != nil {
			return err
		}
	}
	return nil
}

func (w *testDBWriter) writeDeltaName(name, previousName string) error {
	// Calculate name offset (where names start to differ)
	nameOffset := uint8(0)
	minLen := len(name)
	if len(previousName) < minLen {
		minLen = len(previousName)
	}
	for i := 0; i < minLen; i++ {
		if name[i] != previousName[i] {
			break
		}
		nameOffset++
	}

	// Calculate name length (new characters to append)
	nameLen := uint8(len(name) - int(nameOffset))

	// Write offset and length
	if err := w.writeUint8(nameOffset); err != nil {
		return err
	}
	if err := w.writeUint8(nameLen); err != nil {
		return err
	}

	// Write new name characters
	if nameLen > 0 {
		if _, err := w.file.Write([]byte(name[nameOffset:])); err != nil {
			return err
		}
	}

	return nil
}

func (w *testDBWriter) writeUint8(v uint8) error {
	return binary.Write(w.file, binary.LittleEndian, v)
}

func (w *testDBWriter) writeUint16(v uint16) error {
	return binary.Write(w.file, binary.LittleEndian, v)
}

func (w *testDBWriter) writeUint32(v uint32) error {
	return binary.Write(w.file, binary.LittleEndian, v)
}

func (w *testDBWriter) writeUint64(v uint64) error {
	return binary.Write(w.file, binary.LittleEndian, v)
}

func (w *testDBWriter) writeInt64(v int64) error {
	return binary.Write(w.file, binary.LittleEndian, v)
}

type testData struct {
	folders []testFolder
	files   []testFile
}

type testFolder struct {
	name     string
	size     int64
	mtime    time.Time
	parentIdx uint32
}

type testFile struct {
	name     string
	size     int64
	mtime    time.Time
	parentIdx uint32
}

func createTestData() testData {
	now := time.Now()

	// Create folder hierarchy:
	// / (root, index 0)
	// /home (index 1, parent 0)
	// /home/user (index 2, parent 1)
	// /Documents (index 3, parent 0)
	// /Downloads (index 4, parent 0)

	folders := []testFolder{
		{name: "", size: 0, mtime: now, parentIdx: 0},           // root
		{name: "home", size: 0, mtime: now, parentIdx: 0},       // /home
		{name: "user", size: 0, mtime: now, parentIdx: 1},       // /home/user
		{name: "Documents", size: 0, mtime: now, parentIdx: 0},  // /Documents
		{name: "Downloads", size: 0, mtime: now, parentIdx: 0},  // /Downloads
	}

	// Create files:
	// /home/user/test.txt (parent 2)
	// /home/user/readme.txt (parent 2)
	// /Documents/document.pdf (parent 3)
	// /Documents/test.go (parent 3)
	// /Downloads/file.zip (parent 4)

	files := []testFile{
		{name: "test.txt", size: 1024, mtime: now, parentIdx: 2},
		{name: "readme.txt", size: 2048, mtime: now, parentIdx: 2},
		{name: "document.pdf", size: 4096, mtime: now, parentIdx: 3},
		{name: "test.go", size: 8192, mtime: now, parentIdx: 3},
		{name: "file.zip", size: 16384, mtime: now, parentIdx: 4},
	}

	return testData{
		folders: folders,
		files:   files,
	}
}

