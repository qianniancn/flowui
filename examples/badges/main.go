package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	avatarpkg "github.com/qianniancn/flowui/widget/avatar"
	"github.com/qianniancn/flowui/widget/badge"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct{}
type Msg any

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Badge").Size(24),
				section("Basic",
					layout.Row(
						badge.New(avatar("JD"), "5").Color(badge.ColorDanger).Size(badge.SizeSmall),
						badge.New(avatar("AB"), "New").Color(badge.ColorAccent).Size(badge.SizeSmall),
						badge.New(avatar("CD"), "").Color(badge.ColorSuccess).Size(badge.SizeSmall).Placement(badge.PlacementBottomRight).Alt("Online"),
						badge.New(avatar("EF"), "").Content(icon.New(lucide.Bell).Size(10)).Color(badge.ColorAccent).Size(badge.SizeSmall).Alt("Notifications"),
					).Gap(28).AlignMiddle(),
				),
				section("Sizes",
					layout.Row(
						badge.New(avatarpkg.New("SM").Size(avatarpkg.SizeSmall), "5").Color(badge.ColorDanger).Size(badge.SizeSmall),
						badge.New(avatarpkg.New("MD").Size(avatarpkg.SizeMedium), "5").Color(badge.ColorDanger).Size(badge.SizeMedium),
						badge.New(avatarpkg.New("LG").Size(avatarpkg.SizeLarge), "5").Color(badge.ColorDanger).Size(badge.SizeLarge),
					).Gap(32).AlignMiddle(),
				),
				section("Colors",
					layout.Row(
						coloredDot(badge.ColorDefault),
						coloredDot(badge.ColorAccent),
						coloredDot(badge.ColorSuccess),
						coloredDot(badge.ColorWarning),
						coloredDot(badge.ColorDanger),
					).Gap(28).AlignMiddle(),
				),
				section("Variants",
					layout.Row(
						badge.New(avatar("P"), "5").Color(badge.ColorAccent).Size(badge.SizeSmall).Variant(badge.VariantPrimary),
						badge.New(avatar("S"), "5").Color(badge.ColorAccent).Size(badge.SizeSmall).Variant(badge.VariantSecondary),
						badge.New(avatar("SF"), "5").Color(badge.ColorAccent).Size(badge.SizeSmall).Variant(badge.VariantSoft),
					).Gap(28).AlignMiddle(),
				),
				section("Placements",
					layout.Row(
						placedDot(badge.PlacementTopLeft),
						placedDot(badge.PlacementTopRight),
						placedDot(badge.PlacementBottomLeft),
						placedDot(badge.PlacementBottomRight),
					).Gap(28).AlignMiddle(),
				),
			).Gap(20),
		).Style(style.New().FillWidth().MaxWidth(620).Padding(24)),
	)
}

func avatar(label string) ui.Widget {
	return avatarpkg.New(label).Color(avatarpkg.ColorAccent).Variant(avatarpkg.VariantSoft)
}

func coloredDot(color badge.Color) ui.Widget {
	return badge.New(avatar("UI"), "").Color(color).Size(badge.SizeSmall).Alt("Status")
}

func placedDot(placement badge.Placement) ui.Widget {
	return badge.New(avatar("UI"), "").Color(badge.ColorSuccess).Size(badge.SizeSmall).Placement(placement).Alt("Online")
}

func section(title string, content ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(16), content).Gap(10)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Badge"), window.Size(760, 680), window.Theme(base.Light()))
}
