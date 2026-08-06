package state

import "fmt"

// Identity identifies one frame-persistent interaction state slot.
// Key is the component's explicit, scoped MVU identity. Slot distinguishes
// multiple Gio state objects owned by the same component.
type Identity struct {
	Key  string
	Slot string
}

// Store owns Gio interaction state that must survive across frames.
// Application and component values belong in the user's MVU Model instead.
type Store struct {
	values           map[Identity]any
	used             map[Identity]struct{}
	retained         map[string]map[Identity]struct{}
	retainedCurrent  map[Identity]struct{}
	retentionTouched map[string]struct{}
}

// BeginFrame starts state usage tracking and notifies existing frame-aware
// states before widgets process the new frame.
func (s *Store) BeginFrame() {
	if s.used == nil {
		s.used = make(map[Identity]struct{})
	} else {
		clear(s.used)
	}
	if s.retainedCurrent == nil {
		s.retainedCurrent = make(map[Identity]struct{})
	} else {
		clear(s.retainedCurrent)
	}
	if s.retentionTouched == nil {
		s.retentionTouched = make(map[string]struct{})
	} else {
		clear(s.retentionTouched)
	}
	for _, value := range s.values {
		if state, ok := value.(interface{ BeginFrame() }); ok {
			state.BeginFrame()
		}
	}
}

// EndFrame releases states whose components were not rendered this frame.
func (s *Store) EndFrame() {
	for id := range s.values {
		if _, ok := s.used[id]; ok {
			continue
		}
		if _, ok := s.retainedCurrent[id]; !ok {
			delete(s.values, id)
		}
	}
	for scope := range s.retained {
		if _, ok := s.retentionTouched[scope]; !ok {
			delete(s.retained, scope)
		}
	}
}

// Retain marks the states previously observed under scope as live for the
// current frame. A scope must be touched again on every frame that owns it.
func (s *Store) Retain(scope string) {
	if scope == "" {
		panic("flowui: empty state retention scope")
	}
	if s.retained == nil {
		s.retained = make(map[string]map[Identity]struct{})
	}
	if s.retainedCurrent == nil {
		s.retainedCurrent = make(map[Identity]struct{})
	}
	if s.retentionTouched == nil {
		s.retentionTouched = make(map[string]struct{})
	}
	s.retentionTouched[scope] = struct{}{}
	ids := s.retained[scope]
	for id := range ids {
		s.retainedCurrent[id] = struct{}{}
	}
}

// RecordRetention associates a state identity with a retention scope. It is
// called automatically by frame.UseState while a component owns a scope.
func (s *Store) RecordRetention(scope string, id Identity) {
	if scope == "" {
		panic("flowui: empty state retention scope")
	}
	validateIdentity(id)
	if s.retained == nil {
		s.retained = make(map[string]map[Identity]struct{})
	}
	if s.retentionTouched == nil {
		s.retentionTouched = make(map[string]struct{})
	}
	ids := s.retained[scope]
	if ids == nil {
		ids = make(map[Identity]struct{})
		s.retained[scope] = ids
	}
	ids[id] = struct{}{}
	s.retentionTouched[scope] = struct{}{}
	s.retainedCurrent[id] = struct{}{}
}

// ReleaseRetention drops all identities remembered for scope. The scope will
// no longer keep state alive once the current frame ends.
func (s *Store) ReleaseRetention(scope string) {
	if scope == "" {
		return
	}
	delete(s.retained, scope)
	delete(s.retentionTouched, scope)
}

// Use returns an existing typed state or creates it with factory.
func Use[T any](s *Store, id Identity, factory func() *T) *T {
	validateIdentity(id)
	if s.values == nil {
		s.values = make(map[Identity]any)
	}
	if s.used == nil {
		s.used = make(map[Identity]struct{})
	}
	s.used[id] = struct{}{}
	if value, ok := s.values[id]; ok {
		state, ok := value.(*T)
		if !ok {
			panic(fmt.Sprintf("flowui: state slot %q for key %q has type %T, want %T", id.Slot, id.Key, value, (*T)(nil)))
		}
		return state
	}
	if factory == nil {
		factory = func() *T { return new(T) }
	}
	state := factory()
	if state == nil {
		panic(fmt.Sprintf("flowui: state factory returned nil for slot %q and key %q", id.Slot, id.Key))
	}
	s.values[id] = state
	return state
}

// Peek returns a state without retaining it for the current frame.
func Peek[T any](s *Store, id Identity) (*T, bool) {
	validateIdentity(id)
	value, ok := s.values[id]
	if !ok {
		return nil, false
	}
	state, ok := value.(*T)
	if !ok {
		panic(fmt.Sprintf("flowui: state slot %q for key %q has type %T, want %T", id.Slot, id.Key, value, (*T)(nil)))
	}
	return state, true
}

// Delete releases one interaction state slot immediately.
func (s *Store) Delete(id Identity) {
	validateIdentity(id)
	delete(s.values, id)
	delete(s.used, id)
}

// Len reports the number of retained interaction state slots.
func (s *Store) Len() int {
	return len(s.values)
}

func validateIdentity(id Identity) {
	if id.Key == "" {
		panic("flowui: empty state key")
	}
	if id.Slot == "" {
		panic(fmt.Sprintf("flowui: empty state slot for key %q", id.Key))
	}
}
