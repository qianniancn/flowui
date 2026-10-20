package theme

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// TableVariant identifies the visual treatment of a table container without
// importing widget/table into the theme protocol.
type TableVariant uint8

const (
	TablePrimary TableVariant = iota
	TableSecondary
)

// TableSortDirection is the provider-neutral sort arrow direction.
type TableSortDirection uint8

const (
	TableSortAscending TableSortDirection = iota
	TableSortDescending
)

// TableRequest describes the behavior-owned table variant and optional
// instance declaration needed by a visual provider. Selection, hover, sort,
// and disabled state are supplied to Row at render time.
type TableRequest struct {
	Variant  TableVariant
	Override *TableOverride
}

// TableOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type TableOverride struct {
	Declaration style.Style
}

// TableOverrideStyle extracts the provider-owned instance layer.
func TableOverrideStyle(request TableRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// TableMetrics contains geometry and typography consumed by table behavior
// layout. It intentionally contains no colors or drawing policy.
type TableMetrics struct {
	RootPadding           unit.Dp
	RootRadius            unit.Dp
	HeaderRadius          unit.Dp
	BodyRadius            unit.Dp
	HeaderHeight          unit.Dp
	RowMinHeight          unit.Dp
	EmptyHeight           unit.Dp
	MaxHeight             unit.Dp
	MinColumnWidth        unit.Dp
	CellPaddingX          unit.Dp
	CellPaddingY          unit.Dp
	HeaderTextSize        unit.Sp
	HeaderTextWeight      font.Weight
	CellTextSize          unit.Sp
	SeparatorWidth        unit.Dp
	ColumnSeparatorHeight unit.Dp
	ColumnResizerHitSize  unit.Dp
	ColumnResizerWidth    unit.Dp
	ColumnResizeStep      unit.Dp
	FocusRingWidth        unit.Dp
	FocusRadius           unit.Dp
	SelectionColumnWidth  unit.Dp
	SortIconSize          unit.Dp
	SortGap               unit.Dp
	FooterPaddingX        unit.Dp
	FooterPaddingY        unit.Dp
	LoadMoreHeight        unit.Dp
}

// TableColors are semantic fallback colors used by table behavior when a
// composed child does not provide an explicit visual declaration.
type TableColors struct {
	Root            color.NRGBA
	Header          color.NRGBA
	Body            color.NRGBA
	Foreground      color.NRGBA
	Muted           color.NRGBA
	ColumnSeparator color.NRGBA
	HeaderSeparator color.NRGBA
	RowSeparator    color.NRGBA
	Border          color.NRGBA
	Focus           color.NRGBA
	Stripe          color.NRGBA
	ResizeAccent    color.NRGBA
}

// TableRowRequest contains behavior-owned interaction state for one row.
type TableRowRequest struct {
	Variant  TableVariant
	Stripe   color.NRGBA
	Selected bool
	Hovered  bool
	Pressed  bool
	Disabled bool
}

// TableRowVisual is the provider-resolved state used by row composition.
type TableRowVisual struct {
	Background color.NRGBA
	Foreground color.NRGBA
	Focus      color.NRGBA
	Opacity    float32
}

// TableTextRequest describes a provider-owned text primitive.
type TableTextRequest struct {
	Text   string
	Size   unit.Sp
	Color  color.NRGBA
	Weight font.Weight
}

// TableRenderer types keep painting in the theme recipe while the widget
// retains only geometry, semantics, and interaction behavior.
type TableRootRenderer func(layout.Context, image.Point, unit.Dp, color.NRGBA)

type TableBorderRenderer func(layout.Context, image.Point, unit.Dp, unit.Dp, color.NRGBA)

type TableHeaderRenderer func(layout.Context, image.Point, unit.Dp, color.NRGBA, color.NRGBA)

type TableBodyRenderer func(layout.Context, image.Point, unit.Dp, color.NRGBA)

type TableRowRenderer func(layout.Context, image.Point, TableRowVisual, color.NRGBA, bool, float32)

type TableHeaderSeparatorRenderer func(layout.Context, int, int, color.NRGBA, bool)

type TableRowSeparatorRenderer func(layout.Context, int, []int, int, color.NRGBA)

type TableResizerRenderer func(layout.Context, int, int, int, int, color.NRGBA, color.NRGBA, bool, float32)

type TableCellFocusRenderer func(layout.Context, image.Point, float32, color.NRGBA)

type TableSortRenderer func(layout.Context, image.Point, TableSortDirection, color.NRGBA)

type TableTextRenderer func(layout.Context, TableTextRequest) layout.Dimensions

// TableVisual is the complete visual plan consumed by table behavior.
type TableVisual struct {
	Metrics   TableMetrics
	Colors    TableColors
	Motion    MotionTheme
	Override  style.Style
	Row       func(TableRowRequest) TableRowVisual
	Root      TableRootRenderer
	Border    TableBorderRenderer
	Header    TableHeaderRenderer
	Body      TableBodyRenderer
	RowRender TableRowRenderer
	HeaderSep TableHeaderSeparatorRenderer
	RowSep    TableRowSeparatorRenderer
	Resizer   TableResizerRenderer
	CellFocus TableCellFocusRenderer
	Sort      TableSortRenderer
	Text      TableTextRenderer
}

// TableProvider supplies table metrics, state visuals, and theme-owned
// rendering strategies without taking ownership of table behavior.
type TableProvider interface {
	Table(active *Theme, request TableRequest) TableVisual
}

// ValidateTableVisual checks the geometry, row states, motion and renderer
// contract before a table visual enters the shared provider cache. Table
// behavior may choose row selection and column geometry, but it must not
// recover missing theme metrics or drawing strategies with local defaults.
func ValidateTableVisual(visual TableVisual) {
	metrics := visual.Metrics
	if !validNonNegativeDp(metrics.RootPadding) {
		invalidVisual("Table", "Metrics.RootPadding", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.RootRadius) {
		invalidVisual("Table", "Metrics.RootRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.HeaderRadius) {
		invalidVisual("Table", "Metrics.HeaderRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.BodyRadius) {
		invalidVisual("Table", "Metrics.BodyRadius", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.HeaderHeight) {
		invalidVisual("Table", "Metrics.HeaderHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.RowMinHeight) {
		invalidVisual("Table", "Metrics.RowMinHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.EmptyHeight) {
		invalidVisual("Table", "Metrics.EmptyHeight", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.MaxHeight) {
		invalidVisual("Table", "Metrics.MaxHeight", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.MinColumnWidth) {
		invalidVisual("Table", "Metrics.MinColumnWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.CellPaddingX) {
		invalidVisual("Table", "Metrics.CellPaddingX", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.CellPaddingY) {
		invalidVisual("Table", "Metrics.CellPaddingY", "must be finite and non-negative")
	}
	if !validPositiveSp(metrics.HeaderTextSize) {
		invalidVisual("Table", "Metrics.HeaderTextSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(metrics.CellTextSize) {
		invalidVisual("Table", "Metrics.CellTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.SeparatorWidth) {
		invalidVisual("Table", "Metrics.SeparatorWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.ColumnSeparatorHeight) {
		invalidVisual("Table", "Metrics.ColumnSeparatorHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.ColumnResizerHitSize) {
		invalidVisual("Table", "Metrics.ColumnResizerHitSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.ColumnResizerWidth) {
		invalidVisual("Table", "Metrics.ColumnResizerWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.ColumnResizeStep) {
		invalidVisual("Table", "Metrics.ColumnResizeStep", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.FocusRingWidth) {
		invalidVisual("Table", "Metrics.FocusRingWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.FocusRadius) {
		invalidVisual("Table", "Metrics.FocusRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.SelectionColumnWidth) {
		invalidVisual("Table", "Metrics.SelectionColumnWidth", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.SortIconSize) {
		invalidVisual("Table", "Metrics.SortIconSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.SortGap) {
		invalidVisual("Table", "Metrics.SortGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.FooterPaddingX) {
		invalidVisual("Table", "Metrics.FooterPaddingX", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.FooterPaddingY) {
		invalidVisual("Table", "Metrics.FooterPaddingY", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.LoadMoreHeight) {
		invalidVisual("Table", "Metrics.LoadMoreHeight", "must be finite and greater than 0")
	}

	if visual.Row == nil {
		invalidVisual("Table", "Row", "must be provided")
	}
	if visual.Root == nil {
		invalidVisual("Table", "Root", "must be provided")
	}
	if visual.Border == nil {
		invalidVisual("Table", "Border", "must be provided")
	}
	if visual.Header == nil {
		invalidVisual("Table", "Header", "must be provided")
	}
	if visual.Body == nil {
		invalidVisual("Table", "Body", "must be provided")
	}
	if visual.RowRender == nil {
		invalidVisual("Table", "RowRender", "must be provided")
	}
	if visual.HeaderSep == nil {
		invalidVisual("Table", "HeaderSep", "must be provided")
	}
	if visual.RowSep == nil {
		invalidVisual("Table", "RowSep", "must be provided")
	}
	if visual.Resizer == nil {
		invalidVisual("Table", "Resizer", "must be provided")
	}
	if visual.CellFocus == nil {
		invalidVisual("Table", "CellFocus", "must be provided")
	}
	if visual.Sort == nil {
		invalidVisual("Table", "Sort", "must be provided")
	}
	if visual.Text == nil {
		invalidVisual("Table", "Text", "must be provided")
	}

	validateMotion("Table", "Motion", visual.Motion)
	for _, request := range []TableRowRequest{
		{},
		{Hovered: true},
		{Pressed: true},
		{Selected: true},
		{Disabled: true},
		{Hovered: true, Selected: true},
		{Pressed: true, Selected: true},
		{Disabled: true, Selected: true},
	} {
		row := visual.Row(request)
		if !validFiniteFloat32(row.Opacity) {
			invalidVisual("Table", "Row.Opacity", "must be finite")
		}
		if row.Opacity < 0 {
			invalidVisual("Table", "Row.Opacity", "must be at least 0")
		}
		if row.Opacity > 1 {
			invalidVisual("Table", "Row.Opacity", "must be at most 1")
		}
	}
}
