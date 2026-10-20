package main

import (
	"image"
	"image/color"

	"gioui.org/op/paint"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/avatar"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Photo paint.ImageOp
}

type Msg any

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, model Model, _ ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Avatar").Size(24),
				section("Basic",
					layout.Row(
						avatar.New("AM").Image(model.Photo).Alt("Alex Morgan"),
						avatar.New("JR"),
						avatar.New("").Fallback(icon.New(lucide.UserRound).Size(20)),
					).Gap(16).AlignMiddle(),
				),
				section("Sizes",
					layout.Row(
						avatar.New("SM").Image(model.Photo).Alt("Small avatar").Size(avatar.SizeSmall),
						avatar.New("MD").Image(model.Photo).Alt("Medium avatar").Size(avatar.SizeMedium),
						avatar.New("LG").Image(model.Photo).Alt("Large avatar").Size(avatar.SizeLarge),
					).Gap(16).AlignMiddle(),
				),
				section("Colors",
					layout.Row(
						avatar.New("DF").Color(avatar.ColorDefault),
						avatar.New("AC").Color(avatar.ColorAccent),
						avatar.New("SC").Color(avatar.ColorSuccess),
						avatar.New("WR").Color(avatar.ColorWarning),
						avatar.New("DG").Color(avatar.ColorDanger),
					).Gap(16).AlignMiddle(),
				),
				section("Soft",
					layout.Row(
						avatar.New("DF").Color(avatar.ColorDefault).Variant(avatar.VariantSoft),
						avatar.New("AC").Color(avatar.ColorAccent).Variant(avatar.VariantSoft),
						avatar.New("SC").Color(avatar.ColorSuccess).Variant(avatar.VariantSoft),
						avatar.New("WR").Color(avatar.ColorWarning).Variant(avatar.VariantSoft),
						avatar.New("DG").Color(avatar.ColorDanger).Variant(avatar.VariantSoft),
					).Gap(16).AlignMiddle(),
				),
			).Gap(20),
		).Style(style.New().FillWidth().MaxWidth(560).Padding(24)),
	)
}

func section(title string, content ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(16), content).Gap(10)
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
	ui.Run(ui.NewProgram(Model{Photo: paint.NewImageOp(samplePortrait())},
		Update, View), window.Title("FlowUI Avatar"), window.Size(720, 560), window.Theme(base.Light()))
}
