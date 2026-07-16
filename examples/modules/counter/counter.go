// Package counter is a self-contained child MVU module.
package counter

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Count   int
	Loading bool
}

type Msg any

type Increment struct{}
type Load struct{}
type Loaded struct{ Count int }

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Increment:
		model.Count++
	case Load:
		if model.Loading {
			return nil
		}
		model.Loading = true
		return ui.Do(func(send ui.Send[Msg]) {
			send(Loaded{Count: 10})
		})
	case Loaded:
		model.Count = msg.Count
		model.Loading = false
	}
	return nil
}

func View(model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Column(
		ui.Text(fmt.Sprintf("Count: %d", model.Count)).Size(24),
		ui.Row(
			ui.Button("increment", ui.Text("Increment")).OnClick(func() {
				send(Increment{})
			}),
			ui.Button("load", ui.Text("Load saved value")).
				Variant(ui.ButtonSecondary).
				Loading(model.Loading).
				OnClick(func() {
					send(Load{})
				}),
		).Gap(12),
	).Gap(16)
}
