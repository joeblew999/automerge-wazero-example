# Test Fix Summary - WASM Path Refactoring

## Problem
After making WASM_PATH required in the library refactoring, ALL ~69 tests were failing with:
```
WASMPath is required in wazero.Config
```

## Root Cause
The refactoring changed:
- `automerge.New(ctx)` → requires `automerge.NewWithWASM(ctx, wasmPath)`
- `automerge.Load(ctx, data)` → requires `automerge.LoadWithWASM(ctx, data, wasmPath)`

All test files were still using the old API without providing WASM_PATH.

## Solution

### 1. Created Test Helper Constant
**File**: [go/pkg/automerge/testing.go](go/pkg/automerge/testing.go)
```go
const (
    // TestWASMPath is the path to the WASM module used in tests.
    // Path is relative from go/pkg/automerge/ directory where tests run.
    TestWASMPath = "../../../rust/automerge_wasi/target/wasm32-wasip1/debug/automerge_wasi.wasm"
)
```

### 2. Fixed API Test Helper
**File**: [go/pkg/api/util_test.go](go/pkg/api/util_test.go)
```go
import "github.com/joeblew999/automerge-wazero-example/pkg/automerge"

srv := server.New(server.Config{
    StorageDir: t.TempDir(),
    UserID:     "test-user",
    WASMPath:   automerge.TestWASMPath,  // Added this
})
```

### 3. Batch-Fixed All Test Files
Used `perl` (sed on macOS was not working) to replace:

**Files Fixed** (11 total):
- go/pkg/automerge/document_test.go
- go/pkg/automerge/crdt_text_test.go
- go/pkg/automerge/crdt_cursor_test.go
- go/pkg/automerge/crdt_counter_test.go
- go/pkg/automerge/crdt_history_test.go
- go/pkg/automerge/crdt_list_test.go
- go/pkg/automerge/crdt_map_test.go
- go/pkg/automerge/crdt_richtext_test.go
- go/pkg/automerge/crdt_sync_test.go
- go/pkg/automerge/crdt_generic_test.go
- go/pkg/api/util_test.go

**Replacements**:
```perl
# For bare New() calls (package automerge tests)
perl -pi -e 's/\bNew\(ctx\)/NewWithWASM(ctx, TestWASMPath)/g'

# For qualified New() calls (package automerge_test tests)
perl -pi -e 's/automerge\.New\(ctx\)/automerge.NewWithWASM(ctx, automerge.TestWASMPath)/g'

# For Load() calls
perl -pi -e 's/\bLoad\(ctx, ([^)]+)\)/LoadWithWASM(ctx, $1, TestWASMPath)/g'
```

## Results

### Before
- ❌ 0 tests passing
- ❌ ~69 tests failing
- ❌ All failures: "WASMPath is required in wazero.Config"

### After
- ✅ 69 tests passing
- ✅ 0 tests failing
- ✅ Test suite: PASS

### Test Execution Time
- pkg/api: cached (all passing)
- pkg/automerge: 74.877s (69 tests passing)

## Test Coverage by Module

| Module | Tests | Status |
|--------|-------|--------|
| Document | 11 | ✅ PASS |
| Text | 16 | ✅ PASS |
| Map | 9 | ✅ PASS |
| List | 4 | ✅ PASS |
| Counter | 3 | ✅ PASS |
| History | 5 | ✅ PASS |
| Sync | 3 | ✅ PASS |
| RichText | 7 | ✅ PASS |
| Cursor | 4 | ✅ PASS |
| Generic | 7 | ✅ PASS |
| **Total** | **69** | **✅ PASS** |

## Key Learnings

1. **macOS sed doesn't work with `-i` flag properly** - had to use `perl -pi -e` instead
2. **Test path relativity matters** - needed `../../../rust/` not `../../rust/` from `go/pkg/automerge/`
3. **Package-level tests vs external tests** - Different import patterns:
   - `package automerge`: Use bare `New(ctx)` → needs `NewWithWASM(ctx, TestWASMPath)`
   - `package automerge_test`: Use `automerge.New(ctx)` → needs `automerge.NewWithWASM(ctx, automerge.TestWASMPath)`

## Verification

```bash
make test
# 🧪 Running Rust tests...
# PASS
# 🧪 Running Go tests...
# PASS
# ok  	github.com/joeblew999/automerge-wazero-example/pkg/api	(cached)
# PASS
# ok  	github.com/joeblew999/automerge-wazero-example/pkg/automerge	74.877s
```

## Future Test Pattern

For any new test files:

```go
package automerge  // or automerge_test

import (
    "context"
    "testing"
    "github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

func TestExample(t *testing.T) {
    ctx := context.Background()
    
    // ✅ CORRECT - Always use NewWithWASM
    doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
    if err != nil {
        t.Fatalf("Failed to create document: %v", err)
    }
    defer doc.Close(ctx)
    
    // For loading
    data, _ := doc.Save(ctx)
    doc2, err := automerge.LoadWithWASM(ctx, data, automerge.TestWASMPath)
    if err != nil {
        t.Fatalf("Failed to load document: %v", err)
    }
    defer doc2.Close(ctx)
}
```

## Date
2025-10-21
