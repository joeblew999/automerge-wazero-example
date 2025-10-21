//! Automerge WASI FFI Layer
//!
//! This crate provides a C-compatible FFI interface to Automerge for use with WASM/WASI.
//! It exposes the Automerge CRDT functionality to Go via wazero.
//!
//! ## Architecture
//!
//! ```text
//! Go (pkg/automerge) → wazero FFI → WASI exports (this crate) → Automerge Rust
//! ```
//!
//! ## Modules
//!
//! - `memory` - Memory management (alloc/free)
//! - `document` - Document lifecycle (init, save, load, merge)
//! - `text` - Text CRDT operations
//! - `map` - Map operations (M2)
//! - `list` - List operations (M2)
//! - `counter` - Counter CRDT (M2)
//! - `sync` - Sync protocol (M1)
//! - `state` - Global document state management
//!
//! ## Current Status
//!
//! **Implemented** (M0):
//! - Memory management
//! - Document lifecycle
//! - Text operations (splice, get, length)
//! - Save/Load
//! - Merge (partial - needs investigation)
//!
//! **Planned**:
//! - M1: Sync protocol exports
//! - M2: Maps, Lists, Counters
//! - M3: Multi-document support
//! - M4: Rich text formatting

// FFI-specific lint allows
// These are intentional for C ABI compatibility
#![allow(clippy::not_unsafe_ptr_arg_deref)] // FFI functions take raw pointers by design
#![allow(clippy::missing_const_for_thread_local)] // thread_local requires runtime initialization
#![allow(clippy::manual_unwrap_or)] // Match expressions are clearer for error codes
#![allow(clippy::collapsible_if)] // Separate conditionals for clarity
#![allow(clippy::uninlined_format_args)] // Explicit format args in tests for clarity

mod memory;
mod state;
mod document;
mod text;
mod map;
mod list;
mod counter;
mod history;
mod sync;
mod richtext;
mod cursor;
mod generic;

// Re-export all public FFI functions
pub use memory::*;
pub use document::*;
pub use text::*;
pub use map::*;
pub use list::*;
pub use counter::*;
pub use history::*;
pub use sync::*;
pub use richtext::*;
pub use cursor::*;
pub use generic::*;
