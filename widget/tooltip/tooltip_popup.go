package tooltip

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Popup is the controlled presentation layer shared by triggered and
// data-driven tooltips.
type Popup struct {
	content            core.Widget
	placement          core.PopoverPlacement
	offset             unit.Dp
	hasOffset          bool
	shouldFlip         bool
	hasShouldFlip      bool
	avoidOverflow      bool
	hasAvoidOverflow   bool
	arrow              bool
	transformMotion    bool
	hasTransformMotion bool
	progress           float32
	exiting            bool
	override           *theme.TooltipOverride
	visualCache        *tooltipVisualCache
	styleKey           string
	styleState         style.State
}

func NewPopup(content core.Widget) Popup {
	return Popup{
		content:     content,
		placement:   core.PopoverTop,
		progress:    1,
		visualCache: &tooltipVisualCache{},
	}
}

func (p Popup) Placement(placement core.PopoverPlacement) Popup {
	p.placement = placement
	return p
}

func (p Popup) Offset(offset unit.Dp) Popup {
	p.offset = offset
	p.hasOffset = true
	return p
}

func (p Popup) ShouldFlip(shouldFlip bool) Popup {
	p.shouldFlip = shouldFlip
	p.hasShouldFlip = true
	return p
}

func (p Popup) AvoidOverflow(avoidOverflow bool) Popup {
	p.avoidOverflow = avoidOverflow
	p.hasAvoidOverflow = true
	return p
}

func (p Popup) Arrow(show bool) Popup {
	p.arrow = show
	return p
}

// TransformMotion controls the popup scale and slide animation. Opacity is
// still controlled by Progress.
func (p Popup) TransformMotion(enabled bool) Popup {
	p.transformMotion = enabled
	p.hasTransformMotion = true
	return p
}

func (p Popup) Progress(progress float32) Popup {
	p.progress = min(max(progress, 0), 1)
	return p
}

func (p Popup) Exiting(exiting bool) Popup {
	p.exiting = exiting
	return p
}

// Layout paints the panel inline, in the caller's ops and coordinate space.
// Use it only from a host that neither clips nor scrolls: an inline panel joins
// that host's paint and visual-overflow reporting. Prefer Overlay everywhere
// else, including charts, scrolled pages, and anything inside a card.
func (p Popup) Layout(ctx *core.Context, gtx layout.Context, anchor image.Rectangle) layout.Dimensions {
	visual := p.visualWithCache(ctx, p.override)
	return p.layoutWithVisual(ctx, gtx, anchor, visual)
}

// Overlay registers the popup as a root popup overlay anchored at anchor, which
// is in the caller's local coordinates. Key must be stable for the popup's
// lifetime and unique within its key scope.
//
// Use Overlay instead of Layout from content that scrolls or clips. An inline
// panel joins its host's paint and layout, so its shadow reports a visual
// outset to the enclosing clipping viewport; that viewport then reserves space
// for it, which moves the host by twice the outset on the following frame. When
// the host is the trigger, the reservation moves it out from under the pointer,
// the pointer leaves, the panel exits, the reservation disappears, and the
// whole sequence oscillates. A root overlay sits outside that layout and
// painting order, so it neither reports into the host nor gets clipped by an
// ancestor.
func (p Popup) Overlay(ctx *core.Context, key string, anchor image.Rectangle) {
	if key == "" {
		panic("flowui: empty tooltip overlay key")
	}
	if anchor.Empty() || p.content == nil {
		return
	}
	visual := p.visualWithCache(ctx, p.override)
	panel := p
	panel.styleKey = key
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       key,
		Layer:     core.OverlayLayerPopup,
		Anchor:    anchor,
		HasAnchor: true,
		Disabled:  true,
		Passive:   true,
		Layout: func(gtx layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			return panel.layoutWithVisual(ctx, gtx, anchor, visual)
		},
	})
}

// Overlay registers the styled popup as a root overlay; see Popup.Overlay.
func (p Popup) layoutWithVisual(ctx *core.Context, gtx layout.Context, anchor image.Rectangle, visual theme.TooltipVisual) layout.Dimensions {
	if anchor.Empty() || p.content == nil {
		return layout.Dimensions{}
	}
	bounds := gtx.Constraints.Max
	if bounds.X <= 0 || bounds.Y <= 0 {
		return layout.Dimensions{}
	}

	panelGtx := gtx.Disabled()
	panelGtx.Constraints = p.panelConstraints(gtx, bounds, visual)
	panelCall, panelDims, panelPlacement := p.recordPanel(ctx, panelGtx, visual)
	result := p.resolvedPosition(gtx, anchor, panelDims.Size, bounds, visual)
	placement := result.Placement.PopoverPlacement()
	panelPos := result.Position

	panelAffine := p.panelAffine(anchor, panelPos, panelDims.Size, placement, visual)
	panelOffset := panelPos.Add(p.slideOffset(gtx, placement, visual))
	panelTransform := panelAffine.Mul(f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))))
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(p.progress)

	transform := op.Affine(panelAffine).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, p.progress)
	offset := op.Offset(panelOffset).Push(gtx.Ops)
	panelCall.Add(gtx.Ops)
	if p.arrow {
		arrowSize := gtx.Dp(visual.Metrics.ArrowSize)
		panelRadius := min(max(gtx.Dp(visual.Metrics.Radius), 0), min(panelDims.Size.X, panelDims.Size.Y)/2)
		arrowAnchor := core.ArrowAnchor(anchor, panelPos, panelDims.Size, placement, panelRadius, arrowSize)
		resolved := p.resolvePanelStyle(ctx, gtx, style.State{Disabled: !gtx.Enabled(), Open: !p.exiting}, visual)
		visual.Arrow(gtx, theme.TooltipArrowRequest{
			Side:        tooltipSide(placement),
			Panel:       panelDims.Size,
			Anchor:      arrowAnchor,
			ArrowSize:   arrowSize,
			BorderWidth: gtx.Dp(visual.Metrics.BorderWidth),
			Colors:      visual.PanelColor(resolved),
		})
	}
	offset.Pop()
	opacity.Pop()
	transform.Pop()

	return layout.Dimensions{Size: bounds}
}

func (p Popup) flipEnabled() bool {
	return !p.hasShouldFlip || p.shouldFlip
}

func (p Popup) overflowAvoidanceEnabled() bool {
	return !p.hasAvoidOverflow || p.avoidOverflow
}

func (t Widget) popup(progress float32, exiting bool) Popup {
	p := NewPopup(t.content).
		Placement(t.placement).
		Progress(progress).
		Exiting(exiting).
		Arrow(t.arrow)
	if t.hasOffset {
		p = p.Offset(unit.Dp(t.offset))
	}
	if t.hasShouldFlip {
		p = p.ShouldFlip(t.shouldFlip)
	}
	if t.hasAvoidOverflow {
		p = p.AvoidOverflow(t.avoidOverflow)
	}
	return p
}

// PopupTransition animates a controlled tooltip without introducing show or
// close delays.
type PopupTransition struct {
	value         float32
	from          float32
	to            float32
	at            time.Time
	ready         bool
	position      f32.Point
	positionFrom  f32.Point
	positionTo    f32.Point
	positionAt    time.Time
	positionReady bool
	visualCache   tooltipVisualCache
}

func (t *PopupTransition) Progress(gtx layout.Context, visible bool, motions ...theme.MotionTheme) float32 {
	return t.ProgressWithDurations(gtx, visible, 0, 0, motions...)
}

// ProgressWithContext resolves tooltip motion metrics from the active visual
// provider and reuses them for this transition's lifetime.
func (t *PopupTransition) ProgressWithContext(ctx *core.Context, gtx layout.Context, visible bool) float32 {
	visual := tooltipVisualWithCache(ctx, nil, &t.visualCache)
	return t.ProgressWithDurations(gtx, visible, visual.Metrics.EnterDuration, visual.Metrics.ExitDuration, visual.Motion)
}

func (t *PopupTransition) ProgressWithDurations(gtx layout.Context, visible bool, enter, exit time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if visible {
		target = 1
	}
	if !t.ready {
		t.at = gtx.Now
		t.ready = true
	}
	if target != t.to {
		t.from = t.value
		t.to = target
		t.at = gtx.Now
	}
	if t.from == t.to {
		t.value = t.to
		if t.value == 0 {
			t.positionReady = false
		}
		return t.value
	}
	duration := enter
	if t.to == 0 {
		duration = exit
	}
	if len(motions) > 0 {
		duration = theme.ResolveMotionDuration(motions[0], duration)
	}
	if duration <= 0 {
		t.value = t.to
		t.from = t.to
		if t.value == 0 {
			t.positionReady = false
		}
		return t.value
	}
	progress := animation.EaseSmoothstep(animation.Progress(gtx.Now.Sub(t.at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	t.value = animation.LerpFloat(t.from, t.to, progress)
	if t.value == 0 {
		t.positionReady = false
	}
	return t.value
}

// Position follows a pointer target with ECharts-style ease-out motion. The
// first position after a completed exit is shown immediately.
func (t *PopupTransition) Position(gtx layout.Context, target f32.Point, motions ...theme.MotionTheme) f32.Point {
	return t.PositionWithDuration(gtx, target, 0, motions...)
}

// PositionWithContext resolves tooltip motion metrics from the active visual
// provider and reuses them for this transition's lifetime.
func (t *PopupTransition) PositionWithContext(ctx *core.Context, gtx layout.Context, target f32.Point) f32.Point {
	visual := tooltipVisualWithCache(ctx, nil, &t.visualCache)
	return t.PositionWithDuration(gtx, target, visual.Metrics.MoveDuration, visual.Motion)
}

func (t *PopupTransition) PositionWithDuration(gtx layout.Context, target f32.Point, duration time.Duration, motions ...theme.MotionTheme) f32.Point {
	if !t.positionReady {
		t.position = target
		t.positionFrom = target
		t.positionTo = target
		t.positionAt = gtx.Now
		t.positionReady = true
		return target
	}
	if len(motions) > 0 {
		duration = theme.ResolveMotionDuration(motions[0], duration)
	}
	t.advancePosition(gtx.Now, duration)
	if target != t.positionTo {
		t.positionFrom = t.position
		t.positionTo = target
		t.positionAt = gtx.Now
	}
	if t.advancePosition(gtx.Now, duration) {
		gtx.Execute(op.InvalidateCmd{})
	}
	return t.position
}

func (t *PopupTransition) advancePosition(now time.Time, duration time.Duration) bool {
	if t.positionFrom == t.positionTo {
		t.position = t.positionTo
		return false
	}
	if duration <= 0 {
		t.position = t.positionTo
		t.positionFrom = t.positionTo
		return false
	}
	progress := animation.Progress(now.Sub(t.positionAt), duration)
	remaining := 1 - progress
	progress = 1 - remaining*remaining*remaining
	t.position = f32.Pt(
		animation.LerpFloat(t.positionFrom.X, t.positionTo.X, progress),
		animation.LerpFloat(t.positionFrom.Y, t.positionTo.Y, progress),
	)
	if progress >= 1 {
		t.positionFrom = t.positionTo
		return false
	}
	return true
}

func (t *PopupTransition) Value() float32 {
	return t.value
}

func (t *PopupTransition) Exiting() bool {
	return t.to == 0 && t.from > 0
}

func (t *PopupTransition) Reset() {
	*t = PopupTransition{}
}

// Style merges instance declarations into the popup.
func (p Popup) Style(values ...style.Style) Popup {
	declaration := style.Join(values...)
	if p.override == nil {
		p.override = &theme.TooltipOverride{}
	}
	p.override.Declaration = style.Join(p.override.Declaration, declaration)
	return p
}
