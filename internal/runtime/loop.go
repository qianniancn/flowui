package runtime

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/system"
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
	// Recover per message so earlier updates in the batch stand, later messages
	// are discarded, and no new commands start after a panic.
	dropped = c.messages.Drain(func(msg queuedMessage[Msg]) {
		if err != nil {
			return
		}
		if msg.fromSub && (acceptSubscription == nil || !acceptSubscription(msg.subscription)) {
			return
		}
		if msg.fromLatest && !c.latest.accepts(msg.latest) {
			return
		}
		var cmd Cmd[Msg]
		if panicErr := recoverRuntimePanic(RuntimePhaseUpdate, func() {
			updated = true
			cmd = c.update(&c.model, msg.value)
		}); panicErr != nil {
			err = panicErr
			return
		}
		StartCmd(group, ctx, cmd, send, report)
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

// windowCloser is implemented by windows that can request a native close.
// *app.Window implements this; test harnesses may inject DestroyEvent.
type windowCloser interface {
	Perform(system.Action)
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
	return LoopWithExit(w, initial, initialCmd, update, subscriptions, onError, onDestroy, onConfig, frame, nil)
}

// LoopWithExit runs an event loop and reports the final model before
// onDestroy makes the window available for reopening.
func LoopWithExit[M any, Msg any](
	w EventWindow,
	initial M,
	initialCmd Cmd[Msg],
	update Update[M, Msg],
	subscriptions Subscriptions[M, Msg],
	onError func(error),
	onDestroy func(),
	onConfig func(app.Config, func(Msg)),
	frame Frame[M, Msg],
	onExit func(M),
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
	var fatal error
	exitReported := false
	reportExit := func() {
		if exitReported {
			return
		}
		exitReported = true
		if onExit != nil {
			onExit(core.model)
		}
	}
	defer func() {
		reportExit()
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

	beginFatalShutdown := func(err error) {
		if fatal != nil || err == nil {
			return
		}
		fatal = err
		cancel()
		if !requestWindowClose(w) {
			// Harnesses without a native close path return immediately so tests
			// and headless drivers are not stuck waiting for DestroyEvent.
			return
		}
	}

	for {
		e := w.Event()
		if fatal != nil {
			switch e := e.(type) {
			case app.DestroyEvent:
				reportExit()
				if onDestroy != nil {
					onDestroy()
				}
				return fatal
			case app.FrameEvent:
				// Acknowledge frames while waiting for destroy so the platform
				// event path keeps moving.
				var empty op.Ops
				gtx := app.NewContext(&empty, e)
				e.Frame(gtx.Ops)
			}
			continue
		}

		switch e := e.(type) {
		case app.DestroyEvent:
			reportExit()
			if onDestroy != nil {
				onDestroy()
			}
			cancel()
			return e.Err
		case app.ConfigEvent:
			if onConfig != nil {
				if err := recoverRuntimePanic(RuntimePhaseUpdate, func() {
					onConfig(e.Config, send)
				}); err != nil {
					beginFatalShutdown(err)
					if !windowSupportsClose(w) {
						return fatal
					}
					continue
				}
			}
			w.Invalidate()
		case app.FrameEvent:
			droppedErrors := effectErrors.Drain(onError)
			if droppedErrors > 0 {
				onError(&QueueOverflowError{Queue: "error", Dropped: droppedErrors, Limit: errorQueueLimit})
			}
			updated, droppedMessages, err := core.updateMessages(&effects, ctx, send, report, activeSubscriptions.accepts)
			if err != nil {
				beginFatalShutdown(err)
				if !windowSupportsClose(w) {
					return fatal
				}
				continue
			}
			if droppedMessages > 0 {
				onError(&QueueOverflowError{Queue: "message", Dropped: droppedMessages, Limit: messageQueueLimit})
			}
			if subscriptions != nil && (!subscriptionsReady || updated) {
				var desired []Subscription[Msg]
				if err := recoverRuntimePanic(RuntimePhaseSubscriptions, func() {
					desired = subscriptions(core.model)
				}); err != nil {
					beginFatalShutdown(err)
					if !windowSupportsClose(w) {
						return fatal
					}
					continue
				}
				desiredSubscriptions = desired
				subscriptionsReady = true
			}
			if subscriptions != nil && subscriptionsReady {
				if err := recoverRuntimePanic(RuntimePhaseSubscriptions, func() {
					activeSubscriptions.reconcile(ctx, desiredSubscriptions, &effects, sendSubscription, report, wake)
				}); err != nil {
					beginFatalShutdown(err)
					if !windowSupportsClose(w) {
						return fatal
					}
					continue
				}
			}
			viewErr := recoverRuntimePanic(RuntimePhaseView, func() {
				gtx := app.NewContext(&ops, e)
				frame(gtx, core.model, send)
				e.Frame(gtx.Ops)
			})
			if viewErr != nil {
				beginFatalShutdown(viewErr)
				if !windowSupportsClose(w) {
					return fatal
				}
			}
		}
	}
}

func windowSupportsClose(w EventWindow) bool {
	_, ok := w.(windowCloser)
	return ok
}

func requestWindowClose(w EventWindow) bool {
	closer, ok := w.(windowCloser)
	if !ok {
		return false
	}
	closer.Perform(system.ActionClose)
	return true
}
