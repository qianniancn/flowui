package popover

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

func (p Widget) layoutOverlay(ctx *core.Context, gtx layout.Context, state *popoverState, trigger image.Rectangle, progress float32, contentEnabled bool, visual theme.PopoverVisual) layout.Dimensions {
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
	panelGtx.Constraints = p.panelConstraintsWithVisual(gtx, overlaySize, visual)
	panelCall, panelDims, panelPlacement := p.recordPanel(ctx, panelGtx, visual)
	result := p.resolvedPositionWithVisual(ctx, gtx, trigger, panelDims.Size, overlaySize, p.placement, visual)
	placement := result.Placement.PopoverPlacement()
	panelPos := result.Position
	panelRect := image.Rectangle{Min: panelPos, Max: panelPos.Add(panelDims.Size)}
	var arrowSize image.Point
	var arrowRect image.Rectangle
	if p.showArrow() {
		arrowSize = image.Pt(gtx.Dp(visual.Metrics.ArrowWidth), gtx.Dp(visual.Metrics.ArrowHeight))
		arrowRect = popoverArrowRect(image.Rectangle{Max: panelDims.Size}, placement, arrowSize)
	}

	panelAffine := popoverPanelAffineWithVisual(panelRect, progress, placement, visual)
	panelOffset := panelPos.Add(popoverSlideOffsetWithVisual(gtx, progress, placement, visual))
	panelTransform := panelAffine.Mul(f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))))
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(progress)
	protected := []image.Rectangle{
		core.AffineRectBounds(image.Rectangle{Max: panelDims.Size}, panelTransform),
	}
	if !arrowRect.Empty() {
		protected = append(protected, core.AffineRectBounds(arrowRect, panelTransform))
	}
	p.layoutDismissAreas(gtx, state, overlaySize, protected...)
	transform := op.Affine(panelAffine).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	offset := op.Offset(panelOffset).Push(gtx.Ops)
	if !arrowRect.Empty() {
		resolved := p.resolvePanelStyle(ctx, gtx, style.State{Disabled: !contentEnabled, Open: true}, visual)
		visual.Arrow(gtx, theme.PopoverArrowRequest{
			Side: popoverSide(placement), Panel: panelDims.Size, Arrow: arrowSize,
			Color: visual.PanelColor(resolved),
		})
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

func (p Widget) panelConstraintsWithVisual(gtx layout.Context, overlay image.Point, visual theme.PopoverVisual) layout.Constraints {
	maxWidth := gtx.Dp(visual.Metrics.MaxWidth)
	if maxWidth <= 0 || maxWidth > overlay.X {
		maxWidth = overlay.X
	}
	return layout.Constraints{
		Max: image.Pt(maxWidth, overlay.Y),
	}
}

func (p Widget) recordPanel(ctx *core.Context, gtx layout.Context, visual theme.PopoverVisual) (op.CallOp, layout.Dimensions, core.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return p.layoutPanel(ctx, gtx, visual)
	})
	return macro.Stop(), dims, placement
}

func (p Widget) layoutPanel(ctx *core.Context, gtx layout.Context, visual theme.PopoverVisual) layout.Dimensions {
	state := style.State{Disabled: !gtx.Enabled(), Open: p.open}
	resolved := p.resolvePanelStyle(ctx, gtx, state, visual)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return p.layoutDialog(ctx, gtx, visual)
	}))
}

func (p Widget) resolvePanelStyle(ctx *core.Context, gtx layout.Context, state style.State, visual theme.PopoverVisual) style.ResolvedStyle {
	return computed.Resolve(ctx, gtx, core.FullKey(ctx, "popover-style"), state, visual.Defaults, style.Style{}, style.Style{}, visual.Override)
}

func (p Widget) layoutDialog(ctx *core.Context, gtx layout.Context, visual theme.PopoverVisual) layout.Dimensions {
	if p.heading == "" {
		return p.layoutContent(ctx, gtx, visual)
	}
	gap := gtx.Dp(visual.Metrics.SectionGap)
	var headingDims layout.Dimensions
	var contentPlacement core.OverlayPlacement
	dims := layout.Flex{
		Axis: layout.Vertical,
		Gap:  gap,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			heading := text.WithDefaults(text.New(p.heading), visual.Heading)
			headingDims = heading.Layout(ctx, gtx)
			return headingDims
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var contentDims layout.Dimensions
			contentDims, contentPlacement = core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return p.layoutContent(ctx, gtx, visual)
			})
			return contentDims
		}),
	)
	contentPlacement.PlaceOffset(image.Pt(0, headingDims.Size.Y+gap))
	return dims
}

func (p Widget) layoutContent(ctx *core.Context, gtx layout.Context, visual theme.PopoverVisual) layout.Dimensions {
	if p.content == nil {
		return layout.Dimensions{}
	}
	return p.styleContentWithVisual(ctx, p.content, visual).Layout(ctx, gtx)
}

func (p Widget) styleContentWithVisual(ctx *core.Context, content core.Widget, visual theme.PopoverVisual) core.Widget {
	value, ok := content.(text.Widget)
	if ok {
		return text.WithDefaults(value, visual.Content)
	}
	return content
}

func (p Widget) resolvedPositionWithVisual(_ *core.Context, gtx layout.Context, trigger image.Rectangle, panel, bounds image.Point, placement core.PopoverPlacement, visual theme.PopoverVisual) core.PositionResult {
	return core.ResolvePosition(core.PositionConfig{
		Trigger:          trigger.Size(),
		TriggerOrigin:    trigger.Min,
		HasTriggerOrigin: true,
		Panel:            panel,
		Bounds:           bounds,
		Offset:           p.offsetPxWithVisual(gtx, visual),
		Placement:        placement.Placement(),
		Flip:             p.flipEnabled(),
		AvoidOverflow:    p.overflowAvoidanceEnabled(),
	})
}

func (p Widget) offsetPxWithVisual(gtx layout.Context, visual theme.PopoverVisual) int {
	if p.hasOffset {
		return gtx.Dp(unit.Dp(p.offset))
	}
	return gtx.Dp(visual.Metrics.Offset)
}

func popoverPanelAffineWithVisual(rect image.Rectangle, progress float32, placement core.PopoverPlacement, visual theme.PopoverVisual) f32.Affine2D {
	origin := popoverTransformOrigin(rect, placement)
	scale := visual.Metrics.AnimationScale
	if scale <= 0 || scale > 1 {
		scale = .90
	}
	scale += (1 - scale) * progress
	return f32.AffineId().Scale(origin, f32.Pt(scale, scale))
}

func popoverTransformOrigin(rect image.Rectangle, placement core.PopoverPlacement) f32.Point {
	return core.TransformOrigin(rect, placement.Placement())
}

func popoverSlideOffsetWithVisual(gtx layout.Context, progress float32, placement core.PopoverPlacement, visual theme.PopoverVisual) image.Point {
	delta := gtx.Dp(visual.Metrics.AnimationDistance)
	return core.SlideOffset(delta, progress, placement.Placement())
}

func popoverSide(placement core.PopoverPlacement) uint8 {
	switch placement.Placement().Side {
	case core.SideTop:
		return theme.PopoverSideTop
	case core.SideLeft:
		return theme.PopoverSideLeft
	case core.SideRight:
		return theme.PopoverSideRight
	default:
		return theme.PopoverSideBottom
	}
}

func (p Widget) layoutDismissAreas(gtx layout.Context, state *popoverState, bounds image.Point, protected ...image.Rectangle) {
	areas := core.DismissRectsExcluding(image.Rectangle{Max: bounds}, protected...)
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

func (p Widget) layoutArrowBlocker(gtx layout.Context, state *popoverState, rect image.Rectangle) {
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

func popoverArrowRect(panel image.Rectangle, placement core.PopoverPlacement, arrow image.Point) image.Rectangle {
	if panel.Empty() || arrow.X <= 0 || arrow.Y <= 0 {
		return image.Rectangle{}
	}
	switch placement.Placement().Side {
	case core.SideTop:
		x := panel.Min.X + panel.Dx()/2 - arrow.X/2
		return image.Rect(x, panel.Max.Y, x+arrow.X, panel.Max.Y+arrow.Y)
	case core.SideLeft:
		y := panel.Min.Y + panel.Dy()/2 - arrow.X/2
		return image.Rect(panel.Max.X, y, panel.Max.X+arrow.Y, y+arrow.X)
	case core.SideRight:
		y := panel.Min.Y + panel.Dy()/2 - arrow.X/2
		return image.Rect(panel.Min.X-arrow.Y, y, panel.Min.X, y+arrow.X)
	default:
		x := panel.Min.X + panel.Dx()/2 - arrow.X/2
		return image.Rect(x, panel.Min.Y-arrow.Y, x+arrow.X, panel.Min.Y)
	}
}

func (p Widget) layoutDialogBlocker(gtx layout.Context, state *popoverState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}
