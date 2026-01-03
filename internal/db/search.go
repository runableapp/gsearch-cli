package db

import (
	"regexp"
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

	// Keep original query for wildcard detection (case conversion happens in matches())
	query := opts.Query

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

// hasWildcards checks if a string contains wildcard characters (* or ?)
func hasWildcards(s string) bool {
	return strings.Contains(s, "*") || strings.Contains(s, "?")
}

// convertWildcardToRegex converts a wildcard pattern to a regex pattern
// * becomes .* (matches any sequence)
// ? becomes . (matches single character)
// Special regex characters are escaped
// Pattern is anchored with ^ and $ for full string matching
func convertWildcardToRegex(pattern string) string {
	var result strings.Builder
	result.WriteString("^") // Anchor start
	
	for _, char := range pattern {
		switch char {
		case '*':
			result.WriteString(".*")
		case '?':
			result.WriteString(".")
		case '.', '^', '$', '+', '(', ')', '[', ']', '{', '}', '\\', '|':
			// Escape special regex characters
			result.WriteRune('\\')
			result.WriteRune(char)
		default:
			result.WriteRune(char)
		}
	}
	
	result.WriteString("$") // Anchor end
	return result.String()
}

// matches checks if a string matches the query based on the search options
func (db *Database) matches(text, query string, opts SearchOptions) bool {
	// Check for wildcard patterns (before case conversion)
	if hasWildcards(query) {
		regexPattern := convertWildcardToRegex(query)
		var re *regexp.Regexp
		var err error
		if opts.CaseSensitive {
			re, err = regexp.Compile(regexPattern)
		} else {
			// Compile case-insensitive regex
			re, err = regexp.Compile("(?i)" + regexPattern)
		}
		if err != nil {
			// If regex compilation fails, fall back to substring match
			if !opts.CaseSensitive {
				text = strings.ToLower(text)
				query = strings.ToLower(query)
			}
			return strings.Contains(text, query)
		}
		// Regex handles case sensitivity internally
		return re.MatchString(text)
	}

	if !opts.CaseSensitive {
		text = strings.ToLower(text)
		query = strings.ToLower(query)
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
// Supports wildcard patterns (* and ?)
func (db *Database) SearchByPath(pattern string, caseSensitive bool) *SearchResult {
	result := &SearchResult{
		Files:   make([]*Entry, 0),
		Folders: make([]*Folder, 0),
	}

	// Check for wildcard patterns (before case conversion)
	useWildcard := hasWildcards(pattern)
	var re *regexp.Regexp
	if useWildcard {
		// For wildcard patterns, we need to handle case sensitivity in the regex
		regexPattern := convertWildcardToRegex(pattern)
		var err error
		if caseSensitive {
			re, err = regexp.Compile(regexPattern)
		} else {
			// Compile case-insensitive regex
			re, err = regexp.Compile("(?i)" + regexPattern)
		}
		if err != nil {
			// If regex compilation fails, fall back to substring match
			useWildcard = false
		}
	}

	if !caseSensitive && !useWildcard {
		pattern = strings.ToLower(pattern)
	}

	// Search files
	for _, file := range db.Files {
		path := file.GetFullPath()
		var matches bool
		
		if useWildcard {
			// Regex handles case sensitivity internally
			matches = re.MatchString(path)
		} else {
			if !caseSensitive {
				path = strings.ToLower(path)
			}
			matches = strings.Contains(path, pattern)
		}
		
		if matches {
			result.Files = append(result.Files, file)
		}
	}

	// Search folders
	for _, folder := range db.Folders {
		path := folder.GetFullPath()
		var matches bool
		
		if useWildcard {
			// Regex handles case sensitivity internally
			matches = re.MatchString(path)
		} else {
			if !caseSensitive {
				path = strings.ToLower(path)
			}
			matches = strings.Contains(path, pattern)
		}
		
		if matches {
			result.Folders = append(result.Folders, folder)
		}
	}

	return result
}

