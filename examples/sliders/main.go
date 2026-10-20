package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/slider"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Volume     float64
	PriceStart float64
	PriceEnd   float64
	Intensity  float64
}

type Msg any

type VolumeChanged float64
type PriceChanged struct {
	Start float64
	End   float64
}
type IntensityChanged float64

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case VolumeChanged:
		model.Volume = float64(msg)
	case PriceChanged:
		model.PriceStart = msg.Start
		model.PriceEnd = msg.End
	case IntensityChanged:
		model.Intensity = float64(msg)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Sliders").Size(24),
				text.New("Controlled single-value, range, vertical, and disabled sliders."),
				layout.Divider(),
				card.New(
					cardHeader("Volume", "Drag the thumb or use the arrow, Home, End, PageUp, and PageDown keys."),
					slider.New("volume", model.Volume).
						Label("Volume").
						ShowValue().
						OnChange(func(value float64) {
							send(VolumeChanged(value))
						}),
				),
				card.New(
					cardHeader("Price range", "Each half of the track targets its nearest range thumb."),
					slider.RangeSlider("price", model.PriceStart, model.PriceEnd).
						Label("Price range").
						Range(0, 1000).
						Step(50).
						ShowValue().
						FormatValue(func(value float64) string {
							return fmt.Sprintf("$%.0f", value)
						}).
						OnRangeChange(func(start, end float64) {
							send(PriceChanged{Start: start, End: end})
						}),
				).Variant(card.Secondary),
				layout.Row(
					card.New(
						cardHeader("Vertical", "The value increases from bottom to top."),
						layout.Box(
							slider.New("intensity", model.Intensity).
								Label("Intensity").
								ShowValue().
								Vertical().
								OnChange(func(value float64) {
									send(IntensityChanged(value))
								}),
						).Style(style.New().Width(96).Height(260)),
					),
					layout.Expanded(
						card.New(
							cardHeader("Disabled", "Disabled sliders preserve their label while reducing control prominence."),
							slider.New("disabled", 64).
								Label("Storage").
								ValueText("64 GB").
								Disabled(true),
						).Variant(card.Tertiary),
					),
				).Gap(16).AlignStart(),
			).Gap(16),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func cardHeader(title, description string) ui.Widget {
	return layout.Column(text.New(title).Size(16), text.New(description).Size(14)).Gap(4)
}

func main() {
	ui.Run(ui.NewProgram(Model{
		Volume:     30,
		PriceStart: 100,
		PriceEnd:   500,
		Intensity:  45,
	},
		Update, View), window.Title("FlowUI Sliders"),
		window.Size(900, 620),
		window.Theme(base.Light()),
	)
}
