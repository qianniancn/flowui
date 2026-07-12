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
	key                 string
	list                layout.List
	items               map[string]*menuItemState
	frameItems          map[string]struct{}
	itemKeys            map[string]struct{}
	keyFilters          []event.Filter
	pressedKey          key.Name
	pressedActionKey    string
	typeahead           string
	typeaheadAt         time.Time
	typeaheadReady      bool
	anchors             map[string]image.Rectangle
	openSubmenu         string
	submenuFocusVisible bool
	hoverSubmenu        string
	hoverAt             time.Time
	dismiss             [16]overlay.ClickArea
	dialog              overlay.ClickArea
	progressValue       float32
	progressFrom        float32
	progressTo          float32
	progressAt          time.Time
	progressReady       bool
	submenuWasOpen      bool
	focusPending        bool
}

func (s *menuState) reveal(entries []entry, key string) {
	index := menuEntryIndex(entries, key)
	if index < 0 || s.list.Position.Count == 0 {
		return
	}
	first := s.list.Position.First
	if index < first || index >= first+s.list.Position.Count {
		s.list.ScrollTo(index)
	}
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
	value := frame.UseStateWith(ctx, key, stateSlotMenu, func() *menuState {
		return &menuState{key: key}
	})
	return value
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

func (s *menuState) checkItems(items []Item) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty menu item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate menu item key %q", item.Key))
		}
		s.itemKeys[item.Key] = struct{}{}
	}
}

func (s *menuState) updateKeys(gtx layout.Context, items []Item, nested bool) keyResult {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range items {
		tag := &s.item(item.Key).clickable
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
			key.Filter{Focus: tag, Name: key.NameRightArrow},
		)
		if nested {
			s.keyFilters = append(s.keyFilters, key.Filter{Focus: tag, Name: key.NameLeftArrow})
		}
		if item.Disabled {
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
	current := s.focusedIndex(gtx, items)
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
				current = menuMoveIndex(items, current, 1)
				if current >= 0 {
					result.focusKey = items[current].Key
				}
			}
		case key.NameUpArrow:
			if event.State == key.Press {
				current = menuMoveIndex(items, current, -1)
				if current >= 0 {
					result.focusKey = items[current].Key
				}
			}
		case key.NameHome:
			if event.State == key.Press {
				current = menuFirstEnabled(items)
				if current >= 0 {
					result.focusKey = items[current].Key
				}
			}
		case key.NameEnd:
			if event.State == key.Press {
				current = menuLastEnabled(items)
				if current >= 0 {
					result.focusKey = items[current].Key
				}
			}
		case key.NameRightArrow:
			if event.State == key.Press && current >= 0 && items[current].Kind == ItemSubmenu && !items[current].Disabled {
				result.openKey = items[current].Key
			}
		case key.NameLeftArrow:
			if event.State == key.Press && nested {
				result.close = true
			}
		case key.NameEnter, key.NameReturn, key.NameSpace:
			s.updateActivation(event, items, current, &result)
		default:
			if event.State != key.Press || event.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt|key.ModSuper) != 0 {
				continue
			}
			text := menuTypeaheadText(event.Name)
			if text == "" {
				continue
			}
			query := s.appendTypeahead(gtx.Now, text)
			if next := menuTypeaheadIndex(items, current, query); next >= 0 {
				current = next
				result.focusKey = items[next].Key
			}
		}
	}
}

func (s *menuState) updateActivation(event key.Event, items []Item, current int, result *keyResult) {
	switch event.State {
	case key.Press:
		s.pressedKey = event.Name
		s.pressedActionKey = ""
		if current >= 0 && current < len(items) && !items[current].Disabled {
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

func (s *menuState) focusedIndex(gtx layout.Context, items []Item) int {
	for index, item := range items {
		if gtx.Focused(&s.item(item.Key).clickable) {
			return index
		}
	}
	return -1
}

func (s *menuState) appendTypeahead(now time.Time, text string) string {
	if !s.typeaheadReady || now.Before(s.typeaheadAt) || now.Sub(s.typeaheadAt) > menuTypeaheadTimeout {
		s.typeahead = ""
	}
	s.typeahead += text
	s.typeaheadAt = now
	s.typeaheadReady = true
	return s.typeahead
}

func menuTypeaheadText(name key.Name) string {
	runes := []rune(string(name))
	if len(runes) != 1 || unicode.IsControl(runes[0]) {
		return ""
	}
	return strings.ToLower(string(runes[0]))
}

func menuTypeaheadIndex(items []Item, current int, query string) int {
	if len(items) == 0 || query == "" {
		return -1
	}
	query = strings.ToLower(query)
	for step := 1; step <= len(items); step++ {
		index := (current + step + len(items)) % len(items)
		if items[index].Disabled {
			continue
		}
		if strings.HasPrefix(strings.ToLower(items[index].Label), query) {
			return index
		}
	}
	return -1
}

func menuMoveIndex(items []Item, current, delta int) int {
	if len(items) == 0 {
		return -1
	}
	for step := 1; step <= len(items); step++ {
		index := (current + delta*step) % len(items)
		if index < 0 {
			index += len(items)
		}
		if !items[index].Disabled {
			return index
		}
	}
	return -1
}

func menuFirstEnabled(items []Item) int {
	for index, item := range items {
		if !item.Disabled {
			return index
		}
	}
	return -1
}

func menuLastEnabled(items []Item) int {
	for index := len(items) - 1; index >= 0; index-- {
		if !items[index].Disabled {
			return index
		}
	}
	return -1
}

func itemByKey(items []Item, key string) (Item, bool) {
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func menuEntryIndex(entries []entry, key string) int {
	for index, entry := range entries {
		if entry.kind != ItemSeparator && entry.kind != ItemGroupLabel && entry.item.Key == key {
			return index
		}
	}
	return -1
}
