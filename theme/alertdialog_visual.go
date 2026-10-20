package theme

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// AlertDialogRequest contains the semantic axes that affect an alert dialog's
// header visual. Overlay behavior remains owned by modal behavior.
type AlertDialogRequest struct {
	Status   AlertStatus
	Size     ModalSize
	Backdrop ModalBackdrop
	Override *ModalOverride
}

// AlertDialogMetrics contains neutral header geometry consumed by the dialog
// behavior widget.
type AlertDialogMetrics struct {
	IconSize      unit.Dp
	IconGlyphSize unit.Dp
	HeaderGap     unit.Dp
}

// AlertDialogIconFrameRequest contains the neutral geometry and icon content
// needed by a provider-owned status icon frame.
type AlertDialogIconFrameRequest struct {
	Size    image.Point
	Content func(layout.Context) layout.Dimensions
}

// AlertDialogIconFrameRenderer paints and lays out the status icon frame.
// The widget supplies only the icon content; the provider owns the frame
// shape, clipping, and background treatment.
type AlertDialogIconFrameRenderer func(layout.Context, AlertDialogIconFrameRequest) layout.Dimensions

// AlertDialogVisual is the complete provider-owned header visual plan.
// Icon is a theme-owned glyph; colors are returned for both the default and
// caller-supplied icon paths.
type AlertDialogVisual struct {
	Title          style.Style
	IconBackground color.NRGBA
	IconForeground color.NRGBA
	Override       style.Style
	Metrics        AlertDialogMetrics
	Icon           []byte
	IconFrame      AlertDialogIconFrameRenderer
}

// AlertDialogProvider supplies status header declarations and glyphs without
// taking ownership of modal behavior, focus, dismissal, or content slots.
type AlertDialogProvider interface {
	AlertDialog(active *Theme, request AlertDialogRequest) AlertDialogVisual
}

// ValidateAlertDialogVisual checks the provider-owned status header contract
// before a widget stores it in its visual cache. Modal behavior owns the
// overlay lifecycle; AlertDialog only supplies header typography, icon colors,
// and neutral icon geometry.
func ValidateAlertDialogVisual(visual AlertDialogVisual) {
	if visual.IconFrame == nil {
		invalidVisual("AlertDialog", "IconFrame", "must be provided")
	}

	metrics := visual.Metrics
	if !validPositiveDp(metrics.IconSize) {
		invalidVisual("AlertDialog", "Metrics.IconSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.IconGlyphSize) {
		invalidVisual("AlertDialog", "Metrics.IconGlyphSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.HeaderGap) {
		invalidVisual("AlertDialog", "Metrics.HeaderGap", "must be finite and non-negative")
	}

	title := style.Cascade(style.State{}, visual.Title)
	if title.Text == nil {
		invalidVisual("AlertDialog", "Title.Text", "must be provided")
	}
	if title.Text.FontSize == nil {
		invalidVisual("AlertDialog", "Title.Text.FontSize", "must be provided")
	}
	if !validPositiveSp(*title.Text.FontSize) {
		invalidVisual("AlertDialog", "Title.Text.FontSize", "must be finite and greater than 0")
	}
	if title.Text.FontWeight == nil {
		invalidVisual("AlertDialog", "Title.Text.FontWeight", "must be provided")
	}
	if !hasColor(title.Text.Color) {
		invalidVisual("AlertDialog", "Title.Text.Color", "must resolve to a solid color")
	}

	if visual.IconBackground.A == 0 {
		invalidVisual("AlertDialog", "IconBackground.A", "must be nonzero")
	}
	if visual.IconForeground.A == 0 {
		invalidVisual("AlertDialog", "IconForeground.A", "must be nonzero")
	}
}
