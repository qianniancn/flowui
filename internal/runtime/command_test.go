package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestBatchRunsCommandsConcurrentlyAndCollectsMessages(t *testing.T) {
	const commandCount = 8
	var ready sync.WaitGroup
	var finished sync.WaitGroup
	ready.Add(commandCount)
	finished.Add(commandCount)
	gate := make(chan struct{})
	cmds := make([]Cmd[int], commandCount)
	for id := range commandCount {
		cmds[id] = func(context.Context, func(int)) error {
			ready.Done()
			<-gate
			finished.Done()
			return nil
		}
	}
	batch := Batch(append(cmds, nil)...)
	if batch == nil {
		t.Fatal("batch with active commands was nil")
	}

	done := make(chan error, 1)
	go func() {
		done <- batch(context.Background(), func(int) {})
	}()
	readyDone := make(chan struct{})
	go func() {
		ready.Wait()
		close(readyDone)
	}()
	select {
	case <-readyDone:
	case <-time.After(time.Second):
		close(gate)
		t.Fatal("timed out waiting for batch commands to start")
	}
	close(gate)
	finishedDone := make(chan struct{})
	go func() {
		finished.Wait()
		close(finishedDone)
	}()
	select {
	case <-finishedDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for batch commands to finish")
	}
	if err := receiveRuntimeTestValue(t, done); err != nil {
		t.Fatalf("batch error = %v", err)
	}
}

func TestBatchReturnsErrorAndPropagatesPanic(t *testing.T) {
	want := errors.New("command failed")
	batch := Batch(
		func(context.Context, func(int)) error { return want },
		func(context.Context, func(int)) error { return nil },
	)
	if err := batch(context.Background(), func(int) {}); !errors.Is(err, want) {
		t.Fatalf("batch error = %v, want %v", err, want)
	}

	panicking := Batch(
		func(context.Context, func(int)) error { return nil },
		func(context.Context, func(int)) error { panic("batch child") },
	)
	defer func() {
		if recover() == nil {
			t.Fatal("batch did not re-raise child panic")
		}
	}()
	_ = panicking(context.Background(), func(int) {})
}

func TestBatchIgnoresNilCommands(t *testing.T) {
	if Batch[int]() != nil || Batch[int](nil, nil) != nil {
		t.Fatal("empty batch should be nil")
	}
	var ran bool
	single := Batch[int](nil, func(context.Context, func(int)) error {
		ran = true
		return nil
	}, nil)
	if single == nil {
		t.Fatal("single-command batch was nil")
	}
	if err := single(context.Background(), func(int) {}); err != nil || !ran {
		t.Fatalf("single-command batch err=%v ran=%v", err, ran)
	}
}

func TestStartCmdReportsReturnedError(t *testing.T) {
	want := errors.New("load failed")
	reported := make(chan error, 1)
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(context.Context, func(int)) error {
		return want
	}, func(int) {}, func(err error) {
		reported <- err
	})

	err := receiveRuntimeTestValue(t, reported)
	var effectErr *EffectError
	if !errors.As(err, &effectErr) {
		t.Fatalf("reported error type = %T, want *EffectError", err)
	}
	if effectErr.Kind != EffectCommand || effectErr.Key != "" || !errors.Is(effectErr, want) {
		t.Fatalf("reported command error = %#v", effectErr)
	}
}

func TestStartCmdReportsContextCanceledWhileEffectIsActive(t *testing.T) {
	reported := make(chan error, 1)
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(context.Context, func(int)) error {
		return context.Canceled
	}, func(int) {}, func(err error) {
		reported <- err
	})

	err := receiveRuntimeTestValue(t, reported)
	var effectErr *EffectError
	if !errors.As(err, &effectErr) || !errors.Is(effectErr, context.Canceled) {
		t.Fatalf("reported cancellation = %v, want active command error", err)
	}
}

func TestStartCmdRecoversPanic(t *testing.T) {
	reported := make(chan error, 1)
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(context.Context, func(int)) error {
		panic("broken command")
	}, func(int) {}, func(err error) {
		reported <- err
	})

	err := receiveRuntimeTestValue(t, reported)
	var effectErr *EffectError
	if !errors.As(err, &effectErr) {
		t.Fatalf("reported error type = %T, want *EffectError", err)
	}
	if effectErr.Kind != EffectCommand || effectErr.Panic != "broken command" {
		t.Fatalf("reported panic = %#v", effectErr)
	}
	if len(effectErr.Stack) == 0 {
		t.Fatal("panic report has no stack trace")
	}
}

func TestStartCmdCancelsAndDropsLateMessages(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	finished := make(chan struct{})
	sent := make(chan int, 2)
	reported := make(chan error, 1)
	var effects effectGroup
	StartCmd(&effects, ctx, func(ctx context.Context, send func(int)) error {
		close(started)
		<-ctx.Done()
		send(1)
		close(finished)
		return ctx.Err()
	}, func(msg int) {
		sent <- msg
	}, func(err error) {
		reported <- err
	})

	receiveRuntimeTestValue(t, started)
	cancel()
	receiveRuntimeTestValue(t, finished)
	select {
	case msg := <-sent:
		t.Fatalf("canceled command sent late message %d", msg)
	default:
	}
	select {
	case err := <-reported:
		t.Fatalf("canceled command reported error %v", err)
	default:
	}
	if !effects.waitFor(time.Second) {
		t.Fatal("canceled command did not stop")
	}
}

func TestLatestCmdCancelsOlderGenerationAndDropsItsMessages(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	started := make(chan struct{})
	canceled := make(chan struct{})
	sent := make(chan int, 2)
	var effects effectGroup
	old := LatestCmd("search", func(ctx context.Context, send func(int)) error {
		close(started)
		<-ctx.Done()
		send(1)
		close(canceled)
		return ctx.Err()
	})
	newer := LatestCmd("search", func(_ context.Context, send func(int)) error {
		send(2)
		return nil
	})
	StartCmd(&effects, ctx, old, func(msg int) { sent <- msg }, func(error) {})
	receiveRuntimeTestValue(t, started)
	StartCmd(&effects, ctx, newer, func(msg int) { sent <- msg }, func(error) {})
	receiveRuntimeTestValue(t, canceled)
	if got := receiveRuntimeTestValue(t, sent); got != 2 {
		t.Fatalf("latest command message = %d, want 2", got)
	}
	select {
	case got := <-sent:
		t.Fatalf("stale command message = %d", got)
	default:
	}
	if !effects.waitFor(time.Second) {
		t.Fatal("latest commands did not stop")
	}
}

func TestLatestCmdSkipsGenerationStartedAfterNewerOne(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	sent := make(chan int, 2)
	old := LatestCmd("preview", func(_ context.Context, send func(int)) error {
		send(1)
		return nil
	})
	newer := LatestCmd("preview", func(_ context.Context, send func(int)) error {
		send(2)
		return nil
	})
	if err := newer(ctx, func(msg int) { sent <- msg }); err != nil {
		t.Fatalf("newer command error: %v", err)
	}
	if err := old(ctx, func(msg int) { sent <- msg }); err != nil {
		t.Fatalf("old command error: %v", err)
	}
	if got := receiveRuntimeTestValue(t, sent); got != 2 {
		t.Fatalf("latest command message = %d, want 2", got)
	}
	select {
	case got := <-sent:
		t.Fatalf("stale command ran with message %d", got)
	default:
	}
}

func TestLatestCmdFinishCancelsChildContext(t *testing.T) {
	parent := &afterFuncTestContext{Context: context.Background(), done: make(chan struct{})}
	manager := newLatestCommandManager()
	child, token, ok := manager.start(parent, "preview", 1)
	if !ok {
		t.Fatal("LatestCmd did not start")
	}
	manager.finish(token)
	if child == nil || child.Err() != context.Canceled {
		t.Fatalf("finished child context = %v, want canceled", child)
	}
	if !parent.stopped {
		t.Fatal("finished child context was not removed from parent")
	}
}

type afterFuncTestContext struct {
	context.Context
	done    chan struct{}
	stopped bool
}

func (c *afterFuncTestContext) Done() <-chan struct{} { return c.done }

func (c *afterFuncTestContext) AfterFunc(func()) func() bool {
	return func() bool {
		c.stopped = true
		return true
	}
}

func TestLatestCmdKeepsGenerationAfterHighKeyChurn(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	sent := make(chan int, 2)
	old := LatestCmd("preview", func(_ context.Context, send func(int)) error {
		send(1)
		return nil
	})
	newer := LatestCmd("preview", func(_ context.Context, send func(int)) error {
		send(2)
		return nil
	})
	if err := newer(ctx, func(msg int) { sent <- msg }); err != nil {
		t.Fatalf("newer command error: %v", err)
	}
	for index := range 300 {
		cmd := LatestCmd(fmt.Sprintf("churn-%d", index), func(context.Context, func(int)) error { return nil })
		if err := cmd(ctx, func(int) {}); err != nil {
			t.Fatalf("churn command %d error: %v", index, err)
		}
	}
	if err := old(ctx, func(msg int) { sent <- msg }); err != nil {
		t.Fatalf("old command error: %v", err)
	}
	if got := receiveRuntimeTestValue(t, sent); got != 2 {
		t.Fatalf("latest command message = %d, want 2", got)
	}
	select {
	case got := <-sent:
		t.Fatalf("stale command revived with message %d", got)
	default:
	}
}

func TestCancelLatestCmdStopsActiveGeneration(t *testing.T) {
	ctx := withLatestCommandManager(context.Background(), newLatestCommandManager())
	started := make(chan struct{})
	canceled := make(chan struct{})
	var effects effectGroup
	active := LatestCmd("load", func(ctx context.Context, _ func(int)) error {
		close(started)
		<-ctx.Done()
		close(canceled)
		return ctx.Err()
	})
	cancel := CancelLatestCmd[int]("load")
	StartCmd(&effects, ctx, active, func(int) {}, nil)
	receiveRuntimeTestValue(t, started)
	StartCmd(&effects, ctx, cancel, func(int) {}, nil)
	receiveRuntimeTestValue(t, canceled)
	if !effects.waitFor(time.Second) {
		t.Fatal("canceled latest command did not stop")
	}
}

func TestEffectGroupWaitIsBounded(t *testing.T) {
	release := make(chan struct{})
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(context.Context, func(int)) error {
		<-release
		return nil
	}, func(int) {}, nil)
	if effects.waitFor(10 * time.Millisecond) {
		t.Fatal("effect group reported idle while command was still running")
	}
	close(release)
	if !effects.waitFor(time.Second) {
		t.Fatal("effect group did not become idle after command completed")
	}
}

func receiveRuntimeTestValue[T any](t *testing.T, values <-chan T) T {
	t.Helper()
	select {
	case value := <-values:
		return value
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for runtime event")
		var zero T
		return zero
	}
}
