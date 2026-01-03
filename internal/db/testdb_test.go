package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateTestDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create test database
	if err := CreateTestDatabase(dbPath); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("Test database file was not created")
	}

	// Try to load it
	db, err := Load(dbPath)
	if err != nil {
		t.Fatalf("Failed to load created test database: %v", err)
	}

	// Verify structure
	if db == nil {
		t.Fatal("Loaded database is nil")
	}

	// Verify we have the expected number of entries
	expectedFolders := 5
	expectedFiles := 5

	if len(db.Folders) != expectedFolders {
		t.Errorf("Expected %d folders, got %d", expectedFolders, len(db.Folders))
	}

	if len(db.Files) != expectedFiles {
		t.Errorf("Expected %d files, got %d", expectedFiles, len(db.Files))
	}

	// Verify index flags
	expectedFlags := IndexFlagName | IndexFlagSize | IndexFlagModificationTime
	if db.IndexFlags != expectedFlags {
		t.Errorf("Expected index flags %d, got %d", expectedFlags, db.IndexFlags)
	}
}

func TestCreateTestDatabaseMultipleTimes(t *testing.T) {
	// Test that we can create multiple test databases
	for i := 0; i < 3; i++ {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		if err := CreateTestDatabase(dbPath); err != nil {
			t.Fatalf("Failed to create test database #%d: %v", i+1, err)
		}

		db, err := Load(dbPath)
		if err != nil {
			t.Fatalf("Failed to load test database #%d: %v", i+1, err)
		}

		if len(db.Folders) != 5 || len(db.Files) != 5 {
			t.Errorf("Test database #%d has incorrect structure", i+1)
		}
	}
}

