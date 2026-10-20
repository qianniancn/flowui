// Package switchstyle contains the shared visual recipe for switch behavior.
// It owns declarations and theme-token mapping; interaction state and Gio
// event handling remain in widget/switchcontrol.
package switchstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildSwitch builds the provider-owned switch declaration and metrics.
func BuildSwitch(active *theme.Theme, request theme.SwitchRequest) theme.SwitchVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.SwitchVisual{
		Defaults: switchRuleSet(active, request.Size).Compile(),
		Override: theme.SwitchOverrideStyle(request),
		Metrics: theme.SwitchMetrics{
			FocusSpace:     active.Components.Switch.FocusSpace,
			ContentGap:     active.Components.Switch.ContentGap,
			DescriptionGap: active.Components.Switch.DescriptionGap,
		},
		Motion:            active.Motion,
		SelectionDuration: 300 * time.Millisecond,
		FocusDuration:     100 * time.Millisecond,
	}
}

// BuildSwitchGroup builds the provider-owned switch group declaration and
// spacing.
func BuildSwitchGroup(active *theme.Theme, request theme.SwitchGroupRequest) theme.SwitchGroupVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.SwitchGroupVisual{
		Metrics: theme.SwitchGroupMetrics{
			VerticalGap:   active.Components.SwitchGroup.VerticalGap,
			HorizontalGap: active.Components.SwitchGroup.HorizontalGap,
		},
		Override: theme.SwitchOverrideStyle(theme.SwitchRequest{Override: request.Override}),
	}
}

// switchPalette contains semantic switch colors for a given interaction row.
type switchPalette struct {
	TrackOff, TrackOn  color.NRGBA
	Thumb, ThumbOn     color.NRGBA
	ThumbForegroundOff color.NRGBA
	ThumbForeground    color.NRGBA
	Focus              color.NRGBA
}

// switchColors resolves switch colors from the active theme and state.
func switchColors(active *theme.Theme, hovered, pressed, disabled, invalid bool) switchPalette {
	sem := active.Semantic()
	value := switchPalette{
		TrackOff:           sem.SurfaceRaised,
		TrackOn:            sem.Tone(theme.TonePrimary).Fill,
		Thumb:              sem.Surface,
		ThumbOn:            sem.Tone(theme.TonePrimary).OnFill,
		ThumbForegroundOff: sem.MutedForeground,
		ThumbForeground:    sem.Tone(theme.TonePrimary).Fill,
		Focus:              sem.Ring,
	}
	if hovered || pressed {
		value.TrackOff = sem.SurfacePressed
		value.TrackOn = sem.Tone(theme.TonePrimary).Hover
	}
	if invalid {
		value.TrackOn = sem.Tone(theme.ToneDanger).Fill
		value.ThumbForeground = sem.Tone(theme.ToneDanger).Fill
		value.Focus = sem.Tone(theme.ToneDanger).Fill
		value.Focus.A = sem.Ring.A
		if hovered || pressed {
			value.TrackOn = sem.Tone(theme.ToneDanger).Hover
		}
	}
	if disabled {
		value.TrackOff = active.DisabledColor(value.TrackOff)
		value.TrackOn = active.DisabledColor(value.TrackOn)
		value.Thumb = active.DisabledColor(value.Thumb)
		value.ThumbOn = active.DisabledColor(value.ThumbOn)
		value.ThumbForegroundOff = active.DisabledColor(value.ThumbForegroundOff)
		value.ThumbForeground = active.DisabledColor(value.ThumbForeground)
		value.Focus = color.NRGBA{}
	}
	return value
}

type switchSizeStyle struct {
	trackWidth  unit.Dp
	trackHeight unit.Dp
	thumbWidth  unit.Dp
	thumbHeight unit.Dp
}

func switchSize(active *theme.Theme, size theme.SwitchSize) switchSizeStyle {
	tokens := active.Components.Switch
	switch size {
	case theme.SwitchSmall:
		return switchSizeStyle{tokens.SmallTrackWidth, tokens.SmallTrackHeight, tokens.SmallThumbWidth, tokens.SmallThumbHeight}
	case theme.SwitchLarge:
		return switchSizeStyle{tokens.LargeTrackWidth, tokens.LargeTrackHeight, tokens.LargeThumbWidth, tokens.LargeThumbHeight}
	default:
		return switchSizeStyle{tokens.MediumTrackWidth, tokens.MediumTrackHeight, tokens.MediumThumbWidth, tokens.MediumThumbHeight}
	}
}

// switchRuleSet returns the state-aware switch recipe.
func switchRuleSet(active *theme.Theme, size theme.SwitchSize) style.Recipe {
	sizes := switchSize(active, size)
	rest := switchColors(active, false, false, false, false)
	hovered := switchColors(active, true, false, false, false)
	invalid := switchColors(active, false, false, false, true)
	invalidHovered := switchColors(active, true, false, false, true)
	dim := active.DisabledOpacityValue()
	background := func(value style.ColorSource) style.Style { return style.Style{}.Background(value) }
	track := style.Style{}.
		Width(sizes.trackWidth).
		Height(sizes.trackHeight).
		Background(style.Solid(rest.TrackOff)).
		Radius(sizes.trackHeight/2).
		Outline(active.Components.Switch.FocusRingWidth, 0, style.Solid(rest.Focus))
	thumb := switchThumbShadows(style.Style{}.
		Width(sizes.thumbWidth).
		Height(sizes.thumbHeight).
		Background(style.Solid(rest.Thumb)).
		Radius(sizes.thumbHeight/2).
		TextColor(style.Solid(rest.ThumbForegroundOff)), active)
	label := style.Style{}.
		TextColor(style.Solid(active.Semantic().Foreground)).
		FontSize(active.Components.Switch.TextSize).
		FontWeight(int(font.Medium))
	description := style.Style{}.
		TextColor(style.Solid(active.Semantic().MutedForeground)).
		FontSize(active.Components.Description.TextSize)
	return style.Recipe{
		style.PartTrack: {
			Base: track,
			Rules: []style.Rule{
				{When: style.Hovered, Not: style.Checked, Set: background(style.Solid(hovered.TrackOff))},
				{When: style.Pressed, Not: style.Checked, Set: background(style.Solid(hovered.TrackOff))},
				{When: style.Checked, Set: background(style.Solid(rest.TrackOn))},
				{When: style.Checked | style.Hovered, Set: background(style.Solid(hovered.TrackOn))},
				{When: style.Checked | style.Pressed, Set: background(style.Solid(hovered.TrackOn))},
				{When: style.Checked | style.Invalid, Set: background(style.Solid(invalid.TrackOn))},
				{When: style.Checked | style.Invalid | style.Hovered, Set: background(style.Solid(invalidHovered.TrackOn))},
				{When: style.Checked | style.Invalid | style.Pressed, Set: background(style.Solid(invalidHovered.TrackOn))},
				{When: style.Invalid, Not: style.Disabled, Set: style.Style{}.Outline(active.Components.Switch.FocusRingWidth, 0, style.Solid(invalid.Focus))},
				{Rank: 1, When: style.Disabled, Set: style.Style{}.Dim(dim).Outline(active.Components.Switch.FocusRingWidth, 0, style.SolidColor{})},
			},
		},
		style.PartThumb: {
			Base: thumb,
			Rules: []style.Rule{
				{When: style.Checked, Set: style.Style{}.Background(style.Solid(rest.ThumbOn)).TextColor(style.Solid(rest.ThumbForeground))},
				{When: style.Checked | style.Invalid, Set: style.Style{}.TextColor(style.Solid(invalid.ThumbForeground))},
				{Rank: 1, When: style.Disabled, Set: style.Style{}.Dim(dim)},
			},
		},
		style.PartLabel: {
			Base:  label,
			Rules: []style.Rule{{Rank: 1, When: style.Disabled, Set: style.Style{}.Dim(dim)}},
		},
		style.PartDescription: {
			Base:  description,
			Rules: []style.Rule{{Rank: 1, When: style.Disabled, Set: style.Style{}.Dim(dim)}},
		},
	}
}

func switchThumbShadows(thumb style.Style, active *theme.Theme) style.Style {
	shadowColor := active.Semantic().SurfaceShadow
	for _, layer := range active.Shadows.SwitchThumb.Layers {
		col := shadowColor
		col.A = byte(float32(col.A)*min(max(layer.Opacity, 0), 1) + .5)
		if layer.Blur >= 0 && col.A != 0 {
			thumb = thumb.BoxShadow(layer.OffsetX, layer.OffsetY, layer.Blur, layer.Spread, style.Solid(col))
		}
	}
	return thumb
}
