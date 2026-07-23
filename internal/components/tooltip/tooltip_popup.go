package tooltip

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/render"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// Popup is the controlled presentation layer shared by triggered and
// data-driven tooltips.
type Popup struct {
	content            frame.Widget
	placement          overlay.PopoverPlacement
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
	customStyle        flowstyle.Style
	styleKey           string
	styleState         flowstyle.StyleState
}

func NewPopup(content frame.Widget) Popup {
	return Popup{
		content:   content,
		placement: overlay.PopoverTop,
		progress:  1,
	}
}

func (p Popup) Placement(placement overlay.PopoverPlacement) Popup {
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

func (p Popup) Style(value flowstyle.Style) Popup {
	p.customStyle = value
	return p
}

func (p Popup) Layout(ctx *frame.Context, gtx layout.Context, anchor image.Rectangle) layout.Dimensions {
	if anchor.Empty() || p.content == nil {
		return layout.Dimensions{}
	}
	bounds := gtx.Constraints.Max
	if bounds.X <= 0 || bounds.Y <= 0 {
		return layout.Dimensions{}
	}

	panelGtx := gtx.Disabled()
	panelGtx.Constraints = p.panelConstraints(ctx, gtx, bounds)
	panelCall, panelDims, panelPlacement := p.recordPanel(ctx, panelGtx)
	result := p.resolvedPosition(ctx, gtx, anchor, panelDims.Size, bounds)
	placement := result.Placement.PopoverPlacement()
	panelPos := result.Position

	panelAffine := p.panelAffine(ctx, anchor, panelPos, panelDims.Size, placement)
	panelOffset := panelPos.Add(p.slideOffset(ctx, gtx, placement))
	panelTransform := panelAffine.Mul(f32.AffineId().Offset(f32.Pt(float32(panelOffset.X), float32(panelOffset.Y))))
	panelPlacement.PlaceTransform(panelTransform)
	panelPlacement.SetOpacity(p.progress)

	transform := op.Affine(panelAffine).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, p.progress)
	offset := op.Offset(panelOffset).Push(gtx.Ops)
	panelCall.Add(gtx.Ops)
	if p.arrow {
		activeTheme := frame.ActiveTheme(ctx)
		tokens := activeTheme.Components.Tooltip
		arrowSize := gtx.Dp(tokens.ArrowSize)
		panelRadius := min(max(gtx.Dp(tokens.Radius), 0), min(panelDims.Size.X, panelDims.Size.Y)/2)
		arrowAnchor := tooltipArrowAnchor(anchor, panelPos, panelDims.Size, placement, panelRadius, arrowSize)
		drawTooltipArrow(gtx, placement, panelDims.Size, arrowAnchor, arrowSize, gtx.Dp(tokens.BorderWidth), tooltipStyleFor(activeTheme))
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

func (t TooltipWidget) popup(progress float32, exiting bool) Popup {
	p := NewPopup(t.content).
		Placement(t.placement).
		Progress(progress).
		Exiting(exiting).
		Arrow(t.arrow).
		Style(t.customStyle)
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
}

func (t *PopupTransition) Progress(gtx layout.Context, visible bool, motions ...theme.MotionTheme) float32 {
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
	duration := tooltipEnterDuration
	if t.to == 0 {
		duration = tooltipExitDuration
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
	progress := render.Ease(render.Progress(gtx.Now.Sub(t.at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	t.value = render.Lerp(t.from, t.to, progress)
	if t.value == 0 {
		t.positionReady = false
	}
	return t.value
}

// Position follows a pointer target with ECharts-style ease-out motion. The
// first position after a completed exit is shown immediately.
func (t *PopupTransition) Position(gtx layout.Context, target f32.Point, motions ...theme.MotionTheme) f32.Point {
	if !t.positionReady {
		t.position = target
		t.positionFrom = target
		t.positionTo = target
		t.positionAt = gtx.Now
		t.positionReady = true
		return target
	}
	duration := tooltipMoveDuration
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
	progress := render.Progress(now.Sub(t.positionAt), duration)
	remaining := 1 - progress
	progress = 1 - remaining*remaining*remaining
	t.position = f32.Pt(
		render.Lerp(t.positionFrom.X, t.positionTo.X, progress),
		render.Lerp(t.positionFrom.Y, t.positionTo.Y, progress),
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
