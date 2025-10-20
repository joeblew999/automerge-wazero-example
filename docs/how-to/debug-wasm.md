# How to Debug WASM Issues

> **Goal**: Troubleshoot and debug WASI/WASM integration problems

## Prerequisites

- Basic understanding of [WASM/WASI](https://wasi.dev)
- Familiarity with [wazero](https://wazero.io)

## Quick Diagnostics

### 1. Verify WASM Module

```bash
make build-wasi
file rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm
```

Should output: `WebAssembly (wasm) binary module version 0x1 (MVP)`

### 2. Check Export List

```bash
wasm-objdump -x rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm | grep -A 100 "Export"
```

Verify your function appears in the export list.

### 3. Enable Debug Logging

**Go side**:
```go
// In go/pkg/wazero/state.go
log.Printf("Calling WASM export: %s with params: %v", name, params)
```

**Rust side**:
```rust
// Add to your function
eprintln!("WASM: am_your_function called with param={}", param);
```

## Common Issues

### "Export not found"

**Problem**: `callExport` fails with "export X not found"

**Solutions**:
1. Verify `#[no_mangle]` attribute on Rust function
2. Rebuild WASM: `make clean && make build-wasi`
3. Check function signature matches exactly

### Memory Access Violations

**Problem**: Segfault or panic when accessing WASM memory

**Solutions**:
1. Check pointer is valid (not null)
2. Verify buffer size before writing
3. Use `am_alloc` for output buffers
4. Always `am_free` allocated memory

### Type Mismatches

**Problem**: Wrong data returned or garbled values

**Solutions**:
1. Verify integer sizes match (u32 vs u64)
2. Check endianness (WASM is little-endian)
3. Ensure UTF-8 encoding for strings

## Debugging Workflow

### 1. Build Debug WASM

```bash
make build-wasi-debug  # Faster compile, includes symbols
```

### 2. Add Logging

**Rust**:
```rust
eprintln!("DEBUG: variable = {:?}", variable);
```

**Go**:
```go
log.Printf("DEBUG: calling %s", exportName)
```

### 3. Run Tests with Verbose Output

```bash
make test-go 2>&1 | tee test.log
```

### 4. Inspect WASM Memory

```go
// In Go test
data, _ := doc.runtime.modInst.Memory().Read(ptr, size)
log.Printf("Memory at %d: %v", ptr, data)
```

## Tools

### wasm-objdump

```bash
# Install
cargo install wabt

# Disassemble
wasm-objdump -d automerge_wasi.wasm > disasm.txt
```

### wazero Debugging

```go
// Enable stack traces
import "github.com/tetratelabs/wazero"

ctx := context.WithValue(ctx, "debug", true)
```

## Testing Strategy

1. **Unit test Rust** - Test WASI exports directly
2. **Unit test Go** - Test FFI wrappers
3. **Integration test** - Test end-to-end flow

Example:
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_am_your_function() {
        let result = am_your_function(42);
        assert_eq!(result, 0);
    }
}
```

## Performance Profiling

```bash
# Go profiling
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...
go tool pprof cpu.prof
```

## See Also

- [Architecture](../explanation/architecture.md) - Understanding the layers
- [Add WASI Export](add-wasi-export.md) - Step-by-step guide
- [Testing Guide](../development/testing.md) - Test strategies

---

**Last Updated**: 2025-10-20 (Placeholder - to be expanded)
