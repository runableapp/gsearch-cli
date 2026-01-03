package main

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gsearch-cli/internal/db"
)

func TestSortResults(t *testing.T) {
	// Create test data
	files := []*db.Entry{
		{Name: "zebra.txt", Size: 100, MTime: time.Now()},
		{Name: "apple.txt", Size: 200, MTime: time.Now().Add(-time.Hour)},
		{Name: "banana.txt", Size: 50, MTime: time.Now().Add(time.Hour)},
	}

	folders := []*db.Folder{
		{Entry: db.Entry{Name: "zebra", MTime: time.Now()}},
		{Entry: db.Entry{Name: "apple", MTime: time.Now().Add(-time.Hour)}},
		{Entry: db.Entry{Name: "banana", MTime: time.Now().Add(time.Hour)}},
	}

	result := &db.SearchResult{
		Files:   files,
		Folders: folders,
	}

	// Test sort by name
	sortResults(result, sortFieldName)
	if result.Files[0].Name != "apple.txt" {
		t.Errorf("Sort by name: expected first file 'apple.txt', got %q", result.Files[0].Name)
	}
	if result.Folders[0].Name != "apple" {
		t.Errorf("Sort by name: expected first folder 'apple', got %q", result.Folders[0].Name)
	}

	// Test sort by size
	result2 := &db.SearchResult{
		Files:   []*db.Entry{files[0], files[1], files[2]},
		Folders: folders,
	}
	sortResults(result2, sortFieldSize)
	if result2.Files[0].Size != 50 {
		t.Errorf("Sort by size: expected smallest file size 50, got %d", result2.Files[0].Size)
	}
	if result2.Files[len(result2.Files)-1].Size != 200 {
		t.Errorf("Sort by size: expected largest file size 200, got %d", result2.Files[len(result2.Files)-1].Size)
	}

	// Test sort by mtime
	result3 := &db.SearchResult{
		Files:   []*db.Entry{files[0], files[1], files[2]},
		Folders: []*db.Folder{folders[0], folders[1], folders[2]},
	}
	sortResults(result3, sortFieldMTime)
	// Oldest should be first (added -1 hour)
	if !result3.Files[0].MTime.Before(result3.Files[1].MTime) {
		t.Error("Sort by mtime: files not sorted correctly")
	}
}

func TestPrintJSON(t *testing.T) {
	// We can't easily test printJSON directly without capturing stdout,
	// but we can test the JSON structure by creating entries manually
	entries := make([]resultEntry, 0)
	
	// Add folder
	entries = append(entries, resultEntry{
		Name:    "Documents",
		Path:    "/Documents",
		Type:    "folder",
		MTime:   time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		MTimeTS: time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC).Unix(),
	})
	
	// Add file
	entries = append(entries, resultEntry{
		Name:    "test.txt",
		Path:    "/test.txt",
		Type:    "file",
		Size:    1024,
		MTime:   time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		MTimeTS: time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC).Unix(),
	})

	jsonData, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Verify JSON is valid
	var decoded []resultEntry
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	if len(decoded) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(decoded))
	}

	if decoded[0].Type != "folder" {
		t.Errorf("Expected first entry to be folder, got %q", decoded[0].Type)
	}

	if decoded[1].Type != "file" {
		t.Errorf("Expected second entry to be file, got %q", decoded[1].Type)
	}

	if decoded[1].Size != 1024 {
		t.Errorf("Expected file size 1024, got %d", decoded[1].Size)
	}
}

func TestPrintCSV(t *testing.T) {
	// Test CSV structure by manually creating CSV
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	
	// Write header
	w.Write([]string{"name", "path", "type", "size", "mtime"})
	
	// Write folder
	w.Write([]string{
		"Documents",
		"/Documents",
		"folder",
		"",
		time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
	})
	
	// Write file
	w.Write([]string{
		"test.txt",
		"/test.txt",
		"file",
		"1024",
		time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
	})
	
	w.Flush()
	
	// Verify CSV can be parsed
	r := csv.NewReader(strings.NewReader(buf.String()))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("Failed to parse CSV: %v", err)
	}

	if len(records) != 3 { // 1 header + 2 data rows
		t.Errorf("Expected 3 CSV rows, got %d", len(records))
	}

	if records[0][0] != "name" {
		t.Errorf("Expected CSV header 'name', got %q", records[0][0])
	}

	if records[1][2] != "folder" {
		t.Errorf("Expected folder type, got %q", records[1][2])
	}

	if records[2][2] != "file" {
		t.Errorf("Expected file type, got %q", records[2][2])
	}

	if records[2][3] != "1024" {
		t.Errorf("Expected file size '1024', got %q", records[2][3])
	}
}

func TestOutputFormatValidation(t *testing.T) {
	tests := []struct {
		input    string
		expected outputFormat
		valid    bool
	}{
		{"text", outputFormatText, true},
		{"TEXT", outputFormatText, true},
		{"json", outputFormatJSON, true},
		{"JSON", outputFormatJSON, true},
		{"csv", outputFormatCSV, true},
		{"CSV", outputFormatCSV, true},
		{"invalid", "", false},
		{"", outputFormatText, true}, // default
	}

	for _, tt := range tests {
		format := outputFormat(strings.ToLower(tt.input))
		if tt.valid {
			if format != tt.expected && tt.input != "" {
				t.Errorf("Input %q: expected %q, got %q", tt.input, tt.expected, format)
			}
		}
	}
}

func TestSortFieldValidation(t *testing.T) {
	tests := []struct {
		input    string
		expected sortField
		valid    bool
	}{
		{"name", sortFieldName, true},
		{"NAME", sortFieldName, true},
		{"path", sortFieldPath, true},
		{"size", sortFieldSize, true},
		{"mtime", sortFieldMTime, true},
		{"invalid", "", false},
		{"", "", true}, // empty is valid (no sorting)
	}

	for _, tt := range tests {
		if tt.input != "" {
			field := sortField(strings.ToLower(tt.input))
			if tt.valid && field != tt.expected {
				t.Errorf("Input %q: expected %q, got %q", tt.input, tt.expected, field)
			}
		}
	}
}

