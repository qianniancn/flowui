package popover

import (
	"image"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

func (p PopoverWidget) layoutOverlay(ctx *frame.Context, gtx layout.Context, state *popoverState, trigger image.Point, progress float32) layout.Dimensions {
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

func shrinkPoint(p image.Point, amount int) image.Point {
	return image.Pt(max(p.X-amount, 0), max(p.Y-amount, 0))
}

func (p PopoverWidget) panelConstraints(ctx *frame.Context, gtx layout.Context, overlay image.Point) layout.Constraints {
	maxWidth := gtx.Dp(frame.ActiveTheme(ctx).Components.Popover.MaxWidth)
	if maxWidth <= 0 || maxWidth > overlay.X {
		maxWidth = overlay.X
	}
	return layout.Constraints{
		Max: image.Pt(maxWidth, overlay.Y),
	}
}

func (p PopoverWidget) recordPanel(ctx *frame.Context, gtx layout.Context, placement overlay.PopoverPlacement) (op.CallOp, layout.Dimensions) {
	macro := op.Record(gtx.Ops)
	dims := p.layoutPanel(ctx, gtx, placement)
	return macro.Stop(), dims
}

func (p PopoverWidget) layoutPanel(ctx *frame.Context, gtx layout.Context, placement overlay.PopoverPlacement) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Popover
	style := popoverStyleFor(frame.ActiveTheme(ctx))
	padding := gtx.Dp(theme.Padding)
	contentGtx := gtx
	contentGtx.Constraints.Min = image.Point{}
	contentGtx.Constraints.Max = shrinkPoint(contentGtx.Constraints.Max, padding*2)

	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	func() {
		restore := frame.PushColors(ctx, style.text, style.surface)
		defer restore()
		contentDims = p.layoutDialog(ctx, contentGtx, style)
	}()
	contentCall := macro.Stop()

	size := contentDims.Size.Add(image.Pt(padding*2, padding*2))
	size = gtx.Constraints.Constrain(size)
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Radius), 1), min(size.X, size.Y)/2)
	drawPopoverSurface(gtx, frame.ActiveTheme(ctx), rect, radius, style)
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

func (p PopoverWidget) layoutDialog(ctx *frame.Context, gtx layout.Context, style popoverStyle) layout.Dimensions {
	if p.heading == "" {
		return p.layoutContent(ctx, gtx, style)
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(frame.ActiveTheme(ctx).Components.Popover.SectionGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(p.heading).
				Size(float32(frame.ActiveTheme(ctx).Components.Popover.HeadingSize)).
				Weight(font.Medium).
				Color(style.text).
				Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutContent(ctx, gtx, style)
		}),
	)
}

func (p PopoverWidget) layoutContent(ctx *frame.Context, gtx layout.Context, style popoverStyle) layout.Dimensions {
	if p.content == nil {
		return layout.Dimensions{}
	}
	return p.styleContent(ctx, p.content, style).Layout(ctx, gtx)
}

func (p PopoverWidget) styleContent(ctx *frame.Context, content frame.Widget, style popoverStyle) frame.Widget {
	text, ok := content.(text.Widget)
	if ok {
		text = text.DefaultSize(float32(frame.ActiveTheme(ctx).Components.Popover.BodyTextSize))
		text = text.DefaultColor(style.muted)
		return text
	}
	return content
}

func (p PopoverWidget) resolvedPlacement(ctx *frame.Context, gtx layout.Context, trigger, panel, bounds image.Point, placement overlay.PopoverPlacement) overlay.PopoverPlacement {
	if !p.flipEnabled() {
		return placement
	}
	resolved := overlay.ResolvePlacement(trigger, panel, bounds, p.offsetPx(ctx, gtx), placement.Placement())
	return resolved.PopoverPlacement()
}

func (p PopoverWidget) panelPosition(ctx *frame.Context, gtx layout.Context, trigger, panel, bounds image.Point, placement overlay.PopoverPlacement) image.Point {
	result := overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:       trigger,
		Panel:         panel,
		Bounds:        bounds,
		Offset:        p.offsetPx(ctx, gtx),
		Placement:     placement.Placement(),
		AvoidOverflow: p.overflowAvoidanceEnabled(),
	})
	return result.Position
}

func (p PopoverWidget) rawPanelPosition(ctx *frame.Context, gtx layout.Context, trigger, panel image.Point, placement overlay.PopoverPlacement) image.Point {
	return overlay.RawPosition(trigger, panel, p.offsetPx(ctx, gtx), placement.Placement())
}

func (p PopoverWidget) offsetPx(ctx *frame.Context, gtx layout.Context) int {
	if p.hasOffset {
		return gtx.Dp(unit.Dp(p.offset))
	}
	return gtx.Dp(frame.ActiveTheme(ctx).Components.Popover.Offset)
}

func popoverPanelTransform(ctx *frame.Context, rect image.Rectangle, progress float32, placement overlay.PopoverPlacement) op.TransformOp {
	origin := popoverTransformOrigin(rect, placement)
	scale := popoverAnimationScale(ctx, progress)
	return op.Affine(f32.AffineId().Scale(origin, f32.Pt(scale, scale)))
}

func popoverTransformOrigin(rect image.Rectangle, placement overlay.PopoverPlacement) f32.Point {
	return overlay.TransformOrigin(rect, placement.Placement())
}

func popoverAnimationScale(ctx *frame.Context, progress float32) float32 {
	start := frame.ActiveTheme(ctx).Components.Popover.AnimationScale
	if start <= 0 || start > 1 {
		start = 0.90
	}
	return start + (1-start)*progress
}

func popoverSlideOffset(gtx layout.Context, progress float32, placement overlay.PopoverPlacement) image.Point {
	delta := gtx.Dp(unit.Dp(4))
	return overlay.SlideOffset(delta, progress, placement.Placement())
}

func popoverAvoidOverflow(pos, panel, bounds image.Point) image.Point {
	return overlay.AvoidOverflow(pos, panel, bounds)
}

func (p PopoverWidget) layoutDismissAreas(gtx layout.Context, state *popoverState, bounds image.Point, panel image.Rectangle) {
	areas := overlay.DismissRects(image.Rectangle{Max: bounds}, panel)
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
