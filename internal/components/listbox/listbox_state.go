package listbox

import (
	"fmt"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	"github.com/qianniancn/flowui/internal/components/nav"
	"github.com/qianniancn/flowui/internal/components/optionrow"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
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
	disclosureSingle disclosure.Binding[string]
	// For multiple selection, we manage state manually since []string is not comparable
	multipleValue    []string
	multipleReady    bool
	list             layout.List
	bar              widget.Scrollbar
	items            map[string]*listBoxItemState
	frameItems       map[string]struct{}
	itemKeys         map[string]int
	keyFilters       []event.Filter
	dataCache        listBoxDataCache
	selectedKeys     state.StringSetCache
	disabledKeys     state.StringSetCache
	focusedKey       string
	pressedKey       key.Name
	pressedActionKey string
	typeahead        nav.Typeahead
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

func (s *listBoxState) updateKeys(gtx layout.Context, items []ListBoxItem, widget ListBoxWidget, selectedKey string) listBoxKeyResult {
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
		if widget.itemDisabled(items[index]) {
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
				"",
			)...)
		}
	}
	if len(s.keyFilters) == 0 {
		return listBoxKeyResult{}
	}

	// itemDisabled resolves in O(1) via the widget's disabled-key set, so the
	// per-keystroke navigation below does not linearly scan disabledKeys.
	list := nav.List{
		Count:    len(items),
		Disabled: func(i int) bool { return widget.itemDisabled(items[i]) },
		Label:    func(i int) string { return items[i].Label },
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
			if next, ok := nav.Move(list, current, 1, false); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameUpArrow:
			if event.State != key.Press {
				continue
			}
			if next, ok := nav.Move(list, current, -1, false); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameHome:
			if event.State != key.Press {
				continue
			}
			if next, ok := nav.First(list); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameEnd:
			if event.State != key.Press {
				continue
			}
			if next, ok := nav.Last(list); ok {
				current = next
				result.focusKey = items[next].Key
			}
		case key.NameEnter, key.NameReturn, key.NameSpace:
			switch event.State {
			case key.Press:
				s.pressedKey = event.Name
				s.pressedActionKey = ""
				if current >= 0 && current < len(items) && !widget.itemDisabled(items[current]) {
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
				if item, ok := listBoxItemByKey(items, actionKey); ok && !widget.itemDisabled(item) {
					result.actionKey = actionKey
				}
			}
		default:
			if event.State != key.Press || event.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt|key.ModSuper) != 0 {
				continue
			}
			text := nav.Printable(event.Name)
			if text == "" {
				continue
			}
			query := s.typeahead.Append(gtx.Now, text)
			next, ok := nav.Match(list, current, query)
			if !ok && query != text {
				s.typeahead.Set(text)
				next, ok = nav.Match(list, current, text)
			}
			if ok {
				current = next
				result.focusKey = items[next].Key
			}
		}
	}
	return result
}

// appendTypeahead delegates to the shared nav.Typeahead accumulator. It remains
// a method so keyboard tests can drive the state directly.
func (s *listBoxState) appendTypeahead(now time.Time, text string) string {
	return s.typeahead.Append(now, text)
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

func listBoxItemDisabled(item ListBoxItem, disabledKeys []string) bool {
	return item.Disabled || listBoxContainsKey(disabledKeys, item.Key)
}

// sliceList adapts an item slice + disabled-key slice into a nav.List for the
// exported one-shot helpers used by Select/ComboBox. These lookups are not on
// the per-keystroke path, so the slice-based disabled check is acceptable.
func sliceList(items []ListBoxItem, disabledKeys []string) nav.List {
	return nav.List{
		Count:    len(items),
		Disabled: func(i int) bool { return listBoxItemDisabled(items[i], disabledKeys) },
		Label:    func(i int) string { return items[i].Label },
	}
}

func FirstEnabled(items []ListBoxItem, disabledKeys []string) (int, bool) {
	return nav.First(sliceList(items, disabledKeys))
}

func LastEnabled(items []ListBoxItem, disabledKeys []string) (int, bool) {
	return nav.Last(sliceList(items, disabledKeys))
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

type listBoxItemState struct {
	optionrow.FocusableState
	keyFilters state.KeyFilterCache
}

// Disclosure helpers for single selection mode
func listBoxSingleDisclosureCfg(widget ListBoxWidget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasSelectedKey,
		Value:      widget.selectedKey,
		HasDefault: widget.hasDefaultKey,
		Default:    widget.defaultSelectedKey,
		OnChange:   widget.onChange,
	}
}

func (s *listBoxState) currentSingleValue(widget ListBoxWidget) string {
	return s.disclosureSingle.Current(listBoxSingleDisclosureCfg(widget))
}

func (s *listBoxState) bindSingle(widget ListBoxWidget) {
	s.disclosureSingle.Bind(listBoxSingleDisclosureCfg(widget))
}

func (s *listBoxState) requestSingleValue(widget ListBoxWidget, value string) string {
	newValue, _ := s.disclosureSingle.Request(listBoxSingleDisclosureCfg(widget), value)
	return newValue
}

// Manual disclosure for multiple selection mode ([]string is not comparable)
func (s *listBoxState) currentMultipleValue(widget ListBoxWidget) []string {
	if widget.hasSelectedKeys {
		// Controlled mode
		return widget.selectedKeys
	}
	if !s.multipleReady {
		// First frame - initialize from default or empty
		if widget.hasDefaultKeys {
			s.multipleValue = append([]string(nil), widget.defaultSelectedKeys...)
		} else {
			s.multipleValue = nil
		}
		s.multipleReady = true
	}
	return s.multipleValue
}

func (s *listBoxState) bindMultiple(widget ListBoxWidget) {
	// For controlled mode, just validate
	if widget.hasSelectedKeys {
		return
	}
	// For uncontrolled mode, ensure initialized
	if !s.multipleReady {
		if widget.hasDefaultKeys {
			s.multipleValue = append([]string(nil), widget.defaultSelectedKeys...)
		} else {
			s.multipleValue = nil
		}
		s.multipleReady = true
	}
}

func (s *listBoxState) requestMultipleValue(widget ListBoxWidget, value []string) []string {
	if widget.hasSelectedKeys {
		// Controlled mode - just call onChange
		if widget.onSelectionChange != nil {
			widget.onSelectionChange(value)
		}
		return widget.selectedKeys
	}
	// Uncontrolled mode - update internal state
	s.multipleValue = append([]string(nil), value...)
	if widget.onSelectionChange != nil {
		widget.onSelectionChange(value)
	}
	return s.multipleValue
}
