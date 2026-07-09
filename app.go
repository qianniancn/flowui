package flowui

import (
	"fmt"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	flowruntime "github.com/qianniancn/FlowUI/runtime"
)

// Run opens a Gio window and runs an MVU application in it.
func Run[M any, Msg any](initial M, update Update[M, Msg], view View[M, Msg], opts ...Option) {
	RunCmd(initial, func(m *M, msg Msg) Cmd[Msg] {
		update(m, msg)
		return nil
	}, view, opts...)
}

// RunCmd opens a Gio window and runs an MVU application with commands.
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
	ctx := newContextWithThemeAndLanguage(w, theme, language)
	return flowruntime.Loop(w, initial, func(model *M, msg Msg) func(func(Msg)) {
		cmd := update(model, msg)
		if cmd == nil {
			return nil
		}
		return func(send func(Msg)) {
			cmd(Send[Msg](send))
		}
	}, func(gtx layout.Context, model M, send func(Msg)) {
		ctx.beginFrame()
		if root := view(ctx, model, Send[Msg](send)); root != nil {
			root.Layout(ctx, gtx)
		}
		ctx.applyFrameCommands(gtx)
		ctx.endFrame()
	})
}

func startCmd[Msg any](cmd Cmd[Msg], send Send[Msg]) {
	if cmd == nil {
		return
	}
	flowruntime.StartCmd(func(sendFn func(Msg)) {
		cmd(Send[Msg](sendFn))
	}, func(msg Msg) {
		send(msg)
	})
}
