package main

import (
	"fmt"
	"image/color"

	"github.com/qianniancn/flowui/ui"
)

type Field string

const (
	fieldBasic    Field = "basic"
	fieldBrand    Field = "brand"
	fieldAlpha    Field = "alpha"
	fieldArea     Field = "area"
	fieldField    Field = "field"
	fieldSlider   Field = "slider"
	fieldSwatches Field = "swatches"
)

type Model struct {
	Basic      color.NRGBA
	Brand      color.NRGBA
	Alpha      color.NRGBA
	Area       color.NRGBA
	FieldColor color.NRGBA
	Slider     color.NRGBA
	Swatches   color.NRGBA
	Last       string
}

type ColorChanged struct {
	Field Field
	Value color.NRGBA
}

func Update(model *Model, message ColorChanged) ui.Cmd[ColorChanged] {
	switch message.Field {
	case fieldBasic:
		model.Basic = message.Value
	case fieldBrand:
		model.Brand = message.Value
	case fieldAlpha:
		model.Alpha = message.Value
	case fieldArea:
		model.Area = message.Value
	case fieldField:
		model.FieldColor = message.Value
	case fieldSlider:
		model.Slider = message.Value
	case fieldSwatches:
		model.Swatches = message.Value
	}
	model.Last = fmt.Sprintf("%s: %s", message.Field, formatColor(message.Value))
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[ColorChanged]) ui.Widget {
	status := "Select a color"
	if model.Last != "" {
		status = model.Last
	}
	return ui.Center(
		ui.Box(
			ui.Scroll("color-pickers",
				ui.Column(
					ui.Text("FlowUI color components").Size(24),
					ui.Text(status).Size(14),
					ui.Divider(),
					section("ColorArea",
						ui.Row(
							ui.ColorArea("standalone-area", model.Area).
								ShowDots(true).
								OnChange(func(value color.NRGBA) { send(ColorChanged{Field: fieldArea, Value: value}) }),
							ui.Column(
								ui.ColorSwatch(model.Area).Size(ui.ColorSwatchLarge),
								ui.Text(formatColor(model.Area)).Size(14),
							).Gap(8),
						).Gap(16),
					),
					section("ColorSlider",
						ui.Box(ui.Column(
							ui.ColorSlider("standalone-hue", model.Slider, ui.ColorChannelHue).
								OnChange(func(value color.NRGBA) { send(ColorChanged{Field: fieldSlider, Value: value}) }),
							ui.ColorSlider("standalone-alpha", model.Slider, ui.ColorChannelAlpha).
								OnChange(func(value color.NRGBA) { send(ColorChanged{Field: fieldSlider, Value: value}) }),
						).Gap(10)).Style(ui.MaxWidth(300)),
					),
					section("ColorField",
						ui.Box(ui.ColorField("standalone-field", model.FieldColor).
							Label("Brand color").
							Description("Enter a hexadecimal color").
							Swatch(true).
							FullWidth().
							OnChange(func(value color.NRGBA) { send(ColorChanged{Field: fieldField, Value: value}) })).Style(ui.MaxWidth(280)),
					),
					section("ColorSwatch",
						ui.Row(
							ui.ColorSwatch(model.Swatches).Size(ui.ColorSwatchExtraSmall),
							ui.ColorSwatch(model.Swatches).Size(ui.ColorSwatchSmall),
							ui.ColorSwatch(model.Swatches).Size(ui.ColorSwatchMedium),
							ui.ColorSwatch(model.Swatches).Size(ui.ColorSwatchLarge),
							ui.ColorSwatch(model.Swatches).Size(ui.ColorSwatchExtraLarge).Shape(ui.ColorSwatchSquare),
						).Gap(10),
					),
					section("ColorSwatchPicker",
						ui.Column(
							ui.Text("Circle").Size(14),
							ui.ColorSwatchPicker("standalone-swatches", model.Swatches, heroUIPresets).
								OnChange(func(value color.NRGBA) { send(ColorChanged{Field: fieldSwatches, Value: value}) }),
							ui.Text("Square").Size(14),
							ui.ColorSwatchPicker("standalone-square-swatches", model.Swatches, heroUIPresets).
								Shape(ui.ColorSwatchSquare).
								OnChange(func(value color.NRGBA) { send(ColorChanged{Field: fieldSwatches, Value: value}) }),
						).Gap(8),
					),
					ui.Divider(),
					ui.Text("ColorPicker composition").Size(20),
					section("Basic",
						boundPicker("basic", fieldBasic, model.Basic, "Pick a color", send).
							ShowHistory(false),
					),
					section("Presets and field",
						boundPicker("brand", fieldBrand, model.Brand, "Brand color", send).
							Presets(brandPresets).
							ShowField(),
					),
					section("Alpha",
						boundPicker("alpha", fieldAlpha, model.Alpha, "Overlay color", send).
							Alpha(true).
							ShowField(),
					),
					section("Disabled",
						ui.ColorPicker("disabled", color.NRGBA{R: 0x71, G: 0x71, B: 0x7a, A: 255}).
							Label("Unavailable color").
							Disabled(true),
					),
				).Gap(18),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(720)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func boundPicker(key string, field Field, value color.NRGBA, label string, send ui.Send[ColorChanged]) ui.ColorPickerWidget {
	return ui.ColorPicker(key, value).
		Label(label).
		OnChange(func(value color.NRGBA) {
			send(ColorChanged{Field: field, Value: value})
		})
}

var brandPresets = []color.NRGBA{
	{R: 0xef, G: 0x44, B: 0x44, A: 255},
	{R: 0xf9, G: 0x73, B: 0x16, A: 255},
	{R: 0xea, G: 0xb3, B: 0x08, A: 255},
	{R: 0x22, G: 0xc5, B: 0x5e, A: 255},
	{R: 0x06, G: 0xb6, B: 0xd4, A: 255},
	{R: 0x3b, G: 0x82, B: 0xf6, A: 255},
	{R: 0x8b, G: 0x5c, B: 0xf6, A: 255},
	{R: 0xec, G: 0x48, B: 0x99, A: 255},
	{R: 0xf4, G: 0x3f, B: 0x5e, A: 255},
}

var heroUIPresets = []color.NRGBA{
	{R: 0xf4, G: 0x3f, B: 0x5e, A: 255},
	{R: 0xd9, G: 0x46, B: 0xef, A: 255},
	{R: 0x8b, G: 0x5c, B: 0xf6, A: 255},
	{R: 0x3b, G: 0x82, B: 0xf6, A: 255},
	{R: 0x06, G: 0xb6, B: 0xd4, A: 255},
	{R: 0x10, G: 0xb9, B: 0x81, A: 255},
	{R: 0x84, G: 0xcc, B: 0x16, A: 255},
}

func formatColor(value color.NRGBA) string {
	if value.A != 255 {
		return fmt.Sprintf("#%02X%02X%02X%02X", value.R, value.G, value.B, value.A)
	}
	return fmt.Sprintf("#%02X%02X%02X", value.R, value.G, value.B)
}

func main() {
	ui.Run(ui.NewProgram(Model{
		Basic:      color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255},
		Brand:      color.NRGBA{R: 0xf4, G: 0x3f, B: 0x5e, A: 255},
		Alpha:      color.NRGBA{R: 0x32, G: 0x55, B: 0x78, A: 0xcc},
		Area:       color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255},
		FieldColor: color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 255},
		Slider:     color.NRGBA{R: 0xf4, G: 0x3f, B: 0x5e, A: 0xcc},
		Swatches:   color.NRGBA{R: 0x8b, G: 0x5c, B: 0xf6, A: 255},
	},
		Update, View), ui.Title("FlowUI ColorPicker"),
		ui.Size(900, 700),
	)
}
