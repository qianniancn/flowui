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
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

func (p PopoverWidget) layoutOverlay(ctx *frame.Context, gtx layout.Context, state *popoverState, trigger image.Rectangle, progress float32, contentEnabled bool) layout.Dimensions {
	if trigger.Empty() {
		return layout.Dimensions{}
	}
	overlaySize := popoverOverlaySize(gtx)
	if overlaySize.X <= 0 || overlaySize.Y <= 0 {
		return layout.Dimensions{}
	}

	panelGtx := gtx
	if !contentEnabled {
		panelGtx = panelGtx.Disabled()
	}
	panelGtx.Constraints = p.panelConstraints(ctx, gtx, overlaySize)
	panelCall, panelDims, panelPlacement := p.recordPanel(ctx, panelGtx)
	result := p.resolvedPosition(ctx, gtx, trigger, panelDims.Size, overlaySize, p.placement)
	placement := result.Placement.PopoverPlacement()
	panelPos := result.Position
	panelRect := image.Rectangle{Min: panelPos, Max: panelPos.Add(panelDims.Size)}
	var arrowSize image.Point
	var arrowRect image.Rectangle
	if p.showArrow() {
		theme := frame.ActiveTheme(ctx).Components.Popover
		arrowSize = image.Pt(gtx.Dp(theme.ArrowWidth), gtx.Dp(theme.ArrowHeight))
		arrowRect = popoverArrowRect(image.Rectangle{Max: panelDims.Size}, placement, arrowSize)
	}

	panelAffine := popoverPanelAffine(ctx, panelRect, progress, placement)
	panelOffset := panelPos.Add(popoverSlideOffset(ctx, gtx, progress, placement))
	panelTransform := panelAffine.Mul(f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))))
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	protected := []image.Rectangle{
		overlay.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform),
	}
	if !arrowRect.Empty() {
		protected = append(protected, overlay.AffineRectBounds(arrowRect, panelTransform))
	}
	p.layoutDismissAreas(gtx, state, overlaySize, protected...)
	transform := op.Affine(panelAffine).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	offset := op.Offset(panelOffset).Push(gtx.Ops)
	if !arrowRect.Empty() {
		drawPopoverArrowShape(gtx, placement, panelDims.Size, arrowSize, popoverStyleFor(frame.ActiveTheme(ctx)))
		p.layoutArrowBlocker(gtx, state, arrowRect)
	}
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

func (p PopoverWidget) recordPanel(ctx *frame.Context, gtx layout.Context) (op.CallOp, layout.Dimensions, frame.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return layoutui.LayoutStyled(ctx, gtx, frame.FullKey(ctx, p.key), flowstyle.StyleState{
			Disabled: !gtx.Enabled(),
			Open:     p.open,
		}, p.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return p.layoutPanel(ctx, gtx)
		}))
	})
	return macro.Stop(), dims, placement
}

func (p PopoverWidget) layoutPanel(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.Popover
	style := popoverStyleFor(frame.ActiveTheme(ctx))
	padding := gtx.Dp(theme.Padding)
	contentGtx := gtx
	contentGtx.Constraints.Min = image.Point{}
	contentGtx.Constraints.Max = shrinkPoint(contentGtx.Constraints.Max, padding*2)

	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	var contentPlacement frame.OverlayPlacement
	func() {
		restore := frame.PushColors(ctx, style.text, style.surface)
		defer restore()
		contentDims, contentPlacement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return p.layoutDialog(ctx, contentGtx, style)
		})
	}()
	contentCall := macro.Stop()

	size := contentDims.Size.Add(image.Pt(padding*2, padding*2))
	size = gtx.Constraints.Constrain(size)
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Radius), 1), min(size.X, size.Y)/2)
	drawPopoverSurface(gtx, frame.ActiveTheme(ctx), rect, radius, style)

	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	contentOffset := op.Offset(image.Pt(padding, padding)).Push(gtx.Ops)
	contentCall.Add(gtx.Ops)
	contentOffset.Pop()
	clipStack.Pop()
	contentPlacement.PlaceOffset(image.Pt(padding, padding))
	return layout.Dimensions{Size: size}
}

func (p PopoverWidget) layoutDialog(ctx *frame.Context, gtx layout.Context, style popoverStyle) layout.Dimensions {
	if p.heading == "" {
		return p.layoutContent(ctx, gtx, style)
	}
	gap := gtx.Dp(frame.ActiveTheme(ctx).Components.Popover.SectionGap)
	var headingDims layout.Dimensions
	var contentPlacement frame.OverlayPlacement
	dims := layout.Flex{
		Axis: layout.Vertical,
		Gap:  gap,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			headingDims = text.New(p.heading).
				Size(float32(frame.ActiveTheme(ctx).Components.Popover.HeadingSize)).
				Weight(font.Medium).
				Color(style.text).
				Layout(ctx, gtx)
			return headingDims
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var contentDims layout.Dimensions
			contentDims, contentPlacement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return p.layoutContent(ctx, gtx, style)
			})
			return contentDims
		}),
	)
	contentPlacement.PlaceOffset(image.Pt(0, headingDims.Size.Y+gap))
	return dims
}

func (p PopoverWidget) layoutContent(ctx *frame.Context, gtx layout.Context, style popoverStyle) layout.Dimensions {
	if p.content == nil {
		return layout.Dimensions{}
	}
	return p.styleContent(ctx, p.content, style).Layout(ctx, gtx)
}

func (p PopoverWidget) styleContent(ctx *frame.Context, content frame.Widget, style popoverStyle) frame.Widget {
	value, ok := content.(text.Widget)
	if ok {
		return text.WithDefaults(value, flowstyle.Style{}.
			FontSize(frame.ActiveTheme(ctx).Components.Popover.BodyTextSize).
			TextColor(flowstyle.SolidColor{Color: style.muted}),
		)
	}
	return content
}

func (p PopoverWidget) resolvedPosition(ctx *frame.Context, gtx layout.Context, trigger image.Rectangle, panel, bounds image.Point, placement overlay.PopoverPlacement) overlay.PositionResult {
	return overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:          trigger.Size(),
		TriggerOrigin:    trigger.Min,
		HasTriggerOrigin: true,
		Panel:            panel,
		Bounds:           bounds,
		Offset:           p.offsetPx(ctx, gtx),
		Placement:        placement.Placement(),
		Flip:             p.flipEnabled(),
		AvoidOverflow:    p.overflowAvoidanceEnabled(),
	})
}

func (p PopoverWidget) offsetPx(ctx *frame.Context, gtx layout.Context) int {
	if p.hasOffset {
		return gtx.Dp(unit.Dp(p.offset))
	}
	return gtx.Dp(frame.ActiveTheme(ctx).Components.Popover.Offset)
}

func popoverPanelAffine(ctx *frame.Context, rect image.Rectangle, progress float32, placement overlay.PopoverPlacement) f32.Affine2D {
	origin := popoverTransformOrigin(rect, placement)
	scale := popoverAnimationScale(ctx, progress)
	return f32.AffineId().Scale(origin, f32.Pt(scale, scale))
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

func popoverSlideOffset(ctx *frame.Context, gtx layout.Context, progress float32, placement overlay.PopoverPlacement) image.Point {
	delta := gtx.Dp(frame.ActiveTheme(ctx).Components.Popover.AnimationDistance)
	return overlay.SlideOffset(delta, progress, placement.Placement())
}

func popoverAvoidOverflow(pos, panel, bounds image.Point) image.Point {
	return overlay.AvoidOverflow(pos, panel, bounds)
}

func (p PopoverWidget) layoutDismissAreas(gtx layout.Context, state *popoverState, bounds image.Point, protected ...image.Rectangle) {
	areas := overlay.DismissRectsExcluding(image.Rectangle{Max: bounds}, protected...)
	for i, area := range areas {
		if i >= len(state.dismiss) {
			break
		}
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

func (p PopoverWidget) layoutArrowBlocker(gtx layout.Context, state *popoverState, rect image.Rectangle) {
	if rect.Empty() {
		return
	}
	arrowGtx := gtx
	arrowGtx.Constraints = layout.Exact(rect.Size())
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	state.arrow.Layout(arrowGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: rect.Size()}
	})
	stack.Pop()
}

func popoverArrowRect(panel image.Rectangle, placement overlay.PopoverPlacement, arrow image.Point) image.Rectangle {
	if panel.Empty() || arrow.X <= 0 || arrow.Y <= 0 {
		return image.Rectangle{}
	}
	switch placement.Placement().Side {
	case overlay.SideTop:
		x := panel.Min.X + panel.Dx()/2 - arrow.X/2
		return image.Rect(x, panel.Max.Y, x+arrow.X, panel.Max.Y+arrow.Y)
	case overlay.SideLeft:
		y := panel.Min.Y + panel.Dy()/2 - arrow.X/2
		return image.Rect(panel.Max.X, y, panel.Max.X+arrow.Y, y+arrow.X)
	case overlay.SideRight:
		y := panel.Min.Y + panel.Dy()/2 - arrow.X/2
		return image.Rect(panel.Min.X-arrow.Y, y, panel.Min.X, y+arrow.X)
	default:
		x := panel.Min.X + panel.Dx()/2 - arrow.X/2
		return image.Rect(x, panel.Min.Y-arrow.Y, x+arrow.X, panel.Min.Y)
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
