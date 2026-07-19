package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
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

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case VolumeChanged:
		model.Volume = float64(msg)
	case PriceChanged:
		model.PriceStart = msg.Start
		model.PriceEnd = msg.End
	case IntensityChanged:
		model.Intensity = float64(msg)
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Scroll("sliders",
				ui.Column(
					ui.Text("FlowUI Sliders").Size(24),
					ui.Text("Controlled single-value, range, vertical, and disabled sliders."),
					ui.Divider(),
					ui.Card(
						cardHeader("Volume", "Drag the thumb or use the arrow, Home, End, PageUp, and PageDown keys."),
						ui.Slider("volume", model.Volume).
							Label("Volume").
							ShowValue().
							OnChange(func(value float64) {
								send(VolumeChanged(value))
							}),
					),
					ui.Card(
						cardHeader("Price range", "Each half of the track targets its nearest range thumb."),
						ui.RangeSlider("price", model.PriceStart, model.PriceEnd).
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
					).Variant(ui.CardSecondary),
					ui.Row(
						ui.Card(
							cardHeader("Vertical", "The value increases from bottom to top."),
							ui.Box(
								ui.Slider("intensity", model.Intensity).
									Label("Intensity").
									ShowValue().
									Vertical().
									OnChange(func(value float64) {
										send(IntensityChanged(value))
									}),
							).Width(96).Height(260),
						),
						ui.Expanded(
							ui.Card(
								cardHeader("Disabled", "Disabled sliders preserve their label while reducing control prominence."),
								ui.Slider("disabled", 64).
									Label("Storage").
									ValueText("64 GB").
									Disabled(true),
							).Variant(ui.CardTertiary),
						),
					).Gap(16).AlignStart(),
				).Gap(16),
			).Vertical(),
		).FillWidth().MaxWidth(760).Padding(24),
	)
}

func cardHeader(title, description string) ui.Widget {
	return ui.Column(ui.Text(title).Size(16), ui.Text(description).Size(14)).Gap(4)
}

func main() {
	ui.Run(Model{
		Volume:     30,
		PriceStart: 100,
		PriceEnd:   500,
		Intensity:  45,
	}, Update, View,
		ui.Title("FlowUI Sliders"),
		ui.Size(900, 620),
	)
}
