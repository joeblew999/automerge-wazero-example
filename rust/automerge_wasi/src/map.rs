// ==============================================================================
// Layer 2: Rust WASI Exports - Map CRDT
// ==============================================================================
// ARCHITECTURE: This is the WASI export layer (Layer 2/7).
//
// RESPONSIBILITIES:
// - WASI-compatible function exports (C ABI)
// - Memory management (reading/writing linear memory)
// - UTF-8 string marshaling
// - Error code translation (Rust Result → i32)
//
// DEPENDENCIES:
// - Layer 1: automerge crate (CRDT core)
// - crate::state (global document state)
//
// DEPENDENTS:
// - Layer 3: pkg/wazero/crdt_map.go (FFI wrappers)
//
// RELATED FILES (1:1 mapping):
// - Layer 3: pkg/wazero/crdt_map.go (Go FFI wrappers)
// - Layer 4: pkg/automerge/crdt_map.go (Go high-level API)
// - Layer 5: pkg/server/crdt_map.go (stateful server)
// - Layer 6: pkg/api/crdt_map.go (HTTP handlers)
// - Layer 7: web/js/crdt_map.js + web/components/crdt_map.html (TODO)
//
// NOTES:
// - All exports use #[no_mangle] and extern "C"
// - Maps are like JSON objects (string keys → values)
// - Return 0 on success, negative error codes on failure
// ==============================================================================

// WASI exports for Automerge map operations
//
// This module provides C-compatible exports for working with Automerge maps.
//
// ## Map API
//
// Maps in Automerge are like JSON objects - key-value stores where keys are strings.
//
// Example workflow:
// 1. am_init() - Create document
// 2. am_map_set(ROOT, "name", "Alice") - Set key "name" to "Alice"
// 3. am_map_get(ROOT, "name") - Get value "Alice"
// 4. am_map_keys(ROOT) - Get all keys: ["name"]
// 5. am_map_delete(ROOT, "name") - Delete key

use crate::state::with_doc_mut;
use automerge::{transaction::Transactable, ReadDoc, ROOT};

/// Set a string value in a map.
///
/// # Parameters
/// - `key_ptr`: Pointer to key string (UTF-8)
/// - `key_len`: Length of key in bytes
/// - `value_ptr`: Pointer to value string (UTF-8)
/// - `value_len`: Length of value in bytes
///
/// # Returns
/// - `0` on success
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error (e.g., object not a map)
///
/// # Example
/// ```rust
/// let key = "name";
/// let value = "Alice";
/// let result = am_map_set(key.as_ptr(), key.len(), value.as_ptr(), value.len());
/// assert_eq!(result, 0);
/// ```
#[no_mangle]
pub extern "C" fn am_map_set(
    key_ptr: *const u8,
    key_len: usize,
    value_ptr: *const u8,
    value_len: usize,
) -> i32 {
    // Safety: Validate pointers and lengths
    if key_ptr.is_null() || value_ptr.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let value_slice = unsafe { std::slice::from_raw_parts(value_ptr, value_len) };

    // Validate UTF-8
    let key = match std::str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };
    let value = match std::str::from_utf8(value_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    // Put value in ROOT map
    match with_doc_mut(|doc| doc.put(&ROOT, key, value)) {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -2,
        None => -3, // Document not initialized
    }
}

/// Get a string value from a map.
///
/// Call am_map_get_len() first to determine buffer size.
///
/// # Parameters
/// - `key_ptr`: Pointer to key string (UTF-8)
/// - `key_len`: Length of key in bytes
/// - `ptr_out`: Pointer to buffer to receive value
///
/// # Returns
/// - `0` on success (value written to ptr_out)
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error
/// - `-3` if key not found
/// - `-4` if value is not a string
#[no_mangle]
pub extern "C" fn am_map_get(key_ptr: *const u8, key_len: usize, ptr_out: *mut u8) -> i32 {
    if key_ptr.is_null() || ptr_out.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key = match std::str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    // Get value from document - extract string inside closure to avoid lifetime issues
    let result = crate::state::with_doc(|doc| {
        match doc.get(&ROOT, key) {
            Ok(Some((value, _exid))) => {
                // Check if value is a string and extract it
                if let automerge::Value::Scalar(s) = value {
                    if let automerge::ScalarValue::Str(text) = s.as_ref() {
                        return Ok(Some(text.to_string()));
                    }
                }
                Err(-4) // Not a string
            }
            Ok(None) => Err(-3), // Key not found
            Err(_) => Err(-2),   // Automerge error
        }
    });

    match result {
        Some(Ok(Some(text))) => {
            let bytes = text.as_bytes();
            unsafe {
                std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
            }
            0
        }
        Some(Ok(None)) => -4, // Not a string (shouldn't happen with our implementation)
        Some(Err(code)) => code,
        None => -5, // Document not initialized
    }
}

/// Get the length of a string value in the map.
///
/// Use this to allocate a buffer before calling am_map_get().
///
/// # Parameters
/// - `key_ptr`: Pointer to key string (UTF-8)
/// - `key_len`: Length of key in bytes
///
/// # Returns
/// - Length in bytes (>= 0) if key exists and value is a string
/// - `0` if key not found or value is not a string
#[no_mangle]
pub extern "C" fn am_map_get_len(key_ptr: *const u8, key_len: usize) -> u32 {
    if key_ptr.is_null() {
        return 0;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key = match std::str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return 0,
    };

    let result = crate::state::with_doc(|doc| {
        match doc.get(&ROOT, key) {
            Ok(Some((value, _exid))) => {
                if let automerge::Value::Scalar(s) = value {
                    if let automerge::ScalarValue::Str(text) = s.as_ref() {
                        return Some(text.len());
                    }
                }
                None
            }
            _ => None,
        }
    });

    match result {
        Some(Some(len)) => len as u32,
        _ => 0,
    }
}

/// Delete a key from a map.
///
/// # Parameters
/// - `key_ptr`: Pointer to key string (UTF-8)
/// - `key_len`: Length of key in bytes
///
/// # Returns
/// - `0` on success (key deleted, or key didn't exist)
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error
#[no_mangle]
pub extern "C" fn am_map_delete(key_ptr: *const u8, key_len: usize) -> i32 {
    if key_ptr.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key = match std::str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    match with_doc_mut(|doc| doc.delete(&ROOT, key)) {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -2,
        None => -3, // Document not initialized
    }
}

/// Get the number of keys in the map.
///
/// # Returns
/// - Number of keys in the ROOT map
#[no_mangle]
pub extern "C" fn am_map_len() -> u32 {
    crate::state::with_doc(|doc| doc.length(&ROOT))
        .unwrap_or(0) as u32
}

/// Get all keys in the map.
///
/// Call am_map_len() first to determine how many keys exist.
/// Then call am_map_keys_total_size() to determine buffer size.
///
/// # Format
/// Keys are returned as null-terminated strings concatenated together:
/// "key1\0key2\0key3\0"
///
/// # Parameters
/// - `ptr_out`: Pointer to buffer to receive concatenated keys
///
/// # Returns
/// - `0` on success
/// - `-1` on error
#[no_mangle]
pub extern "C" fn am_map_keys(ptr_out: *mut u8) -> i32 {
    if ptr_out.is_null() {
        return -1;
    }

    let keys_vec: Vec<String> = crate::state::with_doc(|doc| {
        doc.keys(&ROOT)
            .map(|k| k.to_string())
            .collect()
    }).unwrap_or_else(Vec::new);

    // Concatenate keys with null terminators
    let mut buffer = Vec::new();
    for key in keys_vec {
        buffer.extend_from_slice(key.as_bytes());
        buffer.push(0); // Null terminator
    }

    unsafe {
        std::ptr::copy_nonoverlapping(buffer.as_ptr(), ptr_out, buffer.len());
    }

    0
}

/// Get the total size needed to store all keys (including null terminators).
///
/// Use this to allocate a buffer before calling am_map_keys().
///
/// # Returns
/// - Total size in bytes needed for all keys
#[no_mangle]
pub extern "C" fn am_map_keys_total_size() -> u32 {
    let keys_vec: Vec<String> = crate::state::with_doc(|doc| {
        doc.keys(&ROOT)
            .map(|k| k.to_string())
            .collect()
    }).unwrap_or_else(Vec::new);

    keys_vec.iter()
        .map(|k| k.len() + 1) // +1 for null terminator
        .sum::<usize>() as u32
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::document::am_init;
    use crate::memory::{am_alloc, am_free};

    #[test]
    fn test_map_set_get() {
        // Initialize document
        assert_eq!(am_init(), 0);

        // Set a key
        let key = "name";
        let value = "Alice";
        let result = am_map_set(
            key.as_ptr(),
            key.len(),
            value.as_ptr(),
            value.len(),
        );
        assert_eq!(result, 0);

        // Get the value length
        let len = am_map_get_len(key.as_ptr(), key.len());
        assert_eq!(len, value.len() as u32);

        // Get the value
        let buf = am_alloc(len as usize);
        assert!(!buf.is_null());
        let result = am_map_get(key.as_ptr(), key.len(), buf);
        assert_eq!(result, 0);

        let retrieved = unsafe {
            std::slice::from_raw_parts(buf, len as usize)
        };
        assert_eq!(retrieved, value.as_bytes());

        am_free(buf, len as usize);
    }

    #[test]
    fn test_map_delete() {
        assert_eq!(am_init(), 0);

        let key = "foo";
        let value = "bar";
        am_map_set(key.as_ptr(), key.len(), value.as_ptr(), value.len());

        // Verify it exists (2 keys: "content" from am_init + "foo")
        assert_eq!(am_map_len(), 2);

        // Delete it
        let result = am_map_delete(key.as_ptr(), key.len());
        assert_eq!(result, 0);

        // Verify it's gone (back to 1 key: "content")
        assert_eq!(am_map_len(), 1);
    }

    #[test]
    fn test_map_keys() {
        assert_eq!(am_init(), 0);

        // Add multiple keys
        am_map_set("a".as_ptr(), 1, "1".as_ptr(), 1);
        am_map_set("b".as_ptr(), 1, "2".as_ptr(), 1);
        am_map_set("c".as_ptr(), 1, "3".as_ptr(), 1);

        // 4 keys: "content" from am_init + "a", "b", "c"
        assert_eq!(am_map_len(), 4);

        // Get total size
        let size = am_map_keys_total_size();
        assert!(size >= 6); // "a\0b\0c\0" = 6 bytes minimum

        // Get keys
        let buf = am_alloc(size as usize);
        assert_eq!(am_map_keys(buf), 0);

        let keys_bytes = unsafe {
            std::slice::from_raw_parts(buf, size as usize)
        };

        // Should contain all keys with null terminators (including "content")
        let keys_str = std::str::from_utf8(keys_bytes).unwrap_or("");
        assert!(keys_str.contains("a"));
        assert!(keys_str.contains("b"));
        assert!(keys_str.contains("c"));
        assert!(keys_str.contains("content")); // From am_init()

        am_free(buf, size as usize);
    }
}
