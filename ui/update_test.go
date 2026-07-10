package ui

import (
	"bytes"
	"context"
	"errors"
	"testing"
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
