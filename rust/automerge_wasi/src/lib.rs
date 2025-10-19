use automerge::{AutoCommit, ReadDoc, transaction::Transactable, ObjType, ObjId};
use std::cell::RefCell;
use std::alloc::{alloc, dealloc, Layout};

// Global document and text object ID
// In a real application, consider thread-local or passed context
thread_local! {
    static DOC: RefCell<Option<AutoCommit>> = RefCell::new(None);
    static TEXT_OBJ_ID: RefCell<Option<ObjId>> = RefCell::new(None);
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

// Initialize a new Automerge document with a "content" Text CRDT object
#[no_mangle]
pub extern "C" fn am_init() -> i32 {
    DOC.with(|doc_cell| {
        TEXT_OBJ_ID.with(|text_id_cell| {
            let mut doc = AutoCommit::new();

            // Create a Text object (CRDT) at key "content"
            let text_obj_id = match doc.put_object(automerge::ROOT, "content", ObjType::Text) {
                Ok(id) => id,
                Err(_) => return -1,
            };

            // Store the text object ID for later operations
            *text_id_cell.borrow_mut() = Some(text_obj_id);
            *doc_cell.borrow_mut() = Some(doc);
            0
        })
    })
}

// Splice text at a given position (proper Text CRDT operation)
// pos: byte position to start
// del_count: number of UTF-8 characters to delete
// insert_ptr: pointer to string to insert
// insert_len: length of string to insert
#[no_mangle]
pub extern "C" fn am_text_splice(pos: usize, del_count: i64, insert_ptr: *const u8, insert_len: usize) -> i32 {
    let insert_text = if insert_len > 0 && !insert_ptr.is_null() {
        let slice = unsafe { std::slice::from_raw_parts(insert_ptr, insert_len) };
        match std::str::from_utf8(slice) {
            Ok(s) => s,
            Err(_) => return -2, // Invalid UTF-8
        }
    } else {
        ""
    };

    DOC.with(|doc_cell| {
        TEXT_OBJ_ID.with(|text_id_cell| {
            let mut doc_opt = doc_cell.borrow_mut();
            let text_id_opt = text_id_cell.borrow();

            let doc = match doc_opt.as_mut() {
                Some(d) => d,
                None => return -3, // Document not initialized
            };

            let text_id = match text_id_opt.as_ref() {
                Some(id) => id,
                None => return -4, // Text object not initialized
            };

            // Convert i64 to isize for delete count
            let del_count_isize = match del_count.try_into() {
                Ok(n) => n,
                Err(_) => return -6, // Invalid delete count
            };

            // Perform splice operation on Text CRDT
            if let Err(_) = doc.splice_text(text_id, pos, del_count_isize, insert_text) {
                return -5;
            }

            0
        })
    })
}

// DEPRECATED: Set the entire text content (for backward compatibility)
// Use am_text_splice for proper Text CRDT operations
#[no_mangle]
pub extern "C" fn am_set_text(ptr: *const u8, len: usize) -> i32 {
    if ptr.is_null() {
        return -1;
    }

    // Get current text length to delete all
    let current_len = am_get_text_len() as usize;

    // Delete all existing text, then insert new text
    // This is inefficient but maintains backward compatibility
    if current_len > 0 {
        // Delete all existing characters
        if am_text_splice(0, current_len as i64, std::ptr::null(), 0) != 0 {
            return -2;
        }
    }

    // Insert new text at position 0
    am_text_splice(0, 0, ptr, len)
}

// Get the length of the text content
#[no_mangle]
pub extern "C" fn am_get_text_len() -> u32 {
    DOC.with(|doc_cell| {
        TEXT_OBJ_ID.with(|text_id_cell| {
            let doc_opt = doc_cell.borrow();
            let text_id_opt = text_id_cell.borrow();

            let doc = match doc_opt.as_ref() {
                Some(d) => d,
                None => return 0,
            };

            let text_id = match text_id_opt.as_ref() {
                Some(id) => id,
                None => return 0,
            };

            // Get text from Text object
            match doc.text(text_id) {
                Ok(s) => s.len() as u32,
                Err(_) => 0,
            }
        })
    })
}

// Get the text content (caller must allocate buffer of size from am_get_text_len)
#[no_mangle]
pub extern "C" fn am_get_text(ptr_out: *mut u8) -> i32 {
    if ptr_out.is_null() {
        return -1;
    }

    DOC.with(|doc_cell| {
        TEXT_OBJ_ID.with(|text_id_cell| {
            let doc_opt = doc_cell.borrow();
            let text_id_opt = text_id_cell.borrow();

            let doc = match doc_opt.as_ref() {
                Some(d) => d,
                None => return -2, // Document not initialized
            };

            let text_id = match text_id_opt.as_ref() {
                Some(id) => id,
                None => return -3, // Text object not initialized
            };

            // Get text from Text object
            match doc.text(text_id) {
                Ok(s) => {
                    let bytes = s.as_bytes();
                    unsafe {
                        std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
                    }
                    0
                }
                Err(_) => -4, // Failed to get text
            }
        })
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
        TEXT_OBJ_ID.with(|text_id_cell| {
            match AutoCommit::load(slice) {
                Ok(doc) => {
                    // After loading, find the text object ID
                    // Assuming the text field is at key "content"
                    match doc.get(automerge::ROOT, "content") {
                        Ok(Some((_, obj_id))) => {
                            *text_id_cell.borrow_mut() = Some(obj_id);
                        }
                        _ => {
                            // Text object not found, this is an error
                            return -3;
                        }
                    }

                    *doc_cell.borrow_mut() = Some(doc);
                    0
                }
                Err(_) => -2, // Failed to load
            }
        })
    })
}

// Merge another document into the current document
// This is the CRDT magic! Two diverged documents can be merged without conflicts
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

    DOC.with(|doc_cell| {
        TEXT_OBJ_ID.with(|text_id_cell| {
            let mut doc_opt = doc_cell.borrow_mut();
            let doc = match doc_opt.as_mut() {
                Some(d) => d,
                None => return -3, // Current document not initialized
            };

            // Perform CRDT merge!
            doc.merge(&mut other_doc.fork()).expect("Merge failed");

            // After merge, update text_id in case it changed
            match doc.get(automerge::ROOT, "content") {
                Ok(Some((_, obj_id))) => {
                    *text_id_cell.borrow_mut() = Some(obj_id);
                }
                _ => {
                    return -4; // Text object not found after merge
                }
            }

            0
        })
    })
}
