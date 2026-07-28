package menu

import (
	"image"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/components/disclosure"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotContextMenu = "context-menu"
const contextMenuLongPressDelay = 500 * time.Millisecond

type contextMenuState struct {
	key          string
	trigger      contextMenuTrigger
	dismiss      [16]overlay.ClickArea
	dialog       overlay.ClickArea
	open         bool // cached effective open, updated by isOpen/requestOpen
	wasOpen      bool
	anchor       image.Rectangle
	hasAnchor    bool
	triggerSize  image.Point
	transition   animation.FloatTransition
	skipRestore  bool
	focusVisible bool
	focusTarget  event.Tag
	disclosure   disclosure.Binding[bool]
}

type contextMenuTrigger struct {
	touchTracking bool
	touchID       pointer.ID
	touchStart    f32.Point
	touchAt       time.Time
}

func contextMenuStateFor(ctx *frame.Context, key string) *contextMenuState {
	key = frame.ClaimKey(ctx, state.KindContextMenu, key)
	value := frame.UseStateWith(ctx, key, stateSlotContextMenu, func() *contextMenuState {
		return &contextMenuState{key: key}
	})
	frame.RegisterExclusive(ctx, "context-menu", key, value.closeForPeer)
	return value
}

func activateContextMenu(ctx *frame.Context, value *contextMenuState) {
	if value != nil && value.key != "" {
		frame.ActivateExclusive(ctx, "context-menu", value.key)
	}
}

func releaseContextMenu(ctx *frame.Context, value *contextMenuState) {
	if value != nil {
		frame.ReleaseExclusive(ctx, "context-menu", value.key)
	}
}

// contextMenuDisclosureCfg builds a disclosure.Config from the widget's open-state fields.
func contextMenuDisclosureCfg(widget ContextMenuWidget) disclosure.Config[bool] {
	return disclosure.Config[bool]{
		Controlled: widget.hasOpen,
		Value:      widget.open,
		HasDefault: widget.hasDefaultOpen,
		Default:    widget.defaultOpen,
		OnChange:   widget.onOpenChange,
	}
}

func (s *contextMenuState) bind(widget ContextMenuWidget) {
	s.disclosure.Bind(contextMenuDisclosureCfg(widget))
}

func (s *contextMenuState) isOpen(widget ContextMenuWidget) bool {
	s.open = s.disclosure.Current(contextMenuDisclosureCfg(widget))
	return s.open
}

func (s *contextMenuState) requestOpen(ctx *frame.Context, widget ContextMenuWidget, open bool) bool {
	if widget.disabled {
		open = false
	}
	if open {
		s.skipRestore = false
		activateContextMenu(ctx, s)
	}
	s.open, _ = s.disclosure.Request(contextMenuDisclosureCfg(widget), open)
	if !s.open {
		releaseContextMenu(ctx, s)
	}
	return s.open
}

func (s *contextMenuState) closeForPeer() {
	s.skipRestore = true
	if s.disclosure.PeerClose(false) {
		s.open = false
	}
}

func (s *contextMenuState) observeOpen(open bool) {
	if !open && s.wasOpen {
		s.trigger.touchTracking = false
	}
	s.wasOpen = open
}

func (s *contextMenuState) progress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	duration := contextMenuExitDuration
	if open {
		target = 1
		duration = contextMenuEnterDuration
	}
	s.transition.Initialize(0, gtx.Now)
	return s.transition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func contextMenuPointRect(point f32.Point, size int) image.Rectangle {
	x := int(point.X + 0.5)
	y := int(point.Y + 0.5)
	size = max(size, 1)
	return image.Rect(x, y, x+size, y+size)
}

func contextMenuMovedBeyond(start, current f32.Point, threshold float32) bool {
	dx := current.X - start.X
	if dx < 0 {
		dx = -dx
	}
	dy := current.Y - start.Y
	if dy < 0 {
		dy = -dy
	}
	return dx > threshold || dy > threshold
}

func contextMenuTriggerFilters(pointerTarget, focusTarget event.Tag, additional []event.Tag) []event.Filter {
	filters := []event.Filter{
		pointer.Filter{Target: pointerTarget, Kinds: pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Cancel},
		key.FocusFilter{Target: focusTarget},
		key.Filter{Focus: focusTarget, Name: key.NameF10, Required: key.ModShift},
	}
	for _, target := range additional {
		if target == nil {
			continue
		}
		filters = append(filters,
			key.Filter{Focus: target, Name: key.NameF10, Required: key.ModShift},
		)
	}
	return filters
}
