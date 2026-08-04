package table

import (
	"fmt"
	"image"
	"math"
	"slices"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/checkbox"
	"github.com/qianniancn/flowui/internal/components/nav"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/interact"
	"github.com/qianniancn/flowui/internal/state"
)

const stateSlotTable = "table"

const tableColorDuration = 100 * time.Millisecond

type tableState struct {
	vertical           layout.List
	verticalBar        widget.Scrollbar
	horizontal         layout.List
	horizontalBar      widget.Scrollbar
	rows               map[string]*tableRowState
	columns            map[string]*tableColumnState
	frameRows          map[string]struct{}
	frameColumns       map[string]struct{}
	rowKeys            map[string]struct{}
	columnKeys         map[string]struct{}
	keyFilters         []event.Filter
	selectedKeys       state.StringSetCache
	disabledKeys       state.StringSetCache
	pressedKey         key.Name
	pressedRowKey      string
	pressedModifiers   key.Modifiers
	selectionAnchor    string
	typeahead          nav.Typeahead
	selectAll          widget.Clickable
	selectAllFocus     state.FocusAnimation
	selectAllSelection checkbox.SelectionAnimation
	loadMoreCount      int
	loadMoreHasMore    bool
	loadMoreRequested  bool
}

func (s *tableState) updateLoadMore(count int, hasMore, loading, visible bool, onLoadMore func()) {
	if !hasMore {
		s.loadMoreCount = count
		s.loadMoreHasMore = false
		s.loadMoreRequested = false
		return
	}
	if count != s.loadMoreCount || !s.loadMoreHasMore {
		s.loadMoreRequested = false
	}
	if loading {
		s.loadMoreRequested = true
	}
	if visible && !loading && !s.loadMoreRequested && onLoadMore != nil {
		s.loadMoreRequested = true
		onLoadMore()
	}
	s.loadMoreCount = count
	s.loadMoreHasMore = hasMore
}

func tableStateFor(ctx *frame.Context, key string) *tableState {
	key = frame.ClaimKey(ctx, state.KindTable, key)
	return frame.UseState[tableState](ctx, key, stateSlotTable)
}

func (s *tableState) beginFrame() {
	state.BeginFrameMap(&s.frameRows)
	state.BeginFrameMap(&s.frameColumns)
}

func (s *tableState) endFrame() {
	state.SweepFrameMap(s.rows, s.frameRows)
	state.SweepFrameMap(s.columns, s.frameColumns)
}

func (s *tableState) row(key string) *tableRowState {
	if s.rows == nil {
		s.rows = make(map[string]*tableRowState)
	}
	s.frameRows[key] = struct{}{}
	if value := s.rows[key]; value != nil {
		return value
	}
	value := &tableRowState{index: -1}
	s.rows[key] = value
	return value
}

func (s *tableState) rowAt(key string, index int) *tableRowState {
	if _, exists := s.frameRows[key]; exists {
		if current := s.rows[key]; current != nil && current.index != index {
			panic(fmt.Sprintf("flowui: duplicate virtual table row key %q", key))
		}
	}
	value := s.row(key)
	value.index = index
	return value
}

func (s *tableState) column(key string) *tableColumnState {
	return state.UseFrameMap(&s.columns, &s.frameColumns, key)
}

func (s *tableState) check(columns []Column, rows []Row) {
	s.checkColumns(columns)
	if s.rowKeys == nil {
		s.rowKeys = make(map[string]struct{}, len(rows))
	} else {
		clear(s.rowKeys)
	}
	for _, row := range rows {
		validateRow(columns, row)
		if _, exists := s.rowKeys[row.Key]; exists {
			panic(fmt.Sprintf("flowui: duplicate table row key %q", row.Key))
		}
		s.rowKeys[row.Key] = struct{}{}
	}
	if _, ok := s.rowKeys[s.selectionAnchor]; !ok {
		s.selectionAnchor = ""
	}
}

func (s *tableState) checkColumns(columns []Column) {
	if len(columns) == 0 {
		panic("flowui: table requires at least one column")
	}
	if s.columnKeys == nil {
		s.columnKeys = make(map[string]struct{}, len(columns))
	} else {
		clear(s.columnKeys)
	}
	for _, column := range columns {
		if column.Key == "" {
			panic("flowui: empty table column key")
		}
		if _, exists := s.columnKeys[column.Key]; exists {
			panic(fmt.Sprintf("flowui: duplicate table column key %q", column.Key))
		}
		s.columnKeys[column.Key] = struct{}{}
	}
}

func validateRow(columns []Column, row Row) {
	if row.Key == "" {
		panic("flowui: empty table row key")
	}
	if len(row.Cells) != len(columns) {
		panic(fmt.Sprintf("flowui: table row %q has %d cells, want %d", row.Key, len(row.Cells), len(columns)))
	}
}

type tableKeyResult struct {
	focusKey  string
	actionKey string
	rangeKey  string
}

func (s *tableState) updateKeys(gtx layout.Context, table Widget) tableKeyResult {
	s.keyFilters = s.keyFilters[:0]
	for keyValue, rowState := range s.rows {
		if rowState.index < 0 || rowState.index >= table.count() {
			continue
		}
		row := table.row(rowState.index)
		if row.Key != keyValue {
			continue
		}
		tag := &rowState.clickable
		if table.rowDisabled(row) {
			s.keyFilters = append(s.keyFilters, rowState.keyFilters.Resolve(tag,
				key.NameDownArrow,
				key.NameUpArrow,
				key.NameHome,
				key.NameEnd,
			)...)
			continue
		}
		s.keyFilters = append(s.keyFilters, rowState.keyFilters.Resolve(tag,
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
	if len(s.keyFilters) == 0 {
		return tableKeyResult{}
	}
	current := s.focusedIndex(gtx, table)
	list := tableNavList(table)
	result := tableKeyResult{}
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
			current = table.keyboardActiveIndex()
		}
		switch event.Name {
		case key.NameDownArrow:
			if event.State == key.Press {
				if next, ok := nav.Move(list, current, 1, false); ok {
					current = next
					result.focusKey = table.row(next).Key
					if event.Modifiers.Contain(key.ModShift) && table.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameUpArrow:
			if event.State == key.Press {
				if next, ok := nav.Move(list, current, -1, false); ok {
					current = next
					result.focusKey = table.row(next).Key
					if event.Modifiers.Contain(key.ModShift) && table.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameHome:
			if event.State == key.Press {
				if next, ok := nav.First(list); ok {
					current = next
					result.focusKey = table.row(next).Key
					if event.Modifiers.Contain(key.ModShift) && table.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameEnd:
			if event.State == key.Press {
				if next, ok := nav.Last(list); ok {
					current = next
					result.focusKey = table.row(next).Key
					if event.Modifiers.Contain(key.ModShift) && table.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameEnter, key.NameReturn, key.NameSpace:
			s.handleActivation(event, table, current, &result)
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
				result.focusKey = table.row(next).Key
			}
		}
	}
	return result
}

func (s *tableState) handleActivation(event key.Event, table Widget, current int, result *tableKeyResult) {
	switch event.State {
	case key.Press:
		s.pressedKey = event.Name
		s.pressedModifiers = event.Modifiers
		s.pressedRowKey = ""
		if current >= 0 && current < table.count() && !table.rowDisabled(table.row(current)) {
			s.pressedRowKey = table.row(current).Key
		}
	case key.Release:
		if s.pressedKey == event.Name && s.pressedRowKey != "" {
			if s.pressedModifiers.Contain(key.ModShift) && table.selectionMode == SelectionMultiple {
				result.rangeKey = s.pressedRowKey
			} else {
				result.actionKey = s.pressedRowKey
			}
		}
		s.pressedKey = ""
		s.pressedRowKey = ""
		s.pressedModifiers = 0
	}
}

func (s *tableState) focusedIndex(gtx layout.Context, table Widget) int {
	for keyValue, rowState := range s.rows {
		if gtx.Focused(&rowState.clickable) {
			if rowState.index >= 0 && rowState.index < table.count() && table.row(rowState.index).Key == keyValue {
				return rowState.index
			}
			return table.rowIndex(keyValue)
		}
	}
	return -1
}

func (s *tableState) ensureVisible(index int) {
	if index < 0 || s.vertical.Position.Count == 0 {
		return
	}
	position := &s.vertical.Position
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

func (t Widget) keyboardActiveIndex() int {
	if t.selectionMode == SelectionSingle {
		return t.rowIndex(t.selectedKey)
	}
	if t.selectionMode == SelectionMultiple {
		for _, key := range t.selectedKeys {
			if index := t.rowIndex(key); index >= 0 {
				return index
			}
		}
	}
	return -1
}

// tableNavList adapts the table's rows into a nav.List. rowDisabled resolves in
// O(1) via the table's disabled-key set, so per-keystroke navigation does not
// linearly scan disabledKeys.
func tableNavList(table Widget) nav.List {
	return nav.List{
		Count:    table.count(),
		Disabled: func(i int) bool { return table.rowDisabled(table.row(i)) },
		Label:    func(i int) string { return table.rowLabel(table.row(i)) },
	}
}

func (t Widget) rowLabel(row Row) string {
	if row.Label != "" {
		return row.Label
	}
	for index, column := range t.columns {
		if column.RowHeader && index < len(row.Cells) {
			return row.Cells[index].Text
		}
	}
	if len(row.Cells) > 0 {
		return row.Cells[0].Text
	}
	return ""
}

type tableRowState struct {
	clickable        widget.Clickable
	focus            state.FocusAnimation
	keyFilters       state.KeyFilterCache
	background       animation.ColorTransition
	selection        checkbox.SelectionAnimation
	interactiveCells []image.Rectangle
	focusTargets     []event.Tag
	handledPress     time.Time
	index            int
}

func (s *tableRowState) clickInInteractiveCell() bool {
	history := s.clickable.History()
	if len(history) == 0 {
		return false
	}
	press := history[len(history)-1]
	if press.Start.IsZero() || !press.Start.After(s.handledPress) {
		return false
	}
	s.handledPress = press.Start
	return slices.ContainsFunc(s.interactiveCells, press.Position.In)
}

type tableColumnState struct {
	clickable widget.Clickable
	focus     state.FocusAnimation
	resize    tableColumnResizeState
}

type tableColumnResizeState struct {
	dragging        bool
	overridden      bool
	pointerID       pointer.ID
	startX          float32
	startWidth      int
	width           int
	configuredWidth int
	ready           bool
	focus           state.FocusAnimation
}

func (s *tableColumnResizeState) interactionWidth(configured int) (int, bool) {
	if configured != s.configuredWidth && !s.dragging && configured > 0 {
		return configured, true
	}
	if s.ready {
		return s.width, true
	}
	if configured > 0 {
		return configured, true
	}
	return 0, false
}

func (s *tableColumnResizeState) resolvedWidth(configured int) (int, bool) {
	if configured != s.configuredWidth && !s.dragging {
		return configured, configured > 0
	}
	if s.overridden {
		return s.width, true
	}
	return configured, configured > 0
}

func (s *tableColumnResizeState) sync(configured, resolved int) {
	if !s.ready {
		s.width = resolved
		s.configuredWidth = configured
		s.ready = true
		return
	}
	if configured != s.configuredWidth && !s.dragging {
		s.width = resolved
		s.configuredWidth = configured
		s.overridden = false
		return
	}
	if !s.overridden && !s.dragging {
		s.width = resolved
	}
}

func (s *tableColumnResizeState) update(ctx *frame.Context, gtx layout.Context, current, minimum, maximum, step int, enabled bool) (int, bool) {
	if !enabled {
		s.dragging = false
		return current, false
	}
	next := current
	changed := false
	for {
		eventValue, ok := interact.NextPointerEvent(gtx, s, pointer.Press|pointer.Drag|pointer.Release|pointer.Cancel)
		if !ok {
			break
		}
		switch eventValue.Kind {
		case pointer.Press:
			if !interact.IsPrimaryPointerPress(eventValue) {
				continue
			}
			s.dragging = true
			s.pointerID = eventValue.PointerID
			s.startX = eventValue.Position.X
			s.startWidth = current
			frame.RequestFocusVisible(ctx, s, false)
			interact.GrabPointer(gtx, s, eventValue)
		case pointer.Drag:
			if !s.dragging || eventValue.PointerID != s.pointerID {
				continue
			}
			next = min(max(s.startWidth+int(math.Round(float64(eventValue.Position.X-s.startX))), minimum), maximum)
			changed = next != current
		case pointer.Release:
			if eventValue.PointerID == s.pointerID {
				s.dragging = false
			}
		case pointer.Cancel:
			if eventValue.PointerID == s.pointerID {
				s.dragging = false
			}
		}
	}
	for {
		e, ok := gtx.Event(
			key.Filter{Focus: s, Name: key.NameLeftArrow},
			key.Filter{Focus: s, Name: key.NameRightArrow},
			key.Filter{Focus: s, Name: key.NameHome},
			key.Filter{Focus: s, Name: key.NameEnd},
		)
		if !ok {
			break
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		next = current
		switch event.Name {
		case key.NameLeftArrow:
			next -= step
		case key.NameRightArrow:
			next += step
		case key.NameHome:
			next = minimum
		case key.NameEnd:
			next = maximum
		}
		next = min(max(next, minimum), maximum)
		if next != current {
			changed = true
		}
	}
	if changed {
		s.width = next
		s.ready = true
		s.overridden = true
	}
	return next, changed
}
