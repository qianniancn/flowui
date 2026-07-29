package main

import (
	"bytes"
	"image"
	"image/jpeg"

	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/assets/images"
	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	Landscape paint.ImageOp
}

type Msg any

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, model Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Image").Size(24),
				section("Fit",
					ui.Wrap(
						preview("Contain", ui.Image(model.Landscape).Fit(ui.ImageContain).Width(160).Height(104).Radius(16).Alt("Contained landscape")),
						preview("Cover", ui.Image(model.Landscape).Fit(ui.ImageCover).Width(160).Height(104).Radius(16).Alt("Cropped landscape")),
						preview("Fill", ui.Image(model.Landscape).Fit(ui.ImageFill).Width(160).Height(104).Radius(16).Alt("Stretched landscape")),
					).Gap(16).LineGap(16),
				),
				section("Position",
					ui.Wrap(
						preview("Start", ui.Image(model.Landscape).Fit(ui.ImageCover).Position(ui.AlignStart).Width(136).Height(96).Radius(16)),
						preview("Center", ui.Image(model.Landscape).Fit(ui.ImageCover).Position(ui.AlignCenter).Width(136).Height(96).Radius(16)),
						preview("End", ui.Image(model.Landscape).Fit(ui.ImageCover).Position(ui.AlignEnd).Width(136).Height(96).Radius(16)),
					).Gap(16).LineGap(16),
				),
				section("Opacity",
					ui.Image(model.Landscape).Fit(ui.ImageCover).Width(240).Height(96).Radius(16).Opacity(0.55).Alt("Translucent landscape"),
				),
			).Gap(20),
		).Style(ui.FillWidth().MaxWidth(620).Padding(24)),
	)
}

func preview(label string, content ui.Widget) ui.Widget {
	return ui.Column(
		ui.Surface(content).Variant(ui.SurfaceSecondary).Style(ui.Radius(16)),
		ui.Text(label).Size(12),
	).Gap(6)
}

func section(title string, content ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(16), content).Gap(10)
}

func loadImage(data []byte) image.Image {
	value, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return value
}

func main() {
	ui.Run(ui.NewProgram(Model{Landscape: paint.NewImageOp(loadImage(images.BGDesertJPG))},
		Update, View), ui.Title("FlowUI Image"),
		ui.Size(760, 620),
	)
}
