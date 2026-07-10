package runtime

import (
	"context"
	"testing"
	"time"
)

func TestQueueDrainsInOrder(t *testing.T) {
	var q Queue[int]
	q.Push(1)
	q.Push(2)

	var got []int
	q.Drain(func(msg int) {
		got = append(got, msg)
	})
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("drained messages = %v, want [1 2]", got)
	}
	q.Drain(func(msg int) {
		t.Fatalf("unexpected message %d", msg)
	})
}

func TestStartCmdSendsMessage(t *testing.T) {
	done := make(chan int, 1)
	var effects effectGroup
	StartCmd(&effects, context.Background(), func(_ context.Context, send func(int)) error {
		send(7)
		return nil
	}, func(msg int) {
		done <- msg
	}, nil)

	select {
	case got := <-done:
		if got != 7 {
			t.Fatalf("message = %d, want 7", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for command")
	}
}
