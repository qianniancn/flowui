package ui

import (
	"fmt"
	"os"
	"sync"

	"gioui.org/app"
	"gioui.org/io/system"
)

// WindowSpec describes one independent FlowUI window.
type WindowSpec struct {
	key     string
	options []app.Option
	run     func(*app.Window, func()) error
	onError func(error)
}

// Key returns the identity used to deduplicate and control the window.
func (w WindowSpec) Key() string {
	return w.key
}

// NewWindow creates a window using a synchronous Update function.
func NewWindow[M any, Msg any](key string, initial M, update Update[M, Msg], view View[M, Msg], opts ...Option) WindowSpec {
	if update == nil {
		panic("flowui: nil window update")
	}
	return NewWindowCmd(key, initial, func(model *M, msg Msg) Cmd[Msg] {
		update(model, msg)
		return nil
	}, view, opts...)
}

// NewWindowCmd creates a window whose Update function may return commands.
func NewWindowCmd[M any, Msg any](key string, initial M, update UpdateCmd[M, Msg], view View[M, Msg], opts ...Option) WindowSpec {
	return newWindowSpec(key, initial, update, nil, view, opts)
}

// NewWindowWithSubscriptions creates a window with commands and subscriptions.
func NewWindowWithSubscriptions[M any, Msg any](
	key string,
	initial M,
	update UpdateCmd[M, Msg],
	subscriptions Subscriptions[M, Msg],
	view View[M, Msg],
	opts ...Option,
) WindowSpec {
	return newWindowSpec(key, initial, update, subscriptions, view, opts)
}

func newWindowSpec[M any, Msg any](
	key string,
	initial M,
	update UpdateCmd[M, Msg],
	subscriptions Subscriptions[M, Msg],
	view View[M, Msg],
	opts []Option,
) WindowSpec {
	if key == "" {
		panic("flowui: empty window key")
	}
	if update == nil {
		panic("flowui: nil window update")
	}
	if view == nil {
		panic("flowui: nil window view")
	}
	cfg := newRunOptions(opts)
	return WindowSpec{
		key:     key,
		options: append([]app.Option(nil), cfg.window...),
		onError: cfg.errorHandler,
		run: func(window *app.Window, onDestroy func()) error {
			return runWindowCmd(window, cfg.newTheme(), cfg.language, initial, update, subscriptions, view, cfg.errorHandler, onDestroy)
		},
	}
}

func (w WindowSpec) validate() {
	if w.key == "" {
		panic("flowui: empty window key")
	}
	if w.run == nil {
		panic(fmt.Sprintf("flowui: invalid window %q", w.key))
	}
}

func (w WindowSpec) report(err error) {
	if err == nil {
		return
	}
	if w.onError != nil {
		w.onError(err)
		return
	}
	writeEffectError(os.Stderr, err)
}

// Application owns the windows in one desktop process.
type Application struct {
	windows windowSet
	exit    func(int)
}

// NewApplication creates an application that can open windows dynamically.
func NewApplication() *Application {
	return &Application{exit: os.Exit}
}

// Run opens the initial windows and hands the main thread to Gio. It must be
// called once from main. Multi-window operation is supported on desktop
// platforms.
func (a *Application) Run(initial ...WindowSpec) {
	if a == nil {
		panic("flowui: nil application")
	}
	if len(initial) == 0 {
		panic("flowui: no initial windows")
	}
	seen := make(map[string]struct{}, len(initial))
	for _, window := range initial {
		window.validate()
		if _, duplicate := seen[window.key]; duplicate {
			panic(fmt.Sprintf("flowui: duplicate initial window key %q", window.key))
		}
		seen[window.key] = struct{}{}
	}
	done := a.windows.begin()
	for _, window := range initial {
		if !a.Open(window) {
			panic(fmt.Sprintf("flowui: failed to open initial window %q", window.key))
		}
	}
	a.windows.finishStarting()
	go func() {
		code := <-done
		exit := a.exit
		if exit == nil {
			exit = os.Exit
		}
		exit(code)
	}()
	app.Main()
}

// Open opens a window or raises the existing window with the same key. It
// returns true only when a new window was started.
func (a *Application) Open(spec WindowSpec) bool {
	if a == nil {
		return false
	}
	spec.validate()
	window := new(app.Window)
	window.Option(spec.options...)
	existing, added := a.windows.add(spec.key, window)
	if existing != nil {
		existing.Perform(system.ActionRaise)
		return false
	}
	if !added {
		return false
	}
	go func() {
		deactivated := false
		deactivate := func() {
			if deactivated {
				return
			}
			deactivated = true
			a.windows.deactivate(spec.key, window)
		}
		err := spec.run(window, deactivate)
		deactivate()
		spec.report(err)
		a.windows.complete(err != nil)
	}()
	return true
}

// IsOpen reports whether a window key is active.
func (a *Application) IsOpen(key string) bool {
	if a == nil {
		return false
	}
	return a.windows.get(key) != nil
}

// Close requests that one window close.
func (a *Application) Close(key string) bool {
	if a == nil {
		return false
	}
	window := a.windows.get(key)
	if window == nil {
		return false
	}
	window.Perform(system.ActionClose)
	return true
}

// CloseAll requests that every active window close.
func (a *Application) CloseAll() {
	if a == nil {
		return
	}
	for _, window := range a.windows.snapshot() {
		window.Perform(system.ActionClose)
	}
}

// RunWindows runs a fixed set of independent FlowUI windows.
func RunWindows(windows ...WindowSpec) {
	NewApplication().Run(windows...)
}

type windowSet struct {
	mu       sync.Mutex
	active   map[string]*app.Window
	done     chan int
	running  bool
	starting bool
	closing  bool
	failed   bool
	loops    int
}

func (s *windowSet) begin() <-chan int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		panic("flowui: application already running")
	}
	s.active = make(map[string]*app.Window)
	s.done = make(chan int, 1)
	s.running = true
	s.starting = true
	s.closing = false
	s.failed = false
	s.loops = 0
	return s.done
}

func (s *windowSet) add(key string, window *app.Window) (*app.Window, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running || s.closing {
		return nil, false
	}
	if existing := s.active[key]; existing != nil {
		return existing, false
	}
	s.active[key] = window
	s.loops++
	return nil, true
}

func (s *windowSet) finishStarting() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.starting = false
	s.finishLocked()
}

func (s *windowSet) deactivate(key string, window *app.Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active[key] != window {
		return
	}
	delete(s.active, key)
}

func (s *windowSet) complete(failed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loops <= 0 {
		panic("flowui: completed window was not running")
	}
	s.loops--
	s.failed = s.failed || failed
	s.finishLocked()
}

func (s *windowSet) finishLocked() {
	if s.starting || s.closing || s.loops != 0 {
		return
	}
	s.closing = true
	code := 0
	if s.failed {
		code = 1
	}
	s.done <- code
}

func (s *windowSet) get(key string) *app.Window {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active[key]
}

func (s *windowSet) snapshot() []*app.Window {
	s.mu.Lock()
	defer s.mu.Unlock()
	windows := make([]*app.Window, 0, len(s.active))
	for _, window := range s.active {
		windows = append(windows, window)
	}
	return windows
}
