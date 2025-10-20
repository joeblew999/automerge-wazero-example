# Testing Guide - Automerge WASI Example

> **Complete testing strategy for Go + Rust + WASM collaborative text editor**

## 📊 Testing Stack Overview

| Test Layer | Tool | Location | Status | Coverage |
|------------|------|----------|--------|----------|
| **Unit Tests (Go)** | `go test` | `go/pkg/automerge/document_test.go` | ✅ 11/12 passing | Core CRDT operations |
| **Unit Tests (Rust)** | `cargo test` | `rust/automerge_wasi/src/*.rs` | ✅ Passing | WASI exports |
| **Integration Test** | Bash script | `testdata/integration/test_merge.sh` | ⚠️ Known issue | Distributed CRDT merge |
| **E2E Tests** | Playwright MCP | Manual via Claude Code | ✅ Working | Browser UI + SSE |

---

## 🧪 1. Go Unit Tests

**Location**: `go/pkg/automerge/document_test.go`

**Run**: `cd go/pkg/automerge && go test -v`

**Coverage**: 11/12 tests passing (1 skipped)

### Test Categories

#### ✅ Document Lifecycle (5 tests)
- `TestNew` - Create empty document
- `TestDocument_SaveAndLoad` - Serialize/deserialize round-trip
- `TestDocument_LoadFromTestData` - Load pre-generated snapshots
  - `empty.am` - Empty document
  - `hello-world.am` - Simple text
  - `simple-text.am` - Longer text
  - `unicode-text.am` - Unicode + emoji

#### ✅ Text CRDT Operations (3 tests)
- `TestDocument_SpliceText` - Insert, append, delete, replace
- `TestDocument_SpliceText_Unicode` - Unicode + emoji handling
- `TestDocument_TextLength` - Byte length calculation

#### ⚠️ CRDT Merge (1 test - SKIPPED)
- `TestDocument_Merge` - **Known Issue**: Only preserves one document's changes
  - Expected: Merge "Hi Hello" + "Hello World" → "Hi Hello World"
  - Actual: Only preserves one side
  - Root cause: Under investigation (Automerge 0.5 merge API)

#### ✅ Error Handling (2 tests)
- `TestDocument_Get_NotImplemented` - Verify NotImplementedError
- `TestNotImplementedError` / `TestDeprecatedError` - Error types

#### ✅ API Design (2 tests)
- `TestPath` - Path construction (`/content`, `/users[0]/name`)
- `TestValue` - Value types (string, int, bool, float)

### Test Data

**Location**: `testdata/snapshots/*.am`

**Generate new test data**: `make generate-test-data`

```bash
testdata/snapshots/
├── empty.am              # Empty CRDT document
├── hello-world.am        # "Hello, World!"
├── simple-text.am        # Longer sentence
└── unicode-text.am       # Unicode + emoji
```

---

## 🦀 2. Rust Unit Tests

**Location**: `rust/automerge_wasi/src/*.rs`

**Run**: `cargo test` or `make test-rust`

**Coverage**: WASI export validation

### Test Modules

Each Rust module has inline tests:

- `src/memory.rs` - `am_alloc`, `am_free` safety
- `src/document.rs` - `am_init`, `am_save`, `am_load`, `am_merge`
- `src/text.rs` - `am_text_splice`, `am_get_text`
- `src/state.rs` - Global document state management

**Example**:
```bash
cd rust/automerge_wasi
cargo test
# Running unittests src/lib.rs (target/debug/deps/automerge_wasi-...)
# test result: ok. X passed; 0 failed; 0 ignored; 0 measured
```

---

## 🚀 3. Integration Test - Alice & Bob Scenario

**Location**: `testdata/integration/test_merge.sh`

**Run**: `./testdata/integration/test_merge.sh`

### What It Tests

**Scenario**: Two independent servers (Alice & Bob) make concurrent edits offline, then merge via CRDT.

**Steps**:
1. Start Alice's server (port 8080)
2. Start Bob's server (port 8081)
3. Alice types "Hello from Alice!"
4. Bob types "Hello from Bob!" (concurrent edit)
5. Download Alice's `doc.am` snapshot
6. Merge Alice's doc into Bob's via `/api/merge`
7. Verify CRDT properties:
   - ✅ Both docs saved as binary `.am` files
   - ✅ Automerge magic bytes `85 6f 4a 83` present
   - ⚠️ **KNOWN ISSUE**: Merge only preserves Bob's text

### Expected vs. Actual

**Expected CRDT behavior**:
```
Before merge:
  Alice: "Hello from Alice!"
  Bob:   "Hello from Bob!"

After merge (both should have):
  "Hello from Alice!Hello from Bob!"  (or other merged result)
```

**Actual behavior**:
```
After merge:
  Bob: "Hello from Bob!"  ← Only Bob's text preserved
```

### Known Issue

**Root Cause**: Under investigation

**Hypothesis**:
- Automerge 0.5 `merge()` API may need `apply_changes()` instead
- Or need to merge changes, not full documents
- Related to TestDocument_Merge (Go test also skipped)

**Tracking**: See `rust/automerge_wasi/src/document.rs:139`
```rust
/// ## Known Issues
/// - Currently only preserves one document's changes (needs investigation)
/// - See TestDocument_Merge in Go tests
```

---

## 🌐 4. End-to-End Tests (Playwright MCP)

**Tool**: Playwright MCP via Claude Code

**Location**: Manual testing session (no saved scripts yet)

**Configuration**: `.claude/settings.json` with 21 Playwright tools auto-approved

### Test Workflow

```bash
# 1. Start server
make run  # http://localhost:8080

# 2. Use Playwright MCP tools in Claude Code session
# - mcp__playwright__browser_navigate(url: "http://localhost:8080")
# - mcp__playwright__browser_snapshot()
# - mcp__playwright__browser_click(element: "textarea", ref: "...")
# - mcp__playwright__browser_type(text: "Test text", ref: "...")
# - mcp__playwright__browser_click(element: "Save Changes", ref: "...")
# - mcp__playwright__browser_take_screenshot(filename: "test.png")

# 3. Verify results
curl -s http://localhost:8080/api/text
# Should return: "Test text"
```

### What E2E Tests Cover

- ✅ Page loads without errors
- ✅ SSE connection (status badge shows "Connected" in green)
- ✅ Typing updates character counter
- ✅ Save button persists changes
- ✅ Clear button clears textarea
- ✅ Screenshot captures working state

### Playwright MCP Tools Available

**21 tools** (see [`mcp-playwright.md`](mcp-playwright.md) for details):
- Navigation: `browser_navigate`, `browser_navigate_back`
- Inspection: `browser_snapshot`, `browser_take_screenshot`, `browser_console_messages`
- Interaction: `browser_click`, `browser_type`, `browser_fill_form`, `browser_select_option`
- Control: `browser_close`, `browser_resize`, `browser_tabs`, `browser_wait_for`
- Others: `browser_evaluate`, `browser_hover`, `browser_drag`, etc.

---

## 🐛 5. Known Issues & Investigations

### Issue #1: CRDT Merge Not Preserving Both Edits

**Status**: 🔴 **CRITICAL** - Blocks distributed collaboration

**Affected**:
- Go test: `TestDocument_Merge` (skipped)
- Integration test: Alice & Bob scenario
- Rust: `am_merge` implementation

**Symptoms**:
```rust
// Alice's doc: "Hello from Alice!"
// Bob's doc:   "Hello from Bob!"
// After merge: "Hello from Bob!"  ← Only Bob's text preserved
```

**Investigation Steps**:

1. **Check Automerge 0.5 API usage**:
   ```rust
   // Current implementation (document.rs:157)
   let _ = doc.merge(&mut other_doc);
   ```
   - Is `merge()` the right API?
   - Should we use `apply_changes()` instead?
   - Does Automerge 0.5 have different merge semantics?

2. **Test with simple concurrent edits at different positions**:
   ```
   // Start with: "Hello"
   // Doc1 edit: prepend "Hi " → "Hi Hello"
   // Doc2 edit: append " World" → "Hello World"
   // After merge: should be "Hi Hello World"
   ```

3. **Verify merge commutativity**:
   ```rust
   merge(doc1, doc2) == merge(doc2, doc1)  // Should be true
   ```

4. **Check if we need to merge changes, not documents**:
   - Automerge has `get_changes()` and `apply_changes()`
   - May need to extract changes from other_doc and apply them

**Next Steps**:
- [ ] Review Automerge 0.5 docs for merge API
- [ ] Compare with Automerge 0.7 merge implementation
- [ ] Test with Automerge.js 3.1.2 merge behavior
- [ ] Add debug logging to see internal CRDT state
- [ ] Consider upgrading to Automerge 0.7 if API changed

---

## 📝 6. Test Coverage Matrix

| Feature | Go Test | Rust Test | Integration | E2E | Status |
|---------|---------|-----------|-------------|-----|--------|
| **Document Create** | ✅ | ✅ | ✅ | ✅ | Working |
| **Text Insert** | ✅ | ✅ | ✅ | ✅ | Working |
| **Text Delete** | ✅ | ✅ | ✅ | ✅ | Working |
| **Text Replace** | ✅ | ✅ | ✅ | ✅ | Working |
| **Unicode/Emoji** | ✅ | ✅ | N/A | ✅ | Working |
| **Save/Load** | ✅ | ✅ | ✅ | ✅ | Working |
| **CRDT Merge** | ⚠️ Skipped | ✅ Compiles | ⚠️ Partial | N/A | **Known Issue** |
| **SSE Broadcast** | N/A | N/A | N/A | ✅ | Working |
| **Persistence** | ✅ | ✅ | ✅ | ✅ | Working |

---

## 🚦 7. Running All Tests

### Quick Test

```bash
make test-go     # Go unit tests
make test-rust   # Rust unit tests
```

### Full Test Suite

```bash
# 1. Rust tests
cargo test
# Expected: ok. X passed; 0 failed; 0 ignored

# 2. Go tests
cd go/pkg/automerge && go test -v
# Expected: ok  	github.com/joeblew999/automerge-wazero-example/pkg/automerge	3.178s
# Note: 1 test skipped (TestDocument_Merge)

# 3. Integration test
./testdata/integration/test_merge.sh
# Expected: ✅ CRDT Merge Test Complete (with known issue warning)

# 4. E2E (manual)
make run
# Open http://localhost:8080
# Use Playwright MCP tools to test UI
```

### CI/CD

**Not yet implemented** - See [`CLAUDE.md`](../../CLAUDE.md) T3:

```yaml
# TODO: GitHub Actions CI
- build-wasi: cargo build --target wasm32-wasip1
- test-rust: cargo test
- test-go: go test -v ./...
- lint: golangci-lint, cargo clippy
```

---

## 📚 8. Related Documentation

- [`CLAUDE.md`](../../CLAUDE.md) - AI agent instructions (section 0.3: Testing Requirements)
- [`mcp-playwright.md`](mcp-playwright.md) - Playwright MCP usage guide
- [`cleanup-analysis.md`](../archive/cleanup-analysis.md) - Testing strategy clarification
- [`api-mapping.md`](../reference/api-mapping.md) - API coverage tracking
- [`TODO.md`](../../TODO.md) - Current tasks and known issues

---

## 🎯 9. Testing Principles

1. **NEVER ASSUME CODE WORKS** - All features must have tests
2. **Test from the outside** - Use Playwright MCP for E2E, not mocks
3. **Keep test data** - `testdata/snapshots/*.am` files are version controlled
4. **Document known issues** - Skip tests with clear comments, track in TODO.md
5. **Reproduce bugs as tests** - Before fixing, write failing test
6. **Test real CRDT properties**:
   - Commutativity: `merge(A, B) == merge(B, A)`
   - Convergence: Both sides converge to same state
   - No data loss: All edits preserved

---

## 🔧 10. Debugging Tests

### Go Test Debugging

```bash
cd go/pkg/automerge

# Run specific test
go test -v -run TestDocument_Merge

# With race detector
go test -race -v

# With coverage
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Rust Test Debugging

```bash
cd rust/automerge_wasi

# Show test output
cargo test -- --nocapture

# Run specific test
cargo test test_name

# With backtrace
RUST_BACKTRACE=1 cargo test
```

### Integration Test Debugging

```bash
# Check logs
cat /tmp/alice.log
cat /tmp/bob.log

# Inspect snapshots
hexdump -C data/alice/doc.am | head -5
hexdump -C data/bob/doc.am | head -5

# Manual merge test
curl -s http://localhost:8080/api/doc > alice.am
curl -s -X POST http://localhost:8081/api/merge \
  -H 'Content-Type: application/octet-stream' \
  --data-binary @alice.am
curl -s http://localhost:8081/api/text
```

---

**Last Updated**: 2025-10-20
**Status**: 11/12 Go tests passing, 1 known issue with CRDT merge
