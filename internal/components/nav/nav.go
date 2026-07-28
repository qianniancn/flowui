// Package nav is a behavioral primitive for keyboard navigation over an indexed
// list: enabled-aware movement, edge/wrap handling, and type-to-search. It is
// parameterized by index rather than a concrete item type, so callers keep their
// own slices and back the Disabled/Label accessors with whatever storage is
// cheapest (e.g. an O(1) set for Disabled). The package is UI-framework-free;
// only Printable depends on gioui.org/io/key for key-name decoding.
package nav

import (
	"strings"
	"time"
	"unicode"

	"gioui.org/io/key"
)

// List describes an indexed, navigable collection. Disabled reports whether an
// index is skippable; Label supplies typeahead text and may be nil when
// typeahead is unused.
type List struct {
	Count    int
	Disabled func(i int) bool
	Label    func(i int) string
}

func (l List) disabled(i int) bool {
	return l.Disabled != nil && l.Disabled(i)
}

// First returns the first enabled index, or (-1, false) when none is enabled.
func First(l List) (int, bool) {
	for i := 0; i < l.Count; i++ {
		if !l.disabled(i) {
			return i, true
		}
	}
	return -1, false
}

// Last returns the last enabled index, or (-1, false) when none is enabled.
func Last(l List) (int, bool) {
	for i := l.Count - 1; i >= 0; i-- {
		if !l.disabled(i) {
			return i, true
		}
	}
	return -1, false
}

// NearestEnabled searches from current+delta in the delta direction, then in the
// opposite direction, for the nearest enabled index. It recovers navigation when
// current itself is disabled. Returns (current, false) when nothing is enabled.
func NearestEnabled(l List, current, delta int) (int, bool) {
	if delta == 0 {
		delta = 1
	}
	for next := current + delta; next >= 0 && next < l.Count; next += delta {
		if !l.disabled(next) {
			return next, true
		}
	}
	for next := current - delta; next >= 0 && next < l.Count; next -= delta {
		if !l.disabled(next) {
			return next, true
		}
	}
	return current, false
}

// Move returns the next enabled index from current stepping by delta (usually
// ±1). Behavior:
//   - current out of range: delta<0 → Last, otherwise First.
//   - wrap=true: wrap around the ends, skipping disabled indices (a disabled
//     current is handled by the same forward wrap scan).
//   - wrap=false: stop at the edges, returning (current, false); a disabled
//     current recovers via NearestEnabled (delta direction first, then reverse).
//
// ok=false means no move happened.
func Move(l List, current, delta int, wrap bool) (int, bool) {
	if l.Count == 0 {
		return -1, false
	}
	if delta == 0 {
		return current, false
	}
	if current < 0 || current >= l.Count {
		if delta < 0 {
			return Last(l)
		}
		return First(l)
	}
	if wrap {
		for step := 1; step < l.Count; step++ {
			next := ((current+delta*step)%l.Count + l.Count) % l.Count
			if !l.disabled(next) {
				return next, true
			}
		}
		return current, false
	}
	if l.disabled(current) {
		return NearestEnabled(l, current, delta)
	}
	for next := current + delta; next >= 0 && next < l.Count; next += delta {
		if !l.disabled(next) {
			return next, true
		}
	}
	return current, false
}

// Match finds, searching forward from current with wraparound, the first enabled
// index whose Label has the given case-insensitive prefix. current may be -1 to
// start the search at index 0. Returns (current, false) on no match; also false
// when the list is empty, the query is empty, or Label is nil.
func Match(l List, current int, query string) (int, bool) {
	if l.Count == 0 || query == "" || l.Label == nil {
		return current, false
	}
	query = strings.ToLower(query)
	for step := 1; step <= l.Count; step++ {
		index := ((current+step)%l.Count + l.Count) % l.Count
		if l.disabled(index) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(l.Label(index)), query) {
			return index, true
		}
	}
	return current, false
}

// Printable maps a key.Name to a single lowercase printable character for
// typeahead, or "" for control or multi-rune names.
func Printable(name key.Name) string {
	runes := []rune(string(name))
	if len(runes) != 1 || unicode.IsControl(runes[0]) {
		return ""
	}
	return strings.ToLower(string(runes[0]))
}

// DefaultTypeaheadTimeout is the idle gap after which a Typeahead query resets.
const DefaultTypeaheadTimeout = 500 * time.Millisecond

// Typeahead accumulates incremental type-to-search input. Store it on the
// component's frame-persistent state; the zero value is ready to use.
type Typeahead struct {
	text  string
	at    time.Time
	ready bool
	// Timeout overrides DefaultTypeaheadTimeout when > 0.
	Timeout time.Duration
}

func (t *Typeahead) timeout() time.Duration {
	if t.Timeout <= 0 {
		return DefaultTypeaheadTimeout
	}
	return t.Timeout
}

// Append adds one character, resetting the accumulated query first when the
// timeout has elapsed since the previous keystroke (or now moved backward). It
// returns the accumulated query.
func (t *Typeahead) Append(now time.Time, text string) string {
	if !t.ready || now.Before(t.at) || now.Sub(t.at) > t.timeout() {
		t.text = ""
	}
	t.text += text
	t.at = now
	t.ready = true
	return t.text
}

// Set overrides the accumulated query, e.g. to fall back to a single character
// when a multi-character query matched nothing.
func (t *Typeahead) Set(text string) {
	t.text = text
}

// Reset clears the accumulated query.
func (t *Typeahead) Reset() {
	t.text = ""
	t.ready = false
}
