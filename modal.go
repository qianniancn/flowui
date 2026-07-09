package flowui

import (
	"time"

	"gioui.org/layout"
)

// ModalWidget renders an overlay dialog controlled by application state.
type ModalWidget struct {
	key                     string
	open                    bool
	title                   string
	body                    Widget
	header                  Widget
	footer                  Widget
	icon                    Widget
	onOpenChange            func(bool)
	size                    ModalSize
	placement               ModalPlacement
	backdrop                ModalBackdropVariant
	scroll                  ModalScroll
	animation               ModalAnimation
	dismissable             bool
	hasDismissable          bool
	keyboardDismissDisabled bool
	closeButton             bool
	hasCloseButton          bool
}

// ModalSize selects the dialog width and full-screen behavior.
type ModalSize int

const (
	// ModalMedium uses the default modal width.
	ModalMedium ModalSize = iota
	// ModalXSmall uses the smallest modal width.
	ModalXSmall
	// ModalSmall uses a compact modal width.
	ModalSmall
	// ModalLarge uses a wider modal width.
	ModalLarge
	// ModalCover fills the available safe area while preserving page margins.
	ModalCover
	// ModalFull fills the entire overlay.
	ModalFull
)

// ModalPlacement selects where the dialog is placed inside the overlay.
type ModalPlacement int

const (
	// ModalAuto places the modal based on viewport size.
	ModalAuto ModalPlacement = iota
	// ModalTop places the modal near the top edge.
	ModalTop
	// ModalCenter places the modal in the center of the overlay.
	ModalCenter
	// ModalBottom places the modal near the bottom edge.
	ModalBottom
)

// ModalBackdropVariant selects the backdrop treatment behind the dialog.
type ModalBackdropVariant int

const (
	// ModalBackdropOpaque draws the default opaque backdrop.
	ModalBackdropOpaque ModalBackdropVariant = iota
	// ModalBackdropBlur draws a translucent blurred backdrop.
	ModalBackdropBlur
	// ModalBackdropTransparent keeps the backdrop visually transparent while still blocking input.
	ModalBackdropTransparent
)

// ModalScroll selects whether the body or the dialog surface owns scrolling.
type ModalScroll int

const (
	// ModalScrollInside scrolls the body content inside a fixed dialog surface.
	ModalScrollInside ModalScroll = iota
	// ModalScrollOutside scrolls the entire dialog content inside the surface.
	ModalScrollOutside
)

// ModalAnimation selects the motion preset used when a modal enters and exits.
type ModalAnimation int

const (
	// ModalAnimationAuto uses FlowUI's default modal animation.
	ModalAnimationAuto ModalAnimation = iota
	// ModalAnimationFadeScale fades the modal while scaling it up on entry and down on exit.
	ModalAnimationFadeScale
	// ModalAnimationSlideDown fades the modal while moving it down from above.
	ModalAnimationSlideDown
	// ModalAnimationSlideUp fades the modal while moving it up from below.
	ModalAnimationSlideUp
	// ModalAnimationFade only fades the modal without moving or scaling it.
	ModalAnimationFade
	// ModalAnimationBounceScale adds a restrained overshoot while opening and a clean scale-down while closing.
	ModalAnimationBounceScale
	// ModalAnimationZoomOut starts slightly larger and settles back to the modal's final size.
	ModalAnimationZoomOut
	// ModalAnimationPop opens with a short pop-style scale overshoot and closes with a clean scale-down.
	ModalAnimationPop
)

const modalEnterDuration = 250 * time.Millisecond

// Modal creates a controlled modal dialog.
func Modal(key string, open bool, title string, body Widget) ModalWidget {
	return ModalWidget{
		key:   key,
		open:  open,
		title: title,
		body:  body,
	}
}

// OnOpenChange registers a callback for close requests.
func (m ModalWidget) OnOpenChange(fn func(bool)) ModalWidget {
	m.onOpenChange = fn
	return m
}

// Header replaces the default title/header area.
func (m ModalWidget) Header(header Widget) ModalWidget {
	m.header = header
	return m
}

// Body replaces the modal body content.
func (m ModalWidget) Body(body Widget) ModalWidget {
	m.body = body
	return m
}

// Footer sets the modal footer content.
func (m ModalWidget) Footer(footer Widget) ModalWidget {
	m.footer = footer
	return m
}

// Icon sets an optional icon shown above the default title.
func (m ModalWidget) Icon(icon Widget) ModalWidget {
	m.icon = icon
	return m
}

// Size sets the modal size preset.
func (m ModalWidget) Size(size ModalSize) ModalWidget {
	m.size = size
	return m
}

// Placement sets the modal placement preset.
func (m ModalWidget) Placement(placement ModalPlacement) ModalWidget {
	m.placement = placement
	return m
}

// Backdrop sets the modal backdrop variant.
func (m ModalWidget) Backdrop(backdrop ModalBackdropVariant) ModalWidget {
	m.backdrop = backdrop
	return m
}

// Scroll sets how overflowing modal content scrolls.
func (m ModalWidget) Scroll(scroll ModalScroll) ModalWidget {
	m.scroll = scroll
	return m
}

// Animation sets the modal enter and exit motion preset.
func (m ModalWidget) Animation(animation ModalAnimation) ModalWidget {
	m.animation = animation
	return m
}

// Dismissable controls whether clicking outside the dialog requests close.
func (m ModalWidget) Dismissable(dismissable bool) ModalWidget {
	m.dismissable = dismissable
	m.hasDismissable = true
	return m
}

// KeyboardDismissDisabled controls whether Escape can request close.
func (m ModalWidget) KeyboardDismissDisabled(disabled bool) ModalWidget {
	m.keyboardDismissDisabled = disabled
	return m
}

// CloseButton controls whether the default close button is shown.
func (m ModalWidget) CloseButton(show bool) ModalWidget {
	m.closeButton = show
	m.hasCloseButton = true
	return m
}

// Layout renders the modal overlay when it is open or exiting.
func (m ModalWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	fullKey := ctx.fullKey(m.key)
	if !m.open && !ctx.hasVisibleModal(fullKey) {
		return layout.Dimensions{}
	}
	state := ctx.modalState(m.key)
	progress := state.progress(gtx, m.open)
	state.syncFocus(ctx, m.open)
	if !m.open && progress <= 0 {
		delete(ctx.modals, fullKey)
		return layout.Dimensions{}
	}
	if m.open {
		m.handleCloseEvents(ctx, gtx, state)
	}

	ctx.deferOverlay(gtx, func(gtx layout.Context) layout.Dimensions {
		return m.layoutOverlay(ctx, gtx, state, progress)
	})
	return layout.Dimensions{}
}

func (m ModalWidget) handleCloseEvents(ctx *Context, gtx layout.Context, state *modalState) {
	for state.dialog.Clicked(gtx) {
	}
	if !m.keyboardDismissDisabled && state.escapePressed(gtx) {
		m.requestClose()
	}
	if m.showCloseButton() {
		for state.close.Clicked(gtx) {
			m.requestClose()
			ctx.requestFocus(&state.close)
		}
	}
	if m.isDismissable() {
		for i := range state.dismiss {
			for state.dismiss[i].Clicked(gtx) {
				m.requestClose()
			}
		}
	}
}

func (m ModalWidget) requestClose() {
	if m.onOpenChange != nil {
		m.onOpenChange(false)
	}
}

func (m ModalWidget) isDismissable() bool {
	if !m.hasDismissable {
		return true
	}
	return m.dismissable
}

func (m ModalWidget) showCloseButton() bool {
	if !m.hasCloseButton {
		return true
	}
	return m.closeButton
}

func (m ModalWidget) resolvedAnimation() ModalAnimation {
	if m.animation == ModalAnimationAuto {
		return ModalAnimationFadeScale
	}
	return m.animation
}
