# gsearch-cli

A command-line interface for searching the FSearch database. This Go application reads and searches the binary database file created by [FSearch](https://github.com/cboxdoerfer/fsearch).

## Features

- Fast search through indexed files and folders
- Case-sensitive and case-insensitive search
- Whole word matching
- Search by name or full path
- Filter by files or folders only
- Display database statistics
- Human-readable file sizes

## Installation

### Prerequisites

- Go 1.21 or later
- Task (optional, for using Taskfile)

### Build

Using Task:
```bash
task build
```

Or using Go directly:
```bash
go build -o bin/gsearch-cli ./cmd/gsearch-cli
```

### Install

Using Task:
```bash
task install
```

Or using Go directly:
```bash
go install ./cmd/gsearch-cli
```

## Usage

### Basic Search

Search for files and folders by name:
```bash
gsearch-cli -q "test"
```

### Search Options

- `-q <query>`: Search query (required unless using `-path`)
- `-path <pattern>`: Search in full path instead of just name
- `-case`: Enable case-sensitive search (default: false)
- `-whole`: Match whole words only (default: false)
- `-files`: Search only files
- `-folders`: Search only folders
- `-max <n>`: Maximum number of results (0 = unlimited)
- `-db <path>`: Path to database file (default: `~/.local/share/fsearch/fsearch.db`)
- `-stats`: Show database statistics

### Examples

Search for files containing "test":
```bash
gsearch-cli -q test -files
```

Case-sensitive search:
```bash
gsearch-cli -q Test -case
```

Search in full path:
```bash
gsearch-cli -path /home/user
```

Show database statistics:
```bash
gsearch-cli -stats
```

Limit results:
```bash
gsearch-cli -q test -max 10
```

## Database Format

See [FSEARCH_DB.md](FSEARCH_DB.md) for detailed documentation of the database file format.

## Development

### Testing

This project uses Go's standard testing package. All tests are self-contained and automatically create test database files using `CreateTestDatabase()`. No real FSearch database file is required for testing.

#### Running Tests

**Run all unit tests:**
```bash
task test
```

Or using the alias:
```bash
task test-unit
```

**Run tests with verbose output using Go directly:**
```bash
go test -v ./...
```

**Run tests for a specific package:**
```bash
task test-package PKG=./internal/db
```

Or:
```bash
go test -v ./internal/db
```

#### Test Options

**Short mode** (skip long-running tests):
```bash
task test-short
```

**Race detector** (detect data races):
```bash
task test-race
```

**Coverage report** (generates HTML coverage report):
```bash
task test-coverage
```

This will:
1. Run all tests with coverage
2. Generate `coverage.out` file
3. Generate `coverage.html` (open in browser to view)

**Run tests with specific flags:**
```bash
# Run tests with count (disable caching)
go test -v -count=1 ./...

# Run a specific test
go test -v -run TestLoadDatabase ./internal/db

# Run tests with timeout
go test -v -timeout 30s ./...
```

#### Test Structure

Tests are located alongside the code they test:
- `internal/db/database_test.go` - Tests for database loading and operations
- `internal/db/testdb_test.go` - Tests for test database creation

All tests follow Go's standard testing conventions:
- Test functions are named `TestXxx` and take `*testing.T`
- Use `t.TempDir()` for temporary files (auto-cleaned)
- Test database files are created automatically per test

#### Creating Test Database Files

To manually create a test database file for inspection or debugging:

**Using Task:**
```bash
task create-testdb
```

**Using the shell script directly:**
```bash
./create-testdb.sh -o test.db
```

**Or using Go directly:**
```bash
go run ./cmd/create-testdb/main.go -o /path/to/test.db
```

The shell script (`create-testdb.sh`) is a wrapper that runs `go run` on the Go source file, so no binary compilation is needed.

**Quick test run:**
```bash
./test-run.sh
```

This script automatically:
1. Creates a test database if it doesn't exist
2. Shows database statistics
3. Runs multiple example searches with explanations

Each search command is displayed with an explanation of what it does, making it useful for learning how to use the CLI.

The test database contains:
- 5 folders: root (/), home, user, Documents, Downloads
- 5 files: test.txt, readme.txt, document.pdf, test.go, file.zip
- Proper parent-child relationships
- Index flags for name, size, and modification time

#### Example Test Output

```
=== RUN   TestLoadDatabase
--- PASS: TestLoadDatabase (0.00s)
=== RUN   TestEntryGetFullPath
--- PASS: TestEntryGetFullPath (0.00s)
=== RUN   TestSearch
--- PASS: TestSearch (0.00s)
=== RUN   TestMatchWholeWord
--- PASS: TestMatchWholeWord (0.00s)
=== RUN   TestCreateTestDatabase
--- PASS: TestCreateTestDatabase (0.00s)
=== RUN   TestCreateTestDatabaseMultipleTimes
--- PASS: TestCreateTestDatabaseMultipleTimes (0.00s)
PASS
ok  	github.com/gsearch-cli/internal/db	0.002s
```

### Code Quality

**Format code:**
```bash
task fmt
```

Or:
```bash
go fmt ./...
```

**Lint code:**
```bash
task lint
```

This runs:
- `go vet` - Reports suspicious constructs
- `golangci-lint` - Comprehensive linter (if installed)

## Project Structure

```
gsearch-cli/
├── cmd/
│   └── gsearch-cli/      # CLI application
│       └── main.go
├── internal/
│   └── db/               # Database access layer
│       ├── database.go   # Database loading and structure
│       ├── search.go     # Search functionality
│       └── database_test.go  # Unit tests
├── Taskfile.yml         # Task build configuration
├── go.mod               # Go module definition
├── README.md            # This file
└── FSEARCH_DB.md        # Database format documentation
```

## License

This project is provided as-is for use with FSearch databases.

