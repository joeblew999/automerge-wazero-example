# Option 3: crdt_ Naming Convention - Rename Plan

**Date**: 2025-10-21
**Status**: Ready for execution

## Summary

Rename all CRDT operation files with `crdt_` prefix to visually separate them from infrastructure files.

## Files to Rename

### Layer 3: Go FFI (go/pkg/wazero/) - 9 files

| Old Name | New Name |
|----------|----------|
| text.go | crdt_text.go |
| map.go | crdt_map.go |
| list.go | crdt_list.go |
| counter.go | crdt_counter.go |
| history.go | crdt_history.go |
| sync.go | crdt_sync.go |
| richtext.go | crdt_richtext.go |
| cursor.go | crdt_cursor.go |
| generic.go | crdt_generic.go |

**Keep as-is** (infrastructure):
- runtime.go
- state.go
- memory.go
- document.go

### Layer 4: Go API (go/pkg/automerge/) - 9 files

| Old Name | New Name |
|----------|----------|
| text.go | crdt_text.go |
| map.go | crdt_map.go |
| list.go | crdt_list.go |
| counter.go | crdt_counter.go |
| history.go | crdt_history.go |
| sync.go | crdt_sync.go |
| richtext.go | crdt_richtext.go |
| cursor.go | crdt_cursor.go |
| generic.go | crdt_generic.go |

**Keep as-is** (infrastructure):
- document.go
- errors.go
- types.go
- doc.go

### Layer 5: Server (go/pkg/server/) - 8 files

| Old Name | New Name |
|----------|----------|
| text.go | crdt_text.go |
| map.go | crdt_map.go |
| list.go | crdt_list.go |
| counter.go | crdt_counter.go |
| history.go | crdt_history.go |
| sync.go | crdt_sync.go |
| richtext.go | crdt_richtext.go |
| cursor.go | crdt_cursor.go |

**Keep as-is** (infrastructure):
- server.go
- broadcast.go
- document.go

### Layer 6: HTTP API (go/pkg/api/) - 7 files

| Old Name | New Name |
|----------|----------|
| map.go | crdt_map.go |
| list.go | crdt_list.go |
| counter.go | crdt_counter.go |
| history.go | crdt_history.go |
| sync.go | crdt_sync.go |
| richtext.go | crdt_richtext.go |
| cursor.go | crdt_cursor.go |

**Keep as-is** (infrastructure):
- handlers.go (legacy text handler - has layer marker)
- util.go
- static.go
- text_test.go (test file)

**Note**: No `text.go` in api/ - handled by handlers.go

### Layer 7: Web JS (web/js/) - 8 files

| Old Name | New Name |
|----------|----------|
| text.js | crdt_text.js |
| map.js | crdt_map.js |
| list.js | crdt_list.js |
| counter.js | crdt_counter.js |
| history.js | crdt_history.js |
| sync.js | crdt_sync.js |
| richtext.js | crdt_richtext.js |
| cursor.js | crdt_cursor.js |

**Keep as-is** (infrastructure):
- app.js

### Layer 7: Web HTML (web/components/) - 8 files

| Old Name | New Name |
|----------|----------|
| text.html | crdt_text.html |
| map.html | crdt_map.html |
| list.html | crdt_list.html |
| counter.html | crdt_counter.html |
| history.html | crdt_history.html |
| sync.html | crdt_sync.html |
| richtext.html | crdt_richtext.html |
| cursor.html | crdt_cursor.html |

**No infrastructure files in components/**

## Total Files to Rename

- **Go backend**: 33 files (9 + 9 + 8 + 7)
- **Web frontend**: 16 files (8 + 8)
- **Grand total**: 49 files

## Import Updates Required

### Go Files

After renaming, need to search and replace in ALL Go files:

```bash
# Example patterns that will break:
import "github.com/joeblew999/automerge-wazero-example/pkg/wazero"
// Uses: wazero.Text... → no change needed (package name stays same)

# But function file references in tests might need updates
```

**Important**: Go imports are by package, not file. Since package names don't change, most imports will work without modification. Only need to update:
- Test file references
- File-level comments/docs
- Build tags if any

### Web Files

Update `web/js/app.js`:

```javascript
// OLD:
import { TextComponent } from './text.js';
import { SyncComponent } from './sync.js';
// ... 6 more

// NEW:
import { TextComponent } from './crdt_text.js';
import { SyncComponent } from './crdt_sync.js';
// ... 6 more
```

Update component loading in `web/js/app.js`:

```javascript
// OLD:
fetch('/web/components/text.html')
fetch('/web/components/sync.html')
// ... 6 more

// NEW:
fetch('/web/components/crdt_text.html')
fetch('/web/components/crdt_sync.html')
// ... 6 more
```

### Makefile

Update `WEB_JS` and `WEB_COMPONENTS` variables:

```makefile
# OLD:
WEB_JS = $(WEB_DIR)/js/text.js $(WEB_DIR)/js/sync.js ...

# NEW:
WEB_JS = $(WEB_DIR)/js/crdt_text.js $(WEB_DIR)/js/crdt_sync.js ...
```

## Execution Order

1. ✅ Create this rename plan document
2. Rename Go backend files (git mv)
3. Rename web frontend files (git mv)
4. Update web/js/app.js imports
5. Update Makefile variables
6. Test: `make build-wasi && make test-go`
7. Test: `make run` → verify web UI loads
8. Commit all changes together
9. Update folder-restructuring-options.md to mark Option 3 as "IMPLEMENTED"

## Benefits

- ✅ **Grep-able**: `ls **/crdt_*.go` shows all CRDT files
- ✅ **Visual separation**: CRDT vs infrastructure immediately obvious
- ✅ **Mobile-friendly**: Useful for gomobile code organization
- ✅ **No import changes**: Go imports by package, not file
- ✅ **Quick to implement**: Just renames + update app.js

## Rollback Plan

If something breaks:

```bash
git reset --hard HEAD  # Undo all changes
```

Since we're using `git mv`, git tracks the renames. Easy to revert.
