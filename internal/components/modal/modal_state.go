package modal

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/components/closebutton"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
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
	dismiss      [4]overlay.ClickArea
	dialog       overlay.ClickArea
	close        widget.Clickable
	closeButton  closebutton.State
	bodyList     layout.List
	bodyBar      widget.Scrollbar
	outsideList  layout.List
	focusStart   modalFocusTag
	focusTarget  modalFocusTag
	focusEnd     modalFocusTag
	transition   animation.FloatTransition
	focusPending bool
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
