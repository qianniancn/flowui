package badge

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
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
	activeTheme := frame.ActiveTheme(ctx)
	style := badgeStyleFor(activeTheme, b.color, b.variant, b.size, activeTheme.Palette.Background)
	minimum := min(gtx.Dp(style.minSize), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	contentGtx := gtx
	contentGtx.Constraints.Min = image.Point{}
	padding := 0
	if b.content == nil && b.label != "" {
		padding = gtx.Dp(style.paddingX)
	}
	contentGtx.Constraints.Max.X = max(contentGtx.Constraints.Max.X-padding*2, 0)
	content := b.recordContent(ctx, contentGtx, style)
	width := max(minimum, content.dims.Size.X+padding*2)
	height := max(minimum, content.dims.Size.Y)
	size := gtx.Constraints.Constrain(image.Pt(width, height))
	radius := min(max(gtx.Dp(style.radius), 0), min(size.X, size.Y)/2)
	b.drawBadge(gtx, size, radius, style)

	root := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
	if label := b.semanticLabel(); label != "" {
		semantic.LabelOp(label).Add(gtx.Ops)
	}
	position := image.Pt((size.X-content.dims.Size.X)/2, (size.Y-content.dims.Size.Y)/2)
	content.placement.PlaceOffset(position)
	offset := op.Offset(position).Push(gtx.Ops)
	content.call.Add(gtx.Ops)
	offset.Pop()
	root.Pop()
	return layout.Dimensions{Size: size}
}

func (b Widget) recordContent(ctx *frame.Context, gtx layout.Context, style badgeStyle) recordedBadgeChild {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		restore := frame.PushColors(ctx, style.foreground, style.background)
		defer restore()
		if b.content != nil {
			return b.content.Layout(ctx, gtx)
		}
		if b.label == "" {
			return layout.Dimensions{}
		}
		label := material.Label(frame.ActiveTheme(ctx).Material, style.textSize, b.label)
		label.Color = style.foreground
		label.Font.Weight = font.Medium
		label.LineHeight = style.lineHeight
		label.MaxLines = 1
		return label.Layout(gtx)
	})
	return recordedBadgeChild{call: macro.Stop(), dims: dims, placement: placement}
}

func (b Widget) drawBadge(gtx layout.Context, size image.Point, radius int, style badgeStyle) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	rect := image.Rectangle{Max: size}
	paint.FillShape(gtx.Ops, style.border, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	border := min(max(gtx.Dp(style.borderWidth), 0), min(size.X, size.Y)/2)
	inner := rect.Inset(border)
	if !inner.Empty() {
		paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(inner, max(radius-border, 0)).Op(gtx.Ops))
	}
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
