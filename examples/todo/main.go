package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	NextID int
	Draft  string
	Items  []Item
}

type Item struct {
	ID   int
	Text string
	Done bool
}

type Msg any

type DraftChanged struct {
	Text string
}

type Add struct{}

type SetDone struct {
	ID   int
	Done bool
}

type Delete struct {
	ID int
}

func Update(m *Model, msg Msg) {
	switch msg := msg.(type) {
	case DraftChanged:
		m.Draft = msg.Text
	case Add:
		text := strings.TrimSpace(m.Draft)
		if text == "" {
			return
		}
		m.Items = append(m.Items, Item{
			ID:   m.NextID,
			Text: text,
		})
		m.NextID++
		m.Draft = ""
	case SetDone:
		for i := range m.Items {
			if m.Items[i].ID == msg.ID {
				m.Items[i].Done = msg.Done
				return
			}
		}
	case Delete:
		m.Items = slices.DeleteFunc(m.Items, func(item Item) bool {
			return item.ID == msg.ID
		})
	}
}

func View(ctx *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	canAdd := strings.TrimSpace(m.Draft) != ""
	children := []ui.Widget{
		ui.Text("FlowUI Todo").Size(24),
		addControls(m, canAdd, send),
	}

	if len(m.Items) == 0 {
		children = append(children,
			ui.Box(
				ui.Text("No todos yet.").Size(16),
			).Style(ui.FillWidth().Height(160)).Align(ui.AlignCenter),
		)
	} else {
		children = append(children,
			ui.Box(
				ui.List("todos", len(m.Items), func(i int) ui.Widget {
					item := m.Items[i]
					return ui.Key(fmt.Sprintf("todo:%d", item.ID), todoRow(item, send))
				}).Gap(12),
			).Style(ui.Height(360)),
		)
	}

	return ui.Center(
		ui.Box(
			ui.Column(children...).Gap(12),
		).Style(ui.FillWidth().MaxWidth(640).Padding(24)),
	)
}

func addControls(m Model, canAdd bool, send ui.Send[Msg]) ui.Widget {
	input := ui.Box(
		ui.Input("draft", m.Draft).
			Hint("What needs to be done?").
			OnChange(func(text string) {
				send(DraftChanged{Text: text})
			}).
			OnSubmit(func(string) {
				send(Add{})
			}),
	).Style(ui.FillWidth())

	button := ui.Button("add", ui.Text("Add")).
		Disabled(!canAdd).
		OnClick(func() {
			send(Add{})
		})

	return ui.Row(
		ui.Expanded(input),
		ui.Box(button).Style(ui.Width(100)),
	).Gap(12).AlignMiddle()
}

func todoRow(item Item, send ui.Send[Msg]) ui.Widget {
	checkbox := ui.Box(
		ui.Checkbox("done", item.Done, item.Text).OnChange(func(done bool) {
			send(SetDone{ID: item.ID, Done: done})
		}),
	).Style(ui.FillWidth())

	button := ui.Button("delete", ui.Text("Delete")).OnClick(func() {
		send(Delete{ID: item.ID})
	}).Variant(ui.ButtonDangerSoft)

	return ui.Row(
		ui.Expanded(checkbox),
		ui.Box(button).Style(ui.Width(100)),
	).Gap(12).AlignMiddle()
}

func main() {
	ui.Run(Model{NextID: 1}, Update, View,
		ui.Title("FlowUI Todo"),
		ui.Size(900, 600),
	)
}
