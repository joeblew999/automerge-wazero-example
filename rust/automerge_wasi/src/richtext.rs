// WASI exports for Automerge rich text operations (Marks and Spans)
//
// Marks allow you to add formatting (bold, italic, links, etc.) to text ranges.
// Marks are CRDT-aware and merge correctly when users concurrently format
// the same text.

use crate::state::{with_doc, with_doc_mut, get_text_obj_id};
use automerge::{marks::{ExpandMark, Mark}, transaction::Transactable, ReadDoc, ScalarValue};

/// Add a mark (formatting) to a range of text.
///
/// # Parameters
/// - `name_ptr`: Pointer to mark name string (UTF-8) (e.g., "bold", "italic")
/// - `name_len`: Length of name in bytes
/// - `value_ptr`: Pointer to mark value string (UTF-8) (e.g., "true", "https://...")
/// - `value_len`: Length of value in bytes
/// - `start`: Start index of the range
/// - `end`: End index of the range (exclusive)
/// - `expand`: Expand mode (0=none, 1=before, 2=after, 3=both)
///
/// # Returns
/// - `0` on success
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_mark(
    name_ptr: *const u8,
    name_len: usize,
    value_ptr: *const u8,
    value_len: usize,
    start: usize,
    end: usize,
    expand: u8,
) -> i32 {
    if name_ptr.is_null() || value_ptr.is_null() {
        return -1;
    }

    let name_slice = unsafe { std::slice::from_raw_parts(name_ptr, name_len) };
    let name = match std::str::from_utf8(name_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    let value_slice = unsafe { std::slice::from_raw_parts(value_ptr, value_len) };
    let value = match std::str::from_utf8(value_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    let expand_mode = match expand {
        0 => ExpandMark::None,
        1 => ExpandMark::Before,
        2 => ExpandMark::After,
        3 => ExpandMark::Both,
        _ => return -1,
    };

    let text_obj_id = match get_text_obj_id() {
        Some(id) => id,
        None => return -3,
    };

    let result = with_doc_mut(|doc| {
        let mark = Mark {
            start,
            end,
            name: name.into(), // String implements Into<SmolStr>
            value: ScalarValue::Str(value.into()),
        };
        doc.mark(&text_obj_id, mark, expand_mode)
    });

    match result {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -2,
        None => -3,
    }
}

/// Remove a mark (formatting) from a range of text.
///
/// # Parameters
/// - `name_ptr`: Pointer to mark name string (UTF-8)
/// - `name_len`: Length of name in bytes
/// - `start`: Start index of the range
/// - `end`: End index of the range (exclusive)
/// - `expand`: Expand mode (0=none, 1=before, 2=after, 3=both)
///
/// # Returns
/// - `0` on success
/// - `-1` on UTF-8 validation error
/// - `-2` on Automerge error
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_unmark(
    name_ptr: *const u8,
    name_len: usize,
    start: usize,
    end: usize,
    expand: u8,
) -> i32 {
    if name_ptr.is_null() {
        return -1;
    }

    let name_slice = unsafe { std::slice::from_raw_parts(name_ptr, name_len) };
    let name = match std::str::from_utf8(name_slice) {
        Ok(s) => s,
        Err(_) => return -1,
    };

    let expand_mode = match expand {
        0 => ExpandMark::None,
        1 => ExpandMark::Before,
        2 => ExpandMark::After,
        3 => ExpandMark::Both,
        _ => return -1,
    };

    let text_obj_id = match get_text_obj_id() {
        Some(id) => id,
        None => return -3,
    };

    let result = with_doc_mut(|doc| {
        doc.unmark(&text_obj_id, name, start, end, expand_mode)
    });

    match result {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -2,
        None => -3,
    }
}

/// Get the number of marks at a specific index.
///
/// Call this before `am_get_marks()` to allocate buffer.
///
/// # Parameters
/// - `index`: Index to query
///
/// # Returns
/// - Number of marks at the index
/// - `0` if no marks or document not initialized
#[no_mangle]
pub extern "C" fn am_get_marks_count(index: usize) -> u32 {
    let text_obj_id = match get_text_obj_id() {
        Some(id) => id,
        None => return 0,
    };

    let result = with_doc(|doc| {
        match doc.marks(&text_obj_id) {
            Ok(marks) => {
                // Count marks that apply at this index
                marks
                    .into_iter()
                    .filter(|mark| {
                        mark.start <= index && index < mark.end
                    })
                    .count()
            }
            Err(_) => 0,
        }
    });

    result.unwrap_or(0) as u32
}

/// Get all marks in the text object.
///
/// Returns marks as JSON array string.
/// Format: [{"name": "bold", "value": "true", "start": 0, "end": 5}, ...]
///
/// Call `am_marks_len()` first to allocate buffer.
///
/// # Parameters
/// - `marks_out`: Pointer to buffer to receive JSON string
///
/// # Returns
/// - `0` on success
/// - `-1` if marks_out is null
/// - `-2` on Automerge error
/// - `-3` if document not initialized
#[no_mangle]
pub extern "C" fn am_marks(marks_out: *mut u8) -> i32 {
    if marks_out.is_null() {
        return -1;
    }

    let text_obj_id = match get_text_obj_id() {
        Some(id) => id,
        None => return -3,
    };

    let result = with_doc(|doc| {
        match doc.marks(&text_obj_id) {
            Ok(marks) => {
                // Build JSON array
                let mut json = String::from("[");
                let mut first = true;

                for mark in marks {
                    if !first {
                        json.push(',');
                    }
                    first = false;

                    let value_str = match mark.value() {
                        ScalarValue::Str(s) => s.to_string(),
                        ScalarValue::Boolean(b) => b.to_string(),
                        ScalarValue::Int(i) => i.to_string(),
                        ScalarValue::Uint(u) => u.to_string(),
                        _ => "null".to_string(),
                    };

                    json.push_str(&format!(
                        r#"{{"name":"{}","value":"{}","start":{},"end":{}}}"#,
                        mark.name(),
                        value_str,
                        mark.start,
                        mark.end
                    ));
                }

                json.push(']');

                let bytes = json.as_bytes();
                unsafe {
                    std::ptr::copy_nonoverlapping(bytes.as_ptr(), marks_out, bytes.len());
                }

                Ok(0)
            }
            Err(_) => Err(-2),
        }
    });

    match result {
        Some(Ok(code)) => code,
        Some(Err(code)) => code,
        None => -3,
    }
}

/// Get the length of the marks JSON string.
///
/// Call this before `am_marks()` to allocate buffer.
///
/// # Returns
/// - Length in bytes of JSON string
/// - `0` if document not initialized or no marks
#[no_mangle]
pub extern "C" fn am_marks_len() -> u32 {
    let text_obj_id = match get_text_obj_id() {
        Some(id) => id,
        None => return 0,
    };

    let result = with_doc(|doc| {
        match doc.marks(&text_obj_id) {
            Ok(marks) => {
                // Estimate JSON size
                let mut len = 2; // "[]"
                let mut first = true;

                for mark in marks {
                    if !first {
                        len += 1; // ","
                    }
                    first = false;

                    let value_str = match mark.value() {
                        ScalarValue::Str(s) => s.len(),
                        _ => 10, // Rough estimate for numbers/bools
                    };

                    // Estimate: {"name":"X","value":"Y","start":Z,"end":W}
                    len += 40 + mark.name().len() + value_str + 20; // Conservative estimate
                }

                len
            }
            Err(_) => 0,
        }
    });

    result.unwrap_or(0) as u32
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::document::am_init;
    use crate::memory::{am_alloc, am_free};
    use crate::text::am_text_splice;

    #[test]
    fn test_mark_basic() {
        assert_eq!(am_init(), 0);

        // Add some text
        let text = "Hello World";
        assert_eq!(am_text_splice(0, 0, text.as_ptr(), text.len()), 0);

        // Mark "Hello" as bold
        let name = "bold";
        let value = "true";
        let result = am_mark(
            name.as_ptr(),
            name.len(),
            value.as_ptr(),
            value.len(),
            0,
            5,
            3, // ExpandBoth
        );
        assert_eq!(result, 0);
    }

    #[test]
    fn test_unmark() {
        assert_eq!(am_init(), 0);

        let text = "Hello World";
        assert_eq!(am_text_splice(0, 0, text.as_ptr(), text.len()), 0);

        // Mark as bold
        let name = "bold";
        let value = "true";
        assert_eq!(
            am_mark(
                name.as_ptr(),
                name.len(),
                value.as_ptr(),
                value.len(),
                0,
                11,
                0
            ),
            0
        );

        // Unmark first word
        let result = am_unmark(name.as_ptr(), name.len(), 0, 5, 0);
        assert_eq!(result, 0);
    }

    #[test]
    fn test_marks_json() {
        assert_eq!(am_init(), 0);

        let text = "Hello";
        assert_eq!(am_text_splice(0, 0, text.as_ptr(), text.len()), 0);

        // Add mark
        let name = "bold";
        let value = "true";
        assert_eq!(
            am_mark(
                name.as_ptr(),
                name.len(),
                value.as_ptr(),
                value.len(),
                0,
                5,
                0
            ),
            0
        );

        // Get marks
        let len = am_marks_len();
        if len > 0 {
            let buf_ptr = am_alloc(len as usize);
            assert!(!buf_ptr.is_null());

            let result = am_marks(buf_ptr);
            assert_eq!(result, 0);

            // Verify it's valid JSON starting with '['
            let json_slice = unsafe { std::slice::from_raw_parts(buf_ptr, len as usize) };
            assert_eq!(json_slice[0], b'[');

            am_free(buf_ptr, len as usize);
        }
    }

    #[test]
    fn test_get_marks_count() {
        assert_eq!(am_init(), 0);

        let text = "Hello";
        assert_eq!(am_text_splice(0, 0, text.as_ptr(), text.len()), 0);

        // Initially no marks
        let count = am_get_marks_count(0);
        assert_eq!(count, 0);

        // Add mark
        let name = "bold";
        let value = "true";
        assert_eq!(
            am_mark(
                name.as_ptr(),
                name.len(),
                value.as_ptr(),
                value.len(),
                0,
                5,
                0
            ),
            0
        );

        // Now should have marks
        let count = am_get_marks_count(0);
        assert!(count >= 1);
    }
}
