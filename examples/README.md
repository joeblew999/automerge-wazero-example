# Automerge WASI Examples

Practical code examples showing how to use Automerge CRDTs with the Go + WASM implementation.

## Prerequisites

```bash
# Build the Rust WASM module
make build-wasi

# Make sure automerge.wasm is available
ls rust/automerge_wasi/target/wasm32-wasip1/release/automerge.wasm
```

## Running Examples

```bash
# Run from the examples/ directory
cd examples

# Example 1: Text CRDT
WASM_PATH=../rust/automerge_wasi/target/wasm32-wasip1/release/automerge.wasm \
  go run 01_text_crdt.go

# Example 2: Map CRDT
WASM_PATH=../rust/automerge_wasi/target/wasm32-wasip1/release/automerge.wasm \
  go run 02_map_crdt.go

# Example 3: Sync Protocol
WASM_PATH=../rust/automerge_wasi/target/wasm32-wasip1/release/automerge.wasm \
  go run 03_sync_protocol.go
```

## Examples Overview

### 01_text_crdt.go ✅
**Basic Text Operations**

Demonstrates:
- Creating documents
- Inserting text with `TextSplice()`
- Reading text with `GetText()`
- Modifying and deleting text
- Saving and loading documents
- Unicode support

**Output**:
```
📝 Text CRDT Example
===================
✅ Inserted: "Hello, Automerge!"
📖 Current text: "Hello, Automerge!"
✏️  After edit: "Hello, CRDT Automerge!"
🗑️  After delete: "Hello, Automerge!"

💾 Persistence
==============
✅ Saved 234 bytes
📂 Loaded text: "Hello, Automerge!"

🌍 Unicode Support
==================
✅ Unicode text: "Hello 世界! 🌟"

🎉 Example complete!
```

### 02_map_crdt.go ✅
**Key-Value Storage**

Demonstrates:
- Storing different value types (string, int, bool, float)
- Getting values with `Get()`
- Listing keys with `Keys()`
- Deleting keys with `Delete()`
- Map length
- Persistence

**Output**:
```
🗺️  Map CRDT Example
===================
✅ Set name = Alice
✅ Set age = 30
✅ Set active = true
✅ Set balance = 123.45

📖 Reading values:
  name: Alice
  age: 30

🔑 All keys:
  name: Alice
  age: 30
  active: true
  balance: 123.45

📏 Map size: 4 entries

🗑️  Deleting 'balance':
✅ Remaining keys: [name age active]

✏️  Updating 'age':
✅ Updated age: 31

💾 Persistence:
✅ Saved 312 bytes
📂 Loaded keys: [name age active]

🎉 Example complete!
```

### 03_sync_protocol.go ✅
**Multi-Peer Synchronization**

Demonstrates:
- Two peers making independent changes
- Initializing sync state for each peer
- Generating sync messages with `GenerateSyncMessage()`
- Receiving sync messages with `ReceiveSyncMessage()`
- CRDT convergence (conflict-free merge)

**Output**:
```
🔄 Sync Protocol Example
========================
✅ Created two independent documents

📝 Alice's changes:
  Alice's document: "Hello from Alice!"

📝 Bob's changes:
  Bob's document: "Hello from Bob!"

⚠️  Documents are now diverged!

🔄 Starting sync:

→ Alice generating sync message...
✅ Alice's message: 156 bytes

← Bob receiving Alice's message...
✅ Bob processed message

→ Bob generating sync message...
✅ Bob's message: 142 bytes

← Alice receiving Bob's message...
✅ Alice processed message

🎯 After sync:
  Alice's document: "Hello from Alice!Hello from Bob!"
  Bob's document:   "Hello from Alice!Hello from Bob!"

✅ SUCCESS: Documents have converged!

📊 Explaining the merge:
  Both Alice and Bob's changes are preserved
  Automerge's CRDT algorithm ensures they see the same result
  This is CONFLICT-FREE merge!

🎉 Example complete!
```

## Additional Examples (Coming Soon)

- **04_list_crdt.go** - Array operations (push, insert, delete)
- **05_counter_crdt.go** - Distributed counters
- **06_rich_text.go** - Formatted text with marks
- **07_history.go** - Version control and time travel
- **08_merge_conflict.go** - How CRDTs resolve conflicts
- **09_http_client.go** - Using the HTTP API
- **10_persistence.go** - Advanced save/load patterns

## API Reference

For complete API documentation, see:
- [docs/reference/api-mapping.md](../docs/reference/api-mapping.md)
- [docs/explanation/architecture.md](../docs/explanation/architecture.md)

## Common Patterns

### Creating a Document
```go
runtime, _ := wazero.NewRuntime(ctx, "automerge.wasm")
defer runtime.Close()

doc := automerge.NewDocument(runtime)
doc.Init(ctx)
```

### Working with Paths
```go
path := automerge.NewPath()           // Root object
nested := path.Append("users")        // /users
deeper := nested.AppendIndex(0)       // /users/0
```

### Error Handling
```go
if err := doc.TextSplice(ctx, 0, 0, "text"); err != nil {
    log.Fatalf("Operation failed: %v", err)
}
```

### Cleanup
```go
defer runtime.Close()                 // Close wazero runtime
defer doc.FreeSyncState(ctx, state)   // Free sync state resources
```

## Troubleshooting

### "automerge.wasm not found"
```bash
# Build the WASM module
make build-wasi

# Or specify path explicitly
WASM_PATH=/path/to/automerge.wasm go run example.go
```

### "runtime creation failed"
Make sure the WASM file is the correct target:
```bash
# Check WASM target
file rust/automerge_wasi/target/wasm32-wasip1/release/automerge.wasm
# Should show: WebAssembly (wasm) binary module version 0x1 (mvp)
```

### "context deadline exceeded"
Increase timeout in runtime creation:
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
```

## Next Steps

- Try modifying the examples
- Combine multiple CRDTs in one document
- Implement real-time collaboration
- Add persistence to your application
- Deploy the HTTP server for multi-device sync

## See Also

- [Main README](../README.md)
- [Architecture Guide](../docs/explanation/architecture.md)
- [HTTP API Documentation](../docs/reference/http-api.md)
- [Project Status](../STATUS.md)
