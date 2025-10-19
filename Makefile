.PHONY: build-wasi run tidy clean clean-snapshots test

# Using wasm32-wasip1 for Rust 1.84+
WASI_TARGET = wasm32-wasip1
WASM_FILE = rust/automerge_wasi/target/$(WASI_TARGET)/release/automerge_wasi.wasm

build-wasi:
	@echo "Building Rust WASI module..."
	cd rust/automerge_wasi && cargo build --target $(WASI_TARGET) --release

run: build-wasi
	@echo "Running Go server with wazero..."
	cd go/cmd/server && go run main.go

tidy:
	@echo "Running go mod tidy..."
	cd go/cmd/server && go mod tidy

test:
	@echo "Running tests..."
	cd rust/automerge_wasi && cargo test
	cd go/cmd/server && go test ./...

clean:
	@echo "Cleaning build artifacts..."
	cd rust/automerge_wasi && cargo clean
	cd go/cmd/server && go clean

clean-snapshots:
	@echo "Cleaning snapshot files..."
	rm -f data/*.am
	rm -f doc.am

dev: build-wasi
	@echo "Running in dev mode..."
	cd go/cmd/server && go run main.go
