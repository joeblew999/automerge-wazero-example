package automerge

// Test helpers for automerge package tests

const (
	// TestWASMPath is the path to the WASM module used in tests.
	// This points to the debug build for faster test compilation.
	// Path is relative from go/pkg/automerge/ directory where tests run.
	TestWASMPath = "../../../rust/automerge_wasi/target/wasm32-wasip1/debug/automerge_wasi.wasm"
)
