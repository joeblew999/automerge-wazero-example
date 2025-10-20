//! Global document state management
//!
//! This module manages the thread-local document state for the WASI interface.
//! Currently supports a single document per thread (sufficient for wazero single-instance use).
//!
//! ## Future: Multi-Document Support (M3)
//!
//! For M3, this will be refactored to support multiple documents identified by doc_id.

use automerge::{AutoCommit, ObjId};
use std::cell::RefCell;

// Global document storage (thread-local for WASI single-threaded execution)
thread_local! {
    pub(crate) static DOC: RefCell<Option<AutoCommit>> = RefCell::new(None);
    pub(crate) static TEXT_OBJ_ID: RefCell<Option<ObjId>> = RefCell::new(None);
}

/// Initialize the global document with a new AutoCommit instance
pub(crate) fn init_doc(doc: AutoCommit) {
    DOC.with(|cell| {
        *cell.borrow_mut() = Some(doc);
    });
}

/// Get a reference to the global document
pub(crate) fn with_doc<F, R>(f: F) -> Option<R>
where
    F: FnOnce(&AutoCommit) -> R,
{
    DOC.with(|cell| {
        let doc_ref = cell.borrow();
        doc_ref.as_ref().map(f)
    })
}

/// Get a mutable reference to the global document
pub(crate) fn with_doc_mut<F, R>(f: F) -> Option<R>
where
    F: FnOnce(&mut AutoCommit) -> R,
{
    DOC.with(|cell| {
        let mut doc_ref = cell.borrow_mut();
        doc_ref.as_mut().map(f)
    })
}

/// Set the text object ID (for ROOT["content"] compatibility)
pub(crate) fn set_text_obj_id(id: ObjId) {
    TEXT_OBJ_ID.with(|cell| {
        *cell.borrow_mut() = Some(id);
    });
}

/// Get the text object ID
pub(crate) fn get_text_obj_id() -> Option<ObjId> {
    TEXT_OBJ_ID.with(|cell| {
        cell.borrow().clone()
    })
}

/// Clear all global state (reserved for future cleanup/reset functionality)
#[allow(dead_code)]
pub(crate) fn clear_state() {
    DOC.with(|cell| *cell.borrow_mut() = None);
    TEXT_OBJ_ID.with(|cell| *cell.borrow_mut() = None);
}
