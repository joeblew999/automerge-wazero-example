// Cursor operations for stable position tracking across concurrent edits
//
// Cursors are essential for collaborative editing - they represent positions
// that remain stable even as other users edit the document.
//
// Unlike character offsets which change when text is inserted/deleted,
// cursors track CRDT positions that survive concurrent modifications.

use crate::state::with_doc;
use automerge::{ObjId, ReadDoc};
use std::str;

/// Get a cursor for a position in a text or list object
///
/// # Arguments
/// * `obj_ptr` - Pointer to object path string (e.g., "ROOT.content")
/// * `obj_len` - Length of object path string
/// * `index` - Position (character index for text, item index for lists)
///
/// # Returns
/// * Positive: Length of cursor string (call am_get_cursor_str to retrieve)
/// * -1: Invalid path
/// * -2: Invalid index
/// * -3: Not a text or list object
#[no_mangle]
pub extern "C" fn am_get_cursor(obj_ptr: *const u8, obj_len: usize, index: usize) -> i32 {
    let path_slice = unsafe { std::slice::from_raw_parts(obj_ptr, obj_len) };
    let path_str = match str::from_utf8(path_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    with_doc(|doc| {
        // Parse path to get object ID
        let obj_id = match parse_path(doc, path_str) {
            Ok(id) => id,
            Err(_) => return -1,
        };

        // Get cursor at index (None means current heads)
        match doc.get_cursor(&obj_id, index, None) {
            Ok(cursor) => {
                let cursor_str = cursor.to_string();
                // Store cursor string in thread-local for retrieval
                LAST_CURSOR.with(|c| {
                    *c.borrow_mut() = cursor_str.clone();
                });
                cursor_str.len() as i32
            }
            Err(_) => -2, // Invalid index or not a sequence
        }
    }).unwrap_or(-1)
}

/// Retrieve the cursor string from last am_get_cursor call
#[no_mangle]
pub extern "C" fn am_get_cursor_str(ptr_out: *mut u8) -> i32 {
    LAST_CURSOR.with(|c| {
        let cursor = c.borrow();
        let bytes = cursor.as_bytes();
        unsafe {
            std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
        }
        0
    })
}

/// Lookup the current index for a cursor
///
/// # Arguments
/// * `obj_ptr` - Pointer to object path string
/// * `obj_len` - Length of object path string
/// * `cursor_ptr` - Pointer to cursor string
/// * `cursor_len` - Length of cursor string
///
/// # Returns
/// * >= 0: Current index for the cursor
/// * -1: Invalid path
/// * -2: Invalid cursor
/// * -3: Cursor not found in object
#[no_mangle]
pub extern "C" fn am_lookup_cursor(
    obj_ptr: *const u8,
    obj_len: usize,
    cursor_ptr: *const u8,
    cursor_len: usize,
) -> i32 {
    let path_slice = unsafe { std::slice::from_raw_parts(obj_ptr, obj_len) };
    let path_str = match str::from_utf8(path_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    let cursor_slice = unsafe { std::slice::from_raw_parts(cursor_ptr, cursor_len) };
    let cursor_str = match str::from_utf8(cursor_slice) {
        Ok(s) => s,
        Err(_) => return -2,
    };

    with_doc(|doc| {
        // Parse path to get object ID
        let obj_id = match parse_path(doc, path_str) {
            Ok(id) => id,
            Err(_) => return -1,
        };

        // Parse cursor string using TryFrom
        // Cursor string format: "s" (start), "e" (end), or "123@actorId" (op cursor)
        use automerge::Cursor;
        let cursor = match Cursor::try_from(cursor_str) {
            Ok(c) => c,
            Err(_) => return -2,
        };

        // Get cursor position (None means current heads)
        match doc.get_cursor_position(&obj_id, &cursor, None) {
            Ok(index) => index as i32,
            Err(_) => -3,
        }
    }).unwrap_or(-1)
}

// Thread-local storage for cursor string
use std::cell::RefCell;
thread_local! {
    static LAST_CURSOR: RefCell<String> = RefCell::new(String::new());
}

// Helper to parse path string to ObjId
fn parse_path(doc: &automerge::AutoCommit, path: &str) -> Result<ObjId, ()> {
    if path == "ROOT" {
        return Ok(automerge::ROOT);
    }

    let parts: Vec<&str> = path.split('.').collect();
    if parts.is_empty() || parts[0] != "ROOT" {
        return Err(());
    }

    let mut current = automerge::ROOT;
    for part in &parts[1..] {
        // Try to get object at this key
        match doc.get(&current, *part) {
            Ok(Some((automerge::Value::Object(_obj_type), obj_id))) => {
                current = obj_id;
            }
            _ => return Err(()),
        }
    }

    Ok(current)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cursor_basic() {
        // Initialize document
        crate::document::am_init();

        // Add some text
        let text = b"Hello World";
        let ptr = crate::memory::am_alloc(text.len());
        unsafe {
            std::ptr::copy_nonoverlapping(text.as_ptr(), ptr, text.len());
        }
        crate::text::am_text_splice(0, 0, ptr, text.len());
        crate::memory::am_free(ptr, text.len());

        // Get cursor at position 5 (before "World")
        let path = b"ROOT.content";
        let cursor_len = am_get_cursor(path.as_ptr(), path.len(), 5);
        assert!(cursor_len > 0, "Failed to get cursor");

        // Retrieve cursor string
        let cursor_ptr = crate::memory::am_alloc(cursor_len as usize);
        let result = am_get_cursor_str(cursor_ptr);
        assert_eq!(result, 0);

        // Lookup cursor - should still be at index 5
        let index = am_lookup_cursor(path.as_ptr(), path.len(), cursor_ptr, cursor_len as usize);
        assert_eq!(index, 5);

        crate::memory::am_free(cursor_ptr, cursor_len as usize);
    }

    #[test]
    fn test_cursor_survives_edits() {
        // Initialize document
        crate::document::am_init();

        // Add text "Hello World"
        let text1 = b"Hello World";
        let ptr1 = crate::memory::am_alloc(text1.len());
        unsafe {
            std::ptr::copy_nonoverlapping(text1.as_ptr(), ptr1, text1.len());
        }
        crate::text::am_text_splice(0, 0, ptr1, text1.len());
        crate::memory::am_free(ptr1, text1.len());

        // Get cursor at position 6 (at "World")
        let path = b"ROOT.content";
        let cursor_len = am_get_cursor(path.as_ptr(), path.len(), 6);
        assert!(cursor_len > 0);

        let cursor_ptr = crate::memory::am_alloc(cursor_len as usize);
        am_get_cursor_str(cursor_ptr);

        // Insert text at beginning "Hi "
        let text2 = b"Hi ";
        let ptr2 = crate::memory::am_alloc(text2.len());
        unsafe {
            std::ptr::copy_nonoverlapping(text2.as_ptr(), ptr2, text2.len());
        }
        crate::text::am_text_splice(0, 0, ptr2, text2.len());
        crate::memory::am_free(ptr2, text2.len());

        // Cursor should now point to index 9 (6 + 3 chars inserted)
        let index = am_lookup_cursor(path.as_ptr(), path.len(), cursor_ptr, cursor_len as usize);
        assert_eq!(index, 9, "Cursor should track position after edits");

        crate::memory::am_free(cursor_ptr, cursor_len as usize);
    }

    #[test]
    fn test_cursor_invalid_path() {
        crate::document::am_init();

        let path = b"INVALID.path";
        let result = am_get_cursor(path.as_ptr(), path.len(), 0);
        assert_eq!(result, -1, "Should return error for invalid path");
    }

    #[test]
    fn test_cursor_invalid_index() {
        crate::document::am_init();

        let path = b"ROOT.content";
        // Try to get cursor at invalid index (no text added yet)
        let result = am_get_cursor(path.as_ptr(), path.len(), 999);
        assert!(result < 0, "Should return error for invalid index");
    }
}
