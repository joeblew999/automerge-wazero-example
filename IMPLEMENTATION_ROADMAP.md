# Automerge WASI Complete Implementation Roadmap

**Goal**: Implement 100% of Automerge 0.7 features with 100% test coverage

**Status**: Working through systematically - no web stuff, pure CRDT implementation

---

## Phase 1: Core Data Types (PRIORITY)

### 1.1 Text CRDT ✅ DONE
- [x] `am_text_splice()` - Insert/delete text
- [x] `am_get_text()` - Read text
- [x] `am_get_text_len()` - Text length
- [x] Tests: text_test.go (15 tests passing)

### 1.2 Map Operations ❌ TODO
**Rust exports needed**:
- `am_map_set(obj_id, key, value_type, value_ptr, value_len) -> i32`
- `am_map_get(obj_id, key, value_out) -> i32`
- `am_map_delete(obj_id, key) -> i32`
- `am_map_keys(obj_id, keys_out) -> i32`
- `am_map_len(obj_id) -> u32`

**Go API**:
- `doc.Put(ctx, path, key, value) -> error`
- `doc.Get(ctx, path, key) -> (Value, error)`
- `doc.Delete(ctx, path, key) -> error`
- `doc.Keys(ctx, path) -> ([]string, error)`

**Tests needed**: `map_test.go`
- TestMap_PutGet
- TestMap_Delete
- TestMap_Keys
- TestMap_Nested

### 1.3 List Operations ❌ TODO
**Rust exports needed**:
- `am_list_push(obj_id, value_type, value_ptr, value_len) -> i32`
- `am_list_insert(obj_id, index, value_type, value_ptr, value_len) -> i32`
- `am_list_get(obj_id, index, value_out) -> i32`
- `am_list_delete(obj_id, index) -> i32`
- `am_list_len(obj_id) -> u32`

**Go API**:
- `doc.ListPush(ctx, path, value) -> error`
- `doc.ListInsert(ctx, path, index, value) -> error`
- `doc.ListGet(ctx, path, index) -> (Value, error)`
- `doc.ListDelete(ctx, path, index) -> error`

**Tests needed**: `list_test.go`
- TestList_PushGet
- TestList_Insert
- TestList_Delete
- TestList_Iteration

### 1.4 Counter Operations ❌ TODO
**Rust exports needed**:
- `am_counter_increment(obj_id, key, delta) -> i32`
- `am_counter_get(obj_id, key) -> i64`

**Go API**:
- `doc.Increment(ctx, path, key, delta) -> error`
- `doc.GetCounter(ctx, path, key) -> (int64, error)`

**Tests needed**: `counter_test.go`
- TestCounter_Increment
- TestCounter_Get
- TestCounter_Concurrent

---

## Phase 2: Document Lifecycle (CRITICAL)

### 2.1 Document Creation ✅ DONE
- [x] `am_init()` - Create new document
- [x] Tests: TestNew (passing)

### 2.2 Save/Load ✅ DONE
- [x] `am_save()` - Serialize to bytes
- [x] `am_load()` - Deserialize from bytes
- [x] Tests: TestDocument_SaveAndLoad (passing)

### 2.3 Merge ❌ BROKEN - ROOT CAUSE IDENTIFIED
- [x] `am_merge()` exists but **DOESN'T WORK**
- [ ] Fix merge implementation
- [ ] Test commutativity: merge(A,B) == merge(B,A)
- [ ] Test convergence: all replicas converge to same state
- [ ] Test no data loss

**ROOT CAUSE** (state.rs:15):
```rust
// THE PROBLEM: ONE global document per WASM instance
thread_local! {
    pub(crate) static DOC: RefCell<Option<AutoCommit>> = RefCell::new(None);
    //                     ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    //                     Shared global state!
}
```

**Why merge fails in Go tests**:
- When Go creates doc1 and doc2, both use the SAME WASM instance
- Second call to `am_init()` REPLACES the first document in global DOC
- Both doc1 and doc2 point to the SAME underlying document
- Merging document with itself = broken behavior

**Why integration test works**:
- `test_merge.sh` runs TWO separate server processes
- Each process has its OWN WASM instance with separate DOC
- Alice's doc and Bob's doc are truly independent
- Merge works correctly between processes

**Solutions**:
1. **Accept limitation**: Document that merge testing requires separate processes (current)
2. **Refactor**: Multi-document support via `HashMap<DocId, AutoCommit>` in state.rs (M3 milestone)
3. **Hybrid**: Keep single-doc for M0, plan multi-doc for M3

### 2.4 Changes/Patches ❌ TODO
**Rust exports needed**:
- `am_get_changes(from_heads, to_heads, changes_out) -> i32`
- `am_apply_changes(changes_ptr, changes_len) -> i32`
- `am_get_heads(heads_out) -> i32`
- `am_get_last_local_change(change_out) -> i32`

**Go API**:
- `doc.GetChanges(ctx, from, to) -> ([]Change, error)`
- `doc.ApplyChanges(ctx, changes) -> error`
- `doc.GetHeads(ctx) -> ([]ChangeHash, error)`

**Tests needed**: `changes_test.go`

---

## Phase 3: Query/Read Operations

### 3.1 Object Inspection ❌ TODO
**Rust exports needed**:
- `am_get_object_type(obj_id) -> i32` (returns: MAP=0, LIST=1, TEXT=2)
- `am_get_object_id(path_ptr, path_len, obj_id_out) -> i32`
- `am_get_parent(obj_id, parent_out) -> i32`

### 3.2 Iteration ❌ TODO
**Rust exports needed**:
- `am_map_iter_init(obj_id) -> i32` (returns iterator handle)
- `am_map_iter_next(iter_handle, key_out, value_out) -> i32`
- `am_list_iter_init(obj_id) -> i32`
- `am_list_iter_next(iter_handle, index_out, value_out) -> i32`

---

## Phase 4: Marks/Spans (Rich Text)

### 4.1 Mark Operations ❌ TODO
**Rust exports needed**:
- `am_mark(obj_id, start, end, name_ptr, name_len, value_ptr, value_len) -> i32`
- `am_unmark(obj_id, start, end, name_ptr, name_len) -> i32`
- `am_marks(obj_id, pos, marks_out) -> i32`
- `am_spans(obj_id, spans_out) -> i32`

**Go API**:
- `doc.Mark(ctx, path, start, end, markName, value) -> error`
- `doc.Unmark(ctx, path, start, end, markName) -> error`
- `doc.Marks(ctx, path, pos) -> ([]Mark, error)`
- `doc.Spans(ctx, path) -> ([]Span, error)`

**Tests needed**: `marks_test.go`

---

## Phase 5: History/Time Travel

### 5.1 Heads ❌ TODO
**Rust exports needed**:
- `am_get_heads(heads_out, heads_len_out) -> i32`
- `am_get_heads_count() -> u32`

### 5.2 At-Heads Queries ❌ TODO
**Rust exports needed**:
- `am_get_at(obj_id, key, heads_ptr, heads_len, value_out) -> i32`
- `am_keys_at(obj_id, heads_ptr, heads_len, keys_out) -> i32`

---

## Phase 6: Sync Protocol (M1 Milestone)

### 6.1 Sync State ❌ TODO
**Rust exports needed**:
- `am_sync_state_init() -> i32` (returns sync state handle)
- `am_sync_state_free(handle) -> i32`
- `am_generate_sync_message(sync_state_handle, msg_out) -> i32`
- `am_receive_sync_message(sync_state_handle, msg_ptr, msg_len) -> i32`

### 6.2 Sync Integration ❌ TODO
**Go API**:
- `doc.GenerateSyncMessage(ctx, syncState) -> ([]byte, error)`
- `doc.ReceiveSyncMessage(ctx, syncState, msg) -> error`

---

## Testing Strategy

### Test File Structure (One per module)
```
go/pkg/automerge/
├── document_test.go  ✅ (12 tests)
├── text_test.go      ✅ (15 tests)
├── map_test.go       ❌ TODO
├── list_test.go      ❌ TODO
├── counter_test.go   ❌ TODO
├── marks_test.go     ❌ TODO
├── changes_test.go   ❌ TODO
├── sync_test.go      ❌ TODO
└── history_test.go   ❌ TODO
```

### Rust Test Coverage
```
rust/automerge_wasi/src/
├── memory.rs         ✅ (3 tests)
├── document.rs       ✅ (2 tests)
├── text.rs           ✅ (3 tests)
├── map.rs            ❌ TODO (+ tests)
├── list.rs           ❌ TODO (+ tests)
├── counter.rs        ❌ TODO (+ tests)
├── marks.rs          ❌ TODO (+ tests)
├── changes.rs        ❌ TODO (+ tests)
└── sync.rs           ❌ TODO (+ tests)
```

---

## Priority Order

### NOW (Phase 1 & 2)
1. ✅ Fix CRDT merge bug (CRITICAL)
2. ✅ Implement Map operations
3. ✅ Implement List operations
4. ✅ Implement Counter operations
5. ✅ Create tests for each

### NEXT (Phase 3 & 4)
6. ✅ Implement object inspection
7. ✅ Implement marks/spans
8. ✅ Create comprehensive tests

### LATER (Phase 5 & 6)
9. ✅ Implement history/time travel
10. ✅ Implement sync protocol
11. ✅ End-to-end sync tests

---

## Success Criteria

**100% Feature Coverage**:
- [ ] All CRDT types working (text, map, list, counter)
- [ ] All document operations (save, load, merge, changes)
- [ ] All query operations (get, keys, iteration)
- [ ] Marks/spans for rich text
- [ ] History/time travel
- [ ] Sync protocol

**100% Test Coverage**:
- [ ] Every WASI export has Rust tests
- [ ] Every Go method has Go tests
- [ ] Table-driven tests for all scenarios
- [ ] CRDT properties verified (commutativity, convergence, no data loss)

**Current Status**: 27 tests passing, ~10% feature coverage
**Target**: 200+ tests passing, 100% feature coverage

---

## Next Actions

1. Start with Map operations (most common after text)
2. Create `rust/automerge_wasi/src/map.rs`
3. Add WASI exports for map operations
4. Create `go/pkg/automerge/map_test.go`
5. Verify all map operations work
6. Repeat for List, Counter, Marks, etc.

**No shortcuts. Every feature. Every test. 100%.**
