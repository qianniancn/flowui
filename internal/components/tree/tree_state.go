package tree

import (
	"fmt"
	"image/color"
	"strings"
	"time"
	"unicode"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotTree = "tree"

const (
	treeTypeaheadTimeout = 500 * time.Millisecond
	treeColorDuration    = 100 * time.Millisecond
	treeExpandDuration   = 200 * time.Millisecond
	treeScaleDuration    = 140 * time.Millisecond
)

type flatItem struct {
	item      Item
	depth     int
	parentKey string
}

func flattenVisibleItems(items []Item, expanded map[string]struct{}) []flatItem {
	result := make([]flatItem, 0)
	var walk func([]Item, int, string)
	walk = func(children []Item, depth int, parent string) {
		for _, item := range children {
			result = append(result, flatItem{item: item, depth: depth, parentKey: parent})
			if _, ok := expanded[item.Key]; ok && len(item.Children) > 0 {
				walk(item.Children, depth+1, item.Key)
			}
		}
	}
	walk(items, 0, "")
	return result
}

type treeState struct {
	list             layout.List
	items            map[string]*treeItemState
	frameItems       map[string]struct{}
	itemKeys         map[string]struct{}
	keyFilters       []event.Filter
	pressedKey       key.Name
	pressedActionKey string
	typeahead        string
	typeaheadAt      time.Time
	typeaheadReady   bool
}

func treeStateFor(ctx *frame.Context, key string) *treeState {
	key = frame.ClaimKey(ctx, state.KindTree, key)
	return frame.UseState[treeState](ctx, key, stateSlotTree)
}

func (s *treeState) beginFrame() {
	if s.frameItems == nil {
		s.frameItems = make(map[string]struct{})
	} else {
		clear(s.frameItems)
	}
}

func (s *treeState) endFrame() {
	for key := range s.items {
		if _, ok := s.frameItems[key]; !ok {
			delete(s.items, key)
		}
	}
}

func (s *treeState) item(key string) *treeItemState {
	if s.items == nil {
		s.items = make(map[string]*treeItemState)
	}
	s.frameItems[key] = struct{}{}
	if item := s.items[key]; item != nil {
		return item
	}
	item := new(treeItemState)
	s.items[key] = item
	return item
}

func (s *treeState) checkItems(items []Item) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{})
	} else {
		clear(s.itemKeys)
	}
	var walk func([]Item)
	walk = func(children []Item) {
		for _, item := range children {
			if item.Key == "" {
				panic("flowui: empty tree item key")
			}
			if _, ok := s.itemKeys[item.Key]; ok {
				panic(fmt.Sprintf("flowui: duplicate tree item key %q", item.Key))
			}
			s.itemKeys[item.Key] = struct{}{}
			walk(item.Children)
		}
	}
	walk(items)
}

type treeKeyResult struct {
	focusKey  string
	actionKey string
	toggleKey string
}

func (s *treeState) updateKeys(gtx layout.Context, tree Widget, visible []flatItem) treeKeyResult {
	s.keyFilters = s.keyFilters[:0]
	for _, entry := range visible {
		tag := &s.item(entry.item.Key).clickable
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
		)
		if tree.itemDisabled(entry.item) {
			continue
		}
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameRightArrow},
			key.Filter{Focus: tag, Name: key.NameLeftArrow},
			key.Filter{Focus: tag, Name: key.NameEnter},
			key.Filter{Focus: tag, Name: key.NameReturn},
			key.Filter{Focus: tag, Name: key.NameSpace},
			key.Filter{Focus: tag},
		)
	}
	if len(s.keyFilters) == 0 {
		return treeKeyResult{}
	}

	current := s.focusedIndex(gtx, visible)
	if current < 0 {
		current = treeVisibleIndex(visible, tree.selectedKey)
	}
	result := treeKeyResult{}
	expanded := treeKeySet(tree.expandedKeys)
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
			if event.State == key.Press {
				if next, ok := treeMoveVisible(visible, tree, current, 1); ok {
					current = next
					result.focusKey = visible[next].item.Key
				}
			}
		case key.NameUpArrow:
			if event.State == key.Press {
				if next, ok := treeMoveVisible(visible, tree, current, -1); ok {
					current = next
					result.focusKey = visible[next].item.Key
				}
			}
		case key.NameHome:
			if event.State == key.Press {
				if next, ok := treeFirstEnabled(visible, tree); ok {
					current = next
					result.focusKey = visible[next].item.Key
				}
			}
		case key.NameEnd:
			if event.State == key.Press {
				if next, ok := treeLastEnabled(visible, tree); ok {
					current = next
					result.focusKey = visible[next].item.Key
				}
			}
		case key.NameRightArrow:
			if event.State == key.Press && current >= 0 && current < len(visible) {
				entry := visible[current]
				if len(entry.item.Children) > 0 {
					if _, ok := expanded[entry.item.Key]; !ok {
						result.toggleKey = entry.item.Key
					} else if child := treeFirstEnabledChild(visible, tree, current); child >= 0 {
						current = child
						result.focusKey = visible[child].item.Key
					}
				}
			}
		case key.NameLeftArrow:
			if event.State == key.Press && current >= 0 && current < len(visible) {
				entry := visible[current]
				if len(entry.item.Children) > 0 {
					if _, ok := expanded[entry.item.Key]; ok {
						result.toggleKey = entry.item.Key
						continue
					}
				}
				if parent := treeEnabledAncestor(visible, tree, current); parent >= 0 {
					current = parent
					result.focusKey = visible[parent].item.Key
				}
			}
		case key.NameEnter, key.NameReturn, key.NameSpace:
			s.handleActivationKey(event, visible, tree, current, &result)
		default:
			if event.State != key.Press || event.Modifiers&(key.ModCtrl|key.ModCommand|key.ModAlt|key.ModSuper) != 0 {
				continue
			}
			text := treeTypeaheadText(event.Name)
			if text == "" {
				continue
			}
			query := s.appendTypeahead(gtx.Now, text)
			next, ok := treeTypeaheadIndex(visible, tree, current, query)
			if !ok && query != text {
				s.typeahead = text
				next, ok = treeTypeaheadIndex(visible, tree, current, text)
			}
			if ok {
				current = next
				result.focusKey = visible[next].item.Key
			}
		}
	}
	return result
}

func (s *treeState) handleActivationKey(event key.Event, visible []flatItem, tree Widget, current int, result *treeKeyResult) {
	switch event.State {
	case key.Press:
		s.pressedKey = event.Name
		s.pressedActionKey = ""
		if current >= 0 && current < len(visible) && !tree.itemDisabled(visible[current].item) {
			s.pressedActionKey = visible[current].item.Key
		}
	case key.Release:
		if s.pressedKey == event.Name && s.pressedActionKey != "" {
			result.actionKey = s.pressedActionKey
		}
		s.pressedKey = ""
		s.pressedActionKey = ""
	}
}

func (s *treeState) focusedIndex(gtx layout.Context, visible []flatItem) int {
	for index, entry := range visible {
		if gtx.Focused(&s.item(entry.item.Key).clickable) {
			return index
		}
	}
	return -1
}

func (s *treeState) ensureVisible(index int) {
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

func (s *treeState) appendTypeahead(now time.Time, value string) string {
	if !s.typeaheadReady || now.Before(s.typeaheadAt) || now.Sub(s.typeaheadAt) > treeTypeaheadTimeout {
		s.typeahead = ""
	}
	s.typeahead += value
	s.typeaheadAt = now
	s.typeaheadReady = true
	return s.typeahead
}

func treeTypeaheadText(name key.Name) string {
	runes := []rune(string(name))
	if len(runes) != 1 || unicode.IsControl(runes[0]) {
		return ""
	}
	return strings.ToLower(string(runes[0]))
}

type treeItemState struct {
	clickable  widget.Clickable
	toggle     overlay.ClickArea
	focus      state.FocusAnimation
	background treeColorAnimation
	expansion  treeFloatAnimation
	scale      treeFloatAnimation
}

type treeColorAnimation struct {
	value color.NRGBA
	from  color.NRGBA
	to    color.NRGBA
	at    time.Time
	ready bool
}

func (a *treeColorAnimation) update(gtx layout.Context, target color.NRGBA) color.NRGBA {
	if !a.ready {
		a.value, a.from, a.to, a.at, a.ready = target, target, target, gtx.Now, true
		return target
	}
	if target != a.to {
		a.from, a.to, a.at = a.value, target, gtx.Now
	}
	if a.from == a.to {
		a.value = a.to
		return a.value
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(a.at), treeColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	a.value = render.LerpColor(a.from, a.to, progress)
	return a.value
}

type treeFloatAnimation struct {
	value float32
	from  float32
	to    float32
	at    time.Time
	ready bool
}

func (a *treeFloatAnimation) update(gtx layout.Context, target float32, duration time.Duration) float32 {
	if !a.ready {
		a.value, a.from, a.to, a.at, a.ready = target, target, target, gtx.Now, true
		return target
	}
	if target != a.to {
		a.from, a.to, a.at = a.value, target, gtx.Now
	}
	if a.from == a.to {
		a.value = a.to
		return a.value
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(a.at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	a.value = render.Lerp(a.from, a.to, progress)
	return a.value
}

func treeVisibleIndex(visible []flatItem, key string) int {
	for index, entry := range visible {
		if entry.item.Key == key {
			return index
		}
	}
	return -1
}

func treeMoveVisible(visible []flatItem, tree Widget, current, delta int) (int, bool) {
	if len(visible) == 0 {
		return -1, false
	}
	if current < 0 || current >= len(visible) {
		if delta < 0 {
			return treeLastEnabled(visible, tree)
		}
		return treeFirstEnabled(visible, tree)
	}
	for next := current + delta; next >= 0 && next < len(visible); next += delta {
		if !tree.itemDisabled(visible[next].item) {
			return next, true
		}
	}
	return current, false
}

func treeFirstEnabled(visible []flatItem, tree Widget) (int, bool) {
	for index, entry := range visible {
		if !tree.itemDisabled(entry.item) {
			return index, true
		}
	}
	return -1, false
}

func treeLastEnabled(visible []flatItem, tree Widget) (int, bool) {
	for index := len(visible) - 1; index >= 0; index-- {
		if !tree.itemDisabled(visible[index].item) {
			return index, true
		}
	}
	return -1, false
}

func treeFirstEnabledChild(visible []flatItem, tree Widget, current int) int {
	depth := visible[current].depth + 1
	for index := current + 1; index < len(visible) && visible[index].depth >= depth; index++ {
		if visible[index].depth == depth && !tree.itemDisabled(visible[index].item) {
			return index
		}
	}
	return -1
}

func treeEnabledAncestor(visible []flatItem, tree Widget, current int) int {
	parent := visible[current].parentKey
	for parent != "" {
		index := treeVisibleIndex(visible, parent)
		if index < 0 {
			return -1
		}
		if !tree.itemDisabled(visible[index].item) {
			return index
		}
		parent = visible[index].parentKey
	}
	return -1
}

func treeTypeaheadIndex(visible []flatItem, tree Widget, current int, query string) (int, bool) {
	if len(visible) == 0 || query == "" {
		return -1, false
	}
	query = strings.ToLower(query)
	for step := 1; step <= len(visible); step++ {
		index := (current + step + len(visible)) % len(visible)
		if tree.itemDisabled(visible[index].item) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(visible[index].item.Label), query) {
			return index, true
		}
	}
	return current, false
}

func treeKeySet(keys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if key != "" {
			set[key] = struct{}{}
		}
	}
	return set
}

func treeContainsKey(keys []string, key string) bool {
	for _, current := range keys {
		if current == key {
			return true
		}
	}
	return false
}

func toggleTreeKey(keys []string, key string) []string {
	next := make([]string, 0, len(keys)+1)
	found := false
	seen := make(map[string]struct{}, len(keys))
	for _, current := range keys {
		if current == "" {
			continue
		}
		if current == key {
			found = true
			continue
		}
		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}
		next = append(next, current)
	}
	if !found {
		next = append(next, key)
	}
	return next
}
