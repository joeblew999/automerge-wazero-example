# Automerge Knowledge Base for AI Agents

**Purpose:** This document provides comprehensive knowledge about Automerge CRDTs for AI agents working on this project. It distills critical concepts from official Automerge documentation to enable proper implementation.

**Last Updated:** 2025-10-19

---

## Table of Contents

1. [What is Automerge?](#what-is-automerge)
2. [Core Concepts](#core-concepts)
3. [The `doc.am` File](#the-docam-file)
4. [Document Data Model](#document-data-model)
5. [Text CRDT - CRITICAL](#text-crdt---critical)
6. [Data Modeling Best Practices](#data-modeling-best-practices)
7. [Sync Protocol Fundamentals](#sync-protocol-fundamentals)
8. [Common Mistakes to Avoid](#common-mistakes-to-avoid)
9. [Rust Implementation Notes](#rust-implementation-notes)
10. [References](#references)

---

## What is Automerge?

**Automerge is a CRDT (Conflict-free Replicated Data Type) library** that enables:
- **Offline-first applications** - Work without network, sync later
- **Conflict-free merging** - Concurrent edits automatically merge without conflicts
- **Peer-to-peer sync** - No central server required (though can use one)
- **Full history** - Every edit is preserved (like git commits)

### Key Analogy
> **Automerge document = JSON object + Git repository**
> - Like JSON: Map from strings to values (maps, arrays, primitives)
> - Like Git: Full history of all changes, can merge any two versions

---

## Core Concepts

### Documents
- **Unit of change** in Automerge
- Always starts with a **root map** (key-value pairs)
- Has full **change history** (every edit ever made)
- Two documents can **always be merged** without conflicts

### Document URLs
Format: `automerge:2akvofn6L1o4RMUEMQi7qzwRjKWZ` (base58 encoded)
- Each document identified by unique URL
- Applications typically have **many documents**, each with UUID

### Repositories (JavaScript-specific)
- Manages connections to remote peers
- Handles local storage
- Provides `DocHandle` for accessing/modifying documents
- **Note:** Rust doesn't have repository concept - we implement manually

### Sync Protocol
- **Transport-agnostic** - works over any connection
- **Per-document basis** - each doc syncs independently
- Syncs **deltas** (changes), not full documents
- Efficient binary format

---

## The `doc.am` File

**Critical Understanding:** `doc.am` is NOT just current state!

### What's Inside?
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

### Key Properties:
1. **Full operation history** - Every single edit ever made
2. **Binary format** - Compact, not human-readable
3. **Grows with edits** - More history = larger file
4. **CRDT magic** - Contains causal ordering for conflict-free merging

### Proper Architecture:
```
Laptop A has: doc_A.am (Alice's personal copy)
Laptop B has: doc_B.am (Bob's personal copy)
Server has: doc_server.am (server's copy)

Alice edits offline → doc_A.am grows with new operations
Bob edits offline → doc_B.am grows with new operations

When they connect:
- Alice syncs → Server merges doc_A.am + doc_server.am
- Bob syncs → Server merges doc_B.am + doc_server.am
- Both get the merged result ✅ NO DATA LOSS!
```

### Why Sync Deltas?
- ❌ Sending entire `doc.am` every edit = huge bandwidth
- ✅ Sending only **new operations** = minimal bandwidth
- Example: "I added 5 characters at position 10"

---

## Document Data Model

### Supported Types

**Composite Types:**
- **Maps** - String keys → Automerge values (like JSON objects)
- **Lists** - Ordered sequences (RGA CRDT, preserves intent on concurrent edits)
- **Text** - Collaborative text (Peritext CRDT with marks)

**Scalar Types:**
- Numbers (float64, uint, int)
- Booleans
- Strings (two types - see below!)
- Timestamps (milliseconds since epoch)
- Counters (CRDT that adds concurrent ops)
- Byte arrays

### Critical: Two String Types

1. **Collaborative Text (plain string in JS)**
   - Represented as `string` in JavaScript
   - Use `Automerge.splice()` or `Automerge.updateText()` to modify
   - **Character-level CRDT** - merges concurrent changes properly
   - **THIS IS WHAT WE NEED FOR TEXT EDITING!**

2. **Immutable String**
   - Created with `new Automerge.ImmutableString("value")`
   - For non-collaborative data (usernames, IDs, metadata)
   - Simple replacement, no conflict resolution

### Document Structure Example
```json
{
  "text": "<Automerge.Text CRDT object>",  // NOT a plain string!
  "metadata": {
    "title": "My Document",                // ImmutableString OK here
    "author": "Alice",
    "lastModified": 1234567890
  }
}
```

---

## Text CRDT - CRITICAL

### The Problem We Had
**WRONG approach (what current demo does):**
```rust
doc.put(&automerge::ROOT, "text", "whole string value")  // ❌
```
This is just **string replacement** - NO conflict resolution!

**RIGHT approach (what we must do):**
```rust
// Create a Text object for collaborative editing
let text_id = doc.put_object(&automerge::ROOT, "text", ObjType::Text)?;
// Then insert/delete individual characters
doc.splice_text(&text_id, 0, 0, "Hello")?;
```

### Why Text Type Matters
- **Text type**: Character-level CRDT, merges concurrent edits properly
- **String replacement**: Just overwrites, loses concurrent changes
- Without `Automerge.Text`, we're not using Automerge's core functionality!

### JavaScript API - `splice()`

```js
import * as Automerge from "@automerge/automerge"

let doc = Automerge.from({text: "hello world"})

// Fork and make concurrent changes
let forked = Automerge.clone(doc)
forked = Automerge.change(forked, d => {
    // Insert ' wonderful' at index 5, delete 0 chars
    Automerge.splice(d, ["text"], 5, 0, " wonderful")
})

doc = Automerge.change(doc, d => {
    // Insert "Greetings" at start, delete 5 chars ("hello")
    Automerge.splice(d, ["text"], 0, 5, "Greetings")
})

// Merge the changes - NO CONFLICTS!
doc = Automerge.merge(doc, forked)
console.log(doc.text) // "Greetings wonderful world"
```

### JavaScript API - `updateText()`

When you can't get individual keystroke events:

```typescript
const handle: DocHandle<{text: string}> = ...
const input = document.getElementById("input")!

input.value = handle.docSync()!.text!

// On every keystroke
input.oninput = (e) => {
    handle.change((doc) => {
        const newValue: string = e.target.value
        Automerge.updateText(doc, ["text"], newValue)
    })
}
```

**IMPORTANT:** `updateText` works best when called **after every keystroke**. If text changes a lot between calls, the diff won't merge well with concurrent changes.

### Peritext CRDT - Text with Marks

Text also supports **marks** (for formatting):
```
Mark = (start, end, name, value)

Example: Bold from chars 1-5
(1, 5, "bold", true)
```

Currently, mark values must be scalars (no objects). Use JSON strings if needed.

---

## Data Modeling Best Practices

### How Many Documents?

**Rule of thumb:** An Automerge document is best suited to being a **unit of collaboration between a person or small group**.

- ✅ Hundreds of docs should be fine
- ❌ Thousands of docs = high sync overhead (see PushPin example)
- Balance granularity: Like choosing JSON files vs SQLite tables vs rows

### Initializing Document Schema

**Method 1: Sync initial change to all devices**
```js
let doc1 = Automerge.change(Automerge.init(), (doc) => {
  doc.cards = [];
});

// Don't create schema again on doc2!
let doc2 = Automerge.merge(Automerge.init(), doc1);
```

**Method 2: Hard-code initial change (for independent initialization)**
```js
// Create schema once, save it
let doc = Automerge.change(Automerge.init(), (doc) => {
  doc.cards = [];
});
const initChange = Automerge.save(doc);

// Hard-code this byte array into your application
const INIT_DOC = new Uint8Array([133, 111, 74, 131, ...])

// Each device loads from same initial change
let [doc] = Automerge.load(INIT_DOC)
```

This ensures all devices start with **identical actorId/timestamp** for the schema init, so they can merge.

### Versioning & Schema Migration

**Challenge:** Two users might independently perform same migration → must be idempotent!

**Solution:** Hard-code migrations as byte arrays (like initial change):
```js
type DocV1 = { version: 1, cards: Card[] }
type DocV2 = { version: 2, title: Automerge.Text, cards: Card[] }

const migrateV1toV2 = new Uint8Array([133, 111, 74, 131, ...])

let doc = getDocumentFromNetwork()
if (doc.version === 1) {
  [doc] = Automerge.applyChange(doc, [migrateV1toV2])
}
```

### Performance

- Documents hold **entire change histories** (like Git)
- Generally performant for reasonable workloads
- History truncation/GC experiments show **minimal space savings**
- Optimization: Sync multiple docs over single connection
- Rule: Fewer docs in memory = faster startup/sync

---

## Sync Protocol Fundamentals

### Transport-Agnostic Design
- Works over WebSocket, HTTP, NATS, custom transports
- Per-document basis (each doc syncs independently)
- Binary format for efficiency

### Sync Operations (from Rust API)
```rust
// Initialize sync with peer
am_sync_init_peer(peer_id)

// Generate sync message (delta to send)
am_sync_gen() -> Vec<u8>

// Receive sync message (delta from peer)
am_sync_recv(message: &[u8])
```

### Sync Flow
```
Peer A                              Peer B
  |                                    |
  |  am_sync_init_peer(B)              |
  |  am_sync_gen() → msg1              |
  |---------------------------------->|
  |                    am_sync_recv(msg1)
  |                    am_sync_gen() → msg2
  |<----------------------------------|
  |  am_sync_recv(msg2)                |
  |  (may generate more messages)      |
```

### Delta Sync Benefits
- ✅ Only sends **new operations** since last sync
- ✅ Minimal bandwidth usage
- ✅ Fast for real-time collaboration
- ❌ Don't send full `doc.am` file every time!

---

## Common Mistakes to Avoid

### ❌ MISTAKE 1: Using String Replacement Instead of Text CRDT
```rust
// WRONG - No conflict resolution!
doc.put(&automerge::ROOT, "text", full_text_string)

// RIGHT - Character-level CRDT
let text_id = doc.put_object(&automerge::ROOT, "text", ObjType::Text)?;
doc.splice_text(&text_id, pos, delete_len, insert_string)?;
```

### ❌ MISTAKE 2: One Shared Document on Server
```
Browser A → Server (ONE doc.am) ← Browser B  // WRONG!
```
**Problem:** This defeats Automerge's offline-first design!

**Correct:** Each client has own doc.am, syncs deltas with server.

### ❌ MISTAKE 3: Calling updateText Infrequently
```js
// WRONG - Only on blur/submit
input.onblur = () => updateText(doc, ["text"], input.value)

// RIGHT - Every keystroke
input.oninput = () => updateText(doc, ["text"], input.value)
```

### ❌ MISTAKE 4: Independent Schema Initialization
```js
// WRONG - Different actorIds = can't merge!
let doc1 = Automerge.change(Automerge.init(), d => { d.cards = [] })
let doc2 = Automerge.change(Automerge.init(), d => { d.cards = [] })

// RIGHT - Hard-code initial change or sync it
let doc1 = Automerge.load(INIT_CHANGE)
let doc2 = Automerge.load(INIT_CHANGE)
```

### ❌ MISTAKE 5: Sending Full Documents Instead of Deltas
```
// WRONG - Huge bandwidth!
sendToServer(doc.save())

// RIGHT - Send only changes
let syncMessage = doc.generateSyncMessage(peer)
sendToServer(syncMessage)
```

---

## Merge Rules - How Automerge Resolves Conflicts

### Understanding Concurrent Changes

Two concurrent versions of a document = changes since common ancestor:
```
A → B → C → D → E
        ↓
        F → G
```
Common ancestor: C
Concurrent changes: (D, E) and (F, G)

### Map Merge Rules

| Scenario | Result |
|----------|--------|
| A sets key `x`, B sets key `y` (x ≠ y) | Both `x` and `y` in merged map |
| A deletes key `x`, B doesn't touch `x` | `x` removed from merged map |
| A deletes key `x`, B sets new value for `x` | `x` has B's new value |
| Both delete key `x` | `x` deleted from merged map |
| **Both set key `x` to different values** | **Randomly choose one (but all nodes agree)** |

**"Randomly choose"** = Arbitrary but deterministic (based on operation ID)

### List Merge Rules

**Key concept:** Every element has an ID. Operations reference IDs, not indices!

| Scenario | Result |
|----------|--------|
| A and B both insert after index `i` | Arbitrarily choose order, but **preserve insertion order per replica** |
| A deletes element at `i`, B updates element at `i` | Keep B's updated value (delete loses) |
| Both delete element `i` | Remove from merged list |

**Example - Preserving insertion order:**
```
Initial: [a, b]
A inserts [d, e] after b
B inserts [f, g] after b

Result: [a, b, d, e, f, g]  OR  [a, b, f, g, d, e]
(One order chosen, but d→e and f→g order preserved)
```

### Text Merge Rules

**Text uses same logic as lists** (each character = list element)

- Characters merged with list algorithm
- Marks (formatting) merged with [Peritext](https://www.inkandswitch.com/peritext/) algorithm

### Counter Merge Rules

**Simplest case:** Just sum all operations from each node!

```
Node A: increment(5)
Node B: increment(3)
Merged: counter = 8
```

### Conflict Detection

**Only real conflict:** Concurrent updates to same property in same object

```js
let doc1 = Automerge.change(Automerge.init(), doc => { doc.x = 1 })
let doc2 = Automerge.change(Automerge.init(), doc => { doc.x = 2 })
doc1 = Automerge.merge(doc1, doc2)

doc1.x // Either 1 or 2 (deterministic across all nodes)
Automerge.getConflicts(doc1, "x")
// {'1@01234567': 1, '1@89abcdef': 2}  // Both values preserved!
```

**Conflict resolution:**
- Uses **LWW (Last Writer Wins)** based on operation ID (not wall clock!)
- Operation ID = counter + actorId
- All conflicting values accessible via `getConflicts()`
- Next assignment automatically resolves conflict

---

## Storage Model - How doc.am Files are Stored

### Storage Key Format

**Not simple key-value!** Keys are arrays:
```
[<document ID>, <chunk type>, <chunk identifier>]

chunk type: "snapshot" | "incremental"
chunk identifier:
  - snapshot: heads of document at compaction time
  - incremental: hash of change bytes
```

### Example Storage Keys
```
["3RFyJzsLsZ7M", "incremental", "0290cdc2..."]  // Single change
["3RFyJzsLsZ7M", "snapshot", "abc123..."]       // Compacted snapshot
```

### Incremental Changes vs Snapshots

**Incremental change:** A single change (or small set of changes) to document
**Snapshot:** All changes compacted into single binary blob

### Storage Lifecycle

1. **Initial document creation:**
   ```
   ["docId", "incremental", "hash1"]  // Init change
   ```

2. **User makes edits:**
   ```
   ["docId", "incremental", "hash1"]
   ["docId", "incremental", "hash2"]
   ["docId", "incremental", "hash3"]
   ```

3. **Compaction triggers:**
   - Load all incremental changes
   - Merge into single snapshot
   - Save snapshot with heads as identifier
   - Delete **only** incremental changes this process loaded

   ```
   ["docId", "snapshot", "heads_abc"]  // Compacted
   ```

### Concurrent Storage Access

**Challenge:** Multiple processes writing to same document storage

**Solution:** Use document heads as part of key
- If two processes compact simultaneously, they only delete changes they loaded
- Overwriting same snapshot key is safe (same changes = same bytes)
- No locks or transactions required!

```
Process A compacts: sees changes 1,2,3 → writes snapshot, deletes 1,2,3
Process B compacts: sees changes 1,2,3,4 → writes snapshot, deletes 1,2,3,4
Process B's snapshot includes all of A's data → safe!
```

### Storage Adapter Interface

```typescript
abstract class StorageAdapter {
  abstract load(key: StorageKey): Promise<Uint8Array | undefined>
  abstract save(key: StorageKey, data: Uint8Array): Promise<void>
  abstract remove(key: StorageKey): Promise<void>
  abstract loadRange(keyPrefix: StorageKey): Promise<{key, data}[]>
  abstract removeRange(keyPrefix: StorageKey): Promise<void>
}
```

**Why range queries?** Load all chunks for a document: `loadRange(["docId"])`

### Storage Backends

Can implement over:
- ✅ IndexedDB (browser)
- ✅ Local filesystem directory
- ✅ S3 bucket
- ✅ PostgreSQL
- ✅ Any key-value store with range queries

### Loading Multiple Snapshots

**Magic of CRDTs:** If storage contains multiple snapshots (from concurrent processes), loading merges them automatically!

```
Storage:
  ["docId", "snapshot", "heads_A"]  // From tab A
  ["docId", "snapshot", "heads_B"]  // From tab B

On load:
  - Load both snapshots
  - Merge them (CRDT merge)
  - Result includes all changes from A and B!
```

---

## Network Sync Protocol

### Transport-Agnostic Design

Automerge sync works over ANY message-passing channel:
- ✅ WebSockets
- ✅ HTTP (SSE/long-polling)
- ✅ MessageChannel (browser tabs)
- ✅ BroadcastChannel (browser)
- ✅ NATS pub/sub
- ✅ Custom transports

### Sync is Point-to-Point

**Important:** Sync protocol is between **two peers** (not broadcast!)
- Each peer has sync state with every other peer
- Broadcast channels are inefficient (must duplicate messages per peer)

### Sync Message Flow

```
Peer A                              Peer B
  |                                    |
  |  Generate sync message             |
  |  (based on what B knows)           |
  |---------------------------------->|
  |                    Receive message
  |                    Apply changes
  |                    Generate response
  |<----------------------------------|
  |  Receive response                 |
  |  Apply changes                    |
  |  (may generate more messages)     |
```

### Sync State Management

Each peer tracks:
- **What I've sent to peer B**
- **What peer B has acknowledged**
- **What changes I need to send next**

### Efficient Delta Sync

**Don't send:**
- ❌ Full document every time
- ❌ Changes the peer already has

**Do send:**
- ✅ Only new operations since last sync
- ✅ Minimal binary representation
- ✅ Compressed format

### Network Adapter Pattern (JavaScript)

```typescript
class NetworkAdapter {
  // Called when repo wants to send sync message
  send(targetId: PeerId, message: Uint8Array): void

  // Call this when message received from network
  this.emit('message', { from: senderId, data: message })
}
```

### Example: WebSocket Sync

**Server side:**
```typescript
import { WebSocketServer } from "ws"
import { NodeWSServerAdapter } from "@automerge/automerge-repo-network-websocket"

const wss = new WebSocketServer({ port: 8080 })
const adapter = new NodeWSServerAdapter(wss)
```

**Client side:**
```typescript
import { BrowserWebSocketClientAdapter } from "@automerge/automerge-repo-network-websocket"

const adapter = new BrowserWebSocketClientAdapter("ws://localhost:8080")
```

### Sync Over HTTP (Our Use Case)

For HTTP/SSE, we need to implement:
1. **Client → Server:** POST sync messages
2. **Server → Client:** SSE stream for sync messages
3. **Per-client sync state** on server

```go
// Pseudo-code for Go server
type SyncState struct {
    clientSyncStates map[string]*automerge.SyncState
}

func (s *Server) handleSyncMessage(clientId string, msg []byte) {
    syncState := s.clientSyncStates[clientId]

    // Apply incoming sync message
    doc.receiveSyncMessage(syncState, msg)

    // Generate response (if needed)
    if response := doc.generateSyncMessage(syncState); len(response) > 0 {
        s.broadcastToClient(clientId, response)
    }
}
```

---

## Rust Implementation Notes

### Rust API Mapping

**JavaScript → Rust Equivalents:**

| JavaScript | Rust |
|------------|------|
| `Automerge.from({text: "hello"})` | `let doc = AutoCommit::new(); doc.put_object(ROOT, "text", ObjType::Text)` |
| `Automerge.splice(d, ["text"], pos, del, ins)` | `doc.splice_text(&text_id, pos, del, ins)` |
| `Automerge.save(doc)` | `doc.save()` |
| `Automerge.load(bytes)` | `Automerge::load(&bytes)` |
| `Automerge.merge(doc1, doc2)` | `doc1.merge(&mut doc2)` |

### Rust Text CRDT Example
```rust
use automerge::{AutoCommit, ObjType, ROOT};

// Create document with Text object
let mut doc = AutoCommit::new();
let text_id = doc.put_object(ROOT, "text", ObjType::Text)
    .expect("Failed to create text object");

// Insert characters
doc.splice_text(&text_id, 0, 0, "Hello")
    .expect("Failed to insert text");

// Delete characters
doc.splice_text(&text_id, 0, 5, "")
    .expect("Failed to delete text");

// Get text value
let text_obj = doc.get(ROOT, "text").unwrap();
let text_value = doc.text(&text_obj.id()).unwrap();
println!("{}", text_value); // Prints the text
```

### WASI FFI Considerations

For Rust WASI → Go integration:
```rust
// Export C-like functions
#[no_mangle]
pub extern "C" fn am_text_insert(pos: u32, char_ptr: *const u8, len: u32) -> i32

#[no_mangle]
pub extern "C" fn am_text_delete(pos: u32, delete_len: u32) -> i32

#[no_mangle]
pub extern "C" fn am_get_text_len() -> u32

#[no_mangle]
pub extern "C" fn am_get_text(ptr_out: *mut u8) -> i32
```

**Key differences from current implementation:**
- Current: `am_set_text(full_string)` - WRONG!
- Needed: `am_text_insert()` and `am_text_delete()` - character-level ops!

---

## Sync Protocol Fundamentals

### Phase 2 Implementation Requirements

**Rust WASI exports needed:**
```rust
#[no_mangle]
pub extern "C" fn am_sync_init_peer(peer_id_ptr: *const u8, len: u32) -> i32

#[no_mangle]
pub extern "C" fn am_sync_gen_len() -> u32

#[no_mangle]
pub extern "C" fn am_sync_gen(ptr_out: *mut u8) -> i32

#[no_mangle]
pub extern "C" fn am_sync_recv(ptr: *const u8, len: u32) -> i32
```

**Go server changes:**
- Per-client document map (not single shared doc)
- `/api/sync` SSE endpoint for sync messages
- Broadcast deltas (not full text)
- Store per-client `doc.am` files

---

## References

### Local Documentation (Cloned Repo)
- **Full docs repo:** `.src/automerge.github.io/` (git clone of official docs)
- **Text CRDT:** `.src/automerge.github.io/content/docs/reference/documents/text.md`
- **Modeling Data:** `.src/automerge.github.io/content/docs/cookbook/modeling-data.md`
- **Document Model:** `.src/automerge.github.io/content/docs/reference/documents/index.md`
- **Concepts:** `.src/automerge.github.io/content/docs/reference/concepts.md`
- **All docs:** Browse `.src/automerge.github.io/content/docs/` for complete documentation tree

### Official Documentation
- **GitHub Repo:** https://github.com/automerge/automerge.github.io (cloned to `.src/`)
- **Text Documentation:** https://automerge.org/docs/reference/documents/text/
- **Modeling Data:** https://automerge.org/docs/cookbook/modeling-data/
- **Document Model:** https://automerge.org/docs/reference/documents/
- **Core Concepts:** https://automerge.org/docs/reference/concepts/

### Rust Crate
- **docs.rs:** https://docs.rs/automerge/latest/automerge/
- **Crate:** https://crates.io/crates/automerge

### Project-Specific
- [TODO.md](TODO.md) - Implementation roadmap
- [CLAUDE.md](CLAUDE.md) - Project requirements
- [README.md](README.md) - Project overview

---

## Key Takeaways for Agents

1. **Automerge is NOT a key-value store** - It's a CRDT with operation history
2. **Use Text type for collaborative editing** - Not string replacement!
3. **Each client needs own doc.am** - Not one shared server document
4. **Sync deltas, not full documents** - Bandwidth matters
5. **Character-level operations** - `splice_text()` not `put()`
6. **Every document has full history** - Like a Git repo for data
7. **Offline-first design** - Work without network, sync later
8. **Conflict-free merging** - Two docs can ALWAYS merge

**Phase 0 Priority:** Fix data model to use `ObjType::Text` before implementing sync protocol!

---

**Maintained by:** AI Agents working on automerge-wazero-example
**Status:** Living document - update as we learn more
