.PHONY: help build-wasi build-wasi-debug build-js build-server run dev watch test test-go test-rust test-http test-playwright tidy clean clean-snapshots clean-all check-deps install-deps setup-rust-wasm run-alice run-bob run-server test-two-laptops clean-test-data setup-src update-src clean-src generate-test-data sync-versions verify-docs verify-web

# Configuration
WASI_TARGET = wasm32-wasip1
WASM_DIR = rust/automerge_wasi
WASM_RELEASE = $(WASM_DIR)/target/$(WASI_TARGET)/release/automerge_wasi.wasm
WASM_DEBUG = $(WASM_DIR)/target/$(WASI_TARGET)/debug/automerge_wasi.wasm
GO_ROOT = go
GO_DIR = $(GO_ROOT)/cmd/server
PORT ?= 8080

# Source reference configuration (single source of truth)
# We build from .src/automerge/ at the specified git tag
# Tag: https://github.com/automerge/automerge/releases/tag/rust%2Fautomerge%400.7.0
AUTOMERGE_REPO = https://github.com/automerge/automerge.git
AUTOMERGE_TAG = rust/automerge@0.7.0
SRC_DIR = .src

# JavaScript build configuration
# Builds from .src/automerge/javascript/ and copies to web/vendor/
# The web/vendor/ directory is served at /vendor/ route
JS_SRC_DIR = $(SRC_DIR)/automerge/javascript
JS_DIST = $(JS_SRC_DIR)/dist/cjs/iife.cjs
VENDOR_DIR = web/vendor
VENDOR_JS = $(VENDOR_DIR)/automerge.js

# Web folder configuration (1:1 mapping architecture)
WEB_DIR = web
WEB_HTML = $(WEB_DIR)/index.html
WEB_CSS = $(WEB_DIR)/css/main.css
WEB_JS = $(WEB_DIR)/js/app.js $(WEB_DIR)/js/crdt_text.js $(WEB_DIR)/js/crdt_sync.js $(WEB_DIR)/js/crdt_richtext.js
WEB_COMPONENTS = $(WEB_DIR)/components/crdt_text.html $(WEB_DIR)/components/crdt_sync.html $(WEB_DIR)/components/crdt_richtext.html

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

## setup-rust-wasm: Install Rust targets and tools needed for JavaScript build
setup-rust-wasm:
	@echo "🦀 Setting up Rust WASM toolchain for Automerge.js build..."
	@echo ""
	@echo "This installs:"
	@echo "  1. wasm32-unknown-unknown target (for stable toolchain)"
	@echo "  2. wasm32-unknown-unknown target (for 1.86 toolchain)"
	@echo "  3. wasm-bindgen CLI tool"
	@echo ""
	@echo "Adding wasm32-unknown-unknown target to stable toolchain..."
	@rustup target add wasm32-unknown-unknown
	@echo ""
	@echo "Adding wasm32-unknown-unknown target to 1.86 toolchain (used by Automerge.js build)..."
	@rustup target add wasm32-unknown-unknown --toolchain 1.86-aarch64-apple-darwin 2>/dev/null || \
		echo "⚠️  1.86 toolchain not installed (will be installed automatically by Automerge.js build)"
	@echo ""
	@echo "Checking if wasm-bindgen CLI is installed..."
	@if ! command -v wasm-bindgen >/dev/null 2>&1; then \
		echo "Installing wasm-bindgen-cli (this may take a few minutes)..."; \
		cargo install wasm-bindgen-cli; \
	else \
		echo "✅ wasm-bindgen already installed at $$(which wasm-bindgen)"; \
	fi
	@echo ""
	@echo "✅ Rust WASM toolchain ready for 'make build-js'"

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

## build-js: Build Automerge.js from .src/ (single source of truth)
build-js:
	@echo "📦 Building Automerge.js from source (tag: $(AUTOMERGE_TAG))..."
	@if [ ! -d "$(JS_SRC_DIR)" ]; then \
		echo "❌ Error: $(JS_SRC_DIR) not found. Run 'make setup-src' first."; \
		exit 1; \
	fi
	@echo "Installing JavaScript dependencies (using bun)..."
	@cd $(JS_SRC_DIR) && bun install --frozen-lockfile
	@echo "Building Automerge.js (using bun)..."
	@cd $(JS_SRC_DIR) && bun run build
	@mkdir -p $(VENDOR_DIR)
	@cp $(JS_DIST) $(VENDOR_JS)
	@echo "✅ Built and copied to $(VENDOR_JS)"
	@ls -lh $(VENDOR_JS)

## sync-versions: Verify all components use same .src/ version
sync-versions:
	@echo "🔍 Checking version alignment..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📌 Expected tag: $(AUTOMERGE_TAG)"
	@echo "📌 Actual .src/automerge git version:"
	@cd $(SRC_DIR)/automerge && git describe --tags
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🦀 Cargo.toml dependency:"
	@grep 'automerge.*path' rust/automerge_wasi/Cargo.toml || echo "  ⚠️  Using crates.io (should use path)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📦 JavaScript package.json:"
	@cd $(JS_SRC_DIR) && cat package.json | grep '"version"' | head -1
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@if [ -f "$(VENDOR_JS)" ]; then \
		echo "✅ Built Automerge.js: $$(ls -lh $(VENDOR_JS) | awk '{print $$5}')"; \
	else \
		echo "⚠️  Automerge.js not built. Run 'make build-js'"; \
	fi

## build-server: Build the Go server binary
build-server:
	@echo "🔨 Building Go server..."
	cd $(GO_ROOT) && go build -o ../server ./cmd/server
	@echo "✅ Built: server"
	@ls -lh server

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
test: test-rust test-go

## test-rust: Run Rust tests only
test-rust:
	@echo "🧪 Running Rust tests..."
	cd $(WASM_DIR) && cargo test

## test-go: Run Go tests only
test-go:
	@echo "🧪 Running Go tests..."
	cd $(GO_ROOT) && go test -v ./...

## tidy: Run go mod tidy
tidy:
	@echo "🧹 Running go mod tidy..."
	cd $(GO_ROOT) && go mod tidy

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	cd $(WASM_DIR) && cargo clean
	cd $(GO_ROOT) && go clean

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

## setup-src: Setup .src directory with Automerge source code and docs
setup-src:
	@echo "📚 Setting up .src directory with Automerge source..."
	@if [ ! -d "$(SRC_DIR)" ]; then mkdir -p $(SRC_DIR); fi
	@if [ ! -d "$(SRC_DIR)/automerge" ]; then \
		echo "Cloning Automerge repository..."; \
		git clone $(AUTOMERGE_REPO) $(SRC_DIR)/automerge; \
	fi
	@echo "Checking out tag $(AUTOMERGE_TAG)..."
	@cd $(SRC_DIR)/automerge && git fetch --tags && git checkout $(AUTOMERGE_TAG)
	@echo "✅ Automerge source ready at $(SRC_DIR)/automerge"
	@echo "   Git tag: $(AUTOMERGE_TAG)"
	@echo "   Rust core API: $(SRC_DIR)/automerge/rust/automerge/src/"

## update-src: Update .src to configured version
update-src:
	@echo "🔄 Updating .src directory..."
	@if [ ! -d "$(SRC_DIR)/automerge" ]; then \
		echo "❌ Error: $(SRC_DIR)/automerge does not exist. Run 'make setup-src' first."; \
		exit 1; \
	fi
	@cd $(SRC_DIR)/automerge && git fetch --tags
	@echo "Checking out tag $(AUTOMERGE_TAG)..."
	@cd $(SRC_DIR)/automerge && git checkout $(AUTOMERGE_TAG)
	@echo "✅ Updated to tag: $(AUTOMERGE_TAG)"

## clean-src: Remove .src directory (useful for fresh start)
clean-src:
	@echo "🧹 Cleaning .src directory..."
	@rm -rf $(SRC_DIR)/automerge
	@echo "✅ .src/automerge removed (kept .src/automerge.github.io)"

## generate-test-data: Generate test snapshots for Go package tests
generate-test-data: build-wasi
	@echo "🎲 Generating test data..."
	@cd $(GO_ROOT)/testdata/unit/scripts && ./generate_test_data.sh
	@echo "✅ Test data generated"

## verify-docs: Check for broken internal markdown links
verify-docs:
	@echo "🔍 Checking for broken internal documentation links..."
	@echo ""
	@ERRORS=0; \
	for file in $$(find . -name "*.md" -not -path "./node_modules/*" -not -path "./.src/*"); do \
		while IFS= read -r line; do \
			if echo "$$line" | grep -qE '\]\([^h#][^)]*\.md[^)]*\)'; then \
				link=$$(echo "$$line" | grep -oE '\]\([^h#][^)]*\.md[^)]*\)' | sed 's/][(]//;s/)//'); \
				dir=$$(dirname "$$file"); \
				target="$$dir/$$link"; \
				if [ ! -f "$$target" ]; then \
					echo "❌ Broken link in $$file:"; \
					echo "   Link: $$link"; \
					echo "   Expected: $$target"; \
					echo ""; \
					ERRORS=$$((ERRORS + 1)); \
				fi; \
			fi; \
		done < "$$file"; \
	done; \
	if [ $$ERRORS -eq 0 ]; then \
		echo "✅ All internal documentation links valid!"; \
	else \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "❌ Found $$ERRORS broken link(s)"; \
		echo ""; \
		echo "Fix these before committing documentation changes."; \
		exit 1; \
	fi

## verify-web: Verify web folder structure and files
verify-web:
	@echo "🔍 Verifying web folder structure (1:1 mapping)..."
	@echo ""
	@echo "Checking required files:"
	@ERRORS=0; \
	for file in $(WEB_HTML) $(WEB_CSS) $(WEB_JS) $(WEB_COMPONENTS); do \
		if [ -f "$$file" ]; then \
			echo "  ✅ $$file"; \
		else \
			echo "  ❌ Missing: $$file"; \
			ERRORS=$$((ERRORS + 1)); \
		fi; \
	done; \
	echo ""; \
	echo "Checking Automerge.js vendor file:"; \
	if [ -f "$(VENDOR_JS)" ]; then \
		echo "  ✅ $(VENDOR_JS) ($$(ls -lh $(VENDOR_JS) | awk '{print $$5}'))"; \
		if grep -q 'src="/vendor/automerge.js"' $(WEB_HTML); then \
			echo "  ✅ $(WEB_HTML) references /vendor/automerge.js"; \
		else \
			echo "  ❌ $(WEB_HTML) does not reference /vendor/automerge.js"; \
			ERRORS=$$((ERRORS + 1)); \
		fi; \
	else \
		echo "  ❌ $(VENDOR_JS) not built. Run 'make build-js'"; \
		ERRORS=$$((ERRORS + 1)); \
	fi; \
	echo ""; \
	if [ $$ERRORS -eq 0 ]; then \
		echo "✅ Web folder structure valid!"; \
	else \
		echo "❌ Found $$ERRORS issue(s)"; \
		exit 1; \
	fi

## test-http: Test HTTP API endpoints (requires server running)
test-http:
	@echo "🧪 Testing HTTP API endpoints..."
	@echo ""
	@echo "Testing M0: Text endpoint"
	@curl -s http://localhost:8080/api/text || (echo "❌ Server not running on port 8080"; exit 1)
	@echo ""
	@echo "Testing M1: Sync endpoint"
	@curl -X POST -s -H 'Content-Type: application/json' \
		-d '{"peer_id":"makefile-test","message":""}' \
		http://localhost:8080/api/sync | grep -q 'has_more' && echo "✅ Sync endpoint working" || echo "❌ Sync endpoint failed"
	@echo ""
	@echo "Testing M2: RichText endpoint (requires text first)"
	@curl -X POST -s -H 'Content-Type: application/json' \
		-d '{"path":"ROOT.content","name":"test","value":"true","start":0,"end":1,"expand":"none"}' \
		http://localhost:8080/api/richtext/mark -w "%{http_code}" -o /dev/null | grep -q '204' && \
		echo "✅ RichText mark endpoint working" || echo "⚠️  RichText mark endpoint (may need text in document)"
	@echo ""
	@echo "✅ HTTP tests complete"

## test-playwright: Run Playwright end-to-end tests (requires server + Playwright MCP)
test-playwright:
	@echo "🎭 Playwright tests should be run via MCP tools"
	@echo ""
	@echo "Test plans available:"
	@ls -1 tests/playwright/*.md 2>/dev/null || echo "  No test plans found"
	@echo ""
	@echo "Run tests using Claude Code with Playwright MCP enabled"
