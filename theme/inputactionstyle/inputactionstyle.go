// Package inputactionstyle contains the theme-owned recipe for compact
// input-group actions: the clear buttons, search icons, and suffix controls a
// field group renders inside its slot.
package inputactionstyle

import (
	"time"

	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const (
	actionSize     = unit.Dp(24)
	actionDuration = 100 * time.Millisecond
)

// BuildInputAction returns the provider-owned declaration and neutral slot
// metric for a compact input-group action.
func BuildInputAction(activeTheme *theme.Theme, request theme.InputActionRequest) theme.InputActionVisual {
	if activeTheme == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.InputActionVisual{
		Declaration: style.Join(actionRules(activeTheme), theme.InputActionOverrideStyle(request)),
		Metrics: theme.InputActionMetrics{
			SlotPadding: activeTheme.Components.InputGroup.ActionPadding,
		},
	}
}

func actionRules(activeTheme *theme.Theme) style.Style {
	if activeTheme == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	sem := activeTheme.Semantic()
	neutral := appearance.ToneColorsOf(sem, theme.ToneDefault)
	return style.Style{}.
		MinWidth(actionSize).
		Height(actionSize).
		PaddingX(4).
		Radius(4).
		TextColor(style.Solid(sem.MutedForeground)).
		Cursor(pointer.CursorPointer).
		Outline(unit.Dp(2), 0, style.WithAlpha(sem.Ring, 0)).
		Transition(style.PropBackgroundColor, actionDuration).
		Transition(style.PropOutlineColor, actionDuration).
		Transition(style.PropTextColor, actionDuration).
		Rules(
			style.Rule{
				When: style.Hovered,
				Set: style.Style{}.
					Background(style.Solid(neutral.Hover)).
					TextColor(style.Solid(sem.Foreground)),
			},
			style.Rule{
				When: style.Pressed,
				Set:  style.Style{}.Background(style.Solid(neutral.Fill)),
			},
			style.Rule{
				When: style.FocusVisible,
				Not:  style.Disabled,
				Set:  style.Style{}.Outline(2, 0, style.Solid(sem.Ring)),
			},
			style.Rule{
				When: style.Disabled,
				Set: style.Style{}.
					Cursor(pointer.CursorDefault).
					Opacity(0.45),
			},
		)
}
