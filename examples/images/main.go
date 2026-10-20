package main

import (
	"bytes"
	"image"
	"image/jpeg"

	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/assets/images"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	imagepkg "github.com/qianniancn/flowui/widget/image"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Landscape paint.ImageOp
}

type Msg any

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, model Model, _ ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Image").Size(24),
				section("Fit",
					layout.Wrap(
						preview("Contain", imagepkg.New(model.Landscape).Fit(imagepkg.FitContain).Width(160).Height(104).Radius(16).Alt("Contained landscape")),
						preview("Cover", imagepkg.New(model.Landscape).Fit(imagepkg.FitCover).Width(160).Height(104).Radius(16).Alt("Cropped landscape")),
						preview("Fill", imagepkg.New(model.Landscape).Fit(imagepkg.FitFill).Width(160).Height(104).Radius(16).Alt("Stretched landscape")),
					).Gap(16).LineGap(16),
				),
				section("Position",
					layout.Wrap(
						preview("Start", imagepkg.New(model.Landscape).Fit(imagepkg.FitCover).Position(layout.AlignStart).Width(136).Height(96).Radius(16)),
						preview("Center", imagepkg.New(model.Landscape).Fit(imagepkg.FitCover).Position(layout.AlignCenter).Width(136).Height(96).Radius(16)),
						preview("End", imagepkg.New(model.Landscape).Fit(imagepkg.FitCover).Position(layout.AlignEnd).Width(136).Height(96).Radius(16)),
					).Gap(16).LineGap(16),
				),
				section("Opacity",
					imagepkg.New(model.Landscape).Fit(imagepkg.FitCover).Width(240).Height(96).Radius(16).Opacity(0.55).Alt("Translucent landscape"),
				),
			).Gap(20),
		).Style(style.New().FillWidth().MaxWidth(620).Padding(24)),
	)
}

func preview(label string, content ui.Widget) ui.Widget {
	return layout.Column(
		(surface.New(content).Variant(surface.Secondary)).Style(style.New().Radius(16)),
		text.New(label).Size(12),
	).Gap(6)
}

func section(title string, content ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(16), content).Gap(10)
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
		Update, View), window.Title("FlowUI Image"),
		window.Size(760, 620),
		window.Theme(base.Light()),
	)
}
