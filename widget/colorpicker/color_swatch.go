package colorpicker

import (
	"github.com/qianniancn/flowui/style"
	"image"
	"image/color"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
)

type ColorSwatchWidget struct {
	override *theme.ColorPickerOverride
	value    color.NRGBA
	size     ColorSwatchSize
	shape    ColorSwatchShape
	alt      string
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

func (swatch ColorSwatchWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return swatch.layoutWithOverride(ctx, gtx, swatch.override)
}

func (swatch ColorSwatchWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ColorPickerOverride) layout.Dimensions {
	visual := resolveStandaloneColorPickerVisual(ctx, override)
	return layoutui.Box(core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		side := min(max(colorSwatchPixels(gtx, visual.Swatch, swatch.size), 0), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
		size := gtx.Constraints.Constrain(image.Pt(side, side))
		if swatch.alt != "" {
			semantic.LabelOp(swatch.alt).Add(gtx.Ops)
		}
		visual.Swatch.Render(gtx, theme.ColorSwatchRenderRequest{
			Size: size, Value: swatch.value,
			Shape: theme.ColorSwatchShape(swatch.shape), SizeKind: uint8(swatch.size),
		})
		return layout.Dimensions{Size: size}
	})).Style(visual.Override).Layout(ctx, gtx)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w ColorSwatchWidget) Style(values ...style.Style) ColorSwatchWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ColorPickerOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
