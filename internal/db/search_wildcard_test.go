package db

import (
	"testing"
)

func TestHasWildcards(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"test", false},
		{"test*", true},
		{"*test", true},
		{"test?", true},
		{"?test", true},
		{"test*file", true},
		{"test?file", true},
		{"test.*", true},
		{"test.txt", false},
		{"", false},
		{"*", true},
		{"?", true},
		{"**", true},
		{"??", true},
	}

	for _, tt := range tests {
		result := hasWildcards(tt.input)
		if result != tt.expected {
			t.Errorf("hasWildcards(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestConvertWildcardToRegex(t *testing.T) {
	tests := []struct {
		wildcard string
		expected string
	}{
		{"test", "^test$"},
		{"test*", "^test.*$"},
		{"*test", "^.*test$"},
		{"test?", "^test.$"},
		{"?test", "^.test$"},
		{"test*file", "^test.*file$"},
		{"test?file", "^test.file$"},
		{"*.txt", "^.*\\.txt$"},
		{"test.*", "^test\\..*$"},
		{"file?.txt", "^file.\\.txt$"},
		{"*.*", "^.*\\..*$"},
		{"test[file]", "^test\\[file\\]$"},
		{"test(file)", "^test\\(file\\)$"},
		{"test{file}", "^test\\{file\\}$"},
		{"test^file", "^test\\^file$"},
		{"test$file", "^test\\$file$"},
		{"test+file", "^test\\+file$"},
		{"test|file", "^test\\|file$"},
		{"test\\file", "^test\\\\file$"},
	}

	for _, tt := range tests {
		result := convertWildcardToRegex(tt.wildcard)
		if result != tt.expected {
			t.Errorf("convertWildcardToRegex(%q) = %q, want %q", tt.wildcard, result, tt.expected)
		}
	}
}

func TestWildcardSearch(t *testing.T) {
	// Load test database
	dbPath := setupTestDB(t)
	db, err := Load(dbPath)
	if err != nil {
		t.Fatalf("Failed to load test database: %v", err)
	}

	tests := []struct {
		name     string
		query    string
		expected int // expected number of files
	}{
		{"Match all .txt files", "*.txt", 2},
		{"Match all .go files", "*.go", 1},
		{"Match all .pdf files", "*.pdf", 1},
		{"Match files starting with test", "test*", 2},
		{"Match files with single char before .txt", "?*.txt", 2}, // Matches test.txt and readme.txt (one char + any chars + .txt)
		{"Match readme.txt", "readme.txt", 1},
		{"Match test.txt", "test.txt", 1},
		{"Match test.go", "test.go", 1},
		{"Match files ending with .zip", "*.zip", 1},
		{"Match files with 'test' anywhere", "*test*", 2},
		{"Match single char + .txt", "?.txt", 0}, // No single char .txt files
		{"Match document.pdf", "document.pdf", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := SearchOptions{
				Query:           tt.query,
				CaseSensitive:   false,
				SearchInFiles:   true,
				SearchInFolders: false,
			}

			result := db.Search(opts)
			if len(result.Files) != tt.expected {
				t.Errorf("Query %q: expected %d files, got %d", tt.query, tt.expected, len(result.Files))
				for _, file := range result.Files {
					t.Logf("  Found: %s", file.Name)
				}
			}
		})
	}
}

func TestWildcardSearchFolders(t *testing.T) {
	// Load test database
	dbPath := setupTestDB(t)
	db, err := Load(dbPath)
	if err != nil {
		t.Fatalf("Failed to load test database: %v", err)
	}

	tests := []struct {
		name     string
		query    string
		expected int // expected number of folders
	}{
		{"Match all folders starting with D", "D*", 2}, // Documents, Downloads
		{"Match folders ending with s", "*s", 2},      // Documents, Downloads
		{"Match 'home' folder", "home", 1},
		{"Match 'user' folder", "user", 1},
		{"Match folders with 'doc'", "*doc*", 1}, // Documents
		{"Match single char folders", "?", 0},    // No single char folders
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := SearchOptions{
				Query:           tt.query,
				CaseSensitive:   false,
				SearchInFiles:   false,
				SearchInFolders: true,
			}

			result := db.Search(opts)
			if len(result.Folders) != tt.expected {
				t.Errorf("Query %q: expected %d folders, got %d", tt.query, tt.expected, len(result.Folders))
				for _, folder := range result.Folders {
					t.Logf("  Found: %s", folder.Name)
				}
			}
		})
	}
}

func TestWildcardSearchByPath(t *testing.T) {
	// Load test database
	dbPath := setupTestDB(t)
	db, err := Load(dbPath)
	if err != nil {
		t.Fatalf("Failed to load test database: %v", err)
	}

	tests := []struct {
		name     string
		pattern  string
		expected int // expected total results
	}{
		{"Match paths with /home/*", "/home/*", 3}, // /home/user folder and 2 files in /home/user
		{"Match paths ending with .txt", "*.txt", 2},
		{"Match paths with user/*", "*/user/*", 2}, // Files in /home/user
		{"Match paths with Documents", "*Documents*", 3}, // Documents folder and 2 files in it (/Documents, /Documents/document.pdf, /Documents/test.go)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.SearchByPath(tt.pattern, false)
			total := len(result.Files) + len(result.Folders)
			if total != tt.expected {
				t.Errorf("Pattern %q: expected %d results, got %d (files: %d, folders: %d)",
					tt.pattern, tt.expected, total, len(result.Files), len(result.Folders))
			}
		})
	}
}

func TestWildcardCaseSensitive(t *testing.T) {
	// Load test database
	dbPath := setupTestDB(t)
	db, err := Load(dbPath)
	if err != nil {
		t.Fatalf("Failed to load test database: %v", err)
	}

	// Case-insensitive wildcard search
	opts1 := SearchOptions{
		Query:           "*.TXT",
		CaseSensitive:   false,
		SearchInFiles:   true,
		SearchInFolders: false,
	}
	result1 := db.Search(opts1)
	if len(result1.Files) != 2 {
		t.Errorf("Case-insensitive *.TXT: expected 2 files, got %d", len(result1.Files))
	}

	// Case-sensitive wildcard search
	opts2 := SearchOptions{
		Query:           "*.TXT",
		CaseSensitive:   true,
		SearchInFiles:   true,
		SearchInFolders: false,
	}
	result2 := db.Search(opts2)
	if len(result2.Files) != 0 {
		t.Errorf("Case-sensitive *.TXT: expected 0 files (all are .txt lowercase), got %d", len(result2.Files))
	}
}

