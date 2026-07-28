package tabs

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
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
)

const (
	stateSlotTabs         = "tabs"
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
	disclosure       disclosure.Binding[string]
	selectedKey      string
	selectionSet     bool
	selectionPending bool
}

// tabsDisclosureCfg builds a disclosure.Config from the widget's selected-key fields.
func tabsDisclosureCfg(widget TabsWidget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasSelectedKey,
		Value:      widget.selectedKey,
		HasDefault: widget.hasDefaultSelected,
		Default:    widget.defaultSelectedKey,
		OnChange:   widget.onChange,
	}
}

func (s *tabsState) currentSelectedKey(widget TabsWidget) string {
	s.selectedKey = s.disclosure.Current(tabsDisclosureCfg(widget))
	return s.selectedKey
}

func (s *tabsState) bind(widget TabsWidget) {
	s.disclosure.Bind(tabsDisclosureCfg(widget))
}

func (s *tabsState) requestSelectedKey(widget TabsWidget, key string) string {
	s.selectedKey, _ = s.disclosure.Request(tabsDisclosureCfg(widget), key)
	return s.selectedKey
}

func (s *tabsState) beginFrame() {
	state.BeginFrameMap(&s.frameItems)
}

func (s *tabsState) endFrame() {
	state.SweepFrameMap(s.items, s.frameItems)
}

func (s *tabsState) item(key string) *tabsItemState {
	return state.UseFrameMap(&s.items, &s.frameItems, key)
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
		itemState := s.item(item.Key)
		tag := &itemState.clickable
		if orientation == TabsVertical {
			s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag,
				key.NameDownArrow,
				key.NameUpArrow,
				key.NameHome,
				key.NameEnd,
			)...)
		} else {
			s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag,
				key.NameRightArrow,
				key.NameLeftArrow,
				key.NameHome,
				key.NameEnd,
			)...)
		}
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
	clickable   widget.Clickable
	interaction state.FocusAnimation
	selection   animation.FloatTransition
	keyFilters  state.KeyFilterCache
}

func tabsStateFor(ctx *frame.Context, key string) *tabsState {
	key = frame.ClaimKey(ctx, state.KindTabs, key)
	return frame.UseState[tabsState](ctx, key, stateSlotTabs)
}

type tabsIndicatorState struct {
	key         string
	orientation TabsOrientation
	from        image.Rectangle
	to          image.Rectangle
	at          time.Time
	set         bool
}

func (s *tabsIndicatorState) transition(gtx layout.Context, key string, orientation TabsOrientation, target image.Rectangle, motions ...theme.MotionTheme) image.Rectangle {
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
		s.from = s.current(gtx, motions...)
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
	return s.current(gtx, motions...)
}

func (s *tabsIndicatorState) current(gtx layout.Context, motions ...theme.MotionTheme) image.Rectangle {
	if s.from == s.to {
		return s.to
	}
	duration := tabsIndicatorDuration
	if len(motions) > 0 {
		duration = theme.ResolveMotionDuration(motions[0], duration)
	}
	if duration <= 0 {
		s.from = s.to
		return s.to
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	} else {
		s.from = s.to
		return s.to
	}
	return image.Rect(
		int(math.Round(float64(render.Lerp(float32(s.from.Min.X), float32(s.to.Min.X), progress)))),
		int(math.Round(float64(render.Lerp(float32(s.from.Min.Y), float32(s.to.Min.Y), progress)))),
		int(math.Round(float64(render.Lerp(float32(s.from.Max.X), float32(s.to.Max.X), progress)))),
		int(math.Round(float64(render.Lerp(float32(s.from.Max.Y), float32(s.to.Max.Y), progress)))),
	)
}

func (s *tabsItemState) selectionProgress(gtx layout.Context, selected bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	return s.selection.Value(gtx, target, tabsColorDuration, animation.EaseSmoothstep, motions...)
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
