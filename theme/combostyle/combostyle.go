package combostyle

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/internal/fieldstyle"
	"github.com/qianniancn/flowui/theme/internal/panelstyle"
	"github.com/qianniancn/flowui/theme/internal/textrender"
)

// BuildCombo builds the visual declaration layers for a combo box. Filtering,
// selection, focus, and popup behavior remain in widget/combobox.
func BuildCombo(active *theme.Theme, request theme.ComboRequest) theme.ComboVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components
	defaults := fieldstyle.FieldDefaultRules(active, request.Variant, fieldstyle.FieldDeclarationOptions{
		TargetPart:          theme.ComboPartTrigger,
		Height:              tokens.ComboBox.Height,
		Radius:              tokens.ComboBox.Radius,
		PaddingX:            tokens.Input.PaddingX,
		TextSize:            tokens.ComboBox.TextSize,
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
		Part(theme.ComboPartIndicator, style.Style{}.
			Width(tokens.ComboBox.TriggerWidth).
			TextColor(style.Solid(sem.FieldPlaceholder))).
		Part(theme.ComboPartIcon, style.Style{}.
			Width(tokens.ComboBox.ItemCheckSize).
			Height(tokens.ComboBox.ItemCheckSize).
			TextColor(style.Solid(sem.FieldPlaceholder)).
			On(style.Disabled, style.Style{}.
				TextColor(style.Solid(active.DisabledColor(sem.FieldPlaceholder))))).
		Part(theme.ComboPartPanel, panelstyle.Popup(
			active,
			tokens.ComboBox.PanelRadius,
			tokens.ComboBox.PanelPadding,
			tokens.ComboBox.PanelBorderWidth,
		)).
		Part(theme.ComboPartDescription, style.Style{}.
			TextColor(style.Solid(sem.MutedForeground)).
			FontSize(tokens.ComboBox.ItemDescriptionSize).
			On(style.Disabled, style.Style{}.
				TextColor(style.Solid(active.DisabledColor(sem.MutedForeground))))).
		Part(theme.ComboPartItem, style.Style{}.
			MinHeight(tokens.ComboBox.ItemHeight).
			PaddingX(tokens.ComboBox.ItemPaddingX).
			PaddingY(tokens.ComboBox.ItemPaddingY).
			Radius(tokens.ComboBox.ItemRadius).
			Background(style.RGBA(0x00000000)).
			TextColor(style.Solid(sem.Foreground)).
			FontSize(tokens.ComboBox.ItemTextSize).
			FontWeight(int(font.Medium)).
			Transition(style.PropBackgroundColor, tokens.ComboBox.ItemColorDuration).
			Transition(style.PropTransform, tokens.ComboBox.ItemPressDuration).
			On(style.Hovered, style.Style{}.Background(style.Solid(sem.SurfaceRaised))).
			On(style.Pressed, style.Style{}.
				Background(style.Solid(sem.SurfacePressed)).
				Scale(0.98).
				Transition(style.PropTransform, tokens.ComboBox.ItemPressInDuration)).
			On(style.Disabled, style.Style{}.
				TextColor(style.Solid(active.DisabledColor(sem.Foreground)))))
	return theme.ComboVisual{
		Defaults: defaults,
		Override: theme.ComboOverrideStyle(request),
		Motion:   active.Motion,
		Metrics: theme.ComboMetrics{
			PanelGap:           tokens.ComboBox.PanelGap,
			PanelMaxHeight:     tokens.ComboBox.PanelMaxHeight,
			ItemHeight:         tokens.ComboBox.ItemHeight,
			ItemTextSize:       tokens.ComboBox.ItemTextSize,
			AnimationDuration:  tokens.ComboBox.AnimationDuration,
			ItemSelectDuration: tokens.ComboBox.ItemSelectDuration,
			IconDuration:       tokens.ComboBox.IconDuration,
			AnimationScale:     tokens.ComboBox.AnimationScale,
		},
		Chevron: func(gtx layout.Context, size image.Point, progress float32, resolved style.ResolvedStyle) {
			renderChevron(gtx, size, progress, textrender.Color(resolved), gtx.Dp(tokens.ComboBox.ChevronStroke*12))
		},
		Check: func(gtx layout.Context, size image.Point, progress float32, resolved style.ResolvedStyle) {
			renderCheck(gtx, size, progress, textrender.Color(resolved),
				gtx.Dp(tokens.ComboBox.ItemCheckInset),
				gtx.Dp(tokens.ComboBox.ItemCheckStroke))
		},
		Text: textRenderer(active),
		EditorStyle: func(content, placeholder, selection style.ResolvedStyle) theme.ComboEditorStyle {
			return theme.ComboEditorStyle{
				TextSize:    textrender.TypographySize(active, content),
				LineHeight:  lineHeight(content),
				Foreground:  textrender.Color(content),
				Placeholder: textrender.Color(placeholder),
				Selection:   textrender.PaintColor(selection),
			}
		},
		IndicatorWidth: func(resolved style.ResolvedStyle) unit.Dp {
			return textrender.PartWidth(resolved)
		},
		IconSize: func(gtx layout.Context, resolved style.ResolvedStyle) image.Point {
			return textrender.PartSize(gtx, resolved)
		},
	}
}

func lineHeight(resolved style.ResolvedStyle) unit.Sp {
	if resolved.Text == nil || resolved.Text.LineHeight == nil {
		return 0
	}
	return *resolved.Text.LineHeight
}

func textRenderer(active *theme.Theme) theme.ComboTextRenderer {
	return func(gtx layout.Context, request theme.ComboTextRequest) layout.Dimensions {
		return textrender.Label(active, gtx, request.Text, request.Style, request.Size)
	}
}
