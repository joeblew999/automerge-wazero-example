//! Document lifecycle management
//!
//! Handles document creation, serialization, loading, and merging.

use automerge::{AutoCommit, ObjType, ReadDoc, transaction::Transactable};
use crate::state::{init_doc, with_doc_mut, set_text_obj_id};

/// Initialize a new Automerge document with a "content" Text CRDT object
///
/// This creates a new `AutoCommit` document and adds a Text object at ROOT["content"]
/// for backward compatibility with the existing single-text API.
///
/// ## Returns
/// - `0` on success
/// - `-1` if failed to create text object
///
/// ## Example
/// ```c
/// int result = am_init();
/// if (result != 0) {
///     // Handle error
/// }
/// ```
#[no_mangle]
pub extern "C" fn am_init() -> i32 {
    let mut doc = AutoCommit::new();

    // Create a Text object (CRDT) at key "content"
    let text_obj_id = match doc.put_object(automerge::ROOT, "content", ObjType::Text) {
        Ok(id) => id,
        Err(_) => return -1,
    };

    // Store the text object ID for later operations
    set_text_obj_id(text_obj_id);
    init_doc(doc);

    0
}

/// Get the length of the serialized document
///
/// Used by Go to allocate a buffer before calling `am_save`.
///
/// ## Returns
/// - Size in bytes of the serialized document
/// - `0` if document not initialized
#[no_mangle]
pub extern "C" fn am_save_len() -> u32 {
    with_doc_mut(|doc| {
        doc.save().len() as u32
    }).unwrap_or(0)
}

/// Save the document to a buffer
///
/// Caller must allocate a buffer of size `am_save_len()` using `am_alloc`.
///
/// ## Parameters
/// - `ptr_out`: Pointer to output buffer
///
/// ## Returns
/// - `0` on success
/// - `-1` if ptr_out is null
/// - `-2` if document not initialized
#[no_mangle]
pub extern "C" fn am_save(ptr_out: *mut u8) -> i32 {
    if ptr_out.is_null() {
        return -1;
    }

    match with_doc_mut(|doc| {
        let bytes = doc.save();
        unsafe {
            std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
        }
    }) {
        Some(_) => 0,
        None => -2, // Document not initialized
    }
}

/// Load a document from a buffer
///
/// Replaces the current document with one loaded from the given bytes.
/// Also extracts the "content" text object ID for backward compatibility.
///
/// ## Parameters
/// - `ptr`: Pointer to serialized document bytes
/// - `len`: Length of the buffer
///
/// ## Returns
/// - `0` on success
/// - `-1` if ptr is null
/// - `-2` if failed to load document
/// - `-3` if "content" text object not found
#[no_mangle]
pub extern "C" fn am_load(ptr: *const u8, len: usize) -> i32 {
    if ptr.is_null() {
        return -1;
    }

    let slice = unsafe { std::slice::from_raw_parts(ptr, len) };

    match AutoCommit::load(slice) {
        Ok(doc) => {
            // After loading, find the text object ID
            // Assuming the text field is at key "content"
            match doc.get(automerge::ROOT, "content") {
                Ok(Some((_, obj_id))) => {
                    set_text_obj_id(obj_id);
                    init_doc(doc);
                    0
                }
                _ => -3, // Text object not found
            }
        }
        Err(_) => -2, // Failed to load
    }
}

/// Merge another document into the current document
///
/// This is the CRDT magic! Two diverged documents can be merged without conflicts.
/// The merge is deterministic and commutative.
///
/// ## Parameters
/// - `other_ptr`: Pointer to serialized document bytes to merge
/// - `other_len`: Length of the buffer
///
/// ## Returns
/// - `0` on success
/// - `-1` if other_ptr is null
/// - `-2` if failed to load other document
/// - `-3` if current document not initialized
/// - `-4` if "content" text object not found after merge
///
/// ## Known Issues
/// - Currently only preserves one document's changes (needs investigation)
/// - See TestDocument_Merge in Go tests
#[no_mangle]
pub extern "C" fn am_merge(other_ptr: *const u8, other_len: usize) -> i32 {
    if other_ptr.is_null() {
        return -1;
    }

    let other_slice = unsafe { std::slice::from_raw_parts(other_ptr, other_len) };

    // Load the other document
    let mut other_doc = match AutoCommit::load(other_slice) {
        Ok(d) => d,
        Err(_) => return -2, // Failed to load other document
    };

    match with_doc_mut(|doc| {
        // Perform CRDT merge
        let _ = doc.merge(&mut other_doc);

        // After merge, update text_id in case it changed
        match doc.get(automerge::ROOT, "content") {
            Ok(Some((_, obj_id))) => {
                set_text_obj_id(obj_id);
                0
            }
            _ => -4, // Text object not found after merge
        }
    }) {
        Some(result) => result,
        None => -3, // Current document not initialized
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_init() {
        let result = am_init();
        assert_eq!(result, 0);
    }

    #[test]
    fn test_save_load() {
        // Init document
        assert_eq!(am_init(), 0);

        // Get save length
        let save_len = am_save_len();
        assert!(save_len > 0);

        // Allocate buffer and save
        let buffer = vec![0u8; save_len as usize];
        let ptr = buffer.as_ptr() as *mut u8;
        assert_eq!(am_save(ptr), 0);

        // Load into new document
        assert_eq!(am_load(buffer.as_ptr(), save_len as usize), 0);
    }
}
