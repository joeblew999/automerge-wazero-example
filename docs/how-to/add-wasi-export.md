# How to Add a New WASI Export

> **Goal**: Add a new Automerge operation exposed via WASI to the Go server

## Prerequisites

- Understanding of the [4-layer architecture](../explanation/architecture.md)
- Familiarity with [Automerge Rust API](../reference/api-mapping.md)

## Steps

### 1. Add Rust WASI Export

**File**: `rust/automerge_wasi/src/<module>.rs`

```rust
#[no_mangle]
pub extern "C" fn am_your_function(param: u32) -> i32 {
    // Get document from state
    // Call Automerge API
    // Return result
    0  // success
}
```

See existing exports in:
- `rust/automerge_wasi/src/text.rs` - Text operations
- `rust/automerge_wasi/src/map.rs` - Map operations
- `rust/automerge_wasi/src/list.rs` - List operations

### 2. Build WASM Module

```bash
make build-wasi
```

### 3. Add Go FFI Wrapper

**File**: `go/pkg/wazero/<module>.go` (match Rust module name!)

```go
// YourFunction calls am_your_function WASI export
func (r *Runtime) AmYourFunction(ctx context.Context, param uint32) error {
    results, err := r.callExport(ctx, "am_your_function", uint64(param))
    if err != nil {
        return err
    }
    return r.checkErrorCode(int32(results[0]))
}
```

### 4. Add High-Level Go API

**File**: `go/pkg/automerge/<category>.go`

```go
// YourOperation performs...
func (d *Document) YourOperation(param uint32) error {
    return d.runtime.AmYourFunction(d.ctx, param)
}
```

### 5. Add Tests

**File**: `go/pkg/automerge/<category>_test.go`

```go
func TestDocument_YourOperation(t *testing.T) {
    doc := setupTestDocument(t)
    defer doc.Close()

    err := doc.YourOperation(42)
    assert.NoError(t, err)
}
```

### 6. Update Documentation

- Update [API Mapping](../reference/api-mapping.md) with new function
- Mark as "Implemented" in coverage matrix
- Update this guide if you found issues!

### 7. Verify

```bash
make build-wasi  # Rebuild WASM
make test-go     # Run tests
```

## Common Patterns

### Memory Management

```rust
// Allocate memory for output
#[no_mangle]
pub extern "C" fn am_get_output_len() -> u32 {
    // Return length needed
}

#[no_mangle]
pub extern "C" fn am_get_output(ptr_out: *mut u8) -> i32 {
    // Write to ptr_out
}
```

### Error Handling

```rust
// Return error codes
const ERR_INVALID_INPUT: i32 = -1;
const ERR_NOT_FOUND: i32 = -2;

if invalid {
    return ERR_INVALID_INPUT;
}
```

## Troubleshooting

- **Export not found**: Did you rebuild WASM? (`make build-wasi`)
- **Memory errors**: Check buffer sizes and null pointers
- **Test failures**: Verify WASI export signature matches Go wrapper

## See Also

- [Architecture Guide](../explanation/architecture.md) - 4-layer design
- [API Mapping](../reference/api-mapping.md) - Existing exports
- [Testing Guide](../development/testing.md) - Writing tests

---

**Last Updated**: 2025-10-20 (Placeholder - to be expanded)
