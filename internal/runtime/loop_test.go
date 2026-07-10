package runtime

import (
	"slices"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestLoopCoreProcessesBatchBeforeView(t *testing.T) {
	type model struct {
		values []int
	}

	var updated []int
	core := newLoopCore(model{}, func(m *model, msg int) func(func(int)) {
		updated = append(updated, msg)
		m.values = append(m.values, msg)
		return nil
	})
	core.send(1)
	core.send(2)
	core.send(3)

	var viewed model
	core.frame(core.send, func(m model) {
		viewed.values = append([]int(nil), m.values...)
	})

	if want := []int{1, 2, 3}; !slices.Equal(updated, want) {
		t.Fatalf("update order = %v, want %v", updated, want)
	}
	if want := []int{1, 2, 3}; !slices.Equal(viewed.values, want) {
		t.Fatalf("view model = %v, want %v", viewed.values, want)
	}
}

func TestLoopCoreFeedsCommandMessagesIntoNextFrame(t *testing.T) {
	commandSent := make(chan struct{})
	core := newLoopCore([]int(nil), func(model *[]int, msg int) func(func(int)) {
		*model = append(*model, msg)
		if msg != 1 {
			return nil
		}
		return func(send func(int)) {
			send(2)
			close(commandSent)
		}
	})
	core.send(1)

	var firstFrame []int
	core.frame(core.send, func(model []int) {
		firstFrame = append([]int(nil), model...)
	})
	if want := []int{1}; !slices.Equal(firstFrame, want) {
		t.Fatalf("first frame model = %v, want %v", firstFrame, want)
	}

	select {
	case <-commandSent:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for command message")
	}

	var secondFrame []int
	core.frame(core.send, func(model []int) {
		secondFrame = append([]int(nil), model...)
	})
	if want := []int{1, 2}; !slices.Equal(secondFrame, want) {
		t.Fatalf("second frame model = %v, want %v", secondFrame, want)
	}
}

func TestLoopCoreCollectsMessagesFromConcurrentCommands(t *testing.T) {
	const commandCount = 128
	type message struct {
		start bool
		id    int
	}
	type model struct {
		received []int
	}

	gate := make(chan struct{})
	var ready sync.WaitGroup
	var commands sync.WaitGroup
	ready.Add(commandCount)
	commands.Add(commandCount)
	core := newLoopCore(model{}, func(model *model, msg message) func(func(message)) {
		if !msg.start {
			model.received = append(model.received, msg.id)
			return nil
		}
		return func(send func(message)) {
			ready.Done()
			<-gate
			send(message{id: msg.id})
			commands.Done()
		}
	})
	for id := range commandCount {
		core.send(message{start: true, id: id})
	}

	var firstFrame model
	core.frame(core.send, func(model model) {
		firstFrame.received = append([]int(nil), model.received...)
	})
	if len(firstFrame.received) != 0 {
		t.Fatalf("first frame received %v, want no command results", firstFrame.received)
	}

	readyDone := make(chan struct{})
	go func() {
		ready.Wait()
		close(readyDone)
	}()
	select {
	case <-readyDone:
	case <-time.After(time.Second):
		close(gate)
		t.Fatal("timed out waiting for commands to start")
	}

	close(gate)
	done := make(chan struct{})
	go func() {
		commands.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for concurrent commands")
	}

	var secondFrame model
	core.frame(core.send, func(model model) {
		secondFrame.received = append([]int(nil), model.received...)
	})
	sort.Ints(secondFrame.received)
	if len(secondFrame.received) != commandCount {
		t.Fatalf("second frame received %d messages, want %d", len(secondFrame.received), commandCount)
	}
	for id, got := range secondFrame.received {
		if got != id {
			t.Fatalf("second frame message[%d] = %d, want %d", id, got, id)
		}
	}
}

func TestLoopCoreDefersMessagesSentByViewUntilNextFrame(t *testing.T) {
	core := newLoopCore([]int(nil), func(model *[]int, msg int) func(func(int)) {
		*model = append(*model, msg)
		return nil
	})
	core.send(1)
	send := core.send

	var firstFrame []int
	core.frame(send, func(model []int) {
		firstFrame = append([]int(nil), model...)
		send(2)
	})
	if want := []int{1}; !slices.Equal(firstFrame, want) {
		t.Fatalf("first frame model = %v, want %v", firstFrame, want)
	}

	var secondFrame []int
	core.frame(send, func(model []int) {
		secondFrame = append([]int(nil), model...)
	})
	if want := []int{1, 2}; !slices.Equal(secondFrame, want) {
		t.Fatalf("second frame model = %v, want %v", secondFrame, want)
	}
}
