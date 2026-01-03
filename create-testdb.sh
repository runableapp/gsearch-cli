#!/bin/bash
# Create a test FSearch database file
# Usage: ./create-testdb.sh [options]
#   -o <path>  Output path for test database file (default: test.db)

set -e

# Default output path
OUTPUT_PATH="test.db"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -o|--output)
            OUTPUT_PATH="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Create a test FSearch database file with sample data."
            echo ""
            echo "Options:"
            echo "  -o, --output <path>  Output path for test database file (default: test.db)"
            echo "  -h, --help           Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0"
            echo "  $0 -o /tmp/test.db"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Run the Go program
go run "$SCRIPT_DIR/cmd/create-testdb/main.go" -o "$OUTPUT_PATH"

