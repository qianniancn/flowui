package ui

import (
	"context"
	"fmt"
	"image"
	"io"
	"os"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	internalexplorer "github.com/qianniancn/flowui/internal/explorer"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/runtime"
)

// Program describes a complete MVU application. Init runs once for each
// native window instance and may return a startup command. Add the
// RetainModelOnClose option to a WindowSpec to keep its model and skip Init
// when the same spec is reopened.
type Program[M any, Msg any] struct {
	Init          func() (M, Cmd[Msg])
	Update        Update[M, Msg]
	Subscriptions Subscriptions[M, Msg]
	View          View[M, Msg]
	// WindowStateMessage maps native configuration changes to messages that
	// Update receives before the following view.
	WindowStateMessage func(WindowState) Msg
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

// Run opens a window and runs a complete MVU Program.
func Run[M any, Msg any](program Program[M, Msg], opts ...Option) {
	NewApplication().Run(NewWindow("main", program, opts...))
}

func runWindowCmd[M any, Msg any](
	w *app.Window,
	appearance *windowAppearance,
	theme *Theme,
	language Language,
	initial M,
	initialCmd Cmd[Msg],
	update Update[M, Msg],
	subscriptions Subscriptions[M, Msg],
	view View[M, Msg],
	windowStateMessage func(WindowState) Msg,
	onError func(error),
	onDestroy func(),
	onWindowState func(WindowState),
	requestClose func(),
	onExit func(M),
) error {
	ctx := frame.New(w, theme, language)
	frame.SetWindowCloseRequest(ctx, requestClose)
	explorerService := internalexplorer.New(w)
	eventWindow := &platformEventWindow{Window: w, explorer: explorerService}
	if onError == nil {
		onError = func(err error) {
			writeEffectError(os.Stderr, err)
		}
	}
	var runtimeSubscriptions runtime.Subscriptions[M, Msg]
	if subscriptions != nil {
		runtimeSubscriptions = func(model M) []runtime.Subscription[Msg] {
			requested := subscriptions(model)
			result := make([]runtime.Subscription[Msg], len(requested))
			for index, subscription := range requested {
				result[index] = runtime.Subscription[Msg]{
					Key: subscription.key,
					Run: func(effectCtx context.Context, send func(Msg)) error {
						effectCtx = internalexplorer.WithService(effectCtx, explorerService)
						return subscription.run(effectCtx, Send[Msg](send))
					},
				}
			}
			return result
		}
	}
	return runtime.LoopWithExit(eventWindow, initial, runtimeCmd(initialCmd, explorerService), func(model *M, msg Msg) runtime.Cmd[Msg] {
		return runtimeCmd(update(model, msg), explorerService)
	}, runtimeSubscriptions, onError, onDestroy, func(config app.Config, send func(Msg)) {
		state := frame.UpdateWindowConfig(ctx, config)
		if onWindowState != nil {
			onWindowState(state)
		}
		if windowStateMessage != nil {
			send(windowStateMessage(state))
		}
	}, func(gtx layout.Context, model M, send func(Msg)) {
		appearance.apply(ctx)
		frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		paint.FillShape(
			gtx.Ops,
			frame.ActiveTheme(ctx).Palette.Background,
			clip.Rect(image.Rectangle{Max: gtx.Constraints.Max}).Op(),
		)
		if root := view(ctx, model, Send[Msg](send)); root != nil {
			root.Layout(ctx, gtx)
		}
		frame.LayoutOverlays(ctx, gtx)
		frame.ApplyFrameCommands(ctx, gtx)
		frame.EndFrame(ctx)
	}, onExit)
}

type platformEventWindow struct {
	*app.Window
	explorer *internalexplorer.Service
}

func (w *platformEventWindow) Event() event.Event {
	value := w.Window.Event()
	w.explorer.ListenEvents(value)
	return value
}

func runtimeCmd[Msg any](cmd Cmd[Msg], explorerService *internalexplorer.Service) runtime.Cmd[Msg] {
	if cmd == nil {
		return nil
	}
	return func(effectCtx context.Context, send func(Msg)) error {
		effectCtx = internalexplorer.WithService(effectCtx, explorerService)
		return cmd(effectCtx, Send[Msg](send))
	}
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
