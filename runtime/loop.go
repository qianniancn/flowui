package runtime

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
)

type Update[M any, Msg any] func(*M, Msg) func(func(Msg))

type Frame[M any, Msg any] func(layout.Context, M, func(Msg))

func Loop[M any, Msg any](w *app.Window, initial M, update Update[M, Msg], frame Frame[M, Msg]) error {
	model := initial
	var messages Queue[Msg]
	var ops op.Ops
	send := func(msg Msg) {
		messages.Push(msg)
		w.Invalidate()
	}

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			messages.Drain(func(msg Msg) {
				StartCmd(update(&model, msg), send)
			})
			gtx := app.NewContext(&ops, e)
			frame(gtx, model, send)
			e.Frame(gtx.Ops)
		}
	}
}
