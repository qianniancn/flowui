package ui_test

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
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
	cmd(func(msg loaded) { results <- msg })
	fmt.Println((<-results).query)

	// Output: FlowUI
}
