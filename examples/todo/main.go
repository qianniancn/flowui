package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	buttonpkg "github.com/qianniancn/flowui/widget/button"
	checkboxpkg "github.com/qianniancn/flowui/widget/checkbox"
	inputpkg "github.com/qianniancn/flowui/widget/input"
	textpkg "github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case DraftChanged:
		m.Draft = msg.Text
	case Add:
		text := strings.TrimSpace(m.Draft)
		if text == "" {
			return nil
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
				return nil
			}
		}
	case Delete:
		m.Items = slices.DeleteFunc(m.Items, func(item Item) bool {
			return item.ID == msg.ID
		})
	}
	return nil
}

func View(ctx *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	canAdd := strings.TrimSpace(m.Draft) != ""
	children := []ui.Widget{
		textpkg.New("FlowUI Todo").Size(24),
		addControls(m, canAdd, send),
	}

	if len(m.Items) == 0 {
		children = append(children,
			layout.Box(
				textpkg.New("No todos yet.").Size(16),
			).Style(style.New().FillWidth().Height(160)).Align(layout.AlignCenter),
		)
	} else {
		children = append(children,
			layout.Box(
				layout.List(len(m.Items), func(i int) ui.Widget {
					item := m.Items[i]
					return layout.Key(fmt.Sprintf("todo:%d", item.ID), todoRow(item, send))
				}).Gap(12),
			).Style(style.New().Height(360)),
		)
	}

	return layout.Center(
		layout.Box(
			layout.Column(children...).Gap(12),
		).Style(style.New().FillWidth().MaxWidth(640).Padding(24)),
	)
}

func addControls(m Model, canAdd bool, send ui.Send[Msg]) ui.Widget {
	input := layout.Box(
		inputpkg.New(m.Draft).Key("draft").
			Hint("What needs to be done?").
			OnChange(func(text string) {
				send(DraftChanged{Text: text})
			}).
			OnSubmit(func(string) {
				send(Add{})
			}),
	).Style(style.New().FillWidth())

	button := buttonpkg.New("Add").
		Disabled(!canAdd).
		OnClick(func() {
			send(Add{})
		})

	return layout.Row(
		layout.Expanded(input),
		layout.Box(button).Style(style.New().Width(100)),
	).Gap(12).AlignMiddle()
}

func todoRow(item Item, send ui.Send[Msg]) ui.Widget {
	checkbox := layout.Box(
		checkboxpkg.New(item.Text).Key("done").Checked(item.Done).OnChange(func(done bool) {
			send(SetDone{ID: item.ID, Done: done})
		}),
	).Style(style.New().FillWidth())

	button := buttonpkg.New("Delete").OnClick(func() {
		send(Delete{ID: item.ID})
	}).Variant(buttonpkg.DangerSoft)

	return layout.Row(
		layout.Expanded(checkbox),
		layout.Box(button).Style(style.New().Width(100)),
	).Gap(12).AlignMiddle()
}

func main() {
	ui.Run(ui.NewProgram(Model{NextID: 1},
		Update, View), window.Title("FlowUI Todo"),
		window.Size(900, 600),
		window.Theme(base.Light()),
	)
}
