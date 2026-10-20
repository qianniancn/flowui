package main

import (
	"image/color"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/surface"
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
				text.New("FlowUI Linear Gradients").Size(24),
				text.New("Reusable brushes with CSS-style angles and multi-stop color transitions."),
				layout.Divider(),
				gradientBand(
					"Left to right",
					style.LinearGradient(
						style.ColorStop(0, style.Color(hex(0x2563eb))),
						style.ColorStop(1, style.Color(hex(0x10b981))),
					).Angle(90),
				),
				gradientBand(
					"Three color stops",
					style.LinearGradient(
						style.ColorStop(0, style.Color(hex(0xdc2626))),
						style.ColorStop(.48, style.Color(hex(0xf59e0b))),
						style.ColorStop(1, style.Color(hex(0x0d9488))),
					).Angle(120),
				),
				gradientBand(
					"Hard transition",
					style.LinearGradient(
						style.ColorStop(0, style.Color(hex(0x18181b))),
						style.ColorStop(.5, style.Color(hex(0x18181b))),
						style.ColorStop(.5, style.Color(hex(0x52525b))),
						style.ColorStop(1, style.Color(hex(0x52525b))),
					).Angle(90),
				),
			).Gap(16),
		).Style(style.New().FillWidth().MaxWidth(720).Padding(24)),
	)
}

func gradientBand(label string, brush style.PaintSource) ui.Widget {
	return (surface.New(
		layout.Box(text.New(label).Size(16)).Style(style.New().FillWidth().Padding(20)),
	)).Style(style.New().Background(brush).
		TextColor(style.Color(color.NRGBA{R: 255, G: 255, B: 255, A: 255})).
		Radius(8))
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
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Linear Gradients"),
		window.Size(820, 500),
		window.Theme(base.Light()),
	)
}
