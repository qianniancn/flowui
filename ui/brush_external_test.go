package ui_test

import (
	"image/color"

	"github.com/qianniancn/FlowUI/ui"
)

func ExampleLinearGradient() {
	gradient := ui.LinearGradient(
		ui.ColorStop(0, color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff}),
		ui.ColorStop(1, color.NRGBA{R: 0x10, G: 0xb9, B: 0x81, A: 0xff}),
	).Angle(90)
	_ = ui.Surface(ui.Text("Status")).Background(gradient)
}
