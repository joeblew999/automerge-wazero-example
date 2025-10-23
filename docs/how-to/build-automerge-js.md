# How To: Build Automerge.js from Source

**CRITICAL**: We build our own Automerge.js from the same source as our Rust WASI!

## Why Build from Source?

- ✅ **Version alignment**: Rust backend and JS frontend use identical Automerge version
- ✅ **Single source of truth**: `.src/automerge/` contains both Rust and JS
- ✅ **Custom builds**: Can create slim/fat builds, IIFE/ESM formats
- ✅ **Debugging**: Full source maps, ability to patch if needed

## Build Process

```bash
# 1. Setup source (first time only)
make setup-src              # Clones .src/automerge/ (rust/automerge@0.7.0)

# 2. Install Rust WASM toolchain
make setup-rust-wasm        # Installs wasm32-unknown-unknown + wasm-bindgen

# 3. Build Automerge.js
make build-js               # Builds .src/automerge/javascript/ → ui/vendor/automerge.js
```

## Build Output

```
.src/automerge/javascript/dist/cjs/iife.cjs  # Built IIFE bundle
         ↓ (copied by make build-js)
ui/vendor/automerge.js                       # 3.4M IIFE format
```

## Usage in Web

**Old UI** (`ui/ui.html`):
```html
<script src="/vendor/automerge.js"></script>
<script>
  console.log('Automerge loaded:', typeof window.Automerge);
</script>
```

**New Web Folder** (`web/index.html`):
```html
<script src="/vendor/automerge.js"></script>
<script type="module" src="/web/js/app.js"></script>
```

**Served by Go**:
```go
// go/cmd/server/main.go
http.Handle("/vendor/", api.VendorHandler(staticCfg))  // Serves ui/vendor/
```

## Version Tracking

```bash
make sync-versions   # Verify all components use same .src/ version
```

**Output**:
```
📌 .src/automerge git version: rust/automerge@0.7.0
🦀 Cargo.toml dependency: automerge = { path = "../../.src/automerge/rust/automerge" }
📦 JavaScript package.json: "version": "3.2.0-alpha.0"
✅ Built Automerge.js: 3.4M
```

## Verification

```bash
make verify-web  # Checks that web/index.html references /vendor/automerge.js
```

## See Also

- [Architecture Guide](../explanation/architecture.md) - Understanding the layers
- [Web Architecture](../explanation/web-architecture.md) - Web folder structure
- [CLAUDE.md](../../CLAUDE.md) - AI agent instructions
