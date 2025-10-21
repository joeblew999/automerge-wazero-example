// ═══════════════════════════════════════════════════════════════
// LAYER 2: Rust WASI Exports (C-ABI for FFI)
//
// Responsibilities:
// - Export C-ABI functions callable from Go via wazero
// - Validate UTF-8 input from Go side
// - Call Automerge Rust API for CRDT operations
// - Return error codes as i32 (0 = success, <0 = error)
//
// Dependencies:
// ⬇️  Calls: automerge crate (Layer 1 - CRDT core)
// ⬆️  Called by: go/pkg/wazero/text.go (Layer 3 - Go FFI wrappers)
//
// Related Files:
// 🔁 Siblings: map.rs, list.rs, counter.rs, sync.rs, richtext.rs
// 📝 Tests: cargo test (Rust unit tests)
// 🔗 Docs: docs/explanation/architecture.md#layer-2-rust-wasi
// ═══════════════════════════════════════════════════════════════

//! Text CRDT operations
//!
//! Provides splice, get, and length operations for Text objects.
//! Currently operates on ROOT["content"] for backward compatibility.

use automerge::{ReadDoc, transaction::Transactable};
use crate::state::{with_doc, with_doc_mut, get_text_obj_id};

/// Splice text at a given position (proper Text CRDT operation)
///
/// This is the primary text editing operation. It can insert, delete, or replace text.
///
/// ## Parameters
/// - `pos`: Byte position to start (0-based)
/// - `del_count`: Number of UTF-8 characters to delete (can be 0)
/// - `insert_ptr`: Pointer to string to insert (can be null if insert_len is 0)
/// - `insert_len`: Length of string to insert (can be 0)
///
/// ## Returns
/// - `0` on success
/// - `-2` if insert text is not valid UTF-8
/// - `-3` if document not initialized
/// - `-4` if text object not initialized
/// - `-5` if splice operation failed
/// - `-6` if del_count conversion failed
///
/// ## Examples
///
/// Insert at position 0:
/// ```c
/// am_text_splice(0, 0, "Hello", 5);
/// ```
///
/// Delete 5 characters at position 0:
/// ```c
/// am_text_splice(0, 5, NULL, 0);
/// ```
///
/// Replace 5 characters at position 0:
/// ```c
/// am_text_splice(0, 5, "Hi", 2);
/// ```
#[no_mangle]
pub extern "C" fn am_text_splice(
    pos: usize,
    del_count: i64,
    insert_ptr: *const u8,
    insert_len: usize
) -> i32 {
    let insert_text = if insert_len > 0 && !insert_ptr.is_null() {
        let slice = unsafe { std::slice::from_raw_parts(insert_ptr, insert_len) };
        match std::str::from_utf8(slice) {
            Ok(s) => s,
            Err(_) => return -2, // Invalid UTF-8
        }
    } else {
        ""
    };

    let text_id = match get_text_obj_id() {
        Some(id) => id,
        None => return -4, // Text object not initialized
    };

    // Convert i64 to isize for delete count
    let del_count_isize = match del_count.try_into() {
        Ok(n) => n,
        Err(_) => return -6, // Invalid delete count
    };

    match with_doc_mut(|doc| {
        doc.splice_text(&text_id, pos, del_count_isize, insert_text)
    }) {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -5, // Splice operation failed
        None => -3, // Document not initialized
    }
}

/// DEPRECATED: Set the entire text content (for backward compatibility)
///
/// Use `am_text_splice` for proper Text CRDT operations.
/// This function deletes all existing text and inserts new text.
///
/// ## Parameters
/// - `ptr`: Pointer to new text content
/// - `len`: Length of new text
///
/// ## Returns
/// - `0` on success
/// - `-1` if ptr is null
/// - `-2` if splice failed
///
/// ## Deprecation
/// This destroys CRDT history. Use `am_text_splice` instead.
#[no_mangle]
pub extern "C" fn am_set_text(ptr: *const u8, len: usize) -> i32 {
    if ptr.is_null() {
        return -1;
    }

    // Get current text length to delete all
    let current_len = am_get_text_len() as usize;

    // Delete all existing text, then insert new text
    if current_len > 0 {
        if am_text_splice(0, current_len as i64, std::ptr::null(), 0) != 0 {
            return -2;
        }
    }

    // Insert new text at position 0
    am_text_splice(0, 0, ptr, len)
}

/// Get the length of the text content (in bytes)
///
/// ## Returns
/// - Length of text in bytes
/// - `0` if document or text object not initialized
#[no_mangle]
pub extern "C" fn am_get_text_len() -> u32 {
    let text_id = match get_text_obj_id() {
        Some(id) => id,
        None => return 0,
    };

    with_doc(|doc| {
        doc.text(&text_id)
            .map(|s| s.len() as u32)
            .unwrap_or(0)
    }).unwrap_or(0)
}

/// Get the text content
///
/// Caller must allocate a buffer of size `am_get_text_len()` using `am_alloc`.
///
/// ## Parameters
/// - `ptr_out`: Pointer to output buffer
///
/// ## Returns
/// - `0` on success
/// - `-1` if ptr_out is null
/// - `-2` if document not initialized
/// - `-3` if text object not initialized
/// - `-4` if failed to get text
#[no_mangle]
pub extern "C" fn am_get_text(ptr_out: *mut u8) -> i32 {
    if ptr_out.is_null() {
        return -1;
    }

    let text_id = match get_text_obj_id() {
        Some(id) => id,
        None => return -3, // Text object not initialized
    };

    match with_doc(|doc| {
        doc.text(&text_id)
            .map(|s| {
                let bytes = s.as_bytes();
                unsafe {
                    std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
                }
            })
    }) {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -4, // Failed to get text
        None => -2, // Document not initialized
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::document::am_init;

    #[test]
    fn test_text_splice() {
        // Initialize document
        assert_eq!(am_init(), 0);

        // Insert "Hello"
        let text = b"Hello";
        assert_eq!(am_text_splice(0, 0, text.as_ptr(), text.len()), 0);

        // Check length
        assert_eq!(am_get_text_len(), 5);

        // Get text
        let mut buffer = vec![0u8; 5];
        assert_eq!(am_get_text(buffer.as_mut_ptr()), 0);
        assert_eq!(&buffer, b"Hello");
    }

    #[test]
    fn test_text_splice_unicode() {
        assert_eq!(am_init(), 0);

        let text = "Hello 世界! 🌍".as_bytes();
        assert_eq!(am_text_splice(0, 0, text.as_ptr(), text.len()), 0);

        let len = am_get_text_len() as usize;
        let mut buffer = vec![0u8; len];
        assert_eq!(am_get_text(buffer.as_mut_ptr()), 0);
        assert_eq!(std::str::from_utf8(&buffer).unwrap(), "Hello 世界! 🌍");
    }

    #[test]
    fn test_set_text_deprecated() {
        assert_eq!(am_init(), 0);

        // Set text
        let text = b"World";
        assert_eq!(am_set_text(text.as_ptr(), text.len()), 0);

        // Verify
        assert_eq!(am_get_text_len(), 5);
    }
}
