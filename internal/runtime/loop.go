package runtime

import (
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
)

type Update[M any, Msg any] func(*M, Msg) func(func(Msg))

type Frame[M any, Msg any] func(layout.Context, M, func(Msg))

type loopCore[M any, Msg any] struct {
	model    M
	messages Queue[Msg]
	update   Update[M, Msg]
}

func newLoopCore[M any, Msg any](initial M, update Update[M, Msg]) *loopCore[M, Msg] {
	return &loopCore[M, Msg]{
		model:  initial,
		update: update,
	}
}

func (c *loopCore[M, Msg]) send(msg Msg) {
	c.messages.Push(msg)
}

func (c *loopCore[M, Msg]) frame(send func(Msg), view func(M)) {
	c.messages.Drain(func(msg Msg) {
		StartCmd(c.update(&c.model, msg), send)
	})
	view(c.model)
}

func Loop[M any, Msg any](w *app.Window, initial M, update Update[M, Msg], frame Frame[M, Msg]) error {
	core := newLoopCore(initial, update)
	var ops op.Ops
	send := func(msg Msg) {
		core.send(msg)
		w.Invalidate()
	}

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			core.frame(send, func(model M) {
				gtx := app.NewContext(&ops, e)
				frame(gtx, model, send)
				e.Frame(gtx.Ops)
			})
		}
	}
}
