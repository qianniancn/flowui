package listbox

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/optionrow"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotListBox = "listbox"

func listBoxStateFor(ctx *frame.Context, key string) *listBoxState {
	key = frame.ClaimKey(ctx, state.KindListBox, key)
	return frame.UseState[listBoxState](ctx, key, stateSlotListBox)
}

func (l ListBoxWidget) stateFor(ctx *frame.Context) *listBoxState {
	if l.derivedOwner == "" {
		return listBoxStateFor(ctx, l.key)
	}
	key := frame.ClaimDerivedResolvedKey(ctx, state.KindListBox, l.derivedOwner, l.derivedRole)
	return frame.UseState[listBoxState](ctx, key, stateSlotListBox)
}

type listBoxState struct {
	list             layout.List
	bar              widget.Scrollbar
	items            map[string]*listBoxItemState
	frameItems       map[string]struct{}
	itemKeys         map[string]int
	keyFilters       []event.Filter
	dataCache        listBoxDataCache
	focusedKey       string
	pressedKey       key.Name
	pressedActionKey string
	typeahead        string
	typeaheadAt      time.Time
	typeaheadReady   bool
}

type listBoxDataCache struct {
	ready   bool
	version uint64
	entries []listBoxEntry
	items   []ListBoxItem
}

const listBoxTypeaheadTimeout = 500 * time.Millisecond

type listBoxKeyResult struct {
	focusKey  string
	actionKey string
}

func (s *listBoxState) beginFrame() {
	state.BeginFrameMap(&s.frameItems)
}

func (s *listBoxState) endFrame() {
	state.SweepFrameMap(s.items, s.frameItems)
}

func (s *listBoxState) item(key string) *listBoxItemState {
	return state.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *listBoxState) resolveEntries(widget ListBoxWidget) ([]listBoxEntry, []ListBoxItem) {
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
	s.dataCache = listBoxDataCache{
		ready:   true,
		version: widget.dataVersion,
		entries: entries,
		items:   items,
	}
	return entries, items
}

func (s *listBoxState) checkItems(items []ListBoxItem) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]int, len(items))
	} else {
		clear(s.itemKeys)
	}
	for index, item := range items {
		if item.Key == "" {
			panic("flowui: empty listbox item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate listbox item key %q", item.Key))
		}
		s.itemKeys[item.Key] = index
	}
}

func (s *listBoxState) updateKeys(gtx layout.Context, items []ListBoxItem, disabledKeys []string, selectedKey string) listBoxKeyResult {
	s.keyFilters = s.keyFilters[:0]
	current := s.focusedIndex(gtx)
	if current < 0 {
		if index, ok := s.itemKeys[selectedKey]; ok {
			current = index
			s.focusedKey = selectedKey
		}
	}
	for itemKey, itemState := range s.items {
		index, ok := s.itemKeys[itemKey]
		if !ok || index < 0 || index >= len(items) {
			continue
		}
		tag := &itemState.Clickable
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
		)
		if !listBoxItemDisabled(items[index], disabledKeys) {
			s.keyFilters = append(s.keyFilters,
				key.Filter{Focus: tag, Name: key.NameEnter},
				key.Filter{Focus: tag, Name: key.NameReturn},
				key.Filter{Focus: tag, Name: key.NameSpace},
				key.Filter{Focus: tag},
			)
		}
	}
	if len(s.keyFilters) == 0 {
		return listBoxKeyResult{}
	}

	result := listBoxKeyResult{}
	for {
		e, ok := gtx.Event(s.keyFilters...)
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if !ok {
			continue
		}
		switch event.Name {
		case key.NameDownArrow:
			if event.State != key.Press {
				continue
			}
			if next, ok := listBoxMoveIndex(items, disabledKeys, current, 1); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameUpArrow:
			if event.State != key.Press {
				continue
			}
			if next, ok := listBoxMoveIndex(items, disabledKeys, current, -1); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameHome:
			if event.State != key.Press {
				continue
			}
			if next, ok := listBoxFirstEnabled(items, disabledKeys); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameEnd:
			if event.State != key.Press {
				continue
			}
			if next, ok := listBoxLastEnabled(items, disabledKeys); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameEnter, key.NameReturn, key.NameSpace:
			switch event.State {
			case key.Press:
				s.pressedKey = event.Name
				s.pressedActionKey = ""
				if current >= 0 && current < len(items) && !listBoxItemDisabled(items[current], disabledKeys) {
					s.pressedActionKey = items[current].Key
				}
			case key.Release:
				if s.pressedKey != event.Name || s.pressedActionKey == "" {
					s.pressedKey = ""
					s.pressedActionKey = ""
					continue
				}
				actionKey := s.pressedActionKey
				s.pressedKey = ""
				s.pressedActionKey = ""
				if item, ok := listBoxItemByKey(items, actionKey); ok && !listBoxItemDisabled(item, disabledKeys) {
					result.actionKey = actionKey
				}
			}
		default:
			if event.State != key.Press || event.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt|key.ModSuper) != 0 {
				continue
			}
			text := listBoxTypeaheadText(event.Name)
			if text == "" {
				continue
			}
			query := s.appendTypeahead(gtx.Now, text)
			next, ok := listBoxTypeaheadIndex(items, disabledKeys, current, query)
			if !ok && query != text {
				s.typeahead = text
				next, ok = listBoxTypeaheadIndex(items, disabledKeys, current, text)
			}
			if ok {
				current = next
				result.focusKey = items[next].Key
			}
		}
	}
	return result
}

func (s *listBoxState) appendTypeahead(now time.Time, text string) string {
	if !s.typeaheadReady || now.Before(s.typeaheadAt) || now.Sub(s.typeaheadAt) > listBoxTypeaheadTimeout {
		s.typeahead = ""
	}
	s.typeahead += text
	s.typeaheadAt = now
	s.typeaheadReady = true
	return s.typeahead
}

func listBoxTypeaheadText(name key.Name) string {
	runes := []rune(string(name))
	if len(runes) != 1 || unicode.IsControl(runes[0]) {
		return ""
	}
	return strings.ToLower(string(runes[0]))
}

func listBoxTypeaheadIndex(items []ListBoxItem, disabledKeys []string, current int, query string) (int, bool) {
	if len(items) == 0 || query == "" {
		return -1, false
	}
	query = strings.ToLower(query)
	for step := 1; step <= len(items); step++ {
		index := (current + step + len(items)) % len(items)
		if listBoxItemDisabled(items[index], disabledKeys) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(items[index].Label), query) {
			return index, true
		}
	}
	return current, false
}

func (s *listBoxState) focusedIndex(gtx layout.Context) int {
	if s.focusedKey != "" {
		if index, ok := s.itemKeys[s.focusedKey]; ok {
			if itemState := s.items[s.focusedKey]; itemState != nil && gtx.Focused(&itemState.Clickable) {
				return index
			}
		}
	}
	for key, itemState := range s.items {
		if index, ok := s.itemKeys[key]; ok && gtx.Focused(&itemState.Clickable) {
			s.focusedKey = key
			return index
		}
	}
	return -1
}

func (s *listBoxState) keyboardActiveKey(widget ListBoxWidget) string {
	if widget.selectionMode != ListBoxSelectionMultiple {
		return widget.selectedKey
	}
	for _, key := range widget.selectedKeys {
		if _, ok := s.itemKeys[key]; ok {
			return key
		}
	}
	return ""
}

func listBoxIndexByKey(items []ListBoxItem, key string) int {
	for i, item := range items {
		if item.Key == key {
			return i
		}
	}
	return -1
}

func listBoxItemByKey(items []ListBoxItem, key string) (ListBoxItem, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return ListBoxItem{}, false
}

func listBoxMoveIndex(items []ListBoxItem, disabledKeys []string, current, delta int) (int, bool) {
	if len(items) == 0 {
		return -1, false
	}
	if current < 0 || current >= len(items) {
		if delta < 0 {
			return listBoxLastEnabled(items, disabledKeys)
		}
		return listBoxFirstEnabled(items, disabledKeys)
	}
	if listBoxItemDisabled(items[current], disabledKeys) {
		return listBoxNearestEnabledFrom(items, disabledKeys, current, delta)
	}
	for next := current + delta; next >= 0 && next < len(items); next += delta {
		if !listBoxItemDisabled(items[next], disabledKeys) {
			return next, true
		}
	}
	return current, false
}

func listBoxNearestEnabledFrom(items []ListBoxItem, disabledKeys []string, current, delta int) (int, bool) {
	for next := current + delta; next >= 0 && next < len(items); next += delta {
		if !listBoxItemDisabled(items[next], disabledKeys) {
			return next, true
		}
	}
	for next := current - delta; next >= 0 && next < len(items); next -= delta {
		if !listBoxItemDisabled(items[next], disabledKeys) {
			return next, true
		}
	}
	return current, false
}

func listBoxFirstEnabled(items []ListBoxItem, disabledKeys []string) (int, bool) {
	for i, item := range items {
		if !listBoxItemDisabled(item, disabledKeys) {
			return i, true
		}
	}
	return -1, false
}

func listBoxLastEnabled(items []ListBoxItem, disabledKeys []string) (int, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		if !listBoxItemDisabled(items[i], disabledKeys) {
			return i, true
		}
	}
	return -1, false
}

func listBoxItemDisabled(item ListBoxItem, disabledKeys []string) bool {
	return item.Disabled || listBoxContainsKey(disabledKeys, item.Key)
}

func FirstEnabled(items []ListBoxItem, disabledKeys []string) (int, bool) {
	return listBoxFirstEnabled(items, disabledKeys)
}

func LastEnabled(items []ListBoxItem, disabledKeys []string) (int, bool) {
	return listBoxLastEnabled(items, disabledKeys)
}

func IndexByKey(items []ListBoxItem, key string) int {
	return listBoxIndexByKey(items, key)
}

func ItemDisabled(item ListBoxItem, disabledKeys []string) bool {
	return listBoxItemDisabled(item, disabledKeys)
}

func FocusItem(ctx *frame.Context, key, itemKey string, visible bool) bool {
	return focusItem(ctx, frame.FullKey(ctx, key), itemKey, visible)
}

func FocusDerivedItem(ctx *frame.Context, owner, role, itemKey string, visible bool) bool {
	return focusItem(ctx, frame.DerivedKey(ctx, owner, role), itemKey, visible)
}

func focusItem(ctx *frame.Context, stateKey, itemKey string, visible bool) bool {
	clickable, _, ok := item(ctx, stateKey, itemKey)
	if !ok {
		return false
	}
	frame.RequestFocusVisible(ctx, clickable, visible)
	return true
}

type listBoxItemState = optionrow.FocusableState
