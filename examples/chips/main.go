package main

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return flowlayout.Center(
		flowlayout.Box(
			flowlayout.Scroll(flowlayout.Column(
				text.New("FlowUI Chip").Size(24),
				flowlayout.Divider(),
				section("Default",
					chipRow(
						chip.New("Label"),
						chip.New("Accent").Color(chip.ColorAccent),
					),
				),
				section("Sizes",
					chipRow(
						chip.New("Small").Color(chip.ColorAccent).Size(chip.SizeSmall),
						chip.New("Medium").Color(chip.ColorAccent),
						chip.New("Large").Color(chip.ColorAccent).Size(chip.SizeLarge),
					),
				),
				section("With icons",
					chipRow(
						chip.New("Leading").Color(chip.ColorAccent).
							StartContent(icon.New(lucide.CircleDashed).Size(16)),
						chip.New("Both sides").Color(chip.ColorAccent).
							StartContent(icon.New(lucide.CircleDashed).Size(16)).
							EndContent(icon.New(lucide.CircleDashed).Size(16)),
					),
				),
				section("Statuses",
					flowlayout.Column(
						statusRow(chip.VariantPrimary),
						statusRow(chip.VariantSecondary),
						statusRow(chip.VariantTertiary),
						statusRow(chip.VariantSoft),
					).Gap(12),
				),
				section("Variants",
					flowlayout.Column(
						variantRow("Primary", chip.VariantPrimary),
						variantRow("Secondary", chip.VariantSecondary),
						variantRow("Tertiary", chip.VariantTertiary),
						variantRow("Soft", chip.VariantSoft),
					).Gap(14),
				),
			).Gap(20),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(860).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return flowlayout.Column(text.New(title).Size(18), child).Gap(10)
}

func chipRow(chips ...ui.Widget) ui.Widget {
	return flowlayout.Wrap(chips...).Gap(12).LineGap(10).AlignMiddle()
}

func statusRow(variant chip.Variant) ui.Widget {
	return chipRow(
		chip.New("Information").Color(chip.ColorAccent).Variant(variant).StartContent(statusDot{size: 6}),
		chip.New("Completed").Color(chip.ColorSuccess).Variant(variant).StartContent(statusDot{size: 6}),
		chip.New("Pending").Color(chip.ColorWarning).Variant(variant).StartContent(statusDot{size: 6}),
		chip.New("Failed").Color(chip.ColorDanger).Variant(variant).StartContent(statusDot{size: 6}),
	)
}

func variantRow(label string, variant chip.Variant) ui.Widget {
	return flowlayout.Row(
		flowlayout.Box(text.New(label).Size(13)).Style(style.New().Width(82)),
		chipRow(
			chip.New("Default").Variant(variant),
			chip.New("Accent").Color(chip.ColorAccent).Variant(variant),
			chip.New("Success").Color(chip.ColorSuccess).Variant(variant),
			chip.New("Warning").Color(chip.ColorWarning).Variant(variant),
			chip.New("Danger").Color(chip.ColorDanger).Variant(variant),
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
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Chip"),
		window.Size(980, 760),
		window.Theme(base.Light()),
	)
}
