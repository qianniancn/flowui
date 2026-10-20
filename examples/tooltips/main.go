package main

import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/tooltip"
	"github.com/qianniancn/flowui/window"
)

type Model struct{}
type Msg any

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("FlowUI Tooltips").Size(24),
				layout.Divider(),
				section("Basic",
					layout.Row(
						tooltipButton("basic", "Hover me", "Tooltip content"),
						tooltipButton("arrow", "With arrow", "Tooltip content").Arrow(true),
						tooltip.New("disabled", button.New("Disabled tooltip").Variant(button.Secondary), "This tooltip is disabled").Delay(0).Disabled(true),
					).Gap(12).AlignMiddle(),
				),
				section("Placement",
					layout.Wrap(
						layout.Box(tooltipButton("top", "Top", "Placed above").Placement(tooltip.Top).Arrow(true)),
						layout.Box(tooltipButton("bottom", "Bottom", "Placed below").Placement(tooltip.Bottom).Arrow(true)),
						layout.Box(tooltipButton("left", "Left", "Placed on the left").Placement(tooltip.Left).Arrow(true)),
						layout.Box(tooltipButton("right", "Right", "Placed on the right").Placement(tooltip.Right).Arrow(true)),
						layout.Box(tooltipButton("top-start", "Top start", "Aligned to the start").Placement(tooltip.TopStart).Arrow(true)),
						layout.Box(tooltipButton("bottom-end", "Bottom end", "Aligned to the end").Placement(tooltip.BottomEnd).Arrow(true)),
					).Gap(12).LineGap(12).AlignMiddle(),
				),
				section("Delay",
					layout.Row(
						tooltip.New("default-delay", button.New("Default delay").Variant(button.Secondary), "Standard delay").Arrow(true),
						tooltip.New("instant", button.New("No delay").Variant(button.Secondary), "Opens immediately").Delay(0).CloseDelay(0).Arrow(true),
					).Gap(12).AlignMiddle(),
				),
			).Gap(20),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func tooltipButton(key, label, content string) tooltip.Widget {
	return tooltip.New(key, button.New(label).Key(key+"-trigger").Variant(button.Secondary), content).Delay(0)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(18),
		child,
	).Gap(12)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Tooltips"),
		window.Size(860, 560),
		window.Theme(base.Light()),
	)
}
