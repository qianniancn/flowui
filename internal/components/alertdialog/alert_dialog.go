package alertdialog

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/modal"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// Status identifies the semantic tone of an AlertDialog icon.
type Status uint8

const (
	StatusDefault Status = iota
	StatusAccent
	StatusSuccess
	StatusWarning
	StatusDanger
)

// Size selects the dialog width and cover behavior.
type Size uint8

const (
	SizeMedium Size = iota
	SizeXSmall
	SizeSmall
	SizeLarge
	SizeCover
)

// Placement selects where the dialog is placed inside the overlay.
type Placement uint8

const (
	PlacementAuto Placement = iota
	PlacementTop
	PlacementCenter
	PlacementBottom
)

// BackdropVariant selects the treatment behind the dialog.
type BackdropVariant uint8

const (
	BackdropOpaque BackdropVariant = iota
	BackdropBlur
	BackdropTransparent
)

// Widget presents a controlled confirmation dialog that requires an explicit action by default.
type Widget struct {
	theme                   func(*theme.Theme)
	key                     string
	open                    bool
	title                   string
	description             string
	status                  Status
	body                    frame.Widget
	header                  frame.Widget
	footer                  frame.Widget
	icon                    frame.Widget
	onOpenChange            func(bool)
	size                    Size
	placement               Placement
	backdrop                BackdropVariant
	dismissable             bool
	keyboardDismissDisabled bool
	closeButton             bool
}

// New creates a controlled alert dialog with HeroUI's danger status defaults.
func New(key string, open bool, title, description string) Widget {
	return Widget{
		key:                     key,
		open:                    open,
		title:                   title,
		description:             description,
		status:                  StatusDanger,
		keyboardDismissDisabled: true,
		closeButton:             true,
	}
}

// OnOpenChange registers a callback for close requests.
func (a Widget) OnOpenChange(fn func(bool)) Widget {
	a.onOpenChange = fn
	return a
}

// Status sets the semantic tone of the default or custom icon.
func (a Widget) Status(status Status) Widget {
	a.status = status
	return a
}

// Body replaces the standard description with custom content.
func (a Widget) Body(body frame.Widget) Widget {
	a.body = body
	return a
}

// Header replaces the default status icon and heading.
func (a Widget) Header(header frame.Widget) Widget {
	a.header = header
	return a
}

// Footer sets the action area displayed below the body.
func (a Widget) Footer(footer frame.Widget) Widget {
	a.footer = footer
	return a
}

// Icon replaces the default Lucide status icon.
func (a Widget) Icon(icon frame.Widget) Widget {
	a.icon = icon
	return a
}

// Size sets the dialog size preset.
func (a Widget) Size(size Size) Widget {
	a.size = size
	return a
}

// Placement sets the dialog placement preset.
func (a Widget) Placement(placement Placement) Widget {
	a.placement = placement
	return a
}

// Backdrop sets the backdrop variant.
func (a Widget) Backdrop(backdrop BackdropVariant) Widget {
	a.backdrop = backdrop
	return a
}

// Dismissable controls whether clicking outside the dialog requests close.
// It defaults to false because alert dialogs normally require an explicit action.
func (a Widget) Dismissable(dismissable bool) Widget {
	a.dismissable = dismissable
	return a
}

// KeyboardDismissDisabled controls whether Escape can request close.
// It defaults to true because alert dialogs normally require an explicit action.
func (a Widget) KeyboardDismissDisabled(disabled bool) Widget {
	a.keyboardDismissDisabled = disabled
	return a
}

// CloseButton controls whether the close button is shown.
func (a Widget) CloseButton(show bool) Widget {
	a.closeButton = show
	return a
}

func (a Widget) Theme(fn func(*theme.Theme)) Widget {
	a.theme = fn
	return a
}

func (a Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, a.theme); restore != nil {
		defer restore()
	}
	return a.modal().Layout(ctx, gtx)
}

func (a Widget) modal() modal.ModalWidget {
	body := a.body
	if body == nil {
		body = text.New(a.description)
	}
	header := a.header
	if header == nil {
		header = a.defaultHeader()
	}
	return modal.Modal(a.key, a.open, "", body).
		Header(header).
		Footer(a.footer).
		OnOpenChange(a.onOpenChange).
		Size(a.modalSize()).
		Placement(a.modalPlacement()).
		Backdrop(a.modalBackdrop()).
		Dismissable(a.dismissable).
		KeyboardDismissDisabled(a.keyboardDismissDisabled).
		CloseButton(a.closeButton)
}

func (a Widget) defaultHeader() dialogHeader {
	description := a.description
	if a.body != nil {
		description = ""
	}
	return dialogHeader{
		title:       a.title,
		description: description,
		status:      a.status,
		icon:        a.icon,
	}
}

func (a Widget) modalSize() modal.ModalSize {
	switch a.size {
	case SizeXSmall:
		return modal.ModalXSmall
	case SizeSmall:
		return modal.ModalSmall
	case SizeLarge:
		return modal.ModalLarge
	case SizeCover:
		return modal.ModalCover
	default:
		return modal.ModalMedium
	}
}

func (a Widget) modalPlacement() modal.ModalPlacement {
	switch a.placement {
	case PlacementTop:
		return modal.ModalTop
	case PlacementCenter:
		return modal.ModalCenter
	case PlacementBottom:
		return modal.ModalBottom
	default:
		return modal.ModalAuto
	}
}

func (a Widget) modalBackdrop() modal.ModalBackdropVariant {
	switch a.backdrop {
	case BackdropBlur:
		return modal.ModalBackdropBlur
	case BackdropTransparent:
		return modal.ModalBackdropTransparent
	default:
		return modal.ModalBackdropOpaque
	}
}
