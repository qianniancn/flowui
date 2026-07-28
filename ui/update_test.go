package ui

import (
	"bytes"
	"context"
	"errors"
	"testing"

	internalexplorer "github.com/qianniancn/flowui/internal/explorer"
)

func TestDoAdaptsContextFreeCommand(t *testing.T) {
	var got int
	cmd := Do(func(send Send[int]) {
		send(7)
	})
	if err := cmd(context.Background(), func(msg int) { got = msg }); err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if got != 7 {
		t.Fatalf("message = %d, want 7", got)
	}
}

func TestDoContextReturnsError(t *testing.T) {
	want := errors.New("failed")
	cmd := DoContext(func(context.Context, Send[int]) error {
		return want
	})
	if err := cmd(context.Background(), func(int) {}); !errors.Is(err, want) {
		t.Fatalf("command error = %v, want %v", err, want)
	}
}

func TestRuntimeCmdBindsWindowExplorerService(t *testing.T) {
	service := internalexplorer.New(nil)
	var got *internalexplorer.Service
	cmd := runtimeCmd(func(ctx context.Context, _ Send[struct{}]) error {
		got = internalexplorer.FromContext(ctx)
		return nil
	}, service)
	if err := cmd(context.Background(), func(struct{}) {}); err != nil {
		t.Fatal(err)
	}
	if got != service {
		t.Fatalf("explorer service = %p, want %p", got, service)
	}
}

func TestMapCmdMapsMessagesAndPreservesContextAndError(t *testing.T) {
	type contextKey struct{}
	type parentMessage struct{ value int }
	wantErr := errors.New("failed")
	ctx := context.WithValue(context.Background(), contextKey{}, "value")
	child := func(cmdCtx context.Context, send Send[int]) error {
		if cmdCtx != ctx || cmdCtx.Value(contextKey{}) != "value" {
			t.Fatal("mapped command did not preserve context")
		}
		send(2)
		send(3)
		return wantErr
	}
	mapped := MapCmd(child, func(msg int) parentMessage { return parentMessage{value: msg} })

	var got []int
	err := mapped(ctx, func(msg parentMessage) { got = append(got, msg.value) })
	if !errors.Is(err, wantErr) {
		t.Fatalf("mapped error = %v, want %v", err, wantErr)
	}
	if len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("mapped messages = %v, want [2 3]", got)
	}
}

func TestMapCmdPreservesNilAndRejectsNilMapper(t *testing.T) {
	if MapCmd[int, string](nil, nil) != nil {
		t.Fatal("nil child command did not remain nil")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("non-nil command accepted a nil mapper")
		}
	}()
	MapCmd(Do(func(Send[int]) {}), (func(int) string)(nil))
}

func TestSubscribeValidatesDefinition(t *testing.T) {
	tests := []struct {
		name string
		make func()
	}{
		{name: "empty key", make: func() { Subscribe[int]("", func(context.Context, Send[int]) error { return nil }) }},
		{name: "nil runner", make: func() { Subscribe[int]("events", nil) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid subscription did not panic")
				}
			}()
			test.make()
		})
	}
}

func TestWriteEffectErrorIncludesPanicStack(t *testing.T) {
	var output bytes.Buffer
	writeEffectError(&output, &EffectError{
		Kind:  EffectSubscription,
		Key:   "events",
		Panic: "broken",
		Stack: []byte("stack trace\n"),
	})
	if got := output.String(); got != "flowui: subscription \"events\" panicked: broken\nstack trace\n" {
		t.Fatalf("error output = %q", got)
	}
}

func TestWriteRuntimePanicIncludesStack(t *testing.T) {
	var output bytes.Buffer
	writeEffectError(&output, &RuntimePanicError{
		Phase: RuntimePhaseView,
		Panic: "broken",
		Stack: []byte("stack trace\n"),
	})
	if got := output.String(); got != "flowui: view panicked: broken\nstack trace\n" {
		t.Fatalf("runtime panic output = %q", got)
	}
}
