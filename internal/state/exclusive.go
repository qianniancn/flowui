package state

// Exclusive coordinates controls where at most one member of a group may be
// active, such as Select overlays within one window.
type Exclusive struct {
	active    map[string]exclusiveMember
	callbacks map[exclusiveID]func()
	seen      map[exclusiveID]struct{}
}

type exclusiveID struct {
	group string
	key   string
}

type exclusiveMember struct {
	key   string
	close func()
}

func (e *Exclusive) BeginFrame() {
	if e.seen == nil {
		e.seen = make(map[exclusiveID]struct{})
	} else {
		clear(e.seen)
	}
}

func (e *Exclusive) Register(group, key string, close func()) {
	if group == "" || key == "" {
		panic("flowui: exclusive group and key must not be empty")
	}
	if e.callbacks == nil {
		e.callbacks = make(map[exclusiveID]func())
	}
	if e.seen == nil {
		e.seen = make(map[exclusiveID]struct{})
	}
	id := exclusiveID{group: group, key: key}
	e.callbacks[id] = close
	e.seen[id] = struct{}{}
	if active, ok := e.active[group]; ok && active.key == key {
		active.close = close
		e.active[group] = active
	}
}

func (e *Exclusive) Activate(group, key string) {
	id := exclusiveID{group: group, key: key}
	close, ok := e.callbacks[id]
	if !ok {
		panic("flowui: exclusive member must be registered before activation")
	}
	if e.active == nil {
		e.active = make(map[string]exclusiveMember)
	}
	if current, ok := e.active[group]; ok {
		if current.key == key {
			current.close = close
			e.active[group] = current
			return
		}
		delete(e.active, group)
		if current.close != nil {
			current.close()
		}
	}
	e.active[group] = exclusiveMember{key: key, close: close}
}

func (e *Exclusive) Release(group, key string) {
	if current, ok := e.active[group]; ok && current.key == key {
		delete(e.active, group)
	}
}

func (e *Exclusive) EndFrame() {
	for group, current := range e.active {
		if _, ok := e.seen[exclusiveID{group: group, key: current.key}]; !ok {
			delete(e.active, group)
			// Unmounted open members must still run close so uncontrolled
			// overlays (Select/Menu) clear open trackers.
			if current.close != nil {
				current.close()
			}
		}
	}
	for id := range e.callbacks {
		if _, ok := e.seen[id]; !ok {
			delete(e.callbacks, id)
		}
	}
}

func (e *Exclusive) Active(group string) string {
	return e.active[group].key
}
