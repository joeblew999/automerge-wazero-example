package automerge

import "context"

// Counter Operations

// Increment increments (or decrements) a counter at a key.
//
// Status: ✅ Implemented for ROOT map
func (d *Document) Increment(ctx context.Context, path Path, key string, delta int64) error {
	if !d.isRootPath(path) {
		return &NotImplementedError{
			Feature:   "Increment (nested objects)",
			Milestone: "M2",
			Message:   "Only ROOT map counters are supported currently",
		}
	}

	// Check if counter exists, create if not
	_, err := d.runtime.AmCounterGet(ctx, key)
	if err != nil {
		// Counter doesn't exist, create it with delta as initial value
		if err := d.runtime.AmCounterCreate(ctx, key, delta); err != nil {
			return err
		}
		return nil
	}

	return d.runtime.AmCounterIncrement(ctx, key, delta)
}

// GetCounter retrieves the current value of a counter.
//
// Status: ✅ Implemented for ROOT map
func (d *Document) GetCounter(ctx context.Context, path Path, key string) (int64, error) {
	if !d.isRootPath(path) {
		return 0, &NotImplementedError{
			Feature:   "GetCounter (nested objects)",
			Milestone: "M2",
			Message:   "Only ROOT map counters are supported currently",
		}
	}

	return d.runtime.AmCounterGet(ctx, key)
}
