// ==============================================================================
// Layer 2: Rust WASI Exports - Counter CRDT
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
// - Layer 3: pkg/wazero/crdt_counter.go (FFI wrappers)
//
// RELATED FILES (1:1 mapping):
// - Layer 3: pkg/wazero/crdt_counter.go (Go FFI wrappers)
// - Layer 4: pkg/automerge/crdt_counter.go (Go high-level API)
// - Layer 5: pkg/server/crdt_counter.go (stateful server)
// - Layer 6: pkg/api/crdt_counter.go (HTTP handlers)
// - Layer 7: web/js/crdt_counter.js + web/components/crdt_counter.html (TODO)
//
// NOTES:
// - All exports use #[no_mangle] and extern "C"
// - Counters support concurrent increment/decrement (automatic merge)
// - Return 0 on success, negative error codes on failure
// ==============================================================================

// WASI exports for Automerge counter operations
//
// Counters are CRDT integers that support concurrent increments/decrements
// and automatically merge changes from multiple peers.

use crate::state::{with_doc, with_doc_mut};
use automerge::{transaction::Transactable, ReadDoc, ScalarValue, ROOT};

/// Create a new counter at a key in ROOT map, initialized to value.
///
/// # Parameters
/// - `key_ptr`: Pointer to key string (UTF-8)
/// - `key_len`: Length of key in bytes
/// - `value`: Initial counter value
///
/// # Returns
/// - `0` on success
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_counter_create(
    key_ptr: *const u8,
    key_len: usize,
    value: i64,
) -> i32 {
    if key_ptr.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key = match std::str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    let result = with_doc_mut(|doc| {
        doc.put(&ROOT, key, ScalarValue::counter(value))
    });

    match result {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -2,
        None => -3,
    }
}

/// Increment a counter at a key in ROOT map.
///
/// # Parameters
/// - `key_ptr`: Pointer to key string (UTF-8)
/// - `key_len`: Length of key in bytes
/// - `delta`: Amount to increment (can be negative to decrement)
///
/// # Returns
/// - `0` on success
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error (e.g., key not found or not a counter)
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_counter_increment(
    key_ptr: *const u8,
    key_len: usize,
    delta: i64,
) -> i32 {
    if key_ptr.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key = match std::str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    let result = with_doc_mut(|doc| {
        doc.increment(&ROOT, key, delta)
    });

    match result {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -2,
        None => -3,
    }
}

/// Get the value of a counter at a key in ROOT map.
///
/// # Parameters
/// - `key_ptr`: Pointer to key string (UTF-8)
/// - `key_len`: Length of key in bytes
/// - `value_out`: Pointer to receive counter value
///
/// # Returns
/// - `0` on success (value written to value_out)
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error (e.g., key not found)
/// - `-3` if document not initialized
/// - `-4` if value is not a counter
#[no_mangle]
pub extern "C" fn am_counter_get(
    key_ptr: *const u8,
    key_len: usize,
    value_out: *mut i64,
) -> i32 {
    if key_ptr.is_null() || value_out.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key = match std::str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    let result = with_doc(|doc| {
        match doc.get(&ROOT, key) {
            Ok(Some((value, _))) => {
                if let automerge::Value::Scalar(s) = value {
                    if let ScalarValue::Counter(c) = s.as_ref() {
                        return Ok(c.into());
                    }
                }
                Err(-4)
            }
            Ok(None) => Err(-2),
            Err(_) => Err(-2),
        }
    });

    match result {
        Some(Ok(val)) => {
            unsafe { *value_out = val };
            0
        }
        Some(Err(code)) => code,
        None => -3,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::document::am_init;

    #[test]
    fn test_counter_create_get() {
        assert_eq!(am_init(), 0);

        let key = "score";
        let result = am_counter_create(key.as_ptr(), key.len(), 100);
        assert_eq!(result, 0);

        let mut value: i64 = 0;
        let result = am_counter_get(key.as_ptr(), key.len(), &mut value);
        assert_eq!(result, 0);
        assert_eq!(value, 100);
    }

    #[test]
    fn test_counter_increment() {
        assert_eq!(am_init(), 0);

        let key = "count";
        am_counter_create(key.as_ptr(), key.len(), 0);

        // Increment by 5
        let result = am_counter_increment(key.as_ptr(), key.len(), 5);
        assert_eq!(result, 0);

        let mut value: i64 = 0;
        am_counter_get(key.as_ptr(), key.len(), &mut value);
        assert_eq!(value, 5);

        // Increment by 3 more
        am_counter_increment(key.as_ptr(), key.len(), 3);
        am_counter_get(key.as_ptr(), key.len(), &mut value);
        assert_eq!(value, 8);
    }

    #[test]
    fn test_counter_decrement() {
        assert_eq!(am_init(), 0);

        let key = "balance";
        am_counter_create(key.as_ptr(), key.len(), 100);

        // Decrement by 30 (negative increment)
        let result = am_counter_increment(key.as_ptr(), key.len(), -30);
        assert_eq!(result, 0);

        let mut value: i64 = 0;
        am_counter_get(key.as_ptr(), key.len(), &mut value);
        assert_eq!(value, 70);
    }
}
