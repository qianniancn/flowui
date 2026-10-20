package radiostyle

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/theme"
)

// radioIndicatorRenderer owns the selected-dot geometry. The widget supplies
// only the resolved foreground, selection progress, and press state.
func radioIndicatorRenderer(gtx layout.Context, request theme.RadioIndicatorRenderRequest) layout.Dimensions {
	if request.Selection <= 0 || request.Size.X <= 0 || request.Size.Y <= 0 {
		return layout.Dimensions{Size: request.Size}
	}

	targetScale := request.DotScale
	if request.Pressed {
		targetScale = request.DotPressedScale
	}
	diameter := min(request.Size.X, request.Size.Y)
	dotSize := max(int(float32(diameter)*targetScale*request.Selection+.5), 1)
	center := image.Pt(request.Size.X/2, request.Size.Y/2)
	dot := image.Rectangle{Min: center.Sub(image.Pt(dotSize/2, dotSize/2))}
	dot.Max = dot.Min.Add(image.Pt(dotSize, dotSize))
	if request.Foreground.A != 0 {
		paint.FillShape(gtx.Ops, request.Foreground, clip.Ellipse(dot).Op(gtx.Ops))
	}
	return layout.Dimensions{Size: request.Size}
}

var _ theme.RadioIndicatorRenderer = radioIndicatorRenderer
