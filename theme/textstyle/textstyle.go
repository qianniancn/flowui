// Package textstyle contains the theme-owned visual recipe for plain text.
// Text behavior remains in widget/text; this package only turns theme tokens
// and the inherited foreground into a provider result.
package textstyle

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildText builds the complete provider-owned text visual plan.
func BuildText(active *theme.Theme, request theme.TextRequest) theme.TextVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.TextVisual{
		Defaults: style.Style{}.
			FontSize(active.Typography.BodySize).
			Typeface(active.Typography.Typeface).
			TextColor(style.Solid(request.Foreground)),
		Selection: active.Semantic().Selection,
		Underline: underlineRenderer,
	}
}

func underlineRenderer(gtx layout.Context, request theme.TextUnderlineRequest) {
	if request.Size.X <= 0 || request.Size.Y <= 0 || request.Color.A == 0 {
		return
	}
	thickness := max(gtx.Dp(1), 1)
	top := request.Baseline + thickness
	if top+thickness > request.Size.Y {
		top = request.Size.Y - thickness
	}
	if top < 0 {
		return
	}
	paint.FillShape(gtx.Ops, request.Color, clip.Rect{
		Min: image.Point{Y: top},
		Max: image.Point{X: request.Size.X, Y: top + thickness},
	}.Op())
}
