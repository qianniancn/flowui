package host

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func layoutTrackedInset(ctx *frame.Context, gtx layout.Context, inset layout.Inset, child layout.Widget) layout.Dimensions {
	var placement frame.OverlayPlacement
	dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var dims layout.Dimensions
		dims, placement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return child(gtx)
		})
		return dims
	})
	placement.PlaceOffset(insetOffset(gtx, inset))
	return dims
}

func insetOffset(gtx layout.Context, inset layout.Inset) image.Point {
	top := gtx.Dp(inset.Top)
	right := gtx.Dp(inset.Right)
	bottom := gtx.Dp(inset.Bottom)
	left := gtx.Dp(inset.Left)
	if gtx.Constraints.Max.X-left-right < 0 {
		left = 0
	}
	if gtx.Constraints.Max.Y-top-bottom < 0 {
		top = 0
	}
	return image.Pt(left, top)
}

func layoutTrackedDirection(ctx *frame.Context, gtx layout.Context, direction layout.Direction, child layout.Widget) layout.Dimensions {
	var childDims layout.Dimensions
	var placement frame.OverlayPlacement
	dims := direction.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		childDims, placement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return child(gtx)
		})
		return childDims
	})
	placement.PlaceOffset(direction.Position(childDims.Size, dims.Size))
	return dims
}
