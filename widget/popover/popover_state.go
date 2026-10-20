package popover

import (
	"time"

	"gioui.org/io/key"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const stateSlotPopover = "popover"

func popoverStateFor(ctx *core.Context, key string) *popoverState {
	key = core.ClaimKey(ctx, core.KindPopover, key)
	return core.UseState[popoverState](ctx, key, stateSlotPopover)
}

func deletePopoverState(ctx *core.Context, key string) {
	core.DeleteState(ctx, key, stateSlotPopover)
}

type popoverState struct {
	dismiss     [16]core.ClickArea
	dialog      core.ClickArea
	arrow       core.ClickArea
	transition  animation.FloatTransition
	disclosure  disclosure.Binding[bool]
	visualCache popoverVisualCache
	open        bool // cached effective open, updated by isOpen/requestOpen
}

// popoverDisclosureCfg builds a disclosure.Config from the widget's open-state fields.
func popoverDisclosureCfg(widget Widget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasOpen,
		Value:      widget.open,
		HasDefault: widget.hasDefaultOpen,
		Default:    widget.defaultOpen,
		OnChange:   widget.onOpenChange,
	}
}

func (s *popoverState) isOpen(widget Widget) bool {
	s.open = s.disclosure.Current(popoverDisclosureCfg(widget))
	return s.open
}

func (s *popoverState) bind(widget Widget) {
	s.disclosure.Bind(popoverDisclosureCfg(widget))
}

func (s *popoverState) requestOpen(widget Widget, open bool) bool {
	s.open, _ = s.disclosure.Request(popoverDisclosureCfg(widget), open)
	return s.open
}

// progressFor is the production path: overlay timing comes from the active
// provider visual.
func (s *popoverState) progressFor(gtx layout.Context, open bool, enter, exit time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	duration := enter
	if !open {
		duration = exit
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
