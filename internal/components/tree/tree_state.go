package tree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotTree = "tree"

const treeDragMIMEPrefix = "application/x-flowui-tree-item/"

const (
	treeTypeaheadTimeout = 500 * time.Millisecond
	treeExpandDuration   = 200 * time.Millisecond
	treeDragExpandDelay  = 600 * time.Millisecond
	treeDragScrollSpeed  = 8
)

type flatItem struct {
	item          Item
	depth         int
	parentKey     string
	isLast        bool
	ancestorsLast []bool
}

func flattenVisibleItems(items []Item, expanded map[string]struct{}) []flatItem {
	result := make([]flatItem, 0)
	var walk func([]Item, int, string, []bool)
	walk = func(children []Item, depth int, parent string, ancestorsLast []bool) {
		for index, item := range children {
			isLast := index == len(children)-1
			result = append(result, flatItem{
				item:          item,
				depth:         depth,
				parentKey:     parent,
				isLast:        isLast,
				ancestorsLast: append([]bool(nil), ancestorsLast...),
			})
			if _, ok := expanded[item.Key]; ok && len(item.Children) > 0 {
				walk(item.Children, depth+1, item.Key, append(append([]bool(nil), ancestorsLast...), isLast))
			}
		}
	}
	walk(items, 0, "", nil)
	return result
}

type treeState struct {
	list               layout.List
	bar                widget.Scrollbar
	items              map[string]*treeItemState
	frameItems         map[string]struct{}
	itemKeys           map[string]struct{}
	keyFilters         []event.Filter
	pressedKey         key.Name
	pressedModifiers   key.Modifiers
	pressedActionKey   string
	selectionAnchor    string
	typeahead          string
	typeaheadAt        time.Time
	typeaheadReady     bool
	dragMIME           string
	dragSource         string
	dragSources        []string
	dropTarget         treeDropTarget
	dragHoverKey       string
	dragHoverAt        time.Time
	dragHoverRequested bool
	dragScrollAt       time.Time
	renameKey          string
	renameOriginal     string
	renameEditor       widget.Editor
	renameFocused      bool
	renameRequestKey   string
	renameRequest      uint64
	renameRequestReady bool
	dataCache          treeDataCache
	selectedKeys       state.StringSetCache
	disabledKeys       state.StringSetCache
}

type treeDataCache struct {
	ready        bool
	version      uint64
	expandedKeys []string
	visible      []flatItem
}

func treeStateFor(ctx *frame.Context, key string) *treeState {
	key = frame.ClaimKey(ctx, state.KindTree, key)
	value := frame.UseState[treeState](ctx, key, stateSlotTree)
	value.dragMIME = treeDragMIMEPrefix + key
	return value
}

func (s *treeState) beginFrame() {
	state.BeginFrameMap(&s.frameItems)
}

func (s *treeState) endFrame() {
	state.SweepFrameMap(s.items, s.frameItems)
}

func (s *treeState) item(key string) *treeItemState {
	return state.UseFrameMap(&s.items, &s.frameItems, key)
}

func (s *treeState) resolveVisible(tree Widget) []flatItem {
	if !tree.hasDataVersion {
		s.dataCache = treeDataCache{}
		s.checkItems(tree.items)
		return flattenVisibleItems(tree.items, treeKeySet(tree.expandedKeys))
	}
	cache := &s.dataCache
	if cache.ready && cache.version == tree.dataVersion && treeKeysEqual(cache.expandedKeys, tree.expandedKeys) {
		return cache.visible
	}
	s.checkItems(tree.items)
	cache.visible = flattenVisibleItems(tree.items, treeKeySet(tree.expandedKeys))
	cache.expandedKeys = append(cache.expandedKeys[:0], tree.expandedKeys...)
	cache.version = tree.dataVersion
	cache.ready = true
	return cache.visible
}

func (s *treeState) visibleIndex(visible []flatItem, key string, item *treeItemState) int {
	if item.index >= 0 && item.index < len(visible) && visible[item.index].item.Key == key {
		return item.index
	}
	item.index = treeVisibleIndex(visible, key)
	return item.index
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
	focusKey        string
	actionKey       string
	actionModifiers key.Modifiers
	rangeKey        string
	toggleKey       string
	loadKey         string
	renameKey       string
}

func (s *treeState) updateKeys(gtx layout.Context, tree Widget, visible []flatItem) treeKeyResult {
	s.keyFilters = s.keyFilters[:0]
	for keyValue, itemState := range s.items {
		index := s.visibleIndex(visible, keyValue, itemState)
		if index < 0 {
			continue
		}
		entry := visible[index]
		tag := &itemState.clickable
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
		if tree.onRename != nil && entry.item.Renamable {
			s.keyFilters = append(s.keyFilters, key.Filter{Focus: tag, Name: key.NameF2})
		}
	}
	if len(s.keyFilters) == 0 {
		return treeKeyResult{}
	}

	current := s.focusedIndex(gtx, visible)
	result := treeKeyResult{}
	var expanded map[string]struct{}
	for {
		e, ok := gtx.Event(s.keyFilters...)
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if !ok {
			continue
		}
		if current < 0 {
			current = tree.keyboardActiveIndex(visible)
		}
		switch event.Name {
		case key.NameDownArrow:
			if event.State == key.Press {
				if next, ok := treeMoveVisible(visible, tree, current, 1); ok {
					current = next
					result.focusKey = visible[next].item.Key
					if event.Modifiers.Contain(key.ModShift) && tree.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameUpArrow:
			if event.State == key.Press {
				if next, ok := treeMoveVisible(visible, tree, current, -1); ok {
					current = next
					result.focusKey = visible[next].item.Key
					if event.Modifiers.Contain(key.ModShift) && tree.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameHome:
			if event.State == key.Press {
				if next, ok := treeFirstEnabled(visible, tree); ok {
					current = next
					result.focusKey = visible[next].item.Key
					if event.Modifiers.Contain(key.ModShift) && tree.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameEnd:
			if event.State == key.Press {
				if next, ok := treeLastEnabled(visible, tree); ok {
					current = next
					result.focusKey = visible[next].item.Key
					if event.Modifiers.Contain(key.ModShift) && tree.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameRightArrow:
			if event.State == key.Press && current >= 0 && current < len(visible) {
				if expanded == nil {
					expanded = treeKeySet(tree.expandedKeys)
				}
				entry := visible[current]
				if treeItemExpandable(entry.item) {
					if _, ok := expanded[entry.item.Key]; !ok {
						result.toggleKey = entry.item.Key
						if entry.item.ChildrenState == ChildrenUnloaded || entry.item.ChildrenState == ChildrenError {
							result.loadKey = entry.item.Key
						}
					} else if entry.item.ChildrenState == ChildrenError {
						result.loadKey = entry.item.Key
					} else if child := treeFirstEnabledChild(visible, tree, current); child >= 0 {
						current = child
						result.focusKey = visible[child].item.Key
					}
				}
			}
		case key.NameLeftArrow:
			if event.State == key.Press && current >= 0 && current < len(visible) {
				if expanded == nil {
					expanded = treeKeySet(tree.expandedKeys)
				}
				entry := visible[current]
				if treeItemExpandable(entry.item) {
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
		case key.NameF2:
			if event.State == key.Press && current >= 0 && current < len(visible) {
				entry := visible[current]
				if entry.item.Renamable && tree.onRename != nil && !tree.itemDisabled(entry.item) {
					result.renameKey = entry.item.Key
				}
			}
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
		s.pressedModifiers = event.Modifiers
		s.pressedActionKey = ""
		if current >= 0 && current < len(visible) && !tree.itemDisabled(visible[current].item) {
			s.pressedActionKey = visible[current].item.Key
		}
	case key.Release:
		if s.pressedKey == event.Name && s.pressedActionKey != "" {
			if s.pressedModifiers.Contain(key.ModShift) && tree.selectionMode == SelectionMultiple {
				result.rangeKey = s.pressedActionKey
			} else {
				result.actionKey = s.pressedActionKey
				result.actionModifiers = s.pressedModifiers
			}
		}
		s.pressedKey = ""
		s.pressedModifiers = 0
		s.pressedActionKey = ""
	}
}

func (s *treeState) focusedIndex(gtx layout.Context, visible []flatItem) int {
	for keyValue, itemState := range s.items {
		index := s.visibleIndex(visible, keyValue, itemState)
		if index >= 0 && gtx.Focused(&itemState.clickable) {
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
	index     int
	clickable widget.Clickable
	toggle    overlay.ClickArea
	focus     state.FocusAnimation
	expansion treeFloatAnimation
	drag      widget.Draggable
	dragPress f32.Point
	dragTag   byte
	dropTags  [3]byte
}

type treeDropTarget struct {
	key      string
	drawKey  string
	depth    int
	position DropPosition
}

func (s *treeItemState) updateDrag(gtx layout.Context, mime string, sourceKeys []string) bool {
	for {
		raw, ok := gtx.Event(pointer.Filter{Target: &s.dragTag, Kinds: pointer.Press})
		if !ok {
			break
		}
		if event, ok := raw.(pointer.Event); ok && (event.Source == pointer.Touch || event.Buttons.Contain(pointer.ButtonPrimary)) {
			s.dragPress = event.Position
		}
	}
	s.drag.Type = mime
	if requestedMIME, requested := s.drag.Update(gtx); requested {
		data, _ := json.Marshal(sourceKeys)
		s.drag.Offer(gtx, requestedMIME, io.NopCloser(bytes.NewReader(data)))
	}
	position := s.drag.Pos()
	slop := float32(gtx.Dp(3))
	return s.drag.Dragging() && position.X*position.X+position.Y*position.Y > slop*slop
}

func (t Widget) updateDropEvents(gtx layout.Context, state *treeState, itemState *treeItemState, visible []flatItem, entry flatItem) {
	for index := range itemState.dropTags {
		for {
			raw, ok := gtx.Event(transfer.TargetFilter{Target: &itemState.dropTags[index], Type: state.dragMIME})
			if !ok {
				break
			}
			event, ok := raw.(transfer.DataEvent)
			if !ok {
				continue
			}
			sourceKeys, ok := treeDropSource(event)
			position := DropPosition(index)
			if ok && t.dropAllowed(visible, sourceKeys, entry.item.Key, position) {
				t.onDrop(treeDropEvent(sourceKeys, entry.item.Key, position))
			}
		}
	}
}

func treeDropSource(event transfer.DataEvent) ([]string, bool) {
	reader := event.Open()
	if reader == nil {
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(reader, 64*1024+1))
	_ = reader.Close()
	if err != nil || len(data) == 0 || len(data) > 64*1024 {
		return nil, false
	}
	var keys []string
	if json.Unmarshal(data, &keys) != nil {
		return nil, false
	}
	unique := keys[:0]
	for _, itemKey := range keys {
		if itemKey != "" && !treeContainsKey(unique, itemKey) {
			unique = append(unique, itemKey)
		}
	}
	return unique, len(unique) > 0
}

func treeDropAllowed(tree Widget, visible []flatItem, sourceKeys []string, targetKey string, position DropPosition) bool {
	targetIndex := treeVisibleIndex(visible, targetKey)
	if len(sourceKeys) == 0 || targetIndex < 0 {
		return false
	}
	if tree.itemDisabled(visible[targetIndex].item) {
		return false
	}
	if position == DropInside && !treeItemAcceptsChildren(visible[targetIndex].item) {
		return false
	}
	for _, sourceKey := range sourceKeys {
		sourceIndex := treeVisibleIndex(visible, sourceKey)
		if sourceIndex < 0 || sourceIndex == targetIndex || tree.itemDisabled(visible[sourceIndex].item) {
			return false
		}
		if targetIndex <= sourceIndex {
			continue
		}
		sourceDepth := visible[sourceIndex].depth
		outsideSource := false
		for index := sourceIndex + 1; index <= targetIndex; index++ {
			if visible[index].depth <= sourceDepth {
				outsideSource = true
				break
			}
		}
		if !outsideSource {
			return false
		}
	}
	return true
}

func (t Widget) dropAllowed(visible []flatItem, sourceKeys []string, targetKey string, position DropPosition) bool {
	if !treeDropAllowed(t, visible, sourceKeys, targetKey, position) {
		return false
	}
	return t.canDrop == nil || t.canDrop(treeDropEvent(sourceKeys, targetKey, position))
}

func treeDropEvent(sourceKeys []string, targetKey string, position DropPosition) DropEvent {
	keys := append([]string(nil), sourceKeys...)
	return DropEvent{SourceKey: keys[0], SourceKeys: keys, TargetKey: targetKey, Position: position}
}

func treeDropTargetAt(visible []flatItem, sourceIndex int, pointerY float32, heights []int, gap int) treeDropTarget {
	if sourceIndex < 0 || sourceIndex >= len(visible) || len(heights) != len(visible) {
		return treeDropTarget{}
	}
	top := float32(0)
	for index := sourceIndex - 1; index >= 0; index-- {
		top -= float32(heights[index] + gap)
	}
	for index, entry := range visible {
		height := heights[index]
		bottom := top + float32(height)
		if pointerY >= top && pointerY < bottom {
			return treeDropTarget{
				key: entry.item.Key, drawKey: entry.item.Key, depth: entry.depth,
				position: treeDropPositionAt(pointerY-top, height, treeItemAcceptsChildren(entry.item)),
			}
		}
		top = bottom + float32(gap)
	}
	return treeDropTarget{}
}

func treeDropIndicatorTarget(visible []flatItem, target treeDropTarget) treeDropTarget {
	if target.position != DropAfter {
		return target
	}
	index := treeVisibleIndex(visible, target.key)
	if index < 0 {
		return treeDropTarget{}
	}
	for next := index + 1; next < len(visible) && visible[next].depth > target.depth; next++ {
		target.drawKey = visible[next].item.Key
	}
	return target
}

func treeDropPositionAt(localY float32, height int, acceptsChildren bool) DropPosition {
	if !acceptsChildren {
		if localY < float32(height)/2 {
			return DropBefore
		}
		return DropAfter
	}
	edge := float32(max(height/4, 1))
	if localY < edge {
		return DropBefore
	}
	if localY >= float32(height)-edge {
		return DropAfter
	}
	return DropInside
}

func treeItemAcceptsChildren(item Item) bool {
	return item.AcceptsChildren || treeItemExpandable(item)
}

func treeItemExpandable(item Item) bool {
	return len(item.Children) > 0 || item.ChildrenState != ChildrenLoaded
}

func treeDragViewportY(heights []int, gap int, position layout.Position, sourceIndex int, sourceY float32) float32 {
	top := float32(-position.Offset)
	if sourceIndex >= position.First {
		for index := position.First; index < sourceIndex && index < len(heights); index++ {
			top += float32(heights[index] + gap)
		}
	} else {
		for index := position.First - 1; index >= sourceIndex && index >= 0; index-- {
			top -= float32(heights[index] + gap)
		}
	}
	return top + sourceY
}

func treeListPositionContains(position layout.Position, index int) bool {
	return index >= position.First && index < position.First+position.Count
}

func treeAutoScrollDirection(pointerY float32, viewportHeight, edge int) int {
	if viewportHeight <= 0 || edge <= 0 {
		return 0
	}
	edge = min(edge, viewportHeight/2)
	if pointerY < float32(edge) {
		return -1
	}
	if pointerY > float32(viewportHeight-edge) {
		return 1
	}
	return 0
}

func (s *treeState) updateDragScroll(gtx layout.Context, pointerY float32, viewportHeight, edge, itemCount int) {
	direction := treeAutoScrollDirection(pointerY, viewportHeight, edge)
	if direction == 0 || s.list.Position.Count >= itemCount {
		s.dragScrollAt = time.Time{}
		return
	}
	delta := time.Second / 60
	if !s.dragScrollAt.IsZero() && !gtx.Now.Before(s.dragScrollAt) {
		delta = min(gtx.Now.Sub(s.dragScrollAt), 50*time.Millisecond)
	}
	s.dragScrollAt = gtx.Now
	s.list.ScrollBy(float32(direction) * treeDragScrollSpeed * float32(delta) / float32(time.Second))
	gtx.Execute(op.InvalidateCmd{})
}

func (s *treeState) updateDragHover(gtx layout.Context, key string, eligible bool) bool {
	if !eligible || key == "" {
		s.dragHoverKey = ""
		s.dragHoverAt = time.Time{}
		s.dragHoverRequested = false
		return false
	}
	if s.dragHoverKey != key {
		s.dragHoverKey = key
		s.dragHoverAt = gtx.Now
		s.dragHoverRequested = false
	}
	if s.dragHoverRequested {
		return false
	}
	deadline := s.dragHoverAt.Add(treeDragExpandDelay)
	if gtx.Now.Before(deadline) {
		gtx.Execute(op.InvalidateCmd{At: deadline})
		return false
	}
	s.dragHoverRequested = true
	return true
}

func (s *treeState) resetDragAssist() {
	s.dragScrollAt = time.Time{}
	s.dragHoverKey = ""
	s.dragHoverAt = time.Time{}
	s.dragHoverRequested = false
}

func (s *treeState) beginRename(item Item) {
	s.renameKey = item.Key
	s.renameOriginal = item.Label
	s.renameFocused = false
	s.renameEditor.SingleLine = true
	s.renameEditor.Submit = true
	s.renameEditor.SetText(item.Label)
	s.renameEditor.SetCaret(0, len([]rune(item.Label)))
}

func (s *treeState) applyRenameRequest(ctx *frame.Context, tree Widget, visible []flatItem) {
	if tree.renameRequestKey == "" {
		s.renameRequestReady = false
		return
	}
	if s.renameRequestReady && s.renameRequestKey == tree.renameRequestKey && s.renameRequest == tree.renameRequest {
		return
	}
	index := treeVisibleIndex(visible, tree.renameRequestKey)
	if index < 0 {
		return
	}
	item := visible[index].item
	if !item.Renamable || tree.onRename == nil || tree.itemDisabled(item) {
		return
	}
	s.renameRequestKey = tree.renameRequestKey
	s.renameRequest = tree.renameRequest
	s.renameRequestReady = true
	s.beginRename(item)
	frame.RequestFocus(ctx, &s.renameEditor)
}

func (s *treeState) updateRename(gtx layout.Context, tree Widget) {
	if s.renameKey == "" {
		return
	}
	for {
		eventValue, ok := s.renameEditor.Update(gtx)
		if !ok {
			break
		}
		if _, ok := eventValue.(widget.SubmitEvent); ok {
			s.finishRename(tree, true)
			return
		}
	}
	for {
		eventValue, ok := gtx.Event(key.Filter{Focus: &s.renameEditor, Name: key.NameEscape})
		if !ok {
			break
		}
		if eventValue, ok := eventValue.(key.Event); ok && eventValue.State == key.Press {
			s.finishRename(tree, false)
			return
		}
	}
	focused := gtx.Focused(&s.renameEditor)
	if focused {
		s.renameFocused = true
	} else if s.renameFocused {
		s.finishRename(tree, true)
	}
}

func (s *treeState) finishRename(tree Widget, commit bool) {
	itemKey := s.renameKey
	original := s.renameOriginal
	value := strings.TrimSpace(s.renameEditor.Text())
	s.renameKey = ""
	s.renameOriginal = ""
	s.renameFocused = false
	if commit && value != "" && value != original && tree.onRename != nil {
		tree.onRename(itemKey, value)
	}
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

func treeKeysEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
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
