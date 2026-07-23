package switches

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

type switchColors struct {
	trackOff   color.NRGBA
	trackOn    color.NRGBA
	thumb      color.NRGBA
	thumbOn    color.NRGBA
	thumbFgOff color.NRGBA
	thumbFg    color.NRGBA
	focusColor color.NRGBA
}

type switchStyle struct {
	trackOff    flowstyle.ResolvedStyle
	trackOn     flowstyle.ResolvedStyle
	thumbOff    flowstyle.ResolvedStyle
	thumbOn     flowstyle.ResolvedStyle
	label       flowstyle.ResolvedStyle
	description flowstyle.ResolvedStyle
	selected    float32
	focus       float32
}

type switchSizeStyle struct {
	trackWidth  unit.Dp
	trackHeight unit.Dp
	thumbWidth  unit.Dp
	thumbHeight unit.Dp
}

func switchStyleFor(theme *theme.Theme, hovered, pressed, disabled, invalid bool) switchColors {
	style := switchColors{
		trackOff:   theme.Palette.SurfaceRaised,
		trackOn:    theme.Palette.Accent,
		thumb:      theme.Palette.Surface,
		thumbOn:    theme.Palette.AccentForeground,
		thumbFgOff: theme.Palette.MutedForeground,
		thumbFg:    theme.Palette.Accent,
		focusColor: theme.Palette.Focus,
	}
	if hovered {
		style.trackOff = theme.Palette.SurfacePressed
		style.trackOn = theme.Palette.AccentHover
	}
	if pressed {
		style.trackOff = theme.Palette.SurfacePressed
		style.trackOn = theme.Palette.AccentHover
	}
	if invalid {
		style.trackOn = theme.Palette.Danger
		style.thumbFg = theme.Palette.Danger
		style.focusColor = theme.Palette.Danger
		style.focusColor.A = theme.Palette.Focus.A
		if hovered || pressed {
			style.trackOn = theme.Palette.DangerHover
		}
	}
	if disabled {
		style.trackOff = theme.DisabledColor(style.trackOff)
		style.trackOn = theme.DisabledColor(style.trackOn)
		style.thumb = theme.DisabledColor(style.thumb)
		style.thumbOn = theme.DisabledColor(style.thumbOn)
		style.thumbFgOff = theme.DisabledColor(style.thumbFgOff)
		style.thumbFg = theme.DisabledColor(style.thumbFg)
		style.focusColor = color.NRGBA{}
	}
	return style
}

func (s SwitchWidget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState) (switchStyle, switchSizeStyle) {
	activeTheme := frame.ActiveTheme(ctx)
	colors := switchStyleFor(activeTheme, state.Hovered, state.Pressed, state.Disabled, state.Invalid)
	size := switchSizeStyleFor(activeTheme, s.size)
	defaults := switchStyleDeclaration(activeTheme, colors, size, state.Disabled)
	offState := state
	offState.Checked, offState.Selected = false, false
	onState := state
	onState.Checked, onState.Selected = true, true

	trackOff := styleruntime.ResolvePartStatic(ctx, flowstyle.PartTrack, offState, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	trackOn := styleruntime.ResolvePartStatic(ctx, flowstyle.PartTrack, onState, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	thumbOff := styleruntime.ResolvePartStatic(ctx, flowstyle.PartThumb, offState, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	thumbOn := styleruntime.ResolvePartStatic(ctx, flowstyle.PartThumb, onState, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	label := styleruntime.ResolvePartStatic(ctx, flowstyle.PartLabel, state, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	description := styleruntime.ResolvePartStatic(ctx, flowstyle.PartDescription, state, defaults, flowstyle.Style{}, flowstyle.Style{}, s.customStyle)
	offOwner := frame.DerivedKey(ctx, key, "switch-off")
	onOwner := frame.DerivedKey(ctx, key, "switch-on")
	trackOff = styleruntime.ApplyPartTransitions(ctx, gtx, offOwner, flowstyle.PartTrack, trackOff)
	trackOn = styleruntime.ApplyPartTransitions(ctx, gtx, onOwner, flowstyle.PartTrack, trackOn)
	thumbOff = styleruntime.ApplyPartTransitions(ctx, gtx, offOwner, flowstyle.PartThumb, thumbOff)
	thumbOn = styleruntime.ApplyPartTransitions(ctx, gtx, onOwner, flowstyle.PartThumb, thumbOn)
	label = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartLabel, label)
	description = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartDescription, description)

	return switchStyle{
		trackOff: trackOff, trackOn: trackOn,
		thumbOff: thumbOff, thumbOn: thumbOn,
		label: label, description: description,
	}, size
}

func switchStyleDeclaration(activeTheme *theme.Theme, value switchColors, size switchSizeStyle, disabled bool) flowstyle.Style {
	tokens := activeTheme.Components.Switch
	track := flowstyle.Style{}.
		Width(size.trackWidth).
		Height(size.trackHeight).
		Background(flowstyle.SolidColor{Color: value.trackOff}).
		Radius(size.trackHeight/2).
		Outline(tokens.FocusRingWidth, 0, flowstyle.SolidColor{Color: value.focusColor}).
		When(flowstyle.Checked, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: value.trackOn}))
	thumb := flowstyle.Style{}.
		Width(size.thumbWidth).
		Height(size.thumbHeight).
		Background(flowstyle.SolidColor{Color: value.thumb}).
		Radius(size.thumbHeight/2).
		TextColor(flowstyle.SolidColor{Color: value.thumbFgOff}).
		When(flowstyle.Checked, flowstyle.Style{}.
			Background(flowstyle.SolidColor{Color: value.thumbOn}).
			TextColor(flowstyle.SolidColor{Color: value.thumbFg}))
	shadowColor := activeTheme.Palette.SurfaceShadow
	for _, layer := range activeTheme.Shadows.SwitchThumb.Layers {
		col := shadowColor
		col.A = byte(float32(col.A)*min(max(layer.Opacity, 0), 1) + .5)
		if layer.Blur >= 0 && col.A != 0 {
			thumb = thumb.BoxShadow(layer.OffsetX, layer.OffsetY, layer.Blur, layer.Spread, flowstyle.SolidColor{Color: col})
		}
	}
	labelColor := activeTheme.Palette.Foreground
	if disabled {
		labelColor = activeTheme.DisabledColor(labelColor)
	}
	label := flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: labelColor}).
		FontSize(tokens.TextSize).
		FontWeight(int(font.Medium))
	return flowstyle.Style{}.
		Part(flowstyle.PartTrack, track).
		Part(flowstyle.PartThumb, thumb).
		Part(flowstyle.PartLabel, label)

}

func switchSizeStyleFor(theme *theme.Theme, size SwitchSize) switchSizeStyle {
	switch size {
	case SwitchSmall:
		return switchSizeStyle{
			trackWidth:  theme.Components.Switch.SmallTrackWidth,
			trackHeight: theme.Components.Switch.SmallTrackHeight,
			thumbWidth:  theme.Components.Switch.SmallThumbWidth,
			thumbHeight: theme.Components.Switch.SmallThumbHeight,
		}
	case SwitchLarge:
		return switchSizeStyle{
			trackWidth:  theme.Components.Switch.LargeTrackWidth,
			trackHeight: theme.Components.Switch.LargeTrackHeight,
			thumbWidth:  theme.Components.Switch.LargeThumbWidth,
			thumbHeight: theme.Components.Switch.LargeThumbHeight,
		}
	default:
		return switchSizeStyle{
			trackWidth:  theme.Components.Switch.MediumTrackWidth,
			trackHeight: theme.Components.Switch.MediumTrackHeight,
			thumbWidth:  theme.Components.Switch.MediumThumbWidth,
			thumbHeight: theme.Components.Switch.MediumThumbHeight,
		}
	}
}
