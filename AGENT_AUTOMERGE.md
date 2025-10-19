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

### Local Documentation
- [.src/automerge-docs/text.md](.src/automerge-docs/text.md) - Text CRDT API
- [.src/automerge-docs/modeling-data.md](.src/automerge-docs/modeling-data.md) - Data modeling guide

### Official Documentation
- **GitHub Repo:** https://github.com/automerge/automerge.github.io
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
