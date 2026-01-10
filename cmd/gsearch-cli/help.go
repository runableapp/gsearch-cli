package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gsearch-cli/version"
)

func showUsage() {
	programName := "gsearch-cli"
	if len(os.Args) > 0 {
		programName = filepath.Base(os.Args[0])
	}

	fmt.Fprintf(os.Stderr, `%s - Command-line interface for searching FSearch database

USAGE:
    %s [OPTIONS]

DESCRIPTION:
    Search the FSearch database file for files and folders matching your query.
    Supports wildcard patterns (* and ?) and various output formats.

SEARCH OPTIONS:
    -q, -query <query>
        Search query (required unless using -path)
        Supports wildcard patterns: * (any sequence) and ? (single character)
        Examples: "test", "*.txt", "test*", "file?.go"

    -path <pattern>
        Search in full path instead of just name
        Also supports wildcard patterns
        Examples: "/home/user", "/home/*", "*.txt"

    -case
        Enable case-sensitive search (default: false)

    -whole
        Match whole words only (default: false)

    -files
        Search only files (exclude folders)

    -folders
        Search only folders (exclude files)

    -max <n>
        Maximum number of results (0 = unlimited, default: 0)

OUTPUT OPTIONS:
    -output <format>
        Output format: text, json, or csv (default: text)
        - text: Human-readable format with emojis
        - json: JSON array with fields: name, path, type, size, mtime
        - csv: CSV format with header row

    -sort <field>
        Sort results by field: name, path, size, or mtime (default: no sorting)
        - name: Sort by file/folder name
        - path: Sort by full path
        - size: Sort by file size (files only, folders sorted by name)
        - mtime: Sort by modification time

DATABASE OPTIONS:
    -db <path>
        Path to FSearch database file
        Default: ~/.local/share/fsearch/fsearch.db

    -stats
        Show database statistics instead of searching

HELP:
    -h, -help
        Show this help message

EXAMPLES:
    # Basic search
    %s -q test

    # Search with wildcard pattern
    %s -q "*.txt"
    %s -q "test*"

    # Case-sensitive search
    %s -q Test -case

    # Search in path
    %s -path /home/user

    # Wildcard path search
    %s -path "/home/*"

    # Output as JSON
    %s -q test -output json

    # Output as CSV
    %s -q test -output csv

    # Sort by size
    %s -q test -sort size

    # Sort by modification time
    %s -q test -sort mtime

    # Combine options
    %s -q "*.go" -files -sort size -output json

    # Show database statistics
    %s -stats

WILDCARD PATTERNS:
    *       Matches any sequence of characters (zero or more)
    ?       Matches a single character

    Examples:
        *.txt        Matches all .txt files
        test*        Matches files starting with "test"
        ?.go         Matches single character + .go (e.g., "a.go")
        *test*       Matches files with "test" anywhere

OUTPUT FORMATS:
    text (default):
        Human-readable format with folder/file indicators
        Example:
            Found 2 result(s):
            📁 /Documents
            📄 /home/user/test.txt (1.0 KB)

    json:
        JSON array with structured data
        Example:
            [
              {
                "name": "test.txt",
                "path": "/home/user/test.txt",
                "type": "file",
                "size": 1024,
                "mtime": "2024-01-03T12:00:00Z",
                "mtime_ts": 1704283200
              }
            ]

    csv:
        CSV format with header row
        Example:
            name,path,type,size,mtime
            test.txt,/home/user/test.txt,file,1024,2024-01-03T12:00:00Z

SORTING:
    Results can be sorted by:
    - name: Alphabetical by file/folder name
    - path: Alphabetical by full path
    - size: By file size (ascending), folders sorted by name
    - mtime: By modification time (oldest first)

    Note: Sorting applies to both files and folders together.

`, programName, programName, programName, programName, programName, programName, programName, programName, programName, programName, programName, programName, programName, programName)

	fmt.Fprintf(os.Stderr, "\n%s v%s\n", programName, version.Get())
	fmt.Fprintf(os.Stderr, "Copyright © 2026 Runable.app. All rights reserved.\n")
}
