package ui_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/qianniancn/flowui/ui"
)

func ExampleDo() {
	type loaded struct {
		query string
	}
	type model struct {
		query string
	}

	m := model{query: "FlowUI"}
	query := m.query // Take the immutable snapshot while still in Update.
	cmd := ui.Do(func(send ui.Send[loaded]) {
		// Communicate the result through send; do not access m or a Context here.
		send(loaded{query: query})
	})

	m.query = "changed after Update"
	results := make(chan loaded, 1)
	_ = cmd(context.Background(), func(msg loaded) { results <- msg })
	fmt.Println((<-results).query)

	// Output: FlowUI
}

func ExampleDoContext() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd := ui.DoContext(func(ctx context.Context, _ ui.Send[string]) error {
		<-ctx.Done()
		return ctx.Err()
	})

	err := cmd(ctx, func(string) {})
	fmt.Println(errors.Is(err, context.Canceled))

	// Output: true
}

func ExampleMapCmd() {
	type childLoaded struct {
		count int
	}
	type parentMessage struct {
		loaded childLoaded
	}

	child := ui.Do(func(send ui.Send[childLoaded]) {
		send(childLoaded{count: 3})
	})
	parent := ui.MapCmd(child, func(msg childLoaded) parentMessage {
		return parentMessage{loaded: msg}
	})

	results := make(chan parentMessage, 1)
	_ = parent(context.Background(), func(msg parentMessage) { results <- msg })
	fmt.Println((<-results).loaded.count)

	// Output: 3
}

func ExampleSubscribe() {
	type model struct {
		listening bool
	}
	type message string

	subscriptions := func(m model) []ui.Subscription[message] {
		if !m.listening {
			return nil
		}
		return []ui.Subscription[message]{
			ui.Subscribe("events", func(ctx context.Context, _ ui.Send[message]) error {
				<-ctx.Done()
				return ctx.Err()
			}),
		}
	}

	fmt.Println(len(subscriptions(model{listening: true})))
	fmt.Println(len(subscriptions(model{})))

	// Output:
	// 1
	// 0
}
