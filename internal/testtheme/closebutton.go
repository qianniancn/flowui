package testtheme

import (
	"image/color"
	"sync"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

var neutralCloseIcon struct {
	sync.Once
	value *widget.Icon
}

func neutralCloseIconRenderer(gtx layout.Context, col color.NRGBA) layout.Dimensions {
	neutralCloseIcon.Do(func() {
		value, err := widget.NewIcon(lucide.X)
		if err == nil {
			neutralCloseIcon.value = value
		}
	})
	if neutralCloseIcon.value == nil {
		return layout.Dimensions{}
	}
	return neutralCloseIcon.value.Layout(gtx, col)
}

func (behaviorProvider) CloseButton(active *theme.Theme, request theme.CloseButtonRequest) theme.CloseButtonVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.CloseButton
	sem := active.Semantic()
	return theme.CloseButtonVisual{
		Defaults: style.Style{}.
			Width(tokens.Size).
			Height(tokens.Size).
			Padding(tokens.Padding).
			Background(style.Solid(sem.SurfaceRaised)).
			TextColor(style.Solid(sem.MutedForeground)).
			Radius(tokens.Radius).
			Outline(tokens.FocusRingWidth, 0, style.WithAlpha(sem.Ring, 0)).
			Opacity(1).
			Cursor(pointer.CursorPointer).
			OnNot(style.Hovered, style.Disabled, style.Style{}.Background(style.Solid(sem.SurfacePressed))).
			On(style.Pressed, style.Style{}.Scale(tokens.PressedScale)).
			OnNot(style.FocusVisible, style.Disabled, style.Style{}.Outline(tokens.FocusRingWidth, 0, style.Solid(sem.Ring))).
			On(style.Disabled, style.Style{}.
				Background(style.Solid(active.DisabledColor(sem.SurfaceRaised))).
				TextColor(style.Solid(active.DisabledColor(sem.MutedForeground))).
				Cursor(pointer.CursorDefault)),
		Icon: style.Style{}.
			Part(style.PartIcon, style.Style{}.
				Width(tokens.IconSize).
				Height(tokens.IconSize)),
		Override:     theme.CloseButtonOverrideStyle(request),
		Motion:       active.Motion,
		IconRenderer: neutralCloseIconRenderer,
	}
}
