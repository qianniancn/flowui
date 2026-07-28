package sidebar

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
)

const stateSlotSidebar = "sidebar"

type sidebarState struct {
	list             layout.List
	bar              widget.Scrollbar
	items            map[string]*sidebarItemState
	frameItems       map[string]struct{}
	itemIndex        map[string]int
	keyFilters       []event.Filter
	dataCache        sidebarDataCache
	disabledKeys     stateutil.StringSetCache
	focusedKey       string
	pressedKey       key.Name
	pressedActionKey string
}

type sidebarDataCache struct {
	ready   bool
	version uint64
	entries []entry
	items   []Item
}

type sidebarItemState struct {
	clickable  widget.Clickable
	focus      stateutil.FocusAnimation
	keyFilters stateutil.KeyFilterCache
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

func (s *sidebarState) resolveEntries(widget Widget) ([]entry, []Item) {
	if !widget.hasDataVersion {
		entries, items := widget.entriesAndItems()
		s.checkItems(items)
		s.dataCache.ready = false
		return entries, items
	}
	if s.dataCache.ready && s.dataCache.version == widget.dataVersion {
		return s.dataCache.entries, s.dataCache.items
	}
	entries, items := widget.entriesAndItems()
	s.checkItems(items)
	s.dataCache = sidebarDataCache{ready: true, version: widget.dataVersion, entries: entries, items: items}
	return entries, items
}

func (s *sidebarState) checkItems(items []Item) {
	validateSidebarItems(items)
	if s.itemIndex == nil {
		s.itemIndex = make(map[string]int, len(items))
	} else {
		clear(s.itemIndex)
	}
	for index, item := range items {
		s.itemIndex[item.Key] = index
	}
}

type sidebarKeyResult struct {
	focusKey  string
	actionKey string
}

func (s *sidebarState) updateKeys(gtx layout.Context, sidebar Widget, items []Item) sidebarKeyResult {
	s.keyFilters = s.keyFilters[:0]
	current := s.focusedIndex(gtx)
	if current < 0 {
		if index, ok := s.itemIndex[sidebar.selectedKey]; ok {
			current = index
			s.focusedKey = sidebar.selectedKey
		}
	}
	for itemKey, itemState := range s.items {
		index, ok := s.itemIndex[itemKey]
		if !ok || index < 0 || index >= len(items) {
			continue
		}
		tag := &itemState.clickable
		if sidebar.itemDisabled(items[index]) {
			s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag,
				key.NameDownArrow,
				key.NameUpArrow,
				key.NameHome,
				key.NameEnd,
			)...)
		} else {
			s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag,
				key.NameDownArrow,
				key.NameUpArrow,
				key.NameHome,
				key.NameEnd,
				key.NameEnter,
				key.NameReturn,
				key.NameSpace,
			)...)
		}
	}
	if len(s.keyFilters) == 0 {
		return sidebarKeyResult{}
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

func (s *sidebarState) focusedIndex(gtx layout.Context) int {
	if s.focusedKey != "" {
		if index, ok := s.itemIndex[s.focusedKey]; ok {
			if itemState := s.items[s.focusedKey]; itemState != nil && gtx.Focused(&itemState.clickable) {
				return index
			}
		}
	}
	for key, itemState := range s.items {
		if index, ok := s.itemIndex[key]; ok && gtx.Focused(&itemState.clickable) {
			s.focusedKey = key
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
