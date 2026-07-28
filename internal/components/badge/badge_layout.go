package badge

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	textui "github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
)

type recordedBadgeChild struct {
	call      op.CallOp
	dims      layout.Dimensions
	placement frame.OverlayPlacement
}

func (b Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if b.anchor == nil {
		return layout.Dimensions{}
	}
	anchor := recordBadgeChild(ctx, gtx, b.anchor)
	badgeGtx := gtx
	badgeGtx.Constraints.Min = image.Point{}
	badge := b.recordBadge(ctx, badgeGtx)

	anchor.placement.PlaceOffset(image.Point{})
	anchor.call.Add(gtx.Ops)
	position := badgePosition(anchor.dims.Size, badge.dims.Size, b.placement, frame.ActiveTheme(ctx).Components.Badge.PlacementOffsetRatio)
	badge.placement.PlaceOffset(position)
	offset := op.Offset(position).Push(gtx.Ops)
	badge.call.Add(gtx.Ops)
	offset.Pop()
	return anchor.dims
}

func (b Widget) recordBadge(ctx *frame.Context, gtx layout.Context) recordedBadgeChild {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return b.layoutBadge(ctx, gtx)
	})
	return recordedBadgeChild{call: macro.Stop(), dims: dims, placement: placement}
}

func recordBadgeChild(ctx *frame.Context, gtx layout.Context, child frame.Widget) recordedBadgeChild {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return child.Layout(ctx, gtx)
	})
	return recordedBadgeChild{call: macro.Stop(), dims: dims, placement: placement}
}

func (b Widget) layoutBadge(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	resolved := b.resolveStyle(ctx, gtx)
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		if label := b.semanticLabel(); label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if b.content != nil {
				return layoutui.LayoutResolved(ctx, gtx, resolved.label, b.content)
			}
			if b.label != "" {
				return layoutui.LayoutResolved(ctx, gtx, resolved.label, textui.New(b.label))
			}
			return layout.Dimensions{}
		})
	})
	return layoutui.LayoutResolved(ctx, gtx, resolved.root, content)
}

func badgePosition(anchor, badge image.Point, placement Placement, offsetRatio float32) image.Point {
	offsetX := int(float32(badge.X)*max(offsetRatio, 0) + 0.5)
	offsetY := int(float32(badge.Y)*max(offsetRatio, 0) + 0.5)
	switch placement {
	case PlacementTopLeft:
		return image.Pt(-offsetX, -offsetY)
	case PlacementBottomRight:
		return image.Pt(anchor.X-badge.X+offsetX, anchor.Y-badge.Y+offsetY)
	case PlacementBottomLeft:
		return image.Pt(-offsetX, anchor.Y-badge.Y+offsetY)
	default:
		return image.Pt(anchor.X-badge.X+offsetX, -offsetY)
	}
}

func (b Widget) semanticLabel() string {
	if b.alt != "" {
		return b.alt
	}
	return b.label
}
