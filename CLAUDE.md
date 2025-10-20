# CLAUDE.md — AI Agent Instructions for Automerge + WASI + wazero (Go)

> **Goal**: Run Automerge (Rust) as a **WASI** module hosted by **wazero** (Go), expose a minimal HTTP API + SSE for collaborative text editing, and provide a path to evolve toward **Automerge sync messages** and **NATS** transport.

This document instructs automation agents (and humans) how to build, run, extend, and test the project. Follow tasks in order unless stated otherwise.

---

NOTES for CLAUDE SETUP:

Claude

~/.claude.json

/Users/apple/.local/bin/claude

/Users/apple/.vscode/extensions/anthropic.claude-code-2.0.22-darwin-arm64/resources/native-binary/claude --dangerously-skip-permissions

/Users/apple/.vscode/extensions/anthropic.claude-code-2.0.22-darwin-arm64/resources/native-binary/claude --help

/Users/apple/.vscode/extensions/anthropic.claude-code-2.0.22-darwin-arm64/resources/native-binary/claude mcp list



## 0) Repository & Path Configuration

**Repository**: `joeblew999/automerge-wazero-example`

https://github.com/joeblew999/automerge-wazero-example

### ⚠️ CRITICAL: File Path Requirements

**CORRECT path** (3 nines in username):
```
/Users/apple/workspace/go/src/github.com/joeblew999/automerge-wazero-example
```

**INCORRECT path** (2 nines - DO NOT USE):
```
/Users/apple/workspace/go/src/github.com/joeblew99/automerge-wazero-example
```

Always use the **3-nines** version (`joeblew999`).

---

## 0.1) Stack Dependencies & Source Code Management

### Automerge (Rust CRDT Library)

**Primary**: https://github.com/automerge/automerge

**Version**: https://github.com/automerge/automerge/releases/tag/rust%2Fautomerge%400.7.0

**Docs**: https://github.com/automerge/automerge.github.io

**Requirements**:
- ✅ MUST keep a copy of Automerge **source code** in `.src/automerge/`
- ✅ MUST keep a copy of Automerge **docs** in `.src/automerge.github.io/`
- ✅ MUST understand the source and docs to use Automerge correctly
- ✅ Use `make setup-src` to clone, `make update-src` to update

**AI Agent Documentation Files** (keep these updated):

1. **`AGENT_AUTOMERGE.MD`** - For AI to understand Automerge concepts, CRDT behavior, and usage patterns
   - Purpose: High-level understanding of how Automerge works
   - Audience: AI agents learning to use Automerge effectively
   - Content: Concepts, best practices, common patterns

2. **`API_MAPPING.MD`** - Technical reference for Automerge API → WASI → Go mapping
   - Purpose: Complete API coverage matrix and implementation status
   - Audience: AI agents implementing features
   - Content: Every Rust method, corresponding WASI export, Go wrapper, implementation status

### Datastar (Go UI Framework)

**Primary**: https://github.com/starfederation/datastar-go

**Website**: https://data-star.dev

**Requirements**:
- ✅ MUST keep a copy of datastar-go in `.src/datastar-go/`
- ✅ MUST understand the docs to use Datastar correctly

**AI Agent Documentation File**:

3. **`AGENT_DATASTAR.MD`** - For AI to understand Datastar concepts and usage
   - Purpose: High-level understanding of Datastar for UI work
   - Audience: AI agents implementing UI features (M4+)
   - Content: Datastar patterns, SSE integration, reactive updates

---

## 0.2) 🔥 CODE SYNCHRONIZATION REQUIREMENTS 🔥

**CRITICAL**: The codebase has **4 layers** that MUST stay synchronized:

```
Layer 1: Automerge Rust Core (in .src/automerge/)
           ↓
Layer 2: WASI Exports (rust/automerge_wasi/src/*.rs)
           ↓
Layer 3: Go FFI Wrappers (go/pkg/wazero/exports.go)
           ↓
Layer 4: Go High-Level API (go/pkg/automerge/*.go)
```

### When You Change Go Code → Update Rust

**Rule**: Adding methods to `go/pkg/automerge/*.go` **REQUIRES**:

1. ✅ Corresponding WASI export(s) in `rust/automerge_wasi/src/*.rs`
2. ✅ FFI wrapper(s) in `go/pkg/wazero/exports.go`
3. ✅ Update `API_MAPPING.MD` with:
   - New Rust Automerge method (if applicable)
   - New WASI export signature
   - New Go wrapper
   - Implementation status (Implemented/Stub/Planned)
4. ✅ Tests for the new functionality

**Example Flow**:
```
1. Add method: go/pkg/automerge/map.go → func (d *Document) Put(...)
2. Add export: rust/automerge_wasi/src/map.rs → am_put(...)
3. Add wrapper: go/pkg/wazero/exports.go → func (r *Runtime) AmPut(...)
4. Update docs: API_MAPPING.MD → document the mapping
5. Add test: go/pkg/automerge/map_test.go → TestDocument_Put
```

### When You Change Rust Code → Update Go

**Rule**: Adding WASI exports in `rust/automerge_wasi/src/*.rs` **REQUIRES**:

1. ✅ FFI wrapper in `go/pkg/wazero/exports.go`
2. ✅ High-level method in `go/pkg/automerge/*.go`
3. ✅ Update `API_MAPPING.MD`
4. ✅ Tests

---

## 0.3) 🔄 UPSTREAM SOURCE SYNCHRONIZATION 🔄

**CRITICAL**: When Automerge upstream changes, we MUST update our code to stay in sync.

### The 5-Layer Dependency Chain

```
Layer 0: Automerge Upstream (.src/automerge/) ← WATCH THIS!
           ↓ (We track changes here)
Layer 1: Our Rust WASI Wrapper (rust/automerge_wasi/src/*.rs)
           ↓
Layer 2: Go FFI Wrappers (go/pkg/wazero/exports.go)
           ↓
Layer 3: Go High-Level API (go/pkg/automerge/*.go)
           ↓
Layer 4: Documentation (API_MAPPING.MD, AGENT_AUTOMERGE.MD)
```

### Version Tracking

| Component | Current Version | Tracked Version | Gap |
|-----------|----------------|-----------------|-----|
| **Automerge Rust (in use)** | 0.5 | 0.7.0 | ⚠️ 2 versions behind |
| **Automerge.js (tracked)** | N/A (not used) | 3.1.2 | Reference only |
| **Our WASI exports** | 11 functions | 65 planned | 17% complete |

**Gap Status**: We're using Automerge Rust 0.5 but tracking 0.7.0 source in `.src/automerge/`. Evaluate upgrade path before M2.

### ⚠️ CRITICAL: Client vs Server Automerge Usage

**Current State (M0)**: Server-side CRDT **ONLY**

| Layer | Automerge Usage | Version | Status |
|-------|----------------|---------|--------|
| **Browser (ui/ui.html)** | ❌ NOT LOADED | N/A | Removed in commit fixing JS errors |
| **Go Server (main.go)** | ✅ ACTIVE via WASM | Rust 0.5 | CRDT operations work |
| **Rust WASI Module** | ✅ ACTIVE | Rust 0.5 | Exports am_* functions |

**Why Removed from Browser**:
- Attempted to load `@automerge/automerge@3.1.2` via CDN
- **Error**: `TypeError: (void 0) is not a function` at WASM init
- **Root Cause**: Browser WASM loading incompatibility
- **Fix**: Removed import, all CRDT operations server-side only
- **Result**: UI now works (SSE, character counter, buttons all fixed)

**Version Alignment Requirements**:

When we **re-add** client-side Automerge.js (M2):

1. ✅ **Match tracked version**: Use `@automerge/automerge@3.1.2` (same as `.src/automerge/`)
2. ✅ **Verify WASM loading**: Test in browser console before deploying
3. ✅ **API alignment**: Ensure `Automerge.updateText()` exists in chosen version
4. ✅ **Server compatibility**: Client sync messages must be compatible with Rust 0.5 server

**Testing Client-Side Automerge.js** (before adding back):

```bash
# Test in browser console (manually)
# Visit: https://esm.sh/@automerge/automerge@3.1.2
# Check: Does it load without errors?
# Check: Does window.Automerge.updateText exist?

# Or test with simple HTML
cat > /tmp/test-automerge.html <<'EOF'
<script type="module">
import * as Automerge from 'https://esm.sh/@automerge/automerge@3.1.2';
console.log('Loaded:', Automerge);
console.log('updateText exists:', typeof Automerge.updateText);
let doc = Automerge.from({ text: "" });
console.log('Doc created:', doc);
</script>
EOF
open /tmp/test-automerge.html  # Check browser console
```

**Current Data Flow** (M0):

```
Browser → POST /api/text (full text) → Go Server → WASM am_text_splice() → CRDT
                                          ↓
Browser ← SSE /api/stream (full text) ← Broadcast ← CRDT state
```

**Future Data Flow** (M2):

```
Browser Automerge.updateText() → Sync Message → POST /api/sync → am_sync_recv()
                                                                      ↓
Browser ← SSE sync messages ← am_sync_gen() ← CRDT merge ← Server CRDT
```

### When `.src/automerge/` Changes → Update Our Code

**Rule**: When you update `.src/automerge/` (via `git pull` or version bump), you MUST:

#### Step 1: Check What Changed in JavaScript API

```bash
cd .src/automerge/javascript
git diff v3.1.2..v3.2.0 src/implementation.ts | grep "^+export function"
```

**Action**: For each new function:
1. Add stub to `go/pkg/automerge/<category>.go`
2. Return `NotImplementedError("Function added in Automerge v3.2.0 - planned for milestone MX")`
3. Update `API_MAPPING.MD` coverage matrix

#### Step 2: Check What Changed in Rust API

```bash
cd .src/automerge/rust/automerge
git diff rust/automerge@0.7.0..rust/automerge@0.8.0 src/
```

**Action**: For each API change:
1. Update `AGENT_AUTOMERGE.MD` if concepts changed
2. Update `API_MAPPING.MD` "Complete Automerge Rust API Reference" section
3. Add TODO comments in stubs: `// TODO: Automerge 0.8.0 changed signature to...`

#### Step 3: Check What Changed in Documentation

```bash
cd .src/automerge.github.io
git diff main..new-version content/docs/
```

**Action**:
1. Update `AGENT_AUTOMERGE.MD` with new concepts/best practices
2. Update examples in comments if API usage changed

### The 65-Function Contract

**Discovery**: Automerge.js has **exactly 65 exported functions** (as of v3.1.2).

**Our Go API**: Maintains **1:1 parity** with 65 methods (13 implemented, 52 stubs).

**Rule When Upstream Adds Functions**:

If Automerge.js v3.3.0 adds `splitDocument()`:
1. ✅ Add `func (d *Document) SplitDocument() error { return NotImplementedError("...") }`
2. ✅ Update count: 66 methods total (13 implemented, 53 stubs)
3. ✅ Update `API_MAPPING.MD` coverage: 11/66 = 16.7%
4. ✅ Keep tracking the ratio

### Function Count Verification (Run Regularly)

**Check if we're still in sync**:

```bash
# JavaScript API count
grep "^export function " .src/automerge/javascript/src/implementation.ts | wc -l
# Expected: 65 (as of v3.1.2)

# Our Go API count (should match!)
grep "^func (" go/pkg/automerge/*.go | wc -l
# Expected: 65

# Implemented count
grep -h "NotImplementedError\|DeprecatedError" go/pkg/automerge/*.go | wc -l
# Current: 52 stubs → 13 implemented

# WASI exports count
grep "^pub extern \"C\" fn am_" rust/automerge_wasi/src/*.rs | wc -l
# Current: 11 (M0 milestone)
```

**If counts don't match**: Upstream added functions! Follow Step 1 above.

### When to Update `.src/automerge/`

**Regular updates**:
```bash
cd .src/automerge
git pull origin main  # Get latest changes

# Check what changed
git log --oneline --since="1 month ago" -- javascript/src/implementation.ts
git log --oneline --since="1 month ago" -- rust/automerge/src/
```

**Before major milestones** (M1, M2):
1. Update `.src/automerge/` to latest stable release
2. Run function count verification (above)
3. Update stubs for new functions
4. Update `API_MAPPING.MD` and `AGENT_AUTOMERGE.MD`
5. Document version gap in CLAUDE.md (this file)

**Before Cargo.toml version bump**:
```bash
# We're upgrading from automerge 0.5 → 0.7
# 1. Check breaking changes
cd .src/automerge/rust/automerge
git log rust/automerge@0.5.0..rust/automerge@0.7.0 | grep -i "breaking"

# 2. Update our WASI wrapper for API changes
# 3. Run all tests
# 4. Update this section with new version numbers
```

### API Signature Tracking

**Before implementing a stub**, verify the signature matches upstream:

**Example: Implementing `GetHeads()`**

1. **Check TypeScript signature**:
```bash
grep -A5 "export function getHeads" .src/automerge/javascript/src/implementation.ts
# export function getHeads<T>(doc: Doc<T>): Heads
```

2. **Check Rust signature**:
```bash
rg "fn get_heads" .src/automerge/rust/automerge/src/
# pub fn get_heads(&mut self) -> Vec<ChangeHash>
```

3. **Design WASI export to match**:
```rust
// rust/automerge_wasi/src/history.rs
#[no_mangle]
pub extern "C" fn am_get_heads_count() -> u32 { ... }

#[no_mangle]
pub extern "C" fn am_get_heads(ptr_out: *mut u8) -> i32 { ... }
```

4. **Design Go API to match TypeScript**:
```go
// go/pkg/automerge/history.go
func (d *Document) GetHeads() ([]string, error) { ... }
```

**This ensures our API feels familiar to Automerge users!**

### Why This Matters

**Without upstream tracking**:
- ❌ We won't know when Automerge adds features we need
- ❌ Our stubs might not match real API signatures
- ❌ Version upgrades could break unexpectedly
- ❌ We'll miss bug fixes and improvements

**With upstream tracking**:
- ✅ Plan milestone features based on actual Automerge API
- ✅ Stubs are accurate placeholders with correct signatures
- ✅ Clear upgrade path when ready (0.5 → 0.7 → 0.8)
- ✅ Can cherry-pick features we need
- ✅ Stay compatible with Automerge ecosystem

### Verification Checklist (Run After API Changes)

After ANY changes to the API layer:

- [ ] Every Go method in `pkg/automerge/` has a clear path to WASI (or is marked as stub)
- [ ] Every WASI export in `rust/automerge_wasi/src/` has a Go wrapper in `pkg/wazero/`
- [ ] Every wrapper in `pkg/wazero/` is used by `pkg/automerge/`
- [ ] `API_MAPPING.MD` is updated with coverage status
- [ ] Tests verify the integration works
- [ ] `make build-wasi && make test-go` passes

---

## 0.3) Testing Requirements

**NEVER ASSUME CODE WORKS!**

All code MUST be tested before declaring completion.

### Test Tools & Requirements

1. **Playwright MCP** - MUST use for end-to-end browser testing
   - Test from the outside (user perspective)
   - **Screenshot Workflow**:
     - Playwright saves to `.playwright-mcp/testdata/screenshots/` (auto-created)
     - Copy test screenshots to `testdata/screenshots/` (for test artifacts)
     - Copy final screenshots to `screenshots/` (for README.md)
     - Use `cp .playwright-mcp/testdata/screenshots/name.png screenshots/screenshot.png`
   - Reference screenshots in `README.md` as `screenshots/screenshot.png`

2. **Go Tests** - `go test -v ./...`
   - Unit tests for each package
   - Integration tests for WASM FFI
   - Test data in `go/testdata/`

3. **Rust Tests** - `cargo test`
   - Unit tests in each module (`src/*.rs`)
   - Test WASI exports work correctly

### MCP Configuration for Testing

#### Global MCP Server Setup

**Location**: Playwright MCP must be configured in `~/.claude.json`

Add to the global `mcpServers` section:
```json
"playwright": {
  "type": "stdio",
  "command": "npx",
  "args": ["@playwright/mcp@latest"],
  "env": {}
}
```

**Verify MCP servers**:
```bash
/Users/apple/.local/bin/claude mcp list
# Should show: playwright: npx @playwright/mcp@latest - ✓ Connected
```

#### Project-Level Auto-Approval (REQUIRED for autonomous testing)

**Location**: `.claude/settings.json` (committed to repo)

```json
{
  "allowedTools": [
    "mcp__playwright__*"
  ]
}
```

**Why**: This auto-approves all Playwright MCP tools so AI agents can run end-to-end tests WITHOUT user prompts. Critical for autonomous testing workflows.

**Verify**:
```bash
# Test that Playwright tools work without prompts
# Agent should be able to call mcp__playwright__browser_navigate without asking
```

**Note**: If you have multiple Claude installations (standalone CLI + VSCode extension), they may use different configurations. The Playwright tools may not be available in the current session until Claude Code restarts to load the MCP server.

### Files to Keep Updated

- ✅ `Makefile` - All build and test targets
- ✅ `README.md` - User-facing documentation, screenshots
- ✅ `.gitignore` - Ignore build artifacts, keep test data
- ✅ `TODO.md` - Current tasks, completed work, next steps
  - **CRITICAL**: Keep TODO.md and code in sync!

---

## 0.4) Branching Strategy

* `main` — stable, protected
* `dev/*` — feature branches, merge via PR

---

## 0.5) Primary File Paths

```
/Makefile                              # Build automation
/README.md                             # User documentation
/TODO.md                               # Task tracking (MUST keep updated!)
/API_MAPPING.MD                        # API coverage matrix
/AGENT_AUTOMERGE.MD                    # AI: Automerge concepts
/AGENT_DATASTAR.MD                     # AI: Datastar concepts
/ui/ui.html                            # Browser UI
/go/cmd/server/main.go                 # HTTP server (should be slim!)
/go/pkg/automerge/*.go                 # High-level Go API
/go/pkg/wazero/*.go                    # Low-level FFI wrappers
/rust/automerge_wasi/Cargo.toml        # Rust WASI crate config
/rust/automerge_wasi/src/lib.rs        # Module orchestrator
/rust/automerge_wasi/src/memory.rs     # Memory management
/rust/automerge_wasi/src/document.rs   # Document lifecycle
/rust/automerge_wasi/src/text.rs       # Text CRDT operations
/.src/automerge/                       # Automerge source (reference)
/.src/automerge.github.io/             # Automerge docs (reference)
/.src/datastar-go/                     # Datastar source (reference)
```

---

## 1) Environment & Prerequisites

* **Rust** (stable): `rustup` installed
* **Targets**:
  * Currently using: `wasm32-wasip1` (Rust 1.84+)
  * Legacy: `wasm32-wasi` (pre-1.84)
  * Target configured in `Makefile`
* **Go**: 1.21+
* **Make**

### Local Bootstrap

```bash
make build-wasi   # builds rust → WASI .wasm
make run          # runs Go server with wazero
# open http://localhost:8080
```

**Build Artifacts**:

* `rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm` (~559KB)
* `doc.am` - Snapshot persisted in repo root (for demo)

---

## 2) Architecture (High-Level)

### Four-Layer Architecture

```
┌─────────────────────────────────────────┐
│  User (Browser)                         │
│  ui/ui.html                             │
└──────────────────┬──────────────────────┘
                   │ HTTP/SSE
┌──────────────────▼──────────────────────┐
│  Go Server (wazero host)                │
│  go/cmd/server/main.go                  │
│  - HTTP endpoints                       │
│  - SSE broadcasting                     │
│  - Document persistence                 │
└──────────────────┬──────────────────────┘
                   │ High-level API
┌──────────────────▼──────────────────────┐
│  Go API Layer (pkg/automerge)           │
│  - Document, Text, Map, List, etc.      │
│  - Type-safe, idiomatic Go              │
└──────────────────┬──────────────────────┘
                   │ FFI calls
┌──────────────────▼──────────────────────┐
│  Go FFI Layer (pkg/wazero)              │
│  - 1:1 WASI export wrappers             │
│  - Memory management                    │
└──────────────────┬──────────────────────┘
                   │ WASM calls
┌──────────────────▼──────────────────────┐
│  Rust WASI Layer (automerge_wasi)       │
│  - WASI exports (am_*)                  │
│  - Modules: memory, document, text      │
└──────────────────┬──────────────────────┘
                   │ Rust API calls
┌──────────────────▼──────────────────────┐
│  Automerge Rust Core                    │
│  - AutoCommit, ReadDoc, Transactable    │
│  - CRDT magic                           │
└─────────────────────────────────────────┘
```

### Component Details

* **Rust crate (`automerge_wasi`)**
  * Wraps Automerge core (`automerge` crate)
  * Exposes C-like ABI over WASI
  * Modular structure: memory, document, text, (future: map, list, sync)
  * Exports: `am_alloc`, `am_free`, `am_init`, `am_text_splice`, `am_save`, `am_load`, `am_merge`

* **Go server (wazero host)**
  * Instantiates WASI module
  * Holds one in-memory document (demo; M3 will support multi-doc)
  * HTTP endpoints: `GET /api/text`, `POST /api/text`, `GET /api/doc`, `POST /api/merge`
  * SSE at `GET /api/stream` for broadcasting updates
  * Persists `doc.am` and reloads on startup

* **UI**
  * `ui/ui.html`: textarea + SSE listener + Save button
  * Future (M4): Datastar-powered reactive UI

---

## 3) Tasks for Agents

### T1 — Ensure Repository Skeleton ✅ DONE

* [x] `Makefile`, `README.md`, `ui/ui.html`, `go/cmd/server/main.go`
* [x] `rust/automerge_wasi/{Cargo.toml, src/lib.rs}`
* [x] `go.mod` with `github.com/tetratelabs/wazero`
* [x] Compile & run: `make build-wasi && make run`

### T2 — Developer DX ✅ DONE

* [x] `make tidy` (runs `go mod tidy`)
* [x] `make test-go`, `make test-rust`
* [x] `make generate-test-data`
* [ ] Optional: file-watcher for hot-reload (e.g., `reflex`, `watchexec`)

### T3 — Quality Gates

* [ ] GitHub Actions CI: build WASI + Go server
* [ ] Lint: `golangci-lint` (Go), `cargo clippy` (Rust)

### T4 — Error Handling & Logging ✅ DONE

* [x] Map negative return codes in Rust to HTTP 4xx/5xx in Go
* [x] Error types: `NotImplementedError`, `DeprecatedError`, `WASMError`
* [x] Structured logging in Go (using std log)

### T5 — Persistence Policy ✅ DONE

* [x] Keep latest snapshot `doc.am`
* [ ] (Optional) Periodic snapshots + rotation

---

## 4) Exported WASI ABI (Current - M0)

### Memory Management

* `am_alloc(size: usize) -> *mut u8` — Allocate buffer in WASM memory
* `am_free(ptr: *mut u8, size: usize)` — Free allocated buffer

### Document Lifecycle

* `am_init() -> i32` — Initialize new document with Text at ROOT["content"]
* `am_save_len() -> u32` — Get serialized document size
* `am_save(ptr_out: *mut u8) -> i32` — Save document to buffer
* `am_load(ptr: *const u8, len: usize) -> i32` — Load document from buffer
* `am_merge(other_ptr: *const u8, other_len: usize) -> i32` — Merge documents

### Text Operations

* `am_text_splice(pos: usize, del: i64, insert_ptr: *const u8, insert_len: usize) -> i32` — CRDT text splice
* `am_set_text(ptr: *const u8, len: usize) -> i32` — Replace entire text (DEPRECATED)
* `am_get_text_len() -> u32` — Get text length in bytes
* `am_get_text(ptr_out: *mut u8) -> i32` — Copy text to buffer

**Return codes**: `0` = success; `<0` = error code

**Module Structure** (rust/automerge_wasi/src/):
- `lib.rs` - Module orchestration
- `memory.rs` - `am_alloc`, `am_free`
- `document.rs` - `am_init`, `am_save`, `am_load`, `am_merge`
- `text.rs` - `am_text_splice`, `am_get_text`, etc.
- `state.rs` - Global document state management

---

## 5) HTTP API (Demo)

* `GET /api/text` → `200 text/plain` returns current text
* `POST /api/text` `{"text": string}` → `204 No Content`; broadcasts SSE `update`
* `GET /api/stream` → SSE with events:
  * `snapshot` on connect: `{ "text": string }`
  * `update` on edits: `{ "text": string }`
* `GET /api/doc` → Download `doc.am` snapshot
* `POST /api/merge` → Merge another `doc.am` (CRDT merge)
* `GET /` → Serve `ui/ui.html`

---

## 6) Roadmap / Next Milestones

### M1 — **Automerge Sync Protocol** (delta-based)

**Why**: Avoid shipping whole text; support true peer-to-peer sync.

**Add to Rust** (`rust/automerge_wasi/src/sync.rs`):

* [ ] `am_sync_init_peer(peer_id_ptr, len) -> i32`
* [ ] `am_sync_gen_len() -> u32`
* [ ] `am_sync_gen(ptr_out: *mut u8) -> i32`
* [ ] `am_sync_recv(ptr: *const u8, len: usize) -> i32`

**Update Go**:

* [ ] On local edit, call `am_sync_gen` and broadcast bytes (SSE or NATS)
* [ ] On receive, call `am_sync_recv` then maybe `am_sync_gen` (Automerge may request reply)
* [ ] Add `/api/sync` SSE channel or reuse `/api/stream` with `event: sync`

**Update Documentation**:

* [ ] Add M1 exports to `API_MAPPING.MD`
* [ ] Update `AGENT_AUTOMERGE.MD` with sync protocol concepts

### M2 — **Multi-Object Support** (Maps, Lists, Counters)

**Why**: Support full Automerge data model (not just single text field).

**Add to Rust**:

* [ ] `rust/automerge_wasi/src/map.rs`:
  * `am_get(path_ptr, path_len, key_ptr, key_len, value_out_ptr) -> i32`
  * `am_put(path_ptr, path_len, key_ptr, key_len, value_ptr, value_len) -> i32`
  * `am_delete(path_ptr, path_len, key_ptr, key_len) -> i32`
  * `am_keys(path_ptr, path_len, keys_out_ptr) -> i32`

* [ ] `rust/automerge_wasi/src/list.rs`:
  * `am_insert(path_ptr, path_len, index: u32, value_ptr, value_len) -> i32`
  * `am_remove(path_ptr, path_len, index: u32) -> i32`
  * `am_splice_list(path_ptr, path_len, pos: u32, del: i64, values_ptr, values_len) -> i32`

* [ ] `rust/automerge_wasi/src/counter.rs`:
  * `am_increment(path_ptr, path_len, key_ptr, key_len, delta: i64) -> i32`

**Update Go**:

* [ ] Implement `pkg/automerge/map.go` (remove stubs)
* [ ] Implement `pkg/automerge/list.go` (remove stubs)
* [ ] Implement `pkg/automerge/counter.go` (remove stubs)
* [ ] FFI wrappers in `pkg/wazero/exports.go`

**Multi-Document Support**:

* [ ] Replace single `DOC` with map keyed by `docId`
* [ ] Expose `am_select(doc_id_ptr, len)` / `am_new_doc(doc_id_ptr, len)`
* [ ] Query param `?doc=<id>` on HTTP routes
* [ ] Snapshot files `data/<docId>.am`

### M3 — **NATS Transport**

**Why**: Production-ready pub/sub, object storage, multi-tenant.

* [ ] Subjects: `automerge.sync.<tenant>.<docId>`
* [ ] Server acts as peer: on msg → `am_sync_recv` → maybe `am_sync_gen`
* [ ] Store snapshots in **NATS Object Store**
* [ ] Latest head in KV per `docId`
* [ ] RBAC via JWT; namespace subjects per tenant/region

### M4 — **Datastar UI** (Reactive Frontend)

**Why**: Modern reactive UI without complex JS frameworks.

* [ ] Browser: minimal JS streaming sync messages via SSE
* [ ] Datastar "action" hooks to send local ops
* [ ] Apply remote updates reactively
* [ ] Reference `AGENT_DATASTAR.MD` for implementation
* [ ] Optional WASM-Go frontends calling HTTP or NATS

### M5 — **Observability & Ops**

* [ ] Metrics: flush counts, message sizes, per-doc peers
* [ ] Tracing hooks around sync cycles
* [ ] Config flags for runtime paths and limits

---

## 7) Conventions & Guardrails

**Commits**: Conventional Commits (`feat:`, `fix:`, `chore:`, etc.)

**PRs**: Small, reviewed, CI green. Include:
* Scope, rationale
* Testing notes
* Backward-compatibility considerations

**Code Style**:
* Go: `gofmt` / `go vet`
* Rust: `cargo fmt` / `cargo clippy`

**Security**:
* Validate payload sizes; cap `am_alloc` usage
* UTF-8 validation (already done in Rust)
* Add content-length bounds in HTTP

**Performance**:
* Single module instance OK for demo
* Production: consider per-doc sharding or doc pool
* Avoid excessive `alloc/free`; measure with pprof

---

## 8) Testing Plan

### Unit Tests

**Rust** (`cargo test`):
* Construct doc, set text, save/load, compare
* Each module has its own tests
* See `src/memory.rs`, `src/document.rs`, `src/text.rs`

**Go** (`go test -v ./...`):
* Package tests in `pkg/automerge/*_test.go`
* In-memory server tests (future)
* Current: 11/12 tests passing (1 skipped for merge investigation)

### Integration Tests

* Start server → connect two SSE clients → POST update → assert second client receives `update`
* Test data in `go/testdata/snapshots/`

### End-to-End Tests (Playwright MCP)

**REQUIRED** before marking features complete. Test from browser perspective and capture screenshots for README.md.

**Workflow**:

```bash
# 1. Start server
make run  # or in background

# 2. Use Playwright MCP tools (available in Claude Code session)
# Example test flow:
# - mcp__playwright__browser_navigate(url: "http://localhost:8080")
# - mcp__playwright__browser_snapshot() # Verify page loaded
# - mcp__playwright__browser_click(element: "textarea", ref: "...")
# - mcp__playwright__browser_type(text: "Test CRDT sync", ref: "...")
# - mcp__playwright__browser_click(element: "Save Changes", ref: "...")
# - mcp__playwright__browser_take_screenshot(filename: "testdata/screenshots/test-save.png")
# - mcp__playwright__browser_close()

# 3. Copy screenshots to appropriate folders
# For test artifacts:
cp .playwright-mcp/testdata/screenshots/*.png testdata/screenshots/

# For README.md (pick the best screenshot):
cp .playwright-mcp/testdata/screenshots/best-shot.png screenshots/screenshot.png

# 4. Update README.md if screenshot changed significantly
```

**Test Checklist**:
- [ ] Page loads without errors
- [ ] SSE connects (status badge shows "Connected" in green)
- [ ] Typing updates character counter
- [ ] Save button persists changes (verify with `/api/text`)
- [ ] Clear button clears textarea
- [ ] Screenshot captures working state
- [ ] Screenshot copied to `screenshots/` for README

### CLI Smoke Tests

```bash
curl -s http://localhost:8080/api/text
curl -s -X POST http://localhost:8080/api/text \
  -H 'content-type: application/json' \
  -d '{"text":"Hello"}' -i
curl -s http://localhost:8080/api/stream  # observe snapshot + updates
```

---

## 9) Future Extensions (Optional)

* CRDT rich text or multiple fields (not just `content`)
* Heads/hash exposure for time travel
* Snapshot compaction/GC strategy
* E2E encryption of sync messages (application layer)
* Rollups to object store per interval

---

## 10) Quick Checklist (Copy/Paste for PRs)

```markdown
- [ ] Builds: `make build-wasi` ✅
- [ ] Tests: `make test-go` ✅
- [ ] Tests: `make test-rust` ✅
- [ ] Runs: `make run` → `GET /api/text` works ✅
- [ ] SSE: two tabs receive `snapshot`/`update` ✅
- [ ] Snapshot persists and reloads ✅
- [ ] Updated: `API_MAPPING.MD` ✅
- [ ] Updated: `TODO.md` ✅
- [ ] Updated: `README.md` (if needed) ✅
- [ ] Playwright tests pass ✅
- [ ] CI green ✅
```

---

## 11) Known Issues & Investigations

### CRDT Merge Behavior

**Status**: Needs investigation

**Issue**: `am_merge` currently only preserves one document's changes instead of merging both concurrent edits.

**Test**: `TestDocument_Merge` (currently skipped)

**Next Steps**:
1. Investigate Automerge 0.5 `merge()` vs `apply_changes()` APIs
2. Test with simple concurrent edits at different positions
3. Verify merge commutativity

---

**Contact / Owner**: @joeblew999
