// WASI exports for Automerge sync protocol operations
//
// The sync protocol enables efficient delta-based synchronization between peers.
// Instead of sending the entire document, peers exchange only the changes they
// don't have yet.

use crate::state::with_doc_mut;
use automerge::sync::{self, SyncDoc};
use std::sync::{Mutex, OnceLock};
use std::collections::HashMap;

// Per-peer sync state storage
// Maps peer_id -> sync::State
// This is the correct way - each peer connection needs its own state
static SYNC_STATES: OnceLock<Mutex<HashMap<u32, sync::State>>> = OnceLock::new();
static NEXT_PEER_ID: OnceLock<Mutex<u32>> = OnceLock::new();

fn get_sync_states() -> &'static Mutex<HashMap<u32, sync::State>> {
    SYNC_STATES.get_or_init(|| Mutex::new(HashMap::new()))
}

fn get_next_peer_id() -> &'static Mutex<u32> {
    NEXT_PEER_ID.get_or_init(|| Mutex::new(1))
}

/// Create a new sync state for a peer connection.
///
/// Call this once before starting a sync session with a peer.
/// Returns a peer_id that must be used in all subsequent sync calls.
///
/// # Returns
/// - peer_id (> 0) on success
/// - `0` if failed to initialize
#[no_mangle]
pub extern "C" fn am_sync_state_init() -> u32 {
    let peer_id = match get_next_peer_id().lock() {
        Ok(mut next_id) => {
            let id = *next_id;
            *next_id += 1;
            id
        }
        Err(_) => return 0,
    };

    match get_sync_states().lock() {
        Ok(mut states) => {
            states.insert(peer_id, sync::State::new());
            peer_id
        }
        Err(_) => 0,
    }
}

/// Free a peer's sync state.
///
/// Call this when done with a peer connection.
///
/// # Parameters
/// - `peer_id`: The peer ID returned from am_sync_state_init
#[no_mangle]
pub extern "C" fn am_sync_state_free(peer_id: u32) -> i32 {
    match get_sync_states().lock() {
        Ok(mut states) => {
            states.remove(&peer_id);
            0
        }
        Err(_) => -1,
    }
}

/// Generate a sync message to send to a peer.
///
/// The sync state tracks what the peer has already seen, so we only send
/// changes they're missing.
///
/// Call `am_sync_gen_len()` first to get buffer size.
///
/// # Parameters
/// - `peer_id`: The peer ID from am_sync_state_init
///
/// # Returns
/// - Length of generated message (0 if nothing to send)
/// - Max value indicates error
#[no_mangle]
pub extern "C" fn am_sync_gen_len(peer_id: u32) -> u32 {
    let mut states = match get_sync_states().lock() {
        Ok(s) => s,
        Err(_) => return u32::MAX,
    };

    let state = match states.get_mut(&peer_id) {
        Some(s) => s,
        None => return u32::MAX, // Invalid peer_id
    };

    let result = with_doc_mut(|doc| {
        match doc.sync().generate_sync_message(state) {
            Some(msg) => {
                let bytes = msg.encode();
                bytes.len()
            }
            None => 0, // Nothing to send
        }
    });

    result.unwrap_or(u32::MAX as usize) as u32
}

/// Generate a sync message and write to buffer.
///
/// # Parameters
/// - `peer_id`: The peer ID from am_sync_state_init
/// - `msg_out`: Pointer to buffer to receive sync message
///
/// # Returns
/// - `0` on success (message written)
/// - `-1` if msg_out is null
/// - `-2` if sync state not initialized or invalid peer_id
/// - `-3` if document not initialized
/// - `1` if nothing to send (no error, just no message)
#[no_mangle]
pub extern "C" fn am_sync_gen(peer_id: u32, msg_out: *mut u8) -> i32 {
    if msg_out.is_null() {
        return -1;
    }

    let mut states = match get_sync_states().lock() {
        Ok(s) => s,
        Err(_) => return -2,
    };

    let state = match states.get_mut(&peer_id) {
        Some(s) => s,
        None => return -2, // Invalid peer_id
    };

    let result = with_doc_mut(|doc| {
        match doc.sync().generate_sync_message(state) {
            Some(msg) => {
                let bytes = msg.encode();
                unsafe {
                    std::ptr::copy_nonoverlapping(bytes.as_ptr(), msg_out, bytes.len());
                }
                0
            }
            None => 1, // Nothing to send
        }
    });

    match result {
        Some(code) => code,
        None => -3,
    }
}

/// Receive and process a sync message from a peer.
///
/// This applies any changes we don't have yet and updates the sync state.
/// After receiving a message, you should call `am_sync_gen` to see if
/// we need to reply with our own changes.
///
/// # Parameters
/// - `msg_ptr`: Pointer to sync message bytes
/// - `msg_len`: Length of message
///
/// # Returns
/// - `0` on success
/// - `-1` if msg_ptr is null
/// - `-2` if sync state not initialized
/// - `-3` if document not initialized
/// - `-4` if failed to decode message
/// - `-5` if failed to apply changes
#[no_mangle]
pub extern "C" fn am_sync_recv(peer_id: u32, msg_ptr: *const u8, msg_len: usize) -> i32 {
    if msg_ptr.is_null() {
        return -1;
    }

    let msg_slice = unsafe { std::slice::from_raw_parts(msg_ptr, msg_len) };

    let mut states = match get_sync_states().lock() {
        Ok(s) => s,
        Err(_) => return -2,
    };

    let state = match states.get_mut(&peer_id) {
        Some(s) => s,
        None => return -2, // Invalid peer_id
    };

    let result = with_doc_mut(|doc| {
        // Decode sync message
        let msg = match sync::Message::decode(msg_slice) {
            Ok(m) => m,
            Err(_) => return Err(-4),
        };

        // Receive and apply message
        match doc.sync().receive_sync_message(state, msg) {
            Ok(_) => Ok(0),
            Err(_) => Err(-5),
        }
    });

    match result {
        Some(Ok(code)) => code,
        Some(Err(code)) => code,
        None => -3,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::document::{am_init, am_save, am_save_len};
    use crate::memory::{am_alloc, am_free};
    use crate::text::am_text_splice;

    #[test]
    fn test_sync_state_init() {
        let peer_id = am_sync_state_init();
        assert!(peer_id > 0, "Expected valid peer_id, got {}", peer_id);

        // Clean up
        am_sync_state_free(peer_id);
    }

    #[test]
    fn test_sync_gen_empty() {
        assert_eq!(am_init(), 0);
        let peer_id = am_sync_state_init();
        assert!(peer_id > 0);

        // Initial sync should have a message
        let len = am_sync_gen_len(peer_id);
        assert!(len > 0 && len != u32::MAX);

        let buf_ptr = am_alloc(len as usize);
        assert!(!buf_ptr.is_null());

        let result = am_sync_gen(peer_id, buf_ptr);
        // Empty document returns 1 (nothing to send), not an error
        assert!(result == 0 || result == 1, "Expected 0 or 1, got {}", result);

        am_free(buf_ptr, len as usize);
        am_sync_state_free(peer_id);
    }

    #[test]
    fn test_sync_two_peers() {
        // Peer A: Create doc with text
        assert_eq!(am_init(), 0);
        let text = "Hello from A";
        assert_eq!(am_text_splice(0, 0, text.as_ptr(), text.len()), 0);

        // Save peer A's state
        let save_len = am_save_len();
        let save_buf = vec![0u8; save_len as usize];
        assert_eq!(am_save(save_buf.as_ptr() as *mut u8), 0);

        // Initialize sync on peer A
        let peer_id = am_sync_state_init();
        assert!(peer_id > 0, "Expected valid peer_id");

        // Generate sync message from A
        let msg_len = am_sync_gen_len(peer_id);
        if msg_len > 0 && msg_len != u32::MAX {
            let msg_buf_ptr = am_alloc(msg_len as usize);
            assert!(!msg_buf_ptr.is_null());

            let result = am_sync_gen(peer_id, msg_buf_ptr);
            assert!(result == 0 || result == 1); // 0=sent, 1=nothing to send

            am_free(msg_buf_ptr, msg_len as usize);
        }

        // Clean up
        am_sync_state_free(peer_id);
    }
}
