package testtheme

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) AlertDialog(active *theme.Theme, request theme.AlertDialogRequest) theme.AlertDialogVisual {
	tokens := active.Components.AlertDialog
	return theme.AlertDialogVisual{
		Title:          style.Style{}.TextColor(style.Solid(foreground)).FontSize(tokens.TitleSize).FontWeight(500),
		IconBackground: background,
		IconForeground: foreground,
		Override:       theme.ModalOverrideStyle(theme.ModalRequest{Override: request.Override}),
		Metrics: theme.AlertDialogMetrics{
			IconSize: tokens.IconSize, IconGlyphSize: tokens.IconGlyphSize,
			HeaderGap: tokens.HeaderGap,
		},
		IconFrame: func(gtx layout.Context, request theme.AlertDialogIconFrameRequest) layout.Dimensions {
			if request.Size.X <= 0 || request.Size.Y <= 0 {
				return layout.Dimensions{}
			}
			gtx.Constraints = layout.Exact(request.Size)
			if request.Content == nil {
				return layout.Dimensions{Size: request.Size}
			}
			return request.Content(gtx)
		},
	}
}
