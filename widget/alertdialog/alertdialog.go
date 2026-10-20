package alertdialog

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/modal"
	"github.com/qianniancn/flowui/widget/text"
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
	override                *theme.ModalOverride
	key                     string
	open                    bool
	hasOpen                 bool
	defaultOpen             bool
	hasDefaultOpen          bool
	title                   string
	description             string
	status                  Status
	body                    core.Widget
	header                  core.Widget
	footer                  core.Widget
	icon                    core.Widget
	onOpenChange            func(bool)
	size                    Size
	placement               Placement
	backdrop                BackdropVariant
	dismissable             bool
	keyboardDismissDisabled bool
	closeButton             bool
	visualCache             *alertDialogVisualCache
}

// New creates an alert dialog. When open is true, it starts in controlled
// mode with the dialog open. When open is false, it starts in uncontrolled mode
// with the dialog closed (use DefaultOpen(true) for uncontrolled mode starting open).
// New creates an alert dialog. It starts uncontrolled and closed; Open sets the
// controlled value and DefaultOpen the uncontrolled starting value.
func New(title, description string) Widget {
	return Widget{
		key:                     core.AutoKey(),
		title:                   title,
		description:             description,
		status:                  StatusDanger,
		keyboardDismissDisabled: true,
		closeButton:             true,
		visualCache:             &alertDialogVisualCache{},
	}
}

// Key sets an explicit identity for the dialog state.
func (a Widget) Key(key string) Widget {
	a.key = key
	return a
}

// Uncontrolled makes the constructor's open value an internal initial state.
func (a Widget) Uncontrolled() Widget {
	a.defaultOpen = a.open
	a.hasDefaultOpen = true
	a.hasOpen = false
	return a
}

// OnOpenChange registers a callback for close requests.
func (a Widget) OnOpenChange(fn func(bool)) Widget {
	a.onOpenChange = fn
	return a
}

// Open sets the alert dialog to controlled mode with the given open state.
func (a Widget) Open(open bool) Widget {
	a.open = open
	a.hasOpen = true
	return a
}

// DefaultOpen sets the initial open state for uncontrolled mode.
func (a Widget) DefaultOpen(open bool) Widget {
	a.defaultOpen = open
	a.hasDefaultOpen = true
	return a
}

// Status sets the semantic tone of the default or custom icon.
func (a Widget) Status(status Status) Widget {
	a.status = status
	return a
}

// Body replaces the standard description with custom content.
func (a Widget) Body(body core.Widget) Widget {
	a.body = body
	return a
}

// Header replaces the default status icon and heading.
func (a Widget) Header(header core.Widget) Widget {
	a.header = header
	return a
}

// Footer sets the action area displayed below the body.
func (a Widget) Footer(footer core.Widget) Widget {
	a.footer = footer
	return a
}

// Icon replaces the default Lucide status icon.
func (a Widget) Icon(icon core.Widget) Widget {
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

func (a Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return a.modal(a.override).Layout(ctx, gtx)
}

// Style merges instance declarations into the dialog. The declaration reaches
// the provider at layout time, so Style can sit anywhere in the chain.
func (a Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if a.override == nil {
		a.override = &theme.ModalOverride{}
	}
	a.override.Declaration = style.Join(a.override.Declaration, declaration)
	return a
}

func (a Widget) modal(override *theme.ModalOverride) modal.Widget {
	body := a.body
	if body == nil {
		body = text.New(a.description)
	}
	header := a.header
	if header == nil {
		header = a.defaultHeader(override)
	}
	m := modal.New("", body).Key(a.key).Open(a.open)
	// Apply open-state fields explicitly if set
	if a.hasOpen {
		m = m.Open(a.open)
	}
	if a.hasDefaultOpen {
		m = m.DefaultOpen(a.defaultOpen)
	}
	return m.
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

func (a Widget) defaultHeader(override *theme.ModalOverride) dialogHeader {
	description := a.description
	if a.body != nil {
		description = ""
	}
	return dialogHeader{
		title:       a.title,
		description: description,
		status:      a.status,
		icon:        a.icon,
		cache:       a.visualCache,
		size:        a.size,
		backdrop:    a.backdrop,
		override:    override,
	}
}

func (a Widget) modalSize() modal.Size {
	switch a.size {
	case SizeXSmall:
		return modal.XSmall
	case SizeSmall:
		return modal.Small
	case SizeLarge:
		return modal.Large
	case SizeCover:
		return modal.Cover
	default:
		return modal.Medium
	}
}

func (a Widget) modalPlacement() modal.Placement {
	switch a.placement {
	case PlacementTop:
		return modal.Top
	case PlacementCenter:
		return modal.Center
	case PlacementBottom:
		return modal.Bottom
	default:
		return modal.Auto
	}
}

func (a Widget) modalBackdrop() modal.BackdropVariant {
	switch a.backdrop {
	case BackdropBlur:
		return modal.BackdropBlur
	case BackdropTransparent:
		return modal.BackdropTransparent
	default:
		return modal.BackdropOpaque
	}
}
