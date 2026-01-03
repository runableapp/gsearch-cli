#!/bin/bash
# Test runner script for gsearch-cli
# Creates test database if needed and runs search queries

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Function to colorize command with highlighted flags
colorize_command() {
    local cmd="$1"
    # Highlight flags (words starting with -)
    echo "$cmd" | sed -E "s/(-[a-zA-Z-]+)/${CYAN}${BOLD}\1${NC}/g" | \
         sed -E "s/(go run)/${MAGENTA}${BOLD}\1${NC}/g" | \
         sed -E "s/(\.\/cmd\/gsearch-cli\/main\.go)/${BLUE}\1${NC}/g" | \
         sed -E "s/(-db [^ ]+)/${CYAN}${BOLD}-db${NC} ${GREEN}\2${NC}/g" | \
         sed -E "s/(-q [^ ]+)/${CYAN}${BOLD}-q${NC} ${GREEN}\2${NC}/g" | \
         sed -E "s/(-path [^ ]+)/${CYAN}${BOLD}-path${NC} ${GREEN}\2${NC}/g" | \
         sed -E "s/(-max [^ ]+)/${CYAN}${BOLD}-max${NC} ${GREEN}\2${NC}/g"
}

# Function to colorize command with highlighted flags
colorize_cmd() {
    local cmd="$1"
    local result=""
    
    # Split command into words
    local words=($cmd)
    local i=0
    local len=${#words[@]}
    
    while [ $i -lt $len ]; do
        local word="${words[$i]}"
        local next_word="${words[$i+1]}"
        
        # Check if this is "go" followed by "run"
        if [[ "$word" == "go" ]] && [[ "$next_word" == "run" ]]; then
            result+="${MAGENTA}${BOLD}go run${NC} "
            ((i += 2))
            continue
        fi
        
        # Check if this is a flag with a value
        if [[ "$word" =~ ^- ]] && [[ -n "$next_word" ]] && [[ ! "$next_word" =~ ^- ]]; then
            # Flag with value
            result+="${CYAN}${BOLD}${word}${NC} ${GREEN}${next_word}${NC} "
            ((i += 2))
            continue
        fi
        
        # Standalone flag
        if [[ "$word" =~ ^- ]]; then
            result+="${CYAN}${BOLD}${word}${NC} "
            ((i++))
            continue
        fi
        
        # File paths
        if [[ "$word" =~ \.go$ ]] || [[ "$word" =~ \.sh$ ]]; then
            result+="${BLUE}${word}${NC} "
            ((i++))
            continue
        fi
        
        # Regular word
        result+="${word} "
        ((i++))
    done
    
    # Remove trailing space
    echo -n "${result% }"
}

# Configuration
TEST_DB="test.db"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${BLUE}=== gsearch-cli Test Runner ===${NC}\n"

# Step 1: Create test database if it doesn't exist
if [ ! -f "$TEST_DB" ]; then
    echo -e "${YELLOW}Step 1: Creating test database...${NC}"
    echo -e "Command: $(colorize_cmd "./create-testdb.sh -o $TEST_DB")"
    echo "Explanation: Creating a test FSearch database file with sample data (5 folders, 5 files)"
    echo ""
    ./create-testdb.sh -o "$TEST_DB"
    echo -e "${GREEN}✓ Test database created: $TEST_DB${NC}\n"
else
    echo -e "${GREEN}✓ Test database already exists: $TEST_DB${NC}\n"
fi

# Step 2: Show database statistics
echo -e "${YELLOW}Step 2: Displaying database statistics...${NC}"
echo -e "Command: $(colorize_cmd "go run ./cmd/gsearch-cli/main.go -db $TEST_DB -stats")"
echo "Explanation: Shows database metadata including number of folders, files, and index flags"
echo ""
go run "$SCRIPT_DIR/cmd/gsearch-cli/main.go" -db "$TEST_DB" -stats
echo ""

# Step 3: Run search queries
echo -e "${YELLOW}Step 3: Running search queries...${NC}\n"

# Search 1: Simple name search
echo -e "${BLUE}--- Search 1: Find files/folders containing 'test' ---${NC}"
echo -e "Command: $(colorize_cmd "go run ./cmd/gsearch-cli/main.go -db $TEST_DB -q test")"
echo "Explanation: Searches for entries with 'test' in the name (case-insensitive by default)"
echo ""
go run "$SCRIPT_DIR/cmd/gsearch-cli/main.go" -db "$TEST_DB" -q test
echo ""

# Search 2: Files only
echo -e "${BLUE}--- Search 2: Find files only containing 'test' ---${NC}"
echo -e "Command: $(colorize_cmd "go run ./cmd/gsearch-cli/main.go -db $TEST_DB -q test -files")"
echo "Explanation: Same search but filters to show only files, excluding folders"
echo ""
go run "$SCRIPT_DIR/cmd/gsearch-cli/main.go" -db "$TEST_DB" -q test -files
echo ""

# Search 3: Folders only
echo -e "${BLUE}--- Search 3: Find folders containing 'doc' ---${NC}"
echo -e "Command: $(colorize_cmd "go run ./cmd/gsearch-cli/main.go -db $TEST_DB -q doc -folders")"
echo "Explanation: Searches for folders with 'doc' in the name, excluding files"
echo ""
go run "$SCRIPT_DIR/cmd/gsearch-cli/main.go" -db "$TEST_DB" -q doc -folders
echo ""

# Search 4: Path search
echo -e "${BLUE}--- Search 4: Search in full path for 'home' ---${NC}"
echo -e "Command: $(colorize_cmd "go run ./cmd/gsearch-cli/main.go -db $TEST_DB -path home")"
echo "Explanation: Searches in the full path (not just name) for entries containing 'home'"
echo ""
go run "$SCRIPT_DIR/cmd/gsearch-cli/main.go" -db "$TEST_DB" -path home
echo ""

# Search 5: Case-sensitive search
echo -e "${BLUE}--- Search 5: Case-sensitive search for 'Test' ---${NC}"
echo -e "Command: $(colorize_cmd "go run ./cmd/gsearch-cli/main.go -db $TEST_DB -q Test -case")"
echo "Explanation: Case-sensitive search that will only match 'Test' (capital T), not 'test'"
echo ""
go run "$SCRIPT_DIR/cmd/gsearch-cli/main.go" -db "$TEST_DB" -q Test -case
echo ""

# Search 6: Limit results
echo -e "${BLUE}--- Search 6: Limit results to 1 ---${NC}"
echo -e "Command: $(colorize_cmd "go run ./cmd/gsearch-cli/main.go -db $TEST_DB -q test -max 1")"
echo "Explanation: Limits the search results to maximum 1 result"
echo ""
go run "$SCRIPT_DIR/cmd/gsearch-cli/main.go" -db "$TEST_DB" -q test -max 1
echo ""

echo -e "${GREEN}=== Test run completed ===${NC}"

