package modal

import (
	"image/color"

	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
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

func modalDefaultDeclaration(activeTheme *theme.Theme, backdrop ModalBackdropVariant, size ModalSize) flowstyle.Style {
	tokens := activeTheme.Components.Modal
	backdropColor := modalStyleFor(activeTheme, backdrop, size).backdrop
	builder := flowstyle.Style{}.
		Background(flowstyle.TokenOverlay).
		TextColor(flowstyle.TokenOverlayForeground).
		Padding(tokens.Padding).
		Overflow(flowstyle.OverflowHidden).
		Part(flowstyle.PartBackdrop, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: backdropColor}))
	if size != ModalFull {
		builder = builder.Radius(tokens.Radius).Shadow(flowstyle.ShadowOverlay)
	}
	return builder
}
