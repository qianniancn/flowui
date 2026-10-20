package selectstyle

import (
	"image"

	"gioui.org/layout"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/internal/fieldstyle"
	"github.com/qianniancn/flowui/theme/internal/panelstyle"
	"github.com/qianniancn/flowui/theme/internal/textrender"
)

// BuildSelect builds the visual declaration layers for a select trigger.
// Selection state, focus state, overlay lifetime, and item behavior remain in
// widget/selectfield; this package only owns the visual parts.
func BuildSelect(active *theme.Theme, request theme.SelectRequest) theme.SelectVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components
	defaults := fieldstyle.FieldDefaultRules(active, request.Variant, fieldstyle.FieldDeclarationOptions{
		TargetPart:          style.PartContent,
		MinHeight:           tokens.Select.Height,
		Radius:              tokens.Select.Radius,
		PaddingX:            tokens.Select.TriggerPaddingX,
		PaddingY:            tokens.Select.TriggerPaddingY,
		TextSize:            tokens.Select.TextSize,
		FocusRingWidth:      tokens.Input.FocusRingWidth,
		FocusGlowBlur:       tokens.Input.FocusGlowBlur,
		FocusHaloBlur:       tokens.Input.FocusHaloBlur,
		FocusHaloOpacity:    tokens.Input.FocusHaloOpacity,
		FocusShadowSpread:   tokens.Input.FocusShadowSpread,
		BorderWidth:         tokens.Input.BorderWidth,
		InvalidOutlineWidth: tokens.Input.InvalidOutlineWidth,
		ShadowColor:         tokens.Input.ShadowColor,
		ShadowOpacity:       tokens.Input.ShadowOpacity,
		ShadowStrength:      tokens.Input.ShadowStrength,
	})
	sem := active.Semantic()
	defaults = defaults.
		Part(theme.SelectPartIndicator, style.Style{}.
			Width(tokens.Select.IndicatorWidth).
			TextColor(style.Solid(sem.FieldPlaceholder))).
		Part(theme.SelectPartIcon, style.Style{}.
			Width(tokens.Select.IndicatorSize).
			Height(tokens.Select.IndicatorSize).
			TextColor(style.Solid(sem.FieldPlaceholder))).
		Part(theme.SelectPartPanel, panel(active)).
		Part(theme.SelectPartDescription, description(active))
	return theme.SelectVisual{
		Defaults: defaults,
		Override: theme.SelectOverrideStyle(request),
		Metrics: theme.SelectMetrics{
			ContentGap:        tokens.Select.ContentGap,
			PanelGap:          tokens.Select.PanelGap,
			PanelMaxHeight:    tokens.Select.PanelMaxHeight,
			AnimationScale:    tokens.Select.AnimationScale,
			AnimationDistance: tokens.Select.AnimationDistance,
			OpenDuration:      tokens.Select.OpenDuration,
			CloseDuration:     tokens.Select.CloseDuration,
			IndicatorDuration: tokens.Select.IndicatorDuration,
		},
		Motion:    active.Motion,
		Indicator: indicatorRenderer(tokens.Select.IndicatorSize, tokens.Select.IndicatorStroke),
		Text:      textRenderer(active),
		IndicatorWidth: func(resolved style.ResolvedStyle) unit.Dp {
			return textrender.PartWidth(resolved)
		},
		IconSize: func(gtx layout.Context, resolved style.ResolvedStyle) image.Point {
			return textrender.PartSize(gtx, resolved)
		},
	}
}

func panel(active *theme.Theme) style.Style {
	tokens := active.Components.Select
	return panelstyle.Popup(active, tokens.PanelRadius, tokens.PanelPadding, tokens.PanelBorderWidth)
}

func description(active *theme.Theme) style.Style {
	sem := active.Semantic()
	danger := sem.Tone(theme.ToneDanger).Fill
	return style.Style{}.
		FontSize(active.Components.Description.TextSize).
		TextColor(style.Solid(sem.MutedForeground)).
		Wrap(giotext.WrapHeuristically).
		On(style.Invalid, style.Style{}.TextColor(style.Solid(danger))).
		On(style.Disabled, style.Style{}.
			TextColor(style.Solid(active.DisabledColor(sem.MutedForeground)))).
		On(style.Invalid|style.Disabled, style.Style{}.
			TextColor(style.Solid(active.DisabledColor(danger))))
}
