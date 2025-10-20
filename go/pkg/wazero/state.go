// Package wazero provides low-level FFI bindings to the Automerge WASI module.
// This package wraps the WASM exports directly with minimal abstraction.
package wazero

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const (
	// Default WASM module path (can be overridden)
	defaultWASMPath = "../../../rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm"
)

// Runtime wraps the Wazero runtime and provides access to WASM exports
type Runtime struct {
	runtime wazero.Runtime
	module  wazero.CompiledModule
	modInst api.Module
}

// Config for runtime initialization
type Config struct {
	WASMPath string // Path to .wasm file (optional, uses default if empty)
}

// New creates a new Wazero runtime and loads the Automerge WASI module
func New(ctx context.Context, cfg Config) (*Runtime, error) {
	if cfg.WASMPath == "" {
		cfg.WASMPath = defaultWASMPath
	}

	// Create Wazero runtime
	runtime := wazero.NewRuntime(ctx)

	// Instantiate WASI
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		runtime.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	// Load WASM bytes (you could embed this with //go:embed)
	wasmBytes, err := loadWASM(cfg.WASMPath)
	if err != nil {
		runtime.Close(ctx)
		return nil, fmt.Errorf("failed to load WASM: %w", err)
	}

	// Compile module
	compiled, err := runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		runtime.Close(ctx)
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Instantiate module
	modInst, err := runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	if err != nil {
		runtime.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	return &Runtime{
		runtime: runtime,
		module:  compiled,
		modInst: modInst,
	}, nil
}

// Close closes the runtime and frees resources
func (r *Runtime) Close(ctx context.Context) error {
	return r.runtime.Close(ctx)
}

// Memory returns the WASM linear memory
func (r *Runtime) Memory() api.Memory {
	return r.modInst.Memory()
}

// callExport is a helper to call a WASM export and check for errors
func (r *Runtime) callExport(ctx context.Context, name string, params ...uint64) ([]uint64, error) {
	fn := r.modInst.ExportedFunction(name)
	if fn == nil {
		return nil, fmt.Errorf("export %s not found", name)
	}

	results, err := fn.Call(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("call to %s failed: %w", name, err)
	}

	return results, nil
}

// checkErrorCode checks if the first result is a non-zero error code
func checkErrorCode(name string, results []uint64) error {
	if len(results) == 0 {
		return nil
	}

	code := int32(results[0])
	if code != 0 {
		return &WASMError{
			Operation: name,
			Code:      code,
		}
	}
	return nil
}

// WASMError represents an error returned from a WASM function
type WASMError struct {
	Operation string
	Code      int32
}

func (e *WASMError) Error() string {
	return fmt.Sprintf("WASM operation %s failed with code %d", e.Operation, e.Code)
}

// loadWASM loads the WASM bytes from a file
func loadWASM(path string) ([]byte, error) {
	// Note: In production, consider embedding the WASM with //go:embed
	return os.ReadFile(path)
}
