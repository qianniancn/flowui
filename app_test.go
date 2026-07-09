package flowui

import (
	"testing"
	"time"

	flowruntime "github.com/qianniancn/FlowUI/runtime"
)

func TestMessageQueueDrainsInOrder(t *testing.T) {
	var q flowruntime.Queue[int]
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
	startCmd(Do(func(send Send[int]) {
		send(7)
	}), func(msg int) {
		done <- msg
	})

	select {
	case got := <-done:
		if got != 7 {
			t.Fatalf("message = %d, want 7", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for command")
	}
}
