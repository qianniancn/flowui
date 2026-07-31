package ui

import (
	"github.com/qianniancn/flowui/internal/components/modal"
	"github.com/qianniancn/flowui/internal/components/popover"
	"github.com/qianniancn/flowui/internal/overlay"
)

// PopoverPlacement identifies the side and alignment used for a popover.
type PopoverPlacement = overlay.PopoverPlacement

type PopoverWidget = popover.PopoverWidget

type ModalWidget = modal.ModalWidget

// ModalSize controls the width of a modal.
type ModalSize = modal.ModalSize

// ModalPlacement controls where a modal is placed in the window.
type ModalPlacement = modal.ModalPlacement

// ModalBackdropVariant controls how the area behind a modal is drawn.
type ModalBackdropVariant = modal.ModalBackdropVariant

// ModalScroll controls where modal overflow is handled.
type ModalScroll = modal.ModalScroll

// ModalAnimation selects the modal enter and exit animation.
type ModalAnimation = modal.ModalAnimation

const (
	PopoverBottom      = overlay.PopoverBottom
	PopoverTop         = overlay.PopoverTop
	PopoverLeft        = overlay.PopoverLeft
	PopoverRight       = overlay.PopoverRight
	PopoverBottomStart = overlay.PopoverBottomStart
	PopoverBottomEnd   = overlay.PopoverBottomEnd
	PopoverTopStart    = overlay.PopoverTopStart
	PopoverTopEnd      = overlay.PopoverTopEnd
	PopoverLeftStart   = overlay.PopoverLeftStart
	PopoverLeftEnd     = overlay.PopoverLeftEnd
	PopoverRightStart  = overlay.PopoverRightStart
	PopoverRightEnd    = overlay.PopoverRightEnd

	ModalMedium = modal.ModalMedium
	ModalXSmall = modal.ModalXSmall
	ModalSmall  = modal.ModalSmall
	ModalLarge  = modal.ModalLarge
	ModalCover  = modal.ModalCover
	ModalFull   = modal.ModalFull

	ModalAuto   = modal.ModalAuto
	ModalTop    = modal.ModalTop
	ModalCenter = modal.ModalCenter
	ModalBottom = modal.ModalBottom

	ModalBackdropOpaque      = modal.ModalBackdropOpaque
	ModalBackdropBlur        = modal.ModalBackdropBlur
	ModalBackdropTransparent = modal.ModalBackdropTransparent

	ModalScrollInside  = modal.ModalScrollInside
	ModalScrollOutside = modal.ModalScrollOutside

	ModalAnimationAuto        = modal.ModalAnimationAuto
	ModalAnimationFadeScale   = modal.ModalAnimationFadeScale
	ModalAnimationSlideDown   = modal.ModalAnimationSlideDown
	ModalAnimationSlideUp     = modal.ModalAnimationSlideUp
	ModalAnimationFade        = modal.ModalAnimationFade
	ModalAnimationBounceScale = modal.ModalAnimationBounceScale
	ModalAnimationZoomOut     = modal.ModalAnimationZoomOut
	ModalAnimationPop         = modal.ModalAnimationPop
)

// Popover creates a positioned popup attached to trigger.
func Popover(key string, open bool, trigger Widget, content Widget) PopoverWidget {
	return popover.Popover(key, open, trigger, content)
}

// Modal creates a modal surface whose visibility is controlled by open.
func Modal(key string, open bool, title string, body Widget) ModalWidget {
	return modal.Modal(key, open, title, body)
}
