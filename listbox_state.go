package flowui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
)

type listBoxState struct {
	list             layout.List
	items            map[string]*listBoxItemState
	frameItems       map[string]struct{}
	itemKeys         map[string]struct{}
	keyFilters       []event.Filter
	pressedKey       key.Name
	pressedActionKey string
}

type listBoxKeyResult struct {
	focusKey  string
	actionKey string
}

func (s *listBoxState) beginFrame() {
	if s.frameItems == nil {
		s.frameItems = make(map[string]struct{})
	} else {
		clear(s.frameItems)
	}
}

func (s *listBoxState) endFrame() {
	for key := range s.items {
		if _, ok := s.frameItems[key]; !ok {
			delete(s.items, key)
		}
	}
}

func (s *listBoxState) item(key string) *listBoxItemState {
	if s.items == nil {
		s.items = make(map[string]*listBoxItemState)
	}
	s.frameItems[key] = struct{}{}
	if item := s.items[key]; item != nil {
		return item
	}
	item := new(listBoxItemState)
	s.items[key] = item
	return item
}

func (s *listBoxState) checkItems(items []ListBoxItem) {
	if s.itemKeys == nil {
		s.itemKeys = make(map[string]struct{}, len(items))
	} else {
		clear(s.itemKeys)
	}
	for _, item := range items {
		if item.Key == "" {
			panic("flowui: empty listbox item key")
		}
		if _, ok := s.itemKeys[item.Key]; ok {
			panic(fmt.Sprintf("flowui: duplicate listbox item key %q", item.Key))
		}
		s.itemKeys[item.Key] = struct{}{}
	}
}

func (s *listBoxState) updateKeys(gtx layout.Context, items []ListBoxItem, disabledKeys []string, selectedKey string) listBoxKeyResult {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range items {
		tag := &s.item(item.Key).clickable
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
		)
		if listBoxItemDisabled(item, disabledKeys) {
			continue
		}
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameEnter},
			key.Filter{Focus: tag, Name: key.NameReturn},
			key.Filter{Focus: tag, Name: key.NameSpace},
		)
	}
	if len(s.keyFilters) == 0 {
		return listBoxKeyResult{}
	}

	current := s.focusedIndex(gtx, items)
	if current < 0 {
		current = listBoxIndexByKey(items, selectedKey)
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
		}
	}
	return result
}

func (s *listBoxState) focusedIndex(gtx layout.Context, items []ListBoxItem) int {
	for i, item := range items {
		if gtx.Focused(&s.item(item.Key).clickable) {
			return i
		}
	}
	return -1
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

type listBoxItemState struct {
	clickable     widget.Clickable
	bg            color.NRGBA
	bgFrom        color.NRGBA
	bgTo          color.NRGBA
	bgAt          time.Time
	bgReady       bool
	selected      float32
	selectedFrom  float32
	selectedTo    float32
	selectedAt    time.Time
	selectedReady bool
	focus         float32
	focusFrom     float32
	focusTo       float32
	focusAt       time.Time
	focusReady    bool
	focused       bool
	pointerFocus  bool
}

func (s *listBoxItemState) background(gtx layout.Context, target color.NRGBA) color.NRGBA {
	return listBoxAnimateColor(gtx, target, &s.bg, &s.bgFrom, &s.bgTo, &s.bgAt, &s.bgReady)
}

func (s *listBoxItemState) selection(gtx layout.Context, selected bool) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	return listBoxAnimateFloat(gtx, target, &s.selected, &s.selectedFrom, &s.selectedTo, &s.selectedAt, &s.selectedReady, listBoxItemSelectDuration)
}

func (s *listBoxItemState) focusOpacity(gtx layout.Context, focused bool) float32 {
	target := float32(0)
	if focused {
		target = 1
	}
	return listBoxAnimateFloat(gtx, target, &s.focus, &s.focusFrom, &s.focusTo, &s.focusAt, &s.focusReady, listBoxItemFocusDuration)
}

func (s *listBoxItemState) focusVisible(focused bool, history []widget.Press) bool {
	if !focused {
		s.focused = false
		s.pointerFocus = false
		return false
	}
	if !s.focused {
		s.focused = true
		s.pointerFocus = len(history) > 0
	}
	return !s.pointerFocus
}

func listBoxAnimateColor(gtx layout.Context, target color.NRGBA, value, from, to *color.NRGBA, at *time.Time, ready *bool) color.NRGBA {
	if !*ready {
		*value = target
		*from = target
		*to = target
		*at = gtx.Now
		*ready = true
		return target
	}
	if target != *to {
		*from = *value
		*to = target
		*at = gtx.Now
	}
	if *from == *to {
		*value = *to
		return *value
	}
	progress := animationEase(animationProgress(gtx.Now.Sub(*at), listBoxItemColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	*value = lerpColor(*from, *to, progress)
	return *value
}

func listBoxAnimateFloat(gtx layout.Context, target float32, value, from, to *float32, at *time.Time, ready *bool, duration time.Duration) float32 {
	if !*ready {
		*value = target
		*from = target
		*to = target
		*at = gtx.Now
		*ready = true
		return target
	}
	if target != *to {
		*from = *value
		*to = target
		*at = gtx.Now
	}
	if *from == *to {
		*value = *to
		return *value
	}
	progress := animationEase(animationProgress(gtx.Now.Sub(*at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	*value = lerp(*from, *to, progress)
	return *value
}

func listBoxItemScale(gtx layout.Context, history []widget.Press, theme *Theme, disabled bool) float32 {
	if disabled || len(history) == 0 {
		return 1
	}
	target := theme.Components.ListBox.PressedScale
	if target <= 0 || target > 1 {
		target = 0.98
	}
	press := history[len(history)-1]
	if press.End.IsZero() {
		progress := animationEase(animationProgress(gtx.Now.Sub(press.Start), listBoxItemPressInDuration))
		if progress < 1 {
			gtx.Execute(op.InvalidateCmd{})
		}
		return lerp(1, target, progress)
	}
	progress := animationEase(animationProgress(gtx.Now.Sub(press.End), listBoxItemPressOutDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	return lerp(target, 1, progress)
}

func listBoxItemTransform(size image.Point, scale float32) op.TransformOp {
	origin := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	return op.Affine(f32.AffineId().Scale(origin, f32.Pt(scale, scale)))
}
