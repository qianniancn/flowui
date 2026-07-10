package flowui

import (
	"image"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func (p PopoverWidget) layoutOverlay(ctx *Context, gtx layout.Context, state *popoverState, trigger image.Point, progress float32) layout.Dimensions {
	if trigger.X <= 0 || trigger.Y <= 0 {
		return layout.Dimensions{}
	}
	overlaySize := popoverOverlaySize(gtx)
	if overlaySize.X <= 0 || overlaySize.Y <= 0 {
		return layout.Dimensions{}
	}

	panelGtx := gtx
	panelGtx.Constraints = p.panelConstraints(ctx, gtx, overlaySize)
	requested := p.placement
	panelCall, panelDims := p.recordPanel(ctx, panelGtx, requested)
	placement := p.resolvedPlacement(ctx, gtx, trigger, panelDims.Size, overlaySize, requested)
	if placement != requested {
		panelCall, panelDims = p.recordPanel(ctx, panelGtx, placement)
	}
	panelPos := p.panelPosition(ctx, gtx, trigger, panelDims.Size, overlaySize, placement)
	panelRect := image.Rectangle{Min: panelPos, Max: panelPos.Add(panelDims.Size)}

	p.layoutDismissAreas(gtx, state, overlaySize, panelRect)

	transform := popoverPanelTransform(ctx, panelRect, progress, placement).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	offset := op.Offset(panelPos.Add(popoverSlideOffset(gtx, progress, placement))).Push(gtx.Ops)
	p.layoutDialogBlocker(gtx, state, panelDims.Size)
	panelCall.Add(gtx.Ops)
	offset.Pop()
	opacity.Pop()
	transform.Pop()

	return layout.Dimensions{Size: overlaySize}
}

func popoverOverlaySize(gtx layout.Context) image.Point {
	size := gtx.Constraints.Max
	if size.X < 0 {
		size.X = 0
	}
	if size.Y < 0 {
		size.Y = 0
	}
	return size
}

func (p PopoverWidget) panelConstraints(ctx *Context, gtx layout.Context, overlay image.Point) layout.Constraints {
	maxWidth := gtx.Dp(ctx.Theme.Components.Popover.MaxWidth)
	if maxWidth <= 0 || maxWidth > overlay.X {
		maxWidth = overlay.X
	}
	return layout.Constraints{
		Max: image.Pt(maxWidth, overlay.Y),
	}
}

func (p PopoverWidget) recordPanel(ctx *Context, gtx layout.Context, placement PopoverPlacement) (op.CallOp, layout.Dimensions) {
	macro := op.Record(gtx.Ops)
	dims := p.layoutPanel(ctx, gtx, placement)
	return macro.Stop(), dims
}

func (p PopoverWidget) layoutPanel(ctx *Context, gtx layout.Context, placement PopoverPlacement) layout.Dimensions {
	theme := ctx.Theme.Components.Popover
	style := popoverStyleFor(ctx.Theme)
	padding := gtx.Dp(theme.Padding)
	contentGtx := gtx
	contentGtx.Constraints.Min = image.Point{}
	contentGtx.Constraints.Max = shrinkPoint(contentGtx.Constraints.Max, padding*2)

	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	func() {
		restore := ctx.pushForeground(style.text)
		defer restore()
		contentDims = p.layoutDialog(ctx, contentGtx, style)
	}()
	contentCall := macro.Stop()

	size := contentDims.Size.Add(image.Pt(padding*2, padding*2))
	size = gtx.Constraints.Constrain(size)
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Radius), 1), min(size.X, size.Y)/2)
	drawPopoverSurface(gtx, ctx.Theme, rect, radius, style)
	if p.showArrow() {
		arrow := image.Pt(gtx.Dp(theme.ArrowWidth), gtx.Dp(theme.ArrowHeight))
		drawPopoverArrowShape(gtx, placement, size, arrow, style)
	}

	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	contentOffset := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	contentCall.Add(gtx.Ops)
	contentOffset.Pop()
	clipStack.Pop()
	return layout.Dimensions{Size: size}
}

func (p PopoverWidget) layoutDialog(ctx *Context, gtx layout.Context, style popoverStyle) layout.Dimensions {
	if p.heading == "" {
		return p.layoutContent(ctx, gtx, style)
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(ctx.Theme.Components.Popover.SectionGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Text(p.heading).
				Size(float32(ctx.Theme.Components.Popover.HeadingSize)).
				Weight(font.Medium).
				Color(style.text).
				Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutContent(ctx, gtx, style)
		}),
	)
}

func (p PopoverWidget) layoutContent(ctx *Context, gtx layout.Context, style popoverStyle) layout.Dimensions {
	if p.content == nil {
		return layout.Dimensions{}
	}
	return p.styleContent(ctx, p.content, style).Layout(ctx, gtx)
}

func (p PopoverWidget) styleContent(ctx *Context, content Widget, style popoverStyle) Widget {
	text, ok := content.(TextWidget)
	if ok {
		if text.size == 0 {
			text = text.Size(float32(ctx.Theme.Components.Popover.BodyTextSize))
		}
		if !text.hasColor {
			text = text.Color(style.muted)
		}
		return text
	}
	return content
}

func (p PopoverWidget) resolvedPlacement(ctx *Context, gtx layout.Context, trigger, panel, overlay image.Point, placement PopoverPlacement) PopoverPlacement {
	if !p.flipEnabled() {
		return placement
	}
	resolved := overlayResolvePlacement(trigger, panel, overlay, p.offsetPx(ctx, gtx), popoverOverlayPlacement(placement))
	return popoverPlacementFromOverlay(resolved)
}

func (p PopoverWidget) panelPosition(ctx *Context, gtx layout.Context, trigger, panel, overlay image.Point, placement PopoverPlacement) image.Point {
	result := overlayResolvePosition(overlayPositionConfig{
		Trigger:       trigger,
		Panel:         panel,
		Bounds:        overlay,
		Offset:        p.offsetPx(ctx, gtx),
		Placement:     popoverOverlayPlacement(placement),
		AvoidOverflow: p.overflowAvoidanceEnabled(),
	})
	return result.Position
}

func (p PopoverWidget) rawPanelPosition(ctx *Context, gtx layout.Context, trigger, panel image.Point, placement PopoverPlacement) image.Point {
	return overlayRawPosition(trigger, panel, p.offsetPx(ctx, gtx), popoverOverlayPlacement(placement))
}

func (p PopoverWidget) offsetPx(ctx *Context, gtx layout.Context) int {
	if p.hasOffset {
		return gtx.Dp(unit.Dp(p.offset))
	}
	return gtx.Dp(ctx.Theme.Components.Popover.Offset)
}

func popoverPanelTransform(ctx *Context, rect image.Rectangle, progress float32, placement PopoverPlacement) op.TransformOp {
	origin := popoverTransformOrigin(rect, placement)
	scale := popoverAnimationScale(ctx, progress)
	return op.Affine(f32.AffineId().Scale(origin, f32.Pt(scale, scale)))
}

func popoverTransformOrigin(rect image.Rectangle, placement PopoverPlacement) f32.Point {
	return overlayTransformOrigin(rect, popoverOverlayPlacement(placement))
}

func popoverAnimationScale(ctx *Context, progress float32) float32 {
	start := ctx.Theme.Components.Popover.AnimationScale
	if start <= 0 || start > 1 {
		start = 0.90
	}
	return start + (1-start)*progress
}

func popoverSlideOffset(gtx layout.Context, progress float32, placement PopoverPlacement) image.Point {
	delta := gtx.Dp(unit.Dp(4))
	return overlaySlideOffset(delta, progress, popoverOverlayPlacement(placement))
}

func popoverSide(placement PopoverPlacement) overlaySide {
	switch placement {
	case PopoverTop, PopoverTopStart, PopoverTopEnd:
		return overlaySideTop
	case PopoverLeft, PopoverLeftStart, PopoverLeftEnd:
		return overlaySideLeft
	case PopoverRight, PopoverRightStart, PopoverRightEnd:
		return overlaySideRight
	default:
		return overlaySideBottom
	}
}

func popoverAlign(placement PopoverPlacement) overlayAlign {
	switch placement {
	case PopoverBottomStart, PopoverTopStart, PopoverLeftStart, PopoverRightStart:
		return overlayAlignStart
	case PopoverBottomEnd, PopoverTopEnd, PopoverLeftEnd, PopoverRightEnd:
		return overlayAlignEnd
	default:
		return overlayAlignCenter
	}
}

func popoverOverlayPlacement(placement PopoverPlacement) overlayPlacement {
	return overlayPlacement{
		side:  popoverSide(placement),
		align: popoverAlign(placement),
	}
}

func popoverPlacementFromOverlay(placement overlayPlacement) PopoverPlacement {
	return popoverWithSideAndAlign(placement.side, placement.align)
}

func popoverWithSide(placement PopoverPlacement, side overlaySide) PopoverPlacement {
	return popoverWithSideAndAlign(side, popoverAlign(placement))
}

func popoverWithSideAndAlign(side overlaySide, align overlayAlign) PopoverPlacement {
	switch side {
	case overlaySideTop:
		switch align {
		case overlayAlignStart:
			return PopoverTopStart
		case overlayAlignEnd:
			return PopoverTopEnd
		default:
			return PopoverTop
		}
	case overlaySideLeft:
		switch align {
		case overlayAlignStart:
			return PopoverLeftStart
		case overlayAlignEnd:
			return PopoverLeftEnd
		default:
			return PopoverLeft
		}
	case overlaySideRight:
		switch align {
		case overlayAlignStart:
			return PopoverRightStart
		case overlayAlignEnd:
			return PopoverRightEnd
		default:
			return PopoverRight
		}
	default:
		switch align {
		case overlayAlignStart:
			return PopoverBottomStart
		case overlayAlignEnd:
			return PopoverBottomEnd
		default:
			return PopoverBottom
		}
	}
}

func popoverAvoidOverflow(pos, panel, overlay image.Point) image.Point {
	return overlayAvoidOverflow(pos, panel, overlay)
}

func (p PopoverWidget) layoutDismissAreas(gtx layout.Context, state *popoverState, overlay image.Point, panel image.Rectangle) {
	areas := overlayDismissRects(image.Rectangle{Max: overlay}, panel)
	for i, area := range areas {
		if area.Empty() {
			continue
		}
		areaGtx := gtx
		areaGtx.Constraints = layout.Exact(area.Size())
		stack := op.Offset(area.Min).Push(gtx.Ops)
		state.dismiss[i].Layout(areaGtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: area.Size()}
		})
		stack.Pop()
	}
}

func (p PopoverWidget) layoutDialogBlocker(gtx layout.Context, state *popoverState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}
