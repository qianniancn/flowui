package tooltip

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

func (p Popup) panelConstraints(ctx *frame.Context, gtx layout.Context, overlaySize image.Point) layout.Constraints {
	maxWidth := gtx.Dp(frame.ActiveTheme(ctx).Components.Tooltip.MaxWidth)
	if maxWidth <= 0 || maxWidth > overlaySize.X {
		maxWidth = overlaySize.X
	}
	return layout.Constraints{Max: image.Pt(maxWidth, overlaySize.Y)}
}

func (p Popup) recordPanel(ctx *frame.Context, gtx layout.Context) (op.CallOp, layout.Dimensions, frame.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return p.layoutPanel(ctx, gtx)
	})
	return macro.Stop(), dims, placement
}

func (p Popup) layoutPanel(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	theme := activeTheme.Components.Tooltip
	style := tooltipStyleFor(activeTheme)
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
			return p.layoutContent(ctx, contentGtx, style)
		})
	}()
	contentCall := macro.Stop()

	size := gtx.Constraints.Constrain(contentDims.Size.Add(image.Pt(padding*2, padding*2)))
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Radius), 0), min(size.X, size.Y)/2)
	drawTooltipSurface(gtx, activeTheme, rect, radius, style)

	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	contentOffset := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	contentCall.Add(gtx.Ops)
	contentOffset.Pop()
	clipStack.Pop()
	contentPlacement.PlaceOffset(image.Pt(padding, padding))
	return layout.Dimensions{Size: size}
}

func (p Popup) layoutContent(ctx *frame.Context, gtx layout.Context, style tooltipStyle) layout.Dimensions {
	if p.content == nil {
		return layout.Dimensions{}
	}
	content := p.content
	if value, ok := content.(text.Widget); ok {
		value = value.DefaultSize(float32(frame.ActiveTheme(ctx).Components.Tooltip.TextSize))
		value = value.DefaultColor(style.text)
		content = value
	}
	return content.Layout(ctx, gtx)
}

func (p Popup) resolvedPosition(ctx *frame.Context, gtx layout.Context, trigger image.Rectangle, panel, bounds image.Point) overlay.PositionResult {
	return overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:          trigger.Size(),
		TriggerOrigin:    trigger.Min,
		HasTriggerOrigin: true,
		Panel:            panel,
		Bounds:           bounds,
		Offset:           p.offsetPx(ctx, gtx),
		Placement:        p.placement.Placement(),
		Flip:             p.flipEnabled(),
		AvoidOverflow:    p.overflowAvoidanceEnabled(),
	})
}

func (p Popup) offsetPx(ctx *frame.Context, gtx layout.Context) int {
	if p.hasOffset {
		return gtx.Dp(p.offset)
	}
	theme := frame.ActiveTheme(ctx).Components.Tooltip
	if p.arrow {
		return gtx.Dp(theme.ArrowOffset)
	}
	return gtx.Dp(theme.Offset)
}

func (p Popup) panelAffine(ctx *frame.Context, trigger image.Rectangle, panelPos, panelSize image.Point, placement overlay.PopoverPlacement) f32.Affine2D {
	if !p.transformMotionEnabled() {
		return f32.AffineId()
	}
	origin := tooltipTransformOrigin(trigger, panelPos, panelSize, placement)
	theme := frame.ActiveTheme(ctx).Components.Tooltip
	scale := theme.AnimationScale
	if p.exiting {
		scale = theme.ExitScale
	}
	if scale <= 0 || scale > 1 {
		scale = 0.90
		if p.exiting {
			scale = 0.95
		}
	}
	scale += (1 - scale) * p.progress
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

func (p Popup) slideOffset(ctx *frame.Context, gtx layout.Context, placement overlay.PopoverPlacement) image.Point {
	if p.exiting || !p.transformMotionEnabled() {
		return image.Point{}
	}
	distance := gtx.Dp(frame.ActiveTheme(ctx).Components.Tooltip.AnimationDistance)
	return overlay.SlideOffset(distance, p.progress, placement.Placement())
}

func (p Popup) transformMotionEnabled() bool {
	return !p.hasTransformMotion || p.transformMotion
}
