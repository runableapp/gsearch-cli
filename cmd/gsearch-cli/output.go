package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/gsearch-cli/internal/db"
)

type resultEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Type    string `json:"type"` // "file" or "folder"
	Size    int64  `json:"size,omitempty"`
	MTime   string `json:"mtime,omitempty"`
	MTimeTS int64  `json:"mtime_ts,omitempty"`
}

// sortResults sorts the search results by the specified field
func sortResults(result *db.SearchResult, field sortField) {
	switch field {
	case sortFieldName:
		sort.Slice(result.Files, func(i, j int) bool {
			return result.Files[i].Name < result.Files[j].Name
		})
		sort.Slice(result.Folders, func(i, j int) bool {
			return result.Folders[i].Name < result.Folders[j].Name
		})
	case sortFieldPath:
		sort.Slice(result.Files, func(i, j int) bool {
			return result.Files[i].GetFullPath() < result.Files[j].GetFullPath()
		})
		sort.Slice(result.Folders, func(i, j int) bool {
			return result.Folders[i].GetFullPath() < result.Folders[j].GetFullPath()
		})
	case sortFieldSize:
		sort.Slice(result.Files, func(i, j int) bool {
			return result.Files[i].Size < result.Files[j].Size
		})
		// Folders don't have meaningful size for sorting
		sort.Slice(result.Folders, func(i, j int) bool {
			return result.Folders[i].Name < result.Folders[j].Name
		})
	case sortFieldMTime:
		sort.Slice(result.Files, func(i, j int) bool {
			return result.Files[i].MTime.Before(result.Files[j].MTime)
		})
		sort.Slice(result.Folders, func(i, j int) bool {
			return result.Folders[i].MTime.Before(result.Folders[j].MTime)
		})
	}
}

// printResults prints search results in the specified format
func printResults(result *db.SearchResult, format outputFormat) {
	total := len(result.Files) + len(result.Folders)
	if total == 0 {
		switch format {
		case outputFormatJSON:
			fmt.Println("[]")
		case outputFormatCSV:
			// Print header only
			w := csv.NewWriter(os.Stdout)
			w.Write([]string{"name", "path", "type", "size", "mtime"})
			w.Flush()
		default:
			fmt.Println("No results found.")
		}
		return
	}

	switch format {
	case outputFormatJSON:
		printJSON(result)
	case outputFormatCSV:
		printCSV(result)
	default:
		printText(result)
	}
}

func printJSON(result *db.SearchResult) {
	entries := make([]resultEntry, 0, len(result.Files)+len(result.Folders))

	// Add folders
	for _, folder := range result.Folders {
		path := folder.GetFullPath()
		if path == "" {
			path = "/"
		}
		entry := resultEntry{
			Name:    folder.Name,
			Path:    path,
			Type:    "folder",
			MTime:   folder.MTime.Format(time.RFC3339),
			MTimeTS: folder.MTime.Unix(),
		}
		entries = append(entries, entry)
	}

	// Add files
	for _, file := range result.Files {
		path := file.GetFullPath()
		entry := resultEntry{
			Name:    file.Name,
			Path:    path,
			Type:    "file",
			Size:    file.Size,
			MTime:   file.MTime.Format(time.RFC3339),
			MTimeTS: file.MTime.Unix(),
		}
		entries = append(entries, entry)
	}

	jsonData, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonData))
}

func printCSV(result *db.SearchResult) {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Write header
	w.Write([]string{"name", "path", "type", "size", "mtime"})

	// Write folders
	for _, folder := range result.Folders {
		path := folder.GetFullPath()
		if path == "" {
			path = "/"
		}
		w.Write([]string{
			folder.Name,
			path,
			"folder",
			"",
			folder.MTime.Format(time.RFC3339),
		})
	}

	// Write files
	for _, file := range result.Files {
		path := file.GetFullPath()
		w.Write([]string{
			file.Name,
			path,
			"file",
			fmt.Sprintf("%d", file.Size),
			file.MTime.Format(time.RFC3339),
		})
	}
}

func printText(result *db.SearchResult) {
	total := len(result.Files) + len(result.Folders)
	fmt.Printf("Found %d result(s):\n\n", total)

	// Print folders first
	for _, folder := range result.Folders {
		path := folder.GetFullPath()
		if path == "" {
			path = "/"
		}
		fmt.Printf("📁 %s\n", path)
	}

	// Print files
	for _, file := range result.Files {
		path := file.GetFullPath()
		fmt.Printf("📄 %s", path)
		if file.Size > 0 {
			fmt.Printf(" (%s)", formatSize(file.Size))
		}
		fmt.Println()
	}
}

