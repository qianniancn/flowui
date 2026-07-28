package main

import (
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct{}
type Msg any

func Update(*Model, Msg) {}

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Badge").Size(24),
				section("Basic",
					ui.Row(
						ui.Badge(avatar("JD"), "5").Color(ui.BadgeDanger).Size(ui.BadgeSmall),
						ui.Badge(avatar("AB"), "New").Color(ui.BadgeAccent).Size(ui.BadgeSmall),
						ui.Badge(avatar("CD"), "").Color(ui.BadgeSuccess).Size(ui.BadgeSmall).Placement(ui.BadgeBottomRight).Alt("Online"),
						ui.Badge(avatar("EF"), "").Content(ui.Icon(lucide.Bell).Size(10)).Color(ui.BadgeAccent).Size(ui.BadgeSmall).Alt("Notifications"),
					).Gap(28).AlignMiddle(),
				),
				section("Sizes",
					ui.Row(
						ui.Badge(ui.Avatar("SM").Size(ui.AvatarSmall), "5").Color(ui.BadgeDanger).Size(ui.BadgeSmall),
						ui.Badge(ui.Avatar("MD").Size(ui.AvatarMedium), "5").Color(ui.BadgeDanger).Size(ui.BadgeMedium),
						ui.Badge(ui.Avatar("LG").Size(ui.AvatarLarge), "5").Color(ui.BadgeDanger).Size(ui.BadgeLarge),
					).Gap(32).AlignMiddle(),
				),
				section("Colors",
					ui.Row(
						coloredDot(ui.BadgeDefault),
						coloredDot(ui.BadgeAccent),
						coloredDot(ui.BadgeSuccess),
						coloredDot(ui.BadgeWarning),
						coloredDot(ui.BadgeDanger),
					).Gap(28).AlignMiddle(),
				),
				section("Variants",
					ui.Row(
						ui.Badge(avatar("P"), "5").Color(ui.BadgeAccent).Size(ui.BadgeSmall).Variant(ui.BadgePrimary),
						ui.Badge(avatar("S"), "5").Color(ui.BadgeAccent).Size(ui.BadgeSmall).Variant(ui.BadgeSecondary),
						ui.Badge(avatar("SF"), "5").Color(ui.BadgeAccent).Size(ui.BadgeSmall).Variant(ui.BadgeSoft),
					).Gap(28).AlignMiddle(),
				),
				section("Placements",
					ui.Row(
						placedDot(ui.BadgeTopLeft),
						placedDot(ui.BadgeTopRight),
						placedDot(ui.BadgeBottomLeft),
						placedDot(ui.BadgeBottomRight),
					).Gap(28).AlignMiddle(),
				),
			).Gap(20),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(620)).Style(ui.Padding(24)),
	)
}

func avatar(label string) ui.Widget {
	return ui.Avatar(label).Color(ui.AvatarAccent).Variant(ui.AvatarSoft)
}

func coloredDot(color ui.BadgeColor) ui.Widget {
	return ui.Badge(avatar("UI"), "").Color(color).Size(ui.BadgeSmall).Alt("Status")
}

func placedDot(placement ui.BadgePlacement) ui.Widget {
	return ui.Badge(avatar("UI"), "").Color(ui.BadgeSuccess).Size(ui.BadgeSmall).Placement(placement).Alt("Online")
}

func section(title string, content ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(16), content).Gap(10)
}

func main() {
	ui.Run(Model{}, Update, View, ui.Title("FlowUI Badge"), ui.Size(760, 680))
}
