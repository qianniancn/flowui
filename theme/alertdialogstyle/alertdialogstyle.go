// Package alertdialogstyle contains the theme-owned visual recipe for alert
// dialog headers. Modal overlay behavior remains in widget/modal.
package alertdialogstyle

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildAlertDialog builds the provider-owned status header visual plan.
func BuildAlertDialog(active *theme.Theme, request theme.AlertDialogRequest) theme.AlertDialogVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.AlertDialog
	sem := active.Semantic()
	background := sem.SurfaceSecondary
	foreground := sem.OverlayForeground
	if tone, tinted := statusTone(active, request.Status); tinted {
		background = tone.Soft
		foreground = tone.OnSoft
	}
	return theme.AlertDialogVisual{
		Title: style.Style{}.
			FontSize(tokens.TitleSize).
			FontWeight(int(font.Medium)).
			TextColor(style.Solid(sem.OverlayForeground)),
		IconBackground: background,
		IconForeground: foreground,
		Override:       theme.ModalOverrideStyle(theme.ModalRequest{Override: request.Override}),
		Metrics: theme.AlertDialogMetrics{
			IconSize:      tokens.IconSize,
			IconGlyphSize: tokens.IconGlyphSize,
			HeaderGap:     tokens.HeaderGap,
		},
		Icon:      statusIcon(request.Status),
		IconFrame: iconFrameRenderer(background),
	}
}

func iconFrameRenderer(background color.NRGBA) theme.AlertDialogIconFrameRenderer {
	return func(gtx layout.Context, request theme.AlertDialogIconFrameRequest) layout.Dimensions {
		if request.Size.X <= 0 || request.Size.Y <= 0 {
			return layout.Dimensions{}
		}
		gtx.Constraints = layout.Exact(request.Size)
		radius := min(request.Size.X, request.Size.Y) / 2
		paint.FillShape(gtx.Ops, background, clip.UniformRRect(image.Rectangle{Max: request.Size}, radius).Op(gtx.Ops))
		stack := clip.UniformRRect(image.Rectangle{Max: request.Size}, radius).Push(gtx.Ops)
		defer stack.Pop()
		if request.Content == nil {
			return layout.Dimensions{Size: request.Size}
		}
		dims := layout.Center.Layout(gtx, request.Content)
		return layout.Dimensions{Size: request.Size, Baseline: dims.Baseline}
	}
}

func statusTone(active *theme.Theme, status theme.AlertStatus) (appearance.ToneColors, bool) {
	switch status {
	case theme.AlertAccent:
		return appearance.Tone(active, theme.TonePrimary), true
	case theme.AlertSuccess:
		return appearance.Tone(active, theme.ToneSuccess), true
	case theme.AlertWarning:
		return appearance.Tone(active, theme.ToneWarning), true
	case theme.AlertDanger:
		return appearance.Tone(active, theme.ToneDanger), true
	default:
		return appearance.Tone(active, theme.ToneDefault), false
	}
}

func statusIcon(status theme.AlertStatus) []byte {
	switch status {
	case theme.AlertSuccess:
		return lucide.CircleCheck
	case theme.AlertWarning:
		return lucide.TriangleAlert
	case theme.AlertDanger:
		return lucide.CircleAlert
	default:
		return lucide.Info
	}
}
