package runtime

import (
	"context"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
)

type Update[M any, Msg any] func(*M, Msg) Cmd[Msg]

type Frame[M any, Msg any] func(layout.Context, M, func(Msg))

type queuedMessage[Msg any] struct {
	value        Msg
	subscription subscriptionToken
	fromSub      bool
}

type loopCore[M any, Msg any] struct {
	model    M
	messages Queue[queuedMessage[Msg]]
	update   Update[M, Msg]
}

func newLoopCore[M any, Msg any](initial M, update Update[M, Msg]) *loopCore[M, Msg] {
	return &loopCore[M, Msg]{
		model:  initial,
		update: update,
	}
}

func (c *loopCore[M, Msg]) send(msg Msg) {
	c.messages.Push(queuedMessage[Msg]{value: msg})
}

func (c *loopCore[M, Msg]) sendSubscription(token subscriptionToken, msg Msg) {
	c.messages.Push(queuedMessage[Msg]{value: msg, subscription: token, fromSub: true})
}

func (c *loopCore[M, Msg]) frame(
	group *effectGroup,
	ctx context.Context,
	send func(Msg),
	report func(error),
	acceptSubscription func(subscriptionToken) bool,
	view func(M),
) {
	c.messages.Drain(func(msg queuedMessage[Msg]) {
		if msg.fromSub && (acceptSubscription == nil || !acceptSubscription(msg.subscription)) {
			return
		}
		StartCmd(group, ctx, c.update(&c.model, msg.value), send, report)
	})
	view(c.model)
}

type eventWindow interface {
	Event() event.Event
	Invalidate()
}

const effectShutdownTimeout = 2 * time.Second

func Loop[M any, Msg any](
	w *app.Window,
	initial M,
	update Update[M, Msg],
	subscriptions Subscriptions[M, Msg],
	onError func(error),
	onDestroy func(),
	onConfig func(app.Config),
	frame Frame[M, Msg],
) error {
	return loop(w, initial, update, subscriptions, onError, onDestroy, onConfig, frame)
}

func loop[M any, Msg any](
	w eventWindow,
	initial M,
	update Update[M, Msg],
	subscriptions Subscriptions[M, Msg],
	onError func(error),
	onDestroy func(),
	onConfig func(app.Config),
	frame Frame[M, Msg],
) error {
	core := newLoopCore(initial, update)
	ctx, cancel := context.WithCancel(context.Background())
	var activeSubscriptions subscriptionSet[Msg]
	var effects effectGroup
	var effectErrors Queue[error]
	var ops op.Ops
	defer func() {
		cancel()
		activeSubscriptions.close()
		if !effects.waitFor(effectShutdownTimeout) {
			onError(ErrEffectShutdownTimeout)
		}
	}()
	if onError == nil {
		onError = func(error) {}
	}
	send := func(msg Msg) {
		if ctx.Err() != nil {
			return
		}
		core.send(msg)
		w.Invalidate()
	}
	sendSubscription := func(token subscriptionToken, msg Msg) {
		if ctx.Err() != nil {
			return
		}
		core.sendSubscription(token, msg)
		w.Invalidate()
	}
	wake := func() {
		if ctx.Err() == nil {
			w.Invalidate()
		}
	}
	report := func(err error) {
		if err == nil || ctx.Err() != nil {
			return
		}
		effectErrors.Push(err)
		w.Invalidate()
	}

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			if onDestroy != nil {
				onDestroy()
			}
			cancel()
			return e.Err
		case app.ConfigEvent:
			if onConfig != nil {
				onConfig(e.Config)
			}
			w.Invalidate()
		case app.FrameEvent:
			effectErrors.Drain(onError)
			core.frame(&effects, ctx, send, report, activeSubscriptions.accepts, func(model M) {
				if subscriptions != nil {
					activeSubscriptions.reconcile(ctx, subscriptions(model), &effects, sendSubscription, report, wake)
				} else {
					activeSubscriptions.reconcile(ctx, nil, &effects, sendSubscription, report, wake)
				}
				gtx := app.NewContext(&ops, e)
				frame(gtx, model, send)
				e.Frame(gtx.Ops)
			})
		}
	}
}
