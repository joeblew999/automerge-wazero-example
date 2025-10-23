# How to Embed Automerge WASI in Your Go Application

This guide shows how to use the Automerge WASI server as a library in your own Go applications.

## Quick Start

### Basic Example (10 lines)

```go
package main

import (
    "log"
    "github.com/joeblew999/automerge-wazero-example/pkg/config"
    "github.com/joeblew999/automerge-wazero-example/pkg/httpserver"
)

func main() {
    cfg := config.NewFromEnv()
    srv, err := httpserver.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    log.Fatal(srv.ListenAndServe())
}
```

Run with environment variables:
```bash
PORT=3000 STORAGE_DIR=/data go run main.go
```

## Configuration

### Environment Variables

All configuration is done via environment variables (with sensible defaults):

| Variable | Default | Description |
|----------|---------|-------------|
| **PORT** | `8080` | HTTP port to listen on |
| **STORAGE_DIR** | `.` | Directory for `.am` snapshot files |
| **USER_ID** | `default` | Server instance identifier (for logging) |
| **WASM_PATH** | `../rust/.../automerge_wasi.wasm` | Path to WASM file |
| **WEB_PATH** | `../web` | Path to web UI folder |
| **ENABLE_UI** | `true` | Enable web UI routes |

### Programmatic Configuration

```go
cfg := config.Config{
    Port:       "3000",
    StorageDir: "/data",
    UserID:     "production-1",
    WASMPath:   "/app/automerge.wasm",
    WebPath:    "/app/ui",
    EnableUI:   true,
}

srv, err := httpserver.New(cfg)
if err != nil {
    log.Fatal(err)
}
log.Fatal(srv.ListenAndServe())
```

## Advanced Usage

### Adding Custom Routes

```go
package main

import (
    "log"
    "net/http"
    "github.com/joeblew999/automerge-wazero-example/pkg/config"
    "github.com/joeblew999/automerge-wazero-example/pkg/httpserver"
)

func main() {
    cfg := config.NewFromEnv()
    srv, err := httpserver.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Add your custom routes
    srv.Mux().HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    srv.Mux().HandleFunc("/custom", customHandler)

    log.Fatal(srv.ListenAndServe())
}

func customHandler(w http.ResponseWriter, r *http.Request) {
    // Your custom logic here
}
```

### Direct Server Access

If you need to call Automerge methods directly (not via HTTP):

```go
package main

import (
    "context"
    "log"
    "github.com/joeblew999/automerge-wazero-example/pkg/config"
    "github.com/joeblew999/automerge-wazero-example/pkg/httpserver"
)

func main() {
    cfg := config.NewFromEnv()
    httpSrv, err := httpserver.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Get direct access to Automerge server
    srv := httpSrv.Server()

    // Call server methods directly
    ctx := context.Background()
    text, err := srv.GetText(ctx)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Current text: %s", text)

    // Start HTTP server
    log.Fatal(httpSrv.ListenAndServe())
}
```

### API-Only Mode (No Web UI)

```bash
ENABLE_UI=false go run main.go
```

Or programmatically:

```go
cfg := config.NewFromEnv()
cfg.EnableUI = false  // No UI routes (/, /web/*, /vendor/*)

srv, err := httpserver.New(cfg)
```

## Embedding WASM Binary (Future)

For mobile/desktop apps, you can embed the WASM file directly:

```go
package main

import (
    _ "embed"
    "log"
    "github.com/joeblew999/automerge-wazero-example/pkg/config"
    "github.com/joeblew999/automerge-wazero-example/pkg/httpserver"
)

//go:embed automerge_wasi.wasm
var wasmBytes []byte

func main() {
    cfg := config.NewFromEnv()
    cfg.WASMBytes = wasmBytes  // Use embedded bytes instead of file path
    cfg.WASMPath = ""           // Clear path

    srv, err := httpserver.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    log.Fatal(srv.ListenAndServe())
}
```

**Note**: Embedding is not yet implemented - WASMBytes field is reserved for future use.

## API Endpoints

The server exposes these HTTP endpoints:

### M0 - Core CRDT Operations

- `POST /api/text` - Set text content
- `GET /api/text` - Get text content
- `GET /api/stream` - SSE stream for real-time updates
- `POST /api/merge` - Merge another document
- `GET /api/doc` - Download `.am` snapshot

### M0 - Data Structures

- `POST /api/map` - Put/delete map values
- `GET /api/map/keys` - Get map keys
- `POST /api/list/push` - Push to list
- `POST /api/list/insert` - Insert into list
- `GET /api/list` - Get list values
- `POST /api/counter` - Update counter
- `GET /api/counter/get` - Get counter value

### M1 - Sync Protocol

- `POST /api/sync` - Generate/receive sync messages

### M2 - Rich Text

- `POST /api/richtext/mark` - Apply formatting marks
- `POST /api/richtext/unmark` - Remove formatting marks
- `GET /api/richtext/marks` - Get marks at position

See [HTTP API Complete](../reference/http-api-complete.md) for full API docs.

## Examples

### Mobile App (gomobile)

```go
package mobileapp

import (
    "github.com/joeblew999/automerge-wazero-example/pkg/config"
    "github.com/joeblew999/automerge-wazero-example/pkg/httpserver"
)

// Start starts the Automerge HTTP server for mobile app
// This will be exposed to mobile via gomobile bind
func Start(port string, dataDir string) error {
    cfg := config.Config{
        Port:       port,
        StorageDir: dataDir,
        UserID:     "mobile-app",
        EnableUI:   false, // No UI for mobile
    }

    srv, err := httpserver.New(cfg)
    if err != nil {
        return err
    }

    go func() {
        _ = srv.ListenAndServe()
    }()

    return nil
}
```

### Desktop App (Wails/Fyne)

```go
package main

import (
    "context"
    "github.com/joeblew999/automerge-wazero-example/pkg/config"
    "github.com/joeblew999/automerge-wazero-example/pkg/httpserver"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
    ctx    context.Context
    server *httpserver.HTTPServer
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx

    cfg := config.Config{
        Port:       "8080",
        StorageDir: "/Users/me/.myapp/data",
        UserID:     "desktop-app",
        EnableUI:   true, // Embed web UI
    }

    srv, err := httpserver.New(cfg)
    if err != nil {
        runtime.LogFatal(ctx, err.Error())
    }

    a.server = srv

    // Start server in background
    go func() {
        _ = srv.ListenAndServe()
    }()
}
```

## Best Practices

1. **Always set STORAGE_DIR** in production to a persistent location
2. **Use unique USER_ID** for each server instance in clustered setups
3. **Disable UI (ENABLE_UI=false)** if only using API endpoints
4. **Add /health endpoint** for monitoring/load balancing
5. **Use embed.FS** for WASM binary in production deployments

## Troubleshooting

### WASM file not found

```
Failed to create server: failed to load WASM: open ../rust/.../automerge_wasi.wasm: no such file or directory
```

**Solution**: Set WASM_PATH environment variable to correct path:

```bash
WASM_PATH=/path/to/automerge_wasi.wasm go run main.go
```

### Port already in use

```
listen tcp :8080: bind: address already in use
```

**Solution**: Change port:

```bash
PORT=3000 go run main.go
```

### Permission denied writing snapshots

```
Failed to save document: open doc.am: permission denied
```

**Solution**: Set STORAGE_DIR to writable directory:

```bash
STORAGE_DIR=/tmp go run main.go
```

## See Also

- [Architecture Guide](../explanation/architecture.md) - Understanding the layers
- [HTTP API Reference](../reference/http-api-complete.md) - Complete API docs
- [Project Status](../../STATUS.md) - Current status and future plans
