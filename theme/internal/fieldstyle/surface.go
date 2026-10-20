package fieldstyle

import (
	"gioui.org/io/pointer"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func surfaceColors(activeTheme *theme.Theme, variant theme.InputVariant, shadowOpacity float32) (background, hover, focused style.ColorSource, outShadowOpacity float32) {
	surface := appearance.ResolveField(activeTheme, appearance.FieldVariant(variant), shadowOpacity)
	return surface.Background, surface.Hover, surface.Focused, surface.ShadowOpacity
}

func disabledSet(activeTheme *theme.Theme, variant theme.InputVariant) style.Style {
	sem := activeTheme.Semantic()
	if sem.DisabledBackground.A == 0 {
		return style.Style{}.Opacity(activeTheme.DisabledOpacityValue())
	}

	set := style.Style{}.
		TextColor(style.Solid(sem.DisabledForeground)).
		Cursor(pointer.CursorNotAllowed).
		Opacity(1)
	if variant == theme.InputBorderless {
		return set.Background(style.RGBA(0))
	}
	return set.Background(style.Solid(sem.DisabledBackground)).
		BorderColor(style.Solid(sem.DisabledBorder))
}

// readOnlySet is the disabled face without the disabled interaction: a read-only
// field stays focusable and selectable, so it keeps the text cursor and the
// focus ring takes over while it is focused.
func readOnlySet(activeTheme *theme.Theme, variant theme.InputVariant) style.Style {
	sem := activeTheme.Semantic()
	set := style.Style{}.
		Background(style.Solid(sem.DisabledBackground)).
		TextColor(style.Solid(sem.DisabledForeground)).
		BorderColor(style.Solid(sem.DisabledBorder)).
		Cursor(pointer.CursorText).
		Opacity(1)
	if variant == theme.InputBorderless {
		return set.Background(style.RGBA(0))
	}
	return set
}
