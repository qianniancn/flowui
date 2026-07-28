package main

import (
	"image"
	"image/color"

	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	Photo paint.ImageOp
}

type Msg any

func Update(*Model, Msg) {}

func View(_ *ui.Context, model Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Avatar").Size(24),
				section("Basic",
					ui.Row(
						ui.Avatar("AM").Image(model.Photo).Alt("Alex Morgan"),
						ui.Avatar("JR"),
						ui.Avatar("").Fallback(ui.Icon(lucide.UserRound).Size(20)),
					).Gap(16).AlignMiddle(),
				),
				section("Sizes",
					ui.Row(
						ui.Avatar("SM").Image(model.Photo).Alt("Small avatar").Size(ui.AvatarSmall),
						ui.Avatar("MD").Image(model.Photo).Alt("Medium avatar").Size(ui.AvatarMedium),
						ui.Avatar("LG").Image(model.Photo).Alt("Large avatar").Size(ui.AvatarLarge),
					).Gap(16).AlignMiddle(),
				),
				section("Colors",
					ui.Row(
						ui.Avatar("DF").Color(ui.AvatarDefault),
						ui.Avatar("AC").Color(ui.AvatarAccent),
						ui.Avatar("SC").Color(ui.AvatarSuccess),
						ui.Avatar("WR").Color(ui.AvatarWarning),
						ui.Avatar("DG").Color(ui.AvatarDanger),
					).Gap(16).AlignMiddle(),
				),
				section("Soft",
					ui.Row(
						ui.Avatar("DF").Color(ui.AvatarDefault).Variant(ui.AvatarSoft),
						ui.Avatar("AC").Color(ui.AvatarAccent).Variant(ui.AvatarSoft),
						ui.Avatar("SC").Color(ui.AvatarSuccess).Variant(ui.AvatarSoft),
						ui.Avatar("WR").Color(ui.AvatarWarning).Variant(ui.AvatarSoft),
						ui.Avatar("DG").Color(ui.AvatarDanger).Variant(ui.AvatarSoft),
					).Gap(16).AlignMiddle(),
				),
			).Gap(20),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(560)).Style(ui.Padding(24)),
	)
}

func section(title string, content ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(16), content).Gap(10)
}

func samplePortrait() image.Image {
	const width, height = 120, 80
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(54 + x*80/width),
				G: uint8(118 + y*70/height),
				B: uint8(180 + x*45/width),
				A: 0xff,
			})
		}
	}
	drawEllipse(img, 60, 38, 18, 22, color.NRGBA{R: 0xf2, G: 0xc6, B: 0xa0, A: 0xff})
	drawEllipse(img, 60, 27, 19, 12, color.NRGBA{R: 0x28, G: 0x24, B: 0x2b, A: 0xff})
	drawEllipse(img, 60, 76, 34, 22, color.NRGBA{R: 0x25, G: 0x47, B: 0x76, A: 0xff})
	return img
}

func drawEllipse(img *image.NRGBA, centerX, centerY, radiusX, radiusY int, value color.NRGBA) {
	for y := centerY - radiusY; y <= centerY+radiusY; y++ {
		for x := centerX - radiusX; x <= centerX+radiusX; x++ {
			dx := float64(x-centerX) / float64(radiusX)
			dy := float64(y-centerY) / float64(radiusY)
			if dx*dx+dy*dy <= 1 && image.Pt(x, y).In(img.Bounds()) {
				img.SetNRGBA(x, y, value)
			}
		}
	}
}

func main() {
	ui.Run(Model{Photo: paint.NewImageOp(samplePortrait())}, Update, View, ui.Title("FlowUI Avatar"), ui.Size(720, 560))
}
