package ui

import (
	"fmt"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
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
	cfg := newRunOptions(opts)
	w := new(app.Window)
	w.Option(cfg.window...)

	go func() {
		err := runWindowCmd(w, cfg.newTheme(), cfg.language, initial, update, view)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func runWindowCmd[M any, Msg any](w *app.Window, theme *Theme, language Language, initial M, update UpdateCmd[M, Msg], view View[M, Msg]) error {
	ctx := frame.New(w, theme, language)
	return runtime.Loop(w, initial, func(model *M, msg Msg) func(func(Msg)) {
		cmd := update(model, msg)
		if cmd == nil {
			return nil
		}
		return func(send func(Msg)) {
			cmd(Send[Msg](send))
		}
	}, func(gtx layout.Context, model M, send func(Msg)) {
		frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
		if root := view(ctx, model, Send[Msg](send)); root != nil {
			root.Layout(ctx, gtx)
		}
		frame.LayoutOverlays(ctx, gtx)
		frame.ApplyFrameCommands(ctx, gtx)
		frame.EndFrame(ctx)
	})
}
