//! Memory management for WASM FFI
//!
//! Provides allocation and deallocation functions for transferring data between Go and Rust.
//!
//! ## Safety
//!
//! - All allocations use 8-byte alignment
//! - Caller MUST call `am_free` with the same pointer and size from `am_alloc`
//! - Failing to free memory will cause leaks

use std::alloc::{alloc, dealloc, Layout};

/// Allocate memory in WASM linear memory for Go→Rust data transfer
///
/// ## Parameters
/// - `size`: Number of bytes to allocate
///
/// ## Returns
/// - Pointer to allocated buffer (8-byte aligned)
/// - `null` if allocation fails or size is 0
///
/// ## Safety
/// Caller MUST call `am_free(ptr, size)` when done
#[no_mangle]
pub extern "C" fn am_alloc(size: usize) -> *mut u8 {
    if size == 0 {
        return std::ptr::null_mut();
    }

    let layout = match Layout::from_size_align(size, 8) {
        Ok(l) => l,
        Err(_) => return std::ptr::null_mut(),
    };

    unsafe { alloc(layout) }
}

/// Free memory allocated by `am_alloc`
///
/// ## Parameters
/// - `ptr`: Pointer returned from `am_alloc`
/// - `size`: Same size passed to `am_alloc`
///
/// ## Safety
/// - Must be called with exact same pointer and size from `am_alloc`
/// - Double-free will cause undefined behavior
/// - Freeing null or zero-size is a no-op
#[no_mangle]
pub extern "C" fn am_free(ptr: *mut u8, size: usize) {
    if ptr.is_null() || size == 0 {
        return;
    }

    let layout = match Layout::from_size_align(size, 8) {
        Ok(l) => l,
        Err(_) => return,
    };

    unsafe { dealloc(ptr, layout) }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_alloc_free() {
        let size = 1024;
        let ptr = am_alloc(size);
        assert!(!ptr.is_null());
        am_free(ptr, size);
    }

    #[test]
    fn test_alloc_zero() {
        let ptr = am_alloc(0);
        assert!(ptr.is_null());
    }

    #[test]
    fn test_free_null() {
        // Should not panic
        am_free(std::ptr::null_mut(), 100);
    }
}
