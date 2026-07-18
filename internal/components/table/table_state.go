package table

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"time"
	"unicode"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/checkbox"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotTable = "table"

const (
	tableTypeaheadTimeout = 500 * time.Millisecond
	tableColorDuration    = 100 * time.Millisecond
)

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
	typeahead          string
	typeaheadAt        time.Time
	typeaheadReady     bool
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
		s.keyFilters = append(s.keyFilters,
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameHome},
			key.Filter{Focus: tag, Name: key.NameEnd},
		)
		if table.rowDisabled(row) {
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
		return tableKeyResult{}
	}
	current := s.focusedIndex(gtx, table)
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
				if next, ok := moveRow(table, current, 1); ok {
					current = next
					result.focusKey = table.row(next).Key
					if event.Modifiers.Contain(key.ModShift) && table.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameUpArrow:
			if event.State == key.Press {
				if next, ok := moveRow(table, current, -1); ok {
					current = next
					result.focusKey = table.row(next).Key
					if event.Modifiers.Contain(key.ModShift) && table.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameHome:
			if event.State == key.Press {
				if next, ok := firstEnabledRow(table); ok {
					current = next
					result.focusKey = table.row(next).Key
					if event.Modifiers.Contain(key.ModShift) && table.selectionMode == SelectionMultiple {
						result.rangeKey = result.focusKey
					}
				}
			}
		case key.NameEnd:
			if event.State == key.Press {
				if next, ok := lastEnabledRow(table); ok {
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
			text := tableTypeaheadText(event.Name)
			if text == "" {
				continue
			}
			query := s.appendTypeahead(gtx.Now, text)
			next, ok := typeaheadRow(table, current, query)
			if !ok && query != text {
				s.typeahead = text
				next, ok = typeaheadRow(table, current, text)
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

func (s *tableState) appendTypeahead(now time.Time, value string) string {
	if !s.typeaheadReady || now.Before(s.typeaheadAt) || now.Sub(s.typeaheadAt) > tableTypeaheadTimeout {
		s.typeahead = ""
	}
	s.typeahead += value
	s.typeaheadAt = now
	s.typeaheadReady = true
	return s.typeahead
}

func tableTypeaheadText(name key.Name) string {
	runes := []rune(string(name))
	if len(runes) != 1 || unicode.IsControl(runes[0]) {
		return ""
	}
	return strings.ToLower(string(runes[0]))
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

func moveRow(table Widget, current, delta int) (int, bool) {
	if current < 0 || current >= table.count() {
		if delta < 0 {
			return lastEnabledRow(table)
		}
		return firstEnabledRow(table)
	}
	for next := current + delta; next >= 0 && next < table.count(); next += delta {
		if !table.rowDisabled(table.row(next)) {
			return next, true
		}
	}
	return current, false
}

func firstEnabledRow(table Widget) (int, bool) {
	for index := range table.count() {
		row := table.row(index)
		if !table.rowDisabled(row) {
			return index, true
		}
	}
	return -1, false
}

func lastEnabledRow(table Widget) (int, bool) {
	for index := table.count() - 1; index >= 0; index-- {
		if !table.rowDisabled(table.row(index)) {
			return index, true
		}
	}
	return -1, false
}

func typeaheadRow(table Widget, current int, query string) (int, bool) {
	if table.count() == 0 || query == "" {
		return -1, false
	}
	query = strings.ToLower(query)
	for step := 1; step <= table.count(); step++ {
		index := (current + step + table.count()) % table.count()
		row := table.row(index)
		if table.rowDisabled(row) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(table.rowLabel(row)), query) {
			return index, true
		}
	}
	return current, false
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
	background       tableColorAnimation
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
	for _, cell := range s.interactiveCells {
		if press.Position.In(cell) {
			return true
		}
	}
	return false
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
		e, ok := gtx.Event(pointer.Filter{Target: s, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
		if !ok {
			break
		}
		event, ok := e.(pointer.Event)
		if !ok {
			continue
		}
		switch event.Kind {
		case pointer.Press:
			if event.Source != pointer.Touch && !event.Buttons.Contain(pointer.ButtonPrimary) {
				continue
			}
			s.dragging = true
			s.pointerID = event.PointerID
			s.startX = event.Position.X
			s.startWidth = current
			frame.RequestFocusVisible(ctx, s, false)
			gtx.Execute(pointer.GrabCmd{Tag: s, ID: event.PointerID})
		case pointer.Drag:
			if !s.dragging || event.PointerID != s.pointerID {
				continue
			}
			next = min(max(s.startWidth+int(math.Round(float64(event.Position.X-s.startX))), minimum), maximum)
			changed = next != current
		case pointer.Release:
			if event.PointerID == s.pointerID {
				s.dragging = false
			}
		case pointer.Cancel:
			if event.PointerID == s.pointerID {
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

type tableColorAnimation struct {
	value color.NRGBA
	from  color.NRGBA
	to    color.NRGBA
	at    time.Time
	ready bool
}

func (a *tableColorAnimation) update(gtx layout.Context, target color.NRGBA) color.NRGBA {
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
	progress := render.Ease(render.Progress(gtx.Now.Sub(a.at), tableColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	a.value = render.LerpColor(a.from, a.to, progress)
	return a.value
}
