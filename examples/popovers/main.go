package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/popover"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Active string
	Choice string
	Last   string
}

type Msg any

type Toggle struct {
	Key string
}

type Close struct{}

type Pick struct {
	Value string
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Toggle:
		if m.Active == msg.Key {
			m.Active = ""
			return nil
		}
		m.Active = msg.Key
	case Close:
		m.Active = ""
	case Pick:
		m.Choice = msg.Value
		m.Last = fmt.Sprintf("Selected %s", msg.Value)
		m.Active = ""
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No popover open"
	if m.Active != "" {
		status = fmt.Sprintf("Open: %s", m.Active)
	} else if m.Last != "" {
		status = m.Last
	}

	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("FlowUI Popover").Size(24),
				text.New(status).Size(16),
				layout.Divider(),
				section("Basic",
					layout.Row(
						popoverButton("basic", "Open popover", m, send,
							text.New("Popover content uses the overlay surface, soft shadow, rounded corners, and fade/zoom animation."),
						),
						popoverButton("arrow", "With arrow", m, send,
							text.New("The arrow points back to the trigger.").Size(14),
						).Arrow(true),
					).Gap(12).AlignMiddle(),
				),
				section("Placement",
					layout.Wrap(
						layout.Box(popoverButton("top", "Top", m, send, placementContent("top")).Placement(popover.Top).ShouldFlip(false).Arrow(true)),
						layout.Box(popoverButton("bottom", "Bottom", m, send, placementContent("bottom")).Placement(popover.Bottom).Arrow(true)),
						layout.Box(popoverButton("left", "Left", m, send, placementContent("left")).Placement(popover.Left).ShouldFlip(false).Arrow(true)),
						layout.Box(popoverButton("right", "Right", m, send, placementContent("right")).Placement(popover.Right).Arrow(true)),
						layout.Box(popoverButton("start", "Bottom start", m, send, placementContent("bottom start")).Placement(popover.BottomStart).Arrow(true)),
						layout.Box(popoverButton("end", "Bottom end", m, send, placementContent("bottom end")).Placement(popover.BottomEnd).Arrow(true)),
					).Gap(12).LineGap(12).AlignMiddle(),
				),
				section("Flip and overflow",
					layout.Row(
						popoverButton("flip", "Bottom flips near edge", m, send,
							text.New("This popover requests bottom placement. When the current layout has no room below, it resolves upward."),
						).Placement(popover.Bottom).Arrow(true),
						popoverButton("no-flip", "Flip disabled", m, send,
							text.New("This popover keeps its requested placement even if space is tight."),
						).Placement(popover.Bottom).ShouldFlip(false).Arrow(true),
					).Gap(12).AlignMiddle(),
				),
				section("Interactive",
					layout.Row(
						popoverButton("choices", "Pick one", m, send, choicesContent(m, send)).
							Heading("Quick action").
							Arrow(true),
						popoverButton("locked", "Manual close", m, send,
							layout.Column(
								text.New("Backdrop clicks are ignored for this popover."),
								button.New("Close").Variant(button.Secondary).OnClick(func() {
									send(Close{})
								}),
							).Gap(12),
						).
							Heading("Dismiss disabled").
							Dismissable(false).
							Arrow(true),
					).Gap(12).AlignMiddle(),
				),
			).Gap(18),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(18),
		child,
	).Gap(10)
}

func popoverButton(key, label string, m Model, send ui.Send[Msg], content ui.Widget) popover.Widget {
	return popover.New(button.New(label).Key("trigger-"+key).
		Variant(button.Secondary).
		OnClick(func() {
			send(Toggle{Key: key})
		}), layout.Box(content).Style(style.New().Width(260))).Key(key).Open(m.Active == key).OnOpenChange(func(open bool) {
		if !open {
			send(Close{})
		}
	})
}

func placementContent(name string) ui.Widget {
	return text.New(fmt.Sprintf("This popover is placed on the %s side of its trigger.", name))
}

func choicesContent(m Model, send ui.Send[Msg]) ui.Widget {
	current := "Nothing selected"
	if m.Choice != "" {
		current = "Current: " + m.Choice
	}
	return layout.Column(
		text.New(current),
		layout.Row(
			button.New("Alpha").OnClick(func() {
				send(Pick{Value: "Alpha"})
			}),
			button.New("Beta").Variant(button.Secondary).OnClick(func() {
				send(Pick{Value: "Beta"})
			}),
		).Gap(8).AlignMiddle(),
	).Gap(12)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Popover"),
		window.Size(900, 640),
		window.Theme(base.Light()),
	)
}
