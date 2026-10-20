package tooltip

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

func (p Popup) panelConstraints(gtx layout.Context, overlaySize image.Point, visual theme.TooltipVisual) layout.Constraints {
	maxWidth := gtx.Dp(visual.Metrics.MaxWidth)
	if maxWidth <= 0 || maxWidth > overlaySize.X {
		maxWidth = overlaySize.X
	}
	return layout.Constraints{Max: image.Pt(maxWidth, overlaySize.Y)}
}

func (p Popup) recordPanel(ctx *core.Context, gtx layout.Context, visual theme.TooltipVisual) (op.CallOp, layout.Dimensions, core.OverlayPlacement) {
	macro := op.Record(gtx.Ops)
	dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return p.layoutPanel(ctx, gtx, visual)
	})
	return macro.Stop(), dims, placement
}

func (p Popup) layoutPanel(ctx *core.Context, gtx layout.Context, visual theme.TooltipVisual) layout.Dimensions {
	state := p.panelState(gtx)
	resolved := p.resolvePanelStyle(ctx, gtx, state, visual)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return p.layoutContent(ctx, gtx, visual)
	}))
}

func (p Popup) layoutContent(ctx *core.Context, gtx layout.Context, visual theme.TooltipVisual) layout.Dimensions {
	if p.content == nil {
		return layout.Dimensions{}
	}
	content := p.content
	if value, ok := content.(text.Widget); ok {
		content = text.WithDefaults(value, visual.Content)
	}
	return content.Layout(ctx, gtx)
}

func (p Popup) resolvedPosition(gtx layout.Context, trigger image.Rectangle, panel, bounds image.Point, visual theme.TooltipVisual) core.PositionResult {
	return core.ResolvePosition(core.PositionConfig{
		Trigger:          trigger.Size(),
		TriggerOrigin:    trigger.Min,
		HasTriggerOrigin: true,
		Panel:            panel,
		Bounds:           bounds,
		Offset:           p.offsetPx(gtx, visual),
		Placement:        p.placement.Placement(),
		Flip:             p.flipEnabled(),
		AvoidOverflow:    p.overflowAvoidanceEnabled(),
	})
}

func (p Popup) offsetPx(gtx layout.Context, visual theme.TooltipVisual) int {
	if p.hasOffset {
		return gtx.Dp(p.offset)
	}
	if p.arrow {
		return gtx.Dp(visual.Metrics.ArrowOffset)
	}
	return gtx.Dp(visual.Metrics.Offset)
}

func (p Popup) panelAffine(trigger image.Rectangle, panelPos, panelSize image.Point, placement core.PopoverPlacement, visual theme.TooltipVisual) f32.Affine2D {
	if !p.transformMotionEnabled() {
		return f32.AffineId()
	}
	origin := core.PanelTransformOriginAt(trigger, panelPos, panelSize, placement.Placement())
	scale := visual.Metrics.AnimationScale
	if p.exiting {
		scale = visual.Metrics.ExitScale
	}
	scale += (1 - scale) * p.progress
	return f32.AffineId().Scale(origin, f32.Pt(scale, scale))
}

func (p Popup) slideOffset(gtx layout.Context, placement core.PopoverPlacement, visual theme.TooltipVisual) image.Point {
	if p.exiting || !p.transformMotionEnabled() {
		return image.Point{}
	}
	distance := gtx.Dp(visual.Metrics.AnimationDistance)
	return core.SlideOffset(distance, p.progress, placement.Placement())
}

func (p Popup) resolvePanelStyle(ctx *core.Context, gtx layout.Context, state style.State, visual theme.TooltipVisual) style.ResolvedStyle {
	key := p.styleKey
	if key == "" {
		key = core.FullKey(ctx, "tooltip-style")
	}
	return computed.Resolve(ctx, gtx, key, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override)
}

func (p Popup) panelState(gtx layout.Context) style.State {
	if p.styleState != (style.State{}) {
		return p.styleState
	}
	return style.State{Disabled: !gtx.Enabled(), Open: !p.exiting}
}

func tooltipSide(placement core.PopoverPlacement) uint8 {
	switch placement.Placement().Side {
	case core.SideTop:
		return theme.TooltipSideTop
	case core.SideLeft:
		return theme.TooltipSideLeft
	case core.SideRight:
		return theme.TooltipSideRight
	default:
		return theme.TooltipSideBottom
	}
}

func (p Popup) transformMotionEnabled() bool {
	return !p.hasTransformMotion || p.transformMotion
}
