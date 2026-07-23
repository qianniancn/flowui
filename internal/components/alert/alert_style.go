package alert

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type alertStyle struct {
	background  color.NRGBA
	foreground  color.NRGBA
	indicator   color.NRGBA
	title       color.NRGBA
	description color.NRGBA
}

type alertResolvedStyle struct {
	root        flowstyle.ResolvedStyle
	title       flowstyle.ResolvedStyle
	description flowstyle.ResolvedStyle
	content     flowstyle.ResolvedStyle
	indicator   flowstyle.ResolvedStyle
}

func (a Widget) resolveStyle(ctx *frame.Context, gtx layout.Context) alertResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	state := flowstyle.StyleState{}
	root := alertRootDeclaration(activeTheme)
	title := alertTitleDeclaration(activeTheme, a.status)
	description := alertDescriptionDeclaration(activeTheme)
	indicator := alertIndicatorDeclaration(activeTheme, a.status)
	resolved := alertResolvedStyle{
		root:        styleruntime.ResolveStatic(ctx, state, root, flowstyle.Style{}, flowstyle.Style{}, a.customStyle),
		title:       styleruntime.ResolvePartStatic(ctx, flowstyle.PartLabel, state, title, flowstyle.Style{}, flowstyle.Style{}, a.customStyle),
		description: styleruntime.ResolvePartStatic(ctx, flowstyle.PartDescription, state, description, flowstyle.Style{}, flowstyle.Style{}, a.customStyle),
		content:     styleruntime.ResolvePartStatic(ctx, flowstyle.PartContent, state, description, flowstyle.Style{}, flowstyle.Style{}, a.customStyle),
		indicator:   styleruntime.ResolvePartStatic(ctx, flowstyle.PartIndicator, state, indicator, flowstyle.Style{}, flowstyle.Style{}, a.customStyle),
	}
	if !styleruntime.HasTransitions(resolved.root, resolved.title, resolved.description, resolved.content, resolved.indicator) {
		return resolved
	}
	key := frame.ClaimKey(ctx, stateutil.KindStyle, "alert")
	resolved.root = styleruntime.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.title = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartLabel, resolved.title)
	resolved.description = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartDescription, resolved.description)
	resolved.content = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartContent, resolved.content)
	resolved.indicator = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartIndicator, resolved.indicator)
	return resolved
}

func alertRootDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	tokens := activeTheme.Components.Alert
	return flowstyle.Style{}.
		FillWidth().
		Background(flowstyle.SolidColor{Color: activeTheme.Palette.Surface}).
		TextColor(flowstyle.SolidColor{Color: activeTheme.Palette.SurfaceForeground}).
		PaddingX(tokens.PaddingX).
		PaddingY(tokens.PaddingY).
		Radius(tokens.Radius).
		Overflow(flowstyle.OverflowHidden).
		Shadow(flowstyle.ShadowSurface).
		Opacity(1)

}

func alertTitleDeclaration(activeTheme *theme.Theme, status Status) flowstyle.Style {
	tokens := activeTheme.Components.Alert
	return flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: alertStyleFor(activeTheme, status).title}).
		FontSize(tokens.TitleSize).
		FontWeight(int(font.Medium)).
		LineHeight(tokens.TitleLineHeight)

}

func alertDescriptionDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	tokens := activeTheme.Components.Alert
	return flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: activeTheme.Palette.MutedForeground}).
		FontSize(tokens.DescriptionSize).
		LineHeight(tokens.DescriptionLineHeight)

}

func alertIndicatorDeclaration(activeTheme *theme.Theme, status Status) flowstyle.Style {
	return flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: alertStyleFor(activeTheme, status).indicator})

}

func alertStyleFor(activeTheme *theme.Theme, status Status) alertStyle {
	foreground := activeTheme.Palette.SurfaceForeground
	statusColor := foreground
	switch status {
	case StatusAccent:
		statusColor = activeTheme.Palette.AccentSoftForeground
	case StatusSuccess:
		statusColor = activeTheme.Palette.SuccessSoftForegroundColor()
	case StatusWarning:
		statusColor = activeTheme.Palette.WarningSoftForegroundColor()
	case StatusDanger:
		statusColor = activeTheme.Palette.DangerSoftForeground
	}
	return alertStyle{
		background:  activeTheme.Palette.Surface,
		foreground:  foreground,
		indicator:   statusColor,
		title:       statusColor,
		description: activeTheme.Palette.MutedForeground,
	}
}
