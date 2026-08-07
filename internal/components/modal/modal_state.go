package modal

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
)

const stateSlotModal = "modal"

func modalStateFor(ctx *frame.Context, key string) *modalState {
	key = frame.ClaimKey(ctx, state.KindModal, key)
	return frame.UseState[modalState](ctx, key, stateSlotModal)
}

func hasVisibleModal(ctx *frame.Context, key string) bool {
	value, ok := frame.PeekState[modalState](ctx, key, stateSlotModal)
	return ok && value.visible()
}

func deleteModalState(ctx *frame.Context, key string) {
	frame.DeleteState(ctx, key, stateSlotModal)
}

type modalState struct {
	dismiss             [4]overlay.ClickArea
	dialog              overlay.ClickArea
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
	focusPending        bool
	disclosure          disclosure.Binding[bool]
	open                bool // cached effective open, updated by isOpen/requestOpen
}

// modalDisclosureCfg builds a disclosure.Config from the widget's open-state fields.
func modalDisclosureCfg(widget ModalWidget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasOpen,
		Value:      widget.open,
		HasDefault: widget.hasDefaultOpen,
		Default:    widget.defaultOpen,
		OnChange:   widget.onOpenChange,
	}
}

func (s *modalState) isOpen(widget ModalWidget) bool {
	s.open = s.disclosure.Current(modalDisclosureCfg(widget))
	return s.open
}

func (s *modalState) bind(widget ModalWidget) {
	s.disclosure.Bind(modalDisclosureCfg(widget))
}

func (s *modalState) requestOpen(widget ModalWidget, open bool) bool {
	s.open, _ = s.disclosure.Request(modalDisclosureCfg(widget), open)
	return s.open
}

// Keep the tag non-zero-sized so distinct focus boundaries have distinct
// addresses when used as Gio event tags.
type modalFocusTag struct{ _ byte }

func (s *modalState) progress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	s.transition.Initialize(0, gtx.Now)
	if target != s.transition.Target() {
		s.focusPending = open
	}
	return s.transition.Value(gtx, target, modalEnterDuration, animation.EaseSmoothstep, motions...)
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
