package runtime

import (
	"context"
	"fmt"
	"runtime/debug"
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
	latest       latestCommandToken
	fromLatest   bool
}

const (
	messageQueueLimit = 256
	errorQueueLimit   = 64
)

type RuntimePhase string

const (
	RuntimePhaseUpdate        RuntimePhase = "update"
	RuntimePhaseSubscriptions RuntimePhase = "subscriptions"
	RuntimePhaseView          RuntimePhase = "view"
)

// RuntimePanicError describes a panic from synchronous application code.
type RuntimePanicError struct {
	Phase RuntimePhase
	Panic any
	Stack []byte
}

func (e *RuntimePanicError) Error() string {
	return fmt.Sprintf("%s panicked: %v", e.Phase, e.Panic)
}

// QueueOverflowError reports messages or errors dropped after a runtime queue
// reached its fixed bound.
type QueueOverflowError struct {
	Queue   string
	Dropped int
	Limit   int
}

func (e *QueueOverflowError) Error() string {
	return fmt.Sprintf("%s queue overflow: dropped %d (limit %d)", e.Queue, e.Dropped, e.Limit)
}

type loopCore[M any, Msg any] struct {
	model    M
	messages Queue[queuedMessage[Msg]]
	latest   *latestCommandManager
	update   Update[M, Msg]
}

func newLoopCore[M any, Msg any](initial M, update Update[M, Msg]) *loopCore[M, Msg] {
	return &loopCore[M, Msg]{
		model:    initial,
		update:   update,
		messages: Queue[queuedMessage[Msg]]{limit: messageQueueLimit},
		latest:   newLatestCommandManager(),
	}
}

func (c *loopCore[M, Msg]) send(msg Msg) {
	c.messages.Push(queuedMessage[Msg]{value: msg})
}

func (c *loopCore[M, Msg]) sendSubscription(token subscriptionToken, msg Msg) {
	c.messages.PushOrReplace(queuedMessage[Msg]{value: msg, subscription: token, fromSub: true}, func(current queuedMessage[Msg]) bool {
		return current.fromSub && current.subscription == token
	})
}

func (c *loopCore[M, Msg]) sendLatest(token latestCommandToken, msg Msg) {
	c.messages.PushOrReplace(queuedMessage[Msg]{value: msg, latest: token, fromLatest: true}, func(current queuedMessage[Msg]) bool {
		return current.fromLatest && current.latest == token
	})
}

func (c *loopCore[M, Msg]) updateMessages(
	group *effectGroup,
	ctx context.Context,
	send func(Msg),
	report func(error),
	acceptSubscription func(subscriptionToken) bool,
) (updated bool, dropped int, err error) {
	err = recoverRuntimePanic(RuntimePhaseUpdate, func() {
		dropped = c.messages.Drain(func(msg queuedMessage[Msg]) {
			if msg.fromSub && (acceptSubscription == nil || !acceptSubscription(msg.subscription)) {
				return
			}
			if msg.fromLatest && !c.latest.accepts(msg.latest) {
				return
			}
			updated = true
			StartCmd(group, ctx, c.update(&c.model, msg.value), send, report)
		})
	})
	return updated, dropped, err
}

func (c *loopCore[M, Msg]) frame(
	group *effectGroup,
	ctx context.Context,
	send func(Msg),
	report func(error),
	acceptSubscription func(subscriptionToken) bool,
	view func(M),
) (updated bool, dropped int, err error) {
	updated, dropped, err = c.updateMessages(group, ctx, send, report, acceptSubscription)
	if err != nil {
		return updated, dropped, err
	}
	err = recoverRuntimePanic(RuntimePhaseView, func() {
		view(c.model)
	})
	return updated, dropped, err
}

func recoverRuntimePanic(phase RuntimePhase, fn func()) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = &RuntimePanicError{Phase: phase, Panic: recovered, Stack: debug.Stack()}
		}
	}()
	fn()
	return nil
}

type EventWindow interface {
	Event() event.Event
	Invalidate()
}

const effectShutdownTimeout = 2 * time.Second

func Loop[M any, Msg any](
	w EventWindow,
	initial M,
	initialCmd Cmd[Msg],
	update Update[M, Msg],
	subscriptions Subscriptions[M, Msg],
	onError func(error),
	onDestroy func(),
	onConfig func(app.Config, func(Msg)),
	frame Frame[M, Msg],
) error {
	core := newLoopCore(initial, update)
	root, cancel := context.WithCancel(context.Background())
	ctx := withLatestCommandManager(root, core.latest)
	var activeSubscriptions subscriptionSet[Msg]
	var effects effectGroup
	effectErrors := Queue[error]{limit: errorQueueLimit}
	var desiredSubscriptions []Subscription[Msg]
	subscriptionsReady := subscriptions == nil
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
	ctx = context.WithValue(ctx, latestDispatchKey{}, func(token latestCommandToken, msg Msg) {
		if ctx.Err() != nil || !core.latest.accepts(token) {
			return
		}
		core.sendLatest(token, msg)
		w.Invalidate()
	})
	StartCmd(&effects, ctx, initialCmd, send, report)

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
				onConfig(e.Config, send)
			}
			w.Invalidate()
		case app.FrameEvent:
			droppedErrors := effectErrors.Drain(onError)
			if droppedErrors > 0 {
				onError(&QueueOverflowError{Queue: "error", Dropped: droppedErrors, Limit: errorQueueLimit})
			}
			updated, droppedMessages, err := core.updateMessages(&effects, ctx, send, report, activeSubscriptions.accepts)
			if err != nil {
				return err
			}
			if droppedMessages > 0 {
				onError(&QueueOverflowError{Queue: "message", Dropped: droppedMessages, Limit: messageQueueLimit})
			}
			if subscriptions != nil && (!subscriptionsReady || updated) {
				var desired []Subscription[Msg]
				if err := recoverRuntimePanic(RuntimePhaseSubscriptions, func() {
					desired = subscriptions(core.model)
				}); err != nil {
					return err
				}
				desiredSubscriptions = desired
				subscriptionsReady = true
			}
			if subscriptions != nil && subscriptionsReady {
				if err := recoverRuntimePanic(RuntimePhaseSubscriptions, func() {
					activeSubscriptions.reconcile(ctx, desiredSubscriptions, &effects, sendSubscription, report, wake)
				}); err != nil {
					return err
				}
			}
			var viewErr error
			viewErr = recoverRuntimePanic(RuntimePhaseView, func() {
				gtx := app.NewContext(&ops, e)
				frame(gtx, core.model, send)
				e.Frame(gtx.Ops)
			})
			if viewErr != nil {
				return viewErr
			}
		}
	}
}
