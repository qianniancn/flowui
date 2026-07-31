package ui

import "github.com/qianniancn/flowui/internal/components/table"

type TableWidget = table.Widget

// TableColumn describes one table column.
type TableColumn = table.Column

// TableRow contains the cells for one table row.
type TableRow = table.Row

// TableRowProvider supplies rows for a virtual table.
type TableRowProvider = table.RowProvider

// TableRowContextMenu describes the menu shown for a table row.
type TableRowContextMenu = table.RowContextMenu

// TableCell contains one table cell's content.
type TableCell = table.Cell

// TableVariant selects the table surface treatment.
type TableVariant = table.Variant

// TableSelectionMode controls how many rows may be selected.
type TableSelectionMode = table.SelectionMode

// TableSortDirection identifies the direction of a sort.
type TableSortDirection = table.SortDirection

// TableSortDescriptor describes the active table sort.
type TableSortDescriptor = table.SortDescriptor

// TableAlignment controls cell content alignment.
type TableAlignment = table.Alignment

const (
	TablePrimary   = table.VariantPrimary
	TableSecondary = table.VariantSecondary

	TableSelectionNone     = table.SelectionNone
	TableSelectionSingle   = table.SelectionSingle
	TableSelectionMultiple = table.SelectionMultiple

	TableSortAscending  = table.SortAscending
	TableSortDescending = table.SortDescending

	TableAlignStart  = table.AlignStart
	TableAlignCenter = table.AlignCenter
	TableAlignEnd    = table.AlignEnd
)

// Table creates a table from its columns and in-memory rows.
func Table(key string, columns []TableColumn, rows []TableRow) TableWidget {
	return table.New(key, columns, rows)
}

// VirtualTable creates a table whose rows are supplied on demand.
func VirtualTable(key string, columns []TableColumn, count int, provider TableRowProvider) TableWidget {
	return table.NewVirtual(key, columns, count, provider)
}
