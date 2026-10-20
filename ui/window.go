package ui

import (
	"fmt"
	"os"
	"sync"

	"gioui.org/app"
	"gioui.org/io/system"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/runtime"
	"github.com/qianniancn/flowui/theme"
	flowwindow "github.com/qianniancn/flowui/window"
)

type windowRun func(
	*app.Window,
	*windowAppearance,
	func(),
	func(WindowState),
	func(),
	func(runtime.ClosingEvent) bool,
	func(),
) error

// WindowSpec describes one independent FlowUI window.
type WindowSpec struct {
	key                 string
	options             []app.Option
	initialActions      system.Action
	run                 windowRun
	onError             func(error)
	closeRequestHandler func() flowwindow.CloseDecision
}

type retainedWindowModel[M any, Msg any] struct {
	mu          sync.Mutex
	initialized bool
	model       M
}

func (state *retainedWindowModel[M, Msg]) start(initialize func() (M, Cmd[Msg])) (M, Cmd[Msg]) {
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.initialized {
		return state.model, nil
	}
	model, cmd := initialize()
	state.model = model
	state.initialized = true
	return model, cmd
}

func (state *retainedWindowModel[M, Msg]) save(model M) {
	state.mu.Lock()
	state.model = model
	state.initialized = true
	state.mu.Unlock()
}

type windowAppearance struct {
	mu          sync.Mutex
	theme       *theme.Theme
	language    flowwindow.Language
	languageSet bool
}

func (appearance *windowAppearance) setTheme(value theme.Theme) {
	theme.DetachMaterial(&value)
	theme.SyncMaterialTheme(&value)
	appearance.mu.Lock()
	appearance.theme = &value
	appearance.mu.Unlock()
}

func (appearance *windowAppearance) setLanguage(value flowwindow.Language) {
	appearance.mu.Lock()
	appearance.language = value
	appearance.languageSet = true
	appearance.mu.Unlock()
}

func (appearance *windowAppearance) apply(ctx *core.Context) {
	if appearance == nil {
		return
	}
	appearance.mu.Lock()
	activeTheme := appearance.theme
	language := appearance.language
	languageSet := appearance.languageSet
	appearance.theme = nil
	appearance.languageSet = false
	appearance.mu.Unlock()
	if activeTheme != nil {
		core.ReplaceTheme(ctx, *activeTheme)
	}
	if languageSet {
		core.ReplaceLanguage(ctx, language)
	}
}

// Key returns the identity used to deduplicate and control the window.
func (w WindowSpec) Key() string {
	return w.key
}

// NewWindow creates a window from a complete MVU Program.
func NewWindow[M any, Msg any](key string, program Program[M, Msg], opts ...flowwindow.Option) WindowSpec {
	if program.Init == nil {
		panic("flowui: nil program init")
	}
	return newWindowSpec(
		key,
		program.Init,
		program.Update,
		program.Subscriptions,
		program.View,
		program.WindowStateMessage,
		program.CloseRequestMessage,
		opts,
	)
}

func newWindowSpec[M any, Msg any](
	key string,
	initialize func() (M, Cmd[Msg]),
	update Update[M, Msg],
	subscriptions Subscriptions[M, Msg],
	view View[M, Msg],
	windowStateMessage func(WindowState) Msg,
	closeRequestMessage func() Msg,
	opts []flowwindow.Option,
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
	cfg := flowwindow.Resolve(opts...)
	var retained *retainedWindowModel[M, Msg]
	if cfg.RetainModel {
		retained = new(retainedWindowModel[M, Msg])
	}
	if !cfg.ThemeSet {
		panic("flowui: window requires an explicit theme; use window.Theme(...)")
	}
	if err := theme.ValidateVisualProvider(cfg.InitialTheme); err != nil {
		panic(err)
	}
	return WindowSpec{
		key:                 key,
		options:             append([]app.Option(nil), cfg.Native...),
		initialActions:      cfg.InitialActions,
		onError:             cfg.ErrorHandler,
		closeRequestHandler: cfg.CloseRequestHandler,
		run: func(
			window *app.Window,
			appearance *windowAppearance,
			onDestroy func(),
			onWindowState func(WindowState),
			requestClose func(),
			onClosing func(runtime.ClosingEvent) bool,
			forceClose func(),
		) error {
			var initial M
			var initialCmd Cmd[Msg]
			var onExit func(M)
			if retained != nil {
				initial, initialCmd = retained.start(initialize)
				onExit = retained.save
			} else {
				initial, initialCmd = initialize()
			}
			activeTheme := cfg.InitialTheme
			for _, apply := range cfg.ThemeOps {
				apply(&activeTheme)
			}
			if err := theme.ValidateVisualProvider(activeTheme); err != nil {
				return err
			}
			if theme.MaterialOf(&activeTheme) == nil {
				theme.SyncMaterialTheme(&activeTheme)
			}
			return runWindowCmd(
				window,
				appearance,
				&activeTheme,
				cfg.Language,
				cfg.Translations,
				initial,
				initialCmd,
				update,
				subscriptions,
				view,
				windowStateMessage,
				closeRequestMessage,
				cfg.ErrorHandler,
				onDestroy,
				onWindowState,
				requestClose,
				onClosing,
				forceClose,
				onExit,
			)
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

// Application owns the windows and process lifecycle for one desktop process.
type Application struct {
	windows windowSet
	exit    func(int)
}

// NewApplication creates an application that can open windows dynamically.
func NewApplication() *Application {
	return &Application{exit: os.Exit}
}

// SetKeepAlive controls whether the application remains running after its
// last window closes. Enable it before Run for tray and background apps that
// may open a new window later.
func (a *Application) SetKeepAlive(enabled bool) {
	if a == nil {
		return
	}
	a.windows.setKeepAlive(enabled)
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
	appearance := new(windowAppearance)
	existing, added := a.windows.addWithCloseRequest(spec.key, window, appearance, spec.closeRequestHandler)
	if existing != nil {
		existing.Perform(system.ActionRaise)
		return false
	}
	if !added {
		return false
	}
	if spec.initialActions != 0 {
		window.Perform(spec.initialActions)
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
		err := spec.run(window, appearance, deactivate, func(state WindowState) {
			a.windows.update(spec.key, window, state)
		}, func() {
			a.RequestClose(spec.key)
		}, func(event runtime.ClosingEvent) bool {
			return a.handleClosing(spec.key, window, event)
		}, func() {
			a.forceClose(window)
		})
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

// Configure applies native window options to an active window.
func (a *Application) Configure(key string, options ...flowwindow.Option) bool {
	if a == nil {
		return false
	}
	window := a.windows.get(key)
	if window == nil {
		return false
	}
	cfg := flowwindow.Resolve(options...)
	window.Option(cfg.Native...)
	return true
}

// SetTheme replaces the theme of an active window on its event loop. It
// returns false when key does not identify an active window.
func (a *Application) SetTheme(key string, value theme.Theme) bool {
	if a == nil {
		return false
	}
	if theme.ValidateVisualProvider(value) != nil {
		return false
	}
	window, appearance := a.windows.appearance(key)
	if window == nil || appearance == nil {
		return false
	}
	appearance.setTheme(value)
	window.Invalidate()
	return true
}

// SetLanguage replaces the language of an active window on its event loop. It
// returns false when key does not identify an active window.
func (a *Application) SetLanguage(key string, value flowwindow.Language) bool {
	if a == nil {
		return false
	}
	window, appearance := a.windows.appearance(key)
	if window == nil || appearance == nil {
		return false
	}
	appearance.setLanguage(value)
	window.Invalidate()
	return true
}

// Perform requests a native action for an active window.
func (a *Application) Perform(key string, action flowwindow.Action) bool {
	if a == nil || action == 0 {
		return false
	}
	window := a.windows.get(key)
	if window == nil {
		return false
	}
	window.Perform(action)
	return true
}

// WindowState returns the latest native state reported for an active window.
func (a *Application) WindowState(key string) (WindowState, bool) {
	if a == nil {
		return WindowState{}, false
	}
	return a.windows.state(key)
}

// RequestClose asks an active window's configured close-request handler what to
// do. It returns false when key does not identify an active window. A true
// result means the request was delivered; the handler may cancel it.
//
// CloseKeepAlive closes the native window and enables application keep-alive.
// Reopen it with Open, and use RetainModelOnClose when the same MVU model should
// resume. Native close commands use the same handler. When no handler is
// configured, RequestClose requests the native close, so a
// Program.CloseRequestMessage hook applies to the resulting closing event.
func (a *Application) RequestClose(key string) bool {
	if a == nil {
		return false
	}
	window, handler := a.windows.closeRequest(key)
	if window == nil {
		return false
	}
	if handler == nil {
		window.Perform(system.ActionClose)
		return true
	}
	decision := handler()
	switch decision {
	case flowwindow.CloseProceed:
		a.windows.markCloseIntent(window, closeIntentApproved)
		window.Perform(system.ActionClose)
	case flowwindow.CloseCancel:
		// The handler intentionally left the window open.
	case flowwindow.CloseKeepAlive:
		a.windows.setKeepAlive(true)
		a.windows.markCloseIntent(window, closeIntentApproved)
		window.Perform(system.ActionClose)
	default:
		// Unknown decisions fail closed so an invalid handler result cannot
		// unexpectedly destroy a window.
	}
	return true
}

// Close force-closes one window without calling its close-request handler.
func (a *Application) Close(key string) bool {
	if a == nil {
		return false
	}
	window := a.windows.get(key)
	if window == nil {
		return false
	}
	a.forceClose(window)
	return true
}

// CloseAll force-closes every active window without calling close-request
// handlers.
func (a *Application) CloseAll() {
	if a == nil {
		return
	}
	for _, window := range a.windows.snapshot() {
		a.forceClose(window)
	}
}

// Quit disables keep-alive and requests that every active window close. If no
// windows are open, the application exits immediately.
func (a *Application) Quit() {
	if a == nil {
		return
	}
	a.windows.quit()
	a.CloseAll()
}

// handleClosing applies a native close request. It returns true when a close
// intent or an explicit close-request handler owns the request, so the caller
// must not fall back to the close-request message hook. A false result means
// the window has no handler and the request is still unhandled.
func (a *Application) handleClosing(key string, window *app.Window, event runtime.ClosingEvent) bool {
	if a == nil || event == nil {
		return true
	}
	if a.windows.shuttingDown() {
		return true
	}
	if _, ok := a.windows.consumeCloseIntent(window); ok {
		return true
	}
	_, handler := a.windows.closeRequest(key)
	if handler == nil {
		return false
	}
	switch handler() {
	case flowwindow.CloseProceed:
	case flowwindow.CloseKeepAlive:
		a.windows.setKeepAlive(true)
	case flowwindow.CloseCancel:
		event.Abort()
	default:
		event.Abort()
	}
	return true
}

func (a *Application) forceClose(window *app.Window) {
	if a == nil || window == nil {
		return
	}
	a.windows.markCloseIntent(window, closeIntentForced)
	window.Perform(system.ActionClose)
}

type windowSet struct {
	mu            sync.Mutex
	active        map[string]*app.Window
	appearances   map[string]*windowAppearance
	states        map[string]WindowState
	closeRequests map[string]func() flowwindow.CloseDecision
	done          chan int
	running       bool
	starting      bool
	quitting      bool
	closing       bool
	failed        bool
	keepAlive     bool
	loops         int
	closeIntents  map[*app.Window]closeIntent
}

type closeIntent uint8

const (
	closeIntentApproved closeIntent = iota + 1
	closeIntentForced
)

func (s *windowSet) setKeepAlive(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.quitting || s.closing {
		return
	}
	s.keepAlive = enabled
	s.finishLocked()
}

func (s *windowSet) quit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return
	}
	s.quitting = true
	s.keepAlive = false
	s.finishLocked()
}

func (s *windowSet) begin() <-chan int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		panic("flowui: application already running")
	}
	s.active = make(map[string]*app.Window)
	s.appearances = make(map[string]*windowAppearance)
	s.states = make(map[string]WindowState)
	s.closeRequests = make(map[string]func() flowwindow.CloseDecision)
	s.closeIntents = make(map[*app.Window]closeIntent)
	s.done = make(chan int, 1)
	s.running = true
	s.starting = true
	s.quitting = false
	s.closing = false
	s.failed = false
	s.loops = 0
	return s.done
}

func (s *windowSet) addWithCloseRequest(key string, window *app.Window, appearance *windowAppearance, handler func() flowwindow.CloseDecision) (*app.Window, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running || s.quitting || s.closing {
		return nil, false
	}
	if existing := s.active[key]; existing != nil {
		return existing, false
	}
	s.active[key] = window
	s.appearances[key] = appearance
	s.closeRequests[key] = handler
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
	delete(s.appearances, key)
	delete(s.states, key)
	delete(s.closeRequests, key)
	delete(s.closeIntents, window)
}

func (s *windowSet) update(key string, window *app.Window, state WindowState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active[key] == window {
		s.states[key] = state
	}
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
	if !s.running || s.starting || s.closing || s.keepAlive || s.loops != 0 {
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

func (s *windowSet) appearance(key string) (*app.Window, *windowAppearance) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active[key], s.appearances[key]
}

func (s *windowSet) state(key string) (WindowState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.states[key]
	return state, ok
}

func (s *windowSet) shuttingDown() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.quitting || s.closing
}

func (s *windowSet) closeRequest(key string) (*app.Window, func() flowwindow.CloseDecision) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.quitting || s.closing {
		return nil, nil
	}
	return s.active[key], s.closeRequests[key]
}

func (s *windowSet) markCloseIntent(window *app.Window, intent closeIntent) {
	if window == nil || intent == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closeIntents == nil {
		s.closeIntents = make(map[*app.Window]closeIntent)
	}
	s.closeIntents[window] = intent
}

func (s *windowSet) consumeCloseIntent(window *app.Window) (closeIntent, bool) {
	if window == nil {
		return 0, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	intent, ok := s.closeIntents[window]
	if ok {
		delete(s.closeIntents, window)
	}
	return intent, ok
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
