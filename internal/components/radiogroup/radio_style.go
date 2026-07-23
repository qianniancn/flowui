package radiogroup

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type radioStyle struct {
	indicatorOff flowstyle.ResolvedStyle
	indicatorOn  flowstyle.ResolvedStyle
	label        flowstyle.ResolvedStyle
	description  flowstyle.ResolvedStyle
	selected     float32
	focus        float32
	pressed      bool
}

type radioColors struct {
	bg          color.NRGBA
	border      color.NRGBA
	selectedBg  color.NRGBA
	dot         color.NRGBA
	fg          color.NRGBA
	description color.NRGBA
	focusColor  color.NRGBA
}

func (r RadioGroupWidget) resolveItemStyle(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState) radioStyle {
	activeTheme := frame.ActiveTheme(ctx)
	colors := radioStyleFor(activeTheme, r.variant, state.Hovered, state.Pressed, state.Disabled, state.Invalid)
	defaults := radioStyleDeclaration(activeTheme, colors)
	offState := state
	offState.Checked, offState.Selected = false, false
	onState := state
	onState.Checked, onState.Selected = true, true

	off := styleruntime.ResolvePartStatic(ctx, flowstyle.PartIndicator, offState, defaults, flowstyle.Style{}, flowstyle.Style{}, r.customStyle)
	on := styleruntime.ResolvePartStatic(ctx, flowstyle.PartIndicator, onState, defaults, flowstyle.Style{}, flowstyle.Style{}, r.customStyle)
	label := styleruntime.ResolvePartStatic(ctx, flowstyle.PartLabel, state, defaults, flowstyle.Style{}, flowstyle.Style{}, r.customStyle)
	description := styleruntime.ResolvePartStatic(ctx, flowstyle.PartDescription, state, defaults, flowstyle.Style{}, flowstyle.Style{}, r.customStyle)
	off = styleruntime.ApplyPartTransitions(ctx, gtx, frame.DerivedKey(ctx, key, "off"), flowstyle.PartIndicator, off)
	on = styleruntime.ApplyPartTransitions(ctx, gtx, frame.DerivedKey(ctx, key, "on"), flowstyle.PartIndicator, on)
	label = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartLabel, label)
	description = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartDescription, description)
	return radioStyle{indicatorOff: off, indicatorOn: on, label: label, description: description, pressed: state.Pressed}
}

func radioStyleDeclaration(activeTheme *theme.Theme, colors radioColors) flowstyle.Style {
	tokens := activeTheme.Components.RadioGroup
	indicator := flowstyle.Style{}.
		Width(tokens.Size).
		Height(tokens.Size).
		Background(flowstyle.SolidColor{Color: colors.bg}).
		BorderWidth(tokens.BorderWidth).
		BorderColor(flowstyle.SolidColor{Color: colors.border}).
		Radius(tokens.Size/2).
		Outline(unit.Dp(2), 0, flowstyle.SolidColor{Color: colors.focusColor}).
		TextColor(flowstyle.SolidColor{Color: colors.dot}).
		When(flowstyle.Checked, flowstyle.Style{}.
			Background(flowstyle.SolidColor{Color: colors.selectedBg}).
			BorderColor(flowstyle.SolidColor{Color: colors.selectedBg}))
	return flowstyle.Style{}.
		Part(flowstyle.PartIndicator, indicator).
		Part(flowstyle.PartLabel, flowstyle.Style{}.
			TextColor(flowstyle.SolidColor{Color: colors.fg}).
			FontSize(tokens.TextSize).
			FontWeight(int(font.Medium))).
		Part(flowstyle.PartDescription, flowstyle.Style{}.
			TextColor(flowstyle.SolidColor{Color: colors.description}).
			FontSize(tokens.DescriptionSize))
}

func radioStyleFor(theme *theme.Theme, variant RadioGroupVariant, hovered, pressed, disabled, invalid bool) radioColors {
	style := radioColors{
		bg:          theme.Palette.Surface,
		border:      theme.Palette.Border,
		selectedBg:  theme.Palette.Accent,
		dot:         theme.Palette.AccentForeground,
		fg:          theme.Palette.Foreground,
		description: theme.Palette.MutedForeground,
		focusColor:  theme.Palette.Focus,
	}
	if variant == RadioSecondary {
		style.bg = theme.Palette.SurfaceRaised
	}
	if hovered {
		style.border = theme.Palette.SurfacePressed
	}
	if pressed {
		style.selectedBg = theme.Palette.AccentHover
	}
	if invalid {
		style.border = theme.Palette.Danger
		style.selectedBg = theme.Palette.Danger
		style.dot = theme.Palette.DangerForeground
		style.focusColor = theme.Palette.Danger
		style.focusColor.A = theme.Palette.Focus.A
		if hovered || pressed {
			style.border = theme.Palette.DangerHover
			style.selectedBg = theme.Palette.DangerHover
		}
	}
	if disabled {
		style.bg = theme.DisabledColor(style.bg)
		style.border = theme.DisabledColor(style.border)
		style.selectedBg = theme.DisabledColor(style.selectedBg)
		style.dot = theme.DisabledColor(style.dot)
		style.fg = theme.DisabledColor(style.fg)
		style.description = theme.DisabledColor(style.description)
		style.focusColor = color.NRGBA{}
	}
	return style
}
