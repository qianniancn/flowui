package main

import (
	"image/color"

	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	favorite bool
}

type ToggleFavorite struct{}

func Update(model *Model, _ ToggleFavorite) {
	model.favorite = !model.favorite
}

func View(_ *ui.Context, model Model, send ui.Send[ToggleFavorite]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Lucide Icons").Size(24),
				ui.Divider(),
				section("Icons",
					ui.Wrap(
						labeled("Search", ui.Icon(lucide.Search)),
						labeled("Heart", ui.Icon(lucide.Heart)),
						labeled("Settings", ui.Icon(lucide.Settings)),
						labeled("Album", ui.Icon(lucide.Album)),
						labeled("Alert", ui.Icon(lucide.BadgeAlert)),
						labeled("Cone", ui.Icon(lucide.Cone)),
						labeled("Navigation", ui.Icon(lucide.Navigation)),
						labeled("Chart", ui.Icon(lucide.ChartScatter)),
					).Gap(32).LineGap(20).AlignMiddle(),
				),
				section("Sizes",
					ui.Row(
						labeled("16", ui.Icon(lucide.Search).Size(16)),
						labeled("20", ui.Icon(lucide.Search).Size(20)),
						labeled("24", ui.Icon(lucide.Search).Size(24)),
					).Gap(32).AlignMiddle(),
				),
				section("Colors",
					ui.Row(
						ui.Icon(lucide.Heart).Color(color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0xff}),
						ui.Icon(lucide.Settings).Color(color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0xff}),
					).Gap(24).AlignMiddle(),
				),
				section("Components",
					ui.Row(
						ui.Button("settings", ui.Row(
							ui.Icon(lucide.Settings).Size(18),
							ui.Text("Settings"),
						).Gap(8).AlignMiddle()),
						ui.ToggleButton("favorite", model.favorite, ui.Icon(lucide.Heart).Size(18)).
							IconOnly().
							Label("Favorite").
							OnChange(func(bool) { send(ToggleFavorite{}) }),
						ui.CloseButton("close").Icon(ui.Icon(lucide.X).Size(16)),
					).Gap(16).AlignMiddle(),
				),
			).Gap(20),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(640)).Style(ui.Padding(24)),
	)
}

func section(title string, content ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), content).Gap(12)
}

func labeled(label string, value ui.Widget) ui.Widget {
	return ui.Column(ui.Center(value), ui.Text(label).Size(12)).Gap(8).AlignMiddle()
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Lucide Icons"),
		ui.Size(720, 680),
	)
}
