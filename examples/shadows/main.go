package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	WheelColor color.NRGBA
	Brightness float64
	Alpha      float64
}

type Msg any

type colorChanged color.NRGBA
type brightnessChanged float64
type alphaChanged float64

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case colorChanged:
		model.WheelColor = color.NRGBA(msg)
	case brightnessChanged:
		model.Brightness = float64(msg)
	case alphaChanged:
		model.Alpha = float64(msg)
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	shadow := shadowColor(model)
	preview := ui.Box(ui.Center(shadowPreview(shadow))).Style(
		ui.Width(416).Height(340),
	)
	controls := ui.Box(
		ui.Column(
			ui.ColorWheel("shadow-color", model.WheelColor).
				Size(190).
				Label("阴影颜色").
				OnChange(func(value color.NRGBA) {
					send(colorChanged(value))
				}),
			ui.Box(
				ui.Column(
					ui.Slider("brightness", model.Brightness).
						Range(0, 1).
						Step(.01).
						Label("明暗").
						ShowValue().
						FormatValue(formatPercent).
						OnChange(func(value float64) {
							send(brightnessChanged(value))
						}),
					ui.Slider("alpha", model.Alpha).
						Range(0, 1).
						Step(.01).
						Label("透明度").
						ShowValue().
						FormatValue(formatPercent).
						OnChange(func(value float64) {
							send(alphaChanged(value))
						}),
				).Gap(12),
			).Style(ui.Width(280).PaddingX(20)),
		).Gap(24).AlignMiddle(),
	).Style(ui.Width(280))

	return ui.Center(
		ui.Box(
			ui.Scroll("shadows",
				ui.Column(
					ui.Text("FlowUI Shadows").Size(24),
					ui.Wrap(preview, controls).
						Gap(40).
						LineGap(24).
						AlignMiddle(),
				).Gap(16).AlignMiddle(),
			).Vertical(),
		).Style(ui.MaxWidth(860).Padding(32)),
	)
}

func shadowPreview(shadow color.NRGBA) ui.Widget {
	ambient := shadow
	ambient.A = uint8(math.Round(float64(shadow.A) * .65))
	return ui.Card(
		ui.Text("Card shadow").Size(20),
		ui.Description(fmt.Sprintf("#%02X%02X%02X%02X", shadow.R, shadow.G, shadow.B, shadow.A)),
	).
		Padding(28).
		Gap(6).
		Style(
			ui.Width(320).
				MinHeight(144).
				Radius(18).
				BoxShadow(0, 1, 8, 1, ui.Color(shadow)).
				BoxShadow(0, 7, 24, -2, ui.Color(ambient)),
		)
}

func shadowColor(model Model) color.NRGBA {
	brightness := min(max(model.Brightness, 0), 1)
	channel := func(value uint8) uint8 {
		return uint8(math.Round(float64(value) * brightness))
	}
	return color.NRGBA{
		R: channel(model.WheelColor.R),
		G: channel(model.WheelColor.G),
		B: channel(model.WheelColor.B),
		A: uint8(math.Round(min(max(model.Alpha, 0), 1) * 255)),
	}
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.0f%%", value*100)
}

func main() {
	ui.Run(Model{
		WheelColor: color.NRGBA{R: 71, G: 194, B: 255, A: 255},
		Brightness: .62,
		Alpha:      .7,
	}, Update, View,
		ui.Title("FlowUI Shadows"),
		ui.Size(920, 560),
	)
}
