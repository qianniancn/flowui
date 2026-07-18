package optionrow

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
)

type State struct {
	Clickable  widget.Clickable
	background animation.ColorTransition
	selection  animation.FloatTransition
}

type FocusableState struct {
	State
	focus stateutil.FocusAnimation
}

func (s *State) Background(gtx layout.Context, target color.NRGBA, duration time.Duration) color.NRGBA {
	return s.background.Value(gtx, target, duration, animation.EaseSmoothstep)
}

func (s *State) Selection(gtx layout.Context, selected bool, duration time.Duration) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	return s.selection.Value(gtx, target, duration, animation.EaseSmoothstep)
}

func (s *FocusableState) FocusOpacity(gtx layout.Context, focused bool) float32 {
	return s.focus.Opacity(gtx, focused)
}

func (s *FocusableState) FocusTargetOpacity() float32 {
	return s.focus.TargetOpacity()
}

func PressScale(gtx layout.Context, history []widget.Press, disabled bool, target float32, pressIn, pressOut time.Duration) float32 {
	if disabled || len(history) == 0 {
		return 1
	}
	press := history[len(history)-1]
	if press.End.IsZero() {
		progress := render.Ease(render.Progress(gtx.Now.Sub(press.Start), pressIn))
		if progress < 1 {
			gtx.Execute(op.InvalidateCmd{})
		}
		return render.Lerp(1, target, progress)
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(press.End), pressOut))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	return render.Lerp(target, 1, progress)
}

func Frame(constraints layout.Constraints, minHeight int, content image.Point) (size, offset image.Point) {
	size = constraints.Constrain(image.Pt(content.X, max(minHeight, content.Y)))
	offset.Y = max((size.Y-content.Y)/2, 0)
	return size, offset
}

func LayoutText(ctx *frame.Context, gtx layout.Context, label, description string, labelSize, descriptionSize float32, foreground, descriptionColor color.NRGBA) layout.Dimensions {
	if description == "" {
		return text.New(label).
			Size(labelSize).
			Weight(font.Medium).
			Color(foreground).
			Layout(ctx, gtx)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(label).
				Size(labelSize).
				Weight(font.Medium).
				Color(foreground).
				Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(description).
				Size(descriptionSize).
				Color(descriptionColor).
				Layout(ctx, gtx)
		}),
	)
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
	path := render.CheckPath(gtx.Ops, points, progress)
	foreground.A = byte(float32(foreground.A)*progress + 0.5)
	stroke := clip.Stroke{
		Path:  path,
		Width: max(render.DpFloat(gtx, strokeWidth), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, foreground)
	stroke.Pop()
}
