.PHONY: help build-wasi build-wasi-debug run dev watch test tidy clean clean-snapshots clean-all check-deps install-deps

# Configuration
WASI_TARGET = wasm32-wasip1
WASM_DIR = rust/automerge_wasi
WASM_RELEASE = $(WASM_DIR)/target/$(WASI_TARGET)/release/automerge_wasi.wasm
WASM_DEBUG = $(WASM_DIR)/target/$(WASI_TARGET)/debug/automerge_wasi.wasm
GO_DIR = go/cmd/server
PORT ?= 8080

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## check-deps: Check if required dependencies are installed
check-deps:
	@echo "Checking dependencies..."
	@command -v cargo >/dev/null 2>&1 || { echo "❌ cargo not found. Install Rust from https://rustup.rs"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "❌ go not found. Install Go from https://go.dev"; exit 1; }
	@rustup target list | grep -q "$(WASI_TARGET) (installed)" || { echo "⚠️  WASI target not installed. Run: make install-deps"; exit 1; }
	@echo "✅ All dependencies installed"

## install-deps: Install required Rust WASI target
install-deps:
	@echo "Installing WASI target..."
	rustup target add $(WASI_TARGET)
	@echo "✅ WASI target installed"

## build-wasi: Build Rust WASI module (release mode)
build-wasi: check-deps
	@echo "🔨 Building Rust WASI module (release)..."
	cd $(WASM_DIR) && cargo build --target $(WASI_TARGET) --release
	@echo "✅ Built: $(WASM_RELEASE)"
	@ls -lh $(WASM_RELEASE)

## build-wasi-debug: Build Rust WASI module (debug mode, faster compile)
build-wasi-debug: check-deps
	@echo "🔨 Building Rust WASI module (debug)..."
	cd $(WASM_DIR) && cargo build --target $(WASI_TARGET)
	@echo "✅ Built: $(WASM_DEBUG)"
	@ls -lh $(WASM_DEBUG)

## run: Build and run the Go server (release build)
run: build-wasi
	@echo "🚀 Starting Go server on port $(PORT)..."
	cd $(GO_DIR) && PORT=$(PORT) go run main.go

## dev: Build (debug) and run the Go server (faster iteration)
dev: build-wasi-debug
	@echo "🚀 Starting Go server in dev mode on port $(PORT)..."
	cd $(GO_DIR) && PORT=$(PORT) WASM_FILE=../../../$(WASM_DEBUG) go run main.go

## watch: Watch for changes and auto-rebuild (requires air)
watch:
	@command -v air >/dev/null 2>&1 || { echo "⚠️  'air' not found. Install: go install github.com/air-verse/air@latest"; exit 1; }
	@echo "👀 Watching for changes..."
	cd $(GO_DIR) && air

## test: Run all tests (Rust + Go)
test:
	@echo "🧪 Running Rust tests..."
	cd $(WASM_DIR) && cargo test
	@echo "🧪 Running Go tests..."
	cd $(GO_DIR) && go test ./...

## tidy: Run go mod tidy
tidy:
	@echo "🧹 Running go mod tidy..."
	cd $(GO_DIR) && go mod tidy

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	cd $(WASM_DIR) && cargo clean
	cd $(GO_DIR) && go clean

## clean-snapshots: Remove all doc.am snapshot files
clean-snapshots:
	@echo "🧹 Cleaning snapshot files..."
	rm -f data/*.am
	rm -f doc.am
	@echo "✅ Snapshots cleaned"

## clean-all: Clean everything (artifacts + snapshots)
clean-all: clean clean-snapshots
	@echo "✅ All clean!"

## verify: Verify the WASM module is valid
verify: build-wasi
	@echo "🔍 Verifying WASM module..."
	@file $(WASM_RELEASE)
	@echo "✅ WASM module is valid"

## size: Show size of built artifacts
size:
	@echo "📊 Build artifacts sizes:"
	@[ -f $(WASM_RELEASE) ] && ls -lh $(WASM_RELEASE) || echo "  Release WASM: not built"
	@[ -f $(WASM_DEBUG) ] && ls -lh $(WASM_DEBUG) || echo "  Debug WASM: not built"
