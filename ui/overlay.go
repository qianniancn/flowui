package ui

import (
	"github.com/qianniancn/FlowUI/internal/components/modal"
	"github.com/qianniancn/FlowUI/internal/components/popover"
	"github.com/qianniancn/FlowUI/internal/overlay"
)

type PopoverPlacement = overlay.PopoverPlacement
type PopoverWidget = popover.PopoverWidget
type ModalWidget = modal.ModalWidget
type ModalSize = modal.ModalSize
type ModalPlacement = modal.ModalPlacement
type ModalBackdropVariant = modal.ModalBackdropVariant
type ModalScroll = modal.ModalScroll
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

func Popover(key string, open bool, trigger Widget, content Widget) PopoverWidget {
	return popover.Popover(key, open, trigger, content)
}

func Modal(key string, open bool, title string, body Widget) ModalWidget {
	return modal.Modal(key, open, title, body)
}
