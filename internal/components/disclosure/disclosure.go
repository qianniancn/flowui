// Package disclosure is a behavioral primitive for controlled/uncontrolled
// open-state bindings. It manages the three-mode contract:
//
//	no declaration   → uncontrolled; widget keeps value internally
//	Controlled=true  → controlled; caller owns the value and handles OnChange
//	HasDefault=true  → seeds uncontrolled initial value only
//
// T is bool for open/close, string for "which item is open" (e.g. Menubar).
// Exclusive-group registration, focus side-effects, and other per-control
// concerns are left to the caller; Binding owns only the value state machine.
package disclosure

// Config is the per-frame snapshot of the widget's open-state declarations.
// Prepare one from the widget fields and pass it to each Binding method.
type Config[T comparable] struct {
	// Controlled reports whether the caller is managing the value externally.
	Controlled bool
	// Value is the externally supplied value, read when Controlled is true.
	Value T
	// HasDefault seeds the uncontrolled initial value on the first frame.
	HasDefault bool
	// Default is the seed value, used only when HasDefault is true.
	Default T
	// OnChange is called when the effective value changes; may be nil.
	OnChange func(T)
}

// Binding manages a single open/expanded value. The zero value is ready to
// use; store it in the component's frame-persistent state struct.
type Binding[T comparable] struct {
	value       T
	initialized bool
	snap        Config[T] // snapshot of the last Bind call, for PeerClose
}

// Current returns the effective value for this frame:
//   - On the first call, seeds the uncontrolled value from Config.Default when
//     Config.HasDefault is true.
//   - When Config.Controlled is true, returns Config.Value directly (the caller
//     owns it; Binding does not mutate its internal copy).
//   - Otherwise returns the internally maintained uncontrolled value.
func (b *Binding[T]) Current(cfg Config[T]) T {
	if !b.initialized {
		if cfg.HasDefault {
			b.value = cfg.Default
		}
		b.initialized = true
	}
	if cfg.Controlled {
		return cfg.Value
	}
	return b.value
}

// Request transitions to next. Returns (effective, changed):
//   - For controlled bindings: fires OnChange only when Config.Value != next;
//     does not mutate internal state (the external model owns the value).
//     Returns the current Config.Value (unchanged until the model feeds it back).
//   - For uncontrolled bindings: mutates internal value and fires OnChange only
//     when b.value != next. Returns the new internal value.
func (b *Binding[T]) Request(cfg Config[T], next T) (effective T, changed bool) {
	if cfg.Controlled {
		if cfg.Value != next {
			changed = true
			if cfg.OnChange != nil {
				cfg.OnChange(next)
			}
		}
		return cfg.Value, changed
	}
	if b.value != next {
		b.value = next
		if cfg.OnChange != nil {
			cfg.OnChange(next)
		}
		changed = true
	}
	return b.value, changed
}

// Bind snapshots cfg so that PeerClose can use last-frame controlled/onChange
// values even when called outside the normal frame flow (e.g. from an
// exclusive-group callback). Call once per frame after the widget is bound.
func (b *Binding[T]) Bind(cfg Config[T]) {
	b.snap = cfg
}

// PeerClose transitions to zero (the "closed" value) using the last snapshotted
// Config. Intended for exclusive-group callbacks that fire between frames.
// Returns true when a change was made or requested.
func (b *Binding[T]) PeerClose(zero T) bool {
	cfg := b.snap
	if cfg.Controlled {
		if cfg.Value != zero && cfg.OnChange != nil {
			cfg.OnChange(zero)
			return true
		}
		return false
	}
	if b.value != zero {
		b.value = zero
		if cfg.OnChange != nil {
			cfg.OnChange(zero)
		}
		return true
	}
	return false
}
