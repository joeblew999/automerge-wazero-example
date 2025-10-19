use automerge::{AutoCommit, ReadDoc, transaction::Transactable};
use std::cell::RefCell;
use std::alloc::{alloc, dealloc, Layout};

// Global document - in a real application, consider thread-local or passed context
thread_local! {
    static DOC: RefCell<Option<AutoCommit>> = RefCell::new(None);
}

// Memory management exports
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

// Initialize a new Automerge document with a "content" text field
#[no_mangle]
pub extern "C" fn am_init() -> i32 {
    DOC.with(|doc_cell| {
        let mut doc = AutoCommit::new();

        // Create a text object at key "content"
        if let Err(_) = doc.put(automerge::ROOT, "content", "") {
            return -1;
        }

        *doc_cell.borrow_mut() = Some(doc);
        0
    })
}

// Set the entire text content
#[no_mangle]
pub extern "C" fn am_set_text(ptr: *const u8, len: usize) -> i32 {
    if ptr.is_null() {
        return -1;
    }

    let slice = unsafe { std::slice::from_raw_parts(ptr, len) };
    let text = match std::str::from_utf8(slice) {
        Ok(s) => s,
        Err(_) => return -2, // Invalid UTF-8
    };

    DOC.with(|doc_cell| {
        let mut doc_opt = doc_cell.borrow_mut();
        let doc = match doc_opt.as_mut() {
            Some(d) => d,
            None => return -3, // Document not initialized
        };

        // Replace the "content" field
        if let Err(_) = doc.put(automerge::ROOT, "content", text) {
            return -4;
        }

        0
    })
}

// Get the length of the text content
#[no_mangle]
pub extern "C" fn am_get_text_len() -> u32 {
    DOC.with(|doc_cell| {
        let doc_opt = doc_cell.borrow();
        let doc = match doc_opt.as_ref() {
            Some(d) => d,
            None => return 0,
        };

        match doc.get(automerge::ROOT, "content") {
            Ok(Some((value, _))) => {
                if let Some(s) = value.to_str() {
                    s.len() as u32
                } else {
                    0
                }
            }
            _ => 0,
        }
    })
}

// Get the text content (caller must allocate buffer of size from am_get_text_len)
#[no_mangle]
pub extern "C" fn am_get_text(ptr_out: *mut u8) -> i32 {
    if ptr_out.is_null() {
        return -1;
    }

    DOC.with(|doc_cell| {
        let doc_opt = doc_cell.borrow();
        let doc = match doc_opt.as_ref() {
            Some(d) => d,
            None => return -2, // Document not initialized
        };

        match doc.get(automerge::ROOT, "content") {
            Ok(Some((value, _))) => {
                if let Some(s) = value.to_str() {
                    let bytes = s.as_bytes();
                    unsafe {
                        std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
                    }
                    0
                } else {
                    -3 // Not a string
                }
            }
            _ => -4, // Field not found
        }
    })
}

// Get the length of the saved document
#[no_mangle]
pub extern "C" fn am_save_len() -> u32 {
    DOC.with(|doc_cell| {
        let mut doc_opt = doc_cell.borrow_mut();
        let doc = match doc_opt.as_mut() {
            Some(d) => d,
            None => return 0,
        };

        doc.save().len() as u32
    })
}

// Save the document to a buffer (caller must allocate buffer of size from am_save_len)
#[no_mangle]
pub extern "C" fn am_save(ptr_out: *mut u8) -> i32 {
    if ptr_out.is_null() {
        return -1;
    }

    DOC.with(|doc_cell| {
        let mut doc_opt = doc_cell.borrow_mut();
        let doc = match doc_opt.as_mut() {
            Some(d) => d,
            None => return -2, // Document not initialized
        };

        let bytes = doc.save();
        unsafe {
            std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
        }
        0
    })
}

// Load a document from a buffer
#[no_mangle]
pub extern "C" fn am_load(ptr: *const u8, len: usize) -> i32 {
    if ptr.is_null() {
        return -1;
    }

    let slice = unsafe { std::slice::from_raw_parts(ptr, len) };

    DOC.with(|doc_cell| {
        match AutoCommit::load(slice) {
            Ok(doc) => {
                *doc_cell.borrow_mut() = Some(doc);
                0
            }
            Err(_) => -2, // Failed to load
        }
    })
}
