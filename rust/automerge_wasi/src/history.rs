// WASI exports for Automerge history and time-travel operations
//
// History operations allow you to query the document's change history,
// get heads (frontier), and fork documents at specific points in time.

use crate::state::with_doc_mut;

/// Get the number of heads (frontier) in the document.
///
/// Heads identify the current state of the document. After merging, a document
/// may have multiple heads temporarily until the next change.
///
/// # Returns
/// - Number of heads (≥ 1)
/// - `0` if document not initialized
#[no_mangle]
pub extern "C" fn am_get_heads_count() -> u32 {
    let result = with_doc_mut(|doc| {
        doc.get_heads().len()
    });

    result.unwrap_or(0) as u32
}

/// Get the heads (change hashes) of the document.
///
/// Call `am_get_heads_count()` first to allocate the correct buffer size.
/// Each head is a 32-byte hash.
///
/// # Parameters
/// - `heads_out`: Pointer to buffer to receive heads (32 bytes per head)
///
/// # Returns
/// - `0` on success
/// - `-1` if heads_out is null
/// - `-2` if document not initialized
#[no_mangle]
pub extern "C" fn am_get_heads(heads_out: *mut u8) -> i32 {
    if heads_out.is_null() {
        return -1;
    }

    let result = with_doc_mut(|doc| {
        let heads = doc.get_heads();
        let mut offset = 0;

        for head in heads {
            let bytes = &head.0; // ChangeHash wraps a [u8; 32]
            unsafe {
                std::ptr::copy_nonoverlapping(
                    bytes.as_ptr(),
                    heads_out.add(offset),
                    bytes.len(),
                );
            }
            offset += bytes.len();
        }

        0
    });

    match result {
        Some(code) => code,
        None => -2,
    }
}

/// Get the number of changes since the given heads.
///
/// Used internally by the sync protocol. Returns count of changes
/// that have occurred since the specified dependencies.
///
/// # Parameters
/// - `have_heads_ptr`: Pointer to array of change hashes (32 bytes each)
/// - `have_heads_count`: Number of heads in the array
///
/// # Returns
/// - Number of changes since the given heads
/// - `0` if document not initialized or no changes
#[no_mangle]
pub extern "C" fn am_get_changes_count(
    have_heads_ptr: *const u8,
    have_heads_count: usize,
) -> u32 {
    if have_heads_count == 0 {
        // No dependencies = all changes
        let result = with_doc_mut(|doc| {
            doc.get_changes(&[]).len()
        });
        return result.unwrap_or(0) as u32;
    }

    if have_heads_ptr.is_null() {
        return 0;
    }

    let result = with_doc_mut(|doc| {
        // Convert byte array to ChangeHash array
        let mut have_heads = Vec::new();
        for i in 0..have_heads_count {
            let offset = i * 32;
            let hash_slice = unsafe {
                std::slice::from_raw_parts(have_heads_ptr.add(offset), 32)
            };

            // Try to convert to ChangeHash
            if let Ok(hash) = automerge::ChangeHash::try_from(hash_slice) {
                have_heads.push(hash);
            } else {
                return 0; // Invalid hash
            }
        }

        doc.get_changes(&have_heads).len()
    });

    result.unwrap_or(0) as u32
}

/// Get the total size of changes since the given heads.
///
/// Call this before `am_get_changes()` to allocate the correct buffer.
///
/// # Parameters
/// - `have_heads_ptr`: Pointer to array of change hashes (32 bytes each)
/// - `have_heads_count`: Number of heads in the array
///
/// # Returns
/// - Total byte size needed for all changes
/// - `0` if document not initialized or no changes
#[no_mangle]
pub extern "C" fn am_get_changes_len(
    have_heads_ptr: *const u8,
    have_heads_count: usize,
) -> u32 {
    if have_heads_count == 0 {
        // No dependencies = all changes
        let result = with_doc_mut(|doc| {
            let changes = doc.get_changes(&[]);
            changes.iter().map(|c| c.raw_bytes().len()).sum::<usize>()
        });
        return result.unwrap_or(0) as u32;
    }

    if have_heads_ptr.is_null() {
        return 0;
    }

    let result = with_doc_mut(|doc| {
        let mut have_heads = Vec::new();
        for i in 0..have_heads_count {
            let offset = i * 32;
            let hash_slice = unsafe {
                std::slice::from_raw_parts(have_heads_ptr.add(offset), 32)
            };

            if let Ok(hash) = automerge::ChangeHash::try_from(hash_slice) {
                have_heads.push(hash);
            } else {
                return 0;
            }
        }

        let changes = doc.get_changes(&have_heads);
        changes.iter().map(|c| c.raw_bytes().len()).sum::<usize>()
    });

    result.unwrap_or(0) as u32
}

/// Get changes since the given heads.
///
/// Call `am_get_changes_count()` and `am_get_changes_len()` first to
/// allocate correct buffers.
///
/// # Parameters
/// - `have_heads_ptr`: Pointer to array of change hashes (32 bytes each)
/// - `have_heads_count`: Number of heads in the array
/// - `changes_out`: Pointer to buffer to receive serialized changes
///
/// # Returns
/// - `0` on success
/// - `-1` if changes_out is null
/// - `-2` if document not initialized
/// - `-3` if invalid change hashes provided
#[no_mangle]
pub extern "C" fn am_get_changes(
    have_heads_ptr: *const u8,
    have_heads_count: usize,
    changes_out: *mut u8,
) -> i32 {
    if changes_out.is_null() {
        return -1;
    }

    let result = with_doc_mut(|doc| {
        let have_heads = if have_heads_count == 0 {
            Vec::new()
        } else {
            if have_heads_ptr.is_null() {
                return Err(-3);
            }

            let mut heads = Vec::new();
            for i in 0..have_heads_count {
                let offset = i * 32;
                let hash_slice = unsafe {
                    std::slice::from_raw_parts(have_heads_ptr.add(offset), 32)
                };

                match automerge::ChangeHash::try_from(hash_slice) {
                    Ok(hash) => heads.push(hash),
                    Err(_) => return Err(-3),
                }
            }
            heads
        };

        let changes = doc.get_changes(&have_heads);

        // Write changes to output buffer
        let mut offset = 0;
        for change in changes {
            let bytes = change.raw_bytes();
            unsafe {
                std::ptr::copy_nonoverlapping(
                    bytes.as_ptr(),
                    changes_out.add(offset),
                    bytes.len(),
                );
            }
            offset += bytes.len();
        }

        Ok(0)
    });

    match result {
        Some(Ok(code)) => code,
        Some(Err(code)) => code,
        None => -2,
    }
}

/// Apply changes to the document.
///
/// This is the low-level API for applying changes. Typically used by
/// the sync protocol.
///
/// # Parameters
/// - `changes_ptr`: Pointer to serialized changes bytes
/// - `changes_len`: Length of changes buffer
///
/// # Returns
/// - `0` on success
/// - `-1` if changes_ptr is null
/// - `-2` if failed to apply changes
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_apply_changes(changes_ptr: *const u8, changes_len: usize) -> i32 {
    if changes_ptr.is_null() {
        return -1;
    }

    let changes_slice = unsafe { std::slice::from_raw_parts(changes_ptr, changes_len) };

    let result = with_doc_mut(|doc| {
        match doc.load_incremental(changes_slice) {
            Ok(_) => 0,
            Err(_) => -2,
        }
    });

    match result {
        Some(code) => code,
        None => -3,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::document::am_init;
    use crate::text::am_text_splice;

    #[test]
    fn test_get_heads() {
        assert_eq!(am_init(), 0);

        // Initial document should have heads
        let count = am_get_heads_count();
        assert!(count > 0);

        // Allocate buffer and get heads
        let buffer = vec![0u8; (count as usize) * 32];
        let result = am_get_heads(buffer.as_ptr() as *mut u8);
        assert_eq!(result, 0);
    }

    #[test]
    fn test_get_changes() {
        assert_eq!(am_init(), 0);

        // Make a change
        let text = "Hello";
        assert_eq!(am_text_splice(0, 0, text.as_ptr(), text.len()), 0);

        // Get all changes (no dependencies)
        let count = am_get_changes_count(std::ptr::null(), 0);
        assert!(count > 0);

        let len = am_get_changes_len(std::ptr::null(), 0);
        assert!(len > 0);

        let buffer = vec![0u8; len as usize];
        let result = am_get_changes(std::ptr::null(), 0, buffer.as_ptr() as *mut u8);
        assert_eq!(result, 0);
    }

    #[test]
    fn test_get_changes_with_heads() {
        assert_eq!(am_init(), 0);

        // Get initial heads
        let head_count = am_get_heads_count();
        let mut heads_buf = vec![0u8; (head_count as usize) * 32];
        assert_eq!(am_get_heads(heads_buf.as_mut_ptr()), 0);

        // Make a change
        let text = "World";
        assert_eq!(am_text_splice(0, 0, text.as_ptr(), text.len()), 0);

        // Get changes since initial heads
        let count = am_get_changes_count(heads_buf.as_ptr(), head_count as usize);
        assert!(count > 0); // Should have new changes

        let len = am_get_changes_len(heads_buf.as_ptr(), head_count as usize);
        assert!(len > 0);
    }
}
