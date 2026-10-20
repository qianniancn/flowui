package modal

import (
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/closebutton"
	"github.com/qianniancn/flowui/widget/disclosure"
)

const stateSlotModal = "modal"

func modalStateFor(ctx *core.Context, key string) *modalState {
	key = core.ClaimKey(ctx, core.KindModal, key)
	return core.UseState[modalState](ctx, key, stateSlotModal)
}

func deleteModalState(ctx *core.Context, key string) {
	core.DeleteState(ctx, key, stateSlotModal)
}

type modalState struct {
	dismiss             [4]core.ClickArea
	dialog              core.ClickArea
	close               widget.Clickable
	bodyList            layout.List
	bodyBar             widget.Scrollbar
	bodyVisualOutset    layoutui.VisualOutsetState
	outsideList         layout.List
	outsideVisualOutset layoutui.VisualOutsetState
	focusStart          modalFocusTag
	focusTarget         modalFocusTag
	focusEnd            modalFocusTag
	transition          animation.FloatTransition
	visualCache         modalVisualCache
	closeButton         *closebutton.Widget
	focusPending        bool
	disclosure          disclosure.Binding[bool]
	open                bool // cached effective open, updated by isOpen/requestOpen
}

// modalDisclosureCfg builds a disclosure.Config from the widget's open-state fields.
func modalDisclosureCfg(widget Widget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasOpen,
		Value:      widget.open,
		HasDefault: widget.hasDefaultOpen,
		Default:    widget.defaultOpen,
		OnChange:   widget.onOpenChange,
	}
}

func (s *modalState) isOpen(widget Widget) bool {
	s.open = s.disclosure.Current(modalDisclosureCfg(widget))
	return s.open
}

func (s *modalState) bind(widget Widget) {
	s.disclosure.Bind(modalDisclosureCfg(widget))
}

func (s *modalState) requestOpen(widget Widget, open bool) bool {
	s.open, _ = s.disclosure.Request(modalDisclosureCfg(widget), open)
	return s.open
}

// Keep the tag non-zero-sized so distinct focus boundaries have distinct
// addresses when used as Gio event tags.
type modalFocusTag struct{ _ byte }

// progressFor is the production path: the provider visual owns modal motion
// timing.
func (s *modalState) progressFor(gtx layout.Context, open bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	s.transition.Initialize(0, gtx.Now)
	if target != s.transition.Target() {
		s.focusPending = open
	}
	easing := animation.EaseSmoothstep
	if len(motions) > 0 {
		motion := motions[0]
		if resolved := animation.EasingFor(motion.Ease); resolved != nil {
			easing = resolved
		}
	}
	return s.transition.Value(gtx, target, duration, easing, motions...)
}

func (s *modalState) visible() bool {
	return s.transition.Ready() && s.transition.Current() > 0
}

func (s *modalState) opening() bool {
	return s.transition.Increasing()
}

func (s *modalState) initialFocusTag() event.Tag {
	return &s.focusTarget
}

func (s *modalState) tabFocusTag(showClose bool) event.Tag {
	if showClose {
		return &s.close
	}
	return &s.focusTarget
}

func (s *modalState) endFocusTag() event.Tag {
	return &s.focusTarget
}

// closeButtonWidget retains the Provider cache across modal frames while
// keeping CloseButton's behavior and clickable state owned by modalState.
func (s *modalState) closeButtonWidget() *closebutton.Widget {
	if s.closeButton == nil {
		value := closebutton.New("close")
		s.closeButton = &value
	}
	return s.closeButton
}

func (s *modalState) escapePressed(gtx layout.Context) bool {
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
