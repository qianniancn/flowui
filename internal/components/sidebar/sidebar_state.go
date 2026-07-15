package sidebar

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotSidebar = "sidebar"

type sidebarState struct {
	list             layout.List
	bar              widget.Scrollbar
	items            map[string]*sidebarItemState
	frameItems       map[string]struct{}
	keyFilters       []event.Filter
	pressedKey       key.Name
	pressedActionKey string
}

type sidebarItemState struct {
	clickable widget.Clickable
	focus     stateutil.FocusAnimation
}

func sidebarStateFor(ctx *frame.Context, key string) *sidebarState {
	key = frame.ClaimKey(ctx, stateutil.KindSidebar, key)
	return frame.UseState[sidebarState](ctx, key, stateSlotSidebar)
}

func (s *sidebarState) beginFrame() {
	stateutil.BeginFrameMap(&s.frameItems)
}

func (s *sidebarState) endFrame() {
	stateutil.SweepFrameMap(s.items, s.frameItems)
}

func (s *sidebarState) item(key string) *sidebarItemState {
	return stateutil.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *sidebarState) checkItems(items []Item) {
	validateSidebarItems(items)
}

type sidebarKeyResult struct {
	focusKey  string
	actionKey string
}

func (s *sidebarState) updateKeys(gtx layout.Context, sidebar Widget, items []Item) sidebarKeyResult {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range items {
		tag := &s.item(item.Key).clickable
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
		)
		if sidebar.itemDisabled(item) {
			continue
		}
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameEnter},
			key.Filter{Focus: tag, Name: key.NameReturn},
			key.Filter{Focus: tag, Name: key.NameSpace},
		)
	}
	if len(s.keyFilters) == 0 {
		return sidebarKeyResult{}
	}

	current := s.focusedIndex(gtx, items)
	if current < 0 {
		current = sidebarItemIndex(items, sidebar.selectedKey)
	}
	result := sidebarKeyResult{}
	for {
		value, ok := gtx.Event(s.keyFilters...)
		if !ok {
			break
		}
		event, ok := value.(key.Event)
		if !ok {
			continue
		}
		switch event.Name {
		case key.NameDownArrow:
			if event.State == key.Press {
				if next, ok := sidebarMoveIndex(items, sidebar, current, 1); ok {
					current = next
					result.focusKey = items[next].Key
				}
			}
		case key.NameUpArrow:
			if event.State == key.Press {
				if next, ok := sidebarMoveIndex(items, sidebar, current, -1); ok {
					current = next
					result.focusKey = items[next].Key
				}
			}
		case key.NameHome:
			if event.State == key.Press {
				if next, ok := sidebarFirstEnabled(items, sidebar); ok {
					current = next
					result.focusKey = items[next].Key
				}
			}
		case key.NameEnd:
			if event.State == key.Press {
				if next, ok := sidebarLastEnabled(items, sidebar); ok {
					current = next
					result.focusKey = items[next].Key
				}
			}
		case key.NameEnter, key.NameReturn, key.NameSpace:
			s.updateActionKey(event, items, sidebar, current, &result)
		}
	}
	return result
}

func (s *sidebarState) updateActionKey(event key.Event, items []Item, sidebar Widget, current int, result *sidebarKeyResult) {
	switch event.State {
	case key.Press:
		s.pressedKey = event.Name
		s.pressedActionKey = ""
		if current >= 0 && current < len(items) && !sidebar.itemDisabled(items[current]) {
			s.pressedActionKey = items[current].Key
		}
	case key.Release:
		if s.pressedKey == event.Name && s.pressedActionKey != "" {
			result.actionKey = s.pressedActionKey
		}
		s.pressedKey = ""
		s.pressedActionKey = ""
	}
}

func (s *sidebarState) focusedIndex(gtx layout.Context, items []Item) int {
	for index, item := range items {
		if gtx.Focused(&s.item(item.Key).clickable) {
			return index
		}
	}
	return -1
}

func (s *sidebarState) ensureVisible(index int) {
	if index < 0 || s.list.Position.Count == 0 {
		return
	}
	position := &s.list.Position
	if index < position.First || index >= position.First+position.Count {
		position.First = index
		position.Offset = 0
		return
	}
	if index == position.First && position.Offset > 0 {
		position.Offset = 0
	} else if index == position.First+position.Count-1 && position.OffsetLast < 0 {
		position.Offset -= position.OffsetLast
	}
}

func sidebarMoveIndex(items []Item, sidebar Widget, current, delta int) (int, bool) {
	if current < 0 || current >= len(items) {
		if delta < 0 {
			return sidebarLastEnabled(items, sidebar)
		}
		return sidebarFirstEnabled(items, sidebar)
	}
	for next := current + delta; next >= 0 && next < len(items); next += delta {
		if !sidebar.itemDisabled(items[next]) {
			return next, true
		}
	}
	return current, false
}

func sidebarFirstEnabled(items []Item, sidebar Widget) (int, bool) {
	for index, item := range items {
		if !sidebar.itemDisabled(item) {
			return index, true
		}
	}
	return -1, false
}

func sidebarLastEnabled(items []Item, sidebar Widget) (int, bool) {
	for index := len(items) - 1; index >= 0; index-- {
		if !sidebar.itemDisabled(items[index]) {
			return index, true
		}
	}
	return -1, false
}

func sidebarItemIndex(items []Item, itemKey string) int {
	for index, item := range items {
		if item.Key == itemKey {
			return index
		}
	}
	return -1
}
