# Automerge WASI Demo - Complete Guide

This demo showcases **Automerge CRDTs** running in **WebAssembly (WASI)** via **Rust**, hosted by **Go** using **wazero**, with real-time browser sync!

---

## 🎯 What This Demo Does

### ✅ What WORKS Right Now:

1. **Text CRDT Implementation** - Proper Automerge Text type (not plain strings)
2. **Real-Time Collaboration** - Multiple browsers editing the same document via SSE
3. **Independent Storage** - 2-laptop architecture (Alice & Bob can run separately)
4. **Binary CRDT Format** - doc.am files contain full operation history
5. **Rust WASI Module** - Automerge library compiled to WebAssembly
6. **Go wazero Runtime** - Pure Go WebAssembly runtime (no CGO!)
7. **Browser Integration** - Automerge.js loaded via CDN

### 🚧 What's NOT Implemented (Yet):

- Full binary sync protocol (still using full-text broadcast)
- Offline editing + eventual merge between Alice & Bob
- Multi-document support
- Presence/cursors
- Conflict visualization

---

## 🚀 Quick Start

### 1. **Real-Time Collaboration** (Single Server)

```bash
# Terminal 1: Start server
make run

# Open multiple browser windows
open http://localhost:8080
open http://localhost:8080

# Type in one window → see it appear in real-time in the other!
```

**How it works:**
- Server broadcasts text changes via Server-Sent Events (SSE)
- All connected browsers receive updates instantly
- Text CRDT ensures proper merge semantics

---

### 2. **2-Laptop Simulation** (Alice & Bob)

```bash
# Terminal 1: Alice's laptop
make run-alice
# Server on port 8080, storage: data/alice/doc.am

# Terminal 2: Bob's laptop
make run-bob
# Server on port 8081, storage: data/bob/doc.am

# Open browsers
open http://localhost:8080  # Alice
open http://localhost:8081  # Bob

# Type different text in each!
# They maintain independent doc.am files
```

**Storage locations:**
```
go/cmd/server/data/
├── alice/
│   └── doc.am (196 bytes - Text CRDT binary)
└── bob/
    └── doc.am (201 bytes - Text CRDT binary)
```

**Verify CRDT format:**
```bash
# Check file sizes (should be >50 bytes, not just text length)
ls -lh go/cmd/server/data/{alice,bob}/doc.am

# Hexdump shows Automerge magic bytes: 85 6f 4a 83
hexdump -C go/cmd/server/data/alice/doc.am | head -3
```

---

## 🧪 Testing

### Run Automated Tests

```bash
# Install dependencies
npm install @automerge/automerge

# Run test suite (8 tests)
node test_text_crdt.mjs
```

**Tests verify:**
- ✅ Automerge.js imports
- ✅ Text CRDT created (not plain string)
- ✅ `updateText()` works
- ✅ Binary format >50 bytes
- ✅ Edit history preserved
- ✅ Concurrent edits merge correctly
- ✅ Server connectivity

### Browser Tests

Open: http://localhost:8080/test_text_crdt.html

Runs 6 tests in browser showing Automerge.js functionality.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Browser                          │
│  ┌──────────────────────────────────────────────┐  │
│  │ Automerge.js (via CDN)                       │  │
│  │  - Maintains local doc = from({text: ""})    │  │
│  │  - On keystroke: updateText(doc, ["text"], val)│  │
│  │  - POST to /api/text                         │  │
│  │  - Listen to /api/stream (SSE)               │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
                         ↕ HTTP/SSE
┌─────────────────────────────────────────────────────┐
│               Go Server (wazero)                    │
│  ┌──────────────────────────────────────────────┐  │
│  │ HTTP Handlers:                               │  │
│  │  - GET  /api/text    → get current text      │  │
│  │  - POST /api/text    → set text & broadcast  │  │
│  │  - GET  /api/stream  → SSE connection        │  │
│  │  - GET  /            → serve ui.html         │  │
│  └──────────────────────────────────────────────┘  │
│                         ↕                            │
│  ┌──────────────────────────────────────────────┐  │
│  │ Rust WASI Module (automerge_wasi.wasm)      │  │
│  │  - am_init()         → create doc with Text  │  │
│  │  - am_text_splice()  → character-level ops   │  │
│  │  - am_set_text()     → set full text         │  │
│  │  - am_get_text()     → retrieve text         │  │
│  │  - am_save()         → serialize to binary   │  │
│  │  - am_load()         → deserialize from file │  │
│  │  - am_merge()        → CRDT merge (NEW!)     │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
                         ↕
┌─────────────────────────────────────────────────────┐
│               Filesystem Storage                    │
│  - doc.am (default single server)                  │
│  - data/alice/doc.am (2-laptop mode)                │
│  - data/bob/doc.am   (2-laptop mode)                │
└─────────────────────────────────────────────────────┘
```

---

## 📁 Key Files

| File | Purpose |
|------|---------|
| `rust/automerge_wasi/src/lib.rs` | Rust WASI module wrapping Automerge |
| `go/cmd/server/main.go` | Go server with wazero runtime |
| `ui/ui.html` | Browser UI with Automerge.js |
| `Makefile` | Build & run commands |
| `TODO.md` | Implementation roadmap (10 phases) |
| `AGENT_AUTOMERGE.md` | AI knowledge base for Automerge |
| `test_text_crdt.mjs` | Automated test suite |
| `test_text_crdt.html` | Browser test page |

---

## 🔧 Development

### Build Commands

```bash
make help               # Show all commands
make build-wasi         # Build Rust WASI module (release)
make build-wasi-debug   # Build debug version (faster)
make run                # Build + run server (port 8080)
make dev                # Use debug WASM for faster iteration
make clean              # Clean build artifacts
make clean-snapshots    # Delete doc.am files
make clean-all          # Clean everything
```

### 2-Laptop Commands

```bash
make run-alice          # Alice on port 8080
make run-bob            # Bob on port 8081
make test-two-laptops   # Start both simultaneously
make clean-test-data    # Clean test directories
```

### Environment Variables

```bash
PORT=8081               # Server port (default: 8080)
STORAGE_DIR=./my/path   # Where to store doc.am (default: ../../../)
USER_ID=alice           # Logging identifier (default: "default")
```

**Example:**
```bash
PORT=9000 STORAGE_DIR=./data/custom USER_ID=charlie make run-server
```

---

## 📊 What is `doc.am`?

`doc.am` is a **binary snapshot file** containing:

1. **Full operation history** - Every single character edit ever made
2. **CRDT metadata** - Actor IDs, timestamps, causal ordering
3. **Automerge header** - Magic bytes: `85 6f 4a 83`
4. **Compressed format** - Efficient binary encoding

**NOT just current text!** It's the entire CRDT state for merging.

Example file structure:
```
00000000  85 6f 4a 83 94 2b f0 43  |.oJ..+.C|  ← Automerge header
00000010  10 52 9d dd 09 ba 92 4e  |.R.....N|  ← Operation history
00000020  34 99 c9 09 52 f2 84 14  |4...R...|  ← CRDT metadata
...
```

---

## 🎓 CRDT Concepts

### What's a CRDT?

**Conflict-free Replicated Data Type** - A data structure that can be edited independently on multiple devices and merged without conflicts.

### Automerge.Text CRDT

- Each character has a unique ID
- Concurrent inserts/deletes merge deterministically
- **No "last write wins"** - all edits preserved
- Uses list CRDT (RGA/Peritext algorithm)

### Example Merge Scenario

```
Alice's doc (offline):  "Hello World"
Bob's doc (offline):    "Hello Everyone"

After merge (CRDT magic!):
Possible result: "Hello World Everyone"
(Exact result depends on operation timestamps)
```

---

## 🚧 Known Limitations

1. **Simplified Sync** - Currently broadcasts full text, not binary deltas
2. **No Merge UI** - Merge function exists but no HTTP endpoint yet
3. **Single Document** - No multi-doc support
4. **No Presence** - Can't see other users' cursors
5. **No Offline Support** - Must stay connected to server

See [TODO.md](TODO.md) for complete implementation roadmap (Phases 1-10).

---

## 🎯 Future Phases

**Phase 1:** Document current state (YOU ARE HERE!)
**Phase 2:** Binary sync protocol
**Phase 3:** Multi-document support
**Phase 4:** Presence & cursors
**Phase 5:** Offline mode
**Phase 6:** NATS integration
**Phase 7:** Storage adapter
**Phase 8:** Conflict visualization
**Phase 9:** Performance optimization
**Phase 10:** Production ready

---

## 🐛 Troubleshooting

### Server won't start - port in use

```bash
# Kill existing server
killall -9 main
lsof -ti :8080 | xargs kill -9

# Or use different port
PORT=8081 make run
```

### Can't build WASI module

```bash
# Install Rust WASI target
make install-deps

# Check dependencies
make check-deps
```

### Browser not updating in real-time

1. Check SSE connection in browser console
2. Look for "Connected" status in UI
3. Check server logs for SSE messages

### doc.am file is plain text

This means Text CRDT isn't working! File should be:
- **>50 bytes** (CRDT overhead)
- **Binary format** (starts with `85 6f 4a 83`)
- **NOT** human-readable

Run tests to verify: `node test_text_crdt.mjs`

---

## 📚 Resources

- **Automerge Docs:** https://automerge.org/docs/
- **Automerge Rust:** https://docs.rs/automerge/
- **wazero:** https://wazero.io/
- **CRDT Explained:** https://crdt.tech/

---

## 🤝 Contributing

This is a demo project showcasing Automerge + WASI + wazero integration.

See [TODO.md](TODO.md) for implementation tasks.

---

## 📄 License

MIT License - See project repository for details.

---

**Built with:**
- 🦀 Rust (Automerge CRDT)
- 🐹 Go (wazero runtime)
- 🌐 WebAssembly (WASI)
- ⚡ Automerge.js (browser)
- 📡 Server-Sent Events (real-time)

**🎉 Automerge CRDTs + WebAssembly = Collaborative Magic!**
