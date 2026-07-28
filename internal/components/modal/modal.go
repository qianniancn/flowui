package modal

import (
	"image"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

// ModalWidget renders an overlay dialog controlled by application state.
type ModalWidget struct {
	key                     string
	open                    bool
	hasOpen                 bool
	defaultOpen             bool
	hasDefaultOpen          bool
	title                   string
	body                    frame.Widget
	header                  frame.Widget
	footer                  frame.Widget
	icon                    frame.Widget
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
	customStyle             flowstyle.Style
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

// Modal creates a modal dialog. When open is true, it starts in controlled
// mode with the modal open. When open is false, it starts in uncontrolled mode
// with the modal closed (use DefaultOpen(true) for uncontrolled mode starting open).
func Modal(key string, open bool, title string, body frame.Widget) ModalWidget {
	if open {
		// open=true means controlled mode, immediately open
		return ModalWidget{
			key:     key,
			open:    true,
			hasOpen: true,
			title:   title,
			body:    body,
		}
	}
	// open=false means uncontrolled mode, initially closed
	return ModalWidget{
		key:   key,
		title: title,
		body:  body,
	}
}

// OnOpenChange registers a callback for close requests.
func (m ModalWidget) OnOpenChange(fn func(bool)) ModalWidget {
	m.onOpenChange = fn
	return m
}

// Open sets the modal to controlled mode with the given open state.
func (m ModalWidget) Open(open bool) ModalWidget {
	m.open = open
	m.hasOpen = true
	return m
}

// DefaultOpen sets the initial open state for uncontrolled mode.
func (m ModalWidget) DefaultOpen(open bool) ModalWidget {
	m.defaultOpen = open
	m.hasDefaultOpen = true
	return m
}

// Header replaces the default title/header area.
func (m ModalWidget) Header(header frame.Widget) ModalWidget {
	m.header = header
	return m
}

// Body replaces the modal body content.
func (m ModalWidget) Body(body frame.Widget) ModalWidget {
	m.body = body
	return m
}

// Footer sets the modal footer content.
func (m ModalWidget) Footer(footer frame.Widget) ModalWidget {
	m.footer = footer
	return m
}

// Icon sets an optional icon shown above the default title.
func (m ModalWidget) Icon(icon frame.Widget) ModalWidget {
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
func (m ModalWidget) Style(value flowstyle.Style) ModalWidget {
	m.customStyle = value
	return m
}

func (m ModalWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	fullKey := frame.FullKey(ctx, m.key)
	naturallyDisabled := frame.OverlayNaturallyDisabled(gtx)

	// Check if we need to evaluate open state (peek first to avoid claiming state)
	needsState := hasVisibleModal(ctx, fullKey)
	if !needsState {
		// For uncontrolled mode, check if we have a default that would open it
		if !m.hasOpen && m.hasDefaultOpen && m.defaultOpen {
			needsState = true
		} else if m.hasOpen && m.open {
			needsState = true
		}
	}

	if !needsState {
		return layout.Dimensions{}
	}

	// Now we know we need state, claim it
	state := modalStateFor(ctx, m.key)
	state.bind(m)
	open := state.isOpen(m)

	if !open && !state.visible() {
		deleteModalState(ctx, fullKey)
		return layout.Dimensions{}
	}

	progress := state.progress(gtx, open, frame.ActiveTheme(ctx).Motion)
	if open && naturallyDisabled {
		state.focusPending = true
	}
	if !open && progress <= 0 {
		deleteModalState(ctx, fullKey)
		return layout.Dimensions{}
	}
	var tail func(layout.Context)
	if !naturallyDisabled {
		tail = func(gtx layout.Context) {
			m.layoutFocusEnd(gtx, state)
		}
	}
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:      fullKey,
		Layer:    frame.OverlayLayerModal,
		Disabled: naturallyDisabled,
		Layout: func(gtx layout.Context, _ image.Rectangle, interactive bool) layout.Dimensions {
			if interactive && gtx.Enabled() {
				if open {
					m.handleCloseEvents(ctx, gtx, state)
				} else {
					m.handleExitSurfaceEvents(ctx, gtx, state)
				}
			}
			return m.layoutOverlay(ctx, gtx, state, progress, open && gtx.Enabled())
		},
		Tail: tail,
	})
	if open && !naturallyDisabled {
		frame.AfterOverlays(ctx, func() {
			becameTopmost := frame.OverlayFocusScopeBecameTopmost(ctx, frame.OverlayLayerModal, fullKey)
			if frame.OverlayFocusScopeTopmost(ctx, frame.OverlayLayerModal, fullKey) && (state.focusPending || becameTopmost) {
				frame.RequestFocus(ctx, state.initialFocusTag())
				state.focusPending = false
			}
		})
	} else if !naturallyDisabled {
		frame.AfterOverlays(ctx, func() {
			if frame.OverlayFocusScopeTopmost(ctx, frame.OverlayLayerModal, fullKey) {
				frame.RequestFocus(ctx, state.initialFocusTag())
			}
		})
	}

	return layout.Dimensions{}
}

func (m ModalWidget) handleExitSurfaceEvents(ctx *frame.Context, gtx layout.Context, state *modalState) {
	for state.dialog.Clicked(gtx) {
	}
	pressed := state.dialog.TakePressed()
	for i := range state.dismiss {
		for state.dismiss[i].Clicked(gtx) {
		}
		pressed = state.dismiss[i].TakePressed() || pressed
	}
	if pressed {
		frame.RequestFocus(ctx, state.endFocusTag())
	}
}

func (m ModalWidget) handleCloseEvents(ctx *frame.Context, gtx layout.Context, modalStateValue *modalState) {
	for modalStateValue.dialog.Clicked(gtx) {
		frame.RequestFocus(ctx, modalStateValue.endFocusTag())
	}
	if modalStateValue.dialog.TakePressed() {
		frame.RequestFocus(ctx, modalStateValue.endFocusTag())
	}
	if !m.keyboardDismissDisabled && modalStateValue.escapePressed(gtx) {
		m.requestClose(modalStateValue)
	}
	if m.showCloseButton() {
		presses := state.SnapshotPresses(modalStateValue.close.History())
		for modalStateValue.close.Clicked(gtx) {
			m.requestClose(modalStateValue)
			frame.RequestFocusVisible(ctx, &modalStateValue.close, presses.ClickFocusVisible(modalStateValue.close.History()))
		}
		frame.FocusOnPress(ctx, &modalStateValue.close, modalStateValue.close.History(), presses.Active())
	}
	for i := range modalStateValue.dismiss {
		for modalStateValue.dismiss[i].Clicked(gtx) {
			if m.isDismissable() {
				m.requestClose(modalStateValue)
			}
		}
		if modalStateValue.dismiss[i].TakePressed() {
			frame.RequestFocus(ctx, modalStateValue.endFocusTag())
		}
	}
}

func (m ModalWidget) requestClose(state *modalState) {
	state.requestOpen(m, false)
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
