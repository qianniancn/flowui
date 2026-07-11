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
	RunCmd(initial, func(m *M, msg Msg) Cmd[Msg] {
		update(m, msg)
		return nil
	}, view, opts...)
}

// RunCmd opens a Gio window and runs an MVU application with commands. Update
// runs serially on the event loop, while each returned Cmd runs concurrently;
// see Cmd for the required capture and message-passing rules.
func RunCmd[M any, Msg any](initial M, update UpdateCmd[M, Msg], view View[M, Msg], opts ...Option) {
	run(initial, update, nil, view, opts...)
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
	run(initial, update, subscriptions, view, opts...)
}

func run[M any, Msg any](initial M, update UpdateCmd[M, Msg], subscriptions Subscriptions[M, Msg], view View[M, Msg], opts ...Option) {
	cfg := newRunOptions(opts)
	w := new(app.Window)
	w.Option(cfg.window...)

	go func() {
		err := runWindowCmd(w, cfg.newTheme(), cfg.language, initial, update, subscriptions, view, cfg.errorHandler)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func runWindowCmd[M any, Msg any](
	w *app.Window,
	theme *Theme,
	language Language,
	initial M,
	update UpdateCmd[M, Msg],
	subscriptions Subscriptions[M, Msg],
	view View[M, Msg],
	onError func(error),
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
	return runtime.Loop(w, initial, func(model *M, msg Msg) runtime.Cmd[Msg] {
		cmd := update(model, msg)
		if cmd == nil {
			return nil
		}
		return func(effectCtx context.Context, send func(Msg)) error {
			return cmd(effectCtx, Send[Msg](send))
		}
	}, runtimeSubscriptions, onError, func(gtx layout.Context, model M, send func(Msg)) {
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

func writeEffectError(w io.Writer, err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(w, "flowui:", err)
	if effectErr, ok := err.(*EffectError); ok && len(effectErr.Stack) > 0 {
		_, _ = w.Write(effectErr.Stack)
	}
}
