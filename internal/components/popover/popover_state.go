package popover

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
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
	disclosure disclosure.Binding[bool]
	open       bool // cached effective open, updated by isOpen/requestOpen
}

// popoverDisclosureCfg builds a disclosure.Config from the widget's open-state fields.
func popoverDisclosureCfg(widget PopoverWidget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasOpen,
		Value:      widget.open,
		HasDefault: widget.hasDefaultOpen,
		Default:    widget.defaultOpen,
		OnChange:   widget.onOpenChange,
	}
}

func (s *popoverState) isOpen(widget PopoverWidget) bool {
	s.open = s.disclosure.Current(popoverDisclosureCfg(widget))
	return s.open
}

func (s *popoverState) bind(widget PopoverWidget) {
	s.disclosure.Bind(popoverDisclosureCfg(widget))
}

func (s *popoverState) requestOpen(widget PopoverWidget, open bool) bool {
	s.open, _ = s.disclosure.Request(popoverDisclosureCfg(widget), open)
	return s.open
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
