package colorpicker

import (
	"image"
	"image/color"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type ColorSwatchWidget struct {
	theme func(*theme.Theme)
	value color.NRGBA
	size  ColorSwatchSize
	shape ColorSwatchShape
	alt   string
}

func ColorSwatch(value color.NRGBA) ColorSwatchWidget {
	return ColorSwatchWidget{value: value, size: ColorSwatchMedium, shape: ColorSwatchCircle}
}

func (swatch ColorSwatchWidget) Size(size ColorSwatchSize) ColorSwatchWidget {
	swatch.size = size
	return swatch
}

func (swatch ColorSwatchWidget) Shape(shape ColorSwatchShape) ColorSwatchWidget {
	swatch.shape = shape
	return swatch
}

func (swatch ColorSwatchWidget) Alt(alt string) ColorSwatchWidget {
	swatch.alt = alt
	return swatch
}

func (swatch ColorSwatchWidget) Theme(fn func(*theme.Theme)) ColorSwatchWidget {
	swatch.theme = fn
	return swatch
}

func (swatch ColorSwatchWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, swatch.theme); restore != nil {
		defer restore()
	}
	side := min(max(colorSwatchPixels(ctx, gtx, swatch.size), 0), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	size := gtx.Constraints.Constrain(image.Pt(side, side))
	if swatch.alt != "" {
		semantic.LabelOp(swatch.alt).Add(gtx.Ops)
	}
	drawColorSwatch(gtx, size, swatch.value, colorSwatchRadius(ctx, gtx, min(size.X, size.Y), swatch.shape))
	return layout.Dimensions{Size: size}
}
