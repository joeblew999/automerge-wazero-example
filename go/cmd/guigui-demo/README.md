# Guigui Demo - Native Desktop UI for Automerge

> **Status**: 🚧 PLANNED (See [STATUS.md](../../../STATUS.md) - M3.5 Milestone)

A pure Go desktop application demonstrating Automerge CRDT operations using the [guigui](https://github.com/guigui-gui/guigui) immediate-mode GUI framework.

## Overview

This demo showcases how to use the Automerge Go API (Layer 4) **directly** without HTTP overhead, providing a native desktop experience for local-first collaborative applications.

## Architecture

```
Guigui App (Pure Go)
    ↓ Direct Go calls
pkg/automerge (Layer 4 - High-level API)
    ↓ WASM calls
pkg/wazero (Layer 3 - FFI)
    ↓ C-ABI
rust/automerge_wasi (Layer 2)
    ↓
Automerge Core (Layer 1)
```

**Key Difference from Web UI**:
| Aspect | Web UI (Current) | Guigui Demo |
|--------|------------------|-------------|
| **Layer** | Layer 6 (HTTP API) | Layer 4 (Go API) |
| **Protocol** | HTTP + SSE | Direct function calls |
| **Deployment** | Browser required | Native executable |
| **Latency** | HTTP overhead | Zero overhead |
| **Use Case** | Multi-user web apps | Local desktop apps |

## Features (Planned)

### Phase 1: Basic CRDT Operations
- Text editor widget using `doc.TextSplice()`
- Real-time character count display
- Save/Load buttons (`doc.Save()`, `doc.Load()`)
- Visual feedback for CRDT operations

### Phase 2: Multi-Object Support
- Map editor (key-value pairs with `doc.Put()`, `doc.Get()`)
- List viewer (array operations with `doc.ListPush()`, `doc.ListInsert()`)
- Counter demo (increment/decrement with `doc.Increment()`)
- Nested object visualization

### Phase 3: History & Time Travel
- Timeline slider showing document versions
- "Undo" button using `doc.GetChanges()` and `doc.LoadIncremental()`
- Visual diff between versions
- Heads display (`doc.GetHeads()`)

### Phase 4: Rich Text
- Formatted text editor with marks
- Bold/Italic/Underline buttons using `doc.Mark()`, `doc.Unmark()`
- Color highlighting
- Spans visualization

## Installation

### Why Separate Workspace?

This demo uses a **separate Go workspace** (`go.work`) to:
- ✅ Isolate guigui dependencies from main codebase
- ✅ Prevent dependency pollution in parent module
- ✅ Allow independent versioning
- ✅ Keep main `go.mod` clean and minimal

The workspace references the parent module via `replace` directive, so it can still use `pkg/automerge` and `pkg/wazero`.

### Setup

```bash
# Clone the repository (if not already done)
git clone https://github.com/joeblew999/automerge-wazero-example
cd automerge-wazero-example

# Build the Rust WASI module (required for automerge.wasm)
make build-wasi

# Navigate to demo directory
cd go/cmd/guigui-demo

# Install dependencies (when implemented)
# go get github.com/guigui-gui/guigui

# Run the demo (uses go.work workspace)
go run .
```

**Note**: The guigui demo **does NOT require the HTTP server** to be running. It uses the Go API directly (Layer 4), not HTTP. This means:
- ✅ You can run the guigui demo while the HTTP server is also running
- ✅ No port conflicts - guigui is a native desktop app, not a web server
- ✅ The HTTP server (`make dev`) is for the web UI only

### Workspace Structure

```
automerge-wazero-example/
├── go/
│   ├── go.mod                    # Main module (wazero only)
│   ├── pkg/automerge/            # Layer 4 API
│   ├── pkg/wazero/               # Layer 3 FFI
│   └── cmd/
│       ├── server/               # HTTP server (uses main go.mod)
│       └── guigui-demo/          # Guigui demo (separate workspace)
│           ├── go.mod            # Demo-specific module
│           ├── go.work           # Workspace config
│           ├── main.go           # Entry point
│           └── README.md         # This file
└── rust/automerge_wasi/          # Rust WASI module
```

The `go.work` file tells Go to use both:
1. Current directory (guigui-demo module with guigui dependency)
2. Parent directory (main module with pkg/automerge, pkg/wazero)

## File Structure

```
go/cmd/guigui-demo/
├── main.go              # App entry point + wazero runtime setup
├── text_widget.go       # Text CRDT widget
├── map_widget.go        # Map CRDT widget
├── list_widget.go       # List CRDT widget
├── counter_widget.go    # Counter widget
├── history_widget.go    # History timeline widget
└── README.md            # This file
```

## Example Code

```go
package main

import (
    "context"
    "github.com/guigui-gui/guigui/app"
    "github.com/joeblew999/automerge-wazero-example/go/pkg/automerge"
    "github.com/joeblew999/automerge-wazero-example/go/pkg/wazero"
)

func main() {
    // Initialize wazero runtime
    runtime, err := wazero.NewRuntime(context.Background(), "automerge.wasm")
    if err != nil {
        panic(err)
    }
    defer runtime.Close()

    // Create Automerge document
    doc := automerge.NewDocument(runtime)
    if err := doc.Init(context.Background()); err != nil {
        panic(err)
    }

    // Run Guigui app
    app.Run(func(ctx *app.Context) {
        // UI rendering happens here each frame
        renderTextEditor(ctx, doc)
    })
}
```

## Benefits

### For Users
- ✅ No web browser required
- ✅ Native desktop performance
- ✅ Offline-first by design
- ✅ Cross-platform (Windows, macOS, Linux)

### For Developers
- ✅ Pure Go codebase (no context switching)
- ✅ Direct API access (Layer 4, no HTTP)
- ✅ Easier debugging (single-process)
- ✅ Type-safe API calls

### For the Project
- ✅ Demonstrates Go API capabilities
- ✅ Validates Layer 4 design
- ✅ Provides alternative to web UI
- ✅ Shows local-first architecture

## Dependencies

**Required**:
- Go 1.21+
- [Guigui](https://github.com/guigui-gui/guigui) - Immediate-mode GUI framework for Go
- Existing `pkg/automerge` API (Layer 4)
- Existing `pkg/wazero` runtime (Layer 3)

**Optional**:
- File dialog library (for Save/Load dialogs)
- System tray integration
- Hot reload for development

## Development Timeline

**Phase 1** (1-2 days): Basic text editor
**Phase 2** (2-3 days): Multi-object widgets
**Phase 3** (1-2 days): History/time travel
**Phase 4** (2-3 days): Rich text formatting

**Total Estimate**: 6-10 days

## Testing

```bash
# Run unit tests
go test ./go/cmd/guigui-demo/...

# Run the demo
go run ./go/cmd/guigui-demo
```

### Acceptance Criteria
- [ ] All M0 features work (Text, Map, List, Counter)
- [ ] Save/Load persists data correctly
- [ ] UI is responsive (no blocking on CRDT ops)
- [ ] Works on macOS, Linux, Windows

## Future Enhancements

- **Mobile UI**: Use `gomobile` to compile Guigui app for iOS/Android
- **NATS Transport**: Add real-time sync between Guigui instances (M3)
- **Plugin System**: Allow users to create custom widgets
- **Datastar Comparison**: Compare Guigui vs Datastar for Go-based UIs (M4)

## Links

- **Guigui GitHub**: https://github.com/guigui-gui/guigui
- **STATUS**: [STATUS.md](../../../STATUS.md) - See M3.5 Guigui Demo milestone
- **Architecture**: [docs/explanation/architecture.md](../../../docs/explanation/architecture.md)
- **API Reference**: [docs/reference/api-mapping.md](../../../docs/reference/api-mapping.md)

## License

MIT (same as parent project)
