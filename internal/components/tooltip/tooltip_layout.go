package tooltip

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

func (t TooltipWidget) layoutOverlay(ctx *frame.Context, gtx layout.Context, state *tooltipState, trigger image.Rectangle, progress float32) layout.Dimensions {
	if trigger.Empty() {
		return layout.Dimensions{}
	}
	overlaySize := gtx.Constraints.Max
	if overlaySize.X <= 0 || overlaySize.Y <= 0 {
		return layout.Dimensions{}
	}

	panelGtx := gtx.Disabled()
	panelGtx.Constraints = t.panelConstraints(ctx, gtx, overlaySize)
	panelCall, panelDims, panelPlacement := t.recordPanel(ctx, panelGtx)
	result := t.resolvedPosition(ctx, gtx, trigger, panelDims.Size, overlaySize)
	placement := result.Placement.PopoverPlacement()
	panelPos := result.Position

	panelAffine := t.panelAffine(ctx, trigger, panelPos, panelDims.Size, placement, state, progress)
	panelOffset := panelPos.Add(t.slideOffset(ctx, gtx, state, progress, placement))
	panelTransform := panelAffine.Mul(f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))))
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)

	transform := op.Affine(panelAffine).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	offset := op.Offset(panelOffset).Push(gtx.Ops)
	panelCall.Add(gtx.Ops)
	if t.showArrow() {
		theme := frame.ActiveTheme(ctx).Components.Tooltip
		arrowSize := gtx.Dp(theme.ArrowSize)
		panelRadius := min(max(gtx.Dp(theme.Radius), 0), min(panelDims.Size.X, panelDims.Size.Y)/2)
		anchor := tooltipArrowAnchor(trigger, panelPos, panelDims.Size, placement, panelRadius, arrowSize)
		drawTooltipArrow(gtx, placement, panelDims.Size, anchor, arrowSize, gtx.Dp(theme.BorderWidth), tooltipStyleFor(frame.ActiveTheme(ctx)))
	}
	offset.Pop()
	opacity.Pop()
	transform.Pop()

	return layout.Dimensions{Size: overlaySize}
}

func (t TooltipWidget) panelConstraints(ctx *frame.Context, gtx layout.Context, overlaySize image.Point) layout.Constraints {
	maxWidth := gtx.Dp(frame.ActiveTheme(ctx).Components.Tooltip.MaxWidth)
	if maxWidth <= 0 || maxWidth > overlaySize.X {
		maxWidth = overlaySize.X
	}
	return layout.Constraints{Max: image.Pt(maxWidth, overlaySize.Y)}
}

func (t TooltipWidget) recordPanel(ctx *frame.Context, gtx layout.Context) (op.CallOp, layout.Dimensions, frame.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return t.layoutPanel(ctx, gtx)
	})
	return macro.Stop(), dims, placement
}

func (t TooltipWidget) layoutPanel(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Tooltip
	style := tooltipStyleFor(frame.ActiveTheme(ctx))
	padding := gtx.Dp(theme.Padding)
	contentGtx := gtx
	contentGtx.Constraints.Min = image.Point{}
	contentGtx.Constraints.Max.X = max(contentGtx.Constraints.Max.X-padding*2, 0)
	contentGtx.Constraints.Max.Y = max(contentGtx.Constraints.Max.Y-padding*2, 0)

	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	var contentPlacement frame.OverlayPlacement
	func() {
		restore := frame.PushColors(ctx, style.text, style.surface)
		defer restore()
		contentDims, contentPlacement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return t.layoutContent(ctx, contentGtx, style)
		})
	}()
	contentCall := macro.Stop()

	size := gtx.Constraints.Constrain(contentDims.Size.Add(image.Pt(padding*2, padding*2)))
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Radius), 0), min(size.X, size.Y)/2)
	drawTooltipSurface(gtx, frame.ActiveTheme(ctx), rect, radius, style)

	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	contentOffset := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	contentCall.Add(gtx.Ops)
	contentOffset.Pop()
	clipStack.Pop()
	contentPlacement.PlaceOffset(image.Pt(padding, padding))
	return layout.Dimensions{Size: size}
}

func (t TooltipWidget) layoutContent(ctx *frame.Context, gtx layout.Context, style tooltipStyle) layout.Dimensions {
	if t.content == nil {
		return layout.Dimensions{}
	}
	content := t.content
	if value, ok := content.(text.Widget); ok {
		value = value.DefaultSize(float32(frame.ActiveTheme(ctx).Components.Tooltip.TextSize))
		value = value.DefaultColor(style.text)
		content = value
	}
	return content.Layout(ctx, gtx)
}

func (t TooltipWidget) resolvedPosition(ctx *frame.Context, gtx layout.Context, trigger image.Rectangle, panel, bounds image.Point) overlay.PositionResult {
	return overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:          trigger.Size(),
		TriggerOrigin:    trigger.Min,
		HasTriggerOrigin: true,
		Panel:            panel,
		Bounds:           bounds,
		Offset:           t.offsetPx(ctx, gtx),
		Placement:        t.placement.Placement(),
		Flip:             t.flipEnabled(),
		AvoidOverflow:    t.overflowAvoidanceEnabled(),
	})
}

func (t TooltipWidget) offsetPx(ctx *frame.Context, gtx layout.Context) int {
	if t.hasOffset {
		return gtx.Dp(unit.Dp(t.offset))
	}
	theme := frame.ActiveTheme(ctx).Components.Tooltip
	if t.showArrow() {
		return gtx.Dp(theme.ArrowOffset)
	}
	return gtx.Dp(theme.Offset)
}

func (t TooltipWidget) panelAffine(ctx *frame.Context, trigger image.Rectangle, panelPos, panelSize image.Point, placement overlay.PopoverPlacement, state *tooltipState, progress float32) f32.Affine2D {
	origin := tooltipTransformOrigin(trigger, panelPos, panelSize, placement)
	scale := frame.ActiveTheme(ctx).Components.Tooltip.AnimationScale
	if state.exiting() {
		scale = frame.ActiveTheme(ctx).Components.Tooltip.ExitScale
	}
	if scale <= 0 || scale > 1 {
		scale = 0.90
		if state.exiting() {
			scale = 0.95
		}
	}
	scale += (1 - scale) * progress
	return f32.AffineId().Scale(origin, f32.Pt(scale, scale))
}

func tooltipTransformOrigin(trigger image.Rectangle, panelPos, panelSize image.Point, placement overlay.PopoverPlacement) f32.Point {
	return overlay.PanelTransformOriginAt(trigger, panelPos, panelSize, placement.Placement())
}

func tooltipArrowAnchor(trigger image.Rectangle, panelPos, panelSize image.Point, placement overlay.PopoverPlacement, radius, arrowSize int) float32 {
	var anchor float32
	var crossSize int
	if placement.Placement().Side == overlay.SideTop || placement.Placement().Side == overlay.SideBottom {
		anchor = float32(trigger.Min.X + trigger.Dx()/2 - panelPos.X)
		crossSize = panelSize.X
	} else {
		anchor = float32(trigger.Min.Y + trigger.Dy()/2 - panelPos.Y)
		crossSize = panelSize.Y
	}
	halfArrow := float32(max(arrowSize, 0)) / 2
	margin := float32(max(radius, 0)) + halfArrow
	if float32(crossSize) < margin*2 {
		return float32(crossSize) / 2
	}
	return min(max(anchor, margin), float32(crossSize)-margin)
}

func (t TooltipWidget) slideOffset(ctx *frame.Context, gtx layout.Context, state *tooltipState, progress float32, placement overlay.PopoverPlacement) image.Point {
	if state.exiting() {
		return image.Point{}
	}
	distance := gtx.Dp(frame.ActiveTheme(ctx).Components.Tooltip.AnimationDistance)
	return overlay.SlideOffset(distance, progress, placement.Placement())
}
