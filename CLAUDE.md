# AGENTS.md — Automerge + WASI + wazero (Go) collaborative text demo

> **Goal**: Run Automerge (Rust) as a **WASI** module hosted by **wazero** (Go), expose a minimal HTTP API + SSE for collaborative text editing, and provide a path to evolve toward **Automerge sync messages** and **NATS** transport.

This document instructs automation agents (and humans) how to build, run, extend, and test the project. Follow tasks in order unless stated otherwise.

---

## 0) Repo assumptions

**Repository**: `joeblew999/automerge-wazero-example`

YOU MUST USE FILE PATH for code: /Users/apple/workspace/go/src/github.com/joeblew999/automerge-wazero-example

**stack**:

automerge and rust

https://github.com/automerge/automerge

https://github.com/automerge/automerge/releases/tag/rust%2Fautomerge%400.7.0

https://automerge.org/docs/hello/

---

datastar and golang

https://github.com/starfederation/datastar-go

https://data-star.dev

**TESTING**

you MUST use your playwright MCP to Test that it works from the outside. 

you MUST keep the makefile and README.md up to date.

**Branches**:

* `main` — stable, protected.
* `dev/*` — feature branches, merge via PR.

**Primary paths**:

```
/Makefile
/README.md
/ui/ui.html
/go/cmd/server/main.go
/rust/automerge_wasi/Cargo.toml
/rust/automerge_wasi/src/lib.rs
```

---

## 1) Environment & prerequisites

* **Rust** (stable): `rustup` installed
* **Targets**:

  * Prefer `wasm32-wasi` (pre-1.84) or `wasm32-wasip1` (Rust 1.84+).
  * We will default to `wasm32-wasi` initially. Switch target in `Makefile` when toolchain updates.
* **Go**: 1.21+
* **Make**

### Local bootstrap

```bash
make build-wasi   # builds rust → WASI .wasm
make run          # runs Go server with wazero
# open http://localhost:8080
```

**Artifacts**:

* `rust/automerge_wasi/target/wasm32-wasi/release/automerge_wasi.wasm`
* Snapshot persisted as `doc.am` in repo root (for demo)

---

## 2) Architecture (high-level)

* **Rust crate (`automerge_wasi`)**

  * Wraps Automerge core (`automerge` crate) and exposes a small C-like ABI over WASI
  * Exports: memory helpers (`am_alloc`, `am_free`), lifecycle (`am_init`), whole-text set/get, and snapshot save/load
* **Go server (wazero host)**

  * Instantiates the WASI module, holds one in-memory document (demo)
  * HTTP endpoints: `GET /api/text`, `POST /api/text`
  * **SSE** at `GET /api/stream` for broadcasting updates
  * Persists `doc.am` and reloads on startup
* **UI**

  * `ui/ui.html`: simple textarea + SSE listener + Save button

---

## 3) Tasks for agents

### T1 — Ensure repository skeleton

* [ ] Create `Makefile`, `README.md`, `ui/ui.html`, `go/cmd/server/main.go`, `rust/automerge_wasi/{Cargo.toml, src/lib.rs}` (see current repo contents)
* [ ] `go.mod` with `github.com/tetratelabs/wazero`
* [ ] Compile & run: `make build-wasi && make run`

### T2 — Developer DX

* [ ] Add `make tidy` (runs `go mod tidy`)
* [ ] Optional: add file-watcher for hot-reload on `.wasm` changes (e.g., `reflex` or `watchexec`). On change: re-instantiate module.

### T3 — Quality gates

* [ ] Add GitHub Actions CI: build WASI target and `go build` server
* [ ] Lint: `golangci-lint` (optional), `cargo clippy` (optional)

### T4 — Error handling & logging

* [ ] Map negative return codes in Rust to HTTP 4xx/5xx in Go
* [ ] Structured logging in Go (std log OK for demo)

### T5 — Persistence policy

* [ ] Keep latest snapshot `doc.am`
* [ ] (Optional) Periodic snapshots + rotation; add `make clean-snapshots`

---

## 4) Exported WASI ABI (current)

**Memory**

* `am_alloc(size: usize) -> *mut u8` — guest allocates buffer for host writes
* `am_free(ptr: *mut u8, size: usize)` — guest frees a prior allocation

**Lifecycle**

* `am_init() -> i32` — initialize a single Automerge doc with a Text object at key `"content"`

**Whole-text (demo)**

* `am_set_text(ptr: *const u8, len: usize) -> i32` — replace entire text
* `am_get_text_len() -> u32` — byte length for next `am_get_text`
* `am_get_text(ptr_out: *mut u8) -> i32` — copies text bytes into guest buffer at `ptr_out`

**Snapshots**

* `am_save_len() -> u32`
* `am_save(ptr_out: *mut u8) -> i32`
* `am_load(ptr: *const u8, len: usize) -> i32`

**Return codes**: `0` success; `<0` error. (Map to structured errors later.)

---

## 5) HTTP API (demo)

* `GET /api/text` → `200 text/plain` returns current buffer
* `POST /api/text` `{"text": string}` → `204 No Content` on success; broadcasts SSE `update`
* `GET /api/stream` → SSE with events:

  * `snapshot` on connect: `{ "text": string }`
  * `update` on edits: `{ "text": string }`

---

## 6) Roadmap / Next milestones

### M1 — **Automerge Sync Protocol** (delta-based)

**Why**: avoid shipping whole text; support true peer-to-peer style sync.

Add to Rust (`src/lib.rs`):

* [ ] `am_sync_init_peer(peer_id_ptr,len) -> i32` (optional if single peer)
* [ ] `am_sync_gen_len() -> u32`
* [ ] `am_sync_gen(ptr_out: *mut u8) -> i32`
* [ ] `am_sync_recv(ptr: *const u8, len: usize) -> i32`

Update Go:

* [ ] On local edit, call `am_sync_gen` and broadcast bytes (SSE or NATS)
* [ ] On receive, call `am_sync_recv` then, if needed, `am_sync_gen` (Automerge may request a reply)
* [ ] Add `/api/sync` SSE channel or reuse `/api/stream` with a new `event: sync`

### M2 — **Multi-document support**

* [ ] Replace single `DOC` with a map keyed by `docId`
* [ ] Expose `am_select(doc_id_ptr,len)` / `am_new_doc(doc_id_ptr,len)`
* [ ] Query param `?doc=<id>` on HTTP routes
* [ ] Snapshot files `data/<docId>.am`

### M3 — **NATS transport** (fits platform stack)

* [ ] Subjects: `automerge.sync.<tenant>.<docId>`
* [ ] Server acts as a peer: on msg → `am_sync_recv` → maybe reply with `am_sync_gen`
* [ ] Store snapshots in **NATS Object Store**; latest head in KV per `docId`
* [ ] RBAC via JWT; namespace subjects per-tenant/region

### M4 — **Datastar UI** (browser or TUI)

* [ ] Browser: minimal JS that streams sync messages via SSE (or NATS WS bridge)
* [ ] Datastar “action” hooks to send local ops and apply remote updates
* [ ] Optional WASM-Go frontends calling HTTP or NATS

### M5 — **Observability & ops**

* [ ] Metrics: flush counts, message sizes, per-doc peers
* [ ] Tracing hooks around sync cycles
* [ ] Config flags for runtime paths and limits

---

## 7) Conventions & guardrails

**Commits**: Conventional Commits (`feat:`, `fix:`, `chore:` …)

**PRs**: Small, reviewed, CI green. Include:

* Scope, rationale
* Testing notes
* Backward-compat considerations

**Code style**:

* Go: `gofmt`/`go vet`
* Rust: `cargo fmt`/`cargo clippy`

**Security**:

* Validate payload sizes; cap `am_alloc` usage
* Don’t trust client text blindly (UTF-8 checked already). Add content-length bounds in HTTP.

**Performance**:

* Single module instance is fine for demo; for prod consider per-doc sharding or doc pool
* Avoid excessive `alloc/free` by reusing buffers; measure with pprof later

---

## 8) Testing plan

**Unit**

* Rust: construct doc, set text, save/load, compare
* Go: in-memory server test that calls handlers and validates SSE frames

**Integration**

* Start server → connect two SSE clients → POST update → assert second client receives `update`

**CLI smoke**

```bash
curl -s http://localhost:8080/api/text
curl -s -X POST http://localhost:8080/api/text -H 'content-type: application/json' -d '{"text":"Hello"}' -i
curl -s http://localhost:8080/api/stream  # observe snapshot + updates
```

---

## 9) Switch to `wasm32-wasip1` when ready

* Update `Makefile` target and Go path to the built `.wasm`
* Confirm CI installs the right target (`rustup target add wasm32-wasip1`)

---

## 10) Future extensions (optional)

* CRDT rich text or multiple fields (not just `content`)
* Heads/hash exposure for time travel
* Snapshot compaction/GC strategy
* E2E encryption of sync messages (app layer)
* Rollups to object store per interval

---

## 11) Quick checklist (copy/paste for PRs)

* [ ] Builds: `make build-wasi` ✅
* [ ] Runs: `make run` → `GET /api/text` works ✅
* [ ] SSE: two tabs receive `snapshot`/`update` ✅
* [ ] Snapshot persists and reloads ✅
* [ ] CI green ✅

---

**Contact / Owner**: @joeblew999
