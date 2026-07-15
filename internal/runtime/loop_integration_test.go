package runtime

import (
	"context"
	"errors"
	"image"
	"sync/atomic"
	"testing"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/layout"
	"gioui.org/op"
)

type runtimeTestWindow struct {
	events      chan event.Event
	invalidated chan struct{}
}

func newRuntimeTestWindow() *runtimeTestWindow {
	return &runtimeTestWindow{
		events:      make(chan event.Event),
		invalidated: make(chan struct{}, 8),
	}
}

func (w *runtimeTestWindow) Event() event.Event {
	return <-w.events
}

func (w *runtimeTestWindow) Invalidate() {
	select {
	case w.invalidated <- struct{}{}:
	default:
	}
}

func TestLoopCancelsAndWaitsForSubscriptionOnDestroy(t *testing.T) {
	window := newRuntimeTestWindow()
	started := make(chan struct{})
	destroyed := make(chan struct{})
	canceled := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		result <- loop(window, struct{}{}, func(*struct{}, struct{}) Cmd[struct{}] {
			return nil
		}, func(struct{}) []Subscription[struct{}] {
			return []Subscription[struct{}]{
				{
					Key: "events",
					Run: func(ctx context.Context, _ func(struct{})) error {
						close(started)
						<-ctx.Done()
						close(canceled)
						return ctx.Err()
					},
				},
			}
		}, nil, func() { close(destroyed) }, func(layout.Context, struct{}, func(struct{})) {})
	}()

	window.events <- runtimeFrameEvent()
	receiveRuntimeTestValue(t, started)
	window.events <- app.DestroyEvent{}
	receiveRuntimeTestValue(t, destroyed)
	if err := receiveRuntimeTestValue(t, result); err != nil {
		t.Fatalf("loop returned error: %v", err)
	}
	select {
	case <-canceled:
	default:
		t.Fatal("loop returned before subscription cleanup completed")
	}
}

func TestLoopDeliversEffectErrorsBeforeNextFrame(t *testing.T) {
	window := newRuntimeTestWindow()
	want := errors.New("stream failed")
	reported := make(chan error, 1)
	result := make(chan error, 1)
	var frames atomic.Int32
	go func() {
		result <- loop(window, struct{}{}, func(*struct{}, struct{}) Cmd[struct{}] {
			return nil
		}, func(struct{}) []Subscription[struct{}] {
			return []Subscription[struct{}]{
				{
					Key: "events",
					Run: func(context.Context, func(struct{})) error { return want },
				},
			}
		}, func(err error) {
			if frames.Load() != 1 {
				t.Errorf("error handler ran after frame %d, want before frame 2", frames.Load()+1)
			}
			reported <- err
		}, nil, func(layout.Context, struct{}, func(struct{})) {
			frames.Add(1)
		})
	}()

	window.events <- runtimeFrameEvent()
	receiveRuntimeTestValue(t, window.invalidated)
	window.events <- runtimeFrameEvent()
	err := receiveRuntimeTestValue(t, reported)
	var effectErr *EffectError
	if !errors.As(err, &effectErr) || !errors.Is(effectErr, want) {
		t.Fatalf("reported error = %v, want %v", err, want)
	}
	window.events <- app.DestroyEvent{}
	if err := receiveRuntimeTestValue(t, result); err != nil {
		t.Fatalf("loop returned error: %v", err)
	}
}

func runtimeFrameEvent() app.FrameEvent {
	return app.FrameEvent{
		Now:   time.Now(),
		Size:  image.Pt(100, 100),
		Frame: func(*op.Ops) {},
	}
}
