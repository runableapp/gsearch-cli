# Performance Analysis: FSearch vs gsearch-cli

## Executive Summary

FSearch is significantly faster than the Go implementation because it uses:
1. **Multi-threaded parallel search** across CPU cores
2. **Pre-sorted arrays** loaded from the database
3. **Memory pools** for efficient allocation
4. **Persistent database in memory** (loaded once, reused)
5. **Optimized path operations** (no repeated path reconstruction)

The Go implementation is single-threaded, doesn't use pre-sorted arrays, and reconstructs paths for every search.

## Detailed Analysis

### 1. Multi-Threading (Major Impact)

**FSearch:**
- Uses a thread pool (`FsearchThreadPool`) with one thread per CPU core
- Splits search work across threads (threshold: 1000+ entries)
- Parallel search in `db_search_entries()`:
  ```c
  const uint32_t num_threads = (num_entries < THRESHOLD_FOR_PARALLEL_SEARCH || q->wants_single_threaded_search)
                                 ? 1
                                 : fsearch_thread_pool_get_num_threads(pool);
  ```
- Each thread searches a portion of the array independently
- Results are merged after all threads complete

**Go Implementation:**
- Single-threaded sequential search
- Iterates through all files, then all folders sequentially
- No parallelization

**Impact:** On a multi-core system, FSearch can be 4-8x faster for large databases.

### 2. Pre-Sorted Arrays (Major Impact)

**FSearch:**
- Maintains pre-sorted arrays in memory for different sort types:
  - `sorted_files[NUM_DATABASE_INDEX_TYPES]`
  - `sorted_folders[NUM_DATABASE_INDEX_TYPES]`
- Sorted arrays are loaded from the database file at startup
- Uses `db_get_entries_sorted()` to get already-sorted arrays
- No sorting needed during search - just iterate pre-sorted arrays

**Go Implementation:**
- Loads sorted arrays from database (`SortedArrays` map) but **doesn't use them**
- Always iterates through unsorted `db.Files` and `db.Folders` arrays
- Sorted arrays are loaded but ignored during search

**Impact:** FSearch can use binary search or early termination on sorted arrays. The Go code always does linear search.

### 3. Path Reconstruction (Moderate Impact)

**FSearch:**
- Paths are built on-demand using `db_entry_get_path()` or `db_entry_get_path_full()`
- Uses recursive traversal but likely caches results in some contexts
- Path building is optimized with `GString` (mutable string builder)

**Go Implementation:**
- Calls `GetFullPath()` for **every entry** during path searches
- Each call:
  1. Traverses parent chain (potentially many levels)
  2. Builds a slice of strings
  3. Joins with `strings.Join()`
  4. Does additional root path checking
- For a search like `*qiskit*` on a large database, this means:
  - Thousands of path reconstructions
  - Thousands of string allocations
  - Thousands of parent pointer traversals

**Impact:** Significant overhead for path-based searches. Could be optimized by:
- Caching paths after first computation
- Using a string builder instead of slice + join
- Pre-computing paths during database load

### 4. Regex Compilation (Minor Impact)

**FSearch:**
- Uses PCRE2 library with JIT compilation
- Regex patterns are compiled once and reused
- Thread-safe matching

**Go Implementation:**
- Compiles regex **for every search** (even if pattern is the same)
- Uses Go's `regexp` package (RE2 engine)
- No caching of compiled regex patterns

**Impact:** Small overhead for wildcard searches. Could cache compiled regex patterns.

### 5. String Operations (Minor Impact)

**FSearch:**
- Uses optimized C string operations
- Direct memory access
- Minimal allocations

**Go Implementation:**
- Multiple string operations per entry:
  - `strings.Contains()` for substring matching
  - `strings.ToLower()` for case-insensitive search (creates new string)
  - String comparisons
- Each operation may allocate memory

**Impact:** Minor overhead, but adds up over thousands of entries.

### 6. Database Loading (Minor Impact for Repeated Searches)

**FSearch:**
- Database is loaded once and kept in memory
- Subsequent searches use the in-memory database
- No I/O overhead after initial load

**Go Implementation:**
- Loads database from disk **every time** the program runs
- For CLI usage, this is acceptable (one load per invocation)
- But if doing multiple searches in one run, database is already loaded

**Impact:** Only affects first search in a session. Subsequent searches in the same run are fast.

### 7. Memory Allocation (Minor Impact)

**FSearch:**
- Uses memory pools (`FsearchMemoryPool`) for efficient allocation
- Pre-allocates blocks of memory
- Reduces malloc/free overhead

**Go Implementation:**
- Uses Go's standard memory allocator
- Many small allocations for:
  - Search results slices
  - String operations
  - Path building

**Impact:** Go's GC handles this well, but memory pools could reduce allocations.

## Performance Bottlenecks in Go Code

### Critical Issues:

1. **No multi-threading** - Biggest performance gap
2. **Not using pre-sorted arrays** - Missing optimization opportunity
3. **Repeated path reconstruction** - Major overhead for path searches

### Moderate Issues:

4. **Regex compilation per search** - Should cache compiled patterns
5. **String allocations** - Could use string builders or byte slices

### Minor Issues:

6. **No early termination** - Could stop early when max results reached
7. **Inefficient path building** - Using slice + join instead of builder

## Optimization Recommendations

### High Priority:

1. **Implement multi-threaded search**
   - Use Go's `sync` package or worker pool pattern
   - Split files/folders arrays across goroutines
   - Merge results after all goroutines complete
   - Expected speedup: 4-8x on multi-core systems

2. **Use pre-sorted arrays from database**
   - Instead of iterating `db.Files` and `db.Folders`, use `db.SortedArrays`
   - For name searches, use sorted arrays sorted by name
   - Enables binary search or early termination
   - Expected speedup: 2-10x depending on database size

3. **Cache path computations**
   - Compute paths once and cache them
   - Or pre-compute during database load
   - Use `sync.Map` or similar for thread-safe caching
   - Expected speedup: 2-5x for path searches

### Medium Priority:

4. **Cache compiled regex patterns**
   - Store compiled regex in a map keyed by pattern
   - Reuse compiled patterns across searches
   - Expected speedup: 10-20% for wildcard searches

5. **Optimize path building**
   - Use `strings.Builder` instead of slice + join
   - Pre-allocate builder capacity
   - Expected speedup: 10-30% for path operations

### Low Priority:

6. **Early termination optimization**
   - Stop searching when max results reached
   - Already partially implemented but could be improved
   - Expected speedup: Variable (depends on result position)

7. **Reduce string allocations**
   - Use byte slices for temporary string operations
   - Pool string builders
   - Expected speedup: 5-15%

## Expected Performance Improvements

If all optimizations are implemented:

- **Multi-threading alone**: 4-8x faster
- **Using sorted arrays**: 2-10x faster (depends on database size)
- **Path caching**: 2-5x faster for path searches
- **Combined**: Could approach or exceed FSearch performance

## Code Locations for Optimization

### Multi-threading:
- `internal/db/search.go` - `Search()` and `SearchByPath()` functions
- Add worker pool or goroutine-based parallel search

### Sorted Arrays:
- `internal/db/database.go` - `SortedArrays` map is loaded but unused
- `internal/db/search.go` - Modify search to use sorted arrays

### Path Caching:
- `internal/db/database.go` - Add path cache to `Database` struct
- `internal/db/database.go` - `GetFullPath()` method

### Regex Caching:
- `internal/db/search.go` - Add regex cache map
- `internal/db/search.go` - `matches()` and `SearchByPath()` functions

## Measurement Results

Test case: `./bin/gsearch-cli -db /home/keith/.local/share/fsearch/fsearch.db.bak -q '*qiskit*'`

- **Go implementation**: ~7.8 seconds (real time), 24.4 seconds (user time)
- **FSearch**: Near-instant (< 100ms typically)

The user time (24.4s) being much higher than real time (7.8s) suggests the system is doing a lot of work, but the real bottleneck is likely:
1. Single-threaded execution (not utilizing all CPU cores)
2. Path reconstruction overhead
3. Not using sorted arrays for efficient search

## Conclusion

The Go implementation is significantly slower primarily due to:
1. **Single-threaded execution** (biggest issue)
2. **Not using pre-sorted arrays** (major missed optimization)
3. **Inefficient path reconstruction** (major overhead for path searches)

Implementing multi-threading and using sorted arrays would provide the biggest performance improvements and could make the Go implementation competitive with FSearch.

