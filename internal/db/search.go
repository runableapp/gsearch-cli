package db

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SearchOptions contains options for searching the database
type SearchOptions struct {
	Query           string
	CaseSensitive   bool
	MatchWholeWord  bool
	SearchInFiles   bool
	SearchInFolders bool
	MaxResults      int // 0 = unlimited
}

// SearchResult contains the results of a search
type SearchResult struct {
	Files   []*Entry
	Folders []*Folder
}

// Search performs a search on the database
func (db *Database) Search(opts SearchOptions) *SearchResult {
	result := &SearchResult{
		Files:   make([]*Entry, 0),
		Folders: make([]*Folder, 0),
	}

	if opts.Query == "" {
		return result
	}

	// Normalize query
	query := opts.Query
	if !opts.CaseSensitive {
		query = strings.ToLower(query)
	}

	// Search files
	if opts.SearchInFiles {
		for _, file := range db.Files {
			if db.matches(file.Name, query, opts) {
				result.Files = append(result.Files, file)
				if opts.MaxResults > 0 && len(result.Files) >= opts.MaxResults {
					break
				}
			}
		}
	}

	// Search folders
	if opts.SearchInFolders {
		for _, folder := range db.Folders {
			if db.matches(folder.Name, query, opts) {
				result.Folders = append(result.Folders, folder)
				if opts.MaxResults > 0 && len(result.Folders)+len(result.Files) >= opts.MaxResults {
					break
				}
			}
		}
	}

	return result
}

// matches checks if a string matches the query based on the search options
func (db *Database) matches(text, query string, opts SearchOptions) bool {
	if !opts.CaseSensitive {
		text = strings.ToLower(text)
	}

	if opts.MatchWholeWord {
		// Check if query appears as a whole word
		return db.matchWholeWord(text, query)
	}

	// Simple substring match
	return strings.Contains(text, query)
}

// matchWholeWord checks if query appears as a complete word in text
func (db *Database) matchWholeWord(text, query string) bool {
	// Find all occurrences of query in text
	idx := 0
	for {
		pos := strings.Index(text[idx:], query)
		if pos == -1 {
			return false
		}
		pos += idx

		// Check if it's a whole word
		// Word boundary: start of string or non-word char before, end of string or non-word char after
		before := pos == 0
		if !before {
			r, _ := utf8.DecodeLastRuneInString(text[:pos])
			before = !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
		}

		after := pos+len(query) == len(text)
		if !after {
			r, _ := utf8.DecodeRuneInString(text[pos+len(query):])
			after = !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
		}

		if before && after {
			return true
		}

		idx = pos + 1
	}
}

// SearchByPath searches for entries matching a path pattern
func (db *Database) SearchByPath(pattern string, caseSensitive bool) *SearchResult {
	result := &SearchResult{
		Files:   make([]*Entry, 0),
		Folders: make([]*Folder, 0),
	}

	if !caseSensitive {
		pattern = strings.ToLower(pattern)
	}

	// Search files
	for _, file := range db.Files {
		path := file.GetFullPath()
		if !caseSensitive {
			path = strings.ToLower(path)
		}
		if strings.Contains(path, pattern) {
			result.Files = append(result.Files, file)
		}
	}

	// Search folders
	for _, folder := range db.Folders {
		path := folder.GetFullPath()
		if !caseSensitive {
			path = strings.ToLower(path)
		}
		if strings.Contains(path, pattern) {
			result.Folders = append(result.Folders, folder)
		}
	}

	return result
}

