// Package automerge provides a Go interface to the Automerge CRDT library.
//
// Automerge is a library for building collaborative applications. It provides
// data structures (documents) that can be modified concurrently by multiple
// users and automatically merged without conflicts.
//
// # Architecture
//
// This package wraps the Automerge Rust library compiled to WebAssembly (WASM).
// The architecture has four layers:
//
//  1. Automerge Rust Core - Full CRDT implementation (~60 methods)
//  2. WASI Wrapper - C-like ABI exports (11 functions in rust/automerge_wasi/src/lib.rs)
//  3. pkg/wazero - Low-level Go FFI to WASM (direct export wrappers)
//  4. pkg/automerge - High-level idiomatic Go API (this package)
//
// # Current Implementation Status
//
// ✅ Implemented Features:
//   - Document lifecycle (New, Load, Save, Close)
//   - Text operations (GetText, SpliceText, UpdateText)
//   - Merging (Merge)
//   - Persistence (Save, Load)
//
// ⚠️ Partially Implemented:
//   - UpdateText (deprecated - use SpliceText)
//   - Length (only for text)
//
// ❌ Not Yet Implemented (return NotImplementedError):
//   - Map operations (Get, Put, Delete, Keys) - M2 milestone
//   - List operations (Insert, Remove, Splice) - M2 milestone
//   - Counter operations (Increment) - M2 milestone
//   - Rich text (Mark, Unmark, GetMarks) - M4 milestone
//   - Sync protocol (GenerateSyncMessage, ReceiveSyncMessage) - M1 milestone
//   - History (GetHeads, GetChanges, Fork) - Future
//
// # Basic Usage
//
// Create a new document:
//
//	ctx := context.Background()
//	doc, err := automerge.New(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer doc.Close(ctx)
//
// Edit text:
//
//	path := automerge.Root().Get("content")
//	err = doc.SpliceText(ctx, path, 0, 0, "Hello, Automerge!")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Read text:
//
//	text, err := doc.GetText(ctx, path)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(text) // "Hello, Automerge!"
//
// Save and load:
//
//	// Save to bytes
//	data, err := doc.Save(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Load from bytes
//	doc2, err := automerge.Load(ctx, data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer doc2.Close(ctx)
//
// Merge two documents:
//
//	// Alice and Bob both edit offline
//	alice, _ := automerge.New(ctx)
//	bob, _ := automerge.New(ctx)
//
//	alice.SpliceText(ctx, path, 0, 0, "Alice was here")
//	bob.SpliceText(ctx, path, 0, 0, "Bob was here")
//
//	// Merge Bob's changes into Alice's document
//	alice.Merge(ctx, bob)
//
//	// Both edits are preserved!
//	text, _ := alice.GetText(ctx, path)
//	fmt.Println(text) // Contains both edits
//
// # Error Handling
//
// This package defines several error types:
//
//   - NotImplementedError: Feature not yet implemented (includes milestone info)
//   - DeprecatedError: Method is deprecated (includes recommended alternative)
//   - WASMError: Error from WASM layer (includes operation name and error code)
//
// You can use errors.As to check for specific error types:
//
//	var notImpl *automerge.NotImplementedError
//	if errors.As(err, &notImpl) {
//	    fmt.Printf("Feature %s planned for %s\n", notImpl.Feature, notImpl.Milestone)
//	}
//
// # Milestones
//
// The roadmap for this package follows the milestones defined in CLAUDE.md:
//
//   - M1: Sync Protocol - delta-based synchronization
//   - M2: Multi-Document - support for maps, lists, and multiple text objects
//   - M3: NATS Transport - integration with NATS messaging
//   - M4: Rich Text - marks, blocks, formatting
//
// See API_MAPPING.md for detailed API coverage and implementation status.
package automerge
