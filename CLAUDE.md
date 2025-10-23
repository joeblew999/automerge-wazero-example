# CLAUDE.md — AI Agent Instructions

> **Goal**: Run Automerge (Rust) as WASI module in wazero (Go), expose HTTP API + SSE, evolve toward Automerge sync + NATS transport.

**For detailed docs**: [Documentation Index](docs/README.md) | **For project status**: [STATUS.md](STATUS.md)

---

## 0) Repository Info

**Repository**: `joeblew999/automerge-wazero-example`  
**URL**: https://github.com/joeblew999/automerge-wazero-example  
**Path**: `/Users/apple/workspace/go/src/github.com/joeblew999/automerge-wazero-example`

⚠️ **CRITICAL**: Always use `joeblew999` (3 nines), not `joeblew99` (2 nines).

**Stack Dependencies**:
- **Automerge**: `.src/automerge/` (v0.7.0) - See [Automerge Guide](docs/ai-agents/automerge-guide.md)
- **Datastar** (M4): `.src/datastar-go/` - See [Datastar Guide](docs/ai-agents/datastar-guide.md)

**Setup**:
```bash
make setup-src        # Clone .src/automerge/
make build-wasi       # Build Rust → WASM
make run              # Start server
```

---

## 1) 🚨 CRITICAL RULES FOR AI AGENTS

**NEVER**:
- ❌ Create new status/tracking/roadmap docs → Use [STATUS.md](STATUS.md) ONLY
- ❌ Skip tests after making changes → Always run `make test`
- ❌ Break 1:1 file mapping → Every CRDT module needs files in ALL 7 layers
- ❌ Assume code works → Verify with actual test results
- ❌ Create session summaries or changelogs → Update STATUS.md "Recent Changes"

**ALWAYS**:
- ✅ Update [STATUS.md](STATUS.md) for any status/milestone changes
- ✅ Run `make build-wasi && make test` after code changes
- ✅ Maintain 1:1 file mapping (see Section 2)
- ✅ Add layer markers to new files (see [docs/templates/layer-markers.md](docs/templates/layer-markers.md))
- ✅ Update [docs/reference/api-mapping.md](docs/reference/api-mapping.md) when adding API methods

---

## 2) 📊 PROJECT TRACKING - STATUS.MD ONLY!

**CRITICAL**: Use ONLY [STATUS.md](STATUS.md) for ALL project tracking.

**STATUS.md contains**:
- Current implementation status
- Milestones (M0, M1, M2, M3, M4, M5)
- Test coverage and statistics
- Future plans and estimates
- Recommended next steps

**Root Folder Structure**:
```
/
├── README.md    # User entry point
├── CLAUDE.md    # AI agent instructions (this file)
├── STATUS.md    # THE ONLY tracking/status document
└── TODO.md      # Active task list
```

**Rule**: If you're about to create a new `.md` about status/roadmap/next-steps → **STOP** and update STATUS.md instead!

---

## 3) 🔥 CODE SYNCHRONIZATION - 7-Layer 1:1 Mapping

**CRITICAL**: Every CRDT operation requires files in ALL 7 layers!

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
           ↓
Layer 7: Web Frontend (web/js/<module>.js + web/components/<module>.html)
```

### Perfect 1:1 File Mapping (10/10 modules)

| Rust | Go FFI | Go API | Server | HTTP | Web JS | Web HTML | Purpose |
|------|--------|--------|--------|------|--------|----------|---------|
| state.rs | state.go | - | server.go | - | - | - | Global state |
| memory.rs | memory.go | - | - | - | - | - | Memory alloc |
| document.rs | document.go | document.go | document.go | - | - | - | Save/Load |
| text.rs | text.go | text.go | text.go | text.go | text.js | text.html | Text CRDT |
| map.rs | map.go | map.go | map.go | map.go | map.js | map.html | Map CRDT |
| list.rs | list.go | list.go | list.go | list.go | list.js | list.html | List CRDT |
| counter.rs | counter.go | counter.go | counter.go | counter.go | counter.js | counter.html | Counter CRDT |
| history.rs | history.go | history.go | history.go | history.go | history.js | history.html | History |
| sync.rs | sync.go | sync.go | sync.go | sync.go | sync.js | sync.html | Sync (M1) |
| richtext.rs | richtext.go | richtext.go | richtext.go | richtext.go | richtext.js | richtext.html | Rich text (M2) |

**Infrastructure files** (NOT 1:1): `server/server.go`, `api/util.go`, `web/js/app.js`, etc.

### crdt_ Prefix Naming

All CRDT operation files use `crdt_` prefix:
```
go/pkg/wazero/crdt_text.go    # CRDT operation
go/pkg/wazero/document.go      # Infrastructure

go/pkg/api/crdt_sync.go        # CRDT operation
go/pkg/api/util.go             # Infrastructure
```

### Layer Responsibilities

**Layer 4 (pkg/automerge)**: Pure CRDT, no state, no mutex
**Layer 5 (pkg/server)**: Stateful, thread-safe, owns `*automerge.Document` + `sync.RWMutex`
**Layer 6 (pkg/api)**: HTTP protocol, parses requests, calls server methods
**Layer 7 (web/)**: Browser UI, calls HTTP API via fetch/SSE

**Why separate Layers 5 & 6?** Enables protocol flexibility - can add gRPC/CLI using same Layer 5 logic.

### When You Change Code

**Adding Go API method** → REQUIRES:
1. ✅ WASI export in `rust/automerge_wasi/src/<module>.rs`
2. ✅ FFI wrapper in `go/pkg/wazero/<module>.go`
3. ✅ Update [docs/reference/api-mapping.md](docs/reference/api-mapping.md)
4. ✅ Tests

**Adding new file** → REQUIRES:
1. ✅ Add layer marker at top (see [docs/templates/layer-markers.md](docs/templates/layer-markers.md))
2. ✅ Verify: `make build-wasi && make test`

---

## 4) 🏗️ DEPLOYMENT ARCHITECTURE

**Model**: Local-first - Go server runs **locally on each device**

```
Browser (JS) → HTTP → Go Server (localhost:8080) → wazero → WASM (Rust Automerge)
```

**Current (M0-M2)**: Single server, multiple browsers  
**Target (M3+)**: Server per device, NATS sync between servers

**Key Points**:
- Custom HTTP/JSON APIs around Automerge (not using Automerge.js)
- Server runs locally (desktop, mobile via gomobile)
- Browser is thin UI connecting to localhost
- NATS syncs between local servers (M3)

**DO NOT SUGGEST**:
- ❌ Running WASI in browser (syscall limitations)
- ❌ Integrating Automerge.js (API mismatch)
- ❌ Changing from local server model

See: [Deployment Architecture](docs/explanation/deployment-architecture.md)

---

## 5) 🧪 TESTING REQUIREMENTS

**NEVER ASSUME CODE WORKS!** All code MUST be tested.

### Testing Philosophy

✅ **WE USE INTEGRATION TESTING** across WASM boundary (intentional!)

**Why?**
- WASM boundary is expensive - don't unit test every FFI call
- Integration tests verify complete stack works
- Catches FFI bugs (memory, pointers) immediately
- Less maintenance, already comprehensive (83 tests)

### Test Workflow

```bash
make build-wasi   # Build Rust → WASM
make test         # Runs test-rust + test-go
make test-rust    # 28 Rust unit tests
make test-go      # 55 Go integration tests
```

### Test Coverage

- **Rust WASI**: 28 tests (unit)
- **Go API**: 48 tests (integration)
- **HTTP API**: 7 tests (integration)
- **Total**: 83 tests, 100% passing ✅

See: [Testing Guide](docs/development/testing.md)

---

## 6) 📋 DOCUMENTATION PRINCIPLES

**Structure** (Diátaxis framework):
```
/
├── README.md    # User entry point
├── CLAUDE.md    # AI agent instructions (this file)
├── STATUS.md    # Project status (ONLY tracking doc!)
├── TODO.md      # Active tasks
└── docs/
    ├── tutorials/      # Learning-oriented
    ├── how-to/         # Goal-oriented recipes
    ├── reference/      # Information lookup
    ├── explanation/    # Understanding concepts
    ├── development/    # Developer workflow
    ├── ai-agents/      # AI-specific guides
    └── templates/      # Code templates
```

**Before moving/renaming files**:
```bash
grep -r "FILENAME.md" . --include="*.md"
git mv OLD.md NEW.md
# Update all references
make verify-docs
git commit
```

### AI-Readability Patterns

**1. Layer Markers**: Every file knows its place
- Shows layer number, responsibilities, dependencies
- Points to tests and related docs

**2. crdt_ Prefix**: Visual separation of CRDT vs infrastructure

**3. 1:1 File Mapping**: Predictable structure across layers

See: [AI Readability Improvements](docs/explanation/ai-readability-improvements.md)

---

## 7) 📂 QUICK REFERENCE

### Primary Files

```
/Makefile                   # Build automation
/README.md                  # User docs
/STATUS.md                  # **THE ONLY** tracking doc
/TODO.md                    # Active tasks

# Go layers
/go/pkg/wazero/*.go         # Layer 3: FFI wrappers
/go/pkg/automerge/*.go      # Layer 4: Pure CRDT API
/go/pkg/server/*.go         # Layer 5: Stateful + thread-safe
/go/pkg/api/*.go            # Layer 6: HTTP handlers

# Rust WASI
/rust/automerge_wasi/src/   # Layer 2: WASI exports

# Web UI
/web/index.html             # Main entry
/web/js/*.js                # Component modules (1:1 with automerge/*.go)
/web/components/*.html      # UI templates (1:1 with api/*.go)
/ui/vendor/automerge.js     # Built from .src/ (3.4M IIFE)
```

### Common Commands

```bash
# Build & Test
make build-wasi         # Rust → WASM
make test               # All tests
make run                # Start server

# Development
make verify-docs        # Check markdown links
make verify-web         # Check web folder structure

# Setup
make setup-src          # Clone .src/automerge/
make build-js           # Build Automerge.js from source
make sync-versions      # Verify version alignment
```

### Environment

**Platform**: darwin (macOS)  
**Go**: 1.21+  
**Rust**: stable, `wasm32-wasip1` target  
**Working Dir**: `/Users/apple/workspace/go/src/github.com/joeblew999/automerge-wazero-example`

---

## 8) 🔗 DETAILED DOCUMENTATION

**For Users**:
- [Getting Started](docs/tutorials/getting-started.md) - Tutorial
- [HTTP API Reference](docs/reference/http-api-complete.md) - All endpoints

**For Developers**:
- [Architecture](docs/explanation/architecture.md) - 7-layer design
- [API Mapping](docs/reference/api-mapping.md) - API coverage matrix
- [Testing Guide](docs/development/testing.md) - Test strategies
- [Web Architecture](docs/explanation/web-architecture.md) - Web folder 1:1 mapping
- [Build Automerge.js](docs/how-to/build-automerge-js.md) - Building from source

**For AI Agents**:
- [Automerge Guide](docs/ai-agents/automerge-guide.md) - CRDT concepts
- [Layer Markers](docs/templates/layer-markers.md) - Code templates
- [AI Readability](docs/explanation/ai-readability-improvements.md) - Patterns

**How-To Guides**:
- [Add WASI Export](docs/how-to/add-wasi-export.md) - Step-by-step
- [Debug WASM](docs/how-to/debug-wasm.md) - Troubleshooting
- [Embed in Your App](docs/how-to/embed-in-your-app.md) - Integration

---

## 9) ✅ PR CHECKLIST

```markdown
- [ ] Builds: `make build-wasi` ✅
- [ ] Tests: `make test` ✅
- [ ] Docs: `make verify-docs` ✅
- [ ] Updated: [STATUS.md](STATUS.md) ✅
- [ ] Updated: [docs/reference/api-mapping.md](docs/reference/api-mapping.md) (if API changed) ✅
- [ ] Layer markers added to new files ✅
```

---

**Contact**: @joeblew999
