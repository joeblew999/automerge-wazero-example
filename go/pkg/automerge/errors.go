package automerge

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// Implementation status
	ErrNotImplemented = errors.New("automerge: feature not yet implemented")
	ErrDeprecated     = errors.New("automerge: deprecated - use alternative method")

	// Runtime errors
	ErrNotInitialized = errors.New("automerge: document not initialized")
	ErrInvalidPath    = errors.New("automerge: invalid object path")
	ErrInvalidUTF8    = errors.New("automerge: invalid UTF-8 data")
	ErrWASMCall       = errors.New("automerge: WASM call failed")
	ErrMemoryAlloc    = errors.New("automerge: WASM memory allocation failed")

	// CRDT errors
	ErrMergeFailed = errors.New("automerge: merge failed")
	ErrLoadFailed  = errors.New("automerge: failed to load document")
	ErrSaveFailed  = errors.New("automerge: failed to save document")

	// Object/value errors
	ErrObjectNotFound = errors.New("automerge: object not found")
	ErrKeyNotFound    = errors.New("automerge: key not found")
	ErrTypeMismatch   = errors.New("automerge: type mismatch")
	ErrIndexOutOfBounds = errors.New("automerge: index out of bounds")
)

// NotImplementedError provides context about unimplemented features
type NotImplementedError struct {
	Feature   string // e.g., "SpliceText", "Put", "Mark"
	Milestone string // "Current", "M1", "M2", "M3", "M4"
	Message   string // Additional context
}

func (e *NotImplementedError) Error() string {
	if e.Milestone != "" && e.Milestone != "Current" {
		return fmt.Sprintf("automerge: %s not implemented (planned for %s): %s",
			e.Feature, e.Milestone, e.Message)
	}
	return fmt.Sprintf("automerge: %s not implemented: %s", e.Feature, e.Message)
}

func (e *NotImplementedError) Is(target error) bool {
	return target == ErrNotImplemented
}

// DeprecatedError indicates a deprecated method with recommended alternative
type DeprecatedError struct {
	Method      string // e.g., "UpdateText"
	Alternative string // e.g., "SpliceText"
	Reason      string // Why it's deprecated
}

func (e *DeprecatedError) Error() string {
	return fmt.Sprintf("automerge: %s is deprecated (%s), use %s instead",
		e.Method, e.Reason, e.Alternative)
}

func (e *DeprecatedError) Is(target error) bool {
	return target == ErrDeprecated
}

// WASMError wraps WASM-level errors with additional context
type WASMError struct {
	Operation string // e.g., "am_text_splice", "am_save"
	Code      int32  // WASM error code
	Err       error  // Underlying error
}

func (e *WASMError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("automerge: WASM operation %s failed (code %d): %v",
			e.Operation, e.Code, e.Err)
	}
	return fmt.Sprintf("automerge: WASM operation %s failed (code %d)",
		e.Operation, e.Code)
}

func (e *WASMError) Unwrap() error {
	return e.Err
}

func (e *WASMError) Is(target error) bool {
	return target == ErrWASMCall
}
