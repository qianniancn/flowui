package main

import (
	"image/color"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct{}
type Msg any

func Update(*Model, Msg) {}

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Linear Gradients").Size(24),
				ui.Text("Reusable brushes with CSS-style angles and multi-stop color transitions."),
				ui.Divider(),
				gradientBand(
					"Left to right",
					ui.LinearGradient(
						ui.ColorStop(0, ui.Color(hex(0x2563eb))),
						ui.ColorStop(1, ui.Color(hex(0x10b981))),
					).Angle(90),
				),
				gradientBand(
					"Three color stops",
					ui.LinearGradient(
						ui.ColorStop(0, ui.Color(hex(0xdc2626))),
						ui.ColorStop(.48, ui.Color(hex(0xf59e0b))),
						ui.ColorStop(1, ui.Color(hex(0x0d9488))),
					).Angle(120),
				),
				gradientBand(
					"Hard transition",
					ui.LinearGradient(
						ui.ColorStop(0, ui.Color(hex(0x18181b))),
						ui.ColorStop(.5, ui.Color(hex(0x18181b))),
						ui.ColorStop(.5, ui.Color(hex(0x52525b))),
						ui.ColorStop(1, ui.Color(hex(0x52525b))),
					).Angle(90),
				),
			).Gap(16),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(720)).Style(ui.Padding(24)),
	)
}

func gradientBand(label string, brush ui.PaintSource) ui.Widget {
	return ui.Surface(
		ui.Box(ui.Text(label).Size(16)).Style(ui.FillWidth()).Style(ui.Padding(20)),
	).Style(ui.Background(brush).
		TextColor(ui.Color(color.NRGBA{R: 255, G: 255, B: 255, A: 255})).
		Radius(8),
	)
}

func hex(value uint32) color.NRGBA {
	return color.NRGBA{
		R: byte(value >> 16),
		G: byte(value >> 8),
		B: byte(value),
		A: 255,
	}
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Linear Gradients"),
		ui.Size(820, 500),
	)
}
