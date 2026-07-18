package table

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/menu"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
)

// Variant selects the Table container treatment.
type Variant uint8

const (
	VariantPrimary Variant = iota
	VariantSecondary
)

// SelectionMode controls row selection behavior.
type SelectionMode uint8

const (
	SelectionNone SelectionMode = iota
	SelectionSingle
	SelectionMultiple
)

// SortDirection is the controlled direction of a sorted column.
type SortDirection uint8

const (
	SortAscending SortDirection = iota
	SortDescending
)

// Alignment controls content alignment within a column.
type Alignment uint8

const (
	AlignStart Alignment = iota
	AlignCenter
	AlignEnd
)

// SortDescriptor identifies the controlled sorted column and direction.
type SortDescriptor struct {
	Column    string
	Direction SortDirection
}

// Column describes a Table column. Width and MinWidth are expressed in dp.
type Column struct {
	Key       string
	Label     string
	Header    frame.Widget
	Width     int
	MinWidth  int
	MaxWidth  int
	Weight    float32
	Align     Alignment
	Sortable  bool
	Resizable bool
	RowHeader bool
}

// Cell contains either plain text or a custom widget. Content takes precedence.
type Cell struct {
	Text        string
	Content     frame.Widget
	Interactive bool
}

// Row describes one controlled Table row.
type Row struct {
	Key      string
	Label    string
	Cells    []Cell
	Disabled bool
}

// RowProvider returns a row by zero-based index for a virtual Table.
type RowProvider func(index int) Row

// RowContextMenu returns the menu shown for a secondary click, long press, or
// Shift+F10 anywhere on a row.
type RowContextMenu func(Row) menu.Widget

// Widget presents structured data with controlled sorting and selection.
type Widget struct {
	key                string
	columns            []Column
	rows               []Row
	rowCount           int
	rowProvider        RowProvider
	virtual            bool
	variant            Variant
	selectionMode      SelectionMode
	selectedKey        string
	selectedKeys       []string
	selectedKeySet     stateutil.StringSet
	sort               SortDescriptor
	disabledKeys       []string
	disabledKeySet     stateutil.StringSet
	emptyText          string
	emptyContent       frame.Widget
	footer             frame.Widget
	loadMoreContent    frame.Widget
	onChange           func(string)
	onSelectionChange  func([]string)
	onSortChange       func(SortDescriptor)
	onAction           func(string)
	onColumnResize     func(string, int)
	onLoadMore         func()
	rowContextMenu     RowContextMenu
	disabled           bool
	allowEmpty         bool
	selectionIndicator bool
	hasMore            bool
	loadingMore        bool
	maxHeight          int
	minWidth           int
	headerHeight       int
	rowHeight          int
	gridLines          bool
	gridLinesSet       bool
	bordered           bool
}

// New creates a controlled Table.
func New(key string, columns []Column, rows []Row) Widget {
	return Widget{
		key:           key,
		columns:       columns,
		rows:          rows,
		emptyText:     "No results",
		selectionMode: SelectionNone,
	}
}

// NewVirtual creates a Table that requests only rows visible in the viewport.
// The provider must return deterministic rows with stable, unique keys. Supply
// disabled virtual row keys through DisabledKeys so select-all state stays
// virtualized without scanning the provider.
func NewVirtual(key string, columns []Column, count int, provider RowProvider) Widget {
	if count < 0 {
		panic("flowui: virtual table row count must not be negative")
	}
	if count > 0 && provider == nil {
		panic("flowui: virtual table requires a row provider")
	}
	return Widget{
		key:           key,
		columns:       columns,
		rowCount:      count,
		rowProvider:   provider,
		virtual:       true,
		emptyText:     "No results",
		selectionMode: SelectionNone,
	}
}

func (t Widget) Variant(variant Variant) Widget {
	t.variant = variant
	return t
}

func (t Widget) SelectionMode(mode SelectionMode) Widget {
	t.selectionMode = mode
	return t
}

func (t Widget) SelectedKey(key string) Widget {
	t.selectedKey = key
	return t
}

func (t Widget) SelectedKeys(keys []string) Widget {
	t.selectedKeys = keys
	return t
}

func (t Widget) SortDescriptor(sort SortDescriptor) Widget {
	t.sort = sort
	return t
}

func (t Widget) DisabledKeys(keys []string) Widget {
	t.disabledKeys = keys
	return t
}

func (t Widget) EmptyText(text string) Widget {
	t.emptyText = text
	return t
}

func (t Widget) EmptyContent(content frame.Widget) Widget {
	t.emptyContent = content
	return t
}

func (t Widget) Footer(footer frame.Widget) Widget {
	t.footer = footer
	return t
}

// LoadMore enables an end-of-list sentinel that requests more rows when visible.
func (t Widget) LoadMore(hasMore, loading bool, fn func()) Widget {
	t.hasMore = hasMore
	t.loadingMore = loading
	t.onLoadMore = fn
	return t
}

// LoadMoreContent replaces the default loading spinner.
func (t Widget) LoadMoreContent(content frame.Widget) Widget {
	t.loadMoreContent = content
	return t
}

func (t Widget) OnChange(fn func(string)) Widget {
	t.onChange = fn
	return t
}

func (t Widget) OnSelectionChange(fn func([]string)) Widget {
	t.onSelectionChange = fn
	return t
}

func (t Widget) OnSortChange(fn func(SortDescriptor)) Widget {
	t.onSortChange = fn
	return t
}

func (t Widget) OnAction(fn func(string)) Widget {
	t.onAction = fn
	return t
}

// OnColumnResize reports the current column width while a resizer is moved.
func (t Widget) OnColumnResize(fn func(string, int)) Widget {
	t.onColumnResize = fn
	return t
}

// RowContextMenu sets the menu available from the complete row area.
func (t Widget) RowContextMenu(menu RowContextMenu) Widget {
	t.rowContextMenu = menu
	return t
}

func (t Widget) Disabled(disabled bool) Widget {
	t.disabled = disabled
	return t
}

func (t Widget) AllowEmptySelection() Widget {
	t.allowEmpty = true
	return t
}

func (t Widget) ShowSelectionIndicator() Widget {
	t.selectionIndicator = true
	return t
}

// MaxHeight limits the header and scrollable body height in dp.
func (t Widget) MaxHeight(dp int) Widget {
	t.maxHeight = max(dp, 0)
	return t
}

// MinWidth sets the minimum content width in dp and enables horizontal overflow.
func (t Widget) MinWidth(dp int) Widget {
	t.minWidth = max(dp, 0)
	return t
}

// HeaderHeight overrides the theme header height in dp.
func (t Widget) HeaderHeight(dp int) Widget {
	t.headerHeight = max(dp, 0)
	return t
}

// RowHeight overrides the minimum row height in dp.
func (t Widget) RowHeight(dp int) Widget {
	t.rowHeight = max(dp, 0)
	return t
}

// GridLines controls internal table lines. When enabled, horizontal and
// vertical lines form a complete cell grid. The default keeps the standard
// Table separators for compatibility.
func (t Widget) GridLines(visible bool) Widget {
	t.gridLines = visible
	t.gridLinesSet = true
	return t
}

// Border controls the outer table border.
func (t Widget) Border(visible bool) Widget {
	t.bordered = visible
	return t
}

func (t Widget) showsGridLines() bool {
	return !t.gridLinesSet || t.gridLines
}

func (t Widget) showsFullGrid() bool {
	return t.gridLinesSet && t.gridLines
}

func (t Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := tableStateFor(ctx, t.key)
	state.beginFrame()
	defer state.endFrame()
	t.selectedKeySet = state.selectedKeys.Resolve(t.selectedKeys)
	t.disabledKeySet = state.disabledKeys.Resolve(t.disabledKeys)
	if t.virtual {
		state.checkColumns(t.columns)
	} else {
		state.check(t.columns, t.rows)
	}
	t.updateColumnResizers(ctx, gtx, state)

	if !t.disabled {
		for _, column := range t.columns {
			if !column.Sortable {
				continue
			}
			for state.column(column.Key).clickable.Clicked(gtx) {
				t.requestSort(column.Key)
			}
		}
		if t.showsSelectionIndicator() && t.selectionMode == SelectionMultiple {
			for state.selectAll.Clicked(gtx) {
				t.toggleAll()
			}
		}
		t.updateRowClicks(gtx, state)
		result := state.updateKeys(gtx, t)
		if result.focusKey != "" {
			index := t.rowIndex(result.focusKey)
			if result.rangeKey == "" && t.selectionMode == SelectionMultiple {
				state.selectionAnchor = result.focusKey
			}
			frame.RequestFocus(ctx, &state.rowAt(result.focusKey, index).clickable)
			state.ensureVisible(index)
		}
		if result.rangeKey != "" {
			t.activateWithModifiers(state, result.rangeKey, key.ModShift)
		}
		if result.actionKey != "" {
			t.activateWithModifiers(state, result.actionKey, 0)
		}
	}
	if t.disabled {
		gtx = gtx.Disabled()
	}
	return t.layout(ctx, gtx, state)
}

func (t Widget) updateColumnResizers(ctx *frame.Context, gtx layout.Context, stateValue *tableState) {
	enabled := gtx.Enabled() && !t.disabled
	step := max(gtx.Dp(frame.ActiveTheme(ctx).Components.Table.ColumnResizeStep), 1)
	for _, column := range t.columns {
		if !column.Resizable {
			continue
		}
		minimum := max(gtx.Dp(frame.ActiveTheme(ctx).Components.Table.MinColumnWidth), 1)
		if column.MinWidth > 0 {
			minimum = max(gtx.Dp(unit.Dp(column.MinWidth)), 1)
		}
		maximum := gtx.Dp(unit.Dp(4096))
		if column.MaxWidth > 0 {
			maximum = max(gtx.Dp(unit.Dp(column.MaxWidth)), minimum)
		}
		resize := &stateValue.column(column.Key).resize
		configuredWidth := 0
		if column.Width > 0 {
			configuredWidth = gtx.Dp(unit.Dp(column.Width))
		}
		current, ok := resize.interactionWidth(configuredWidth)
		if !ok {
			current = minimum
		}
		current = min(max(current, minimum), maximum)
		if next, changed := resize.update(ctx, gtx, current, minimum, maximum, step, enabled); changed && t.onColumnResize != nil {
			t.onColumnResize(column.Key, int(gtx.Metric.PxToDp(next)+0.5))
		}
	}
}

func (t Widget) activate(key string) {
	t.activateWithModifiers(nil, key, 0)
}

func (t Widget) activateWithModifiers(stateValue *tableState, rowKey string, modifiers key.Modifiers) {
	rangeSelection := t.selectionMode == SelectionMultiple && modifiers.Contain(key.ModShift) && stateValue != nil
	if t.onAction != nil && !rangeSelection {
		t.onAction(rowKey)
	}
	switch t.selectionMode {
	case SelectionSingle:
		next := rowKey
		if t.allowEmpty && rowKey == t.selectedKey {
			next = ""
		}
		if next != t.selectedKey && t.onChange != nil {
			t.onChange(next)
		}
	case SelectionMultiple:
		if rangeSelection {
			next := t.rangeKeys(stateValue.selectionAnchor, rowKey)
			if !sameKeys(next, t.selectedKeys) && t.onSelectionChange != nil {
				t.onSelectionChange(next)
			}
			return
		}
		if stateValue != nil {
			stateValue.selectionAnchor = rowKey
		}
		next := toggleKey(t.selectedKeys, rowKey)
		if !sameKeys(next, t.selectedKeys) && t.onSelectionChange != nil {
			t.onSelectionChange(next)
		}
	}
}

func (t Widget) requestSort(column string) {
	if t.onSortChange == nil {
		return
	}
	next := SortDescriptor{Column: column, Direction: SortAscending}
	if t.sort.Column == column && t.sort.Direction == SortAscending {
		next.Direction = SortDescending
	}
	t.onSortChange(next)
}

func (t Widget) toggleAll() {
	if t.onSelectionChange == nil {
		return
	}
	all, _ := t.selectionSummary()
	next := make([]string, 0, t.count())
	for index := range t.count() {
		row := t.row(index)
		if t.rowDisabled(row) {
			if containsKey(t.selectedKeys, row.Key) {
				next = append(next, row.Key)
			}
			continue
		}
		if !all {
			next = append(next, row.Key)
		}
	}
	if !sameKeys(next, t.selectedKeys) {
		t.onSelectionChange(next)
	}
}

func (t Widget) selectionSummary() (all, some bool) {
	if t.virtual {
		return t.virtualSelectionSummary()
	}
	enabled := 0
	selected := 0
	for index := range t.count() {
		row := t.row(index)
		if t.rowDisabled(row) {
			continue
		}
		enabled++
		if t.isSelected(row.Key) {
			selected++
		}
	}
	return enabled > 0 && selected == enabled, selected > 0 && selected < enabled
}

func (t Widget) virtualSelectionSummary() (all, some bool) {
	disabled := make(map[string]struct{}, len(t.disabledKeys))
	for _, key := range t.disabledKeys {
		if key != "" {
			disabled[key] = struct{}{}
		}
	}
	enabled := max(t.count()-len(disabled), 0)
	selected := make(map[string]struct{}, len(t.selectedKeys))
	for _, key := range t.selectedKeys {
		if key == "" {
			continue
		}
		if _, excluded := disabled[key]; !excluded {
			selected[key] = struct{}{}
		}
	}
	selectedCount := len(selected)
	all = enabled > 0 && selectedCount >= enabled
	return all, selectedCount > 0 && !all
}

func (t Widget) isSelected(key string) bool {
	switch t.selectionMode {
	case SelectionSingle:
		return key == t.selectedKey
	case SelectionMultiple:
		return stateutil.StringSetContains(t.selectedKeys, t.selectedKeySet, key)
	default:
		return false
	}
}

func (t Widget) rowDisabled(row Row) bool {
	return t.disabled || row.Disabled || stateutil.StringSetContains(t.disabledKeys, t.disabledKeySet, row.Key)
}

func (t Widget) showsSelectionIndicator() bool {
	return t.selectionIndicator && t.selectionMode != SelectionNone
}

func (t Widget) count() int {
	if t.virtual {
		return t.rowCount
	}
	return len(t.rows)
}

func (t Widget) row(index int) Row {
	if index < 0 || index >= t.count() {
		panic("flowui: table row index out of range")
	}
	if t.virtual {
		row := t.rowProvider(index)
		validateRow(t.columns, row)
		return row
	}
	return t.rows[index]
}

func (t Widget) rowIndex(key string) int {
	for index := range t.count() {
		if t.row(index).Key == key {
			return index
		}
	}
	return -1
}

func (t Widget) updateRowClicks(gtx layout.Context, stateValue *tableState) {
	if !t.virtual {
		for index, row := range t.rows {
			rowState := stateValue.rowAt(row.Key, index)
			t.consumeRowClicks(gtx, stateValue, row, rowState)
		}
		return
	}
	for key, rowState := range stateValue.rows {
		index := rowState.index
		if index < 0 || index >= t.count() {
			continue
		}
		row := t.row(index)
		if row.Key != key {
			continue
		}
		t.consumeRowClicks(gtx, stateValue, row, rowState)
	}
}

func (t Widget) consumeRowClicks(gtx layout.Context, stateValue *tableState, row Row, rowState *tableRowState) {
	for {
		click, ok := rowState.clickable.Update(gtx)
		if !ok {
			return
		}
		if rowState.clickInInteractiveCell() {
			continue
		}
		if !t.rowDisabled(row) {
			t.activateWithModifiers(stateValue, row.Key, click.Modifiers)
		}
	}
}

func (t Widget) rangeKeys(anchor, target string) []string {
	selectedIndexes := make(map[string]int, len(t.selectedKeys))
	selectedDisabled := make(map[string]bool, len(t.selectedKeys))
	selectedSet := make(map[string]struct{}, len(t.selectedKeys))
	for _, key := range t.selectedKeys {
		selectedSet[key] = struct{}{}
	}
	anchorIndex, targetIndex := -1, -1
	for index := range t.count() {
		row := t.row(index)
		if row.Key == anchor {
			anchorIndex = index
		}
		if row.Key == target {
			targetIndex = index
		}
		if _, selected := selectedSet[row.Key]; selected {
			selectedIndexes[row.Key] = index
			selectedDisabled[row.Key] = t.rowDisabled(row)
		}
	}
	if anchorIndex < 0 {
		for _, selected := range t.selectedKeys {
			if index, ok := selectedIndexes[selected]; ok {
				anchorIndex = index
				break
			}
		}
	}
	if anchorIndex < 0 {
		anchorIndex = targetIndex
	}
	if targetIndex < 0 {
		return append([]string(nil), t.selectedKeys...)
	}
	start, end := min(anchorIndex, targetIndex), max(anchorIndex, targetIndex)
	next := make([]string, 0, end-start+1)
	for _, selected := range t.selectedKeys {
		if selectedDisabled[selected] {
			next = append(next, selected)
		}
	}
	for index := start; index <= end; index++ {
		row := t.row(index)
		if !t.rowDisabled(row) && !containsKey(next, row.Key) {
			next = append(next, row.Key)
		}
	}
	return next
}

func toggleKey(keys []string, key string) []string {
	next := make([]string, 0, len(keys)+1)
	removed := false
	for _, current := range keys {
		if current == key {
			removed = true
			continue
		}
		if current != "" && !containsKey(next, current) {
			next = append(next, current)
		}
	}
	if !removed {
		next = append(next, key)
	}
	return next
}

func containsKey(keys []string, key string) bool {
	for _, current := range keys {
		if current == key {
			return true
		}
	}
	return false
}

func sameKeys(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}
