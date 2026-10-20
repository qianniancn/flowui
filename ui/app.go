package ui

import (
	"context"
	"fmt"
	"image"
	"io"
	"os"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	internalexplorer "github.com/qianniancn/flowui/internal/explorer"
	"github.com/qianniancn/flowui/internal/runtime"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/window"
)

// Program describes a complete MVU application. Init runs once for each
// native window instance and may return a startup command. Add the
// RetainModelOnClose option to a WindowSpec to keep its model and skip Init
// when the same spec is reopened.
type Program[M any, Msg any] struct {
	// Init creates the model for a newly opened native window and may return
	// the command that starts its initial asynchronous work.
	Init func() (M, Cmd[Msg])
	// Update applies one message to the model and may return a command to run
	// after the update completes.
	Update Update[M, Msg]
	// Subscriptions derives the long-lived asynchronous work for the current
	// model. A nil function means that the program has no subscriptions.
	Subscriptions Subscriptions[M, Msg]
	// View renders the current model and receives a Send function for user
	// interactions.
	View View[M, Msg]
	// WindowStateMessage maps native configuration changes to messages that
	// Update receives before the following view.
	WindowStateMessage func(WindowState) Msg
	// CloseRequestMessage turns a native close request into an MVU message.
	// When non-nil and the window has no OnWindowCloseRequest handler, FlowUI
	// aborts the request and queues the returned message; the window stays
	// open until the application calls Application.Close or RequestClose.
	// Without a handler, Application.RequestClose requests the native close and
	// reaches this hook too. Forced closes skip this hook. Repeated native
	// requests while a confirmation is pending are delivered again, so Update
	// should ignore duplicates. The message must be safe for Update to handle;
	// use a message value that Update can recognize even when no confirmation
	// is pending.
	CloseRequestMessage func() Msg
}

// NewProgram creates a Program with a fixed initial model. Use a Program
// literal when initialization must return a startup command or create fresh
// reference-backed model data each time a window opens.
func NewProgram[M any, Msg any](initial M, update Update[M, Msg], view View[M, Msg]) Program[M, Msg] {
	return Program[M, Msg]{
		Init:   func() (M, Cmd[Msg]) { return initial, nil },
		Update: update,
		View:   view,
	}
}

// Run opens a single window for program and blocks until the application exits.
// Use Application when the process needs multiple windows or explicit
// lifecycle control. Run panics when Program.CloseRequestMessage is set,
// because the caller has no Application handle for the final close; use
// NewApplication and Application.Run instead.
func Run[M any, Msg any](program Program[M, Msg], opts ...window.Option) {
	if program.CloseRequestMessage != nil {
		panic("flowui: ui.Run cannot deliver CloseRequestMessage; use ui.NewApplication")
	}
	NewApplication().Run(NewWindow("main", program, opts...))
}

func runWindowCmd[M any, Msg any](
	w *app.Window,
	appearance *windowAppearance,
	activeTheme *theme.Theme,
	language window.Language,
	translations i18n.Catalog,
	initial M,
	initialCmd Cmd[Msg],
	update Update[M, Msg],
	subscriptions Subscriptions[M, Msg],
	view View[M, Msg],
	windowStateMessage func(WindowState) Msg,
	closeRequestMessage func() Msg,
	onError func(error),
	onDestroy func(),
	onWindowState func(WindowState),
	requestClose func(),
	onClosing func(runtime.ClosingEvent) bool,
	forceClose func(),
	onExit func(M),
) error {
	ctx, err := core.NewWithCatalogStrict(w, activeTheme, language, translations)
	if err != nil {
		return err
	}
	core.SetWindowCloseRequest(ctx, requestClose)
	explorerService := internalexplorer.New(w)
	eventWindow := &platformEventWindow{
		Window:     w,
		explorer:   explorerService,
		forceClose: forceClose,
	}
	if onError == nil {
		onError = func(err error) {
			writeEffectError(os.Stderr, err)
		}
	}
	return runtime.LoopWithExit(eventWindow, initial, initialCmd, update, subscriptions, onError, onDestroy, func(config app.Config, send Send[Msg]) {
		state := core.UpdateWindowConfig(ctx, config)
		if onWindowState != nil {
			onWindowState(state)
		}
		if windowStateMessage != nil {
			send(windowStateMessage(state))
		}
	}, nativeCloseHandler(onClosing, closeRequestMessage), func(gtx layout.Context, model M, send Send[Msg]) {
		appearance.apply(ctx)
		core.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		paint.FillShape(
			gtx.Ops,
			core.ActiveTheme(ctx).Semantic().Background,
			clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Op(),
		)
		if root := view(ctx, model, send); root != nil {
			root.Layout(ctx, gtx)
		}
		core.LayoutOverlays(ctx, gtx)
		core.ApplyFrameCommands(ctx, gtx)
		core.EndFrame(ctx)
	}, onExit)
}

// nativeCloseHandler combines the explicit window close handler with the
// program's close-request message hook. An explicit handler that returns true
// owns the request. Otherwise a non-nil message aborts the native close and
// queues the message.
func nativeCloseHandler[Msg any](
	handle func(runtime.ClosingEvent) bool,
	message func() Msg,
) func(runtime.ClosingEvent, Send[Msg]) {
	return func(event runtime.ClosingEvent, send Send[Msg]) {
		if handle != nil && handle(event) {
			return
		}
		if message == nil {
			return
		}
		event.Abort()
		send(message())
	}
}

type platformEventWindow struct {
	*app.Window
	explorer   *internalexplorer.Service
	forceClose func()
}

// Event forwards native events while allowing the explorer service to observe them.
func (w *platformEventWindow) Event() event.Event {
	value := w.Window.Event()
	w.explorer.ListenEvents(value)
	return value
}

// ForceClose requests a native close without invoking the window close handler.
func (w *platformEventWindow) ForceClose() {
	if w.forceClose != nil {
		w.forceClose()
		return
	}
	if w.Window != nil {
		w.Window.Perform(system.ActionClose)
	}
}

// DecorateEffectContext attaches the window's explorer service to the effect
// context, so every command and subscription can resolve it from the context
// the runtime hands them.
func (w *platformEventWindow) DecorateEffectContext(ctx context.Context) context.Context {
	return internalexplorer.WithService(ctx, w.explorer)
}

func writeEffectError(w io.Writer, err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(w, "flowui:", err)
	var stack []byte
	switch value := err.(type) {
	case *EffectError:
		stack = value.Stack
	case *RuntimePanicError:
		stack = value.Stack
	}
	if len(stack) > 0 {
		_, _ = w.Write(stack)
	}
}
