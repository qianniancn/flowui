package runtime

import (
	"context"
	"errors"
	"testing"
	"time"
)

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
	sent := make(chan int, 1)
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
