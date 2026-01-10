# FSearch Database File Format

This document describes the binary file format used by FSearch to store its file index database.

## File Location

The database file is stored at: `~/.local/share/fsearch/fsearch.db`

## File Format Overview

The database file is a binary format with the following structure:

### Header

```
Offset  Size  Type     Description
------  ----  ----     -----------
0       4     char[]   Magic number: "FSDB"
4       1     uint8    Major version (currently 0)
5       1     uint8    Minor version (currently 9)
```

### Database Metadata

```
Offset  Size  Type     Description
------  ----  ----     -----------
6       8     uint64   Index flags (bitmask indicating which fields are indexed)
14      4     uint32   Number of folders
18      4     uint32   Number of files
22      8     uint64   Folder block size (in bytes)
30      8     uint64   File block size (in bytes)
38      4     uint32   Number of indexes (currently unused, always 0)
42      4     uint32   Number of excludes (currently unused, always 0)
```

### Index Flags

The index flags bitmask indicates which metadata fields are stored for each entry:

- `DATABASE_INDEX_FLAG_NAME` (1 << 0): Name is always stored
- `DATABASE_INDEX_FLAG_SIZE` (1 << 2): File/folder size
- `DATABASE_INDEX_FLAG_MODIFICATION_TIME` (1 << 3): Modification time (mtime)

Currently, the database always stores:
- Name (always present)
- Size (if flag is set)
- Modification time (if flag is set)

### Folder Block

The folder block contains all folder entries, stored sequentially. Each folder entry has the following structure:

```
Size  Type     Description
----  ----     -----------
2     uint16   Database index (currently unused, always 0)
1     uint8    Name offset (character position where name differs from previous entry)
1     uint8    Name length (length of new name characters)
N     char[]   Name characters (only the differing part from previous entry)
8     int64    Size (if DATABASE_INDEX_FLAG_SIZE is set)
8     int64    Modification time (if DATABASE_INDEX_FLAG_MODIFICATION_TIME is set)
4     uint32   Parent folder index
```

**Name Compression**: To save space, folder/file names are stored using delta compression:
- `name_offset`: The character position where the current name differs from the previous entry's name
- `name_len`: The number of new characters to append
- `name`: Only the new characters (not the full name)

The full name is reconstructed by:
1. Taking the previous entry's name
2. Truncating it at `name_offset`
3. Appending the new `name` characters

### File Block

The file block contains all file entries, stored sequentially. Each file entry has the same structure as a folder entry, except:
- No `db_index` field (folders only)
- Files always have a parent folder

```
Size  Type     Description
----  ----     -----------
1     uint8    Name offset
1     uint8    Name length
N     char[]   Name characters
8     int64    Size (if DATABASE_INDEX_FLAG_SIZE is set)
8     int64    Modification time (if DATABASE_INDEX_FLAG_MODIFICATION_TIME is set)
4     uint32   Parent folder index
```

### Sorted Arrays

After the file block, the database stores pre-sorted index arrays for efficient searching. The format is:

```
Size  Type     Description
----  ----     -----------
4     uint32   Number of sorted arrays
```

For each sorted array:

```
Size  Type     Description
----  ----     -----------
4     uint32   Array ID (sort type: 1=Path, 2=Size, 3=ModificationTime, 4=Extension, etc.)
4*N   uint32[] Folder indices (N = number of folders)
4*M   uint32[] File indices (M = number of files)
```

The sorted arrays contain indices into the original folder/file arrays, sorted by different criteria:
- `DATABASE_INDEX_TYPE_NAME` (0): Original order (not stored as sorted array)
- `DATABASE_INDEX_TYPE_PATH` (1): Sorted by full path
- `DATABASE_INDEX_TYPE_SIZE` (2): Sorted by size
- `DATABASE_INDEX_TYPE_MODIFICATION_TIME` (3): Sorted by modification time
- `DATABASE_INDEX_TYPE_EXTENSION` (4): Sorted by file extension

## Data Structures

### Entry Structure

Each entry (file or folder) contains:
- `name`: Full name of the file/folder
- `size`: Size in bytes (for folders, this is the sum of all children)
- `mtime`: Modification time (Unix timestamp)
- `parent`: Pointer/index to parent folder
- `idx`: Index in the original name-sorted array
- `type`: Entry type (FILE or FOLDER)

### Folder-Specific Fields

Folders additionally contain:
- `db_idx`: Database index (currently unused)
- `num_files`: Number of files in this folder
- `num_folders`: Number of subfolders

## Reading the Database

### Loading Process

1. **Read Header**: Verify magic number and version
2. **Read Metadata**: Get counts and block sizes
3. **Pre-allocate Folders**: Create folder entries with indices
4. **Load Folders**: Read folder block and reconstruct names using delta compression
5. **Load Files**: Read file block and reconstruct names
6. **Load Sorted Arrays**: Read pre-sorted indices for efficient searching

### Name Reconstruction Example

If entries are stored as:
- Entry 1: offset=0, len=5, name="home"
- Entry 2: offset=4, len=4, name="user"
- Entry 3: offset=0, len=3, name="etc"

The reconstructed names are:
- Entry 1: "home"
- Entry 2: "home"[0:4] + "user" = "homeuser" (wait, this seems wrong...)

Actually, looking at the code more carefully:
- The offset is where the names start to differ
- Previous name is truncated at offset, then new characters are appended

So if:
- Previous: "home/user"
- Current: offset=5, len=4, name="docs"
- Result: "home/user"[0:5] + "docs" = "home/docs"

## Notes

- The database uses little-endian byte order
- All integers are stored in network byte order (little-endian on x86/x64)
- The database file is locked during writes using `flock()`
- Temporary files (`.tmp`) are used during saves for atomic updates
- Parent indices reference the folder's index in the name-sorted array
- Root folders have `parent_idx == idx` (self-reference indicates no parent)

