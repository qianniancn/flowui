package menu

import (
	"fmt"
	"image"
	"strings"
	"time"
	"unicode"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotMenu = "menu"
const menuTypeaheadTimeout = 500 * time.Millisecond

type menuState struct {
	key                   string
	list                  layout.List
	bar                   widget.Scrollbar
	items                 map[string]*menuItemState
	frameItems            map[string]struct{}
	itemKeys              map[string]struct{}
	keyFilters            []event.Filter
	pressedKey            key.Name
	pressedActionKey      string
	typeaheadText         string
	typeaheadAt           time.Time
	typeaheadReady        bool
	anchors               map[string]image.Rectangle
	openSubmenu           string
	submenuActive         bool
	submenuFocusVisible   bool
	hoverSubmenu          string
	hoverAt               time.Time
	dismiss               [16]overlay.ClickArea
	dialog                overlay.ClickArea
	progressValue         float32
	progressFrom          float32
	progressTo            float32
	progressAt            time.Time
	progressReady         bool
	submenuWasOpen        bool
	focusPending          bool
	requestedFocusVisible bool
}

type menuItemState struct {
	clickable widget.Clickable
	focus     state.FocusAnimation
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
	if s.frameItems == nil {
		s.frameItems = make(map[string]struct{})
	} else {
		clear(s.frameItems)
	}
	if s.anchors == nil {
		s.anchors = make(map[string]image.Rectangle)
	} else {
		clear(s.anchors)
	}
}

func (s *menuState) endFrame() {
	for key := range s.items {
		if _, ok := s.frameItems[key]; !ok {
			delete(s.items, key)
		}
	}
	if s.openSubmenu != "" {
		if _, ok := s.frameItems[s.openSubmenu]; !ok {
			s.openSubmenu = ""
		}
	}
}

func (s *menuState) item(key string) *menuItemState {
	if s.items == nil {
		s.items = make(map[string]*menuItemState)
	}
	if s.frameItems == nil {
		s.frameItems = make(map[string]struct{})
	}
	s.frameItems[key] = struct{}{}
	if item := s.items[key]; item != nil {
		return item
	}
	item := new(menuItemState)
	s.items[key] = item
	return item
}

func (s *menuState) checkEntries(entries []entry) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(entries))
	} else {
		clear(s.itemKeys)
	}
	for _, entry := range entries {
		if entry.item.Key == "" {
			panic("flowui: empty menu item key")
		}
		if _, ok := s.itemKeys[entry.item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate menu item key %q", entry.item.Key))
		}
		s.itemKeys[entry.item.Key] = struct{}{}
	}
}

func (s *menuState) updateKeys(gtx layout.Context, widget Widget, entries []entry, nested bool) keyResult {
	s.keyFilters = s.keyFilters[:0]
	for _, entry := range entries {
		tag := &s.item(entry.item.Key).clickable
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
			key.Filter{Focus: tag, Name: key.NameRightArrow},
		)
		if nested || widget.onRootPrevious != nil {
			s.keyFilters = append(s.keyFilters, key.Filter{Focus: tag, Name: key.NameLeftArrow})
		}
		if widget.itemDisabled(entry.item) {
			continue
		}
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameEnter},
			key.Filter{Focus: tag, Name: key.NameReturn},
			key.Filter{Focus: tag, Name: key.NameSpace},
			key.Filter{Focus: tag},
		)
	}
	if len(s.keyFilters) == 0 {
		return keyResult{}
	}
	current := s.focusedIndex(gtx, entries)
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
			text := menuKeyText(event.Name)
			if text == "" {
				continue
			}
			query := s.appendTypeahead(gtx.Now, text)
			next := menuTypeaheadIndex(widget, entries, current, query)
			if next < 0 && query != text {
				s.typeaheadText = text
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

func (s *menuState) focusedIndex(gtx layout.Context, entries []entry) int {
	for index, entry := range entries {
		if gtx.Focused(&s.item(entry.item.Key).clickable) {
			return index
		}
	}
	return -1
}

func (s *menuState) appendTypeahead(now time.Time, text string) string {
	if !s.typeaheadReady || now.Before(s.typeaheadAt) || now.Sub(s.typeaheadAt) > menuTypeaheadTimeout {
		s.typeaheadText = ""
	}
	s.typeaheadText += text
	s.typeaheadAt = now
	s.typeaheadReady = true
	return s.typeaheadText
}

func menuKeyText(name key.Name) string {
	runes := []rune(string(name))
	if len(runes) != 1 || unicode.IsControl(runes[0]) {
		return ""
	}
	return strings.ToLower(string(runes[0]))
}

func menuTypeaheadIndex(widget Widget, entries []entry, current int, query string) int {
	if len(entries) == 0 || query == "" {
		return -1
	}
	query = strings.ToLower(query)
	for step := 1; step <= len(entries); step++ {
		index := (current + step + len(entries)) % len(entries)
		if widget.itemDisabled(entries[index].item) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(entries[index].item.Label), query) {
			return index
		}
	}
	return -1
}

func menuMoveIndex(widget Widget, entries []entry, current, delta int) int {
	if len(entries) == 0 {
		return -1
	}
	for step := 1; step <= len(entries); step++ {
		index := (current + delta*step) % len(entries)
		if index < 0 {
			index += len(entries)
		}
		if !widget.itemDisabled(entries[index].item) {
			return index
		}
	}
	return -1
}

func menuFirstEnabled(widget Widget, entries []entry) int {
	for index, entry := range entries {
		if !widget.itemDisabled(entry.item) {
			return index
		}
	}
	return -1
}

func menuLastEnabled(widget Widget, entries []entry) int {
	for index := len(entries) - 1; index >= 0; index-- {
		if !widget.itemDisabled(entries[index].item) {
			return index
		}
	}
	return -1
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
	item.focus.Prepare(visible)
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
