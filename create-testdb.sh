#!/bin/bash
# Shell script wrapper for create-testdb tool
# Runs the Go source directly without building a binary

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_FILE="test.db"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [-o|--output OUTPUT_FILE]"
            echo ""
            echo "Create a test FSearch database file."
            echo ""
            echo "Options:"
            echo "  -o, --output FILE    Output file path (default: test.db)"
            echo "  -h, --help          Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Run the Go program
go run "$SCRIPT_DIR/cmd/create-testdb/main.go" -o "$OUTPUT_FILE"

