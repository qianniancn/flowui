package uitest

import (
	"context"
	"image"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/runtime"
	"github.com/qianniancn/flowui/ui"
)

// AppConfig configures an AppHarness. Update is required; Init is optional.
type AppConfig[M any, Msg any] struct {
	Initial M
	Init    ui.Cmd[Msg]
	Update  ui.Update[M, Msg]
}

// AppHarness drives FlowUI's production MVU runtime without opening a window.
// Close must be called to cancel and join outstanding commands.
type AppHarness[M any, Msg any] struct {
	window appHarnessWindow

	mu      sync.RWMutex
	model   M
	send    ui.Send[Msg]
	errors  []error
	runErr  error
	stopped chan struct{}
	close   sync.Once
}

// NewApp creates a harness for an Update function.
func NewApp[M any, Msg any](initial M, update ui.Update[M, Msg]) *AppHarness[M, Msg] {
	if update == nil {
		panic("flowui/uitest: nil app update")
	}
	return NewAppWithConfig(AppConfig[M, Msg]{
		Initial: initial,
		Update:  update,
	})
}

// NewAppWithConfig creates a configured application harness.
func NewAppWithConfig[M any, Msg any](config AppConfig[M, Msg]) *AppHarness[M, Msg] {
	if config.Update == nil {
		panic("flowui/uitest: nil app update")
	}
	harness := &AppHarness[M, Msg]{
		window: appHarnessWindow{
			events:      make(chan event.Event),
			invalidated: make(chan struct{}, 16),
		},
		model:   config.Initial,
		stopped: make(chan struct{}),
	}
	go func() {
		err := runtime.Loop(
			&harness.window,
			config.Initial,
			appRuntimeCmd(config.Init),
			func(model *M, msg Msg) runtime.Cmd[Msg] {
				return appRuntimeCmd(config.Update(model, msg))
			},
			nil,
			func(err error) {
				harness.mu.Lock()
				harness.errors = append(harness.errors, err)
				harness.mu.Unlock()
			},
			nil,
			nil,
			func(_ layout.Context, model M, send func(Msg)) {
				harness.mu.Lock()
				harness.model = model
				harness.send = ui.Send[Msg](send)
				harness.mu.Unlock()
			},
		)
		harness.mu.Lock()
		harness.runErr = err
		harness.send = nil
		harness.mu.Unlock()
		close(harness.stopped)
	}()
	harness.Frame()
	return harness
}

// Send queues a message for the next Frame.
func (h *AppHarness[M, Msg]) Send(msg Msg) {
	h.mu.RLock()
	send := h.send
	h.mu.RUnlock()
	if send == nil {
		panic("flowui/uitest: app harness is closed")
	}
	send(msg)
}

// Frame processes queued messages, starts their commands, and returns the
// model snapshot passed to the view phase.
func (h *AppHarness[M, Msg]) Frame() M {
	h.drainInvalidations()
	rendered := make(chan struct{})
	e := app.FrameEvent{
		Now:  time.Now(),
		Size: image.Pt(1, 1),
		Frame: func(*op.Ops) {
			close(rendered)
		},
	}
	select {
	case h.window.events <- e:
	case <-h.stopped:
		panic("flowui/uitest: app harness is closed")
	}
	select {
	case <-rendered:
		return h.Model()
	case <-h.stopped:
		panic("flowui/uitest: app harness stopped before completing a frame")
	}
}

// Wait reports whether the runtime requested another frame before timeout.
func (h *AppHarness[M, Msg]) Wait(timeout time.Duration) bool {
	if timeout < 0 {
		panic("flowui/uitest: negative app wait")
	}
	select {
	case <-h.stopped:
		return false
	default:
	}
	if timeout == 0 {
		select {
		case <-h.window.invalidated:
			return true
		default:
			return false
		}
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-h.window.invalidated:
		return true
	case <-h.stopped:
		return false
	case <-timer.C:
		return false
	}
}

// Model returns the most recent model snapshot produced by Frame.
func (h *AppHarness[M, Msg]) Model() M {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.model
}

// Errors returns effect errors delivered on completed frames.
func (h *AppHarness[M, Msg]) Errors() []error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]error(nil), h.errors...)
}

// Close cancels outstanding commands and waits for runtime cleanup.
func (h *AppHarness[M, Msg]) Close() error {
	h.close.Do(func() {
		select {
		case h.window.events <- app.DestroyEvent{}:
		case <-h.stopped:
		}
	})
	<-h.stopped
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.runErr
}

func (h *AppHarness[M, Msg]) drainInvalidations() {
	for {
		select {
		case <-h.window.invalidated:
		default:
			return
		}
	}
}

func appRuntimeCmd[Msg any](cmd ui.Cmd[Msg]) runtime.Cmd[Msg] {
	if cmd == nil {
		return nil
	}
	return func(ctx context.Context, send func(Msg)) error {
		return cmd(ctx, ui.Send[Msg](send))
	}
}

type appHarnessWindow struct {
	events      chan event.Event
	invalidated chan struct{}
}

func (w *appHarnessWindow) Event() event.Event {
	return <-w.events
}

func (w *appHarnessWindow) Invalidate() {
	select {
	case w.invalidated <- struct{}{}:
	default:
	}
}
