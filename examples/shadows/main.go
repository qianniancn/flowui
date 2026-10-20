package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/colorpicker"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/slider"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case colorChanged:
		model.WheelColor = color.NRGBA(msg)
	case brightnessChanged:
		model.Brightness = float64(msg)
	case alphaChanged:
		model.Alpha = float64(msg)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	shadow := shadowColor(model)
	preview := layout.Box(layout.Center(shadowPreview(shadow))).Style(
		style.New().Width(416).Height(340),
	)
	controls := layout.Box(
		layout.Column(
			colorpicker.ColorWheel("shadow-color", model.WheelColor).
				Size(190).
				Label("阴影颜色").
				OnChange(func(value color.NRGBA) {
					send(colorChanged(value))
				}),
			layout.Box(
				layout.Column(
					slider.New("brightness", model.Brightness).
						Range(0, 1).
						Step(.01).
						Label("明暗").
						ShowValue().
						FormatValue(formatPercent).
						OnChange(func(value float64) {
							send(brightnessChanged(value))
						}),
					slider.New("alpha", model.Alpha).
						Range(0, 1).
						Step(.01).
						Label("透明度").
						ShowValue().
						FormatValue(formatPercent).
						OnChange(func(value float64) {
							send(alphaChanged(value))
						}),
				).Gap(12),
			).Style(style.New().Width(280).PaddingX(20)),
		).Gap(24).AlignMiddle(),
	).Style(style.New().Width(280))

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Shadows").Size(24),
				layout.Wrap(preview, controls).
					Gap(40).
					LineGap(24).
					AlignMiddle(),
			).Gap(16).AlignMiddle(),
			).Vertical(),
		).Style(style.New().MaxWidth(860).Padding(32)),
	)
}

func shadowPreview(shadow color.NRGBA) ui.Widget {
	ambient := shadow
	ambient.A = uint8(math.Round(float64(shadow.A) * .65))
	return (card.New(
		text.New("Card shadow").Size(20),
		description.New(fmt.Sprintf("#%02X%02X%02X%02X", shadow.R, shadow.G, shadow.B, shadow.A)),
	)).StyleWithMetrics(style.New().Width(320).
		MinHeight(144).
		Radius(18).
		BoxShadow(0, 1, 8, 1, style.Color(shadow)).
		BoxShadow(0, 7, 24, -2, style.Color(ambient)), 28, 6)
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
	ui.Run(ui.NewProgram(Model{
		WheelColor: color.NRGBA{R: 71, G: 194, B: 255, A: 255},
		Brightness: .62,
		Alpha:      .7,
	},
		Update, View), window.Title("FlowUI Shadows"),
		window.Size(920, 560),
		window.Theme(base.Light()),
	)
}
