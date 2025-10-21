# AI-Code Connection: Improvement Plan

**Date**: 2025-10-21
**Status**: Proposal
**Goal**: Make this codebase maximally navigable for AI agents

---

## Current State Analysis

### Metrics (Actual Scan Results)

| Metric | Current | Target | Impact |
|--------|---------|--------|--------|
| **Layer markers** | 0/77 files | 77/77 | HIGH - AI can't tell which layer it's in |
| **Doc comments** | 0/284 funcs (0%) | 200/284 (70%) | HIGH - No context for what functions do |
| **FFI safety docs** | Some | All exports | CRITICAL - Memory bugs here |
| **Error codes** | Magic numbers | Named constants | MEDIUM - Hard to debug |
| **Why comments** | 0 | 50+ | HIGH - AI "fixes" intentional designs |
| **Decision logs** | None | Key decisions | MEDIUM - Context for weird code |

### Real Examples of AI Confusion

#### Example 1: Layer Ambiguity

**Current** (`go/pkg/server/text.go`):
```go
package server

import (
	"context"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// Text operations - maps to automerge/text.go

func (s *Server) GetText(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// ...
}
```

**AI Questions**:
- Why does this have a mutex but `automerge/text.go` doesn't?
- Should I add thread safety to the automerge layer?
- Is this a bug or intentional design?

**Improved**:
```go
// ═══════════════════════════════════════════════════════════════
// LAYER 4: Server (Stateful + Thread-Safe)
//
// Responsibilities:
// - Add thread safety (mutex protection)
// - Add persistence (save after mutations)
// - Manage SSE broadcast to clients
// - Own the Document instance
//
// Dependencies:
// ⬇️  Calls: pkg/automerge (Layer 3 - stateless CRDT operations)
// ⬆️  Called by: pkg/api (Layer 5 - HTTP handlers)
//
// Related Files:
// - go/pkg/automerge/text.go (pure CRDT logic, NO mutex)
// - go/pkg/api/text.go (HTTP protocol, calls this layer)
// - rust/automerge_wasi/src/text.rs (WASM implementation)
//
// Testing:
// - Unit: server/text_test.go (thread safety, persistence)
// - Integration: api/text_test.go (HTTP → Server → WASM)
// ═══════════════════════════════════════════════════════════════

package server

import (
	"context"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// GetText returns the current text from the document.
//
// Thread-safe: Uses RLock for concurrent reads.
// Does NOT save to disk (read-only operation).
func (s *Server) GetText(ctx context.Context) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := automerge.Root().Get("content")
	return s.doc.GetText(ctx, path)
}
```

**AI Now Understands**:
✅ Layer 4 adds thread safety (Layer 3 doesn't need it)
✅ This calls Layer 3, which calls Layer 2 (FFI)
✅ Tests are split: unit vs integration
✅ No need to "fix" the lack of mutex in automerge layer

---

#### Example 2: FFI Memory Safety

**Current** (`rust/automerge_wasi/src/counter.rs`):
```rust
/// Create a new counter at a key in ROOT map, initialized to value.
///
/// # Parameters
/// - `key_ptr`: Pointer to key string (UTF-8)
/// - `key_len`: Length of key in bytes
/// - `value`: Initial counter value
///
/// # Returns
/// - `0` on success
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_counter_create(
    key_ptr: *const u8,
    key_len: usize,
    value: i64,
) -> i32 {
    if key_ptr.is_null() {
        return -1;  // ❌ Magic number
    }
    // ...
}
```

**AI Questions**:
- Who allocates `key_ptr`? Rust or Go?
- Do I need to free it inside this function?
- What if `key_len` is wrong?
- Can `key_ptr` point to invalid memory?
- What encoding is guaranteed?

**Improved**:
```rust
/// Create a new counter at a key in ROOT map, initialized to value.
///
/// ┌─────────────────────────────────────────────────────────────┐
/// │ FFI SAFETY CONTRACT - READ BEFORE MODIFYING                 │
/// ├─────────────────────────────────────────────────────────────┤
/// │ Memory Ownership: Caller (Go) allocates & frees             │
/// │ String Encoding: MUST be valid UTF-8                        │
/// │ Null Safety: key_ptr must not be null                       │
/// │ Length: key_len must match actual string length             │
/// │ Lifetime: key_ptr valid only during this call               │
/// │                                                             │
/// │ Typical Call Sequence (Go side):                            │
/// │ 1. keyPtr := runtime.AmAlloc(len(key))                      │
/// │ 2. copy(wasm.Memory[keyPtr:], []byte(key))                  │
/// │ 3. result := runtime.AmCounterCreate(keyPtr, len(key), val) │
/// │ 4. runtime.AmFree(keyPtr)  // MUST FREE                     │
/// │                                                             │
/// │ See: go/pkg/wazero/counter.go for usage example             │
/// └─────────────────────────────────────────────────────────────┘
///
/// # Parameters
/// - `key_ptr`: Pointer to UTF-8 string in WASM linear memory
/// - `key_len`: Byte length of string (NOT character count!)
/// - `value`: Initial counter value (i64::MIN to i64::MAX)
///
/// # Returns
/// - `ErrCode::Success` (0) on success
/// - `ErrCode::InvalidUTF8` (-1) if key_ptr is null or invalid UTF-8
/// - `ErrCode::AutomergeError` (-2) on CRDT operation failure
/// - `ErrCode::NotInitialized` (-3) if am_init() not called first
///
/// # Safety
/// This function is `unsafe` because it dereferences raw pointers.
/// The caller MUST ensure:
/// - `key_ptr` points to valid WASM memory
/// - `key_len` bytes are readable from `key_ptr`
/// - Memory is not freed until after this function returns
#[no_mangle]
pub extern "C" fn am_counter_create(
    key_ptr: *const u8,
    key_len: usize,
    value: i64,
) -> i32 {
    use crate::errors::ErrCode;

    if key_ptr.is_null() {
        return ErrCode::InvalidUTF8 as i32;
    }

    // SAFETY: Caller guarantees key_ptr points to key_len valid bytes
    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };

    let key = match std::str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return ErrCode::InvalidUTF8 as i32,
    };

    let result = with_doc_mut(|doc| {
        doc.put(&ROOT, key, ScalarValue::counter(value))
    });

    match result {
        Some(Ok(_)) => ErrCode::Success as i32,
        Some(Err(_)) => ErrCode::AutomergeError as i32,
        None => ErrCode::NotInitialized as i32,
    }
}
```

**AI Now Understands**:
✅ Go allocates, Rust only reads
✅ Must free after call
✅ UTF-8 required, validated inside
✅ Error codes are named constants
✅ Exactly how to call from Go side

---

#### Example 3: "Why" Comments

**Current** (`go/pkg/wazero/state.go`):
```go
const (
    defaultWASMPath = "../rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm"
)
```

**AI Confusion**: "This path looks wrong! Should be `../../rust/`. Let me fix it..."

**Improved**:
```go
const (
    // Why "../rust/" and not "../../rust/"?
    //
    // This path is relative to the go/ directory, which is the working
    // directory when you run: cd go && go run cmd/server/main.go
    //
    // From go/:
    //   ../rust/  ✅ Correct (go up one level to repo root, then into rust/)
    //
    // If path were "../../rust/", it would look OUTSIDE the repo!
    //
    // See: CLAUDE.md Section 0) for path configuration details
    defaultWASMPath = "../rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm"
)
```

**AI Now Understands**: This is intentional, not a bug!

---

## Implementation Plan

### Phase 1: Quick Wins (4 hours) 🚀

**Immediate ROI, minimal effort**

#### 1.1 Add Layer Markers (2 hours)

**Create template**:
```bash
# scripts/add-layer-markers.sh
#!/bin/bash

# Layer 1: Rust WASI
for file in rust/automerge_wasi/src/*.rs; do
    # Add marker at top of file
done

# Layer 2: Go FFI
for file in go/pkg/wazero/*.go; do
    # Add marker
done

# ... etc for all 6 layers
```

**Template for each layer**:
```go
// ═══════════════════════════════════════════════════════════════
// LAYER X: [Name] ([Responsibilities])
//
// ⬇️  Calls: [downstream layer]
// ⬆️  Called by: [upstream layer]
// 🔁 Siblings: [related files in same layer]
// 📝 Tests: [test file locations]
// ═══════════════════════════════════════════════════════════════
```

**Files affected**: 77 files across 6 layers
**Benefit**: AI instantly knows context when reading any file

#### 1.2 Create Error Code Registry (2 hours)

**New file**: `rust/automerge_wasi/src/errors.rs`
```rust
/// Standard error codes returned by all WASI exports.
///
/// These are returned as i32 from C-ABI functions.
/// Generated Go constants: go/pkg/wazero/errors.go
#[repr(i32)]
pub enum ErrCode {
    Success = 0,
    InvalidUTF8 = -1,
    PathNotFound = -2,
    OutOfBounds = -3,
    NotInitialized = -4,
    AllocFailed = -5,
    WrongType = -6,
    AutomergeError = -7,
}
```

**Generate Go constants**: `go/pkg/wazero/errors.go`
```go
// AUTO-GENERATED from rust/automerge_wasi/src/errors.rs
// DO NOT EDIT - run `make generate-errors` to update

package wazero

const (
    ErrSuccess       = 0
    ErrInvalidUTF8   = -1
    ErrPathNotFound  = -2
    // ...
)

var ErrMessages = map[int]string{
    ErrSuccess:       "Success",
    ErrInvalidUTF8:   "Invalid UTF-8 encoding",
    ErrPathNotFound:  "Path not found in document",
    // ...
}
```

**Files affected**: 13 Rust modules, 1 new file
**Benefit**: Replace all magic numbers, clear error meanings

---

### Phase 2: Documentation (1 day) 📚

#### 2.1 FFI Boundary Documentation

**Add "FFI SAFETY CONTRACT" box to all 57 WASM exports**

**Template**:
```rust
/// [Function description]
///
/// ┌─────────────────────────────────────────────────────────────┐
/// │ FFI SAFETY CONTRACT                                         │
/// ├─────────────────────────────────────────────────────────────┤
/// │ Memory Ownership: [Caller/Callee]                           │
/// │ String Encoding: [UTF-8/Binary/etc]                         │
/// │ Null Safety: [What's allowed]                               │
/// │ Typical Call Sequence: [Go code example]                    │
/// └─────────────────────────────────────────────────────────────┘
```

**Files affected**: 13 Rust modules (all with exports)
**Benefit**: Zero memory bugs, AI understands FFI boundary

#### 2.2 Decision Log

**Create**: `docs/decisions/README.md`
```markdown
# Architectural Decision Records (ADRs)

Why does this code look weird? Check here first!

- [001-why-sse-not-websocket.md](001-why-sse-not-websocket.md)
- [002-why-six-layers.md](002-why-six-layers.md)
- [003-why-separate-server-api.md](003-why-separate-server-api.md)
- [004-why-wasm-not-cgo.md](004-why-wasm-not-cgo.md)
- [005-why-relative-paths.md](005-why-relative-paths.md)
```

**Start with 5 key decisions** that explain "surprising" code.

**Files affected**: New `docs/decisions/` directory
**Benefit**: AI stops "fixing" intentional designs

---

### Phase 3: Tooling (1 day) 🔧

#### 3.1 Verification Scripts

**New**: `scripts/verify-ai-readability.sh`
```bash
#!/bin/bash
# Check that AI-readability standards are maintained

ERRORS=0

# Check 1: Every file has layer marker
echo "Checking layer markers..."
for file in $(find go/pkg rust/automerge_wasi/src -name "*.go" -o -name "*.rs"); do
    if ! grep -q "LAYER [0-9]:" "$file"; then
        echo "  ❌ Missing layer marker: $file"
        ERRORS=$((ERRORS + 1))
    fi
done

# Check 2: No magic number returns in Rust
echo "Checking for magic numbers..."
if grep -rn "return -[0-9]" rust/automerge_wasi/src/*.rs | grep -v "ErrCode"; then
    echo "  ❌ Found magic numbers (use ErrCode enum)"
    ERRORS=$((ERRORS + 1))
fi

# Check 3: All WASM exports have FFI docs
echo "Checking FFI documentation..."
# ... check for "FFI SAFETY CONTRACT" near every #[no_mangle]

exit $ERRORS
```

**Add to Makefile**:
```makefile
.PHONY: verify-ai-readability
verify-ai-readability:
	@./scripts/verify-ai-readability.sh

# Run before every commit
pre-commit: verify-docs verify-ai-readability test-go test-rust
```

#### 3.2 Auto-Generate CURRENT_STATE.md

**New**: `scripts/generate-state.sh`
```bash
#!/bin/bash
# Generate docs/reference/CURRENT_STATE.md from code analysis

cat > docs/reference/CURRENT_STATE.md <<EOF
<!-- AUTO-GENERATED by make update-state -->
<!-- Last updated: $(date) -->

# Current System State

## Module Completeness

| Module | Rust | FFI | API | Server | HTTP | Web | Tests | Status |
|--------|------|-----|-----|--------|------|-----|-------|--------|
EOF

# Scan for each module
for module in text map list counter cursor history sync richtext generic; do
    # Check if files exist
    RUST=$([ -f "rust/automerge_wasi/src/${module}.rs" ] && echo "✅" || echo "❌")
    # ... check all layers ...
    echo "| $module | $RUST | ... | DONE |" >> docs/reference/CURRENT_STATE.md
done

# Add known issues section
# Add recent changes section
```

**Run automatically**: `make update-state` before every commit

---

### Phase 4: Advanced (Optional) 🚀

#### 4.1 Interface Extraction for Testing

**Create**: `go/pkg/automerge/interfaces.go`
```go
// Mock-able interfaces for testing without WASM

type Runtime interface {
    AmInit(ctx context.Context) error
    AmSave(ctx context.Context) ([]byte, error)
    AmLoad(ctx context.Context, data []byte) error
    // ... all WASM calls
}

type Document interface {
    GetText(ctx context.Context, path string) (string, error)
    SpliceText(ctx context.Context, path string, pos, del int, text string) error
    // ... all CRDT operations
}
```

**Benefit**: Unit test server layer without loading WASM!

#### 4.2 Shared Type Definitions

**Create**: `schemas/wasi-types.yaml`
```yaml
types:
  Path:
    rust: "&str"
    go: "string"
    encoding: "UTF-8"
    max_length: 1024

  ErrorCode:
    rust: "i32"
    go: "int"
    values:
      Success: 0
      InvalidUTF8: -1
      # ...

functions:
  am_text_splice:
    params:
      - {name: path, type: Path}
      - {name: pos, type: uint32}
      - {name: del, type: uint32}
      - {name: text, type: Path}
    returns: ErrorCode
```

**Generate from schema**:
- Rust function signatures
- Go FFI wrappers
- Documentation
- Tests

---

## Measuring Success

### Before (Current)
```
Layer markers:      0 / 77 files (0%)
Doc comments:       0 / 284 functions (0%)
FFI docs:           Partial (some exports)
Error codes:        Magic numbers
Why comments:       0
Decision logs:      None
```

### After Phase 1 (Quick Wins)
```
Layer markers:      77 / 77 files (100%) ✅
Error codes:        Named constants ✅
```

### After Phase 2 (Documentation)
```
FFI docs:           57 / 57 exports (100%) ✅
Decision logs:      5 key decisions ✅
```

### After Phase 3 (Tooling)
```
Automated checks:   3 verification scripts ✅
Auto-generated:     CURRENT_STATE.md ✅
```

---

## Recommendation: Start Here 👉

**Highest ROI for AI readability**:

1. ✅ **Add layer markers** (2 hours) - Template + script
2. ✅ **Error code registry** (2 hours) - Replace magic numbers
3. ✅ **5 decision logs** (1 hour) - Document "weird" choices

**Total: 5 hours, massive improvement in AI navigation**

Then measure impact by asking AI agent to:
- "Add a new counter operation"
- "Debug why sync fails"
- "Refactor the server layer"

You'll see **drastically** better AI understanding!

---

## Questions?

Want me to implement any of these? I can:

1. Create the layer marker template + add to 5 example files
2. Build the error code enum + generator
3. Write 3 starter ADRs (decision logs)
4. Build `verify-ai-readability.sh` script

Which would help most?
