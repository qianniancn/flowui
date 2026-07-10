package popover

import (
	"time"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotPopover = "popover"

func popoverStateFor(ctx *frame.Context, key string) *popoverState {
	key = frame.ClaimKey(ctx, state.KindPopover, key)
	return frame.UseState[popoverState](ctx, key, stateSlotPopover)
}

func hasVisiblePopover(ctx *frame.Context, key string) bool {
	value, ok := frame.PeekState[popoverState](ctx, key, stateSlotPopover)
	return ok && value.visible()
}

func deletePopoverState(ctx *frame.Context, key string) {
	frame.DeleteState(ctx, key, stateSlotPopover)
}

type popoverState struct {
	dismiss [4]widget.Clickable
	dialog  widget.Clickable
	value   float32
	from    float32
	to      float32
	at      time.Time
	ready   bool
}

func (s *popoverState) progress(gtx layout.Context, open bool) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	if !s.ready {
		s.value = 0
		s.from = 0
		s.to = 0
		s.at = gtx.Now
		s.ready = true
	}
	if target != s.to {
		s.from = s.value
		s.to = target
		s.at = gtx.Now
	}
	if s.from == s.to {
		s.value = s.to
		return s.value
	}
	duration := popoverEnterDuration
	if s.to == 0 {
		duration = popoverExitDuration
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.value = render.Lerp(s.from, s.to, progress)
	return s.value
}

func (s *popoverState) visible() bool {
	return s.ready && s.value > 0
}

func (s *popoverState) escapePressed(gtx layout.Context) bool {
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			return false
		}
		event, ok := e.(key.Event)
		if ok && event.State == key.Press {
			return true
		}
	}
}
