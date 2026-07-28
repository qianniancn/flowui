package runtime

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
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

var latestCommandGeneration atomic.Uint64

func newLatestCommandManager() *latestCommandManager {
	return &latestCommandManager{states: make(map[string]latestCommandState)}
}

func withLatestCommandManager(ctx context.Context, manager *latestCommandManager) context.Context {
	return context.WithValue(ctx, latestCommandKey{}, manager)
}

func (m *latestCommandManager) start(ctx context.Context, key string, generation uint64) (context.Context, latestCommandToken, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.states[key]
	if current.generation >= generation {
		return nil, latestCommandToken{}, false
	}
	if current.cancel != nil {
		current.cancel()
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

func (m *latestCommandManager) cancel(key string, generation uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.states[key]
	if current.generation > generation {
		return
	}
	if current.cancel != nil {
		current.cancel()
	}
	// Keep a tombstone generation so in-flight sends from older tokens are
	// rejected, but do not retain a cancel func.
	if current.generation < generation {
		current.generation = generation
	}
	current.cancel = nil
	m.states[key] = current
}

// LatestCmd cancels an older command with the same key and drops messages sent
// by an older generation. Outside a FlowUI runtime it behaves like cmd.
//
// IMPORTANT: The generation is captured at construction time. Reusing the same
// Cmd value across multiple Update calls will cause it to be silently rejected.
// Always construct a new LatestCmd inside each Update call:
//
//	// Correct:
//	case SomeEvent:
//	    return m, LatestCmd("fetch", fetchCmd())
//
//	// Wrong (reused cmd will be rejected):
//	var cmd = LatestCmd("fetch", fetchCmd())
//	case SomeEvent:
//	    return m, cmd
func LatestCmd[Msg any](key string, cmd Cmd[Msg]) Cmd[Msg] {
	if key == "" {
		panic("flowui: empty latest command key")
	}
	if cmd == nil {
		return nil
	}
	generation := latestCommandGeneration.Add(1)
	return func(ctx context.Context, send func(Msg)) error {
		manager, _ := ctx.Value(latestCommandKey{}).(*latestCommandManager)
		if manager == nil {
			return cmd(ctx, send)
		}
		child, token, ok := manager.start(ctx, key, generation)
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
// IMPORTANT: The generation is captured at construction time. Always construct
// a new CancelLatestCmd inside each Update call, not as a package-level variable.
// See LatestCmd for details.
func CancelLatestCmd[Msg any](key string) Cmd[Msg] {
	if key == "" {
		panic("flowui: empty latest command key")
	}
	generation := latestCommandGeneration.Add(1)
	return func(ctx context.Context, _ func(Msg)) error {
		manager, _ := ctx.Value(latestCommandKey{}).(*latestCommandManager)
		if manager != nil {
			manager.cancel(key, generation)
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
	return func(ctx context.Context, send func(Msg)) error {
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
	send func(Msg),
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
