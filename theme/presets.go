package theme

import (
	"image/color"
	"time"

	"github.com/qianniancn/flowui/style"
)

// Action presets provide the four canonical action looks as ready declarations.
// They use the theme's semantic palette. A component Provider can give its
// variants a different treatment; presets remain generic starting values.
//
// A preset is a starting value, not a component. It carries the variant
// colors and the state rows a control needs to read as an action (hover,
// pressed, disabled) plus the color transition the built-in button uses; it
// deliberately leaves geometry (radius, height, padding) and the focus ring
// to the component that adopts it. Presets are values: chain further
// declarations or rows onto them freely.
//
//	t.SecondaryAction().On(style.Loading, style.New().Opacity(0.7))

// actionColorDuration matches the built-in button's color transition.
const actionColorDuration = 100 * time.Millisecond

// actionPalette is the variant-specific color set a preset is built from.
type actionPalette struct {
	bg, hover, pressed, fg color.NRGBA
}

// PrimaryAction returns the accent-filled primary action look.
func (t *Theme) PrimaryAction() style.Style {
	return t.actionPreset(actionPalette{
		bg:      t.Palette.Accent,
		hover:   t.Palette.AccentHover,
		pressed: t.Palette.AccentPressed,
		fg:      t.Palette.AccentForeground,
	})
}

// SecondaryAction returns the soft accent-on-surface secondary look.
func (t *Theme) SecondaryAction() style.Style {
	return t.actionPreset(actionPalette{
		bg:      t.Palette.DefaultColor(),
		hover:   t.Palette.DefaultHoverColor(),
		pressed: t.Palette.DefaultHoverColor(),
		fg:      t.Palette.AccentSoftForeground,
	})
}

// DangerAction returns the filled destructive action look.
func (t *Theme) DangerAction() style.Style {
	return t.actionPreset(actionPalette{
		bg:      t.Palette.Danger,
		hover:   t.Palette.DangerHover,
		pressed: t.Palette.DangerPressed,
		fg:      t.Palette.DangerForeground,
	})
}

// GhostAction returns the transparent quiet action look.
func (t *Theme) GhostAction() style.Style {
	return t.actionPreset(actionPalette{
		bg:      color.NRGBA{},
		hover:   t.Palette.DefaultColor(),
		pressed: t.Palette.DefaultColor(),
		fg:      t.Palette.DefaultForegroundColor(),
	})
}

// actionPreset builds the shared declaration shape: the variant colors on the
// base plus one row per interaction state. The color transition mirrors the
// button's chrome layer so hover feedback animates out of the box.
func (t *Theme) actionPreset(c actionPalette) style.Style {
	return style.New().
		Background(style.Solid(c.bg)).
		TextColor(style.Solid(c.fg)).
		Transition(style.PropBackgroundColor, actionColorDuration).
		On(style.Hovered, style.New().Background(style.Solid(c.hover))).
		On(style.Pressed, style.New().Background(style.Solid(c.pressed))).
		On(style.Disabled, style.New().Opacity(t.DisabledOpacityValue()))
}
