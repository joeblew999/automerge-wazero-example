# CLAUDE.md — AI Agent Instructions for Automerge + WASI + wazero (Go)

> **Goal**: Run Automerge (Rust) as a **WASI** module hosted by **wazero** (Go), expose a minimal HTTP API + SSE for collaborative text editing, and provide a path to evolve toward **Automerge sync messages** and **NATS** transport.

This document provides **essential instructions** for AI agents. For detailed explanations, see **[Documentation Index](docs/README.md)**.

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

**CRITICAL**: 4-layer architecture - ALL layers must stay synchronized!

```
Layer 1: Automerge Rust Core (.src/automerge/)
           ↓
Layer 2: Rust WASI Exports (rust/automerge_wasi/src/<module>.rs)
           ↓
Layer 3: Go FFI Wrappers (go/pkg/wazero/<module>.go - 1:1 with Layer 2)
           ↓
Layer 4: Go High-Level API (go/pkg/automerge/<module>.go)
```

### 🎯 Perfect 1:1 File Mapping (10/10) ✅

| Rust Module | Go FFI Wrapper | Purpose |
|-------------|----------------|---------|
| state.rs    | state.go       | Global state management |
| memory.rs   | memory.go      | Memory allocation |
| document.rs | document.go    | Document lifecycle |
| text.rs     | text.go        | Text CRDT operations |
| map.rs      | map.go         | Map operations |
| list.rs     | list.go        | List operations |
| counter.rs  | counter.go     | Counter operations |
| history.rs  | history.go     | History/changes |
| sync.rs     | sync.go        | Sync protocol |
| richtext.rs | richtext.go    | Rich text marks |

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

## 0.4) Testing Requirements

**NEVER ASSUME CODE WORKS!** All code MUST be tested.

### Test Workflow

```bash
# Build + test
make build-wasi
make test-go        # Must pass
make test-rust      # Must pass
```

### End-to-End Testing (Playwright MCP)

**REQUIRED** before marking features complete.

See **[Testing Guide](docs/development/testing.md)** for:
- Unit test strategies
- Integration test patterns
- Playwright MCP usage
- Test data generation

---

## 0.5) Primary File Paths (Quick Reference)

```
/Makefile                              # Build automation + verify-docs
/README.md                             # User documentation
/TODO.md                               # Task tracking
/docs/reference/api-mapping.md         # API coverage matrix
/docs/ai-agents/automerge-guide.md     # AI: Automerge concepts
/docs/development/testing.md           # Testing guide
/docs/development/roadmap.md           # Milestones M0-M5
/docs/explanation/architecture.md      # 4-layer architecture deep dive
/ui/ui.html                            # Browser UI
/go/cmd/server/main.go                 # HTTP server
/go/pkg/automerge/*.go                 # High-level Go API
/go/pkg/wazero/*.go                    # Low-level FFI wrappers (1:1 with Rust)
/go/testdata/                          # All test data (unit/integration/e2e)
/rust/automerge_wasi/Cargo.toml        # Rust WASI crate config
/rust/automerge_wasi/src/lib.rs        # Module orchestrator
/rust/automerge_wasi/src/*.rs          # WASI modules (1:1 with Go)
/.src/automerge/                       # Automerge source (reference)
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
