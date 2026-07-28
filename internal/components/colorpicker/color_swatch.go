package colorpicker

import (
	"image"
	"image/color"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

type ColorSwatchWidget struct {
	value       color.NRGBA
	size        ColorSwatchSize
	shape       ColorSwatchShape
	alt         string
	customStyle flowstyle.Style
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

func (swatch ColorSwatchWidget) Style(value flowstyle.Style) ColorSwatchWidget {
	swatch.customStyle = value
	return swatch
}

func (swatch ColorSwatchWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return layoutui.Box(frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		side := min(max(colorSwatchPixels(ctx, gtx, swatch.size), 0), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
		size := gtx.Constraints.Constrain(image.Pt(side, side))
		if swatch.alt != "" {
			semantic.LabelOp(swatch.alt).Add(gtx.Ops)
		}
		drawColorSwatch(gtx, size, swatch.value, colorSwatchRadius(ctx, gtx, min(size.X, size.Y), swatch.shape))
		return layout.Dimensions{Size: size}
	})).Style(swatch.customStyle).Layout(ctx, gtx)
}
