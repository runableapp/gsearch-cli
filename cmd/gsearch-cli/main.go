package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gsearch-cli/internal/db"
)

const (
	defaultDBPath = "~/.local/share/fsearch/fsearch.db"
)

func main() {
	var (
		dbPath         = flag.String("db", defaultDBPath, "Path to fsearch database file")
		query          = flag.String("q", "", "Search query")
		caseSensitive  = flag.Bool("case", false, "Case-sensitive search")
		wholeWord      = flag.Bool("whole", false, "Match whole words only")
		searchPath     = flag.String("path", "", "Search in full path (instead of just name)")
		filesOnly      = flag.Bool("files", false, "Search only files")
		foldersOnly    = flag.Bool("folders", false, "Search only folders")
		maxResults = flag.Int("max", 0, "Maximum number of results (0 = unlimited)")
		showStats  = flag.Bool("stats", false, "Show database statistics")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Search the FSearch database from the command line.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -q test\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -q test -files\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -path /home/user\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -stats\n", os.Args[0])
	}

	flag.Parse()

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

	// Print results
	printResults(result)
}

func showDatabaseStats(database *db.Database) {
	fmt.Printf("Database Statistics:\n")
	fmt.Printf("  Folders: %d\n", len(database.Folders))
	fmt.Printf("  Files: %d\n", len(database.Files))
	fmt.Printf("  Total entries: %d\n", len(database.Folders)+len(database.Files))
	fmt.Printf("  Index flags: %d\n", database.IndexFlags)
	fmt.Printf("  Sorted arrays: %d\n", len(database.SortedArrays))
}

func printResults(result *db.SearchResult) {
	total := len(result.Files) + len(result.Folders)
	if total == 0 {
		fmt.Println("No results found.")
		return
	}

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

