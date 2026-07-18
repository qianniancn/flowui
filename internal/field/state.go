package field

import (
	"image"
	"image/color"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const colorDuration = 150 * time.Millisecond

type State struct {
	Hovered bool
	bg      animation.ColorTransition
	border  animation.ColorTransition
}

func (s *State) Update(ctx *frame.Context, gtx layout.Context, disabled bool, tag event.Tag) {
	if disabled {
		s.Hovered = false
		return
	}
	s.updatePointer(ctx, gtx, tag)
}

func (s *State) updatePointer(ctx *frame.Context, gtx layout.Context, tag event.Tag) {
	for {
		e, ok := gtx.Event(pointer.Filter{
			Target: s,
			Kinds:  pointer.Enter | pointer.Leave | pointer.Cancel | pointer.Press,
		})
		if !ok {
			return
		}
		event, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		switch event.Kind {
		case pointer.Enter:
			s.Hovered = true
		case pointer.Leave, pointer.Cancel:
			s.Hovered = false
		case pointer.Press:
			frame.RequestFocus(ctx, tag)
		}
	}
}

func (s *State) AddPointer(gtx layout.Context, size image.Point, disabled bool) {
	if disabled {
		return
	}
	area := clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops)
	pointer.CursorText.Add(gtx.Ops)
	stack := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, s)
	stack.Pop()
	area.Pop()
}

func (s *State) Background(gtx layout.Context, target color.NRGBA, motions ...theme.MotionTheme) color.NRGBA {
	return s.bg.Value(gtx, target, colorDuration, animation.EaseSmoothstep, motions...)
}

func (s *State) BorderColor(gtx layout.Context, target color.NRGBA, motions ...theme.MotionTheme) color.NRGBA {
	return s.border.Value(gtx, target, colorDuration, animation.EaseSmoothstep, motions...)
}
