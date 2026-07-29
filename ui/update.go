package ui

import (
	"context"

	"github.com/qianniancn/flowui/internal/runtime"
)

// Send dispatches a message to Update. It is safe for concurrent commands to
// call Send; the messages are queued for a later frame.
type Send[Msg any] func(Msg)

// A Cmd returned by Update runs in its own goroutine after Update returns
// and may send messages back later. Its context is canceled when the window
// closes.
//
// A command may overlap later frames. It must capture only immutable value
// snapshots prepared by Update, must not retain or access the model pointer
// or a Context, and must report model-facing results through Send. Copy slices,
// maps, and other reference-backed model data before capturing them.
type Cmd[Msg any] func(context.Context, Send[Msg]) error

// LatestCmd cancels an older command with the same key and drops messages from
// an older generation. It is intended for search, preview, and autocomplete
// workflows where only the newest result is useful.
func LatestCmd[Msg any](key string, cmd Cmd[Msg]) Cmd[Msg] {
	wrapped := runtime.LatestCmd(key, func(ctx context.Context, send func(Msg)) error {
		if cmd == nil {
			return nil
		}
		return cmd(ctx, Send[Msg](send))
	})
	return func(ctx context.Context, send Send[Msg]) error {
		return wrapped(ctx, func(msg Msg) { send(msg) })
	}
}

// CancelLatestCmd cancels the active LatestCmd with key and invalidates its
// queued messages.
func CancelLatestCmd[Msg any](key string) Cmd[Msg] {
	wrapped := runtime.CancelLatestCmd[Msg](key)
	return func(ctx context.Context, send Send[Msg]) error {
		return wrapped(ctx, func(msg Msg) { send(msg) })
	}
}

// Update applies a message and may return a command to run. It must finish
// all model mutation before returning; a returned Cmd must follow the Cmd
// capture rules.
type Update[M any, Msg any] func(*M, Msg) Cmd[Msg]

// Do turns a context-free function into a command. Use DoContext for work that
// blocks, can fail, or must respond promptly to application shutdown.
func Do[Msg any](fn func(Send[Msg])) Cmd[Msg] {
	if fn == nil {
		return nil
	}
	return func(_ context.Context, send Send[Msg]) error {
		fn(send)
		return nil
	}
}

// DoContext creates a cancellable command that may return an error to the
// application's error handler.
func DoContext[Msg any](fn func(context.Context, Send[Msg]) error) Cmd[Msg] {
	return fn
}

// MapCmd adapts a child command to a parent message type without changing its
// context, execution, or error behavior.
func MapCmd[ChildMsg any, ParentMsg any](cmd Cmd[ChildMsg], mapMsg func(ChildMsg) ParentMsg) Cmd[ParentMsg] {
	if cmd == nil {
		return nil
	}
	if mapMsg == nil {
		panic("flowui: nil command mapper")
	}
	return func(ctx context.Context, send Send[ParentMsg]) error {
		return cmd(ctx, func(msg ChildMsg) {
			send(mapMsg(msg))
		})
	}
}

// EffectKind identifies asynchronous work managed by FlowUI.
type EffectKind = runtime.EffectKind

const (
	EffectCommand      = runtime.EffectCommand
	EffectSubscription = runtime.EffectSubscription
)

// EffectError describes an error or panic from a command or subscription.
type EffectError = runtime.EffectError

type RuntimePhase = runtime.RuntimePhase
type RuntimePanicError = runtime.RuntimePanicError
type QueueOverflowError = runtime.QueueOverflowError

const (
	RuntimePhaseUpdate        = runtime.RuntimePhaseUpdate
	RuntimePhaseSubscriptions = runtime.RuntimePhaseSubscriptions
	RuntimePhaseView          = runtime.RuntimePhaseView
)

// ErrEffectShutdownTimeout is reported when asynchronous work does not stop
// within the runtime's bounded shutdown period.
var ErrEffectShutdownTimeout = runtime.ErrEffectShutdownTimeout

// Subscription is long-lived asynchronous work with a stable identity.
type Subscription[Msg any] struct {
	key string
	run func(context.Context, Send[Msg]) error
}

// Subscriptions derives the active subscription set from the current model.
type Subscriptions[M any, Msg any] func(M) []Subscription[Msg]

// Subscribe creates a subscription. The key is its lifecycle identity: keep it
// stable to retain the running subscription and change it to restart with new
// parameters. A subscription should run until ctx is canceled. After its
// runner returns, the key remains retained until it is removed from the desired
// set; remove and re-add it or change the key to retry.
func Subscribe[Msg any](key string, run func(context.Context, Send[Msg]) error) Subscription[Msg] {
	if key == "" {
		panic("flowui: empty subscription key")
	}
	if run == nil {
		panic("flowui: nil subscription")
	}
	return Subscription[Msg]{key: key, run: run}
}
