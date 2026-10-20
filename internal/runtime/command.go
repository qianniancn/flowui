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

// Send delivers a message to the runtime, which queues it for the next frame.
// It is safe for concurrent effects to call Send. The queue is bounded; a
// burst beyond that bound drops the newest messages and reports
// QueueOverflowError, so a sender must not assume unlimited buffering.
type Send[Msg any] func(Msg)

// Cmd is asynchronous work started by Update, Init, or a Subscription. It runs
// after the callback that returned it and may send messages back later. Its
// context is canceled when the window closes.
type Cmd[Msg any] func(context.Context, Send[Msg]) error

type latestCommandKey struct{}
type latestDispatchKey struct{}

type latestCommandToken struct {
	key        string
	generation uint64
}

type latestCommandState struct {
	generation uint64
	cancel     context.CancelFunc
}

type latestCommandManager struct {
	mu     sync.Mutex
	states map[string]latestCommandState
}

func newLatestCommandManager() *latestCommandManager {
	return &latestCommandManager{states: make(map[string]latestCommandState)}
}

func withLatestCommandManager(ctx context.Context, manager *latestCommandManager) context.Context {
	return context.WithValue(ctx, latestCommandKey{}, manager)
}

func (m *latestCommandManager) start(ctx context.Context, key string) (context.Context, latestCommandToken, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.states[key]
	if current.cancel != nil {
		current.cancel()
	}
	generation := current.generation + 1
	if generation == 0 {
		// Reserve zero for the manager's untouched state if a process ever
		// executes more than MaxUint64 latest commands for one key.
		generation = 1
	}
	child, cancel := context.WithCancel(ctx)
	m.states[key] = latestCommandState{generation: generation, cancel: cancel}
	return child, latestCommandToken{key: key, generation: generation}, true
}

func (m *latestCommandManager) accepts(token latestCommandToken) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.states[token.key]
	return ok && current.generation == token.generation
}

func (m *latestCommandManager) finish(token latestCommandToken) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.states[token.key]
	if ok && current.generation == token.generation {
		// Keep a generation tombstone so older LatestCmd closures cannot start
		// after a newer generation has finished. Cancel the child before
		// releasing its cancel function from the state.
		if current.cancel != nil {
			current.cancel()
		}
		current.cancel = nil
		m.states[token.key] = current
	}
}

func (m *latestCommandManager) cancel(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.states[key]
	if current.cancel != nil {
		current.cancel()
	}
	// Keep a tombstone generation so in-flight sends from older tokens are
	// rejected, but do not retain a cancel func.
	current.generation++
	if current.generation == 0 {
		current.generation = 1
	}
	current.cancel = nil
	m.states[key] = current
}

// LatestCmd cancels an older command with the same key and drops messages sent
// by an older generation. Outside a FlowUI runtime it behaves like cmd.
//
// A generation is assigned when the returned command starts in a runtime, so
// the same Cmd value may safely be reused across Update calls.
func LatestCmd[Msg any](key string, cmd Cmd[Msg]) Cmd[Msg] {
	if key == "" {
		panic("flowui: empty latest command key")
	}
	if cmd == nil {
		return nil
	}
	return func(ctx context.Context, send Send[Msg]) error {
		manager, _ := ctx.Value(latestCommandKey{}).(*latestCommandManager)
		if manager == nil {
			return cmd(ctx, send)
		}
		child, token, ok := manager.start(ctx, key)
		if !ok {
			return nil
		}
		defer manager.finish(token)
		dispatch, _ := ctx.Value(latestDispatchKey{}).(func(latestCommandToken, Msg))
		err := cmd(child, func(msg Msg) {
			if !manager.accepts(token) || child.Err() != nil {
				return
			}
			if dispatch != nil {
				dispatch(token, msg)
				return
			}
			send(msg)
		})
		if child.Err() != nil {
			return nil
		}
		return err
	}
}

// CancelLatestCmd cancels the active command with key and invalidates its
// queued messages.
//
// The cancellation generation is assigned when the returned command runs, so
// the same cancellation command may safely be reused.
func CancelLatestCmd[Msg any](key string) Cmd[Msg] {
	if key == "" {
		panic("flowui: empty latest command key")
	}
	return func(ctx context.Context, _ Send[Msg]) error {
		manager, _ := ctx.Value(latestCommandKey{}).(*latestCommandManager)
		if manager != nil {
			manager.cancel(key)
		}
		return nil
	}
}

// Batch combines commands so Update can start several independent effects at
// once. Nil commands are ignored. Remaining commands run concurrently after
// Update returns, share the effect context, and may each call send. The first
// non-nil error is returned after every command finishes; a panic in any
// command is re-raised so the runtime reports it like a single command panic.
func Batch[Msg any](cmds ...Cmd[Msg]) Cmd[Msg] {
	active := make([]Cmd[Msg], 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			active = append(active, cmd)
		}
	}
	switch len(active) {
	case 0:
		return nil
	case 1:
		return active[0]
	}
	return func(ctx context.Context, send Send[Msg]) error {
		var (
			wg       sync.WaitGroup
			mu       sync.Mutex
			firstErr error
			panicked any
		)
		for _, cmd := range active {
			wg.Go(func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						mu.Lock()
						if panicked == nil {
							panicked = recovered
						}
						mu.Unlock()
					}
				}()
				err := cmd(ctx, send)
				if err == nil || ctx.Err() != nil {
					return
				}
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			})
		}
		wg.Wait()
		if panicked != nil {
			panic(panicked)
		}
		return firstErr
	}
}

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
	send Send[Msg],
	report func(error),
	onDone func(),
) <-chan struct{} {
	done := make(chan struct{})
	group.wait.Go(func() {
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
	})
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
