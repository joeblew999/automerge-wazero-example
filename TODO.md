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

## What is a `doc.am` file?

**`doc.am` is a binary snapshot file** that contains the entire state and history of an Automerge CRDT document.

### Key Concepts:

**1. Not Just Current State**
- It doesn't just store "Hello World" (the current text)
- It stores **every single edit** that was ever made to the document
- Example: "H" added, then "e", then "llo", then someone deleted "l", etc.

**2. Binary Format**
- The `.am` extension stands for "**Automerge**"
- It's a compact binary format (not human-readable text)
- Contains the full operation history compressed efficiently

**3. CRDT Magic - Conflict-Free Merging**

When you have two different `doc.am` files that diverged:
```
Laptop A's doc.am: "Hello World" (edited offline)
Laptop B's doc.am: "Hello Everyone" (edited offline)
```

Automerge can **merge them without conflicts** because each file contains:
- Which character was inserted when
- Who inserted it (which replica/user)
- The causal order of all operations

**4. Inside a `doc.am` file (simplified):**
```
[Header]
ActorID: user-123
Operations: [
  { op: "insert", pos: 0, char: "H", timestamp: 1000, actor: user-123 }
  { op: "insert", pos: 1, char: "e", timestamp: 1001, actor: user-123 }
  { op: "insert", pos: 2, char: "l", timestamp: 1050, actor: user-456 }
  { op: "delete", pos: 2, timestamp: 1100, actor: user-123 }
  ...
]
```

### Current Demo vs. Real Automerge

**What the current demo does WRONG:**
```
Server has: doc.am (ONE shared file)
Client 1 → POST "Hello World" → overwrites doc.am
Client 2 → POST "Hello Everyone" → overwrites doc.am  ❌ LOST DATA!
```

**What real Automerge should do:**
```
Laptop A has: doc_A.am (Alice's personal copy)
Laptop B has: doc_B.am (Bob's personal copy)
Server has: doc_server.am (server's copy)

Alice edits offline → doc_A.am grows with new operations
Bob edits offline → doc_B.am grows with new operations

When they connect:
- Alice syncs → Server merges doc_A.am + doc_server.am = new doc_server.am
- Bob syncs → Server merges doc_B.am + doc_server.am = final doc_server.am
- Both get the merged result ✅ NO DATA LOSS!
```

### Why Sync Deltas (Not Full Files)?

Sending entire `doc.am` files every edit would be:
- ❌ Huge bandwidth (file grows with every edit ever made)
- ❌ Slow performance

Instead, Automerge syncs **deltas** (just new operations):
- ✅ "I added 5 characters at position 10"
- ✅ Minimal bandwidth
- ✅ Fast real-time sync

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

## Proper Data Modeling (Critical!)

### Current Demo Uses Wrong Automerge API

**Current (WRONG) approach:**
```rust
// lib.rs - Using string replacement
doc.put(&automerge::ROOT, "text", "whole string value")  // ❌
```

**Should use Automerge.Text type:**
```rust
// Create a Text object for collaborative editing
let text_id = doc.put_object(&automerge::ROOT, "text", ObjType::Text)?;
// Then insert/delete individual characters
doc.splice_text(&text_id, 0, 0, "Hello")?;
```

### Why This Matters

- **Text type**: Character-level CRDT that merges concurrent edits properly
- **String replacement**: Just overwrites, NO conflict resolution
- Without `Automerge.Text`, we're not using Automerge's core functionality!

### Document Structure

Automerge documents are JSON-like:
```json
{
  "text": <Automerge.Text object>,  // NOT a plain string!
  "metadata": {
    "title": "My Document",         // Immutable strings OK here
    "lastModified": 1234567890
  }
}
```

**Reference:** https://automerge.org/docs/cookbook/modeling-data/

## TODO List

### Phase 0: Fix Data Model (BLOCKING!)
- [ ] **Replace string-based text with Automerge.Text type**
  - [ ] Update Rust lib.rs to use `ObjType::Text` instead of string value
  - [ ] Add `am_text_insert(pos, char)` and `am_text_delete(pos, len)` exports
  - [ ] Update `am_get_text()` to read from Text object (not plain string)
- [ ] **Test character-level operations work correctly**
- [ ] **Verify text properly stored in Text CRDT, not as string**

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
