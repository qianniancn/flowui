package theme

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// PaginationSize identifies the provider-owned density tier for pagination.
// Page calculation and selection remain behavior-owned by widget/pagination.
type PaginationSize uint8

const (
	PaginationMedium PaginationSize = iota
	PaginationSmall
	PaginationLarge
)

// PaginationRequest describes the semantic axes needed by a pagination
// visual provider. Labels, page ranges, and callbacks remain widget inputs.
type PaginationRequest struct {
	Size     PaginationSize
	Override *PaginationOverride
}

// PaginationOverride carries caller customization across the provider
// boundary without storing visual state on the behavior widget.
type PaginationOverride struct {
	Declaration style.Style
}

// PaginationOverrideStyle extracts the provider-owned instance layer.
func PaginationOverrideStyle(request PaginationRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// PaginationMetrics contains only geometry consumed by pagination layout.
// Button chrome geometry remains private to the provider-owned renderer.
type PaginationMetrics struct {
	Size         unit.Dp
	ItemGap      unit.Dp
	ContentGap   unit.Dp
	CompactWidth unit.Dp
}

// PaginationButtonKind identifies the semantic slot being painted by a
// provider-owned renderer.
type PaginationButtonKind uint8

const (
	PaginationPage PaginationButtonKind = iota
	PaginationPrevious
	PaginationNext
	PaginationEllipsis
)

// PaginationButtonRequest contains behavior-owned state for one pagination
// slot. The provider decides colors, typography, iconography, shape, focus,
// and press treatment.
type PaginationButtonRequest struct {
	Kind         PaginationButtonKind
	Label        string
	Active       bool
	Disabled     bool
	Hovered      bool
	Pressed      bool
	FocusVisible bool
	Focus        float32
}

// PaginationButtonRenderer paints one pagination slot and returns its
// dimensions under the supplied constraints.
type PaginationButtonRenderer func(layout.Context, PaginationButtonRequest) layout.Dimensions

// PaginationVisual is the complete visual plan consumed by pagination
// behavior. The widget owns page math, semantics, interaction and callbacks;
// all visual decisions live behind this protocol.
type PaginationVisual struct {
	Defaults style.Style
	Summary  style.Style
	Override style.Style
	Metrics  PaginationMetrics
	Motion   MotionTheme
	Button   PaginationButtonRenderer
}

// PaginationProvider supplies pagination declarations, geometry and button
// rendering without taking ownership of page state or interaction behavior.
type PaginationProvider interface {
	Pagination(active *Theme, request PaginationRequest) PaginationVisual
}

// ValidatePaginationVisual checks the provider-owned pagination declaration
// before a widget stores it in its cache.
func ValidatePaginationVisual(visual PaginationVisual) {
	m := visual.Metrics
	if m.Size <= 0 {
		invalidVisual("Pagination", "Metrics.Size", "must be greater than 0")
	}
	if m.ItemGap < 0 {
		invalidVisual("Pagination", "Metrics.ItemGap", "must be at least 0")
	}
	if m.ContentGap < 0 {
		invalidVisual("Pagination", "Metrics.ContentGap", "must be at least 0")
	}
	if m.CompactWidth <= 0 {
		invalidVisual("Pagination", "Metrics.CompactWidth", "must be greater than 0")
	}
	if visual.Button == nil {
		invalidVisual("Pagination", "Button", "must be provided")
	}

	summary := style.Cascade(style.State{}, visual.Summary, visual.Override)
	if summary.Text == nil {
		invalidVisual("Pagination", "Summary.Text", "must be provided")
	}
	if summary.Text.FontSize == nil {
		invalidVisual("Pagination", "Summary.Text.FontSize", "must be provided")
	}
	if !validPositiveSp(*summary.Text.FontSize) {
		invalidVisual("Pagination", "Summary.Text.FontSize", "must be finite and greater than 0")
	}
	if !hasColor(summary.Text.Color) {
		invalidVisual("Pagination", "Summary.Text.Color", "must resolve to a solid color")
	}

	validateMotion("Pagination", "Motion", visual.Motion)
}
