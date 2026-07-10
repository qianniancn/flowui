package flowui

import (
	"fmt"
	"image"
	"math"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

const (
	tabsColorDuration     = 150 * time.Millisecond
	tabsIndicatorDuration = 250 * time.Millisecond
)

type tabsState struct {
	list             layout.List
	items            map[string]*tabsItemState
	frameItems       map[string]struct{}
	itemKeys         map[string]struct{}
	keyFilters       []event.Filter
	previous         widget.Clickable
	next             widget.Clickable
	indicator        tabsIndicatorState
	selectedKey      string
	selectionSet     bool
	selectionPending bool
}

func (s *tabsState) beginFrame() {
	if s.frameItems == nil {
		s.frameItems = make(map[string]struct{})
	} else {
		clear(s.frameItems)
	}
}

func (s *tabsState) endFrame() {
	for key := range s.items {
		if _, ok := s.frameItems[key]; !ok {
			delete(s.items, key)
		}
	}
}

func (s *tabsState) item(key string) *tabsItemState {
	if s.items == nil {
		s.items = make(map[string]*tabsItemState)
	}
	s.frameItems[key] = struct{}{}
	if item := s.items[key]; item != nil {
		return item
	}
	item := new(tabsItemState)
	s.items[key] = item
	return item
}

func (s *tabsState) checkItems(items []TabItem) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty tab item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate tab item key %q", item.Key))
		}
		s.itemKeys[item.Key] = struct{}{}
		s.item(item.Key)
	}
}

func (s *tabsState) syncSelection(items []TabItem, selectedKey string) {
	if s.selectionSet && s.selectedKey == selectedKey {
		return
	}
	s.selectionSet = true
	s.selectedKey = selectedKey
	s.selectionPending = true
	index := tabsIndexByKey(items, selectedKey)
	if index < 0 {
		s.selectionPending = false
		return
	}
}

func (s *tabsState) ensureSelectionVisible(items []TabItem, selectedKey string) bool {
	if !s.selectionPending {
		return false
	}
	index := tabsIndexByKey(items, selectedKey)
	if index < 0 || s.list.Position.Count == 0 {
		return false
	}
	if s.selectionFullyVisible(index) {
		s.selectionPending = false
		return false
	}
	previous := s.list.Position
	s.ensureVisible(index)
	return previous.First != s.list.Position.First || previous.Offset != s.list.Position.Offset
}

func (s *tabsState) selectionFullyVisible(index int) bool {
	position := s.list.Position
	if position.Count == 0 || index < position.First || index >= position.First+position.Count {
		return false
	}
	if index == position.First && position.Offset > 0 {
		return false
	}
	if position.Count == 1 && index == position.First && position.Offset == 0 {
		return true
	}
	return index != position.First+position.Count-1 || position.OffsetLast >= 0
}

func (s *tabsState) updateKeys(gtx layout.Context, items []TabItem, selectedKey string, orientation TabsOrientation) (string, bool) {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range items {
		if item.Disabled {
			continue
		}
		tag := &s.item(item.Key).clickable
		if orientation == TabsVertical {
			s.keyFilters = append(s.keyFilters,
				key.Filter{Focus: tag, Name: key.NameDownArrow},
				key.Filter{Focus: tag, Name: key.NameUpArrow},
			)
		} else {
			s.keyFilters = append(s.keyFilters,
				key.Filter{Focus: tag, Name: key.NameRightArrow},
				key.Filter{Focus: tag, Name: key.NameLeftArrow},
			)
		}
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
		)
	}
	if len(s.keyFilters) == 0 {
		return "", false
	}

	current := s.focusedIndex(gtx, items)
	if current < 0 {
		current = tabsIndexByKey(items, selectedKey)
	}
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
		case key.NameRightArrow, key.NameDownArrow:
			target, _ = tabsMoveIndex(items, current, 1)
		case key.NameLeftArrow, key.NameUpArrow:
			target, _ = tabsMoveIndex(items, current, -1)
		case key.NameHome:
			target, _ = tabsFirstEnabled(items)
		case key.NameEnd:
			target, _ = tabsLastEnabled(items)
		}
		if target >= 0 {
			current = target
		}
	}
	if target < 0 || items[target].Key == selectedKey {
		return "", false
	}
	s.ensureVisible(target)
	return items[target].Key, true
}

func (s *tabsState) focusedIndex(gtx layout.Context, items []TabItem) int {
	for index, item := range items {
		if item.Disabled {
			continue
		}
		if gtx.Focused(&s.item(item.Key).clickable) {
			return index
		}
	}
	return -1
}

func (s *tabsState) ensureVisible(index int) {
	position := &s.list.Position
	if position.Count == 0 {
		return
	}
	if index < position.First {
		position.First = index
		position.Offset = 0
		return
	}
	if index == position.First && position.Offset > 0 {
		position.Offset = 0
		return
	}
	if index >= position.First+position.Count {
		position.First = index
		position.Offset = 0
		return
	}
	if index == position.First+position.Count-1 && position.OffsetLast < 0 {
		position.Offset -= position.OffsetLast
	}
}

func (s *tabsState) scrollPage(direction, count int) {
	step := max(s.list.Position.Count-1, 1)
	first := s.list.Position.First + direction*step
	s.list.Position.First = min(max(first, 0), max(count-1, 0))
	s.list.Position.Offset = 0
}

func (s *tabsState) canScrollPrevious() bool {
	return s.list.Position.First > 0 || s.list.Position.Offset > 0
}

func (s *tabsState) canScrollNext(count int) bool {
	position := s.list.Position
	return position.First+position.Count < count || position.OffsetLast < 0
}

type tabsItemState struct {
	clickable     widget.Clickable
	interaction   buttonState
	selection     float32
	selectionFrom float32
	selectionTo   float32
	selectionAt   time.Time
	selectionSet  bool
}

type tabsIndicatorState struct {
	key         string
	orientation TabsOrientation
	from        image.Rectangle
	to          image.Rectangle
	at          time.Time
	set         bool
}

func (s *tabsIndicatorState) transition(gtx layout.Context, key string, orientation TabsOrientation, target image.Rectangle) image.Rectangle {
	if key == "" || target.Empty() {
		s.set = false
		return image.Rectangle{}
	}
	if !s.set || s.orientation != orientation {
		s.key = key
		s.orientation = orientation
		s.from = target
		s.to = target
		s.at = gtx.Now
		s.set = true
		return target
	}

	if s.key != key {
		s.from = s.current(gtx)
		s.to = target
		s.at = gtx.Now
		s.key = key
	} else if s.to != target {
		if s.to.Size() == target.Size() {
			delta := target.Min.Sub(s.to.Min)
			s.from = s.from.Add(delta)
			s.to = target
		} else {
			s.from = target
			s.to = target
			s.at = gtx.Now
		}
	}
	return s.current(gtx)
}

func (s *tabsIndicatorState) current(gtx layout.Context) image.Rectangle {
	if s.from == s.to {
		return s.to
	}
	progress := animationEase(animationProgress(gtx.Now.Sub(s.at), tabsIndicatorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	} else {
		s.from = s.to
		return s.to
	}
	return image.Rect(
		int(math.Round(float64(lerp(float32(s.from.Min.X), float32(s.to.Min.X), progress)))),
		int(math.Round(float64(lerp(float32(s.from.Min.Y), float32(s.to.Min.Y), progress)))),
		int(math.Round(float64(lerp(float32(s.from.Max.X), float32(s.to.Max.X), progress)))),
		int(math.Round(float64(lerp(float32(s.from.Max.Y), float32(s.to.Max.Y), progress)))),
	)
}

func (s *tabsItemState) selectionProgress(gtx layout.Context, selected bool) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	if !s.selectionSet {
		s.selection = target
		s.selectionFrom = target
		s.selectionTo = target
		s.selectionAt = gtx.Now
		s.selectionSet = true
		return target
	}
	if target != s.selectionTo {
		s.selectionFrom = s.selection
		s.selectionTo = target
		s.selectionAt = gtx.Now
	}
	if s.selectionFrom == s.selectionTo {
		s.selection = s.selectionTo
		return s.selection
	}
	progress := animationEase(animationProgress(gtx.Now.Sub(s.selectionAt), tabsColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.selection = lerp(s.selectionFrom, s.selectionTo, progress)
	return s.selection
}

func tabsIndexByKey(items []TabItem, key string) int {
	for index, item := range items {
		if item.Key == key {
			return index
		}
	}
	return -1
}

func tabsMoveIndex(items []TabItem, current, delta int) (int, bool) {
	if len(items) == 0 {
		return -1, false
	}
	if current < 0 || current >= len(items) {
		if delta < 0 {
			return tabsLastEnabled(items)
		}
		return tabsFirstEnabled(items)
	}
	index := current
	for range len(items) {
		index = (index + delta + len(items)) % len(items)
		if !items[index].Disabled {
			return index, true
		}
	}
	return -1, false
}

func tabsFirstEnabled(items []TabItem) (int, bool) {
	for index, item := range items {
		if !item.Disabled {
			return index, true
		}
	}
	return -1, false
}

func tabsLastEnabled(items []TabItem) (int, bool) {
	for index := len(items) - 1; index >= 0; index-- {
		if !items[index].Disabled {
			return index, true
		}
	}
	return -1, false
}
