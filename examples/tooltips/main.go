package main

import "github.com/qianniancn/flowui/ui"

type Model struct{}
type Msg any

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Tooltips").Size(24),
				ui.Divider(),
				section("Basic",
					ui.Row(
						tooltipButton("basic", "Hover me", "Tooltip content"),
						tooltipButton("arrow", "With arrow", "Tooltip content").Arrow(true),
						ui.Tooltip("disabled",
							ui.Button("disabled-trigger", ui.Text("Disabled tooltip")).Variant(ui.ButtonSecondary),
							ui.Text("This tooltip is disabled"),
						).Delay(0).Disabled(true),
					).Gap(12).AlignMiddle(),
				),
				section("Placement",
					ui.Wrap(
						ui.Box(tooltipButton("top", "Top", "Placed above").Placement(ui.TooltipTop).Arrow(true)),
						ui.Box(tooltipButton("bottom", "Bottom", "Placed below").Placement(ui.TooltipBottom).Arrow(true)),
						ui.Box(tooltipButton("left", "Left", "Placed on the left").Placement(ui.TooltipLeft).Arrow(true)),
						ui.Box(tooltipButton("right", "Right", "Placed on the right").Placement(ui.TooltipRight).Arrow(true)),
						ui.Box(tooltipButton("top-start", "Top start", "Aligned to the start").Placement(ui.TooltipTopStart).Arrow(true)),
						ui.Box(tooltipButton("bottom-end", "Bottom end", "Aligned to the end").Placement(ui.TooltipBottomEnd).Arrow(true)),
					).Gap(12).LineGap(12).AlignMiddle(),
				),
				section("Delay",
					ui.Row(
						ui.Tooltip("default-delay",
							ui.Button("default-delay-trigger", ui.Text("Default delay")).Variant(ui.ButtonSecondary),
							ui.Text("HeroUI default delay"),
						).Arrow(true),
						ui.Tooltip("instant",
							ui.Button("instant-trigger", ui.Text("No delay")).Variant(ui.ButtonSecondary),
							ui.Text("Opens immediately"),
						).Delay(0).CloseDelay(0).Arrow(true),
					).Gap(12).AlignMiddle(),
				),
			).Gap(20),
		).Style(ui.FillWidth().MaxWidth(760).Padding(24)),
	)
}

func tooltipButton(key, label, content string) ui.TooltipWidget {
	return ui.Tooltip(
		key,
		ui.Button(key+"-trigger", ui.Text(label)).Variant(ui.ButtonSecondary),
		ui.Text(content),
	).Delay(0)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(12)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), ui.Title("FlowUI Tooltips"),
		ui.Size(860, 560),
	)
}
