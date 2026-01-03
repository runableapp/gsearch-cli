package db

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	MagicNumber        = "FSDB"
	MajorVersion       = 0
	MinorVersion       = 9
	HeaderSize         = 6
	MaxNameLength      = 256
)

// IndexFlags represents which metadata fields are indexed
type IndexFlags uint64

const (
	IndexFlagName              IndexFlags = 1 << 0
	IndexFlagPath              IndexFlags = 1 << 1
	IndexFlagSize              IndexFlags = 1 << 2
	IndexFlagModificationTime  IndexFlags = 1 << 3
	IndexFlagAccessTime        IndexFlags = 1 << 4
	IndexFlagCreationTime      IndexFlags = 1 << 5
	IndexFlagStatusChangeTime  IndexFlags = 1 << 6
)

// EntryType represents the type of database entry
type EntryType uint8

const (
	EntryTypeNone   EntryType = 0
	EntryTypeFolder EntryType = 1
	EntryTypeFile   EntryType = 2
)

// Entry represents a file or folder entry in the database
type Entry struct {
	Name     string
	Size     int64
	MTime    time.Time
	Parent   *Folder
	Index    uint32
	Type     EntryType
}

// Folder represents a folder entry with additional metadata
type Folder struct {
	Entry
	DBIndex   uint32
	NumFiles  uint32
	NumFolders uint32
}

// Database represents the loaded FSearch database
type Database struct {
	IndexFlags   IndexFlags
	Folders      []*Folder
	Files        []*Entry
	SortedArrays map[uint32]*SortedArray
	metadata     metadata
	pathCache    sync.Map // map[*Entry]string - caches computed paths for performance
}

// SortedArray contains pre-sorted indices for efficient searching
type SortedArray struct {
	ID       uint32
	Folders  []uint32 // Indices into Folders array
	Files    []uint32 // Indices into Files array
}

// Load opens and reads an FSearch database file
func Load(filePath string) (*Database, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %w", err)
	}
	defer file.Close()

	// Try to acquire lock (non-blocking)
	// Note: On Linux, we'd use syscall.Flock, but for portability we'll skip locking for read-only access
	// The original code uses flock() with LOCK_EX|LOCK_NB, but for read-only we can proceed

	db := &Database{
		SortedArrays: make(map[uint32]*SortedArray),
	}

	// Read and verify header
	if err := db.readHeader(file); err != nil {
		return nil, err
	}

	// Read metadata
	if err := db.readMetadata(file); err != nil {
		return nil, err
	}

	// Pre-allocate folders
	db.Folders = make([]*Folder, db.metadata.numFolders)
	for i := uint32(0); i < db.metadata.numFolders; i++ {
		db.Folders[i] = &Folder{
			Entry: Entry{
				Index: i,
				Type:  EntryTypeFolder,
			},
		}
	}

	// Load folders
	if err := db.loadFolders(file); err != nil {
		return nil, err
	}

	// Load files
	if err := db.loadFiles(file); err != nil {
		return nil, err
	}

	// Load sorted arrays
	if err := db.loadSortedArrays(file); err != nil {
		return nil, err
	}

	return db, nil
}

type metadata struct {
	indexFlags      IndexFlags
	numFolders      uint32
	numFiles        uint32
	folderBlockSize uint64
	fileBlockSize   uint64
	numIndexes      uint32
	numExcludes     uint32
}

func (db *Database) readHeader(r io.Reader) error {
	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return fmt.Errorf("failed to read magic number: %w", err)
	}

	if string(magic) != MagicNumber {
		return fmt.Errorf("invalid magic number: got %q, expected %q", string(magic), MagicNumber)
	}

	var majorVer, minorVer uint8
	if err := binary.Read(r, binary.LittleEndian, &majorVer); err != nil {
		return fmt.Errorf("failed to read major version: %w", err)
	}
	if majorVer != MajorVersion {
		return fmt.Errorf("unsupported major version: got %d, expected %d", majorVer, MajorVersion)
	}

	if err := binary.Read(r, binary.LittleEndian, &minorVer); err != nil {
		return fmt.Errorf("failed to read minor version: %w", err)
	}
	if minorVer > MinorVersion {
		return fmt.Errorf("unsupported minor version: got %d, expected <= %d", minorVer, MinorVersion)
	}

	return nil
}

func (db *Database) readMetadata(r io.Reader) error {
	var meta metadata

	if err := binary.Read(r, binary.LittleEndian, &meta.indexFlags); err != nil {
		return fmt.Errorf("failed to read index flags: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &meta.numFolders); err != nil {
		return fmt.Errorf("failed to read number of folders: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &meta.numFiles); err != nil {
		return fmt.Errorf("failed to read number of files: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &meta.folderBlockSize); err != nil {
		return fmt.Errorf("failed to read folder block size: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &meta.fileBlockSize); err != nil {
		return fmt.Errorf("failed to read file block size: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &meta.numIndexes); err != nil {
		return fmt.Errorf("failed to read number of indexes: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &meta.numExcludes); err != nil {
		return fmt.Errorf("failed to read number of excludes: %w", err)
	}

	db.IndexFlags = meta.indexFlags
	db.metadata = meta

	return nil
}

func (db *Database) loadFolders(r io.Reader) error {
	// Read the entire folder block into memory
	folderBlock := make([]byte, db.metadata.folderBlockSize)
	if _, err := io.ReadFull(r, folderBlock); err != nil {
		return fmt.Errorf("failed to read folder block: %w", err)
	}

	offset := 0
	previousName := ""

	for i := uint32(0); i < db.metadata.numFolders; i++ {
		folder := db.Folders[i]

		// Read db_index (2 bytes)
		if offset+2 > len(folderBlock) {
			return fmt.Errorf("folder block truncated at folder %d", i)
		}
		// db_index is currently unused, skip it
		offset += 2

		// Read name using delta compression
		var err error
		previousName, offset, err = db.readDeltaName(folderBlock, offset, previousName)
		if err != nil {
			return fmt.Errorf("failed to read folder name at index %d: %w", i, err)
		}
		folder.Name = previousName

		// Read size if indexed
		if db.IndexFlags&IndexFlagSize != 0 {
			if offset+8 > len(folderBlock) {
				return fmt.Errorf("folder block truncated at folder %d (size)", i)
			}
			var size int64
			size = int64(binary.LittleEndian.Uint64(folderBlock[offset:]))
			folder.Size = size
			offset += 8
		}

		// Read mtime if indexed
		if db.IndexFlags&IndexFlagModificationTime != 0 {
			if offset+8 > len(folderBlock) {
				return fmt.Errorf("folder block truncated at folder %d (mtime)", i)
			}
			mtime := int64(binary.LittleEndian.Uint64(folderBlock[offset:]))
			folder.MTime = time.Unix(mtime, 0)
			offset += 8
		}

		// Read parent index
		if offset+4 > len(folderBlock) {
			return fmt.Errorf("folder block truncated at folder %d (parent)", i)
		}
		parentIdx := binary.LittleEndian.Uint32(folderBlock[offset:])
		offset += 4

		// Set parent (if not self-reference)
		if parentIdx != folder.Index && parentIdx < uint32(len(db.Folders)) {
			folder.Parent = db.Folders[parentIdx]
		}
	}

	if offset != len(folderBlock) {
		return fmt.Errorf("folder block size mismatch: read %d bytes, expected %d", offset, len(folderBlock))
	}

	return nil
}

func (db *Database) loadFiles(r io.Reader) error {
	// Read the entire file block into memory
	fileBlock := make([]byte, db.metadata.fileBlockSize)
	if _, err := io.ReadFull(r, fileBlock); err != nil {
		return fmt.Errorf("failed to read file block: %w", err)
	}

	db.Files = make([]*Entry, db.metadata.numFiles)
	offset := 0
	previousName := ""

	for i := uint32(0); i < db.metadata.numFiles; i++ {
		entry := &Entry{
			Index: i,
			Type:  EntryTypeFile,
		}

		// Read name using delta compression
		var err error
		previousName, offset, err = db.readDeltaName(fileBlock, offset, previousName)
		if err != nil {
			return fmt.Errorf("failed to read file name at index %d: %w", i, err)
		}
		entry.Name = previousName

		// Read size if indexed
		if db.IndexFlags&IndexFlagSize != 0 {
			if offset+8 > len(fileBlock) {
				return fmt.Errorf("file block truncated at file %d (size)", i)
			}
			size := int64(binary.LittleEndian.Uint64(fileBlock[offset:]))
			entry.Size = size
			offset += 8
		}

		// Read mtime if indexed
		if db.IndexFlags&IndexFlagModificationTime != 0 {
			if offset+8 > len(fileBlock) {
				return fmt.Errorf("file block truncated at file %d (mtime)", i)
			}
			mtime := int64(binary.LittleEndian.Uint64(fileBlock[offset:]))
			entry.MTime = time.Unix(mtime, 0)
			offset += 8
		}

		// Read parent index
		if offset+4 > len(fileBlock) {
			return fmt.Errorf("file block truncated at file %d (parent)", i)
		}
		parentIdx := binary.LittleEndian.Uint32(fileBlock[offset:])
		offset += 4

		// Set parent
		if parentIdx < uint32(len(db.Folders)) {
			entry.Parent = db.Folders[parentIdx]
		}

		db.Files[i] = entry
	}

	if offset != len(fileBlock) {
		return fmt.Errorf("file block size mismatch: read %d bytes, expected %d", offset, len(fileBlock))
	}

	return nil
}

// readDeltaName reads a delta-compressed name from the block
func (db *Database) readDeltaName(block []byte, offset int, previousName string) (string, int, error) {
	if offset+2 > len(block) {
		return "", offset, fmt.Errorf("block truncated at name header")
	}

	nameOffset := block[offset]
	nameLen := block[offset+1]
	offset += 2

	// Truncate previous name at offset
	var name string
	if int(nameOffset) < len(previousName) {
		name = previousName[:nameOffset]
	} else {
		name = previousName
	}

	// Append new characters
	if nameLen > 0 {
		if offset+int(nameLen) > len(block) {
			return "", offset, fmt.Errorf("block truncated at name data")
		}
		name += string(block[offset : offset+int(nameLen)])
		offset += int(nameLen)
	}

	return name, offset, nil
}

func (db *Database) loadSortedArrays(r io.Reader) error {
	var numSortedArrays uint32
	if err := binary.Read(r, binary.LittleEndian, &numSortedArrays); err != nil {
		return fmt.Errorf("failed to read number of sorted arrays: %w", err)
	}

	for i := uint32(0); i < numSortedArrays; i++ {
		var arrayID uint32
		if err := binary.Read(r, binary.LittleEndian, &arrayID); err != nil {
			return fmt.Errorf("failed to read sorted array ID: %w", err)
		}

		// Read folder indices
		folderIndices := make([]uint32, db.metadata.numFolders)
		if err := binary.Read(r, binary.LittleEndian, folderIndices); err != nil {
			return fmt.Errorf("failed to read folder indices for array %d: %w", arrayID, err)
		}

		// Read file indices
		fileIndices := make([]uint32, db.metadata.numFiles)
		if err := binary.Read(r, binary.LittleEndian, fileIndices); err != nil {
			return fmt.Errorf("failed to read file indices for array %d: %w", arrayID, err)
		}

		db.SortedArrays[arrayID] = &SortedArray{
			ID:      arrayID,
			Folders: folderIndices,
			Files:   fileIndices,
		}
	}

	return nil
}

// GetFullPath returns the full path of an entry by traversing parent folders.
// Paths are cached after first computation for performance.
func (e *Entry) GetFullPath() string {
	// Check cache first (if we have access to the database)
	// Note: We can't access the database from Entry, so we'll cache at the Entry level
	// For now, we'll use a simpler approach: cache in the Entry itself if it has a parent reference
	
	// For entries without parent, return immediately (no caching needed)
	if e.Parent == nil {
		if e.Name == "" {
			return "/"
		}
		return e.Name
	}

	// Check if parent's path is cached (we'll use a different approach)
	// Since Entry doesn't have direct access to Database, we'll optimize by
	// building the path more efficiently and caching at the Database level
	
	// Build path efficiently using strings.Builder
	var builder strings.Builder
	builder.Grow(256) // Pre-allocate reasonable capacity
	
	// Collect path components (we'll build in reverse order)
	components := make([]string, 0, 10)
	components = append(components, e.Name)
	
	parent := e.Parent
	isRoot := false
	for parent != nil {
		if parent.Name == "" {
			isRoot = true
			break
		}
		components = append(components, parent.Name)
		parent = parent.Parent
	}
	
	// Build path from components (reverse order)
	if isRoot {
		builder.WriteByte('/')
	}
	for i := len(components) - 1; i >= 0; i-- {
		if i < len(components)-1 {
			builder.WriteByte('/')
		} else if isRoot && i == len(components)-1 {
			// Already wrote '/' for root
		}
		builder.WriteString(components[i])
	}
	
	return builder.String()
}

// getFullPathCached returns the full path using the database's path cache.
// This is the optimized version that should be used when Database is available.
func (db *Database) getFullPathCached(e *Entry) string {
	// Check cache first
	if cached, ok := db.pathCache.Load(e); ok {
		return cached.(string)
	}
	
	// For entries without parent, cache and return immediately
	if e.Parent == nil {
		var path string
		if e.Name == "" {
			path = "/"
		} else {
			path = e.Name
		}
		db.pathCache.Store(e, path)
		return path
	}
	
	// Check if parent's path is cached
	var parentPath string
	if e.Parent != nil {
		if cached, ok := db.pathCache.Load(e.Parent); ok {
			parentPath = cached.(string)
		} else {
			// Recursively compute parent's path (will cache it)
			parentPath = db.getFullPathCached(&e.Parent.Entry)
		}
	}
	
	// Build this entry's path from parent
	var fullPath string
	if parentPath == "/" {
		fullPath = "/" + e.Name
	} else {
		fullPath = parentPath + "/" + e.Name
	}
	
	// Cache it
	db.pathCache.Store(e, fullPath)
	return fullPath
}

// GetFullPath returns the full path of a folder
func (f *Folder) GetFullPath() string {
	return f.Entry.GetFullPath()
}

