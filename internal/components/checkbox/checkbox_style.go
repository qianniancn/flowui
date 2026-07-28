package checkbox

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

type checkboxResolvedStyle struct {
	indicator   checkboxIndicatorStyle
	label       flowstyle.ResolvedStyle
	description flowstyle.ResolvedStyle
}

type checkboxIndicatorStyle struct {
	off flowstyle.ResolvedStyle
	on  flowstyle.ResolvedStyle
}

func (c CheckboxWidget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState) checkboxResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	visual := checkboxStyleFor(activeTheme, c.variant, state.Hovered, state.Pressed, state.Disabled, state.Invalid)
	defaults := checkboxStyleDeclaration(activeTheme, visual, state)
	offState := state
	offState.Checked, offState.Selected, offState.Indeterminate = false, false, false
	onState := state
	onState.Checked, onState.Selected = true, true

	off := styleruntime.ResolvePartStatic(ctx, flowstyle.PartIndicator, offState, defaults, flowstyle.Style{}, flowstyle.Style{}, c.customStyle)
	on := styleruntime.ResolvePartStatic(ctx, flowstyle.PartIndicator, onState, defaults, flowstyle.Style{}, flowstyle.Style{}, c.customStyle)
	label := styleruntime.ResolvePartStatic(ctx, flowstyle.PartLabel, state, defaults, flowstyle.Style{}, flowstyle.Style{}, c.customStyle)
	description := styleruntime.ResolvePartStatic(ctx, flowstyle.PartDescription, state, defaults, flowstyle.Style{}, flowstyle.Style{}, c.customStyle)
	offOwner := frame.DerivedKey(ctx, key, "checkbox-indicator-off")
	onOwner := frame.DerivedKey(ctx, key, "checkbox-indicator-on")
	off = styleruntime.ApplyPartTransitions(ctx, gtx, offOwner, flowstyle.PartIndicator, off)
	on = styleruntime.ApplyPartTransitions(ctx, gtx, onOwner, flowstyle.PartIndicator, on)
	label = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartLabel, label)
	description = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartDescription, description)

	return checkboxResolvedStyle{
		indicator:   checkboxIndicatorStyle{off: off, on: on},
		label:       label,
		description: description,
	}
}

func checkboxStyleDeclaration(activeTheme *theme.Theme, value checkboxStyle, state flowstyle.StyleState) flowstyle.Style {
	indicator := flowstyle.Style{}.
		Width(activeTheme.Components.Checkbox.Size).
		Height(activeTheme.Components.Checkbox.Size).
		Background(flowstyle.SolidColor{Color: value.bg}).
		BorderWidth(activeTheme.Components.Checkbox.BorderWidth).
		BorderColor(flowstyle.SolidColor{Color: value.border}).
		Radius(activeTheme.Shape.CheckboxRadius).
		Outline(activeTheme.Components.Checkbox.FocusRingWidth, 0, flowstyle.SolidColor{Color: value.focusColor}).
		TextColor(flowstyle.SolidColor{Color: value.fg}).
		When(flowstyle.Any(flowstyle.Checked, flowstyle.Indeterminate), flowstyle.Style{}.
			Background(flowstyle.SolidColor{Color: value.accent}).
			BorderColor(flowstyle.SolidColor{Color: value.accent}).
			TextColor(flowstyle.SolidColor{Color: value.accentFg}))
	shadowColor := activeTheme.Palette.SurfaceShadow
	for _, layer := range activeTheme.Shadows.Checkbox.Layers {
		col := shadowColor
		col.A = byte(float32(col.A)*min(max(layer.Opacity*value.shadow, 0), 1) + .5)
		if layer.Blur >= 0 && col.A != 0 {
			indicator = indicator.BoxShadow(layer.OffsetX, layer.OffsetY, layer.Blur, layer.Spread, flowstyle.SolidColor{Color: col})
		}
	}
	label := flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: value.fg}).
		FontSize(activeTheme.Typography.ControlSize).
		FontWeight(int(font.Medium))
	descriptionColor := activeTheme.Palette.MutedForeground
	if state.Invalid {
		descriptionColor = activeTheme.Palette.Danger
	}
	if state.Disabled {
		descriptionColor = activeTheme.DisabledColor(descriptionColor)
	}
	return flowstyle.Style{}.
		Part(flowstyle.PartIndicator, indicator).
		Part(flowstyle.PartLabel, label).
		Part(flowstyle.PartDescription, flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: descriptionColor}))

}

func checkboxStyleFor(theme *theme.Theme, variant CheckboxVariant, hovered, pressed, disabled, invalid bool) checkboxStyle {
	danger := theme.Palette.Danger
	dangerHover := theme.Palette.DangerHover
	dangerFg := theme.Palette.DangerForeground
	style := checkboxStyle{
		bg:          theme.Palette.Surface,
		border:      theme.Palette.Border,
		accent:      theme.Palette.Accent,
		accentHover: theme.Palette.AccentHover,
		accentFg:    theme.Palette.AccentForeground,
		fg:          theme.Palette.Foreground,
		focusColor:  theme.Palette.Focus,
		shadow:      theme.Components.Checkbox.ShadowOpacity,
	}
	if variant == CheckboxSecondary {
		style.bg = theme.Palette.SurfaceTertiary
		style.shadow = 0
	}
	if invalid {
		style.border = danger
		style.accent = danger
		style.accentHover = dangerHover
		style.accentFg = dangerFg
		style.focusColor = danger
		style.focusColor.A = theme.Palette.Focus.A
	}
	if disabled {
		style.bg = theme.DisabledColor(style.bg)
		style.border = theme.DisabledColor(style.border)
		style.accent = theme.DisabledColor(style.accent)
		style.accentHover = theme.DisabledColor(style.accentHover)
		style.accentFg = theme.DisabledColor(style.accentFg)
		style.fg = theme.DisabledColor(style.fg)
		style.focusColor = color.NRGBA{}
		style.shadow = 0
		return style
	}
	if hovered || pressed {
		if invalid {
			style.border = dangerHover
		} else {
			style.border = theme.Palette.SurfacePressed
		}
		style.accent = style.accentHover
	}
	return style
}

type checkboxStyle struct {
	bg          color.NRGBA
	border      color.NRGBA
	accent      color.NRGBA
	accentHover color.NRGBA
	accentFg    color.NRGBA
	fg          color.NRGBA
	focusColor  color.NRGBA
	shadow      float32
}

func checkboxLabelColor(activeTheme *theme.Theme, disabled bool) color.NRGBA {
	if disabled {
		return activeTheme.DisabledColor(activeTheme.Palette.Foreground)
	}
	return activeTheme.Palette.Foreground
}

func checkboxSupportingColor(activeTheme *theme.Theme, value color.NRGBA, disabled bool) color.NRGBA {
	if disabled {
		return activeTheme.DisabledColor(value)
	}
	return value
}
