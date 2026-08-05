package menu

import (
	"fmt"
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/nav"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
	"github.com/qianniancn/flowui/internal/state"
)

const stateSlotMenu = "menu"

type menuState struct {
	key                   string
	list                  layout.List
	bar                   widget.Scrollbar
	items                 map[string]*menuItemState
	frameItems            map[string]struct{}
	itemKeys              map[string]struct{}
	itemIndex             map[string]int
	keyFilters            []event.Filter
	dataCache             menuDataCache
	widthCache            menuWidthCache
	frameActionable       []entry
	focusedKey            string
	pressedKey            key.Name
	pressedActionKey      string
	typeahead             nav.Typeahead
	anchors               map[string]image.Rectangle
	openSubmenu           string
	submenuActive         bool
	submenuFocusVisible   bool
	hoverSubmenu          string
	hoverAt               time.Time
	dismiss               [16]overlay.ClickArea
	dialog                overlay.ClickArea
	transition            animation.FloatTransition
	submenuWasOpen        bool
	focusPending          bool
	requestedFocusVisible bool
}

type menuDataCache struct {
	ready                bool
	version              uint64
	autoSeparateSections bool
	entries              []entry
	actionable           []entry
}

type menuItemState struct {
	clickable  widget.Clickable
	focus      state.FocusAnimation
	keyFilters state.KeyFilterCache
}

type keyResult struct {
	focusKey  string
	actionKey string
	openKey   string
	close     bool
}

func (m Widget) stateFor(ctx *frame.Context) *menuState {
	var key string
	if m.derivedOwner == "" {
		key = frame.ClaimKey(ctx, state.KindMenu, m.key)
	} else {
		key = frame.ClaimDerivedResolvedKey(ctx, state.KindMenu, m.derivedOwner, m.derivedRole)
	}
	return frame.UseStateWith(ctx, key, stateSlotMenu, func() *menuState {
		return &menuState{key: key}
	})
}

func (s *menuState) beginFrame() {
	state.BeginFrameMap(&s.frameItems)
	if s.anchors == nil {
		s.anchors = make(map[string]image.Rectangle)
	} else {
		clear(s.anchors)
	}
}

func (s *menuState) endFrame() {
	if s.openSubmenu != "" {
		if _, ok := s.frameItems[s.openSubmenu]; !ok {
			s.openSubmenu = ""
		}
	}
	state.SweepFrameMap(s.items, s.frameItems)
}

func (s *menuState) item(key string) *menuItemState {
	return state.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *menuState) resolveEntries(widget Widget) []entry {
	if !widget.hasDataVersion {
		entries := widget.entries()
		s.frameActionable = actionableEntries(entries)
		s.checkEntries(s.frameActionable)
		s.dataCache.ready = false
		return entries
	}
	if s.dataCache.ready && s.dataCache.version == widget.dataVersion &&
		s.dataCache.autoSeparateSections == widget.autoSeparateSections {
		return s.dataCache.entries
	}
	entries := widget.entries()
	actionable := actionableEntries(entries)
	s.checkEntries(actionable)
	s.dataCache = menuDataCache{
		ready:                true,
		version:              widget.dataVersion,
		autoSeparateSections: widget.autoSeparateSections,
		entries:              entries,
		actionable:           actionable,
	}
	return entries
}

func (s *menuState) actionableEntries(entries []entry) []entry {
	if s.dataCache.ready {
		return s.dataCache.actionable
	}
	return s.frameActionable
}

func (s *menuState) checkEntries(entries []entry) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(entries))
	} else {
		clear(s.itemKeys)
	}
	if s.itemIndex == nil {
		s.itemIndex = make(map[string]int, len(entries))
	} else {
		clear(s.itemIndex)
	}
	for index, entry := range entries {
		if entry.item.Key == "" {
			panic("flowui: empty menu item key")
		}
		if _, ok := s.itemKeys[entry.item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate menu item key %q", entry.item.Key))
		}
		s.itemKeys[entry.item.Key] = struct{}{}
		s.itemIndex[entry.item.Key] = index
	}
}

func (s *menuState) updateKeys(gtx layout.Context, widget Widget, entries []entry, nested bool) keyResult {
	s.keyFilters = s.keyFilters[:0]
	current := s.focusedIndex(gtx)
	first := max(s.list.Position.First-1, 0)
	last := min(first+s.list.Position.Count+2, len(entries))
	for index := first; index < last; index++ {
		entry := entries[index]
		itemState := s.item(entry.item.Key)
		tag := &itemState.clickable
		names := [10]key.Name{
			key.NameDownArrow,
			key.NameUpArrow,
			key.NameHome,
			key.NameEnd,
			key.NameRightArrow,
		}
		count := 5
		if nested || widget.onRootPrevious != nil {
			names[count] = key.NameLeftArrow
			count++
		}
		if !widget.itemDisabled(entries[index].item) {
			names[count] = key.NameEnter
			names[count+1] = key.NameReturn
			names[count+2] = key.NameSpace
			names[count+3] = ""
			count += 4
		}
		s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag, names[:count]...)...)
	}
	if len(s.keyFilters) == 0 {
		return keyResult{}
	}
	result := keyResult{}
	for {
		e, ok := gtx.Event(s.keyFilters...)
		if !ok {
			return result
		}
		event, ok := e.(key.Event)
		if !ok {
			continue
		}
		switch event.Name {
		case key.NameDownArrow:
			if event.State == key.Press {
				current = menuMoveIndex(widget, entries, current, 1)
				if current >= 0 {
					result.focusKey = entries[current].item.Key
				}
			}
		case key.NameUpArrow:
			if event.State == key.Press {
				current = menuMoveIndex(widget, entries, current, -1)
				if current >= 0 {
					result.focusKey = entries[current].item.Key
				}
			}
		case key.NameHome:
			if event.State == key.Press {
				current = menuFirstEnabled(widget, entries)
				if current >= 0 {
					result.focusKey = entries[current].item.Key
				}
			}
		case key.NameEnd:
			if event.State == key.Press {
				current = menuLastEnabled(widget, entries)
				if current >= 0 {
					result.focusKey = entries[current].item.Key
				}
			}
		case key.NameRightArrow:
			if event.State != key.Press {
				continue
			}
			if current >= 0 && itemHasSubmenu(entries[current].item) && !widget.itemDisabled(entries[current].item) {
				result.openKey = entries[current].item.Key
			} else if widget.onRootNext != nil {
				widget.onRootNext()
			}
		case key.NameLeftArrow:
			if event.State != key.Press {
				continue
			}
			if nested {
				result.close = true
			} else if widget.onRootPrevious != nil {
				widget.onRootPrevious()
			}
		case key.NameEnter, key.NameReturn, key.NameSpace:
			s.updateActivation(event, widget, entries, current, &result)
		default:
			if event.State != key.Press || event.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt|key.ModSuper) != 0 {
				continue
			}
			text := nav.Printable(event.Name)
			if text == "" {
				continue
			}
			query := s.typeahead.Append(gtx.Now, text)
			next := menuTypeaheadIndex(widget, entries, current, query)
			if next < 0 && query != text {
				s.typeahead.Set(text)
				next = menuTypeaheadIndex(widget, entries, current, text)
			}
			if next >= 0 {
				current = next
				result.focusKey = entries[next].item.Key
			}
		}
	}
}

func (s *menuState) updateActivation(event key.Event, widget Widget, entries []entry, current int, result *keyResult) {
	switch event.State {
	case key.Press:
		s.pressedKey = event.Name
		s.pressedActionKey = ""
		if current >= 0 && current < len(entries) && !widget.itemDisabled(entries[current].item) {
			s.pressedActionKey = entries[current].item.Key
		}
	case key.Release:
		if s.pressedKey == event.Name && s.pressedActionKey != "" {
			result.actionKey = s.pressedActionKey
		}
		s.pressedKey = ""
		s.pressedActionKey = ""
	}
}

func (s *menuState) focusedIndex(gtx layout.Context) int {
	if index, ok := s.itemIndex[s.focusedKey]; ok {
		if itemState := s.items[s.focusedKey]; itemState != nil && gtx.Focused(&itemState.clickable) {
			return index
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

func actionableEntries(entries []entry) []entry {
	result := make([]entry, 0, len(entries))
	for _, entry := range entries {
		if !entry.separator && entry.sectionTitle == "" {
			result = append(result, entry)
		}
	}
	return result
}

// menuNavList adapts menu entries into a nav.List; itemDisabled also skips
// non-actionable entries (separators/section titles), preserving prior behavior.
func menuNavList(widget Widget, entries []entry) nav.List {
	return nav.List{
		Count:    len(entries),
		Disabled: func(i int) bool { return widget.itemDisabled(entries[i].item) },
		Label:    func(i int) string { return entries[i].item.Label },
	}
}

func menuTypeaheadIndex(widget Widget, entries []entry, current int, query string) int {
	if index, ok := nav.Match(menuNavList(widget, entries), current, query); ok {
		return index
	}
	return -1
}

// menuMoveIndex wraps arrow navigation (menus loop around the ends).
func menuMoveIndex(widget Widget, entries []entry, current, delta int) int {
	if next, ok := nav.Move(menuNavList(widget, entries), current, delta, true); ok {
		return next
	}
	return -1
}

func menuFirstEnabled(widget Widget, entries []entry) int {
	index, _ := nav.First(menuNavList(widget, entries))
	return index
}

func menuLastEnabled(widget Widget, entries []entry) int {
	index, _ := nav.Last(menuNavList(widget, entries))
	return index
}

func entryByKey(entries []entry, key string) (entry, bool) {
	for _, entry := range entries {
		if entry.item.Key == key {
			return entry, true
		}
	}
	return entry{}, false
}

func (s *menuState) focus(ctx *frame.Context, entry entry, visible bool) {
	item := s.item(entry.item.Key)
	frame.RequestFocusVisible(ctx, &item.clickable, visible)
}

func (s *menuState) focusFirstEntry(ctx *frame.Context, widget Widget, visible bool) bool {
	entries := widget.actionableEntries()
	index := menuFirstEnabled(widget, entries)
	if index < 0 {
		return false
	}
	s.focus(ctx, entries[index], visible)
	return true
}

func (s *menuState) focusLastEntry(ctx *frame.Context, widget Widget, visible bool) bool {
	entries := widget.actionableEntries()
	index := menuLastEnabled(widget, entries)
	if index < 0 {
		return false
	}
	s.focus(ctx, entries[index], visible)
	return true
}

func (s *menuState) reveal(allEntries []entry, key string) {
	entryIndex := -1
	for index, entry := range allEntries {
		if entry.item.Key == key {
			entryIndex = index
			break
		}
	}
	if entryIndex < 0 || s.list.Position.Count == 0 {
		return
	}
	first := s.list.Position.First
	if entryIndex < first || entryIndex >= first+s.list.Position.Count {
		s.list.ScrollTo(entryIndex)
	}
}
