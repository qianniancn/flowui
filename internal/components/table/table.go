package table

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
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
	Weight    float32
	Align     Alignment
	Sortable  bool
	RowHeader bool
}

// Cell contains either plain text or a custom widget. Content takes precedence.
type Cell struct {
	Text    string
	Content frame.Widget
}

// Row describes one controlled Table row.
type Row struct {
	Key      string
	Label    string
	Cells    []Cell
	Disabled bool
}

// Widget presents structured data with controlled sorting and selection.
type Widget struct {
	key                string
	columns            []Column
	rows               []Row
	variant            Variant
	selectionMode      SelectionMode
	selectedKey        string
	selectedKeys       []string
	sort               SortDescriptor
	disabledKeys       []string
	emptyText          string
	emptyContent       frame.Widget
	footer             frame.Widget
	onChange           func(string)
	onSelectionChange  func([]string)
	onSortChange       func(SortDescriptor)
	onAction           func(string)
	disabled           bool
	allowEmpty         bool
	selectionIndicator bool
	maxHeight          int
	minWidth           int
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

func (t Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := tableStateFor(ctx, t.key)
	state.beginFrame()
	defer state.endFrame()
	state.check(t.columns, t.rows)

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
		for _, row := range t.rows {
			rowState := state.row(row.Key)
			for rowState.clickable.Clicked(gtx) {
				if !t.rowDisabled(row) {
					t.activate(row.Key)
				}
			}
		}
		result := state.updateKeys(gtx, t)
		if result.focusKey != "" {
			frame.RequestFocus(ctx, &state.row(result.focusKey).clickable)
			state.ensureVisible(rowIndex(t.rows, result.focusKey))
		}
		if result.actionKey != "" {
			t.activate(result.actionKey)
		}
	}
	if t.disabled {
		gtx = gtx.Disabled()
	}
	return t.layout(ctx, gtx, state)
}

func (t Widget) activate(key string) {
	if t.onAction != nil {
		t.onAction(key)
	}
	switch t.selectionMode {
	case SelectionSingle:
		next := key
		if t.allowEmpty && key == t.selectedKey {
			next = ""
		}
		if next != t.selectedKey && t.onChange != nil {
			t.onChange(next)
		}
	case SelectionMultiple:
		next := toggleKey(t.selectedKeys, key)
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
	next := make([]string, 0, len(t.rows))
	for _, row := range t.rows {
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
	enabled := 0
	selected := 0
	for _, row := range t.rows {
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

func (t Widget) isSelected(key string) bool {
	switch t.selectionMode {
	case SelectionSingle:
		return key == t.selectedKey
	case SelectionMultiple:
		return containsKey(t.selectedKeys, key)
	default:
		return false
	}
}

func (t Widget) rowDisabled(row Row) bool {
	return t.disabled || row.Disabled || containsKey(t.disabledKeys, row.Key)
}

func (t Widget) showsSelectionIndicator() bool {
	return t.selectionIndicator && t.selectionMode != SelectionNone
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

func rowIndex(rows []Row, key string) int {
	for index, row := range rows {
		if row.Key == key {
			return index
		}
	}
	return -1
}
