// ==============================================================================
// Layer 7: Web Frontend - Sync Protocol (JavaScript Module)
// ==============================================================================
// ARCHITECTURE: This is the web frontend layer (Layer 7/7).
//
// RESPONSIBILITIES:
// - User interaction (DOM events, form handling)
// - HTTP API calls to Layer 6 (fetch /api/sync)
// - UI updates (rendering sync status, log messages)
// - State management (peer ID, sync state, log entries)
//
// DEPENDENCIES:
// - Layer 6: pkg/api/crdt_sync.go (HTTP handlers)
// - Browser APIs: fetch(), DOM
//
// RELATED FILES (1:1 mapping):
// - Layer 2: rust/automerge_wasi/src/sync.rs (WASI exports)
// - Layer 3: pkg/wazero/crdt_sync.go (FFI wrappers)
// - Layer 4: pkg/automerge/crdt_sync.go (pure CRDT API)
// - Layer 5: pkg/server/crdt_sync.go (stateful server operations)
// - Layer 6: pkg/api/crdt_sync.go (HTTP handlers)
// - Layer 7: web/components/crdt_sync.html (HTML template)
//
// NOTES:
// - This module exports a class that's instantiated by app.js
// - All state is local to this component (not shared)
// - HTTP calls use base64-encoded sync messages
// ==============================================================================

// M1: Sync Protocol Module
// Maps to: go/pkg/automerge/sync.go, go/pkg/api/sync.go

export class SyncComponent {
    constructor() {
        this.peerID = null;
        this.syncState = null;
        this.logEntries = [];
    }

    init() {
        const peerInput = document.getElementById('sync-peer-id');
        const initBtn = document.getElementById('sync-init');
        const sendBtn = document.getElementById('sync-send');
        const clearLogBtn = document.getElementById('sync-clear-log');

        if (!peerInput) return; // Not on this tab

        // Set default peer ID
        if (peerInput.value === 'browser-peer') {
            peerInput.value = `browser-${Date.now()}`;
        }

        initBtn?.addEventListener('click', () => this.initSync());
        sendBtn?.addEventListener('click', () => this.sendSync());
        clearLogBtn?.addEventListener('click', () => this.clearLog());

        this.log('Sync component initialized. Click "Initialize Sync" to begin.');
    }

    async initSync() {
        const peerInput = document.getElementById('sync-peer-id');
        this.peerID = peerInput.value.trim();

        if (!this.peerID) {
            alert('Please enter a peer ID');
            return;
        }

        this.log(`Initializing sync for peer: ${this.peerID}`, 'info');

        // Send initial sync message
        try {
            const response = await fetch('/api/sync', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    peer_id: this.peerID,
                    message: '', // Empty message for init
                }),
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${await response.text()}`);
            }

            const data = await response.json();
            this.log(`✓ Sync initialized. has_more=${data.has_more}`, 'success');

            if (data.message) {
                this.log(`Received sync message (${data.message.length} chars base64)`, 'success');
                this.displayResponse(data);
            }

            const statusEl = document.getElementById('sync-peer-status');
            if (statusEl) {
                statusEl.textContent = `✓ Initialized as ${this.peerID}`;
                statusEl.className = 'status-text status-success';
            }
        } catch (error) {
            this.log(`✗ Sync init failed: ${error.message}`, 'error');
        }
    }

    async sendSync() {
        if (!this.peerID) {
            alert('Please initialize sync first');
            return;
        }

        this.log(`Sending sync message for peer: ${this.peerID}`, 'info');

        try {
            const response = await fetch('/api/sync', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    peer_id: this.peerID,
                    message: '', // Server will generate sync message
                }),
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${await response.text()}`);
            }

            const data = await response.json();
            this.log(`✓ Sync message sent. has_more=${data.has_more}`, 'success');

            if (data.message) {
                this.log(`Received response (${data.message.length} chars base64)`, 'success');
                this.displayResponse(data);
            } else {
                this.log('No response message (peer is up to date)', 'info');
            }
        } catch (error) {
            this.log(`✗ Sync failed: ${error.message}`, 'error');
        }
    }

    displayResponse(data) {
        const responseEl = document.getElementById('sync-response');
        if (!responseEl) return;

        const formatted = {
            has_more: data.has_more,
            message_length: data.message ? data.message.length : 0,
            message_preview: data.message ? data.message.substring(0, 50) + '...' : null,
        };

        responseEl.textContent = JSON.stringify(formatted, null, 2);
    }

    log(message, type = 'info') {
        this.logEntries.push({ message, type, timestamp: new Date() });
        this.updateLogDisplay();
    }

    updateLogDisplay() {
        const logEl = document.getElementById('sync-log');
        if (!logEl) return;

        logEl.innerHTML = this.logEntries.map((entry, i) => {
            const time = entry.timestamp.toLocaleTimeString();
            return `<div class="log-entry log-${entry.type}">
                <span class="log-time">[${time}]</span>
                <span class="log-message">${entry.message}</span>
            </div>`;
        }).reverse().join('');
    }

    clearLog() {
        this.logEntries = [];
        this.updateLogDisplay();
    }
}
