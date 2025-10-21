# CLAUDE.md — AI Agent Instructions for Automerge + WASI + wazero (Go)

> **Goal**: Run Automerge (Rust) as a **WASI** module hosted by **wazero** (Go), expose a minimal HTTP API + SSE for collaborative text editing, and provide a path to evolve toward **Automerge sync messages** and **NATS** transport.

This document provides **essential instructions** for AI agents. For detailed explanations, see **[Documentation Index](docs/README.md)**.

## ✅ CURRENT STATUS: M0/M1/M2 COMPLETE

**Date**: 2025-10-21
**Test Status**: 83/83 tests passing (100%)
**Production Ready**: YES (for M0/M1/M2 features)

**Completed Milestones**:
- ✅ **M0**: Core CRDT (Text, Map, List, Counter, History, Document) - ALL DONE
- ✅ **M1**: Sync Protocol (per-peer state, binary messages) - ALL DONE
- ✅ **M2**: Rich Text Marks (CRDT-aware formatting) - ALL DONE

**System Health**:
- ✅ 10/10 modules with perfect 1:1 file mapping across all 6 layers
- ✅ 83 automated tests, 100% passing
- ✅ 23 HTTP endpoints, all functional
- ✅ Web UI with components for M0/M1/M2
- ✅ Integration testing strategy across WASM boundary

**See**: [STATUS.md](STATUS.md) for complete details

---

## 0) Repository & Path Configuration

**Repository**: `joeblew999/automerge-wazero-example`
**URL**: https://github.com/joeblew999/automerge-wazero-example

### ⚠️ CRITICAL: File Path (3 nines!)

```
/Users/apple/workspace/go/src/github.com/joeblew999/automerge-wazero-example
```

Always use **`joeblew999`** (3 nines), not `joeblew99` (2 nines).

---

## 0.1) Stack Dependencies

### Automerge (Rust CRDT Library)

- **Source**: `.src/automerge/` (v0.7.0 reference, using v0.5 in production)
- **Docs**: `.src/automerge.github.io/`
- **Setup**: `make setup-src` to clone, `make update-src` to update

**AI Agent Documentation**:
1. **[Automerge Guide](docs/ai-agents/automerge-guide.md)** - CRDT concepts, patterns, best practices
2. **[API Mapping](docs/reference/api-mapping.md)** - Complete API coverage matrix

### Datastar (Go UI Framework) - M4

- **Source**: `.src/datastar-go/`
- **Guide**: [docs/ai-agents/datastar-guide.md](docs/ai-agents/datastar-guide.md) (placeholder for M4)

---

## 0.2) 🔥 CODE SYNCHRONIZATION REQUIREMENTS

**CRITICAL**: 6-layer architecture - ALL layers must stay synchronized!

```
Layer 1: Automerge Rust Core (.src/automerge/)
           ↓
Layer 2: Rust WASI Exports (rust/automerge_wasi/src/<module>.rs)
           ↓
Layer 3: Go FFI Wrappers (go/pkg/wazero/<module>.go - 1:1 with Layer 2)
           ↓
Layer 4: Go High-Level API (go/pkg/automerge/<module>.go - Pure CRDT)
           ↓
Layer 5: Go Server Layer (go/pkg/server/<module>.go - Stateful + Thread-safe)
           ↓
Layer 6: Go HTTP API (go/pkg/api/<module>.go - HTTP handlers)
```

### 🎯 Perfect 1:1 File Mapping Across ALL Layers ✅

**Core CRDT Modules (10/10)**:

| Rust Module | Go FFI | Go API | Go Server | Go HTTP | Purpose |
|-------------|--------|--------|-----------|---------|---------|
| state.rs | state.go | - | server.go | - | Global state |
| memory.rs | memory.go | - | - | - | Memory allocation |
| document.rs | document.go | document.go | document.go | - | Lifecycle/Save/Load |
| text.rs | text.go | text.go | text.go | text.go | Text CRDT |
| map.rs | map.go | map.go | map.go | map.go | Map CRDT |
| list.rs | list.go | list.go | list.go | list.go | List CRDT |
| counter.rs | counter.go | counter.go | counter.go | counter.go | Counter CRDT |
| history.rs | history.go | history.go | history.go | history.go | Version control |
| sync.rs | sync.go | sync.go | sync.go | sync.go | Sync protocol (M1) |
| richtext.rs | richtext.go | richtext.go | richtext.go | richtext.go | Rich text (M2) |

**Additional Server Modules**:

| File | Purpose | Layer |
|------|---------|-------|
| server/broadcast.go | SSE client management | Layer 5 only |
| api/util.go | HTTP helper functions (parsePathString) | Layer 6 only |

### Layer Responsibilities

**Layer 4 (pkg/automerge)**: Pure CRDT operations
- Takes `context.Context` and `*Runtime` directly
- No state, no mutex, no persistence
- Example: `func (d *Document) Put(ctx, path, key, value) error`

**Layer 5 (pkg/server)**: Stateful + Thread-safe
- Owns `*automerge.Document` and `sync.RWMutex`
- Handles persistence (saveDocument after mutations)
- Manages SSE broadcast to clients
- Example: `func (s *Server) PutMapValue(ctx, path, key, value) error { s.mu.Lock(); defer s.mu.Unlock(); ... }`

**Layer 6 (pkg/api)**: HTTP protocol
- Parses HTTP requests (JSON body, query params)
- Calls server methods
- Formats HTTP responses (JSON, status codes)
- Example: `func MapHandler(srv *server.Server) http.HandlerFunc`

### Why Separate Layers 5 & 6?

**DON'T combine pkg/server and pkg/api!**

| Concern | pkg/server | pkg/api |
|---------|------------|---------|
| **State** | Owns document, mutex, clients | Stateless |
| **Protocol** | Go functions | HTTP JSON/query params |
| **Reusability** | Can be used by HTTP, gRPC, CLI | HTTP-specific |
| **Testing** | Unit tests with direct calls | HTTP integration tests |
| **Thread Safety** | Manages mutex | Delegates to server |
| **Persistence** | Calls saveDocument() | No knowledge of persistence |

This separation enables **protocol flexibility** - we could add gRPC handlers or a CLI tool that calls `pkg/server` directly without duplicating business logic.

### When You Change Go Code → Update Rust

**Rule**: Adding methods to `go/pkg/automerge/*.go` **REQUIRES**:

1. ✅ Corresponding WASI export(s) in `rust/automerge_wasi/src/<module>.rs`
2. ✅ FFI wrapper(s) in `go/pkg/wazero/<module>.go` (matching filename!)
3. ✅ Update `docs/reference/api-mapping.md` with coverage status
4. ✅ Tests for the new functionality

**Example Flow**:
```
1. Add method: go/pkg/automerge/map.go → func (d *Document) Put(...)
2. Add export: rust/automerge_wasi/src/map.rs → am_put(...)
3. Add wrapper: go/pkg/wazero/map.go → func (r *Runtime) AmPut(...)
4. Update docs: docs/reference/api-mapping.md → document the mapping
5. Add test: go/pkg/automerge/map_test.go → TestDocument_Put
```

### When You Change Rust Code → Update Go

**Rule**: Adding WASI exports in `rust/automerge_wasi/src/<module>.rs` **REQUIRES**:

1. ✅ FFI wrapper in corresponding `go/pkg/wazero/<module>.go` (same filename!)
2. ✅ High-level method in `go/pkg/automerge/<module>.go`
3. ✅ Update `docs/reference/api-mapping.md`
4. ✅ Tests

### Verification Checklist

After ANY changes to the API layer:

- [ ] Every Go method in `pkg/automerge/` has a clear path to WASI (or is marked as stub)
- [ ] Every WASI export in `rust/automerge_wasi/src/` has a Go wrapper in `pkg/wazero/`
- [ ] Every wrapper in `pkg/wazero/` is used by `pkg/automerge/`
- [ ] `docs/reference/api-mapping.md` is updated with coverage status
- [ ] Tests verify the integration works
- [ ] `make build-wasi && make test-go` passes

---

## 0.3) 📋 DOCUMENTATION PRINCIPLES - SINGLE SOURCE OF TRUTH

**CRITICAL**: Follow these to prevent documentation drift and broken links.

### Structure

```
/
├── README.md           # User entry point
├── CLAUDE.md           # AI agent instructions (this file)
├── TODO.md             # Active task tracking
└── docs/               # ALL other documentation (Diátaxis framework)
    ├── README.md       # Documentation index
    ├── tutorials/      # Learning-oriented
    ├── how-to/         # Goal-oriented recipes
    ├── reference/      # Information lookup
    ├── explanation/    # Understanding concepts
    ├── development/    # Developer workflow
    ├── ai-agents/      # AI-specific guides
    └── archive/        # Historical docs
```

See **[docs/README.md](docs/README.md)** for complete documentation index.

### Before Moving/Renaming Files

```bash
# 1. Find ALL references
grep -r "FILENAME.md" . --include="*.md" --include="*.go" --include="*.rs"

# 2. Move file
git mv OLD.md NEW.md

# 3. Update ALL references
# 4. Verify links work
make verify-docs

# 5. Commit together
git commit -m "docs: move FILENAME.md"
```

### After Any Documentation Changes

**ALWAYS run**:
```bash
make verify-docs  # Checks for broken internal markdown links
```

**Workflow**:
```bash
# Before committing docs
make verify-docs && git add docs/ *.md && git commit
```

---

## 0.4) Testing Requirements & Strategy

**NEVER ASSUME CODE WORKS!** All code MUST be tested.

### Testing Philosophy: Integration Over Unit

**✅ WE USE INTEGRATION TESTING** across the WASM boundary. This is intentional and correct.

**Why Integration Tests?**
1. **WASM boundary is expensive** - Don't want unit tests for every FFI call
2. **Real-world coverage** - Tests verify complete stack works together
3. **Catches FFI bugs** - Memory management, pointer errors surface immediately
4. **Less maintenance** - No need to mock WASM calls
5. **Already comprehensive** - 83 tests covering M0, M1, M2

### Test Coverage by Layer

| Layer | Tests | Type | Status |
|-------|-------|------|--------|
| **Rust WASI** (rust/automerge_wasi) | 28 tests | Unit | ✅ 100% PASS |
| **Go FFI** (pkg/wazero) | 0 explicit | Tested via automerge | ✅ Covered |
| **Go API** (pkg/automerge) | 48 tests | **Integration** | ✅ 100% PASS |
| **Go Server** (pkg/server) | 0 explicit | Tested via api | ✅ Covered |
| **HTTP API** (pkg/api) | 7 tests | **Integration** | ✅ 100% PASS |
| **Web UI** | Manual + Playwright | E2E | ✅ Verified |

**Total: 83 automated tests, 100% passing** 🎉

### Test Workflow

```bash
# Build + run all tests
make build-wasi
make test           # Runs test-rust + test-go
make test-rust      # 28 Rust unit tests
make test-go        # 55 Go integration tests (48 automerge + 7 api)
```

### Milestone Test Coverage

**M0 - Core CRDT Operations** ✅
- Text: 3 test suites (splice, unicode, length)
- Map: 9 tests (put, get, delete, keys, nested paths)
- List: 4 tests (push, insert, get, delete)
- Counter: 3 tests (increment, decrement, get)
- History: 5 tests (heads, changes, load snapshots)
- Document: 12 tests (save, load, merge, lifecycle)

**M1 - Sync Protocol** ✅
- Sync: 4 tests (init state, generate message, receive message, two-peer sync)
- HTTP: 1 test (POST /api/sync)

**M2 - Rich Text Marks** ✅
- RichText: 8 tests (mark, unmark, get marks, overlapping marks)
- HTTP: 1 test (POST /api/richtext/mark, GET /api/richtext/marks)

### End-to-End Testing (Playwright MCP)

**REQUIRED** before marking features complete.

See **[Testing Guide](docs/development/testing.md)** for:
- Unit test strategies
- Integration test patterns
- Playwright MCP usage
- Test data generation

**Playwright MCP Testing Workflow**:
```bash
# 1. Ensure Playwright MCP is configured (in ~/.claude.json)
/Users/apple/.local/bin/claude mcp list  # Should show playwright

# 2. Ensure auto-approval (in .claude/settings.json)
# All 21 Playwright tools must be in allowedTools list

# 3. Start server
make run &

# 4. Use Playwright MCP tools to test
# - mcp__playwright__browser_navigate(url: "http://localhost:8080")
# - mcp__playwright__browser_snapshot()
# - mcp__playwright__browser_evaluate() # Run JavaScript
# - mcp__playwright__browser_take_screenshot()

# 5. Test plans in tests/playwright/
# - M1_SYNC_TEST_PLAN.md
# - M2_RICHTEXT_TEST_PLAN.md
```

**1:1 Mapping for Tests**:
```
tests/playwright/
├── M1_SYNC_TEST_PLAN.md       # Maps to go/pkg/api/sync.go, web/js/sync.js
└── M2_RICHTEXT_TEST_PLAN.md   # Maps to go/pkg/api/richtext.go, web/js/richtext.js
```

**Makefile Targets**:
```bash
make test-http        # Test HTTP endpoints (curl-based, requires server running)
make test-playwright  # Show Playwright test info
make verify-web       # Verify web folder structure
```

---

## 0.4.1) Building Automerge.js from Source

**CRITICAL**: We build our own Automerge.js from the same source as our Rust WASI!

### Why Build from Source?

- ✅ **Version alignment**: Rust backend and JS frontend use identical Automerge version
- ✅ **Single source of truth**: `.src/automerge/` contains both Rust and JS
- ✅ **Custom builds**: Can create slim/fat builds, IIFE/ESM formats
- ✅ **Debugging**: Full source maps, ability to patch if needed

### Build Process

```bash
# 1. Setup source (first time only)
make setup-src              # Clones .src/automerge/ (rust/automerge@0.7.0)

# 2. Install Rust WASM toolchain
make setup-rust-wasm        # Installs wasm32-unknown-unknown + wasm-bindgen

# 3. Build Automerge.js
make build-js               # Builds .src/automerge/javascript/ → ui/vendor/automerge.js
```

### Build Output

```
.src/automerge/javascript/dist/cjs/iife.cjs  # Built IIFE bundle
         ↓ (copied by make build-js)
ui/vendor/automerge.js                       # 3.4M IIFE format
```

### Usage in Web

**Old UI** (`ui/ui.html`):
```html
<script src="/vendor/automerge.js"></script>
<script>
  console.log('Automerge loaded:', typeof window.Automerge);
</script>
```

**New Web Folder** (`web/index.html`):
```html
<script src="/vendor/automerge.js"></script>
<script type="module" src="/web/js/app.js"></script>
```

**Served by Go**:
```go
// go/cmd/server/main.go
http.Handle("/vendor/", api.VendorHandler(staticCfg))  // Serves ui/vendor/
```

### Version Tracking

```bash
make sync-versions   # Verify all components use same .src/ version
```

**Output**:
```
📌 .src/automerge git version: rust/automerge@0.7.0
🦀 Cargo.toml dependency: automerge = { path = "../../.src/automerge/rust/automerge" }
📦 JavaScript package.json: "version": "3.2.0-alpha.0"
✅ Built Automerge.js: 3.4M
```

**Verification**:
```bash
make verify-web  # Checks that web/index.html references /vendor/automerge.js
```

---

## 0.4.2) Web Folder Structure (1:1 Mapping)

**NEW**: The `web/` folder follows the same 1:1 file mapping as the rest of the codebase.

### Architecture

```
web/
├── index.html          # Main entry point with tab navigation
├── css/
│   └── main.css        # Shared styles (600+ lines, gradient UI)
├── js/                 # 1:1 with go/pkg/automerge/*.go
│   ├── app.js          # Orchestrator (tab switching, SSE, init)
│   ├── text.js         # Maps to text.go (M0)
│   ├── map.js          # Maps to map.go (M0) - TODO
│   ├── list.js         # Maps to list.go (M0) - TODO
│   ├── counter.js      # Maps to counter.go (M0) - TODO
│   ├── history.js      # Maps to history.go (M0) - TODO
│   ├── sync.js         # Maps to sync.go (M1) ✅ COMPLETE
│   └── richtext.js     # Maps to richtext.go (M2) ✅ COMPLETE
└── components/         # 1:1 with go/pkg/api/*.go
    ├── text.html       # Maps to api/text.go (M0)
    ├── sync.html       # Maps to api/sync.go (M1) ✅ COMPLETE
    └── richtext.html   # Maps to api/richtext.go (M2) ✅ COMPLETE
```

### 1:1 Mapping Table

| Go API Handler | Web Component | Web JS Module | Status |
|----------------|---------------|---------------|--------|
| api/text.go | text.html | text.js | ✅ M0 |
| api/map.go | map.html | map.js | 🚧 TODO |
| api/list.go | list.html | list.js | 🚧 TODO |
| api/counter.go | counter.html | counter.js | 🚧 TODO |
| api/history.go | history.html | history.js | 🚧 TODO |
| api/sync.go | sync.html | sync.js | ✅ M1 |
| api/richtext.go | richtext.html | richtext.js | ✅ M2 |

### Adding New Web Components

When creating a new web component, maintain 1:1 mapping:

**Example: Adding Map component**

1. Create `web/components/map.html` (UI template)
2. Create `web/js/map.js` (exports `class MapComponent`)
3. Update `web/js/app.js` to import and initialize
4. Update `Makefile` variables:
   ```makefile
   WEB_JS = ... $(WEB_DIR)/js/map.js
   WEB_COMPONENTS = ... $(WEB_DIR)/components/map.html
   ```
5. Run `make verify-web` to ensure structure is correct

### Web Module Pattern

**Every `web/js/*.js` file exports a class**:

```javascript
// web/js/sync.js (M1 example)
export class SyncComponent {
    constructor() {
        this.peerID = null;
    }

    init() {
        // Setup event listeners
        // Initialize UI
    }

    async initSync() {
        // Call /api/sync endpoint
    }

    destroy() {
        // Cleanup when switching tabs
    }
}
```

**Orchestrated by app.js**:

```javascript
// web/js/app.js
import { SyncComponent } from './sync.js';

class App {
    constructor() {
        this.components = {
            sync: new SyncComponent(),
            // ...
        };
    }

    switchTab(tabName) {
        this.components[tabName].init();
    }
}
```

### Verification

```bash
make verify-web
```

**Output**:
```
🔍 Verifying web folder structure (1:1 mapping)...

Checking required files:
  ✅ web/index.html
  ✅ web/css/main.css
  ✅ web/js/app.js
  ✅ web/js/text.js
  ✅ web/js/sync.js
  ✅ web/js/richtext.js
  ✅ web/components/text.html
  ✅ web/components/sync.html
  ✅ web/components/richtext.html
  ✅ ui/vendor/automerge.js

Checking Automerge.js:
  ✅ ui/vendor/automerge.js (3.4M)
  ✅ web/index.html references /vendor/automerge.js

✅ Web folder structure valid!
```

---

## 0.5) Primary File Paths (Quick Reference)

```
/Makefile                              # Build automation + verify-web + test-http
/README.md                             # User documentation
/TODO.md                               # Task tracking
/FINAL_SUMMARY.md                      # Complete session summary (M0, M1, M2 complete)
/docs/reference/api-mapping.md         # API coverage matrix
/docs/ai-agents/automerge-guide.md     # AI: Automerge concepts
/docs/development/testing.md           # Testing guide
/docs/development/roadmap.md           # Milestones M0-M5
/docs/explanation/architecture.md      # 4-layer architecture deep dive

# Web UI (1:1 mapping)
/web/index.html                        # Main entry (tab navigation)
/web/css/main.css                      # Shared styles
/web/js/app.js                         # Orchestrator
/web/js/*.js                           # Component modules (1:1 with automerge/*.go)
/web/components/*.html                 # Component templates (1:1 with api/*.go)
/ui/ui.html                            # Old UI (prototype)
/ui/vendor/automerge.js                # Built from .src/ (3.4M, IIFE format)

# Go server
/go/cmd/server/main.go                 # HTTP server (23 routes)
/go/pkg/automerge/*.go                 # High-level Go API (1:1 with Rust)
/go/pkg/server/*.go                    # Server layer (1:1 with automerge)
/go/pkg/api/*.go                       # HTTP handlers (1:1 with automerge)
/go/pkg/wazero/*.go                    # Low-level FFI wrappers (1:1 with Rust)
/go/testdata/                          # All test data (unit/integration)

# Rust WASI
/rust/automerge_wasi/Cargo.toml        # Rust WASI crate config
/rust/automerge_wasi/src/lib.rs        # Module orchestrator
/rust/automerge_wasi/src/*.rs          # WASI modules (1:1 with Go)

# Source reference
/.src/automerge/                       # Automerge source (Rust + JS, v0.7.0)
/.src/automerge.github.io/             # Automerge docs

# Testing
/tests/playwright/M1_SYNC_TEST_PLAN.md      # M1 Playwright test plan
/tests/playwright/M2_RICHTEXT_TEST_PLAN.md  # M2 Playwright test plan
/screenshots/                          # UI screenshots for README
```

---

## 1) Environment & Prerequisites

- **Rust** (stable): `rustup` installed
- **Target**: `wasm32-wasip1` (Rust 1.84+)
- **Go**: 1.21+
- **Make**

### Quick Start

```bash
make build-wasi   # Build Rust → WASI .wasm
make run          # Run Go server with wazero
# Open http://localhost:8080
```

---

## 2) Architecture Quick Reference

**4-Layer Design**: See full details in **[Architecture Guide](docs/explanation/architecture.md)**

```
Browser (ui/ui.html)
    ↓ HTTP/SSE
Go Server (main.go) + wazero runtime
    ↓ FFI calls
Go FFI Layer (pkg/wazero/*.go)
    ↓ WASM calls
Rust WASI Layer (automerge_wasi/src/*.rs)
    ↓ Rust API
Automerge Core (CRDT magic)
```

**Key Points**:
- Rust compiled to WASM (`wasm32-wasip1`)
- Go loads WASM via wazero
- HTTP + SSE for browser communication
- Binary `.am` snapshots for persistence

---

## 3) Exported WASI ABI (Current - M0)

See **[API Mapping](docs/reference/api-mapping.md)** for complete API coverage.

### Current Exports (11 functions)

**Memory**: `am_alloc`, `am_free`
**Document**: `am_init`, `am_save`, `am_save_len`, `am_load`, `am_merge`
**Text**: `am_text_splice`, `am_set_text` (deprecated), `am_get_text`, `am_get_text_len`

**Return codes**: `0` = success; `<0` = error code

---

## 4) HTTP API (Demo)

**Current endpoints**:
- `GET /` - Serve UI
- `GET /api/text` - Get current text
- `POST /api/text` - Update text (JSON: `{"text": "..."}`)
- `GET /api/stream` - SSE (events: `snapshot`, `update`)
- `GET /api/doc` - Download `.am` snapshot
- `POST /api/merge` - Merge another `.am` (CRDT merge)

---

## 5) Roadmap / Next Milestones

See **[Development Roadmap](docs/development/roadmap.md)** for complete details.

### Current: M0 Complete ✅
- Text CRDT implementation
- HTTP API + SSE broadcasting
- Binary persistence (`.am` format)
- CRDT merge capability

### Next: M1 — Automerge Sync Protocol
- Per-peer sync state
- Delta-based sync (not whole text)
- `am_sync_gen`, `am_sync_recv` exports

### Future: M2-M5
- **M2**: Multi-object (Maps, Lists, Counters)
- **M3**: NATS Transport
- **M4**: Datastar UI (reactive frontend)
- **M5**: Observability & Ops

---

## 6) Conventions & Guardrails

**Code Style**:
- Go: `gofmt` / `go vet`
- Rust: `cargo fmt` / `cargo clippy`

**Commits**: Conventional Commits (`feat:`, `fix:`, `docs:`, etc.)

**PRs**: Small, reviewed, CI green. Include scope, rationale, testing notes.

**Security**:
- Validate payload sizes
- Cap `am_alloc` usage
- UTF-8 validation (done in Rust)

---

## 7) Quick Checklist (Copy/Paste for PRs)

```markdown
- [ ] Builds: `make build-wasi` ✅
- [ ] Tests: `make test-go` ✅
- [ ] Tests: `make test-rust` ✅
- [ ] Runs: `make run` → `GET /api/text` works ✅
- [ ] SSE: two tabs receive updates ✅
- [ ] Updated: `docs/reference/api-mapping.md` ✅
- [ ] Updated: `TODO.md` ✅
- [ ] Verified: `make verify-docs` ✅
- [ ] CI green ✅
```

---

## 📝 RECENT CHANGES

### 2025-10-20: Test Data Consolidation ✅

Consolidated all test data under `go/testdata/` with clear structure:
- `unit/` - Go package tests (snapshots, scripts)
- `integration/` - Bash integration tests
- `e2e/` - Playwright screenshots

### 2025-10-20: Documentation Reorganization ✅

Applied Diátaxis framework:
- Moved 10 files from root → `docs/`
- Created organized structure (tutorials, how-to, reference, explanation, development, ai-agents, archive)
- Added `make verify-docs` to catch broken links
- Fixed 13 broken internal links

### 2025-10-20: Refactoring - Split exports.go ✅

Achieved perfect 10/10 file mapping between Rust and Go:
- Split `go/pkg/wazero/exports.go` (1,149 lines) → 10 module files
- Each file matches corresponding Rust module exactly
- `runtime.go` → `state.go` to align with `state.rs`

### 2025-10-20: Sync Protocol - Per-Peer State ✅

Fixed incorrect global sync state:
- Changed to proper per-peer state (as Automerge requires)
- `InitSyncState()` now returns `*SyncState` with peer_id
- Added `FreeSyncState()` for cleanup

---

## 📚 Detailed Documentation

For comprehensive guides, see:

- **[Documentation Index](docs/README.md)** - Master index of all documentation
- **[Architecture Guide](docs/explanation/architecture.md)** - Complete 4-layer design
- **[API Mapping](docs/reference/api-mapping.md)** - Full API coverage matrix
- **[Testing Guide](docs/development/testing.md)** - Unit, integration, E2E testing
- **[Automerge Guide](docs/ai-agents/automerge-guide.md)** - CRDT concepts for AI agents
- **[Roadmap](docs/development/roadmap.md)** - Milestones M0-M5 detailed plans
- **[How-To: Add WASI Export](docs/how-to/add-wasi-export.md)** - Step-by-step guide
- **[How-To: Debug WASM](docs/how-to/debug-wasm.md)** - Troubleshooting workflow

---

**Contact / Owner**: @joeblew999
