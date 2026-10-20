// Package closebuttonstyle contains the theme-owned visual recipe for
// close-button behavior. The widget owns interaction and accessibility; this
// package owns the root/icon declarations and geometry.
package closebuttonstyle

import (
	"image/color"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
)

const (
	colorTransition = 100 * time.Millisecond
	scaleTransition = 250 * time.Millisecond
)

func closeIconRenderer(gtx layout.Context, col color.NRGBA) layout.Dimensions {
	return icon.Layout(lucide.X, gtx, col)
}

// BuildCloseButton builds the complete provider-owned close-button visual.
func BuildCloseButton(active *theme.Theme, request theme.CloseButtonRequest) theme.CloseButtonVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.CloseButton
	sem := active.Semantic()
	pressedScale := tokens.PressedScale
	if pressedScale <= 0 || pressedScale > 1 {
		panic(theme.ErrIncompleteVisualProvider)
	}
	root := style.Style{}.
		Width(tokens.Size).
		Height(tokens.Size).
		Padding(tokens.Padding).
		Background(style.Solid(sem.SurfaceRaised)).
		TextColor(style.Solid(sem.MutedForeground)).
		Radius(tokens.Radius).
		Outline(tokens.FocusRingWidth, 0, style.WithAlpha(sem.Ring, 0)).
		Opacity(1).
		Scale(1).
		Cursor(pointer.CursorPointer).
		Transition(style.PropBackgroundColor, colorTransition).
		Transition(style.PropOutlineColor, colorTransition).
		Transition(style.PropTransform, scaleTransition).
		OnNot(style.Hovered, style.Disabled, style.Style{}.Background(style.Solid(sem.SurfacePressed))).
		On(style.Pressed, style.Style{}.Scale(pressedScale)).
		OnNot(style.FocusVisible, style.Disabled, style.Style{}.Outline(tokens.FocusRingWidth, 0, style.Solid(sem.Ring))).
		On(style.Disabled, style.Style{}.
			Background(style.Solid(active.DisabledColor(sem.SurfaceRaised))).
			TextColor(style.Solid(active.DisabledColor(sem.MutedForeground))).
			Cursor(pointer.CursorDefault))
	return theme.CloseButtonVisual{
		Defaults: root,
		Icon: style.Style{}.
			Part(style.PartIcon, style.Style{}.
				Width(tokens.IconSize).
				Height(tokens.IconSize)),
		Override:     theme.CloseButtonOverrideStyle(request),
		Motion:       active.Motion,
		IconRenderer: closeIconRenderer,
	}
}
