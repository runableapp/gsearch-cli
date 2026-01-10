package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gsearch-cli/internal/db"
	"github.com/gsearch-cli/version"
)

const (
	defaultDBPath = "~/.local/share/fsearch/fsearch.db"
)

type outputFormat string

const (
	outputFormatText outputFormat = "text"
	outputFormatJSON outputFormat = "json"
	outputFormatCSV  outputFormat = "csv"
)

type sortField string

const (
	sortFieldName sortField = "name"
	sortFieldPath sortField = "path"
	sortFieldSize sortField = "size"
	sortFieldMTime sortField = "mtime"
)

func showVersion() {
	programName := "gsearch-cli"
	if len(os.Args) > 0 {
		programName = filepath.Base(os.Args[0])
	}
	fmt.Printf("%s v%s\n", programName, version.Get())
	fmt.Printf("Copyright © 2026 Runable.app. All rights reserved.\n")
}

func main() {
	// Check for "help" or "version" as single argument before flag parsing
	if len(os.Args) == 2 {
		arg := os.Args[1]
		if arg == "help" {
			showUsage()
			os.Exit(0)
		}
		if arg == "version" {
			showVersion()
			os.Exit(0)
		}
	}

	var (
		dbPath         = flag.String("db", defaultDBPath, "Path to fsearch database file")
		query          = flag.String("q", "", "Search query (supports wildcards: * and ?)")
		caseSensitive  = flag.Bool("case", false, "Case-sensitive search")
		wholeWord      = flag.Bool("whole", false, "Match whole words only")
		searchPath     = flag.String("path", "", "Search in full path (instead of just name, supports wildcards)")
		filesOnly      = flag.Bool("files", false, "Search only files")
		foldersOnly    = flag.Bool("folders", false, "Search only folders")
		maxResults     = flag.Int("max", 0, "Maximum number of results (0 = unlimited)")
		showStats      = flag.Bool("stats", false, "Show database statistics")
		outputFormatStr = flag.String("output", "text", "Output format: text, json, or csv")
		sortBy          = flag.String("sort", "", "Sort results by: name, path, size, or mtime")
		showHelp        = flag.Bool("help", false, "Show detailed help")
		flagHelp        = flag.Bool("h", false, "Show detailed help (alias for -help)")
	)

	flag.Usage = func() {
		showUsage()
	}

	flag.Parse()

	// Show help if requested
	if *showHelp || *flagHelp {
		showUsage()
		os.Exit(0)
	}

	// Validate output format
	format := outputFormat(strings.ToLower(*outputFormatStr))
	if format != outputFormatText && format != outputFormatJSON && format != outputFormatCSV {
		fmt.Fprintf(os.Stderr, "Error: invalid output format %q. Must be: text, json, or csv\n", *outputFormatStr)
		os.Exit(1)
	}

	// Validate sort field
	var sortFieldVal sortField
	if *sortBy != "" {
		sortFieldVal = sortField(strings.ToLower(*sortBy))
		if sortFieldVal != sortFieldName && sortFieldVal != sortFieldPath && sortFieldVal != sortFieldSize && sortFieldVal != sortFieldMTime {
			fmt.Fprintf(os.Stderr, "Error: invalid sort field %q. Must be: name, path, size, or mtime\n", *sortBy)
			os.Exit(1)
		}
	}

	// Expand ~ in path
	if strings.HasPrefix(*dbPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		*dbPath = filepath.Join(home, strings.TrimPrefix(*dbPath, "~"+string(filepath.Separator)))
	}

	// Load database
	database, err := db.Load(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load database: %v\n", err)
		os.Exit(1)
	}

	// Show statistics if requested
	if *showStats {
		showDatabaseStats(database)
		return
	}

	// Perform search
	if *query == "" && *searchPath == "" {
		fmt.Fprintf(os.Stderr, "Error: must provide either -q (query) or -path (path search)\n")
		flag.Usage()
		os.Exit(1)
	}

	var result *db.SearchResult
	if *searchPath != "" {
		result = database.SearchByPath(*searchPath, *caseSensitive)
	} else {
		opts := db.SearchOptions{
			Query:           *query,
			CaseSensitive:   *caseSensitive,
			MatchWholeWord:  *wholeWord,
			SearchInFiles:   !*foldersOnly,
			SearchInFolders: !*filesOnly,
			MaxResults:      *maxResults,
		}
		result = database.Search(opts)
	}

	// Sort results if requested
	if *sortBy != "" {
		sortResults(result, sortFieldVal)
	}

	// Print results in requested format
	printResults(result, format)
}

func showDatabaseStats(database *db.Database) {
	fmt.Printf("Database Statistics:\n")
	fmt.Printf("  Folders: %d\n", len(database.Folders))
	fmt.Printf("  Files: %d\n", len(database.Files))
	fmt.Printf("  Total entries: %d\n", len(database.Folders)+len(database.Files))
	fmt.Printf("  Index flags: %d\n", database.IndexFlags)
	fmt.Printf("  Sorted arrays: %d\n", len(database.SortedArrays))
}


func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

