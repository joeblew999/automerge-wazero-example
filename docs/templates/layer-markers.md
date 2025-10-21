# Layer Marker Templates

Templates for adding layer markers to files across the 7-layer architecture.

## Layer 2: Rust WASI Exports

```rust
// ═══════════════════════════════════════════════════════════════
// LAYER 2: Rust WASI Exports (C-ABI for FFI)
//
// Responsibilities:
// - Export C-ABI functions callable from Go via wazero
// - Validate UTF-8 input from Go side
// - Call Automerge Rust API for CRDT operations
// - Return error codes as i32 (0 = success, <0 = error)
//
// Dependencies:
// ⬇️  Calls: automerge crate (Layer 1 - CRDT core)
// ⬆️  Called by: go/pkg/wazero/<module>.go (Layer 3 - Go FFI wrappers)
//
// Related Files:
// 🔁 Siblings: [list other .rs files]
// 📝 Tests: cargo test (Rust unit tests)
// 🔗 Docs: docs/explanation/architecture.md#layer-2-rust-wasi
// ═══════════════════════════════════════════════════════════════
```

## Layer 3: Go FFI Wrappers

```go
// ═══════════════════════════════════════════════════════════════
// LAYER 3: Go FFI Wrappers (wazero → WASM)
//
// Responsibilities:
// - Call WASM functions via wazero runtime
// - Marshal Go strings/data to WASM linear memory
// - Translate WASM error codes to Go errors
// - Manage WASM memory allocation/deallocation
//
// Dependencies:
// ⬇️  Calls: WASM functions (am_<module>_*)
//           Implemented in: rust/automerge_wasi/src/<module>.rs (Layer 2)
// ⬆️  Called by: go/pkg/automerge/<module>.go (Layer 4 - high-level API)
//
// Related Files:
// 🔁 Siblings: [list other wazero/*.go files]
// 📝 Tests: <module>_test.go (FFI boundary tests)
// 🔗 Docs: docs/explanation/architecture.md#layer-3-go-ffi
// ═══════════════════════════════════════════════════════════════
```

## Layer 4: Go High-Level API

```go
// ═══════════════════════════════════════════════════════════════
// LAYER 4: Go High-Level CRDT API (Pure Functions)
//
// Responsibilities:
// - Provide idiomatic Go API for Automerge <Module> CRDT
// - Pure CRDT operations (NO state, NO mutex, NO persistence)
// - Path-based API for document navigation
// - Error handling and validation
//
// Dependencies:
// ⬇️  Calls: go/pkg/wazero/<module>.go (Layer 3 - FFI wrappers)
// ⬆️  Called by: go/pkg/server/<module>.go (Layer 5 - adds state + thread safety)
//
// Related Files:
// 🔁 Siblings: [list other automerge/*.go files]
// 📝 Tests: <module>_test.go (CRDT property tests)
// 🔗 Docs: docs/explanation/architecture.md#layer-4-go-api
//
// Design Note:
// This layer is STATELESS - it doesn't own the Document or manage
// concurrency. Layer 5 (server) adds mutex protection and persistence.
// ═══════════════════════════════════════════════════════════════
```

## Layer 5: Go Server Layer

```go
// ═══════════════════════════════════════════════════════════════
// LAYER 5: Go Server Layer (Stateful + Thread-Safe)
//
// Responsibilities:
// - Own the Document instance and manage its lifecycle
// - Add thread safety with mutex protection (s.mu.Lock/RLock)
// - Add persistence (call saveDocument after mutations)
// - Manage SSE broadcast to connected clients
//
// Dependencies:
// ⬇️  Calls: go/pkg/automerge/<module>.go (Layer 4 - stateless CRDT API)
// ⬆️  Called by: go/pkg/api/<module>.go (Layer 6 - HTTP handlers)
//
// Related Files:
// 🔁 Siblings: [list other server/*.go files]
// 📝 Tests: <module>_test.go (concurrency + persistence tests)
// 🔗 Docs: docs/explanation/architecture.md#layer-5-server
//
// Design Note:
// This layer adds MUTATION side effects that Layer 4 doesn't have:
// - Mutex locking (thread safety for concurrent HTTP requests)
// - Disk writes (saveDocument after each mutation)
// - SSE broadcasts (notify all connected clients)
// ═══════════════════════════════════════════════════════════════
```

## Layer 6: HTTP API Handlers

```go
// ═══════════════════════════════════════════════════════════════
// LAYER 6: HTTP API Handlers (Protocol Layer)
//
// Responsibilities:
// - Parse HTTP requests (JSON body, query params, headers)
// - Call server layer methods (Layer 5)
// - Format HTTP responses (JSON, status codes, headers)
// - Handle HTTP-specific concerns (CORS, content-type, etc)
//
// Dependencies:
// ⬇️  Calls: go/pkg/server/*.go (Layer 5 - stateful operations)
// ⬆️  Called by: HTTP router in cmd/server/main.go
//
// Related Files:
// 🔁 Siblings: [list other api/*.go files]
// 📝 Tests: <module>_test.go (HTTP integration tests)
// 🔗 Docs: docs/explanation/architecture.md#layer-6-http-api
//           docs/reference/http-api-complete.md
//
// Design Note:
// This layer is STATELESS - it doesn't own any data.
// All state management delegated to Layer 5 (server).
// Multiple HTTP handlers can share the same server instance.
// ═══════════════════════════════════════════════════════════════
```

## Layer 7: Web Frontend (JavaScript)

```javascript
// ═══════════════════════════════════════════════════════════════
// LAYER 7: Web Frontend (JavaScript Client)
//
// Responsibilities:
// - Handle user interactions for <Module> CRDT
// - Call HTTP API endpoints (Layer 6)
// - Update DOM based on SSE events
// - Manage client-side state (peer ID, connection status, etc)
//
// Dependencies:
// ⬇️  Calls: /api/<module> (Layer 6 - HTTP API)
//           SSE: /api/stream (server events)
// ⬆️  Called by: web/js/app.js (orchestrator)
//
// Related Files:
// 🔁 Component: web/components/<module>.html (UI template)
// 🔁 Backend: go/pkg/api/<module>.go (Layer 6)
// 📝 Tests: tests/playwright/<module>_test_plan.md
// 🔗 Docs: docs/explanation/architecture.md#layer-7-web
//
// Design Note:
// This layer provides CRDT-specific UI logic. Infrastructure
// concerns (tab switching, SSE setup, routing) live in app.js.
// ═══════════════════════════════════════════════════════════════

export class <Module>Component {
    constructor() {
        // Component state
    }

    init() {
        // Setup event listeners
    }

    destroy() {
        // Cleanup when switching tabs
    }
}
```

## Layer 7: Web Frontend (HTML)

```html
<!--
═══════════════════════════════════════════════════════════════
LAYER 7: Web Frontend (HTML Template)

Responsibilities:
- UI template for <Module> CRDT operations
- Form inputs, buttons, display areas
- Loaded by web/js/app.js into main DOM

Related Files:
🔁 Logic: web/js/<module>.js (component class)
🔁 Backend: go/pkg/api/<module>.go (Layer 6)
🔗 Docs: docs/explanation/architecture.md#layer-7-web

Design Note:
This is a pure HTML fragment loaded via fetch().
No inline JavaScript - all logic in web/js/<module>.js
═══════════════════════════════════════════════════════════════
-->

<div class="<module>-container">
    <!-- UI elements here -->
</div>
```

## Usage

1. Copy the appropriate template for your layer
2. Replace `<module>` with the actual module name (text, map, list, etc)
3. Update the sibling list with actual related files
4. Add at the **top** of the file, before `package` declaration
5. Keep existing module-level comments below the layer marker

## Example

```go
// ═══════════════════════════════════════════════════════════════
// LAYER 4: Go High-Level CRDT API (Pure Functions)
// ... (template content)
// ═══════════════════════════════════════════════════════════════

package automerge

import "context"

// GetText retrieves the text content at the given path.
// ... (existing function docs)
```

## See Also

- [Architecture Guide](../explanation/architecture.md) - Understanding the 7-layer design
- [AI Readability Guide](../explanation/ai-readability-improvements.md) - Why layer markers matter
- [CLAUDE.md Section 0.3.1](../../CLAUDE.md#031-ai-code-connection-strategy) - AI navigation strategy
