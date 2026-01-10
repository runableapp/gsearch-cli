# Path Reconstruction: Visual Example

## Scenario

Database with 5 files in the same directory:
- `/home/user/docs/file1.txt`
- `/home/user/docs/file2.txt`
- `/home/user/docs/file3.txt`
- `/home/user/docs/file4.txt`
- `/home/user/docs/file5.txt`

Search query: `*file3*` (matches only `file3.txt`)

## What Happens in Current Implementation

### Step 1: Search file1.txt

```
1. Call GetFullPath() for file1.txt
   ├─ Traverse parent chain:
   │  ├─ file1.txt (depth 0)
   │  ├─ docs (depth 1) ← parent
   │  ├─ user (depth 2) ← parent.parent
   │  ├─ home (depth 3) ← parent.parent.parent
   │  └─ / (depth 4) ← parent.parent.parent.parent
   │
   ├─ Build slice: ["file1.txt", "docs", "user", "home", "/"]
   ├─ Join: "/home/user/docs/file1.txt"
   ├─ Traverse parent chain AGAIN to check for root
   └─ Return: "/home/user/docs/file1.txt"

2. Check if matches "*file3*": NO
```

**Operations:**
- 2 parent chain traversals (8 pointer dereferences)
- 5 slice allocations
- 1 string join
- 1 regex match

### Step 2: Search file2.txt

```
1. Call GetFullPath() for file2.txt
   ├─ Traverse parent chain:
   │  ├─ file2.txt (depth 0)
   │  ├─ docs (depth 1) ← SAME PARENT as file1!
   │  ├─ user (depth 2) ← SAME PARENT as file1!
   │  ├─ home (depth 3) ← SAME PARENT as file1!
   │  └─ / (depth 4) ← SAME PARENT as file1!
   │
   ├─ Build slice: ["file2.txt", "docs", "user", "home", "/"]
   ├─ Join: "/home/user/docs/file2.txt"
   ├─ Traverse parent chain AGAIN
   └─ Return: "/home/user/docs/file2.txt"

2. Check if matches "*file3*": NO
```

**Operations:**
- 2 parent chain traversals (8 pointer dereferences) ← **REDUNDANT!**
- 5 slice allocations ← **REDUNDANT!**
- 1 string join ← **REDUNDANT!**
- 1 regex match

**We just traversed the same parent chain (`docs` → `user` → `home` → `/`) that we traversed for file1!**

### Step 3: Search file3.txt

```
1. Call GetFullPath() for file3.txt
   ├─ Traverse parent chain: (SAME AS ABOVE!)
   ├─ Build slice: ["file3.txt", "docs", "user", "home", "/"]
   ├─ Join: "/home/user/docs/file3.txt"
   └─ Return: "/home/user/docs/file3.txt"

2. Check if matches "*file3*": YES ✓
3. Add to results
```

### Steps 4-5: file4.txt and file5.txt

Same redundant work as file1 and file2.

## Total Work Done

For 5 files in the same directory:

**Without caching:**
- 5 × 2 = **10 parent chain traversals**
- 5 × 5 = **25 slice allocations**
- 5 × 1 = **5 string joins**
- 5 × 1 = **5 regex matches**

**With caching (parent paths):**
- 1 × 2 = **2 parent chain traversals** (for first file)
- 4 × 1 = **4 parent chain traversals** (only for file name)
- 1 × 5 = **5 slice allocations** (for first file)
- 4 × 1 = **4 slice allocations** (only for file name)
- 1 × 1 = **1 string join** (for first file)
- 4 × 1 = **4 string joins** (only for file name)
- 5 × 1 = **5 regex matches** (still needed)

**Savings:**
- 8 fewer parent chain traversals (80% reduction)
- 16 fewer slice allocations (64% reduction)
- 4 fewer string joins (80% reduction)

## Real-World Scale

For a database with 100,000 files:

**Without caching:**
- 100,000 × 2 = **200,000 parent chain traversals**
- 100,000 × 5 = **500,000 slice allocations** (assuming depth 5)
- 100,000 × 1 = **100,000 string joins**

**With caching:**
- Depends on directory structure, but typically:
- ~10,000-20,000 parent chain traversals (90% reduction)
- ~100,000-200,000 slice allocations (60-80% reduction)
- ~100,000 string joins (no reduction, but faster due to smaller slices)

## Memory Impact

**Without caching:**
- Each path reconstruction creates temporary objects
- GC pressure from millions of allocations
- Memory fragmentation

**With caching:**
- One-time cost: ~50-100 bytes per path
- For 100,000 files: ~5-10 MB additional memory
- **Worth it for the speedup!**

## The Key Insight

**The same parent directories are traversed over and over again.**

If we cache the path for `/home/user/docs`, then:
- file1.txt: Get cached path for `docs` + "file1.txt" = fast
- file2.txt: Get cached path for `docs` + "file2.txt" = fast
- file3.txt: Get cached path for `docs` + "file3.txt" = fast

Instead of:
- file1.txt: Traverse docs → user → home → /, build path = slow
- file2.txt: Traverse docs → user → home → /, build path = slow (same work!)
- file3.txt: Traverse docs → user → home → /, build path = slow (same work!)

## Implementation Sketch

```go
type Database struct {
    // ... existing fields ...
    pathCache sync.Map  // map[*Entry]string
}

func (e *Entry) GetFullPath() string {
    // Check if this entry's path is cached
    if cached, ok := db.pathCache.Load(e); ok {
        return cached.(string)
    }
    
    // Check if parent's path is cached
    var parentPath string
    if e.Parent != nil {
        if cached, ok := db.pathCache.Load(e.Parent); ok {
            parentPath = cached.(string)
        } else {
            // Recursively compute parent's path (will cache it)
            parentPath = e.Parent.GetFullPath()
        }
    }
    
    // Build this entry's path from parent
    var fullPath string
    if parentPath == "/" {
        fullPath = "/" + e.Name
    } else {
        fullPath = parentPath + "/" + e.Name
    }
    
    // Cache it
    db.pathCache.Store(e, fullPath)
    return fullPath
}
```

**Result:** Each directory path is computed once, then reused for all children.

