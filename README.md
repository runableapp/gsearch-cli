# gsearch-cli

A command-line interface for searching the FSearch database. This Go application reads and searches the binary database file created by [FSearch](https://github.com/cboxdoerfer/fsearch).

## Features

- Fast search through indexed files and folders
- Case-sensitive and case-insensitive search
- Whole word matching
- Search by name or full path
- Wildcard pattern matching (`*` and `?`)
- Filter by files or folders only
- Multiple output formats (text, JSON, CSV)
- Sort results by name, path, size, or modification time
- Display database statistics
- Human-readable file sizes
- Comprehensive help documentation

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
  - Supports wildcard patterns: `*` (any sequence) and `?` (single character)
  - Examples: `*.txt`, `test*`, `file?.go`
- `-path <pattern>`: Search in full path instead of just name
  - Also supports wildcard patterns
  - Examples: `/home/*`, `*/Documents/*`
- `-case`: Enable case-sensitive search (default: false)
- `-whole`: Match whole words only (default: false)
- `-files`: Search only files
- `-folders`: Search only folders
- `-max <n>`: Maximum number of results (0 = unlimited)
- `-db <path>`: Path to database file (default: `~/.local/share/fsearch/fsearch.db`)
- `-stats`: Show database statistics

### Output Options

- `-output <format>`: Output format (default: `text`)
  - `text`: Human-readable format with folder/file indicators
  - `json`: JSON array with structured fields (name, path, type, size, mtime)
  - `csv`: CSV format with header row, suitable for spreadsheet import
- `-sort <field>`: Sort results by field (default: no sorting)
  - `name`: Sort by file/folder name (alphabetical)
  - `path`: Sort by full path (alphabetical)
  - `size`: Sort by file size (ascending), folders sorted by name
  - `mtime`: Sort by modification time (oldest first)

### Help

- `-h`, `-help`: Show detailed help message with all options and examples

### Examples

**Basic search:**
```bash
gsearch-cli -q test
```

**Search for files containing "test":**
```bash
gsearch-cli -q test -files
```

**Wildcard patterns:**
```bash
# Find all .txt files
gsearch-cli -q "*.txt"

# Find files starting with "test"
gsearch-cli -q "test*"

# Find files with single character before .go
gsearch-cli -q "?.go"

# Find files with "test" anywhere in name
gsearch-cli -q "*test*"
```

**Case-sensitive search:**
```bash
gsearch-cli -q Test -case
```

**Search in full path:**
```bash
gsearch-cli -path /home/user
```

**Wildcard path search:**
```bash
# Find all files in /home directory
gsearch-cli -path "/home/*"

# Find all .txt files in any path
gsearch-cli -path "*.txt"
```

**Show database statistics:**
```bash
gsearch-cli -stats
```

**Limit results:**
```bash
gsearch-cli -q test -max 10
```

**Output formats:**
```bash
# JSON output
gsearch-cli -q test -output json

# CSV output
gsearch-cli -q test -output csv

# Text output (default)
gsearch-cli -q test -output text
```

**Sort results:**
```bash
# Sort by name
gsearch-cli -q test -sort name

# Sort by size
gsearch-cli -q test -sort size

# Sort by modification time
gsearch-cli -q test -sort mtime

# Sort by path
gsearch-cli -q test -sort path
```

**Combine options:**
```bash
# Wildcard search, files only, sorted by size, JSON output
gsearch-cli -q "*.go" -files -sort size -output json

# Path search, sorted by modification time, CSV output
gsearch-cli -path "/home/*" -sort mtime -output csv
```

**Show help:**
```bash
gsearch-cli -help
# or
gsearch-cli -h
```

### Wildcard Patterns

gsearch-cli supports wildcard patterns for flexible searching:

- `*` - Matches any sequence of characters (zero or more)
- `?` - Matches a single character

**Wildcard Examples:**

```bash
# Find all text files
gsearch-cli -q "*.txt"

# Find all files starting with "test"
gsearch-cli -q "test*"

# Find files with exactly one character before .go
gsearch-cli -q "?.go"

# Find files containing "test" anywhere
gsearch-cli -q "*test*"

# Wildcard in path search
gsearch-cli -path "/home/*"
gsearch-cli -path "*.txt"  # All .txt files in any path
```

**Note:** Wildcard patterns are automatically detected when `*` or `?` characters are present in the query. Special regex characters (`.`, `^`, `$`, etc.) are automatically escaped, so you can use them literally in your patterns.

## Output Formats

### Text Format (Default)

Human-readable format with folder/file indicators:
```
Found 2 result(s):
📁 /Documents
📄 /home/user/test.txt (1.0 KB)
```

### JSON Format

Structured JSON array with all available fields:
```json
[
  {
    "name": "test.txt",
    "path": "/home/user/test.txt",
    "type": "file",
    "size": 1024,
    "mtime": "2024-01-03T12:00:00Z",
    "mtime_ts": 1704283200
  },
  {
    "name": "Documents",
    "path": "/Documents",
    "type": "folder",
    "mtime": "2024-01-03T12:00:00Z",
    "mtime_ts": 1704283200
  }
]
```

Fields:
- `name`: File or folder name
- `path`: Full path
- `type`: `"file"` or `"folder"`
- `size`: File size in bytes (only for files)
- `mtime`: Modification time in RFC3339 format
- `mtime_ts`: Modification time as Unix timestamp

### CSV Format

CSV format with header row, suitable for spreadsheet import:
```csv
name,path,type,size,mtime
test.txt,/home/user/test.txt,file,1024,2024-01-03T12:00:00Z
Documents,/Documents,folder,,2024-01-03T12:00:00Z
```

Note: Folder entries have an empty `size` field.

## Sorting

Results can be sorted by any of the following fields:

- **name**: Alphabetical sorting by file/folder name
- **path**: Alphabetical sorting by full path
- **size**: By file size (ascending). Folders are sorted by name when sorting by size
- **mtime**: By modification time (oldest first)

Sorting applies to both files and folders together. When sorting by size, folders (which don't have a size) are sorted by name instead.

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
- `internal/db/search_wildcard_test.go` - Tests for wildcard pattern matching
- `cmd/gsearch-cli/output_test.go` - Tests for output formats and sorting

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
├── LICENSE              # GNU General Public License v3.0
├── README.md            # This file
└── FSEARCH_DB.md        # Database format documentation
```

## License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0).

See the [LICENSE](LICENSE) file for the full license text.

### What this means:

- **You are free to use, modify, and distribute** this software
- **You must disclose source code** when distributing modified versions
- **You must license derivative works** under the same GPL-3.0 license
- **You must preserve copyright notices** and license information

For more information about the GPL-3.0 license, visit: https://www.gnu.org/licenses/gpl-3.0.html

