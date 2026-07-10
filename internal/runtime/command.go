package runtime

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

// EffectKind identifies asynchronous work managed by the runtime.
type EffectKind string

const (
	EffectCommand      EffectKind = "command"
	EffectSubscription EffectKind = "subscription"
)

// EffectError describes an error or panic produced by asynchronous work.
type EffectError struct {
	Kind  EffectKind
	Key   string
	Err   error
	Panic any
	Stack []byte
}

func (e *EffectError) Error() string {
	source := string(e.Kind)
	if e.Key != "" {
		source += fmt.Sprintf(" %q", e.Key)
	}
	if e.Panic != nil {
		return fmt.Sprintf("%s panicked: %v", source, e.Panic)
	}
	return fmt.Sprintf("%s failed: %v", source, e.Err)
}

func (e *EffectError) Unwrap() error {
	return e.Err
}

type Cmd[Msg any] func(context.Context, func(Msg)) error

// ErrEffectShutdownTimeout reports asynchronous work that did not stop within
// the runtime's bounded shutdown period.
var ErrEffectShutdownTimeout = fmt.Errorf("timed out waiting for asynchronous effects to stop")

type effectGroup struct {
	wait sync.WaitGroup
}

func StartCmd[Msg any](group *effectGroup, ctx context.Context, cmd Cmd[Msg], send func(Msg), report func(error)) {
	if cmd != nil {
		startEffect(group, ctx, EffectCommand, "", cmd, send, report, nil)
	}
}

func startEffect[Msg any](
	group *effectGroup,
	ctx context.Context,
	kind EffectKind,
	key string,
	run Cmd[Msg],
	send func(Msg),
	report func(error),
	onDone func(),
) <-chan struct{} {
	done := make(chan struct{})
	group.wait.Add(1)
	go func() {
		defer group.wait.Done()
		defer func() {
			close(done)
			if onDone != nil {
				onDone()
			}
		}()
		defer func() {
			if recovered := recover(); recovered != nil {
				reportEffectError(report, &EffectError{
					Kind:  kind,
					Key:   key,
					Panic: recovered,
					Stack: debug.Stack(),
				})
			}
		}()

		err := run(ctx, func(msg Msg) {
			if ctx.Err() == nil {
				send(msg)
			}
		})
		if err == nil || ctx.Err() != nil {
			return
		}
		reportEffectError(report, &EffectError{Kind: kind, Key: key, Err: err})
	}()
	return done
}

func reportEffectError(report func(error), err error) {
	if report != nil && err != nil {
		report(err)
	}
}

func (g *effectGroup) waitFor(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		g.wait.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
