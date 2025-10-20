# Automerge API Mapping: Rust → WASI → Wazero → Go

**Purpose:** This document maps the Automerge Rust API to our WASI wrapper and Go implementation. It serves as the authoritative reference for AI agents and developers working with Automerge via Wazero.

**Status:** Living document - updated as the API evolves
**Last Updated:** 2025-10-20
**Automerge Version:** 0.7.0 (tag: `rust/automerge@0.7.0`)

---

## Table of Contents

1. [Quick Answer: Do We Have 1:1 Go↔Rust API?](#quick-answer)
2. [Architecture Layers](#architecture-layers)
3. [Complete Automerge Rust API Reference](#complete-automerge-rust-api-reference)
4. [Current WASI Exports](#current-wasi-exports)
5. [Current Go Implementation](#current-go-implementation)
6. [API Coverage Matrix](#api-coverage-matrix)
7. [How to Add New API Functions](#how-to-add-new-api-functions)
8. [Wazero FFI Patterns](#wazero-ffi-patterns)
9. [Future Roadmap](#future-roadmap)
10. [References](#references)

---

## Quick Answer

**"Do we have a 1:1 API in Go that maps to the Rust API?"**

### NO - By Design

| Layer | Methods | Coverage | Purpose |
|-------|---------|----------|---------|
| **Automerge Rust Core** | ~60+ | 100% (reference) | Full CRDT implementation |
| **WASI Wrapper (C ABI)** | 11 | ~18% | Minimal FFI bridge |
| **Go Server** | 12 | ~20% | HTTP/SSE/business logic |

**This is correct.** You expose only what's needed for collaborative text editing with a clean WASM boundary.

---

## Architecture Layers

```
┌────────────────────────────────────────────────────────────────┐
│                   Layer 4: HTTP/Application                     │
│  • REST API (GET/POST /api/text, /api/merge, /api/doc)        │
│  • SSE broadcasting (/api/stream)                              │
│  • Business logic, routing, persistence                        │
│                         Go Code                                 │
│                   (go/cmd/server/main.go)                       │
└──────────────────────┬─────────────────────────────────────────┘
                       │ HTTP Handlers call Go methods
                       │
┌──────────────────────▼─────────────────────────────────────────┐
│                   Layer 3: Go FFI Wrappers                      │
│  • getText(ctx) → string                                       │
│  • setText(ctx, text) → error                                  │
│  • mergeDocument(ctx, []byte) → error                          │
│  • saveDocument(ctx) → error                                   │
│                         Go Code                                 │
│              Wraps wazero.Module calls                          │
└──────────────────────┬─────────────────────────────────────────┘
                       │ wazero FFI (memory copy + function call)
                       │
┌──────────────────────▼─────────────────────────────────────────┐
│                   Layer 2: WASI C ABI Exports                   │
│  • am_alloc(size) → *mut u8                                    │
│  • am_free(ptr, size)                                          │
│  • am_init() → i32                                             │
│  • am_text_splice(pos, del, ptr, len) → i32                   │
│  • am_get_text_len() → u32                                     │
│  • am_get_text(ptr) → i32                                      │
│  • am_save_len() → u32                                         │
│  • am_save(ptr) → i32                                          │
│  • am_load(ptr, len) → i32                                     │
│  • am_merge(ptr, len) → i32                                    │
│                    Rust WASM Module                             │
│              (rust/automerge_wasi/src/lib.rs)                   │
│           Compiled to wasm32-wasip1 target                      │
└──────────────────────┬─────────────────────────────────────────┘
                       │ Internal Rust API calls
                       │
┌──────────────────────▼─────────────────────────────────────────┐
│                Layer 1: Automerge Rust Core API                 │
│  • Traits: ReadDoc, Transactable                               │
│  • Types: AutoCommit, Automerge, ObjId, Prop, Value            │
│  • Full CRDT implementation (text, maps, lists, counters)      │
│  • Sync protocol (SyncState, SyncMessage)                      │
│  • Persistence (save, load, load_incremental)                  │
│  • History (heads, changes, patches)                           │
│                    Rust Library                                 │
│           (.src/automerge/rust/automerge/)                      │
└────────────────────────────────────────────────────────────────┘
```

---

## Complete Automerge Rust API Reference

This section documents the **full Automerge Rust API** from version 0.7.0. Use this as a reference when adding new WASI exports.

**Source:** `.src/automerge/rust/automerge/src/`

### Core Types

```rust
// Main document types
pub struct Automerge { ... }      // Low-level, manual transactions
pub struct AutoCommit { ... }     // High-level, auto-transactions (USED IN OUR WASI WRAPPER)
pub const ROOT: ObjId;            // Root object ID

// Object/value types
pub enum ObjType { Map, List, Text }
pub struct ObjId { ... }          // Object ID (internal)
pub struct ExId { ... }           // External ID (for API users)
pub enum Prop { Map(String), Seq(usize) }
pub enum Value { Object(ObjType), Scalar(ScalarValue) }
pub enum ScalarValue {
    Bytes(Vec<u8>),
    Str(String),
    Int(i64),
    Uint(u64),
    F64(f64),
    Counter(i64),
    Timestamp(i64),
    Boolean(bool),
    Null,
}

// Change/history types
pub struct ChangeHash([u8; 32]);
pub struct Change { ... }
pub struct ActorId(Vec<u8>);
```

### Trait: `ReadDoc` (Reading Values)

**Purpose:** Read values from an Automerge document at current state or historical point.

**Source:** `.src/automerge/rust/automerge/src/read.rs`

```rust
pub trait ReadDoc {
    // Get value at a property/key
    fn get<O: AsRef<ExId>, P: Into<Prop>>(
        &self,
        obj: O,
        prop: P
    ) -> Result<Option<(Value, ExId)>, AutomergeError>;

    fn get_at<O: AsRef<ExId>, P: Into<Prop>>(
        &self,
        obj: O,
        prop: P,
        heads: &[ChangeHash]
    ) -> Result<Option<(Value, ExId)>, AutomergeError>;

    // Get all conflicting values at a key
    fn get_all<O: AsRef<ExId>, P: Into<Prop>>(
        &self,
        obj: O,
        prop: P
    ) -> Result<Vec<(Value, ExId)>, AutomergeError>;

    // Object metadata
    fn object_type<O: AsRef<ExId>>(&self, obj: O) -> Result<ObjType, AutomergeError>;
    fn length<O: AsRef<ExId>>(&self, obj: O) -> usize;
    fn length_at<O: AsRef<ExId>>(&self, obj: O, heads: &[ChangeHash]) -> usize;

    // Iteration
    fn keys<O: AsRef<ExId>>(&self, obj: O) -> Keys<'_>;
    fn keys_at<O: AsRef<ExId>>(&self, obj: O, heads: &[ChangeHash]) -> Keys<'_>;
    fn values<O: AsRef<ExId>>(&self, obj: O) -> Values<'_>;
    fn values_at<O: AsRef<ExId>>(&self, obj: O, heads: &[ChangeHash]) -> Values<'_>;
    fn map_range<'a, O: AsRef<ExId>, R: RangeBounds<String> + 'a>(
        &'a self,
        obj: O,
        range: R
    ) -> MapRange<'a>;
    fn list_range<O: AsRef<ExId>, R: RangeBounds<usize>>(
        &self,
        obj: O,
        range: R
    ) -> ListRange<'_>;

    // Text operations (READ)
    fn text<O: AsRef<ExId>>(&self, obj: O) -> Result<String, AutomergeError>;
    fn text_at<O: AsRef<ExId>>(
        &self,
        obj: O,
        heads: &[ChangeHash]
    ) -> Result<String, AutomergeError>;

    // Marks (rich text formatting)
    fn marks<O: AsRef<ExId>>(&self, obj: O) -> Result<Vec<Mark>, AutomergeError>;
    fn marks_at<O: AsRef<ExId>>(
        &self,
        obj: O,
        heads: &[ChangeHash]
    ) -> Result<Vec<Mark>, AutomergeError>;
    fn get_marks<O: AsRef<ExId>>(
        &self,
        obj: O,
        index: usize
    ) -> Result<MarkSet, AutomergeError>;

    // Spans (text + blocks)
    fn spans<O: AsRef<ExId>>(&self, obj: O) -> Spans<'_>;

    // Parents/tree navigation
    fn parents<O: AsRef<ExId>>(&self, obj: O) -> Result<Parents<'_>, AutomergeError>;
    fn parents_at<O: AsRef<ExId>>(
        &self,
        obj: O,
        heads: &[ChangeHash]
    ) -> Result<Parents<'_>, AutomergeError>;
}
```

### Trait: `Transactable` (Mutating Values)

**Purpose:** Mutate an Automerge document within a transaction.

**Source:** `.src/automerge/rust/automerge/src/transaction/transactable.rs`

```rust
pub trait Transactable: ReadDoc {
    // Get pending ops count
    fn pending_ops(&self) -> usize;

    // MAP operations
    fn put<O: AsRef<ExId>, P: Into<Prop>, V: Into<ScalarValue>>(
        &mut self,
        obj: O,
        prop: P,
        value: V
    ) -> Result<(), AutomergeError>;

    fn put_object<O: AsRef<ExId>, P: Into<Prop>>(
        &mut self,
        obj: O,
        prop: P,
        object: ObjType
    ) -> Result<ExId, AutomergeError>;

    // LIST operations
    fn insert<O: AsRef<ExId>, V: Into<ScalarValue>>(
        &mut self,
        obj: O,
        index: usize,
        value: V
    ) -> Result<(), AutomergeError>;

    fn insert_object<O: AsRef<ExId>>(
        &mut self,
        obj: O,
        index: usize,
        object: ObjType
    ) -> Result<ExId, AutomergeError>;

    fn splice<O: AsRef<ExId>, V: IntoIterator<Item = ScalarValue>>(
        &mut self,
        obj: O,
        pos: usize,
        del: isize,
        vals: V
    ) -> Result<(), AutomergeError>;

    // TEXT operations (CRITICAL - USED IN OUR WASI WRAPPER)
    fn splice_text<O: AsRef<ExId>>(
        &mut self,
        obj: O,
        pos: usize,
        del: isize,
        text: &str
    ) -> Result<(), AutomergeError>;

    fn update_text<S: AsRef<str>>(
        &mut self,
        obj: &ExId,
        new_text: S
    ) -> Result<(), AutomergeError>;

    // MARKS (rich text formatting)
    fn mark<O: AsRef<ExId>>(
        &mut self,
        obj: O,
        mark: Mark,
        expand: ExpandMark
    ) -> Result<(), AutomergeError>;

    fn unmark<O: AsRef<ExId>>(
        &mut self,
        obj: O,
        key: &str,
        start: usize,
        end: usize,
        expand: ExpandMark
    ) -> Result<(), AutomergeError>;

    // BLOCKS (rich text structure)
    fn split_block<O>(&mut self, obj: O, index: usize) -> Result<ExId, AutomergeError>
    where O: AsRef<ExId>;

    fn join_block<O: AsRef<ExId>>(&mut self, text: O, index: usize) -> Result<(), AutomergeError>;

    fn replace_block<O>(&mut self, text: O, index: usize) -> Result<ExId, AutomergeError>
    where O: AsRef<ExId>;

    fn update_spans<O: AsRef<ExId>, I: IntoIterator<Item = Span>>(
        &mut self,
        text: O,
        config: UpdateSpansConfig,
        new_text: I
    ) -> Result<(), AutomergeError>;

    // COUNTER operations
    fn increment<O: AsRef<ExId>, P: Into<Prop>>(
        &mut self,
        obj: O,
        prop: P,
        value: i64
    ) -> Result<(), AutomergeError>;

    // DELETE
    fn delete<O: AsRef<ExId>, P: Into<Prop>>(
        &mut self,
        obj: O,
        prop: P
    ) -> Result<(), AutomergeError>;

    // Transaction metadata
    fn base_heads(&self) -> Vec<ChangeHash>;
}
```

### `AutoCommit` Methods (High-Level API)

**Purpose:** Main document type with automatic transaction management.

**Source:** `.src/automerge/rust/automerge/src/autocommit.rs`

**Note:** `AutoCommit` implements both `ReadDoc` and `Transactable`, plus these additional methods:

```rust
impl AutoCommit {
    // Creation
    pub fn new() -> Self;
    pub fn new_with_encoding(encoding: TextEncoding) -> Self;

    // Actor management
    pub fn with_actor(mut self, actor: ActorId) -> Self;
    pub fn set_actor(&mut self, actor: ActorId) -> &mut Self;
    pub fn get_actor(&self) -> &ActorId;

    // Persistence (USED IN OUR WASI WRAPPER)
    pub fn save(&mut self) -> Vec<u8>;
    pub fn load(data: &[u8]) -> Result<Self, AutomergeError>;
    pub fn load_with(data: &[u8], options: LoadOptions) -> Result<Self, AutomergeError>;
    pub fn load_incremental(&mut self, data: &[u8]) -> Result<usize, AutomergeError>;

    // Merging (USED IN OUR WASI WRAPPER)
    pub fn merge(&mut self, other: &mut Self) -> Result<Vec<ChangeHash>, AutomergeError>;
    pub fn merge_with<O>(&mut self, other: &mut O, options: MergeOptions)
        -> Result<Vec<ChangeHash>, AutomergeError>
    where O: ReadDoc + Send + Sync;

    // History/heads
    pub fn get_heads(&mut self) -> Vec<ChangeHash>;
    pub fn get_changes(&self, have_deps: &[ChangeHash]) -> Vec<&Change>;
    pub fn get_change_by_hash(&self, hash: &ChangeHash) -> Option<&Change>;

    // Forking
    pub fn fork(&self) -> Self;
    pub fn fork_at(&self, heads: &[ChangeHash]) -> Result<Self, AutomergeError>;

    // Patches (for UI updates)
    pub fn make_patches(&self, patch_log: &mut PatchLog) -> Vec<Patch>;
    pub fn current_state(&self) -> Vec<Patch>;

    // Commit management
    pub fn commit(&mut self) -> ChangeHash;
    pub fn commit_with(&mut self, opts: CommitOptions) -> ChangeHash;
    pub fn empty_commit(&mut self, opts: CommitOptions) -> ChangeHash;
    pub fn rollback(&mut self) -> usize;

    // Sync protocol (FOR FUTURE M1 MILESTONE)
    pub fn sync(&mut self) -> sync::SyncDoc<'_, Self>;
}
```

### Sync Protocol (Future M1)

**Purpose:** Efficient delta-based synchronization between peers.

**Source:** `.src/automerge/rust/automerge/src/sync.rs`

```rust
pub struct SyncState { ... }
pub struct SyncMessage { ... }

pub trait SyncDoc {
    fn generate_sync_message(&self, state: &mut SyncState) -> Option<SyncMessage>;
    fn receive_sync_message(
        &mut self,
        state: &mut SyncState,
        message: SyncMessage
    ) -> Result<(), AutomergeError>;
    fn encode_sync_message(message: &SyncMessage) -> Vec<u8>;
    fn decode_sync_message(bytes: &[u8]) -> Result<SyncMessage, AutomergeError>;
}
```

---

## Current WASI Exports

**File:** `rust/automerge_wasi/src/lib.rs`

These are the **only** functions exposed from Rust to Go via the WASM module.

### Memory Management

```rust
#[no_mangle]
pub extern "C" fn am_alloc(size: usize) -> *mut u8
```
- **Purpose:** Allocate memory in WASM linear memory for Go→Rust data transfer
- **Returns:** Pointer to allocated buffer, or `null` on failure
- **Go usage:** Before writing data to WASM (e.g., text content, snapshots)
- **Must call:** `am_free()` when done

```rust
#[no_mangle]
pub extern "C" fn am_free(ptr: *mut u8, size: usize)
```
- **Purpose:** Free memory allocated by `am_alloc`
- **Parameters:** Same pointer and size from `am_alloc`
- **Go usage:** Always called in `defer` after `am_alloc`

### Document Lifecycle

```rust
#[no_mangle]
pub extern "C" fn am_init() -> i32
```
- **Purpose:** Initialize a new `AutoCommit` document with a Text object at `ROOT["content"]`
- **Returns:** `0` on success, `<0` on error
- **Internal:**
  - Creates `AutoCommit::new()`
  - Calls `doc.put_object(ROOT, "content", ObjType::Text)`
  - Stores in `thread_local! { static DOC }`
  - Stores text object ID in `thread_local! { static TEXT_OBJ_ID }`

### Text Operations

```rust
#[no_mangle]
pub extern "C" fn am_text_splice(
    pos: usize,
    del_count: i64,
    insert_ptr: *const u8,
    insert_len: usize
) -> i32
```
- **Purpose:** Perform proper Text CRDT splice operation
- **Parameters:**
  - `pos`: Character position to start splice
  - `del_count`: Number of characters to delete (can be 0)
  - `insert_ptr`: Pointer to UTF-8 text to insert (can be null if `insert_len == 0`)
  - `insert_len`: Byte length of text to insert
- **Returns:** `0` on success, `<0` on error
- **Internal:** Calls `doc.splice_text(&text_obj_id, pos, del_count, insert_text)`
- **This is the PROPER way** to edit text CRDTs (not `am_set_text`)

```rust
#[no_mangle]
pub extern "C" fn am_set_text(ptr: *const u8, len: usize) -> i32
```
- **Status:** **DEPRECATED** - Use `am_text_splice` instead
- **Purpose:** Replace entire text content (inefficient, poor merging)
- **Parameters:** Pointer to UTF-8 text and byte length
- **Returns:** `0` on success, `<0` on error
- **Internal:**
  - Gets current text length
  - Calls `am_text_splice(0, current_len, ptr, len)`
  - Deletes all, then inserts new text
- **Why deprecated:** Destroys fine-grained CRDT history

```rust
#[no_mangle]
pub extern "C" fn am_get_text_len() -> u32
```
- **Purpose:** Get byte length of current text content
- **Returns:** UTF-8 byte length (not character count!)
- **Go usage:** Call before `am_alloc` to size the buffer for `am_get_text`

```rust
#[no_mangle]
pub extern "C" fn am_get_text(ptr_out: *mut u8) -> i32
```
- **Purpose:** Copy text content to provided buffer
- **Parameters:** Pointer to buffer (must be allocated via `am_alloc`)
- **Returns:** `0` on success, `<0` on error
- **Go usage:**
  1. Call `am_get_text_len()` → `textLen`
  2. Call `am_alloc(textLen)` → `ptr`
  3. Call `am_get_text(ptr)`
  4. Read from WASM memory at `ptr` for `textLen` bytes
  5. Call `am_free(ptr, textLen)`

### Persistence

```rust
#[no_mangle]
pub extern "C" fn am_save_len() -> u32
```
- **Purpose:** Get byte size of serialized document
- **Returns:** Size of binary snapshot
- **Internal:** Calls `doc.save().len()`

```rust
#[no_mangle]
pub extern "C" fn am_save(ptr_out: *mut u8) -> i32
```
- **Purpose:** Save document to binary format
- **Parameters:** Pointer to buffer (allocated via `am_alloc`)
- **Returns:** `0` on success, `<0` on error
- **Internal:** Calls `doc.save()` and copies to buffer
- **Format:** Automerge binary format (includes full CRDT history)

```rust
#[no_mangle]
pub extern "C" fn am_load(ptr: *const u8, len: usize) -> i32
```
- **Purpose:** Load document from binary snapshot
- **Parameters:** Pointer to snapshot data and byte length
- **Returns:** `0` on success, `<0` on error
- **Internal:**
  - Calls `AutoCommit::load(slice)`
  - Finds text object ID at `ROOT["content"]`
  - Stores both in thread-local storage
- **Replaces** current document entirely

### Merging

```rust
#[no_mangle]
pub extern "C" fn am_merge(other_ptr: *const u8, other_len: usize) -> i32
```
- **Purpose:** Merge another document into current document (CRDT magic!)
- **Parameters:** Pointer to other document's binary snapshot and length
- **Returns:** `0` on success, `<0` on error
- **Internal:**
  - Loads other document: `AutoCommit::load(other_slice)`
  - Merges: `doc.merge(&mut other_doc.fork())`
  - Updates text object ID (may change after merge)
- **CRDT guarantee:** Conflict-free merge, deterministic result

---

## Current Go Implementation

**File:** `go/cmd/server/main.go`

### Server Struct

```go
type Server struct {
    runtime wazero.Runtime      // Wazero runtime
    module  wazero.CompiledModule // Compiled WASM module
    modInst api.Module           // Instantiated module
    mu      sync.RWMutex         // Protects document access
    clients []chan string        // SSE clients
}
```

### Go Methods (FFI Wrappers)

These methods wrap the WASI exports and handle memory management.

#### `getText(ctx context.Context) (string, error)`

**Purpose:** Get current text from document

**Steps:**
1. Call `am_get_text_len()` → `textLen`
2. Call `am_alloc(textLen)` → `ptr`
3. Call `am_get_text(ptr)`
4. Read WASM memory at `ptr` for `textLen` bytes
5. Call `am_free(ptr, textLen)`
6. Return string

**Wazero calls:**
```go
modInst.ExportedFunction("am_get_text_len").Call(ctx)
modInst.ExportedFunction("am_alloc").Call(ctx, textLen)
modInst.ExportedFunction("am_get_text").Call(ctx, ptr)
modInst.Memory().Read(ptr, textLen)
modInst.ExportedFunction("am_free").Call(ctx, ptr, textLen)
```

#### `setText(ctx context.Context, text string) error`

**Purpose:** Replace entire text (uses deprecated `am_set_text`)

**Steps:**
1. Convert string to `[]byte`
2. Call `am_alloc(len)` → `ptr`
3. Write bytes to WASM memory at `ptr`
4. Call `am_set_text(ptr, len)`
5. Call `am_free(ptr, len)`
6. Call `saveDocument(ctx)` to persist

**Note:** Should be refactored to use `am_text_splice` for proper CRDT operations

#### `saveDocument(ctx context.Context) error`

**Purpose:** Persist document to `doc.am` file

**Steps:**
1. Call `am_save_len()` → `saveLen`
2. Call `am_alloc(saveLen)` → `ptr`
3. Call `am_save(ptr)`
4. Read WASM memory at `ptr` for `saveLen` bytes
5. Call `am_free(ptr, saveLen)`
6. Write bytes to file: `os.WriteFile("doc.am", data, 0644)`

**File location:** Configured by `STORAGE_DIR` env var (default: `./`)

#### `saveDocumentToBytes(ctx context.Context) ([]byte, error)`

**Purpose:** Get snapshot as bytes without writing to file

**Same as `saveDocument` but returns bytes instead of writing to disk**

**Used by:** `GET /api/doc` endpoint

#### `loadDocument(ctx context.Context, data []byte) error`

**Purpose:** Load document from binary snapshot

**Steps:**
1. Call `am_alloc(len(data))` → `ptr`
2. Write `data` to WASM memory at `ptr`
3. Call `am_load(ptr, len)`
4. Call `am_free(ptr, len)`

**Called:** On startup to restore from `doc.am`

#### `mergeDocument(ctx context.Context, otherDoc []byte) error`

**Purpose:** Merge another document into this one (CRDT magic!)

**Steps:**
1. Call `am_alloc(len(otherDoc))` → `ptr`
2. Write `otherDoc` to WASM memory at `ptr`
3. Call `am_merge(ptr, len)`
4. Call `am_free(ptr, len)`

**Used by:** `POST /api/merge` endpoint (for 2-laptop demo)

#### `initializeDocument(ctx context.Context) error`

**Purpose:** Initialize or load document on startup

**Logic:**
```go
if doc.am exists:
    data = read("doc.am")
    loadDocument(ctx, data)
else:
    call am_init() // Creates new document
```

### HTTP API

#### `GET /api/text`
- **Handler:** `handleText`
- **Response:** `text/plain` - current document text
- **Calls:** `s.getText(ctx)`

#### `POST /api/text`
- **Handler:** `handleText`
- **Body:** `{"text":"..."}`
- **Calls:** `s.setText(ctx, payload.Text)`
- **Then:** `s.broadcast(text)` to SSE clients
- **Response:** `204 No Content`

#### `GET /api/stream`
- **Handler:** `handleStream`
- **Response:** `text/event-stream` (SSE)
- **Events:**
  - `snapshot` (on connect): Current text
  - `update` (on POST): New text after edit
- **Format:** `event: snapshot\ndata: {"text":"..."}\n\n`

#### `POST /api/merge`
- **Handler:** `handleMerge`
- **Body:** Raw binary (`application/octet-stream`) - another `doc.am` file
- **Calls:** `s.mergeDocument(ctx, otherDoc)`
- **Then:** `s.broadcast(newText)`
- **Response:** `200 OK` with merged text

#### `GET /api/doc`
- **Handler:** `handleDoc`
- **Response:** `application/octet-stream` - current `doc.am` snapshot
- **Calls:** `s.saveDocumentToBytes(ctx)`
- **Filename:** `Content-Disposition: attachment; filename="<userID>-doc.am"`

#### `GET /`
- **Handler:** `handleUI`
- **Response:** `text/html` - serves `ui/ui.html`

---

## API Coverage Matrix

| Automerge Rust Feature | Rust API Method | WASI Export | Go Method | HTTP API | Status |
|------------------------|-----------------|-------------|-----------|----------|--------|
| **Document Lifecycle** |||||
| Create new document | `AutoCommit::new()` | `am_init()` | `initializeDocument()` | On startup | ✅ |
| **Text Operations** |||||
| Read text | `doc.text(&obj)` | `am_get_text_len()`, `am_get_text()` | `getText()` | `GET /api/text` | ✅ |
| Splice text (proper CRDT) | `doc.splice_text(&obj, pos, del, text)` | `am_text_splice()` | ❌ Not wrapped yet | ❌ | ⚠️ **Partial** |
| Replace all text (deprecated) | N/A (wrapper-only) | `am_set_text()` | `setText()` | `POST /api/text` | ✅ But deprecated |
| **Persistence** |||||
| Save snapshot | `doc.save()` | `am_save_len()`, `am_save()` | `saveDocument()` | `GET /api/doc` | ✅ |
| Load snapshot | `AutoCommit::load()` | `am_load()` | `loadDocument()` | On startup | ✅ |
| **Merging** |||||
| Merge documents | `doc.merge(&mut other)` | `am_merge()` | `mergeDocument()` | `POST /api/merge` | ✅ |
| **Rich Text** |||||
| Add marks (bold, italic) | `doc.mark(&obj, mark, expand)` | ❌ | ❌ | ❌ | ❌ Missing |
| Remove marks | `doc.unmark(&obj, key, start, end, expand)` | ❌ | ❌ | ❌ | ❌ Missing |
| Get marks | `doc.marks(&obj)` | ❌ | ❌ | ❌ | ❌ Missing |
| Split block | `doc.split_block(&obj, index)` | ❌ | ❌ | ❌ | ❌ Missing |
| **Maps/Objects** |||||
| Put value | `doc.put(&obj, key, value)` | ❌ | ❌ | ❌ | ❌ Missing |
| Put object | `doc.put_object(&obj, key, objtype)` | ⚠️ Only at init | ❌ | ❌ | ❌ Missing |
| Get value | `doc.get(&obj, key)` | ❌ | ❌ | ❌ | ❌ Missing |
| Delete key | `doc.delete(&obj, key)` | ❌ | ❌ | ❌ | ❌ Missing |
| **Lists** |||||
| Insert at index | `doc.insert(&obj, index, value)` | ❌ | ❌ | ❌ | ❌ Missing |
| Splice list | `doc.splice(&obj, pos, del, vals)` | ❌ | ❌ | ❌ | ❌ Missing |
| **Counters** |||||
| Increment counter | `doc.increment(&obj, key, value)` | ❌ | ❌ | ❌ | ❌ Missing |
| **Sync Protocol** (M1 Milestone) |||||
| Generate sync message | `doc.sync().generate_sync_message(&state)` | ❌ Planned: `am_sync_gen()` | ❌ | ❌ | 🚧 Roadmap |
| Receive sync message | `doc.sync().receive_sync_message(&state, msg)` | ❌ Planned: `am_sync_recv()` | ❌ | ❌ | 🚧 Roadmap |
| **History** |||||
| Get heads | `doc.get_heads()` | ❌ | ❌ | ❌ | ❌ Missing |
| Get changes | `doc.get_changes(&have_deps)` | ❌ | ❌ | ❌ | ❌ Missing |

**Legend:**
- ✅ Fully implemented
- ⚠️ Partially implemented or deprecated
- ❌ Not implemented
- 🚧 Planned for future milestone

**Current Coverage:** ~18% of Automerge Rust API

---

## How to Add New API Functions

Follow these steps to expose a new Automerge feature via Wazero.

### Example: Adding `am_get_heads()` to get document history

#### Step 1: Add WASI Export (Rust)

**File:** `rust/automerge_wasi/src/lib.rs`

```rust
// Get the number of heads
#[no_mangle]
pub extern "C" fn am_get_heads_count() -> u32 {
    DOC.with(|doc_cell| {
        let mut doc_opt = doc_cell.borrow_mut();
        let doc = match doc_opt.as_mut() {
            Some(d) => d,
            None => return 0,
        };

        let heads = doc.get_heads();
        heads.len() as u32
    })
}

// Get heads as JSON array of hex strings
// Caller must allocate buffer via am_alloc
#[no_mangle]
pub extern "C" fn am_get_heads(ptr_out: *mut u8) -> i32 {
    if ptr_out.is_null() {
        return -1;
    }

    DOC.with(|doc_cell| {
        let mut doc_opt = doc_cell.borrow_mut();
        let doc = match doc_opt.as_mut() {
            Some(d) => d,
            None => return -2,
        };

        let heads = doc.get_heads();

        // Serialize to JSON
        let json = serde_json::to_string(&heads)
            .unwrap_or_else(|_| "[]".to_string());

        let bytes = json.as_bytes();
        unsafe {
            std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
        }
        0
    })
}
```

#### Step 2: Add Go Wrapper

**File:** `go/cmd/server/main.go`

```go
func (s *Server) getHeads(ctx context.Context) ([]string, error) {
    // Get count
    getCountFn := s.modInst.ExportedFunction("am_get_heads_count")
    if getCountFn == nil {
        return nil, fmt.Errorf("am_get_heads_count function not found")
    }

    results, err := getCountFn.Call(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get heads count: %w", err)
    }

    count := uint32(results[0])
    if count == 0 {
        return []string{}, nil
    }

    // Allocate buffer (estimate ~64 bytes per hash in JSON)
    bufSize := count * 64 + 100 // Extra for JSON overhead
    allocFn := s.modInst.ExportedFunction("am_alloc")
    if allocFn == nil {
        return nil, fmt.Errorf("am_alloc function not found")
    }

    results, err = allocFn.Call(ctx, uint64(bufSize))
    if err != nil {
        return nil, fmt.Errorf("failed to allocate memory: %w", err)
    }

    ptr := uint32(results[0])
    if ptr == 0 {
        return nil, fmt.Errorf("allocation failed")
    }

    defer func() {
        freeFn := s.modInst.ExportedFunction("am_free")
        if freeFn != nil {
            freeFn.Call(ctx, uint64(ptr), uint64(bufSize))
        }
    }()

    // Get heads JSON
    getHeadsFn := s.modInst.ExportedFunction("am_get_heads")
    if getHeadsFn == nil {
        return nil, fmt.Errorf("am_get_heads function not found")
    }

    results, err = getHeadsFn.Call(ctx, uint64(ptr))
    if err != nil {
        return nil, fmt.Errorf("failed to get heads: %w", err)
    }

    if results[0] != 0 {
        return nil, fmt.Errorf("am_get_heads returned error: %d", results[0])
    }

    // Read JSON from memory
    mem := s.modInst.Memory()
    if mem == nil {
        return nil, fmt.Errorf("memory not found")
    }

    data, ok := mem.Read(ptr, bufSize)
    if !ok {
        return nil, fmt.Errorf("failed to read memory")
    }

    // Parse JSON
    var heads []string
    if err := json.Unmarshal(data, &heads); err != nil {
        return nil, fmt.Errorf("failed to parse heads JSON: %w", err)
    }

    return heads, nil
}
```

#### Step 3: Add HTTP Endpoint (Optional)

**File:** `go/cmd/server/main.go`

```go
// In main()
http.HandleFunc("/api/heads", s.handleHeads)

// New handler
func (s *Server) handleHeads(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    ctx := r.Context()

    s.mu.RLock()
    heads, err := s.getHeads(ctx)
    s.mu.RUnlock()

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "heads": heads,
        "count": len(heads),
    })
}
```

#### Step 4: Test

```bash
# Rebuild WASM
make build-wasi

# Run server
make run

# Test endpoint
curl http://localhost:8080/api/heads
```

---

## Wazero FFI Patterns

Common patterns for calling WASM functions from Go.

### Pattern 1: Call Simple Function (No Args)

```go
fn := s.modInst.ExportedFunction("am_init")
if fn == nil {
    return fmt.Errorf("function not found")
}

results, err := fn.Call(ctx)
if err != nil {
    return fmt.Errorf("call failed: %w", err)
}

// Check return code
if len(results) > 0 && results[0] != 0 {
    return fmt.Errorf("function returned error: %d", results[0])
}
```

### Pattern 2: Pass Data to WASM (Go → Rust)

```go
data := []byte("Hello, WASM!")
dataLen := uint32(len(data))

// 1. Allocate in WASM
allocFn := s.modInst.ExportedFunction("am_alloc")
results, err := allocFn.Call(ctx, uint64(dataLen))
if err != nil {
    return fmt.Errorf("alloc failed: %w", err)
}
ptr := uint32(results[0])

// 2. Defer free
defer func() {
    freeFn := s.modInst.ExportedFunction("am_free")
    if freeFn != nil {
        freeFn.Call(ctx, uint64(ptr), uint64(dataLen))
    }
}()

// 3. Write to WASM memory
mem := s.modInst.Memory()
if !mem.Write(ptr, data) {
    return fmt.Errorf("failed to write to memory")
}

// 4. Call WASM function with pointer
someFn := s.modInst.ExportedFunction("am_some_function")
results, err = someFn.Call(ctx, uint64(ptr), uint64(dataLen))
```

### Pattern 3: Read Data from WASM (Rust → Go)

```go
// 1. Get size
getLenFn := s.modInst.ExportedFunction("am_get_size")
results, err := getLenFn.Call(ctx)
size := uint32(results[0])

// 2. Allocate in WASM
allocFn := s.modInst.ExportedFunction("am_alloc")
results, err = allocFn.Call(ctx, uint64(size))
ptr := uint32(results[0])

// 3. Defer free
defer func() {
    freeFn := s.modInst.ExportedFunction("am_free")
    if freeFn != nil {
        freeFn.Call(ctx, uint64(ptr), uint64(size))
    }
}()

// 4. Call function to populate buffer
getFn := s.modInst.ExportedFunction("am_get_data")
results, err = getFn.Call(ctx, uint64(ptr))

// 5. Read from WASM memory
mem := s.modInst.Memory()
data, ok := mem.Read(ptr, size)
if !ok {
    return fmt.Errorf("failed to read memory")
}

// 6. Use data
result := string(data)
```

### Pattern 4: Error Handling

All WASI exports return `i32` error codes:

```rust
// Rust conventions
return 0;    // Success
return -1;   // Null pointer
return -2;   // Not initialized
return -3;   // Invalid UTF-8
return -4;   // Operation failed
return -5;   // Other error
```

```go
// Go error checking
results, err := fn.Call(ctx, ...)
if err != nil {
    return fmt.Errorf("WASM call failed: %w", err)
}

errorCode := int32(results[0])
if errorCode != 0 {
    return fmt.Errorf("operation failed with code: %d", errorCode)
}
```

---

## Future Roadmap

### Milestone 1: Sync Protocol (M1)

**Goal:** Replace full-document merging with efficient delta sync.

**New WASI Exports:**
```rust
#[no_mangle]
pub extern "C" fn am_sync_state_init() -> i32;

#[no_mangle]
pub extern "C" fn am_sync_gen_len() -> u32;

#[no_mangle]
pub extern "C" fn am_sync_gen(ptr_out: *mut u8) -> i32;

#[no_mangle]
pub extern "C" fn am_sync_recv(ptr: *const u8, len: usize) -> i32;
```

**Go Integration:**
```go
func (s *Server) generateSyncMessage(ctx context.Context) ([]byte, error)
func (s *Server) receiveSyncMessage(ctx context.Context, msg []byte) error
```

**HTTP API:**
- Keep existing `/api/stream` for SSE
- Change payload from full text to sync messages
- Much more efficient over network

### Milestone 2: Multi-Document (M2)

**New WASI Exports:**
```rust
#[no_mangle]
pub extern "C" fn am_new_doc(doc_id_ptr: *const u8, len: usize) -> i32;

#[no_mangle]
pub extern "C" fn am_select_doc(doc_id_ptr: *const u8, len: usize) -> i32;

#[no_mangle]
pub extern "C" fn am_list_docs(ptr_out: *mut u8) -> i32;
```

**Go Changes:**
- Replace single `DOC` with `map[string]*Document`
- Add `?doc=<id>` query param to all endpoints
- Snapshot files: `data/<docId>.am`

### Milestone 3: NATS Transport (M3)

**Architecture:**
```
Go Server
  ↓ receives NATS message on automerge.sync.<tenant>.<docId>
  ↓ calls am_sync_recv(message)
  ↓ if needed, calls am_sync_gen() for reply
  ↓ publishes reply to NATS
```

**No WASM changes needed!** Go handles NATS, WASM handles sync protocol.

### Milestone 4: Rich Text (M4)

**New WASI Exports:**
```rust
#[no_mangle]
pub extern "C" fn am_mark(
    start: usize,
    end: usize,
    name_ptr: *const u8,
    name_len: usize,
    value_ptr: *const u8,
    value_len: usize
) -> i32;

#[no_mangle]
pub extern "C" fn am_unmark(
    start: usize,
    end: usize,
    name_ptr: *const u8,
    name_len: usize
) -> i32;

#[no_mangle]
pub extern "C" fn am_get_marks_len() -> u32;

#[no_mangle]
pub extern "C" fn am_get_marks(ptr_out: *mut u8) -> i32;
```

---

## References

### Source Files

| Component | Path | Purpose |
|-----------|------|---------|
| **Automerge Rust Core** | `.src/automerge/rust/automerge/src/` | Full CRDT implementation (reference) |
| **ReadDoc trait** | `.src/automerge/rust/automerge/src/read.rs` | Read API |
| **Transactable trait** | `.src/automerge/rust/automerge/src/transaction/transactable.rs` | Mutation API |
| **AutoCommit** | `.src/automerge/rust/automerge/src/autocommit.rs` | High-level document API |
| **Sync protocol** | `.src/automerge/rust/automerge/src/sync.rs` | Sync state machines |
| **WASI Wrapper** | `rust/automerge_wasi/src/lib.rs` | C ABI exports |
| **Go Server** | `go/cmd/server/main.go` | Wazero integration |
| **UI** | `ui/ui.html` | Browser interface |

### Documentation

- **Automerge Docs:** `.src/automerge.github.io/`
- **Agent Knowledge Base:** `AGENT_AUTOMERGE.md` (AI-focused Automerge concepts)
- **Project Instructions:** `CLAUDE.md` (Development workflow, milestones)
- **Setup Source:** Run `make setup-src` to clone Automerge v0.7.0

### External Links

- **Automerge GitHub:** https://github.com/automerge/automerge
- **Automerge Docs:** https://automerge.org/docs
- **Wazero:** https://wazero.io
- **WASI:** https://wasi.dev

---

## Maintenance Notes

**This document should be updated when:**

1. ✅ New WASI export is added → Update "Current WASI Exports" section
2. ✅ New Go method is added → Update "Current Go Implementation" section
3. ✅ New HTTP endpoint is added → Update "Current Go Implementation" section
4. ✅ API coverage changes → Update "API Coverage Matrix"
5. ✅ Automerge version changes → Update version at top, re-check API compatibility
6. ✅ New milestone feature is implemented → Move from "Future Roadmap" to "Current"

**Keep this document synchronized with:**
- `AGENT_AUTOMERGE.md` (concepts and best practices)
- `CLAUDE.md` (milestones and project instructions)
- `rust/automerge_wasi/src/lib.rs` (actual exports)
- `go/cmd/server/main.go` (actual implementation)

---

**Document Version:** 2.0
**Last Updated:** 2025-10-20
**Maintainer:** AI Agent + Human Developers
