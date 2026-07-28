package systray

import (
	"sync"
	"sync/atomic"
)

// SystemTray represents a system tray icon
type SystemTray struct {
	id      uint
	label   string
	tooltip string
	icon    []byte

	clickHandler       func()
	rightClickHandler  func()
	doubleClickHandler func()
	mouseEnterHandler  func()
	mouseLeaveHandler  func()

	// Platform-specific implementation
	impl systemTrayImpl
	menu *Menu

	running   bool
	destroyed bool
	mu        sync.Mutex

	ready     chan struct{}
	errors    chan error
	done      chan struct{}
	readyOnce sync.Once
	doneOnce  sync.Once
}

// systemTrayImpl is the platform-specific interface
type systemTrayImpl interface {
	setLabel(label string)
	setTooltip(tooltip string)
	setIcon(icon []byte)
	setMenu(menu *Menu)
	run()
	destroy()
	show()
	hide()
}

var (
	trayIDCounter uint32
	trayMap       = make(map[uint]*SystemTray)
	trayMapMu     sync.RWMutex
)

// New creates a new system tray icon
func New() *SystemTray {
	id := atomic.AddUint32(&trayIDCounter, 1)
	tray := &SystemTray{
		id:     uint(id),
		ready:  make(chan struct{}),
		errors: make(chan error, 8),
		done:   make(chan struct{}),
	}

	trayMapMu.Lock()
	trayMap[tray.id] = tray
	trayMapMu.Unlock()

	return tray
}

// Ready is closed after the native tray icon has been registered and can be
// used. It is not closed when startup fails; observe Errors and Done as well.
func (t *SystemTray) Ready() <-chan struct{} {
	return t.ready
}

// Errors reports asynchronous startup and runtime errors. The channel remains
// open for the lifetime of the object; Done signals the end of that lifetime.
func (t *SystemTray) Errors() <-chan error {
	return t.errors
}

// Done is closed when the tray is destroyed or cannot continue running.
func (t *SystemTray) Done() <-chan struct{} {
	return t.done
}

// GetTrayByID retrieves a tray by its ID (internal use)
func GetTrayByID(id uint) *SystemTray {
	trayMapMu.RLock()
	defer trayMapMu.RUnlock()
	return trayMap[id]
}

// SetIcon sets the icon for the system tray
func (t *SystemTray) SetIcon(icon []byte) *SystemTray {
	owned := append([]byte(nil), icon...)
	t.mu.Lock()
	t.icon = owned
	impl := t.impl
	t.mu.Unlock()

	if impl != nil {
		impl.setIcon(owned)
	}
	return t
}

// SetLabel sets the label for the system tray (macOS only)
func (t *SystemTray) SetLabel(label string) *SystemTray {
	t.mu.Lock()
	t.label = label
	impl := t.impl
	t.mu.Unlock()

	if impl != nil {
		impl.setLabel(label)
	}
	return t
}

// SetTooltip sets the tooltip for the system tray
func (t *SystemTray) SetTooltip(tooltip string) *SystemTray {
	t.mu.Lock()
	t.tooltip = tooltip
	impl := t.impl
	t.mu.Unlock()

	if impl != nil {
		impl.setTooltip(tooltip)
	}
	return t
}

// SetMenu sets the menu for the system tray
func (t *SystemTray) SetMenu(menu *Menu) *SystemTray {
	t.mu.Lock()
	t.menu = menu
	impl := t.impl
	t.mu.Unlock()

	if impl != nil {
		impl.setMenu(menu)
	}
	return t
}

// OnClick sets the click handler
func (t *SystemTray) OnClick(handler func()) *SystemTray {
	t.mu.Lock()
	t.clickHandler = handler
	t.mu.Unlock()
	return t
}

// OnRightClick sets a custom right-click handler. A custom handler replaces
// automatic menu display; call ShowMenu from the handler to display it.
func (t *SystemTray) OnRightClick(handler func()) *SystemTray {
	t.mu.Lock()
	t.rightClickHandler = handler
	t.mu.Unlock()
	return t
}

// OnDoubleClick sets the double-click handler
func (t *SystemTray) OnDoubleClick(handler func()) *SystemTray {
	t.mu.Lock()
	t.doubleClickHandler = handler
	t.mu.Unlock()
	return t
}

// OnMouseEnter sets the mouse enter handler
func (t *SystemTray) OnMouseEnter(handler func()) *SystemTray {
	t.mu.Lock()
	t.mouseEnterHandler = handler
	t.mu.Unlock()
	return t
}

// OnMouseLeave sets the mouse leave handler
func (t *SystemTray) OnMouseLeave(handler func()) *SystemTray {
	t.mu.Lock()
	t.mouseLeaveHandler = handler
	t.mu.Unlock()
	return t
}

// Run starts the system tray
func (t *SystemTray) Run() {
	t.mu.Lock()
	if t.running || t.destroyed {
		t.mu.Unlock()
		return
	}
	t.running = true

	// Create platform implementation
	impl := newSystemTrayImpl(t)
	t.impl = impl
	t.mu.Unlock()

	// Run platform-specific implementation
	impl.run()
}

func (t *SystemTray) markReady() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.running || t.destroyed {
		return false
	}
	t.readyOnce.Do(func() { close(t.ready) })
	return true
}

func (t *SystemTray) reportError(err error) {
	if err == nil {
		return
	}
	t.mu.Lock()
	active := !t.destroyed
	t.mu.Unlock()
	if !active {
		return
	}
	select {
	case t.errors <- err:
	default:
	}
}

func (t *SystemTray) fail(err error) {
	t.mu.Lock()
	if t.destroyed {
		t.mu.Unlock()
		return
	}
	t.running = false
	t.destroyed = true
	t.mu.Unlock()

	if err != nil {
		select {
		case t.errors <- err:
		default:
		}
	}
	t.unregister()
	t.doneOnce.Do(func() { close(t.done) })
}

func (t *SystemTray) finish() {
	t.mu.Lock()
	if t.destroyed {
		t.mu.Unlock()
		t.unregister()
		t.doneOnce.Do(func() { close(t.done) })
		return
	}
	t.running = false
	t.destroyed = true
	t.mu.Unlock()
	t.unregister()
	t.doneOnce.Do(func() { close(t.done) })
}

func (t *SystemTray) unregister() {
	trayMapMu.Lock()
	delete(trayMap, t.id)
	trayMapMu.Unlock()
}

// ShowMenu shows the system tray menu
func (t *SystemTray) ShowMenu() {
	t.mu.Lock()
	menu := t.menu
	t.mu.Unlock()

	if menu != nil {
		menu.show()
	}
}

// Show shows the system tray icon
func (t *SystemTray) Show() {
	t.mu.Lock()
	impl := t.impl
	t.mu.Unlock()

	if impl != nil {
		impl.show()
	}
}

// Hide hides the system tray icon
func (t *SystemTray) Hide() {
	t.mu.Lock()
	impl := t.impl
	t.mu.Unlock()

	if impl != nil {
		impl.hide()
	}
}

// Destroy destroys the system tray icon
func (t *SystemTray) Destroy() {
	t.mu.Lock()
	if t.destroyed {
		t.mu.Unlock()
		return
	}
	t.destroyed = true
	impl := t.impl
	running := t.running
	t.running = false
	t.mu.Unlock()

	if impl != nil && running {
		impl.destroy()
	}
	t.unregister()
	t.doneOnce.Do(func() { close(t.done) })
}

// ID returns the tray ID
func (t *SystemTray) ID() uint {
	return t.id
}
