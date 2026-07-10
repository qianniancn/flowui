package modal

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

type modalStyle struct {
	backdrop color.NRGBA
}

func modalStyleFor(theme *theme.Theme, backdrop ModalBackdropVariant, size ModalSize) modalStyle {
	style := modalStyle{
		backdrop: theme.Components.Modal.Backdrop,
	}
	switch backdrop {
	case ModalBackdropTransparent:
		style.backdrop = color.NRGBA{}
	case ModalBackdropBlur:
		style.backdrop = theme.Components.Modal.BlurBackdrop
	}
	if size == ModalFull && backdrop == ModalBackdropTransparent {
		style.backdrop = color.NRGBA{}
	}
	return style
}
