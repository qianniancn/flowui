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
		result <- Loop(window, struct{}{}, nil, func(*struct{}, struct{}) Cmd[struct{}] {
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
		}, nil, func() { close(destroyed) }, nil, func(layout.Context, struct{}, func(struct{})) {})
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
		result <- Loop(window, struct{}{}, nil, func(*struct{}, struct{}) Cmd[struct{}] {
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
		}, nil, nil, func(layout.Context, struct{}, func(struct{})) {
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

func TestLoopReportsConfigChangesAndInvalidates(t *testing.T) {
	window := newRuntimeTestWindow()
	configured := make(chan app.Config, 1)
	result := make(chan error, 1)
	go func() {
		result <- Loop(
			window,
			struct{}{},
			nil,
			func(*struct{}, struct{}) Cmd[struct{}] { return nil },
			nil,
			nil,
			nil,
			func(config app.Config, _ func(struct{})) { configured <- config },
			func(layout.Context, struct{}, func(struct{})) {},
		)
	}()

	want := app.Config{Size: image.Pt(640, 480), Mode: app.Maximized, Focused: true, TopMost: true}
	window.events <- app.ConfigEvent{Config: want}
	if got := receiveRuntimeTestValue(t, configured); got != want {
		t.Fatalf("reported config = %#v, want %#v", got, want)
	}
	receiveRuntimeTestValue(t, window.invalidated)
	window.events <- app.DestroyEvent{}
	if err := receiveRuntimeTestValue(t, result); err != nil {
		t.Fatalf("loop returned error: %v", err)
	}
}

func TestLoopRunsInitialCommandAndQueuesItsMessage(t *testing.T) {
	window := newRuntimeTestWindow()
	commandSent := make(chan struct{})
	viewed := make(chan int, 1)
	result := make(chan error, 1)
	go func() {
		result <- Loop(
			window,
			0,
			func(_ context.Context, send func(int)) error {
				send(7)
				close(commandSent)
				return nil
			},
			func(model *int, msg int) Cmd[int] {
				*model = msg
				return nil
			},
			nil,
			nil,
			nil,
			nil,
			func(_ layout.Context, model int, _ func(int)) { viewed <- model },
		)
	}()

	receiveRuntimeTestValue(t, commandSent)
	window.events <- runtimeFrameEvent()
	if got := receiveRuntimeTestValue(t, viewed); got != 7 {
		t.Fatalf("initial command model = %d, want 7", got)
	}
	window.events <- app.DestroyEvent{}
	if err := receiveRuntimeTestValue(t, result); err != nil {
		t.Fatalf("loop returned error: %v", err)
	}
}

func TestLoopReturnsUpdatePanic(t *testing.T) {
	window := newRuntimeTestWindow()
	result := make(chan error, 1)
	go func() {
		result <- Loop(window, 0, nil, func(*int, struct{}) Cmd[struct{}] {
			panic("update broken")
		}, nil, nil, nil, func(_ app.Config, send func(struct{})) { send(struct{}{}) }, func(layout.Context, int, func(struct{})) {})
	}()
	window.events <- app.ConfigEvent{}
	window.events <- runtimeFrameEvent()
	err := receiveRuntimeTestValue(t, result)
	var panicErr *RuntimePanicError
	if !errors.As(err, &panicErr) || panicErr.Phase != RuntimePhaseUpdate || len(panicErr.Stack) == 0 {
		t.Fatalf("update panic = %#v", err)
	}
}

func TestLoopReturnsSubscriptionPanic(t *testing.T) {
	window := newRuntimeTestWindow()
	result := make(chan error, 1)
	go func() {
		result <- Loop(window, 0, nil, func(*int, struct{}) Cmd[struct{}] { return nil }, func(int) []Subscription[struct{}] {
			panic("subscriptions broken")
		}, nil, nil, nil, func(layout.Context, int, func(struct{})) {})
	}()
	window.events <- runtimeFrameEvent()
	err := receiveRuntimeTestValue(t, result)
	var panicErr *RuntimePanicError
	if !errors.As(err, &panicErr) || panicErr.Phase != RuntimePhaseSubscriptions || len(panicErr.Stack) == 0 {
		t.Fatalf("subscription panic = %#v", err)
	}
}

func TestLoopReturnsViewPanic(t *testing.T) {
	window := newRuntimeTestWindow()
	result := make(chan error, 1)
	go func() {
		result <- Loop(window, 0, nil, func(*int, struct{}) Cmd[struct{}] { return nil }, nil, nil, nil, nil, func(layout.Context, int, func(struct{})) {
			panic("view broken")
		})
	}()
	window.events <- runtimeFrameEvent()
	err := receiveRuntimeTestValue(t, result)
	var panicErr *RuntimePanicError
	if !errors.As(err, &panicErr) || panicErr.Phase != RuntimePhaseView || len(panicErr.Stack) == 0 {
		t.Fatalf("view panic = %#v", err)
	}
}

func TestLoopDerivesSubscriptionsOnlyAfterModelUpdates(t *testing.T) {
	window := newRuntimeTestWindow()
	viewed := make(chan struct{}, 3)
	var derived atomic.Int32
	result := make(chan error, 1)
	go func() {
		result <- Loop(window, 0, nil, func(model *int, msg int) Cmd[int] {
			*model = msg
			return nil
		}, func(int) []Subscription[int] {
			derived.Add(1)
			return []Subscription[int]{{Key: "events", Run: func(ctx context.Context, _ func(int)) error {
				<-ctx.Done()
				return ctx.Err()
			}}}
		}, nil, nil, func(_ app.Config, send func(int)) { send(1) }, func(layout.Context, int, func(int)) { viewed <- struct{}{} })
	}()
	window.events <- runtimeFrameEvent()
	receiveRuntimeTestValue(t, viewed)
	window.events <- runtimeFrameEvent()
	receiveRuntimeTestValue(t, viewed)
	if got := derived.Load(); got != 1 {
		t.Fatalf("subscription derivations after idle frames = %d, want 1", got)
	}
	window.events <- app.ConfigEvent{}
	window.events <- runtimeFrameEvent()
	receiveRuntimeTestValue(t, viewed)
	if got := derived.Load(); got != 2 {
		t.Fatalf("subscription derivations after model update = %d, want 2", got)
	}
	window.events <- app.DestroyEvent{}
	if err := receiveRuntimeTestValue(t, result); err != nil {
		t.Fatalf("loop returned error: %v", err)
	}
}

func TestLoopReportsBoundedMessageQueueDrops(t *testing.T) {
	window := newRuntimeTestWindow()
	viewed := make(chan struct{}, 1)
	errorsReported := make(chan error, 1)
	result := make(chan error, 1)
	var updates atomic.Int32
	go func() {
		result <- Loop(window, 0, nil, func(model *int, _ int) Cmd[int] {
			updates.Add(1)
			(*model)++
			return nil
		}, nil, func(err error) {
			if _, ok := err.(*QueueOverflowError); ok {
				errorsReported <- err
			}
		}, nil, func(_ app.Config, send func(int)) {
			for index := range messageQueueLimit + 4 {
				send(index)
			}
		}, func(layout.Context, int, func(int)) { viewed <- struct{}{} })
	}()
	window.events <- app.ConfigEvent{}
	window.events <- runtimeFrameEvent()
	receiveRuntimeTestValue(t, viewed)
	overflow := receiveRuntimeTestValue(t, errorsReported)
	var overflowErr *QueueOverflowError
	if !errors.As(overflow, &overflowErr) || overflowErr.Dropped != 4 || updates.Load() != messageQueueLimit {
		t.Fatalf("queue overflow = %#v, updates = %d", overflow, updates.Load())
	}
	window.events <- app.DestroyEvent{}
	if err := receiveRuntimeTestValue(t, result); err != nil {
		t.Fatalf("loop returned error: %v", err)
	}
}

func TestLoopLatestCmdDropsReplacedResultFromDispatchQueue(t *testing.T) {
	type model struct{ result int }
	window := newRuntimeTestWindow()
	viewed := make(chan model, 3)
	oldStarted := make(chan struct{})
	oldCanceled := make(chan struct{})
	newDone := make(chan struct{})
	var query atomic.Int32
	result := make(chan error, 1)
	go func() {
		result <- Loop(window, model{}, nil, func(model *model, msg int) Cmd[int] {
			switch msg {
			case 1:
				return LatestCmd("search", func(ctx context.Context, send func(int)) error {
					close(oldStarted)
					<-ctx.Done()
					send(101)
					close(oldCanceled)
					return ctx.Err()
				})
			case 2:
				return LatestCmd("search", func(_ context.Context, send func(int)) error {
					send(202)
					close(newDone)
					return nil
				})
			default:
				model.result = msg
				return nil
			}
		}, nil, nil, nil, func(_ app.Config, send func(int)) { send(int(query.Load())) }, func(_ layout.Context, model model, _ func(int)) { viewed <- model })
	}()
	query.Store(1)
	window.events <- app.ConfigEvent{}
	window.events <- runtimeFrameEvent()
	receiveRuntimeTestValue(t, viewed)
	receiveRuntimeTestValue(t, oldStarted)
	query.Store(2)
	window.events <- app.ConfigEvent{}
	window.events <- runtimeFrameEvent()
	receiveRuntimeTestValue(t, viewed)
	receiveRuntimeTestValue(t, oldCanceled)
	receiveRuntimeTestValue(t, newDone)
	window.events <- runtimeFrameEvent()
	if got := receiveRuntimeTestValue(t, viewed).result; got != 202 {
		t.Fatalf("latest result = %d, want 202", got)
	}
	window.events <- app.DestroyEvent{}
	if err := receiveRuntimeTestValue(t, result); err != nil {
		t.Fatalf("loop returned error: %v", err)
	}
}

func TestLoopQueuesConfigMessagesForNextFrame(t *testing.T) {
	window := newRuntimeTestWindow()
	viewed := make(chan image.Point, 1)
	result := make(chan error, 1)
	go func() {
		result <- Loop(
			window,
			image.Point{},
			nil,
			func(model *image.Point, msg image.Point) Cmd[image.Point] {
				*model = msg
				return nil
			},
			nil,
			nil,
			nil,
			func(config app.Config, send func(image.Point)) { send(config.Size) },
			func(_ layout.Context, model image.Point, _ func(image.Point)) { viewed <- model },
		)
	}()

	want := image.Pt(800, 600)
	window.events <- app.ConfigEvent{Config: app.Config{Size: want}}
	window.events <- runtimeFrameEvent()
	if got := receiveRuntimeTestValue(t, viewed); got != want {
		t.Fatalf("config message model = %v, want %v", got, want)
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
