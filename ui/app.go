package ui

import (
	"context"
	"fmt"
	"image"
	"io"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/runtime"
)

// Run opens a Gio window and runs an MVU application in it.
func Run[M any, Msg any](initial M, update Update[M, Msg], view View[M, Msg], opts ...Option) {
	RunWindows(NewWindow("main", func() M { return initial }, update, view, opts...))
}

// RunCmd opens a Gio window and runs an MVU application with commands. Update
// runs serially on the event loop, while each returned Cmd runs concurrently;
// see Cmd for the required capture and message-passing rules.
func RunCmd[M any, Msg any](initial M, update UpdateCmd[M, Msg], view View[M, Msg], opts ...Option) {
	RunWindows(NewWindowCmd("main", func() M { return initial }, update, view, opts...))
}

// RunWithSubscriptions opens a Gio window and runs an MVU application with
// commands and a model-derived set of long-lived subscriptions.
func RunWithSubscriptions[M any, Msg any](
	initial M,
	update UpdateCmd[M, Msg],
	subscriptions Subscriptions[M, Msg],
	view View[M, Msg],
	opts ...Option,
) {
	RunWindows(NewWindowWithSubscriptions("main", func() M { return initial }, update, subscriptions, view, opts...))
}

// Program describes a complete MVU application. Init runs once for each
// window instance and may return a startup command.
type Program[M any, Msg any] struct {
	Init          func() (M, Cmd[Msg])
	Update        UpdateCmd[M, Msg]
	Subscriptions Subscriptions[M, Msg]
	View          View[M, Msg]
	// WindowStateMessage maps native configuration changes to messages that
	// Update receives before the following view.
	WindowStateMessage func(WindowState) Msg
}

// RunProgram opens a Gio window and runs a complete MVU Program.
func RunProgram[M any, Msg any](program Program[M, Msg], opts ...Option) {
	RunWindows(NewProgramWindow("main", program, opts...))
}

func runWindowCmd[M any, Msg any](
	w *app.Window,
	appearance *windowAppearance,
	theme *Theme,
	language Language,
	initial M,
	initialCmd Cmd[Msg],
	update UpdateCmd[M, Msg],
	subscriptions Subscriptions[M, Msg],
	view View[M, Msg],
	windowStateMessage func(WindowState) Msg,
	onError func(error),
	onDestroy func(),
	onWindowState func(WindowState),
) error {
	ctx := frame.New(w, theme, language)
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
				subscription := subscription
				result[index] = runtime.Subscription[Msg]{
					Key: subscription.key,
					Run: func(effectCtx context.Context, send func(Msg)) error {
						return subscription.run(effectCtx, Send[Msg](send))
					},
				}
			}
			return result
		}
	}
	return runtime.Loop(w, initial, runtimeCmd(initialCmd), func(model *M, msg Msg) runtime.Cmd[Msg] {
		return runtimeCmd(update(model, msg))
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
	})
}

func runtimeCmd[Msg any](cmd Cmd[Msg]) runtime.Cmd[Msg] {
	if cmd == nil {
		return nil
	}
	return func(effectCtx context.Context, send func(Msg)) error {
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
