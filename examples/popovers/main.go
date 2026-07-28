package main

import (
	"fmt"

	"github.com/qianniancn/flowui/ui"
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

func Update(m *Model, msg Msg) {
	switch msg := msg.(type) {
	case Toggle:
		if m.Active == msg.Key {
			m.Active = ""
			return
		}
		m.Active = msg.Key
	case Close:
		m.Active = ""
	case Pick:
		m.Choice = msg.Value
		m.Last = fmt.Sprintf("Selected %s", msg.Value)
		m.Active = ""
	}
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No popover open"
	if m.Active != "" {
		status = fmt.Sprintf("Open: %s", m.Active)
	} else if m.Last != "" {
		status = m.Last
	}

	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Popover").Size(24),
				ui.Text(status).Size(16),
				ui.Divider(),
				section("Basic",
					ui.Row(
						popoverButton("basic", "Open popover", m, send,
							ui.Text("Popover content uses the overlay surface, soft shadow, rounded corners, and fade/zoom animation."),
						),
						popoverButton("arrow", "With arrow", m, send,
							ui.Text("The arrow points back to the trigger.").Size(14),
						).Arrow(true),
					).Gap(12).AlignMiddle(),
				),
				section("Placement",
					ui.Wrap(
						ui.Box(popoverButton("top", "Top", m, send, placementContent("top")).Placement(ui.PopoverTop).ShouldFlip(false).Arrow(true)),
						ui.Box(popoverButton("bottom", "Bottom", m, send, placementContent("bottom")).Placement(ui.PopoverBottom).Arrow(true)),
						ui.Box(popoverButton("left", "Left", m, send, placementContent("left")).Placement(ui.PopoverLeft).ShouldFlip(false).Arrow(true)),
						ui.Box(popoverButton("right", "Right", m, send, placementContent("right")).Placement(ui.PopoverRight).Arrow(true)),
						ui.Box(popoverButton("start", "Bottom start", m, send, placementContent("bottom start")).Placement(ui.PopoverBottomStart).Arrow(true)),
						ui.Box(popoverButton("end", "Bottom end", m, send, placementContent("bottom end")).Placement(ui.PopoverBottomEnd).Arrow(true)),
					).Gap(12).LineGap(12).AlignMiddle(),
				),
				section("Flip and overflow",
					ui.Row(
						popoverButton("flip", "Bottom flips near edge", m, send,
							ui.Text("This popover requests bottom placement. When the current layout has no room below, it resolves upward."),
						).Placement(ui.PopoverBottom).Arrow(true),
						popoverButton("no-flip", "Flip disabled", m, send,
							ui.Text("This popover keeps its requested placement even if space is tight."),
						).Placement(ui.PopoverBottom).ShouldFlip(false).Arrow(true),
					).Gap(12).AlignMiddle(),
				),
				section("Interactive",
					ui.Row(
						popoverButton("choices", "Pick one", m, send, choicesContent(m, send)).
							Heading("Quick action").
							Arrow(true),
						popoverButton("locked", "Manual close", m, send,
							ui.Column(
								ui.Text("Backdrop clicks are ignored for this popover."),
								ui.Button("locked-close", ui.Text("Close")).Variant(ui.ButtonSecondary).OnClick(func() {
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
		).Style(ui.FillWidth()).Style(ui.MaxWidth(760)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func popoverButton(key, label string, m Model, send ui.Send[Msg], content ui.Widget) ui.PopoverWidget {
	return ui.Popover(
		key,
		m.Active == key,
		ui.Button("trigger-"+key, ui.Text(label)).
			Variant(ui.ButtonSecondary).
			OnClick(func() {
				send(Toggle{Key: key})
			}),
		ui.Box(content).Style(ui.Width(260)),
	).OnOpenChange(func(open bool) {
		if !open {
			send(Close{})
		}
	})
}

func placementContent(name string) ui.Widget {
	return ui.Text(fmt.Sprintf("This popover is placed on the %s side of its trigger.", name))
}

func choicesContent(m Model, send ui.Send[Msg]) ui.Widget {
	current := "Nothing selected"
	if m.Choice != "" {
		current = "Current: " + m.Choice
	}
	return ui.Column(
		ui.Text(current),
		ui.Row(
			ui.Button("choice-alpha", ui.Text("Alpha")).OnClick(func() {
				send(Pick{Value: "Alpha"})
			}),
			ui.Button("choice-beta", ui.Text("Beta")).Variant(ui.ButtonSecondary).OnClick(func() {
				send(Pick{Value: "Beta"})
			}),
		).Gap(8).AlignMiddle(),
	).Gap(12)
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Popover"),
		ui.Size(900, 640),
	)
}
