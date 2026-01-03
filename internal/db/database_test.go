package db

import (
	"path/filepath"
	"testing"
)

// This file uses Go's standard testing package.
// Test functions are named TestXxx and take *testing.T as parameter.
// This is the standard Go unit testing approach.

// setupTestDB creates a temporary test database file and returns its path
func setupTestDB(t *testing.T) string {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	if err := CreateTestDatabase(dbPath); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return dbPath
}

func TestLoadDatabase(t *testing.T) {
	dbPath := setupTestDB(t)

	db, err := Load(dbPath)
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	if db == nil {
		t.Fatal("Database is nil")
	}

	// Verify expected test data
	if len(db.Folders) != 5 {
		t.Errorf("Expected 5 folders, got %d", len(db.Folders))
	}

	if len(db.Files) != 5 {
		t.Errorf("Expected 5 files, got %d", len(db.Files))
	}

	// Verify index flags
	expectedFlags := IndexFlagName | IndexFlagSize | IndexFlagModificationTime
	if db.IndexFlags != expectedFlags {
		t.Errorf("Expected index flags %d, got %d", expectedFlags, db.IndexFlags)
	}

	// Verify folder names
	expectedFolderNames := []string{"", "home", "user", "Documents", "Downloads"}
	for i, expected := range expectedFolderNames {
		if i >= len(db.Folders) {
			t.Fatalf("Not enough folders, expected at least %d", i+1)
		}
		if db.Folders[i].Name != expected {
			t.Errorf("Folder %d: expected name %q, got %q", i, expected, db.Folders[i].Name)
		}
	}

	// Verify file names
	expectedFileNames := []string{"test.txt", "readme.txt", "document.pdf", "test.go", "file.zip"}
	for i, expected := range expectedFileNames {
		if i >= len(db.Files) {
			t.Fatalf("Not enough files, expected at least %d", i+1)
		}
		if db.Files[i].Name != expected {
			t.Errorf("File %d: expected name %q, got %q", i, expected, db.Files[i].Name)
		}
	}
}

func TestEntryGetFullPath(t *testing.T) {
	// Create a test folder hierarchy
	root := &Folder{
		Entry: Entry{
			Name:  "",
			Index: 0,
			Type:  EntryTypeFolder,
		},
	}

	home := &Folder{
		Entry: Entry{
			Name:   "home",
			Parent: root,
			Index:  1,
			Type:   EntryTypeFolder,
		},
	}

	user := &Folder{
		Entry: Entry{
			Name:   "user",
			Parent: home,
			Index:  2,
			Type:   EntryTypeFolder,
		},
	}

	file := &Entry{
		Name:   "test.txt",
		Parent: user,
		Index:  0,
		Type:   EntryTypeFile,
	}

	// Test root folder path (empty name means root, should show as "/")
	rootPath := root.GetFullPath()
	if rootPath != "/" {
		t.Errorf("Expected root path to be \"/\", got %q", rootPath)
	}

	// Test nested folder path (root has empty name, so path starts with "/")
	homePath := home.GetFullPath()
	expectedHomePath := "/home"
	if homePath != expectedHomePath {
		t.Errorf("Expected home path %q, got %q", expectedHomePath, homePath)
	}

	// Test deeper folder path
	userPath := user.GetFullPath()
	expectedUserPath := "/home/user"
	if userPath != expectedUserPath {
		t.Errorf("Expected user path %q, got %q", expectedUserPath, userPath)
	}

	// Test file path
	filePath := file.GetFullPath()
	expectedFilePath := "/home/user/test.txt"
	if filePath != expectedFilePath {
		t.Errorf("Expected file path %q, got %q", expectedFilePath, filePath)
	}
}

func TestSearch(t *testing.T) {
	// Load test database
	dbPath := setupTestDB(t)
	db, err := Load(dbPath)
	if err != nil {
		t.Fatalf("Failed to load test database: %v", err)
	}

	// Test case-insensitive search for "test"
	opts := SearchOptions{
		Query:           "test",
		CaseSensitive:   false,
		SearchInFiles:   true,
		SearchInFolders: true,
	}

	result := db.Search(opts)
	// Should find: test.txt, test.go
	if len(result.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result.Files))
	}

	// Verify we found the expected files
	foundNames := make(map[string]bool)
	for _, file := range result.Files {
		foundNames[file.Name] = true
	}
	if !foundNames["test.txt"] {
		t.Error("Expected to find test.txt")
	}
	if !foundNames["test.go"] {
		t.Error("Expected to find test.go")
	}

	// Test case-sensitive search
	opts.CaseSensitive = true
	opts.Query = "Test"
	result = db.Search(opts)
	if len(result.Files) != 0 {
		t.Errorf("Expected 0 files (case-sensitive), got %d", len(result.Files))
	}

	// Test folder search for "doc"
	opts.Query = "doc"
	opts.CaseSensitive = false
	result = db.Search(opts)
	if len(result.Folders) != 1 {
		t.Errorf("Expected 1 folder, got %d", len(result.Folders))
	}
	if result.Folders[0].Name != "Documents" {
		t.Errorf("Expected Documents folder, got %s", result.Folders[0].Name)
	}
}

func TestMatchWholeWord(t *testing.T) {
	db := &Database{}

	tests := []struct {
		text   string
		query  string
		result bool
	}{
		{"test file", "test", true},
		{"testfile", "test", false},
		{"my test file", "test", true},
		{"testing", "test", false},
		{"test", "test", true},
		{"test 123", "test", true},
		{"123test", "test", false},
	}

	for _, tt := range tests {
		opts := SearchOptions{
			Query:          tt.query,
			CaseSensitive:  false,
			MatchWholeWord: true,
		}
		result := db.matches(tt.text, tt.query, opts)
		if result != tt.result {
			t.Errorf("matches(%q, %q, wholeWord=true) = %v, want %v", tt.text, tt.query, result, tt.result)
		}
	}
}
