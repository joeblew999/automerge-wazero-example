# Folder Restructuring Options - Separating Infrastructure from 1:1 Mapping

**Date Created**: 2025-10-21
**Status**: ✅ **IMPLEMENTED** - Option 3 (crdt_ naming convention)

**Goal**: Physically separate infrastructure files from CRDT operation files to make the 1:1 mapping crystal clear.

**Solution**: Renamed all CRDT operation files with `crdt_` prefix (49 files total - 33 Go + 16 web).

---

## ⚠️ LESSON LEARNED: Go Package System Constraints

**We attempted Option 1 (subfolder approach)** and discovered a critical issue:

**In Go, subdirectories MUST be separate packages.** You cannot have:
```
pkg/automerge/
├── types.go         (package automerge)
└── crdt/
    └── text.go      (package automerge) ❌ DOESN'T WORK
```

Files in `crdt/` subdirectory become `package crdt`, which means:
- They can't access `Document`, `Path` types from parent `package automerge`
- Would need to import parent: `import "pkg/automerge"` → creates circular dependency
- Breaks the entire architecture

**Conclusion**: Option 1 as originally described is **impossible in Go**.

**Alternatives that DO work**:
- Option 3: Naming convention (`crdt_text.go`) ✅ Works immediately
- Option 2: Separate packages (`automerge_crdt`, `automerge_types`) ✅ Works but requires major refactor
- Option 4: Status quo + documentation ✅ Already done

---

---

## Option 1: **Subfolder Approach** (Cleanest) ⭐ RECOMMENDED

Create `crdt/` subfolders to hold only 1:1 mapped files:

### Proposed Structure

```
go/pkg/
├── wazero/
│   ├── crdt/              # NEW - Pure 1:1 CRDT mappings
│   │   ├── text.go
│   │   ├── map.go
│   │   ├── list.go
│   │   ├── counter.go
│   │   ├── cursor.go
│   │   ├── history.go
│   │   ├── sync.go
│   │   ├── richtext.go
│   │   └── generic.go
│   ├── state.go           # Infrastructure (stays at root)
│   ├── memory.go          # Infrastructure
│   └── document.go        # Infrastructure
│
├── automerge/
│   ├── crdt/              # NEW - Pure 1:1 CRDT mappings
│   │   ├── text.go
│   │   ├── map.go
│   │   ├── list.go
│   │   ├── counter.go
│   │   ├── cursor.go
│   │   ├── history.go
│   │   ├── sync.go
│   │   ├── richtext.go
│   │   └── generic.go
│   ├── document.go        # Infrastructure
│   ├── errors.go          # Infrastructure
│   └── types.go           # Infrastructure
│
├── server/
│   ├── crdt/              # NEW - Pure 1:1 CRDT mappings
│   │   ├── text.go
│   │   ├── map.go
│   │   ├── list.go
│   │   ├── counter.go
│   │   ├── cursor.go
│   │   ├── history.go
│   │   ├── sync.go
│   │   ├── richtext.go
│   │   └── generic.go
│   ├── server.go          # Infrastructure (stays at root)
│   ├── broadcast.go       # Infrastructure
│   └── document.go        # Infrastructure
│
└── api/
    ├── crdt/              # NEW - Pure 1:1 CRDT mappings
    │   ├── text.go        # (move from handlers.go)
    │   ├── map.go
    │   ├── list.go
    │   ├── counter.go
    │   ├── cursor.go
    │   ├── history.go
    │   ├── sync.go
    │   └── richtext.go
    ├── handlers.go        # Infrastructure (deprecate eventually)
    ├── util.go            # Infrastructure
    └── static.go          # Infrastructure
```

### Benefits
- ✅ **Crystal clear**: `crdt/` = 1:1 mapping, root = infrastructure
- ✅ **grep-able**: `ls */crdt/` shows all 1:1 files instantly
- ✅ **Self-documenting**: Folder structure enforces architectural rule
- ✅ **Scalable**: Easy to add more infrastructure without confusion

### Downsides
- ⚠️ Moderate refactor (~30 files to move)
- ⚠️ Import paths change (automerge.Text → automerge/crdt.Text)
- ⚠️ Need to update all imports across codebase

### Migration Effort
**Time**: 2-3 hours
**Risk**: Medium (need to update imports, run tests)

---

## Option 2: **Separate Packages** (Go Idiomatic) ⭐⭐

Create separate packages for infrastructure vs CRDT:

### Proposed Structure

```
go/pkg/
├── wazero/                # Rename to wazero_ffi
│   ├── text.go
│   ├── map.go
│   └── ...                # All 1:1 CRDT wrappers
│
├── wazero_runtime/        # NEW - Infrastructure only
│   ├── runtime.go         # (was wazero/state.go)
│   ├── memory.go
│   └── module.go
│
├── automerge/             # Rename to automerge_crdt
│   ├── text.go
│   ├── map.go
│   └── ...                # All 1:1 CRDT operations
│
├── automerge_types/       # NEW - Infrastructure only
│   ├── document.go
│   ├── errors.go
│   ├── path.go
│   └── types.go
│
├── server/                # Rename to server_crdt
│   ├── text.go
│   ├── map.go
│   └── ...                # All 1:1 CRDT with state
│
├── server_core/           # NEW - Infrastructure only
│   ├── server.go
│   ├── broadcast.go
│   ├── lifecycle.go
│   └── storage.go
│
├── api/                   # Rename to api_crdt
│   ├── text.go
│   ├── map.go
│   └── ...                # All 1:1 HTTP handlers
│
└── api_infra/             # NEW - Infrastructure only
    ├── routing.go
    ├── middleware.go
    ├── static.go
    └── util.go
```

### Benefits
- ✅ **Go best practice**: Separate packages by concern
- ✅ **Type safety**: Can't accidentally mix infrastructure with CRDT
- ✅ **Import clarity**: `import "pkg/automerge_crdt"` vs `import "pkg/automerge_types"`

### Downsides
- ⚠️ Major refactor (~50 files)
- ⚠️ Many new package names to learn
- ⚠️ More verbose imports

### Migration Effort
**Time**: 4-6 hours
**Risk**: High (major restructure, lots of testing)

---

## Option 3: **Naming Convention** (Minimal Change) ⭐⭐⭐

Keep current structure, use prefixes to distinguish:

### Proposed Structure

```
go/pkg/
├── wazero/
│   ├── crdt_text.go       # Rename text.go
│   ├── crdt_map.go        # Rename map.go
│   ├── crdt_list.go       # ...
│   ├── crdt_counter.go
│   ├── crdt_cursor.go
│   ├── crdt_history.go
│   ├── crdt_sync.go
│   ├── crdt_richtext.go
│   ├── crdt_generic.go
│   ├── state.go           # Infrastructure (no prefix)
│   ├── memory.go          # Infrastructure
│   └── document.go        # Infrastructure
│
├── automerge/
│   ├── crdt_text.go       # Rename text.go
│   ├── crdt_map.go        # ...
│   ├── document.go        # Infrastructure (no prefix)
│   ├── errors.go
│   └── types.go
│
├── server/
│   ├── crdt_text.go       # Rename text.go
│   ├── crdt_map.go        # ...
│   ├── server.go          # Infrastructure (no prefix)
│   ├── broadcast.go
│   └── document.go
│
└── api/
    ├── crdt_text.go       # New (from handlers.go)
    ├── crdt_map.go
    ├── handlers.go        # Infrastructure (no prefix)
    ├── util.go
    └── static.go
```

### Benefits
- ✅ **Minimal change**: Just rename files
- ✅ **grep-able**: `ls **/crdt_*.go` shows all 1:1 files
- ✅ **No import changes**: Packages stay the same
- ✅ **Quick to implement**: 1-2 hours

### Downsides
- ⚠️ Less clean than subfolder approach
- ⚠️ Prefix adds noise to filenames
- ⚠️ Doesn't prevent mixing (just naming convention)

### Migration Effort
**Time**: 1-2 hours
**Risk**: Low (just renames, imports unchanged)

---

## Option 4: **Status Quo + Documentation** (Current)

Keep everything as-is, rely on CLAUDE.md documentation:

### Current Structure
```
go/pkg/
├── wazero/
│   ├── text.go           # 1:1 CRDT
│   ├── map.go            # 1:1 CRDT
│   ├── state.go          # Infrastructure (documented exception)
│   └── memory.go         # Infrastructure (documented exception)
```

### Benefits
- ✅ **No work**: Already done
- ✅ **Documented**: CLAUDE.md explains exceptions

### Downsides
- ❌ **Not obvious**: Need to read CLAUDE.md to understand
- ❌ **No enforcement**: Easy to accidentally put wrong file in wrong place

---

## Recommendation Matrix

| Criterion | Option 1 (Subfolder) | Option 2 (Packages) | Option 3 (Prefix) | Option 4 (Status Quo) |
|-----------|---------------------|-------------------|------------------|---------------------|
| **Clarity** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| **Ease of Migration** | ⭐⭐⭐ | ⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Maintainability** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| **Go Idioms** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |

---

## ✅ IMPLEMENTED: **Option 3 (Naming Convention)**

**Implementation Date**: 2025-10-21
**Commit**: `784f101` - "feat: implement Option 3 - crdt_ naming convention (49 files renamed)"

**What was done**:
- ✅ Renamed 49 files with `crdt_` prefix (33 Go + 16 web)
- ✅ Updated web/js/app.js imports
- ✅ Updated web/index.html component loading
- ✅ Updated Makefile variables
- ✅ All tests pass, server runs correctly

**Results**:
- ✅ **Grep-able**: `ls **/crdt_*.go` shows all CRDT files
- ✅ **Visual clarity**: CRDT vs infrastructure immediately obvious
- ✅ **Mobile-friendly**: Clean separation for gomobile
- ✅ **Self-documenting**: File names indicate CRDT operations
- ✅ **No import changes**: Go imports by package, not filename

**Why Option 1 failed**:
- ❌ **Doesn't work with Go's package system** (discovered during implementation)
- ❌ Subdirectories must be separate packages
- ❌ Creates circular dependency issues

See [Option 3 Rename Plan](option3-rename-plan.md) for complete implementation details.

---

## Implementation Plan for Option 1

If you choose to do this, here's the exact process:

### Step 1: Create Folders (5 min)
```bash
mkdir -p go/pkg/wazero/crdt
mkdir -p go/pkg/automerge/crdt
mkdir -p go/pkg/server/crdt
mkdir -p go/pkg/api/crdt
```

### Step 2: Move Files (10 min)
```bash
# wazero
git mv go/pkg/wazero/{text,map,list,counter,cursor,history,sync,richtext,generic}.go go/pkg/wazero/crdt/

# automerge
git mv go/pkg/automerge/{text,map,list,counter,cursor,history,sync,richtext,generic}.go go/pkg/automerge/crdt/

# server
git mv go/pkg/server/{text,map,list,counter,cursor,history,sync,richtext}.go go/pkg/server/crdt/

# api
git mv go/pkg/api/{text,map,list,counter,cursor,history,sync,richtext}.go go/pkg/api/crdt/
```

### Step 3: Update Imports (90 min)
```bash
# Update all files that import these packages
# Change: import "pkg/automerge"
# To: import "pkg/automerge/crdt"

# Can use sed or your IDE's refactor tool
```

### Step 4: Test (30 min)
```bash
make test-go
make build-wasi
make run
```

### Step 5: Update CLAUDE.md (15 min)
Update the 1:1 mapping table to show `crdt/` subfolders.

**Total Time**: ~2.5 hours

---

## Questions?

Which option do you prefer? Or want me to elaborate on any of them?
