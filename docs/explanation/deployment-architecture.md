# Deployment Architecture

## Overview

This project wraps **Automerge** (Rust CRDT library) with custom APIs:

```
Browser (HTML/JS) → HTTP/JSON → Go Server → wazero → WASM (Rust Automerge)
```

**Key Components**:
- **Rust**: Automerge core compiled to `wasm32-wasip1` (WASI)
- **Go**: wazero hosts WASM, provides HTTP/JSON APIs
- **Browser**: Thin UI layer (HTML/CSS/JS)

**Key Insight**: The Go server IS the application. Browser is just a UI client.

---

## Deployment Models

### Current: Centralized Server (M0/M1/M2)

```
┌─────────┐     ┌─────────┐     ┌─────────┐
│Browser 1│────▶│         │◀────│Browser 2│
└─────────┘     │         │     └─────────┘
                │ Go      │
┌─────────┐     │ Server  │     ┌─────────┐
│Browser 3│────▶│ (WASM)  │◀────│Browser 4│
└─────────┘     │         │     └─────────┘
                └─────────┘
                     │
                  doc.am
```

**Characteristics**:
- Single Go server instance
- Multiple browser clients
- Server is CRDT authority
- SSE for real-time updates
- Intentional design for demo/testing

### Target: Local-First (M3+)

```
Laptop                    Phone                     Desktop
┌─────────────┐          ┌─────────────┐          ┌─────────────┐
│ Browser     │          │ Browser     │          │ Browser     │
│ localhost   │          │ localhost   │          │ localhost   │
└──────┬──────┘          └──────┬──────┘          └──────┬──────┘
       │                        │                        │
┌──────▼──────┐          ┌──────▼──────┐          ┌──────▼──────┐
│ Go Server   │          │ Go Server   │          │ Go Server   │
│   (WASM)    │◀────────▶│   (WASM)    │◀────────▶│   (WASM)    │
└──────┬──────┘   NATS   └──────┬──────┘   NATS   └──────┬──────┘
       │                        │                        │
    local.am                 local.am                 local.am
```

**Characteristics**:
- Go server on **each device**
- Browser connects to \`localhost:8080\`
- Local \`.am\` persistence
- NATS for peer-to-peer sync
- True offline-first

---

## Why This Architecture?

### Custom APIs Required

We needed features that Automerge.js doesn't provide:
- **Path-based operations**: \`/content/text\` navigation
- **HTTP/JSON protocol**: Not binary sync messages
- **Server-side CRDT**: Go owns the document state
- **Custom endpoints**: Text, Map, List, Counter, History, Sync, RichText

**Solution**: Build Go wrapper around Rust Automerge, compile to WASM.

### Code Reuse Everywhere

Same Go binary runs on:
- Linux servers (centralized mode)
- macOS/Linux desktop (local mode)
- iOS (embedded via gomobile)
- Android (embedded via gomobile)

**Zero platform-specific code.**

### Offline-First Benefits

Local server + NATS sync provides:
- ✅ Works without network (local CRDT operations)
- ✅ Fast (no network latency)
- ✅ Secure (data stays on device)
- ✅ Sync when online (NATS connects peers)

---

## Architecture Decisions

### Why Not Browser WASM?

**Option**: Run WASI module directly in browser via polyfills

**Why Rejected**:
- WASI syscalls (\`fd_write\`, \`random_get\`, etc.) need polyfills
- Performance overhead (JS-based polyfills)
- Already have Go wrapper that works everywhere
- No benefit vs local server approach

**Technical Note**: Our WASM uses \`wasm32-wasip1\` (WASI target). Browsers need \`wasm32-unknown-unknown\`. Would require separate build.

### Why Not Automerge.js?

**Option**: Use Automerge.js in browser, binary sync to server

**Why Rejected**:
- API mismatch: Our HTTP/JSON ≠ Automerge.js binary sync
- Would need to rewrite all server logic in browser JS
- Local server gives same benefits (offline, fast) with **zero code changes**

**Key Insight**: We already built the wrapper (Go + WASM). Just run it locally!

---

## M3 Implementation Plan

### Architecture Shift

**Change**: From centralized server to local-first deployment

**No Code Changes Required**:
- ✅ WASM module works as-is
- ✅ Go server works as-is
- ✅ Web UI works as-is (connects to localhost)

**Only Additions**:
- Add NATS client to Go server
- Add service configs (systemd/launchd)
- Add mobile bindings (gomobile)

### NATS Sync Protocol

**Each device**:
1. Runs local Go server
2. Maintains local \`.am\` file
3. Connects to NATS cluster
4. Publishes sync messages when document changes
5. Subscribes to sync messages from other peers
6. Applies incoming changes via Automerge CRDT merge

**Already Implemented (M1)**:
- Binary sync message generation (\`am_sync_gen\`)
- Binary sync message receiving (\`am_sync_recv\`)
- Per-peer sync state
- HTTP endpoint: \`POST /api/sync\`

**Need to Add**:
- NATS client library
- Subscribe to \`automerge.sync.*\` topics
- Publish sync messages to NATS
- Online/offline transition handling

### Mobile Deployment

**iOS**:
\`\`\`swift
import Automerge  // gomobile binding

// Start Go server on app launch
AutomergeStartServer(documentsPath)

// Load web UI in WKWebView
webView.load(URLRequest(url: URL(string: "http://localhost:8080")!))
\`\`\`

**Android**:
\`\`\`kotlin
import automerge.Automerge  // gomobile binding

// Start Go server
Automerge.startServer(filesDir.absolutePath)

// Load web UI in WebView
webView.loadUrl("http://localhost:8080")
\`\`\`

### Desktop Deployment

**Linux (systemd)**:
\`\`\`ini
[Unit]
Description=Automerge Local Server

[Service]
ExecStart=/usr/local/bin/automerge-server --data ~/.automerge/doc.am --nats nats://sync.example.com

[Install]
WantedBy=default.target
\`\`\`

**macOS (launchd)**:
\`\`\`xml
<plist>
  <dict>
    <key>Label</key>
    <string>com.example.automerge</string>
    <key>ProgramArguments</key>
    <array>
      <string>/usr/local/bin/automerge-server</string>
      <string>--data</string>
      <string>~/Library/Application Support/Automerge/doc.am</string>
    </array>
  </dict>
</plist>
\`\`\`

---

## Comparison: Centralized vs Local-First

| Aspect | Centralized (M0-M2) | Local-First (M3+) |
|--------|---------------------|-------------------|
| **Server Location** | Remote VPS/cloud | localhost on each device |
| **Offline Support** | ❌ No | ✅ Yes |
| **Network Latency** | 50-200ms | 0ms (local) |
| **Data Privacy** | Server sees all data | Data stays on device |
| **Deployment** | Single instance | Per-device service |
| **Sync Protocol** | SSE (server push) | NATS (peer-to-peer) |
| **Code Changes** | Current codebase | Zero changes needed! |
| **Mobile Support** | Web-only | Native apps (gomobile) |

---

## Frequently Asked Questions

### What about web-only users who can't install software?

Run the **centralized model** (M0-M2 architecture). Same codebase supports both:
- Deploy single Go server to cloud (Fly.io, Railway, etc.)
- Users visit \`https://yourapp.com\`
- No installation required

The local-first model is **optional** for users who want offline support.

### How does sync work between devices?

1. User edits text on Laptop → Local server updates \`local.am\`
2. Local server generates sync message (M1 protocol)
3. Server publishes to NATS topic \`automerge.sync.laptop-001\`
4. NATS broadcasts to all subscribed peers
5. Phone receives sync message → Applies to local \`local.am\`
6. Automerge CRDT ensures conflict-free merge

**No central server required!** NATS is just a message bus.

### Can I run both models simultaneously?

Yes! Example:
- **Office**: Centralized server for team collaboration
- **Mobile**: Local server for offline work
- **Sync**: When mobile comes online, merge with office server

Same \`.am\` format, same merge logic.

### What if two people edit offline?

**Automerge CRDT handles this!**

\`\`\`
Laptop (offline): "Hello World" → "Hello Alice"
Phone (offline):  "Hello World" → "Hello Bob"

When they sync:
Result: "Hello Alice Bob"  (deterministic merge)
\`\`\`

No "last write wins", no conflicts. Both edits preserved.

### Performance: Local vs Centralized?

**Local-first is faster**:
- Text edit → 0ms (local WASM call)
- Save → ~1ms (write to local file)
- Sync → Background (doesn't block UI)

**Centralized**:
- Text edit → 50-200ms (HTTP round-trip)
- Save → Server-side (network dependent)
- SSE → Push to all clients (network bandwidth)

---

## Design Principles

### Single Codebase

Same \`automerge-server\` binary everywhere:
- Cloud VPS (centralized mode)
- Desktop (local service)
- Mobile (embedded via gomobile)

**No platform-specific forks.**

### WASM as Universal Runtime

Rust Automerge compiled to WASM runs:
- On Linux x86_64 (wazero)
- On macOS ARM64 (wazero)
- On iOS (wazero via gomobile)
- On Android (wazero via gomobile)

**No need for C FFI or platform bindings.**

### Progressive Enhancement

- **Start**: Centralized server (easiest deployment)
- **Add**: Local server (better performance, offline)
- **Add**: NATS sync (multi-device)
- **Add**: Mobile apps (native experience)

Each step builds on previous architecture.

---

## Implementation Status

- ✅ **M0**: Core CRDT (Text, Map, List, Counter, History)
- ✅ **M1**: Sync protocol (binary messages, per-peer state)
- ✅ **M2**: Rich text (marks, formatting)
- 🚧 **M3**: Multi-device (NATS transport, deployment) ← **Next**
- 🚧 **M4**: Datastar UI (reactive frontend)
- 🚧 **M5**: Observability (metrics, tracing)

---

## See Also

- [STATUS.md](../../STATUS.md) - M3 NATS Transport milestone
- [Architecture](architecture.md) - Go package structure
- [Testing Guide](../development/testing.md) - Multi-device test scenarios
- [Automerge Guide](../ai-agents/automerge-guide.md) - CRDT concepts
