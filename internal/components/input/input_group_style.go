package input

import (
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/field"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type inputGroupResolvedStyle struct {
	root        flowstyle.ResolvedStyle
	prefix      flowstyle.ResolvedStyle
	suffix      flowstyle.ResolvedStyle
	divider     flowstyle.ResolvedStyle
	selection   flowstyle.ResolvedStyle
	placeholder flowstyle.ResolvedStyle
}

func inputGroupDefaultDeclaration(activeTheme *theme.Theme, variant InputVariant, fullWidth bool, minHeight unit.Dp) flowstyle.Style {
	tokens := activeTheme.Components.InputGroup
	root := field.DefaultDeclaration(activeTheme, variant, field.DeclarationOptions{
		MinHeight: max(tokens.MinHeight, minHeight), Radius: tokens.Radius,
		TextSize: tokens.TextSize, LineHeight: tokens.LineHeight,
		FocusRingWidth: tokens.FocusRingWidth, InvalidOutlineWidth: tokens.InvalidOutlineWidth,
		ShadowColor: tokens.ShadowColor, ShadowOpacity: tokens.ShadowOpacity,
		ShadowStrength: tokens.ShadowStrength, FillWidth: fullWidth,
	})
	return root.
		Part(flowstyle.PartPrefix, flowstyle.Style{}.TextColor(flowstyle.TokenMutedForeground)).
		Part(flowstyle.PartSuffix, flowstyle.Style{}.TextColor(flowstyle.TokenMutedForeground)).
		Part(flowstyle.PartIndicator, flowstyle.Style{}.Width(tokens.DividerWidth).Background(flowstyle.TokenBorder))

}
