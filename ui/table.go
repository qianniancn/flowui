package ui

import "github.com/qianniancn/FlowUI/internal/components/table"

type TableWidget = table.Widget
type TableColumn = table.Column
type TableRow = table.Row
type TableCell = table.Cell
type TableVariant = table.Variant
type TableSelectionMode = table.SelectionMode
type TableSortDirection = table.SortDirection
type TableSortDescriptor = table.SortDescriptor
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

func Table(key string, columns []TableColumn, rows []TableRow) TableWidget {
	return table.New(key, columns, rows)
}
