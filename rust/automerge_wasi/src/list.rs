// WASI exports for Automerge list operations
//
// Lists in Automerge are CRDT ordered sequences that support concurrent insertions
// and deletions while maintaining eventual consistency.
//
// ## List API
//
// Example workflow:
// 1. am_list_create(ROOT, "items") - Create list at ROOT["items"]
// 2. am_list_push(list_id, "item1") - Append to end
// 3. am_list_insert(list_id, 0, "item0") - Insert at beginning
// 4. am_list_get(list_id, 1) - Get item at index 1
// 5. am_list_delete(list_id, 0) - Delete item at index 0
// 6. am_list_len(list_id) - Get length

use crate::state::{with_doc, with_doc_mut};
use automerge::{transaction::Transactable, ObjType, ReadDoc, ROOT};

/// Create a new List object at a key in ROOT map.
///
/// # Parameters
/// - `key_ptr`: Pointer to key string (UTF-8)
/// - `key_len`: Length of key in bytes
/// - `obj_id_out`: Pointer to buffer to receive object ID (as string)
///
/// # Returns
/// - `0` on success (object ID written to obj_id_out)
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_list_create(
    key_ptr: *const u8,
    key_len: usize,
    obj_id_out: *mut u8,
) -> i32 {
    if key_ptr.is_null() || obj_id_out.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key = match std::str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    // Create list object
    let result = with_doc_mut(|doc| {
        doc.put_object(&ROOT, key, ObjType::List)
    });

    match result {
        Some(Ok(obj_id)) => {
            // Convert ObjId to string
            let id_str = obj_id.to_string();
            let id_bytes = id_str.as_bytes();

            unsafe {
                std::ptr::copy_nonoverlapping(id_bytes.as_ptr(), obj_id_out, id_bytes.len());
            }
            0
        }
        Some(Err(_)) => -2,
        None => -3,
    }
}

/// Get the length of an object ID string for a created list.
///
/// Call this before am_list_create to allocate the buffer.
///
/// # Returns
/// - Length in bytes (typically 32-64 bytes for ExId)
#[no_mangle]
pub extern "C" fn am_list_obj_id_len() -> u32 {
    // ExId string representation is typically around 32-64 bytes
    // We'll use a safe upper bound
    128
}

/// Push a string value to the end of a list.
///
/// Note: Current implementation uses a simplified approach with ROOT["list_items"].
/// Full implementation would use proper object IDs.
///
/// # Parameters
/// - `value_ptr`: Pointer to value string (UTF-8)
/// - `value_len`: Length of value in bytes
///
/// # Returns
/// - `0` on success
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_list_push(value_ptr: *const u8, value_len: usize) -> i32 {
    if value_ptr.is_null() {
        return -1;
    }

    let value_slice = unsafe { std::slice::from_raw_parts(value_ptr, value_len) };
    let value = match std::str::from_utf8(value_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    // For now, use a global list at ROOT["list_items"]
    // Full implementation would track object IDs
    let result = with_doc_mut(|doc| {
        // Get or create list
        let list_id = match doc.get(&ROOT, "list_items") {
            Ok(Some((val, id))) => {
                // Check if it's an object (list)
                if val.is_object() {
                    id
                } else {
                    // Replace with a list
                    doc.put_object(&ROOT, "list_items", ObjType::List)?
                }
            }
            _ => {
                // Create new list
                doc.put_object(&ROOT, "list_items", ObjType::List)?
            }
        };

        // Get current length
        let len = doc.length(&list_id);

        // Insert at end
        doc.insert(&list_id, len, value)
    });

    match result {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -2,
        None => -3,
    }
}

/// Insert a string value at a specific index in a list.
///
/// # Parameters
/// - `index`: Index to insert at (0-based)
/// - `value_ptr`: Pointer to value string (UTF-8)
/// - `value_len`: Length of value in bytes
///
/// # Returns
/// - `0` on success
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error (e.g., index out of bounds)
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_list_insert(
    index: usize,
    value_ptr: *const u8,
    value_len: usize,
) -> i32 {
    if value_ptr.is_null() {
        return -1;
    }

    let value_slice = unsafe { std::slice::from_raw_parts(value_ptr, value_len) };
    let value = match std::str::from_utf8(value_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    let result = with_doc_mut(|doc| {
        // Get list
        let list_id = match doc.get(&ROOT, "list_items") {
            Ok(Some((_, id))) => id,
            _ => return Err(automerge::AutomergeError::InvalidOp(ObjType::Map)),
        };

        doc.insert(&list_id, index, value)
    });

    match result {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -2,
        None => -3,
    }
}

/// Get a string value from a list at a specific index.
///
/// Call am_list_get_len() first to determine buffer size.
///
/// # Parameters
/// - `index`: Index to get (0-based)
/// - `value_out`: Pointer to buffer to receive value
///
/// # Returns
/// - `0` on success
/// - `-1` on null pointer
/// - `-2` on Automerge error (e.g., index out of bounds)
/// - `-3` if document not initialized
/// - `-4` if value is not a string
#[no_mangle]
pub extern "C" fn am_list_get(index: usize, value_out: *mut u8) -> i32 {
    if value_out.is_null() {
        return -1;
    }

    let result = with_doc(|doc| {
        // Get list
        let list_id = match doc.get(&ROOT, "list_items") {
            Ok(Some((_, id))) => id,
            _ => return Err(-2),
        };

        // Get value at index
        match doc.get(&list_id, index) {
            Ok(Some((value, _))) => {
                if let automerge::Value::Scalar(s) = value {
                    if let automerge::ScalarValue::Str(text) = s.as_ref() {
                        return Ok(text.to_string());
                    }
                }
                Err(-4)
            }
            Ok(None) => Err(-2), // Index out of bounds
            Err(_) => Err(-2),
        }
    });

    match result {
        Some(Ok(text)) => {
            let bytes = text.as_bytes();
            unsafe {
                std::ptr::copy_nonoverlapping(bytes.as_ptr(), value_out, bytes.len());
            }
            0
        }
        Some(Err(code)) => code,
        None => -3,
    }
}

/// Get the length of a string value at a specific index.
///
/// # Parameters
/// - `index`: Index to check (0-based)
///
/// # Returns
/// - Length in bytes (>= 0) if value exists and is a string
/// - `0` if index out of bounds or value is not a string
#[no_mangle]
pub extern "C" fn am_list_get_len(index: usize) -> u32 {
    let result = with_doc(|doc| {
        let list_id = match doc.get(&ROOT, "list_items") {
            Ok(Some((_, id))) => id,
            _ => return 0,
        };

        match doc.get(&list_id, index) {
            Ok(Some((value, _))) => {
                if let automerge::Value::Scalar(s) = value {
                    if let automerge::ScalarValue::Str(text) = s.as_ref() {
                        return text.len();
                    }
                }
                0
            }
            _ => 0,
        }
    });

    result.unwrap_or(0) as u32
}

/// Delete a value from a list at a specific index.
///
/// # Parameters
/// - `index`: Index to delete (0-based)
///
/// # Returns
/// - `0` on success
/// - `-2` on Automerge error (e.g., index out of bounds)
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_list_delete(index: usize) -> i32 {
    let result = with_doc_mut(|doc| {
        let list_id = match doc.get(&ROOT, "list_items") {
            Ok(Some((_, id))) => id,
            _ => return Err(automerge::AutomergeError::InvalidOp(ObjType::Map)),
        };

        doc.delete(&list_id, index)
    });

    match result {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -2,
        None => -3,
    }
}

/// Get the length of a list.
///
/// # Returns
/// - Number of elements in the list
/// - `0` if list doesn't exist or document not initialized
#[no_mangle]
pub extern "C" fn am_list_len() -> u32 {
    let result = with_doc(|doc| {
        let list_id = match doc.get(&ROOT, "list_items") {
            Ok(Some((_, id))) => id,
            _ => return 0,
        };

        doc.length(&list_id)
    });

    result.unwrap_or(0) as u32
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::document::am_init;
    use crate::memory::{am_alloc, am_free};

    #[test]
    fn test_list_push_get() {
        assert_eq!(am_init(), 0);

        // Push values
        let values = ["first", "second", "third"];
        for value in &values {
            let result = am_list_push(value.as_ptr(), value.len());
            assert_eq!(result, 0, "Failed to push {}", value);
        }

        // Verify length
        assert_eq!(am_list_len(), 3);

        // Get and verify values
        for (i, expected) in values.iter().enumerate() {
            let len = am_list_get_len(i);
            assert_eq!(len, expected.len() as u32);

            let buf = am_alloc(len as usize);
            assert!(!buf.is_null());

            let result = am_list_get(i, buf);
            assert_eq!(result, 0);

            let retrieved = unsafe {
                std::slice::from_raw_parts(buf, len as usize)
            };
            assert_eq!(retrieved, expected.as_bytes());

            am_free(buf, len as usize);
        }
    }

    #[test]
    fn test_list_insert() {
        assert_eq!(am_init(), 0);

        // Push initial values
        am_list_push("a".as_ptr(), 1);
        am_list_push("c".as_ptr(), 1);

        // Insert in the middle
        let result = am_list_insert(1, "b".as_ptr(), 1);
        assert_eq!(result, 0);

        // Verify order: [a, b, c]
        assert_eq!(am_list_len(), 3);

        let expected = ["a", "b", "c"];
        for (i, exp) in expected.iter().enumerate() {
            let len = am_list_get_len(i);
            let buf = am_alloc(len as usize);
            am_list_get(i, buf);
            let got = unsafe {
                std::str::from_utf8(std::slice::from_raw_parts(buf, len as usize)).unwrap()
            };
            assert_eq!(got, *exp);
            am_free(buf, len as usize);
        }
    }

    #[test]
    fn test_list_delete() {
        assert_eq!(am_init(), 0);

        // Push values
        am_list_push("a".as_ptr(), 1);
        am_list_push("b".as_ptr(), 1);
        am_list_push("c".as_ptr(), 1);

        assert_eq!(am_list_len(), 3);

        // Delete middle element
        let result = am_list_delete(1);
        assert_eq!(result, 0);

        // Verify length and remaining values
        assert_eq!(am_list_len(), 2);

        // Should have [a, c]
        let len = am_list_get_len(0);
        let buf = am_alloc(len as usize);
        am_list_get(0, buf);
        let got = unsafe {
            std::str::from_utf8(std::slice::from_raw_parts(buf, len as usize)).unwrap()
        };
        assert_eq!(got, "a");
        am_free(buf, len as usize);

        let len = am_list_get_len(1);
        let buf = am_alloc(len as usize);
        am_list_get(1, buf);
        let got = unsafe {
            std::str::from_utf8(std::slice::from_raw_parts(buf, len as usize)).unwrap()
        };
        assert_eq!(got, "c");
        am_free(buf, len as usize);
    }

    #[test]
    fn test_list_empty() {
        assert_eq!(am_init(), 0);

        // Empty list
        assert_eq!(am_list_len(), 0);

        // Getting from empty list should fail
        let buf = am_alloc(10);
        let result = am_list_get(0, buf);
        assert_ne!(result, 0);
        am_free(buf, 10);
    }
}
