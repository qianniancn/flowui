package table

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
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotTable = "table"

const (
	tableTypeaheadTimeout = 500 * time.Millisecond
	tableColorDuration    = 100 * time.Millisecond
)

type tableState struct {
	vertical       layout.List
	verticalBar    widget.Scrollbar
	horizontal     layout.List
	horizontalBar  widget.Scrollbar
	rows           map[string]*tableRowState
	columns        map[string]*tableColumnState
	frameRows      map[string]struct{}
	frameColumns   map[string]struct{}
	rowKeys        map[string]struct{}
	columnKeys     map[string]struct{}
	keyFilters     []event.Filter
	pressedKey     key.Name
	pressedRowKey  string
	typeahead      string
	typeaheadAt    time.Time
	typeaheadReady bool
	selectAll      widget.Clickable
	selectAllFocus state.FocusAnimation
}

func tableStateFor(ctx *frame.Context, key string) *tableState {
	key = frame.ClaimKey(ctx, state.KindTable, key)
	return frame.UseState[tableState](ctx, key, stateSlotTable)
}

func (s *tableState) beginFrame() {
	if s.frameRows == nil {
		s.frameRows = make(map[string]struct{})
		s.frameColumns = make(map[string]struct{})
	} else {
		clear(s.frameRows)
		clear(s.frameColumns)
	}
}

func (s *tableState) endFrame() {
	for key := range s.rows {
		if _, ok := s.frameRows[key]; !ok {
			delete(s.rows, key)
		}
	}
	for key := range s.columns {
		if _, ok := s.frameColumns[key]; !ok {
			delete(s.columns, key)
		}
	}
}

func (s *tableState) row(key string) *tableRowState {
	if s.rows == nil {
		s.rows = make(map[string]*tableRowState)
	}
	s.frameRows[key] = struct{}{}
	if value := s.rows[key]; value != nil {
		return value
	}
	value := new(tableRowState)
	s.rows[key] = value
	return value
}

func (s *tableState) column(key string) *tableColumnState {
	if s.columns == nil {
		s.columns = make(map[string]*tableColumnState)
	}
	s.frameColumns[key] = struct{}{}
	if value := s.columns[key]; value != nil {
		return value
	}
	value := new(tableColumnState)
	s.columns[key] = value
	return value
}

func (s *tableState) check(columns []Column, rows []Row) {
	if len(columns) == 0 {
		panic("flowui: table requires at least one column")
	}
	if s.columnKeys == nil {
		s.columnKeys = make(map[string]struct{}, len(columns))
		s.rowKeys = make(map[string]struct{}, len(rows))
	} else {
		clear(s.columnKeys)
		clear(s.rowKeys)
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
	for _, row := range rows {
		if row.Key == "" {
			panic("flowui: empty table row key")
		}
		if _, exists := s.rowKeys[row.Key]; exists {
			panic(fmt.Sprintf("flowui: duplicate table row key %q", row.Key))
		}
		if len(row.Cells) != len(columns) {
			panic(fmt.Sprintf("flowui: table row %q has %d cells, want %d", row.Key, len(row.Cells), len(columns)))
		}
		s.rowKeys[row.Key] = struct{}{}
	}
}

type tableKeyResult struct {
	focusKey  string
	actionKey string
}

func (s *tableState) updateKeys(gtx layout.Context, table Widget) tableKeyResult {
	s.keyFilters = s.keyFilters[:0]
	for _, row := range table.rows {
		tag := &s.row(row.Key).clickable
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
	current := s.focusedIndex(gtx, table.rows)
	if current < 0 {
		current = table.keyboardActiveIndex()
	}
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
		switch event.Name {
		case key.NameDownArrow:
			if event.State == key.Press {
				if next, ok := moveRow(table, current, 1); ok {
					current = next
					result.focusKey = table.rows[next].Key
				}
			}
		case key.NameUpArrow:
			if event.State == key.Press {
				if next, ok := moveRow(table, current, -1); ok {
					current = next
					result.focusKey = table.rows[next].Key
				}
			}
		case key.NameHome:
			if event.State == key.Press {
				if next, ok := firstEnabledRow(table); ok {
					current = next
					result.focusKey = table.rows[next].Key
				}
			}
		case key.NameEnd:
			if event.State == key.Press {
				if next, ok := lastEnabledRow(table); ok {
					current = next
					result.focusKey = table.rows[next].Key
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
				result.focusKey = table.rows[next].Key
			}
		}
	}
	return result
}

func (s *tableState) handleActivation(event key.Event, table Widget, current int, result *tableKeyResult) {
	switch event.State {
	case key.Press:
		s.pressedKey = event.Name
		s.pressedRowKey = ""
		if current >= 0 && current < len(table.rows) && !table.rowDisabled(table.rows[current]) {
			s.pressedRowKey = table.rows[current].Key
		}
	case key.Release:
		if s.pressedKey == event.Name && s.pressedRowKey != "" {
			result.actionKey = s.pressedRowKey
		}
		s.pressedKey = ""
		s.pressedRowKey = ""
	}
}

func (s *tableState) focusedIndex(gtx layout.Context, rows []Row) int {
	for index, row := range rows {
		if gtx.Focused(&s.row(row.Key).clickable) {
			return index
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
		return rowIndex(t.rows, t.selectedKey)
	}
	if t.selectionMode == SelectionMultiple {
		for _, key := range t.selectedKeys {
			if index := rowIndex(t.rows, key); index >= 0 {
				return index
			}
		}
	}
	return -1
}

func moveRow(table Widget, current, delta int) (int, bool) {
	if current < 0 || current >= len(table.rows) {
		if delta < 0 {
			return lastEnabledRow(table)
		}
		return firstEnabledRow(table)
	}
	for next := current + delta; next >= 0 && next < len(table.rows); next += delta {
		if !table.rowDisabled(table.rows[next]) {
			return next, true
		}
	}
	return current, false
}

func firstEnabledRow(table Widget) (int, bool) {
	for index, row := range table.rows {
		if !table.rowDisabled(row) {
			return index, true
		}
	}
	return -1, false
}

func lastEnabledRow(table Widget) (int, bool) {
	for index := len(table.rows) - 1; index >= 0; index-- {
		if !table.rowDisabled(table.rows[index]) {
			return index, true
		}
	}
	return -1, false
}

func typeaheadRow(table Widget, current int, query string) (int, bool) {
	if len(table.rows) == 0 || query == "" {
		return -1, false
	}
	query = strings.ToLower(query)
	for step := 1; step <= len(table.rows); step++ {
		index := (current + step + len(table.rows)) % len(table.rows)
		row := table.rows[index]
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
	clickable  widget.Clickable
	focus      state.FocusAnimation
	background tableColorAnimation
}

type tableColumnState struct {
	clickable widget.Clickable
	focus     state.FocusAnimation
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
