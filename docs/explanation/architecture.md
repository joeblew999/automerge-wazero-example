# Go Package Architecture

This document describes the Go package architecture for the Automerge-wazero implementation.

## Overview

The Go codebase is organized into three main components:

```
go/
├── cmd/server/           # HTTP server binary
├── pkg/automerge/        # High-level Automerge API
├── pkg/wazero/          # Low-level WASM FFI
└── testdata/            # Test data and scripts
```

## Package Structure

### `pkg/wazero` - Low-Level FFI Layer

**Purpose**: Direct 1:1 wrapper around WASI exports with minimal abstraction.

**Files**:
- `runtime.go` - Wazero runtime lifecycle and export calling
- `exports.go` - Direct wrappers for all 11 WASI exports

**Responsibilities**:
- WASM module loading and instantiation
- Memory management (`am_alloc`, `am_free`)
- Raw FFI calls to WASI exports
- Error code checking

**Example Usage**:
```go
runtime, err := wazero.New(ctx, wazero.Config{})
if err != nil {
    return err
}
defer runtime.Close(ctx)

// Direct WASI call
err = runtime.AmInit(ctx)
text, err := runtime.AmGetText(ctx)
```

### `pkg/automerge` - High-Level API Layer

**Purpose**: Idiomatic Go API matching Automerge Rust semantics.

**Files**:
- `document.go` - Document lifecycle (New, Load, Save, Close)
- `text.go` - Text CRDT operations (implemented)
- `map.go` - Map operations (stubs, M2 milestone)
- `list.go` - List operations (stubs, M2 milestone)
- `counter.go` - Counter CRDT (stub, M2 milestone)
- `richtext.go` - Rich text formatting (stubs, M4 milestone)
- `sync.go` - Sync protocol (stubs, M1 milestone)
- `history.go` - Time-travel operations (stubs)
- `types.go` - Type system (Path, Value, ObjType, etc.)
- `errors.go` - Error types (NotImplementedError, DeprecatedError, WASMError)
- `doc.go` - Package documentation

**Responsibilities**:
- Path-based navigation (Root().Get("content"))
- Type-safe Value system
- Milestone-aware error messages
- Proper context propagation

**Example Usage**:
```go
doc, err := automerge.New(ctx)
if err != nil {
    return err
}
defer doc.Close(ctx)

// Work with text
path := automerge.Root().Get("content")
err = doc.SpliceText(ctx, path, 0, 0, "Hello")
text, err := doc.GetText(ctx, path)

// Save and merge
data, err := doc.Save(ctx)
other, err := automerge.Load(ctx, data)
err = doc.Merge(ctx, other)
```

### `cmd/server` - HTTP Server

**Purpose**: REST API and SSE server for collaborative editing.

**Endpoints**:
- `GET /api/text` - Get current text
- `POST /api/text` - Update text (broadcasts via SSE)
- `GET /api/stream` - SSE stream for live updates
- `GET /api/doc` - Download doc.am snapshot
- `POST /api/merge` - Merge another doc.am (CRDT merge)
- `GET /` - Serve UI

**Before Refactoring** (main.go ~716 lines):
- Direct wazero FFI calls inline
- Manual memory management in every handler
- Repetitive error handling
- Hard to test

**After Refactoring** (main.go ~358 lines, **50% smaller**):
- Clean `automerge.Document` API
- No manual FFI or memory management
- Simple, readable handlers
- Testable with mock documents

**Key Improvements**:
```go
// Before: ~50 lines of FFI code
func (s *Server) getText(ctx context.Context) (string, error) {
    getLenFn := s.modInst.ExportedFunction("am_get_text_len")
    results, err := getLenFn.Call(ctx)
    // ... 40 more lines ...
}

// After: 3 lines
func (s *Server) getText(ctx context.Context) (string, error) {
    path := automerge.Root().Get("content")
    return s.doc.GetText(ctx, path)
}
```

## API Coverage

### Currently Implemented (M0 - Phase 0 Complete)

✅ **Text Operations**:
- `GetText(ctx, path)` - Read text content
- `SpliceText(ctx, path, pos, del, text)` - CRDT text splice
- `TextLength(ctx, path)` - Get text length

✅ **Persistence**:
- `Save(ctx)` - Serialize to .am format
- `Load(ctx, data)` - Deserialize from .am

✅ **Merging**:
- `Merge(ctx, other)` - CRDT merge of two documents

✅ **Type System**:
- `Path` - Document navigation (Root().Get("key"))
- `Value` - Typed values (String, Int, Bool, etc.)
- `ObjType` - Object types (Map, List, Text)

### Not Implemented (Stubs with Milestone Info)

❌ **Map Operations** (M2):
- `Get`, `Put`, `Delete`, `Keys`, `PutObject`

❌ **List Operations** (M2):
- `Insert`, `Remove`, `Splice`, `InsertObject`

❌ **Counter CRDT** (M2):
- `Increment`

❌ **Sync Protocol** (M1):
- `GenerateSyncMessage`, `ReceiveSyncMessage`

❌ **Rich Text** (M4):
- `Mark`, `Unmark`, `GetMarks`, `SplitBlock`, `JoinBlock`

❌ **History** (Future):
- `GetHeads`, `GetChanges`, `Fork`, `ForkAt`

All unimplemented methods return `NotImplementedError` with:
- Feature name
- Target milestone (M1, M2, M3, M4)
- Helpful message explaining requirements

## Error Handling

### Three Error Types

1. **`NotImplementedError`** - Feature exists but not implemented yet
   ```go
   err := doc.Put(ctx, path, "key", value)
   // Returns: "automerge: feature 'Put' not yet implemented (planned for M2)"
   ```

2. **`DeprecatedError`** - Method exists but shouldn't be used
   ```go
   err := doc.UpdateText(ctx, path, text)
   // Returns: "automerge: UpdateText is deprecated (use SpliceText instead): destroys CRDT history"
   ```

3. **`WASMError`** - WASM operation failed
   ```go
   // From wazero package
   err := runtime.AmInit(ctx)
   // Returns: "WASM operation am_init failed with code -1"
   ```

## Testing Strategy

### Test Organization

```
go/
├── pkg/automerge/
│   └── document_test.go    # Unit tests for automerge package
├── pkg/wazero/
│   └── (no tests yet)      # Future: FFI-level tests
└── testdata/
    ├── snapshots/          # Binary .am test files
    ├── expected/           # Expected output files
    ├── scripts/            # Test data generators
    └── README.md           # Test data documentation
```

### Current Test Coverage

✅ **Type System Tests**:
- `TestPath` - Path construction and navigation
- `TestValue` - Value creation and type checking

✅ **Error Type Tests**:
- `TestNotImplementedError` - Error formatting and matching
- `TestDeprecatedError` - Deprecation warnings

⏸️ **Integration Tests** (skipped until WASM built):
- `TestNew` - Document creation
- `TestDocument_SpliceText_NotImplementedYet` - Text operations

### Test Data Generation

```bash
# Generate test snapshots (requires WASM built)
make generate-test-data

# Or manually:
cd go/testdata/scripts
./generate_test_data.sh
```

Generates:
- `empty.am` - Empty document
- `hello-world.am` - "Hello, World!"
- `simple-text.am` - ASCII text
- `unicode-text.am` - UTF-8/emoji
- `large-text.am` - ~10KB performance test

## Build and Development

### Building

```bash
# Build WASM module
make build-wasi

# Build Go server
make build-server

# Build and run
make run
```

### Testing

```bash
# Run all tests
make test

# Go tests only
make test-go

# Rust tests only
make test-rust
```

### Development Workflow

```bash
# Fast iteration (debug WASM build)
make dev

# Watch mode (requires air)
make watch

# Generate test data
make generate-test-data
```

## Module Structure

The Go module is structured as:

```
module github.com/joeblew999/automerge-wazero-example

go 1.25.3

require github.com/tetratelabs/wazero v1.9.0
```

**Important**:
- Module root is `go/` directory
- Packages are under `go/pkg/`
- Import paths are `github.com/joeblew999/automerge-wazero-example/pkg/{automerge,wazero}`
- **NOT** `go/pkg/...` (the `go/` is part of the directory structure, not the import path)

## Migration from Monolithic Server

The refactoring from monolithic `main.go` to the package structure:

**Before** (Single 716-line file):
```
go/cmd/server/
└── main.go               # Everything in one file
```

**After** (Organized packages):
```
go/
├── cmd/server/
│   └── main.go           # 358 lines, 50% reduction
├── pkg/automerge/        # 9 files, ~1200 lines
└── pkg/wazero/           # 2 files, ~265 lines
```

**Benefits**:
- ✅ Testable with mock documents
- ✅ Reusable across different server implementations
- ✅ Clear separation of concerns (FFI vs API)
- ✅ Type-safe Path and Value system
- ✅ Helpful error messages with milestones
- ✅ Easy to extend with new operations

## Future Milestones

### M1 - Sync Protocol
- Implement `GenerateSyncMessage`, `ReceiveSyncMessage`
- Add `am_sync_gen`, `am_sync_recv` WASI exports
- Update server to use sync messages instead of full snapshots

### M2 - Multi-Document Support
- Implement Map, List, Counter operations
- Add `am_get`, `am_put`, `am_insert`, etc. WASI exports
- Support arbitrary document structures (not just root["content"])

### M3 - NATS Transport
- Replace HTTP with NATS pub/sub
- Use NATS Object Store for snapshots
- JWT-based RBAC

### M4 - Rich Text
- Implement Mark, Unmark, SplitBlock operations
- Add formatting support (bold, italic, links)
- Block-based editor

## Contributing

When adding new API methods:

1. Add WASI export to `rust/automerge_wasi/src/lib.rs`
2. Add FFI wrapper to `go/pkg/wazero/exports.go`
3. Add high-level method to appropriate `go/pkg/automerge/*.go` file
4. Add tests to `go/pkg/automerge/*_test.go`
5. Update `API_MAPPING.md` with coverage matrix
6. Update this document if architecture changes

See [`api-mapping.md`](../reference/api-mapping.md) for detailed API reference.
