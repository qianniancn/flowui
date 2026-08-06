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
	add              widget.Clickable
	indicator        tabsIndicatorState
	disclosure       disclosure.Binding[string]
	selectedKey      string
	selectionSet     bool
	selectionPending bool
	listLayoutSet    bool
	lastListSize     image.Point
	lastListAxis     TabsOrientation
	lastListCount    int
	lastListWidths   []int
	retainedPanels   map[string]struct{}
	renderedPanels   map[string]struct{}
	forceRender      bool
	destroyOnHidden  bool
	editingKey       string
	panelKey         string
	panelOpacity     animation.FloatTransition
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

// normalizeSelection keeps the effective selection usable when a controlled
// value points at a removed or disabled item. The requested fallback is still
// sent through the disclosure binding so controlled callers can update their
// model; the fallback is used optimistically for the current frame.
func (s *tabsState) normalizeSelection(widget TabsWidget, selectedKey string) string {
	index := tabsIndexByKey(widget.items, selectedKey)
	if index >= 0 && !widget.items[index].Disabled {
		return selectedKey
	}
	fallback, ok := tabsFirstEnabled(widget.items)
	if !ok {
		s.selectionPending = false
		s.selectedKey = ""
		return ""
	}
	if index >= 0 {
		fallback, ok = tabsCloseFallback(widget.items, index)
		if !ok {
			fallback, _ = tabsFirstEnabled(widget.items)
		}
	}
	effective := s.requestSelectedKey(widget, widget.items[fallback].Key)
	if widget.hasSelectedKey {
		s.selectedKey = widget.items[fallback].Key
		return s.selectedKey
	}
	return effective
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

// noteListLayout remembers the geometry that controls which items are visible.
// A changed geometry should re-run selection visibility once, while ordinary
// scroll-button movement must remain under the user's control.
func (s *tabsState) noteListLayout(size image.Point, axis TabsOrientation, widths []int, count int) bool {
	changed := !s.listLayoutSet || s.lastListSize != size || s.lastListAxis != axis || s.lastListCount != count || !sameTabWidths(s.lastListWidths, widths)
	s.listLayoutSet = true
	s.lastListSize = size
	s.lastListAxis = axis
	s.lastListCount = count
	s.lastListWidths = append(s.lastListWidths[:0], widths...)
	return changed
}

func sameTabWidths(first, second []int) bool {
	if len(first) != len(second) {
		return false
	}
	for index, width := range first {
		if second[index] != width {
			return false
		}
	}
	return true
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

func (s *tabsState) updateKeys(gtx layout.Context, items []TabItem, selectedKey string, orientation TabsOrientation, activation TabsActivationMode) (string, string, bool) {
	s.keyFilters = s.keyFilters[:0]
	for _, item := range items {
		if item.Disabled {
			continue
		}
		itemState := s.item(item.Key)
		tag := &itemState.clickable
		names := []key.Name{key.NameHome, key.NameEnd}
		if orientation == TabsVertical {
			names = append([]key.Name{key.NameDownArrow, key.NameUpArrow}, names...)
		} else {
			names = append([]key.Name{key.NameRightArrow, key.NameLeftArrow}, names...)
		}
		if activation == TabsActivationManual {
			names = append(names, key.NameReturn, key.NameEnter, key.NameSpace)
		}
		s.keyFilters = append(s.keyFilters, itemState.keyFilters.Resolve(tag, names...)...)
		// Workbench-style tab strips conventionally support Ctrl/Cmd+Tab and
		// Ctrl/Cmd+PageUp/PageDown in addition to the roving arrow keys.
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NamePageUp, Required: key.ModShortcut},
			key.Filter{Focus: tag, Name: key.NamePageDown, Required: key.ModShortcut},
			key.Filter{Focus: tag, Name: key.NameTab, Required: key.ModShortcut},
		)
	}
	if len(s.keyFilters) == 0 {
		return "", "", false
	}

	current := s.focusedIndex(gtx, items)
	if current < 0 {
		current = tabsIndexByKey(items, selectedKey)
	}
	selectionTarget := -1
	focusTarget := -1
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
			current, _ = tabsMoveIndex(items, current, 1)
			focusTarget = current
		case key.NameLeftArrow, key.NameUpArrow:
			current, _ = tabsMoveIndex(items, current, -1)
			focusTarget = current
		case key.NamePageDown, key.NameTab:
			current, _ = tabsMoveIndex(items, current, 1)
			focusTarget = current
		case key.NamePageUp:
			current, _ = tabsMoveIndex(items, current, -1)
			focusTarget = current
		case key.NameHome:
			current, _ = tabsFirstEnabled(items)
			focusTarget = current
		case key.NameEnd:
			current, _ = tabsLastEnabled(items)
			focusTarget = current
		case key.NameReturn, key.NameEnter, key.NameSpace:
			if activation == TabsActivationManual {
				selectionTarget = current
				focusTarget = current
			}
		}
		if activation == TabsActivationAutomatic && focusTarget >= 0 {
			selectionTarget = focusTarget
		}
	}
	if focusTarget >= 0 {
		s.ensureVisible(focusTarget)
	}
	selectionKey := ""
	if selectionTarget >= 0 && items[selectionTarget].Key != selectedKey {
		selectionKey = items[selectionTarget].Key
	}
	focusKey := ""
	if focusTarget >= 0 {
		focusKey = items[focusTarget].Key
	}
	return selectionKey, focusKey, selectionKey != "" || focusKey != ""
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
	close       widget.Clickable
	editor      widget.Editor
	editReady   bool
	editFocused bool
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
	return s.transitionWithDuration(gtx, key, orientation, target, tabsIndicatorDuration, motions...)
}

func (s *tabsIndicatorState) transitionWithDuration(gtx layout.Context, key string, orientation TabsOrientation, target image.Rectangle, duration time.Duration, motions ...theme.MotionTheme) image.Rectangle {
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
		s.from = s.currentWithDuration(gtx, duration, motions...)
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
	return s.currentWithDuration(gtx, duration, motions...)
}

func (s *tabsIndicatorState) currentWithDuration(gtx layout.Context, duration time.Duration, motions ...theme.MotionTheme) image.Rectangle {
	if s.from == s.to {
		return s.to
	}
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
	return s.selectionProgressWithDuration(gtx, selected, tabsColorDuration, motions...)
}

func (s *tabsItemState) selectionProgressWithDuration(gtx layout.Context, selected bool, duration time.Duration, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	if selected {
		target = 1
	}
	return s.selection.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *tabsState) panelOpacityFor(gtx layout.Context, key string, transition TabsPanelTransition, duration time.Duration, motion theme.MotionTheme) float32 {
	if transition != TabsPanelFade || key == "" {
		s.panelKey = key
		s.panelOpacity.Reset()
		return 1
	}
	if s.panelKey != key {
		if s.panelKey == "" {
			s.panelOpacity.Initialize(1, gtx.Now)
		} else {
			s.panelOpacity.Set(0, 1, gtx.Now)
		}
		s.panelKey = key
	}
	return s.panelOpacity.Value(gtx, 1, duration, animation.EaseCubicOut, motion)
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

// tabsCloseFallback prefers the next enabled item and then the previous one.
// It returns false when no other enabled item remains.
func tabsCloseFallback(items []TabItem, closed int) (int, bool) {
	if closed < 0 || closed >= len(items) {
		return tabsFirstEnabled(items)
	}
	for index := closed + 1; index < len(items); index++ {
		if !items[index].Disabled {
			return index, true
		}
	}
	for index := closed - 1; index >= 0; index-- {
		if !items[index].Disabled {
			return index, true
		}
	}
	return -1, false
}
