package main

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) {}

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Scroll("chips",
				ui.Column(
					ui.Text("FlowUI Chip").Size(24),
					ui.Divider(),
					section("Default",
						chipRow(
							ui.Chip("Label"),
							ui.Chip("Accent").Color(ui.ChipAccent),
						),
					),
					section("Sizes",
						chipRow(
							ui.Chip("Small").Color(ui.ChipAccent).Size(ui.ChipSmall),
							ui.Chip("Medium").Color(ui.ChipAccent),
							ui.Chip("Large").Color(ui.ChipAccent).Size(ui.ChipLarge),
						),
					),
					section("With icons",
						chipRow(
							ui.Chip("Leading").Color(ui.ChipAccent).
								StartContent(ui.Icon(lucide.CircleDashed).Size(16)),
							ui.Chip("Both sides").Color(ui.ChipAccent).
								StartContent(ui.Icon(lucide.CircleDashed).Size(16)).
								EndContent(ui.Icon(lucide.CircleDashed).Size(16)),
						),
					),
					section("Statuses",
						ui.Column(
							statusRow(ui.ChipPrimary),
							statusRow(ui.ChipSecondary),
							statusRow(ui.ChipTertiary),
							statusRow(ui.ChipSoft),
						).Gap(12),
					),
					section("Variants",
						ui.Column(
							variantRow("Primary", ui.ChipPrimary),
							variantRow("Secondary", ui.ChipSecondary),
							variantRow("Tertiary", ui.ChipTertiary),
							variantRow("Soft", ui.ChipSoft),
						).Gap(14),
					),
				).Gap(20),
			).Vertical(),
		).FillWidth().MaxWidth(860).Padding(24),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func chipRow(chips ...ui.Widget) ui.Widget {
	return ui.Wrap(chips...).Gap(12).LineGap(10).AlignMiddle()
}

func statusRow(variant ui.ChipVariant) ui.Widget {
	return chipRow(
		ui.Chip("Information").Color(ui.ChipAccent).Variant(variant).StartContent(statusDot{size: 6}),
		ui.Chip("Completed").Color(ui.ChipSuccess).Variant(variant).StartContent(statusDot{size: 6}),
		ui.Chip("Pending").Color(ui.ChipWarning).Variant(variant).StartContent(statusDot{size: 6}),
		ui.Chip("Failed").Color(ui.ChipDanger).Variant(variant).StartContent(statusDot{size: 6}),
	)
}

func variantRow(label string, variant ui.ChipVariant) ui.Widget {
	return ui.Row(
		ui.Box(ui.Text(label).Size(13)).Width(82),
		chipRow(
			ui.Chip("Default").Variant(variant),
			ui.Chip("Accent").Color(ui.ChipAccent).Variant(variant),
			ui.Chip("Success").Color(ui.ChipSuccess).Variant(variant),
			ui.Chip("Warning").Color(ui.ChipWarning).Variant(variant),
			ui.Chip("Danger").Color(ui.ChipDanger).Variant(variant),
		),
	).Gap(12).AlignMiddle()
}

type statusDot struct {
	size unit.Dp
}

func (d statusDot) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	diameter := min(gtx.Dp(d.size), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	diameter = max(diameter, 0)
	size := image.Pt(diameter, diameter)
	paint.FillShape(gtx.Ops, ctx.ForegroundColor(), clip.Ellipse(image.Rectangle{Max: size}).Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Chip"),
		ui.Size(980, 760),
	)
}
