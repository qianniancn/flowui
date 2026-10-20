package badge

import (
	"image"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
)

type recordedBadgeChild struct {
	call      op.CallOp
	dims      layout.Dimensions
	placement core.OverlayPlacement
}

func (b Widget) layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if b.anchor == nil {
		return layout.Dimensions{}
	}
	return b.layoutWithVisual(ctx, gtx, b.visualWithCache(ctx, nil))
}

func (b Widget) layoutWithVisual(ctx *core.Context, gtx layout.Context, visual theme.BadgeVisual) layout.Dimensions {
	anchor := recordBadgeChild(ctx, gtx, b.anchor)
	badgeGtx := gtx
	badgeGtx.Constraints.Min = image.Point{}
	badge := b.recordBadge(ctx, badgeGtx, visual)

	anchor.placement.PlaceOffset(image.Point{})
	anchor.call.Add(gtx.Ops)
	position := badgePosition(anchor.dims.Size, badge.dims.Size, b.placement, visual.Metrics.PlacementOffsetRatio)
	badge.placement.PlaceOffset(position)
	offset := op.Offset(position).Push(gtx.Ops)
	badge.call.Add(gtx.Ops)
	offset.Pop()
	return anchor.dims
}

func (b Widget) recordBadge(ctx *core.Context, gtx layout.Context, visual theme.BadgeVisual) recordedBadgeChild {
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return b.layoutBadgeWithVisual(ctx, gtx, visual)
	})
	return recordedBadgeChild{call: macro.Stop(), dims: dims, placement: placement}
}

func recordBadgeChild(ctx *core.Context, gtx layout.Context, child core.Widget) recordedBadgeChild {
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return child.Layout(ctx, gtx)
	})
	return recordedBadgeChild{call: macro.Stop(), dims: dims, placement: placement}
}

func (b Widget) layoutBadge(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return b.layoutBadgeWithVisual(ctx, gtx, b.visualWithCache(ctx, nil))
}

func (b Widget) layoutBadgeWithVisual(ctx *core.Context, gtx layout.Context, visual theme.BadgeVisual) layout.Dimensions {
	resolved := b.resolveStyle(ctx, gtx, visual)
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if label := b.semanticLabel(); label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if b.content != nil {
				return layoutui.Resolved(ctx, gtx, resolved.label, b.content)
			}
			if b.label != "" {
				return layoutui.Resolved(ctx, gtx, resolved.label, b.labelWidget)
			}
			return layout.Dimensions{}
		})
	})
	return layoutui.Resolved(ctx, gtx, resolved.root, content)
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
