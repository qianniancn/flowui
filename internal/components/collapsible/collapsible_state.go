package collapsible

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotCollapsible = "collapsible"

type collapsibleItemState struct {
	clickable  widget.Clickable
	keyFilters stateutil.KeyFilterCache
}

type collapsibleState struct {
	item       collapsibleItemState
	items      map[string]*collapsibleItemState
	frameItems map[string]struct{}
	keyFilters []event.Filter
}

func collapsibleStateFor(ctx *frame.Context, key string) *collapsibleState {
	key = frame.ClaimKey(ctx, stateutil.KindCollapsible, key)
	return frame.UseState[collapsibleState](ctx, key, stateSlotCollapsible)
}

func (s *collapsibleState) beginFrame() {
	stateutil.BeginFrameMap(&s.frameItems)
}

func (s *collapsibleState) endFrame() {
	stateutil.SweepFrameMap(s.items, s.frameItems)
}

func (s *collapsibleState) itemFor(key string) *collapsibleItemState {
	return stateutil.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *collapsibleState) checkItems(items []Item) {
	validateItems(items)
	for _, item := range items {
		s.itemFor(item.Key)
	}
}

func (s *collapsibleState) updateKeys(gtx layout.Context, items []Item) string {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range items {
		if item.Disabled {
			continue
		}
		itemState := s.itemFor(item.Key)
		tag := &itemState.clickable
		s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag,
			key.NameDownArrow,
			key.NameUpArrow,
			key.NameHome,
			key.NameEnd,
		)...)
	}
	if len(s.keyFilters) == 0 {
		return ""
	}

	current := s.focusedIndex(gtx, items)
	target := -1
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
		case key.NameDownArrow:
			target = moveEnabledIndex(items, current, 1)
		case key.NameUpArrow:
			target = moveEnabledIndex(items, current, -1)
		case key.NameHome:
			target = firstEnabledIndex(items)
		case key.NameEnd:
			target = lastEnabledIndex(items)
		}
		if target >= 0 {
			current = target
		}
	}
	if target < 0 {
		return ""
	}
	return items[target].Key
}

func (s *collapsibleState) focusedIndex(gtx layout.Context, items []Item) int {
	for index, item := range items {
		if !item.Disabled && gtx.Focused(&s.itemFor(item.Key).clickable) {
			return index
		}
	}
	return -1
}

func moveEnabledIndex(items []Item, current, delta int) int {
	if len(items) == 0 {
		return -1
	}
	if current < 0 || current >= len(items) {
		if delta < 0 {
			return lastEnabledIndex(items)
		}
		return firstEnabledIndex(items)
	}
	index := current
	for range len(items) {
		index = (index + delta + len(items)) % len(items)
		if !items[index].Disabled {
			return index
		}
	}
	return -1
}

func firstEnabledIndex(items []Item) int {
	for index, item := range items {
		if !item.Disabled {
			return index
		}
	}
	return -1
}

func lastEnabledIndex(items []Item) int {
	for index := len(items) - 1; index >= 0; index-- {
		if !items[index].Disabled {
			return index
		}
	}
	return -1
}

func activePresses(state *collapsibleItemState) int {
	return stateutil.ActivePresses(state.clickable.History())
}

func focusOnPress(ctx *frame.Context, state *collapsibleItemState, before int) {
	frame.FocusOnPress(ctx, &state.clickable, state.clickable.History(), before)
}
