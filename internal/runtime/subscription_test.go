package runtime

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestSubscriptionSetStartsOnceAndCancelsWhenRemoved(t *testing.T) {
	root := t.Context()
	var set subscriptionSet[int]
	var effects effectGroup
	var starts atomic.Int32
	started := make(chan struct{})
	canceled := make(chan struct{})
	sent := make(chan int, 1)
	subscription := Subscription[int]{
		Key: "clock",
		Run: func(ctx context.Context, send func(int)) error {
			starts.Add(1)
			close(started)
			<-ctx.Done()
			send(1)
			close(canceled)
			return ctx.Err()
		},
	}

	set.reconcile(root, []Subscription[int]{subscription}, &effects, func(_ subscriptionToken, msg int) {
		sent <- msg
	}, nil, nil)
	receiveRuntimeTestValue(t, started)
	set.reconcile(root, []Subscription[int]{subscription}, &effects, func(_ subscriptionToken, msg int) {
		sent <- msg
	}, nil, nil)
	if got := starts.Load(); got != 1 {
		t.Fatalf("subscription starts = %d, want 1", got)
	}

	set.reconcile(root, nil, &effects, func(_ subscriptionToken, msg int) {
		sent <- msg
	}, nil, nil)
	receiveRuntimeTestValue(t, canceled)
	select {
	case msg := <-sent:
		t.Fatalf("removed subscription sent late message %d", msg)
	default:
	}
}

func TestSubscriptionSetReportsKeyedError(t *testing.T) {
	root := t.Context()
	var set subscriptionSet[int]
	var effects effectGroup
	want := errors.New("stream failed")
	reported := make(chan error, 1)
	set.reconcile(root, []Subscription[int]{
		{
			Key: "updates",
			Run: func(context.Context, func(int)) error {
				return want
			},
		},
	}, &effects, func(subscriptionToken, int) {}, func(err error) {
		reported <- err
	}, nil)

	err := receiveRuntimeTestValue(t, reported)
	var effectErr *EffectError
	if !errors.As(err, &effectErr) {
		t.Fatalf("reported error type = %T, want *EffectError", err)
	}
	if effectErr.Kind != EffectSubscription || effectErr.Key != "updates" || !errors.Is(effectErr, want) {
		t.Fatalf("reported subscription error = %#v", effectErr)
	}
	set.close()
}

func TestSubscriptionSetCancelsRemovedKeyBeforeStartingReplacement(t *testing.T) {
	root := t.Context()
	var set subscriptionSet[int]
	var effects effectGroup
	oldStarted := make(chan struct{})
	newStarted := make(chan struct{})
	wake := make(chan struct{}, 2)
	var oldContext context.Context
	set.reconcile(root, []Subscription[int]{
		{
			Key: "old",
			Run: func(ctx context.Context, _ func(int)) error {
				oldContext = ctx
				close(oldStarted)
				<-ctx.Done()
				return ctx.Err()
			},
		},
	}, &effects, func(subscriptionToken, int) {}, nil, func() { wake <- struct{}{} })
	receiveRuntimeTestValue(t, oldStarted)

	set.reconcile(root, []Subscription[int]{
		{
			Key: "new",
			Run: func(ctx context.Context, _ func(int)) error {
				if oldContext.Err() == nil {
					t.Error("replacement started before removed subscription was canceled")
				}
				close(newStarted)
				<-ctx.Done()
				return ctx.Err()
			},
		},
	}, &effects, func(subscriptionToken, int) {}, nil, func() { wake <- struct{}{} })
	receiveRuntimeTestValue(t, wake)
	set.reconcile(root, []Subscription[int]{
		{
			Key: "new",
			Run: func(ctx context.Context, _ func(int)) error {
				if oldContext.Err() == nil {
					t.Error("replacement started before removed subscription was canceled")
				}
				close(newStarted)
				<-ctx.Done()
				return ctx.Err()
			},
		},
	}, &effects, func(subscriptionToken, int) {}, nil, func() { wake <- struct{}{} })
	receiveRuntimeTestValue(t, newStarted)
	set.close()
}

func TestSubscriptionSetStartsUnrelatedKeyWhileAnotherIsStopping(t *testing.T) {
	// A hung subscription on "old" must not delay starting a different key.
	root := t.Context()
	set := subscriptionSet[int]{stopGracePeriod: time.Hour}
	var effects effectGroup
	oldStarted := make(chan struct{})
	newStarted := make(chan struct{})
	releaseOld := make(chan struct{})
	set.reconcile(root, []Subscription[int]{
		{
			Key: "old",
			Run: func(context.Context, func(int)) error {
				close(oldStarted)
				<-releaseOld
				return nil
			},
		},
	}, &effects, func(subscriptionToken, int) {}, nil, nil)
	receiveRuntimeTestValue(t, oldStarted)

	set.reconcile(root, []Subscription[int]{
		{
			Key: "new",
			Run: func(ctx context.Context, _ func(int)) error {
				close(newStarted)
				<-ctx.Done()
				return ctx.Err()
			},
		},
	}, &effects, func(subscriptionToken, int) {}, nil, nil)
	receiveRuntimeTestValue(t, newStarted)

	set.close()
	close(releaseOld)
	if !effects.waitFor(time.Second) {
		t.Fatal("subscription effects did not stop")
	}
}

func TestSubscriptionSetSameKeyWaitsForStopGrace(t *testing.T) {
	root := t.Context()
	set := subscriptionSet[int]{stopGracePeriod: 20 * time.Millisecond}
	var effects effectGroup
	oldStarted := make(chan struct{})
	newStarted := make(chan struct{})
	releaseOld := make(chan struct{})
	wake := make(chan struct{}, 4)
	set.reconcile(root, []Subscription[int]{
		{
			Key: "same",
			Run: func(context.Context, func(int)) error {
				close(oldStarted)
				<-releaseOld
				return nil
			},
		},
	}, &effects, func(subscriptionToken, int) {}, nil, func() { wake <- struct{}{} })
	receiveRuntimeTestValue(t, oldStarted)

	// Remove then re-add the same key while the old Run is hung.
	set.reconcile(root, nil, &effects, func(subscriptionToken, int) {}, nil, func() { wake <- struct{}{} })
	desired := []Subscription[int]{
		{
			Key: "same",
			Run: func(ctx context.Context, _ func(int)) error {
				close(newStarted)
				<-ctx.Done()
				return ctx.Err()
			},
		},
	}
	set.reconcile(root, desired, &effects, func(subscriptionToken, int) {}, nil, func() { wake <- struct{}{} })
	select {
	case <-newStarted:
		t.Fatal("same-key replacement started before stop grace elapsed")
	case <-time.After(5 * time.Millisecond):
	}
	receiveRuntimeTestValue(t, wake)
	set.reconcile(root, desired, &effects, func(subscriptionToken, int) {}, nil, func() { wake <- struct{}{} })
	receiveRuntimeTestValue(t, newStarted)

	set.close()
	close(releaseOld)
	if !effects.waitFor(time.Second) {
		t.Fatal("subscription effects did not stop")
	}
}

func TestLoopCoreDropsQueuedMessageFromStoppedSubscription(t *testing.T) {
	type model struct{ messages []int }
	core := newLoopCore(model{}, func(model *model, msg int) Cmd[int] {
		model.messages = append(model.messages, msg)
		return nil
	})
	root := t.Context()
	var set subscriptionSet[int]
	var effects effectGroup
	started := make(chan struct{})
	set.reconcile(root, []Subscription[int]{
		{
			Key: "events",
			Run: func(ctx context.Context, _ func(int)) error {
				close(started)
				<-ctx.Done()
				return ctx.Err()
			},
		},
	}, &effects, core.sendSubscription, nil, nil)
	receiveRuntimeTestValue(t, started)
	token := set.active["events"].token
	core.sendSubscription(token, 1)
	set.reconcile(root, nil, &effects, core.sendSubscription, nil, nil)

	var viewed model
	core.frame(&effects, root, core.send, nil, set.accepts, func(model model) {
		viewed = model
	})
	if len(viewed.messages) != 0 {
		t.Fatalf("stopped subscription messages = %v, want none", viewed.messages)
	}
	set.close()
	if !effects.waitFor(time.Second) {
		t.Fatal("stopped subscription did not exit")
	}
}

func TestSubscriptionSetRejectsDuplicateKeys(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("duplicate subscription key did not panic")
		}
	}()
	var set subscriptionSet[int]
	var effects effectGroup
	run := func(context.Context, func(int)) error { return nil }
	set.reconcile(context.Background(), []Subscription[int]{
		{Key: "events", Run: run},
		{Key: "events", Run: run},
	}, &effects, func(subscriptionToken, int) {}, nil, nil)
}

func TestCompletedSubscriptionIsRetainedUntilRemoved(t *testing.T) {
	var set subscriptionSet[int]
	var effects effectGroup
	var starts atomic.Int32
	completed := make(chan struct{})
	subscription := Subscription[int]{
		Key: "one-shot",
		Run: func(context.Context, func(int)) error {
			starts.Add(1)
			close(completed)
			return nil
		},
	}
	set.reconcile(context.Background(), []Subscription[int]{subscription}, &effects, func(subscriptionToken, int) {}, nil, nil)
	receiveRuntimeTestValue(t, completed)
	set.reconcile(context.Background(), []Subscription[int]{subscription}, &effects, func(subscriptionToken, int) {}, nil, nil)
	if got := starts.Load(); got != 1 {
		t.Fatalf("completed subscription starts = %d, want 1", got)
	}
	set.close()
}
