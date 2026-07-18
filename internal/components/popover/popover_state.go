package popover

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
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
	dismiss    [16]overlay.ClickArea
	dialog     overlay.ClickArea
	arrow      overlay.ClickArea
	transition animation.FloatTransition
}

func (s *popoverState) progress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	duration := popoverEnterDuration
	if !open {
		duration = popoverExitDuration
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *popoverState) visible() bool {
	return s.transition.Ready() && s.transition.Current() > 0
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
