package automerge

import "context"

// Counter Operations - For M2 Milestone
//
// Counters are special CRDT values that merge by addition rather than
// last-write-wins. Perfect for like counts, view counts, etc.

// Increment increments a counter at a key by the given delta.
//
// Counters can be incremented (positive delta) or decremented (negative delta).
// When two peers concurrently increment a counter, both changes are preserved
// and the final value is the sum of all increments.
//
// Example:
//   // Peer A: counter = 0, increment by 5 → 5
//   // Peer B: counter = 0, increment by 3 → 3
//   // After merge: counter = 8 (not 5 or 3!)
//
// Status: ❌ Not implemented (requires M2)
func (d *Document) Increment(ctx context.Context, path Path, key string, delta int64) error {
	return &NotImplementedError{
		Feature:   "Increment",
		Milestone: "M2",
		Message:   "Requires am_increment WASI export",
	}
}
