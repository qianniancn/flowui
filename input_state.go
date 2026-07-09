package flowui

import (
	"image"
	"image/color"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

type inputState struct {
	hovered     bool
	bg          color.NRGBA
	bgFrom      color.NRGBA
	bgTo        color.NRGBA
	bgAt        time.Time
	bgReady     bool
	border      color.NRGBA
	borderFrom  color.NRGBA
	borderTo    color.NRGBA
	borderAt    time.Time
	borderReady bool
}

func (s *inputState) update(ctx *Context, gtx layout.Context, disabled bool, tag event.Tag) {
	if disabled {
		s.hovered = false
		return
	}
	s.updatePointer(ctx, gtx, tag)
}

func (s *inputState) updatePointer(ctx *Context, gtx layout.Context, tag event.Tag) {
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
			s.hovered = true
		case pointer.Leave, pointer.Cancel:
			s.hovered = false
		case pointer.Press:
			ctx.requestFocus(tag)
		}
	}
}

func (s *inputState) addPointer(gtx layout.Context, size image.Point, disabled bool) {
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

func (s *inputState) background(gtx layout.Context, target color.NRGBA) color.NRGBA {
	return inputColor(gtx, target, &s.bg, &s.bgFrom, &s.bgTo, &s.bgAt, &s.bgReady)
}

func (s *inputState) borderColor(gtx layout.Context, target color.NRGBA) color.NRGBA {
	return inputColor(gtx, target, &s.border, &s.borderFrom, &s.borderTo, &s.borderAt, &s.borderReady)
}

func inputColor(gtx layout.Context, target color.NRGBA, value, from, to *color.NRGBA, at *time.Time, ready *bool) color.NRGBA {
	if !*ready {
		*value = target
		*from = target
		*to = target
		*at = gtx.Now
		*ready = true
		return target
	}
	if target != *to {
		*from = *value
		*to = target
		*at = gtx.Now
	}
	if *from == *to {
		*value = *to
		return *value
	}
	progress := animationEase(animationProgress(gtx.Now.Sub(*at), inputColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	*value = lerpColor(*from, *to, progress)
	return *value
}
