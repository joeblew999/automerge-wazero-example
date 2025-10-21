//! Generic object operations - SIMPLIFIED VERSION
//! 
//! Basic put/get/delete operations that work with ROOT for now.
//! Full nested path support can be added later.

use crate::state::with_doc_mut;
use automerge::{ObjType, ScalarValue, ReadDoc, transaction::Transactable, ROOT};
use std::str;
use std::cell::RefCell;

thread_local! {
    static LAST_VALUE: RefCell<String> = RefCell::new(String::new());
}

/// Put a scalar value at ROOT level
#[no_mangle]
pub extern "C" fn am_put_root(
    key_ptr: *const u8,
    key_len: usize,
    value_ptr: *const u8,
    value_len: usize,
) -> i32 {
    if key_ptr.is_null() || value_ptr.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key_str = match str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -2,
    };

    let value_slice = unsafe { std::slice::from_raw_parts(value_ptr, value_len) };
    let value_str = match str::from_utf8(value_slice) {
        Ok(s) => s,
        Err(_) => return -3,
    };

    // Simple value parsing
    let scalar: ScalarValue = if let Ok(n) = value_str.parse::<i64>() {
        ScalarValue::Int(n)
    } else if value_str == "true" {
        ScalarValue::Boolean(true)
    } else if value_str == "false" {
        ScalarValue::Boolean(false)
    } else {
        ScalarValue::Str(value_str.trim_matches('"').into())
    };

    match with_doc_mut(|doc| doc.put(&ROOT, key_str, scalar)) {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -4,
        None => -5,
    }
}

/// Get a value from ROOT level
#[no_mangle]
pub extern "C" fn am_get_root(key_ptr: *const u8, key_len: usize) -> i32 {
    if key_ptr.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key_str = match str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -2,
    };

    let value_str = match with_doc_mut(|doc| {
        match doc.get(&ROOT, key_str) {
            Ok(Some((value, _))) => {
                use automerge::Value;
                let s = match value {
                    Value::Scalar(sc) => match sc.as_ref() {
                        automerge::ScalarValue::Str(st) => st.to_string(),
                        automerge::ScalarValue::Int(n) => n.to_string(),
                        automerge::ScalarValue::Boolean(b) => b.to_string(),
                        _ => format!("{:?}", sc),
                    },
                    Value::Object(_) => "<object>".to_string(),
                };
                Ok(s)
            }
            Ok(None) => Err(-3),
            Err(_) => Err(-4),
        }
    }) {
        Some(Ok(s)) => s,
        Some(Err(code)) => return code,
        None => return -5,
    };

    LAST_VALUE.with(|v| *v.borrow_mut() = value_str.clone());
    value_str.len() as i32
}

/// Retrieve value from last get
#[no_mangle]
pub extern "C" fn am_get_root_value(ptr_out: *mut u8) -> i32 {
    if ptr_out.is_null() {
        return -1;
    }
    LAST_VALUE.with(|v| {
        let value = v.borrow();
        let bytes = value.as_bytes();
        unsafe {
            std::ptr::copy_nonoverlapping(bytes.as_ptr(), ptr_out, bytes.len());
        }
        0
    })
}

/// Delete from ROOT
#[no_mangle]
pub extern "C" fn am_delete_root(key_ptr: *const u8, key_len: usize) -> i32 {
    if key_ptr.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key_str = match str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -2,
    };

    match with_doc_mut(|doc| doc.delete(&ROOT, key_str)) {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -3,
        None => -4,
    }
}

/// Put object at ROOT level
#[no_mangle]
pub extern "C" fn am_put_object_root(
    key_ptr: *const u8,
    key_len: usize,
    obj_type_ptr: *const u8,
    obj_type_len: usize,
) -> i32 {
    if key_ptr.is_null() || obj_type_ptr.is_null() {
        return -1;
    }

    let key_slice = unsafe { std::slice::from_raw_parts(key_ptr, key_len) };
    let key_str = match str::from_utf8(key_slice) {
        Ok(s) => s,
        Err(_) => return -2,
    };

    let type_slice = unsafe { std::slice::from_raw_parts(obj_type_ptr, obj_type_len) };
    let type_str = match str::from_utf8(type_slice) {
        Ok(s) => s,
        Err(_) => return -3,
    };

    let obj_type = match type_str {
        "map" => ObjType::Map,
        "list" => ObjType::List,
        "text" => ObjType::Text,
        _ => return -4,
    };

    match with_doc_mut(|doc| doc.put_object(&ROOT, key_str, obj_type)) {
        Some(Ok(_)) => 0,
        Some(Err(_)) => -5,
        None => -6,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::document::am_init;

    #[test]
    fn test_put_get_root() {
        assert_eq!(am_init(), 0);

        let key = "name";
        let value = "Alice";

        assert_eq!(am_put_root(key.as_ptr(), key.len(), value.as_ptr(), value.len()), 0);

        let len = am_get_root(key.as_ptr(), key.len());
        assert!(len > 0);

        let mut buffer = vec![0u8; len as usize];
        assert_eq!(am_get_root_value(buffer.as_mut_ptr()), 0);
        let result = String::from_utf8(buffer).unwrap();
        assert_eq!(result, "Alice");
    }

    #[test]
    fn test_delete_root() {
        assert_eq!(am_init(), 0);

        let key = "temp";
        let value = "test";
        assert_eq!(am_put_root(key.as_ptr(), key.len(), value.as_ptr(), value.len()), 0);
        assert_eq!(am_delete_root(key.as_ptr(), key.len()), 0);
    }
}
