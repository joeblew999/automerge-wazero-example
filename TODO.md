# TODO - Automerge WASI Demo

## Current State

✅ **Working:** Basic collaborative text editor
- Rust WASI module wraps Automerge
- Go server with wazero hosts WASM
- SSE broadcasts text changes to all tabs
- Single shared `doc.am` on server

❌ **NOT Implemented:** True Automerge sync protocol

## The Problem

**Current demo is NOT using Automerge correctly!**

- All clients edit ONE shared document on the server
- Uses whole-text replacement (POST full text)
- NOT using Automerge's sync protocol (deltas)
- Each "laptop" should have its own `doc.am` and sync changes

## What Automerge SHOULD Do

### Real Architecture:
```
Laptop A                 Laptop B
  doc.am                   doc.am
    |                        |
    | Sync deltas            |
    | (just changes)         |
    └────────┬───────────────┘
             |
        NATS/WebSocket
```

### Current (Wrong) Architecture:
```
Browser Tab A    Browser Tab B
      |                |
      └────► Server ◄──┘
          ONE doc.am
```

## TODO List

### Phase 1: Document Current State
- [ ] Add section to README explaining this is a simplified demo
- [ ] Note that true Automerge sync is NOT implemented
- [ ] Reference CLAUDE.md Milestone M1 for full sync implementation

### Phase 2: Implement Automerge Sync Protocol (M1)

#### Rust WASI Module
- [ ] Add `am_sync_init_peer(peer_id, len) -> i32`
- [ ] Add `am_sync_gen_len() -> u32`
- [ ] Add `am_sync_gen(ptr_out) -> i32` - Generate sync message (delta)
- [ ] Add `am_sync_recv(ptr, len) -> i32` - Receive and apply sync message
- [ ] Support multiple Automerge document instances (not just one global)

#### Go Server
- [ ] Replace single shared doc with per-client document map
- [ ] Add `/api/sync` SSE endpoint for Automerge sync messages
- [ ] On local edit: call `am_sync_gen` and broadcast delta (not full text)
- [ ] On receive: call `am_sync_recv` then maybe `am_sync_gen` (Automerge may request reply)
- [ ] Store per-client `doc.am` snapshots in `data/<clientId>.am`

#### UI
- [ ] Replace POST /api/text with Automerge sync message exchange
- [ ] Send local edits as sync deltas
- [ ] Apply received sync deltas to local document
- [ ] Keep local Automerge state (not just textarea)

### Phase 3: Multi-Document Support (M2)
- [ ] Support multiple documents via `?doc=<id>` query param
- [ ] Add `am_select(doc_id)` / `am_new_doc(doc_id)` to Rust
- [ ] Store snapshots per document: `data/<docId>.am`

### Phase 4: Real-Time Collaborative Editing Features
- [ ] **Multiple Cursors/Carets** - Show where each user is typing
  - [ ] Track cursor position for each connected user
  - [ ] Assign unique color to each user
  - [ ] Display remote cursors in textarea with user name/color
- [ ] **User Presence** - Who's online
  - [ ] Show list of connected users
  - [ ] Display user status (typing, idle, offline)
  - [ ] Show user avatars or initials with their color
- [ ] **Real-Time Character-by-Character Updates**
  - [ ] Sync on every keystroke (not just Save button)
  - [ ] Use Automerge operational transforms for fine-grained edits
  - [ ] Show typing indicators ("User X is typing...")
- [ ] **Remove "Save" button** - everything auto-syncs
- [ ] **Conflict visualization** - Highlight conflicting edits being merged

### Phase 5: NATS Transport (M3)
- [ ] Replace HTTP SSE with NATS subjects: `automerge.sync.<tenant>.<docId>`
- [ ] Store snapshots in NATS Object Store
- [ ] Add RBAC via JWT

### Phase 6: Testing
- [ ] Test with two separate processes (simulating two laptops)
- [ ] Verify offline edits merge correctly when reconnected
- [ ] Test conflict resolution (both edit same line)
- [ ] Add automated tests

## Why This Matters

**Automerge is designed for:**
- Offline-first applications
- Conflict-free merging of concurrent edits
- Peer-to-peer sync without central server
- Minimal bandwidth (only sync deltas, not full state)

**Current demo just shows:**
- The tech stack works (Rust WASI + Go + wazero)
- Basic SSE broadcasting
- NOT the real power of Automerge!

## References

- CLAUDE.md - Full implementation guide (see Milestone M1-M5)
- https://automerge.org/docs/how-it-works/ - Automerge concepts
- https://automerge.org/docs/cookbook/real-time/ - Real-time collaboration guide

## Notes

This demo is a **proof-of-concept** for the tech stack. To make it a real Automerge application, implement Phase 2 (Milestone M1) from CLAUDE.md.
