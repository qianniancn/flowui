// Package optionrow provides reusable state, animation, and drawing helpers
// for selectable rows. It is a building block for components such as ListBox
// and ComboBox rather than a standalone control.
package optionrow

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/theme"
)

type State struct {
	Clickable  widget.Clickable
	background animation.ColorTransition
	selection  animation.FloatTransition
}

type FocusableState struct {
	State
	focus animation.FocusAnimation
}

func (s *State) Background(gtx layout.Context, target color.NRGBA, duration time.Duration, motions ...theme.MotionTheme) color.NRGBA {
	return s.background.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *State) Selection(gtx layout.Context, selected bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	return s.selection.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *FocusableState) FocusOpacity(gtx layout.Context, focused bool, motions ...theme.MotionTheme) float32 {
	return s.focus.Opacity(gtx, focused, motions...)
}

func (s *FocusableState) FocusTargetOpacity() float32 {
	return s.focus.TargetOpacity()
}

func PressScale(gtx layout.Context, history []widget.Press, disabled bool, target float32, pressIn, pressOut time.Duration, motions ...theme.MotionTheme) float32 {
	return PressScaleWithEasing(gtx, history, disabled, target, pressIn, pressOut, animation.EaseSmoothstep, motions...)
}

// PressScaleWithEasing is the provider-aware variant used by components whose
// theme owns the transition curve. PressScale keeps the historical smoothstep
// behavior for callers that do not supply a visual easing policy.
func PressScaleWithEasing(gtx layout.Context, history []widget.Press, disabled bool, target float32, pressIn, pressOut time.Duration, easing animation.Easing, motions ...theme.MotionTheme) float32 {
	if disabled || len(history) == 0 {
		return 1
	}
	if len(motions) > 0 {
		pressIn = theme.ResolveMotionDuration(motions[0], pressIn)
		pressOut = theme.ResolveMotionDuration(motions[0], pressOut)
	}
	if easing == nil {
		easing = animation.EaseCubicOut
	}
	press := history[len(history)-1]
	if press.End.IsZero() {
		if pressIn <= 0 {
			return target
		}
		progress := easing(animation.Progress(gtx.Now.Sub(press.Start), pressIn))
		if progress < 1 {
			gtx.Execute(op.InvalidateCmd{})
		}
		return animation.LerpFloat(1, target, progress)
	}
	if pressOut <= 0 {
		return 1
	}
	progress := easing(animation.Progress(gtx.Now.Sub(press.End), pressOut))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	return animation.LerpFloat(target, 1, progress)
}

func Frame(constraints layout.Constraints, minHeight int, content image.Point) (size, offset image.Point) {
	size = constraints.Constrain(image.Pt(content.X, max(minHeight, content.Y)))
	offset.Y = max((size.Y-content.Y)/2, 0)
	return size, offset
}

func DrawBackground(gtx layout.Context, size image.Point, radius unit.Dp, background color.NRGBA) {
	if background.A == 0 {
		return
	}
	rect := image.Rectangle{Max: size}
	radiusPx := min(max(gtx.Dp(radius), 1), min(size.X, size.Y)/2)
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, radiusPx).Op(gtx.Ops))
}

func DrawCheck(gtx layout.Context, size image.Point, progress float32, inset, strokeWidth unit.Dp, foreground color.NRGBA) {
	progress = min(max(progress, 0), 1)
	if progress == 0 {
		return
	}
	rect := image.Rectangle{Max: size}.Inset(max(gtx.Dp(inset), 0))
	if rect.Empty() {
		return
	}
	x := float32(rect.Min.X)
	y := float32(rect.Min.Y)
	width := float32(rect.Dx())
	height := float32(rect.Dy())
	points := [3]f32.Point{
		f32.Pt(x+width*0.05, y+height*0.56),
		f32.Pt(x+width*0.40, y+height*0.86),
		f32.Pt(x+width*0.95, y+height*0.14),
	}
	path := draw.CheckPath(gtx.Ops, points, progress)
	foreground.A = byte(float32(foreground.A)*progress + 0.5)
	stroke := clip.Stroke{
		Path:  path,
		Width: max(draw.DpFloat(gtx, strokeWidth), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, foreground)
	stroke.Pop()
}
