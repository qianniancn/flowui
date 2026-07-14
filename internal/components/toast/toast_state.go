package toast

import (
	"fmt"
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotToast = "toast"

func toastStateFor(ctx *frame.Context, key string) (string, *toastProviderState) {
	key = frame.ClaimKey(ctx, state.KindToast, key)
	return key, frame.UseState[toastProviderState](ctx, key, stateSlotToast)
}

type toastProviderState struct {
	entries        map[string]*toastEntryState
	order          []string
	region         toastRegionTag
	regionHovered  bool
	touchMode      bool
	expansionValue float32
	expansionFrom  float32
	expansionTo    float32
	expansionAt    time.Time
	expansionReady bool
}

type toastEntryState struct {
	item              ToastItem
	present           bool
	closeRequested    bool
	root              toastRootTag
	hovered           bool
	close             widget.Clickable
	action            widget.Clickable
	rootFocus         state.FocusAnimation
	closeFocus        state.FocusAnimation
	remaining         time.Duration
	configuredTimeout time.Duration
	deadline          time.Time
	timerRunning      bool
	value             float32
	from              float32
	to                float32
	at                time.Time
	ready             bool
	stackValue        float32
	stackFrom         float32
	stackTo           float32
	stackAt           time.Time
	stackReady        bool
}

type toastRootTag struct{ _ byte }
type toastRegionTag struct{ _ byte }

func (s *toastProviderState) sync(gtx layout.Context, items []ToastItem, defaultTimeout time.Duration) {
	if s.entries == nil {
		s.entries = make(map[string]*toastEntryState)
	}
	seen := make(map[string]struct{}, len(items))
	nextOrder := make([]string, 0, len(items)+len(s.order))
	for _, item := range items {
		if item.key == "" {
			panic("flowui: empty toast key")
		}
		if _, duplicate := seen[item.key]; duplicate {
			panic(fmt.Sprintf("flowui: duplicate toast key %q", item.key))
		}
		seen[item.key] = struct{}{}
		nextOrder = append(nextOrder, item.key)
		entry := s.entries[item.key]
		timeout := defaultTimeout
		if item.hasTimeout {
			timeout = item.timeout
		}
		timeout = max(timeout, 0)
		if entry == nil {
			entry = &toastEntryState{
				item:              item,
				present:           true,
				remaining:         timeout,
				configuredTimeout: timeout,
			}
			if timeout > 0 {
				entry.deadline = gtx.Now.Add(timeout)
				entry.timerRunning = true
			}
			s.entries[item.key] = entry
			continue
		}
		restartTimer := !entry.present || entry.configuredTimeout != timeout || entry.item.loading != item.loading
		entry.item = item
		entry.present = true
		if restartTimer && !entry.closeRequested {
			entry.configuredTimeout = timeout
			entry.remaining = timeout
			entry.timerRunning = timeout > 0
			if entry.timerRunning {
				entry.deadline = gtx.Now.Add(timeout)
			}
		}
	}

	for _, key := range s.order {
		if _, current := seen[key]; current {
			continue
		}
		entry := s.entries[key]
		if entry == nil {
			continue
		}
		entry.present = false
		entry.timerRunning = false
		nextOrder = append(nextOrder, key)
	}
	s.order = nextOrder
}

func (s *toastProviderState) visible() bool {
	for _, entry := range s.entries {
		if entry.value > 0 || (entry.present && !entry.closeRequested) {
			return true
		}
	}
	return false
}

func (s *toastProviderState) resetRegion() {
	s.regionHovered = false
	s.expansionValue = 0
	s.expansionFrom = 0
	s.expansionTo = 0
	s.expansionReady = false
}

func (s *toastProviderState) cleanup() {
	if len(s.entries) == 0 {
		return
	}
	nextOrder := s.order[:0]
	for _, key := range s.order {
		entry := s.entries[key]
		if entry == nil {
			continue
		}
		if !entry.present && entry.value <= 0 {
			delete(s.entries, key)
			continue
		}
		nextOrder = append(nextOrder, key)
	}
	s.order = nextOrder
}

func (s *toastProviderState) entry(key string) *toastEntryState {
	return s.entries[key]
}

func (s *toastProviderState) paused(gtx layout.Context) bool {
	for _, key := range s.order {
		entry := s.entries[key]
		if entry != nil && (entry.hovered || entry.close.Hovered() || gtx.Focused(&entry.root) || gtx.Focused(&entry.close) || gtx.Focused(&entry.action)) {
			return true
		}
	}
	return false
}

func (s *toastProviderState) updateRegionEvents(gtx layout.Context) {
	for {
		eventValue, ok := gtx.Event(pointer.Filter{
			Target: &s.region,
			Kinds:  pointer.Enter | pointer.Leave | pointer.Press | pointer.Cancel,
		})
		if !ok {
			return
		}
		if pointerEvent, ok := eventValue.(pointer.Event); ok {
			switch pointerEvent.Kind {
			case pointer.Enter:
				s.regionHovered = true
			case pointer.Leave, pointer.Cancel:
				s.regionHovered = false
			case pointer.Press:
				if pointerEvent.Source == pointer.Touch {
					s.touchMode = true
				}
			}
		}
	}
}

func (s *toastProviderState) addRegionInput(gtx layout.Context, bounds image.Rectangle) {
	if bounds.Empty() {
		return
	}
	area := clip.Rect(bounds).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &s.region)
	pass.Pop()
	area.Pop()
}

func (s *toastProviderState) expansionProgress(gtx layout.Context, expanded bool, duration time.Duration) float32 {
	target := float32(0)
	if expanded {
		target = 1
	}
	if !s.expansionReady {
		s.expansionValue = target
		s.expansionFrom = target
		s.expansionTo = target
		s.expansionAt = gtx.Now
		s.expansionReady = true
		return target
	}
	if target != s.expansionTo {
		s.expansionFrom = s.expansionValue
		s.expansionTo = target
		s.expansionAt = gtx.Now
	}
	if s.expansionFrom == s.expansionTo {
		s.expansionValue = s.expansionTo
		return s.expansionValue
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.expansionAt), max(duration, time.Millisecond)))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.expansionValue = render.Lerp(s.expansionFrom, s.expansionTo, progress)
	return s.expansionValue
}

func (e *toastEntryState) updateRootEvents(gtx layout.Context) {
	for {
		eventValue, ok := gtx.Event(pointer.Filter{
			Target: &e.root,
			Kinds:  pointer.Enter | pointer.Leave | pointer.Cancel,
		})
		if !ok {
			break
		}
		if pointerEvent, ok := eventValue.(pointer.Event); ok {
			switch pointerEvent.Kind {
			case pointer.Enter:
				e.hovered = true
			case pointer.Leave, pointer.Cancel:
				e.hovered = false
			}
		}
	}
}

func (e *toastEntryState) addRootInput(gtx layout.Context, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &e.root)
	pass.Pop()
	area.Pop()
}

func (s *toastProviderState) updateTimers(gtx layout.Context, paused bool, close func(*toastEntryState)) {
	for _, key := range s.order {
		entry := s.entries[key]
		if entry == nil || !entry.present || entry.closeRequested || entry.remaining <= 0 {
			continue
		}
		if paused {
			entry.pauseTimer(gtx.Now)
			continue
		}
		entry.resumeTimer(gtx.Now)
		if !gtx.Now.Before(entry.deadline) {
			close(entry)
			continue
		}
		gtx.Execute(op.InvalidateCmd{At: entry.deadline})
	}
}

func (e *toastEntryState) pauseTimer(now time.Time) {
	if !e.timerRunning {
		return
	}
	e.remaining = max(e.deadline.Sub(now), 0)
	e.timerRunning = false
}

func (e *toastEntryState) resumeTimer(now time.Time) {
	if e.timerRunning || e.remaining <= 0 {
		return
	}
	e.deadline = now.Add(e.remaining)
	e.timerRunning = true
}

func (e *toastEntryState) requestClose(providerClose func(string)) {
	if e.closeRequested {
		return
	}
	e.closeRequested = true
	e.timerRunning = false
	if providerClose != nil {
		providerClose(e.item.key)
	}
}

func (e *toastEntryState) progress(gtx layout.Context, duration time.Duration) float32 {
	target := float32(0)
	if e.present && !e.closeRequested {
		target = 1
	}
	if !e.ready {
		e.value = 0
		e.from = 0
		e.to = 0
		e.at = gtx.Now
		e.ready = true
	}
	if target != e.to {
		e.from = e.value
		e.to = target
		e.at = gtx.Now
	}
	if e.from == e.to {
		e.value = e.to
		return e.value
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(e.at), max(duration, time.Millisecond)))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	e.value = render.Lerp(e.from, e.to, progress)
	return e.value
}

func (e *toastEntryState) stackPosition(gtx layout.Context, target float32, duration time.Duration) float32 {
	if !e.stackReady {
		e.stackValue = target
		e.stackFrom = target
		e.stackTo = target
		e.stackAt = gtx.Now
		e.stackReady = true
		return target
	}
	if target != e.stackTo {
		e.stackFrom = e.stackValue
		e.stackTo = target
		e.stackAt = gtx.Now
	}
	if e.stackFrom == e.stackTo {
		e.stackValue = e.stackTo
		return e.stackValue
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(e.stackAt), max(duration, time.Millisecond)))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	e.stackValue = render.Lerp(e.stackFrom, e.stackTo, progress)
	return e.stackValue
}
