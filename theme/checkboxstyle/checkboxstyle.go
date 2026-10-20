// Package checkboxstyle owns the visual recipe shared by checkbox providers.
// It maps theme tokens to declarations; interaction and semantic state remain
// in widget/checkbox.
package checkboxstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildCheckbox builds the provider-owned checkbox declaration and metrics.
func BuildCheckbox(active *theme.Theme, request theme.CheckboxRequest) theme.CheckboxVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	recipe := ruleSet(active, request.Variant)
	indicator := recipe[style.PartIndicator]
	delete(recipe, style.PartIndicator)
	return theme.CheckboxVisual{
		Defaults: recipe.Compile(),
		Variant:  style.Recipe{style.PartIndicator: indicator}.Compile(),
		Override: theme.CheckboxOverrideStyle(request),
		Metrics: theme.CheckboxMetrics{
			FocusSpace:          active.Components.Checkbox.FocusSpace,
			IndicatorSize:       active.Components.Checkbox.IndicatorSize,
			IndeterminateStroke: active.Components.Checkbox.IndeterminateStroke,
			CheckStroke:         active.Components.Checkbox.CheckStroke,
			DescriptionGap:      active.Components.Checkbox.DescriptionGap,
			DescriptionIndent:   active.Components.Checkbox.DescriptionIndent,
			LabelGap:            active.Components.Checkbox.LabelGap,
			RequiredMarkOffset:  active.Components.Label.RequiredMarkOffset,
		},
		Colors:            checkboxColors(active),
		IndicatorRenderer: checkboxIndicatorRenderer,
		Motion:            active.Motion,
		SelectionDuration: 200 * time.Millisecond,
		FocusDuration:     100 * time.Millisecond,
	}
}

// checkboxColors returns the semantic colors used by the required mark.
func checkboxColors(active *theme.Theme) theme.CheckboxColors {
	sem := active.Semantic()
	value := sem.Tone(theme.ToneDanger).Fill
	return theme.CheckboxColors{
		RequiredMark:         value,
		DisabledRequiredMark: active.DisabledColor(value),
	}
}

// palette contains the resolved semantic colors used by the checkbox recipe.
type palette struct {
	Bg, Border, Accent, AccentHover, AccentFg, Fg, FocusColor color.NRGBA
	Shadow                                                    float32
}

// colors resolves checkbox colors from the active theme and state.
func colors(active *theme.Theme, variant theme.CheckboxVariant, hovered, pressed, disabled, invalid bool) palette {
	sem := active.Semantic()
	danger := sem.Tone(theme.ToneDanger).Fill
	dangerHover := sem.Tone(theme.ToneDanger).Hover
	dangerFg := sem.Tone(theme.ToneDanger).OnFill
	value := palette{
		Bg: sem.Surface, Border: sem.Border,
		Accent:      sem.Tone(theme.TonePrimary).Fill,
		AccentHover: sem.Tone(theme.TonePrimary).Hover,
		AccentFg:    sem.Tone(theme.TonePrimary).OnFill,
		Fg:          sem.Foreground, FocusColor: sem.Ring,
		Shadow: active.Components.Checkbox.ShadowOpacity,
	}
	if variant == theme.CheckboxSecondary {
		value.Bg = sem.SurfaceTertiary
		value.Shadow = 0
	}
	if invalid {
		value.Border, value.Accent = danger, danger
		value.AccentHover, value.AccentFg = dangerHover, dangerFg
		value.FocusColor = danger
		value.FocusColor.A = sem.Ring.A
	}
	if disabled {
		value.Bg = active.DisabledColor(value.Bg)
		value.Border = active.DisabledColor(value.Border)
		value.Accent = active.DisabledColor(value.Accent)
		value.AccentHover = active.DisabledColor(value.AccentHover)
		value.AccentFg = active.DisabledColor(value.AccentFg)
		value.Fg = active.DisabledColor(value.Fg)
		value.FocusColor = color.NRGBA{}
		value.Shadow = 0
		return value
	}
	if hovered || pressed {
		if invalid {
			value.Border = dangerHover
		} else {
			value.Border = sem.SurfacePressed
		}
		value.Accent = value.AccentHover
	}
	return value
}

// ruleSet returns the state-aware checkbox recipe.
func ruleSet(active *theme.Theme, variant theme.CheckboxVariant) style.Recipe {
	tokens := active.Components.Checkbox
	ringWidth := tokens.FocusRingWidth
	base := colors(active, variant, false, false, false, false)
	hover := colors(active, variant, true, false, false, false)
	invalid := colors(active, variant, false, false, false, true)
	invalidHover := colors(active, variant, true, false, false, true)
	disabled := colors(active, variant, false, false, true, false)
	disabledInvalid := colors(active, variant, false, false, true, true)
	selected := func(p palette) style.Style {
		return style.Style{}.OnAny(style.Checked|style.Indeterminate, style.Style{}.
			Background(style.Solid(p.Accent)).
			BorderColor(style.Solid(p.Accent)).
			TextColor(style.Solid(p.AccentFg)))
	}
	indicator := style.Style{}.
		Width(tokens.Size).Height(tokens.Size).
		BorderWidth(tokens.BorderWidth).
		Background(style.Solid(base.Bg)).
		BorderColor(style.Solid(base.Border)).
		Radius(tokens.Radius).
		Outline(ringWidth, 0, style.Solid(base.FocusColor)).
		TextColor(style.Solid(base.Fg))
	shadowRow := indicatorShadow(style.Style{}, active, base.Shadow)
	row := func(p palette) style.Style {
		return style.Style{}.
			BorderColor(style.Solid(p.Border)).
			Outline(ringWidth, 0, style.Solid(p.FocusColor)).
			TextColor(style.Solid(p.Fg)).
			OnAny(style.Checked|style.Indeterminate, style.Style{}.
				Background(style.Solid(p.Accent)).
				BorderColor(style.Solid(p.Accent)).
				TextColor(style.Solid(p.AccentFg)))
	}
	disabledPaint := style.Join(
		style.Style{}.
			Background(style.Solid(disabled.Bg)).
			BorderColor(style.Solid(disabled.Border)).
			Outline(ringWidth, 0, style.Solid(disabled.FocusColor)).
			TextColor(style.Solid(disabled.Fg)),
		selected(disabled),
	)
	disabledRow := disabledPaint.On(style.Invalid, style.Join(
		style.Style{}.
			Background(style.Solid(disabledInvalid.Bg)).
			BorderColor(style.Solid(disabledInvalid.Border)).
			Outline(ringWidth, 0, style.Solid(disabledInvalid.FocusColor)).
			TextColor(style.Solid(disabledInvalid.Fg)),
		selected(disabledInvalid),
	))
	sem := active.Semantic()
	return style.Recipe{
		style.PartIndicator: {
			Base: indicator,
			Rules: []style.Rule{
				{Not: style.Disabled, Set: shadowRow},
				{Set: selected(base)},
				{When: style.Invalid, Set: row(invalid)},
				{When: style.Hovered, Not: style.Invalid, Set: row(hover)},
				{When: style.Pressed, Not: style.Invalid, Set: row(hover)},
				{When: style.Invalid | style.Hovered, Set: row(invalidHover)},
				{When: style.Invalid | style.Pressed, Set: row(invalidHover)},
				{Rank: 1, When: style.Disabled, Set: disabledRow},
			},
		},
		style.PartLabel: {
			Base: style.Style{}.
				TextColor(style.Solid(base.Fg)).
				FontSize(active.Typography.ControlSize).
				FontWeight(int(font.Medium)),
			Rules: []style.Rule{{Rank: 1, When: style.Disabled, Set: style.Style{}.
				TextColor(style.Solid(disabled.Fg))}},
		},
		style.PartDescription: {
			Base: style.Style{}.TextColor(style.Solid(sem.MutedForeground)),
			Rules: []style.Rule{
				{When: style.Invalid, Set: style.Style{}.
					TextColor(style.Solid(sem.Tone(theme.ToneDanger).Fill))},
				{Rank: 1, When: style.Disabled, Set: style.Style{}.
					TextColor(style.Solid(active.DisabledColor(sem.MutedForeground))).
					On(style.Invalid, style.Style{}.
						TextColor(style.Solid(active.DisabledColor(sem.Tone(theme.ToneDanger).Fill))))},
			},
		},
	}
}

func indicatorShadow(declaration style.Style, active *theme.Theme, shadowOpacity float32) style.Style {
	shadowColor := active.Semantic().SurfaceShadow
	for _, layer := range active.Shadows.Checkbox.Layers {
		col := shadowColor
		col.A = byte(float32(col.A)*min(max(layer.Opacity*shadowOpacity, 0), 1) + .5)
		if layer.Blur >= 0 && col.A != 0 {
			declaration = declaration.BoxShadow(layer.OffsetX, layer.OffsetY, layer.Blur, layer.Spread, style.Solid(col))
		}
	}
	return declaration
}
