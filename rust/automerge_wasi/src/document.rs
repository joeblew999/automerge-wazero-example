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

/// Get the actor ID length for the current document
///
/// Used to allocate a buffer before calling `am_get_actor`.
///
/// ## Returns
/// - Length of actor ID string in bytes
/// - `0` if document not initialized
#[no_mangle]
pub extern "C" fn am_get_actor_len() -> u32 {
    with_doc_mut(|doc| {
        doc.get_actor().to_string().len() as u32
    }).unwrap_or(0)
}

/// Get the actor ID for the current document
///
/// The actor ID uniquely identifies this peer in the distributed system.
/// It's used to track which changes came from which peer.
///
/// ## Parameters
/// - `ptr_out`: Pointer to output buffer (must be at least `am_get_actor_len()` bytes)
///
/// ## Returns
/// - `0` on success
/// - `-1` if ptr_out is null
/// - `-2` if document not initialized
#[no_mangle]
pub extern "C" fn am_get_actor(ptr_out: *mut u8) -> i32 {
    if ptr_out.is_null() {
        return -1;
    }

    match with_doc_mut(|doc| {
        let actor_str = doc.get_actor().to_string();
        let bytes = actor_str.as_bytes();
        unsafe {
            std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
        }
    }) {
        Some(_) => 0,
        None => -2, // Document not initialized
    }
}

/// Set the actor ID for the current document
///
/// Changes the actor ID that will be used for future operations.
/// This should be set before making any changes to the document.
///
/// ## Parameters
/// - `actor_ptr`: Pointer to actor ID string (hex format, e.g., "0123456789abcdef...")
/// - `actor_len`: Length of the actor ID string
///
/// ## Returns
/// - `0` on success
/// - `-1` if actor_ptr is null
/// - `-2` if invalid actor ID format
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_set_actor(actor_ptr: *const u8, actor_len: usize) -> i32 {
    if actor_ptr.is_null() {
        return -1;
    }

    let actor_slice = unsafe { std::slice::from_raw_parts(actor_ptr, actor_len) };
    let actor_str = match std::str::from_utf8(actor_slice) {
        Ok(s) => s,
        Err(_) => return -2,
    };

    let actor_id = match actor_str.try_into() {
        Ok(id) => id,
        Err(_) => return -2, // Invalid actor ID format
    };

    match with_doc_mut(|doc| {
        doc.set_actor(actor_id);
    }) {
        Some(_) => 0,
        None => -3, // Document not initialized
    }
}

/// Fork the current document
///
/// Creates an independent copy of the document with a new actor ID.
/// The forked document can diverge independently and be merged back later.
///
/// This replaces the current document with the fork.
///
/// ## Returns
/// - `0` on success
/// - `-1` if document not initialized
#[no_mangle]
pub extern "C" fn am_fork() -> i32 {
    match with_doc_mut(|doc| {
        let forked = doc.fork();
        *doc = forked;
    }) {
        Some(_) => 0,
        None => -1, // Document not initialized
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

    #[test]
    fn test_get_set_actor() {
        assert_eq!(am_init(), 0);

        // Get initial actor
        let actor_len = am_get_actor_len();
        assert!(actor_len > 0);

        let mut buffer = vec![0u8; actor_len as usize];
        assert_eq!(am_get_actor(buffer.as_mut_ptr()), 0);
        let actor1 = String::from_utf8(buffer).unwrap();

        // Set new actor
        let new_actor = "0123456789abcdef0123456789abcdef";
        assert_eq!(am_set_actor(new_actor.as_ptr(), new_actor.len()), 0);

        // Verify it changed
        let actor_len2 = am_get_actor_len();
        let mut buffer2 = vec![0u8; actor_len2 as usize];
        assert_eq!(am_get_actor(buffer2.as_mut_ptr()), 0);
        let actor2 = String::from_utf8(buffer2).unwrap();

        assert_ne!(actor1, actor2);
    }

    #[test]
    fn test_fork() {
        assert_eq!(am_init(), 0);

        // Get original actor
        let actor_len = am_get_actor_len();
        let mut buffer = vec![0u8; actor_len as usize];
        assert_eq!(am_get_actor(buffer.as_mut_ptr()), 0);
        let actor1 = String::from_utf8(buffer).unwrap();

        // Fork
        assert_eq!(am_fork(), 0);

        // Verify fork has different actor
        let actor_len2 = am_get_actor_len();
        let mut buffer2 = vec![0u8; actor_len2 as usize];
        assert_eq!(am_get_actor(buffer2.as_mut_ptr()), 0);
        let actor2 = String::from_utf8(buffer2).unwrap();

        assert_ne!(actor1, actor2, "Forked document should have different actor ID");
    }
}
