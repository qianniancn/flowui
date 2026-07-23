package ui_test

import "github.com/qianniancn/FlowUI/ui"

func ExampleLinearGradient() {
	gradient := ui.LinearGradient(
		ui.ColorStop(0, ui.RGB(0x2563eb)),
		ui.ColorStop(1, ui.TokenAccent),
	).Angle(90)
	_ = ui.Surface(ui.Text("Status")).Style(ui.Background(gradient))
}
