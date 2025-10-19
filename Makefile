.PHONY: help build-wasi build-wasi-debug run dev watch test tidy clean clean-snapshots clean-all check-deps install-deps run-alice run-bob run-server test-two-laptops clean-test-data

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

## run-alice: Run server as Laptop A (port 8080, storage: ./data/alice/)
run-alice: build-wasi
	@echo "🚀 Starting Laptop A (Alice) on port 8080..."
	@mkdir -p data/alice
	PORT=8080 STORAGE_DIR=./data/alice USER_ID=alice $(MAKE) -s run-server

## run-bob: Run server as Laptop B (port 8081, storage: ./data/bob/)
run-bob: build-wasi
	@echo "🚀 Starting Laptop B (Bob) on port 8081..."
	@mkdir -p data/bob
	PORT=8081 STORAGE_DIR=./data/bob USER_ID=bob $(MAKE) -s run-server

## run-server: Internal target to run Go server (uses env vars)
run-server:
	cd $(GO_DIR) && go run main.go

## test-two-laptops: Start both Alice and Bob servers for testing
test-two-laptops: build-wasi
	@echo "🧪 Starting 2-laptop test environment..."
	@mkdir -p data/alice data/bob
	@echo "  Alice: http://localhost:8080 (storage: ./data/alice/)"
	@echo "  Bob:   http://localhost:8081 (storage: ./data/bob/)"
	@echo ""
	@echo "Starting Alice..."
	@PORT=8080 STORAGE_DIR=./data/alice USER_ID=alice $(MAKE) -s run-server &
	@sleep 2
	@echo "Starting Bob..."
	@PORT=8081 STORAGE_DIR=./data/bob USER_ID=bob $(MAKE) -s run-server &
	@sleep 2
	@echo ""
	@echo "✅ Both servers running!"
	@echo "   Open http://localhost:8080 (Alice's laptop)"
	@echo "   Open http://localhost:8081 (Bob's laptop)"
	@echo ""
	@echo "Press Ctrl+C to stop both servers"
	@wait

## clean-test-data: Clean test laptop data directories
clean-test-data:
	@echo "🧹 Cleaning test data..."
	rm -rf data/alice data/bob
	@echo "✅ Test data cleaned"
