package radiogroup

import (
	"fmt"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/components/optionrow"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotRadioGroup = "radio-group"

func radioGroupStateFor(ctx *frame.Context, key string) *radioGroupState {
	key = frame.ClaimKey(ctx, state.KindRadioGroup, key)
	return frame.UseState[radioGroupState](ctx, key, stateSlotRadioGroup)
}

type radioGroupState struct {
	items      map[string]*radioItemState
	frameItems map[string]struct{}
	itemKeys   map[string]struct{}
	keyFilters []event.Filter
}

func (s *radioGroupState) beginFrame() {
	state.BeginFrameMap(&s.frameItems)
}

func (s *radioGroupState) endFrame() {
	state.SweepFrameMap(s.items, s.frameItems)
}

func (s *radioGroupState) item(key string) *radioItemState {
	return state.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *radioGroupState) checkItems(items []RadioItem) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty radio item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate radio item key %q", item.Key))
		}
		s.itemKeys[item.Key] = struct{}{}
	}
}

func (s *radioGroupState) updateKeys(gtx layout.Context, items []RadioItem, selectedKey string) (string, bool) {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range items {
		if item.Disabled {
			continue
		}
		tag := &s.item(item.Key).clickable
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameRightArrow},
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameLeftArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
		)
	}
	if len(s.keyFilters) == 0 {
		return "", false
	}

	target := selectedKey
	current := s.focusedIndex(gtx, items)
	if current < 0 {
		current = radioIndexByKey(items, target)
	}
	changed := false
	for {
		e, ok := gtx.Event(s.keyFilters...)
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		switch event.Name {
		case key.NameRightArrow, key.NameDownArrow:
			if next, ok := radioMoveIndex(items, current, 1); ok {
				current = next
				target = items[next].Key
				changed = true
			}
		case key.NameLeftArrow, key.NameUpArrow:
			if next, ok := radioMoveIndex(items, current, -1); ok {
				current = next
				target = items[next].Key
				changed = true
			}
		case key.NameHome:
			if next, ok := radioFirstEnabled(items); ok {
				current = next
				target = items[next].Key
				changed = true
			}
		case key.NameEnd:
			if next, ok := radioLastEnabled(items); ok {
				current = next
				target = items[next].Key
				changed = true
			}
		}
	}
	return target, changed
}

func (s *radioGroupState) focusedIndex(gtx layout.Context, items []RadioItem) int {
	for i, item := range items {
		if item.Disabled {
			continue
		}
		if gtx.Focused(&s.item(item.Key).clickable) {
			return i
		}
	}
	return -1
}

func radioIndexByKey(items []RadioItem, key string) int {
	for i, item := range items {
		if item.Key == key {
			return i
		}
	}
	return -1
}

func radioMoveIndex(items []RadioItem, current, delta int) (int, bool) {
	count := len(items)
	if count == 0 {
		return -1, false
	}
	if current < 0 || current >= count {
		if delta < 0 {
			return radioLastEnabled(items)
		}
		return radioFirstEnabled(items)
	}
	next := current
	for range count {
		next = (next + delta + count) % count
		if !items[next].Disabled {
			return next, true
		}
	}
	return -1, false
}

func radioFirstEnabled(items []RadioItem) (int, bool) {
	for i, item := range items {
		if !item.Disabled {
			return i, true
		}
	}
	return -1, false
}

func radioLastEnabled(items []RadioItem) (int, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		if !items[i].Disabled {
			return i, true
		}
	}
	return -1, false
}

type radioItemState struct {
	clickable widget.Clickable
	selected  animation.FloatTransition
	focus     state.FocusAnimation
}

func (s *radioItemState) selection(gtx layout.Context, selected bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	return s.selected.Value(gtx, target, radioSelectDuration, animation.EaseSmoothstep, motions...)
}

func (s *radioItemState) focusOpacity(gtx layout.Context, focused bool, motions ...theme.MotionTheme) float32 {
	return s.focus.Opacity(gtx, focused, motions...)
}

func radioPressScale(gtx layout.Context, history []widget.Press, activeTheme *theme.Theme, disabled bool) float32 {
	target := activeTheme.Components.RadioGroup.PressedScale
	if target <= 0 || target > 1 {
		target = 0.95
	}
	return optionrow.PressScale(gtx, history, disabled, target, radioPressInDuration, radioPressOutDuration, activeTheme.Motion)
}
