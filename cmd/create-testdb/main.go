package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gsearch-cli/internal/db"
)

func main() {
	var outputPath = flag.String("o", "test.db", "Output path for test database file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Create a test FSearch database file with sample data.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := db.CreateTestDatabase(*outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create test database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test database created successfully: %s\n", *outputPath)
	fmt.Printf("Contains: 5 folders, 5 files\n")
}

