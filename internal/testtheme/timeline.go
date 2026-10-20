package testtheme

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Timeline(active *theme.Theme, request theme.TimelineRequest) theme.TimelineVisual {
	tokens := active.Components.Timeline
	return theme.TimelineVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Opacity(active.DisabledOpacityValue())}),
		Override: theme.TimelineOverrideStyle(request),
		Metrics: theme.TimelineMetrics{
			TailWidth:            tokens.TailWidth,
			DotBorderWidth:       tokens.DotBorderWidth,
			DotSize:              tokens.DotSize,
			ItemPaddingBottom:    tokens.ItemPaddingBottom,
			TitleContentGap:      tokens.TitleContentGap,
			SectionGap:           tokens.SectionGap,
			InlineInset:          tokens.InlineInset,
			HorizontalTitleGap:   tokens.HorizontalTitleGap,
			HorizontalContentGap: tokens.HorizontalContentGap,
			HorizontalItemGap:    tokens.HorizontalItemGap,
			ProcessDashLength:    tokens.ProcessDashLength,
			ProcessDashGap:       tokens.ProcessDashGap,
		},
		Colors: theme.TimelineColors{
			Background:    background,
			Tail:          muted,
			DotBackground: background,
			Title:         foreground,
			Content:       foreground,
			Muted:         muted,
			Primary:       accent,
			Error:         focus,
			Success:       accent,
			Disabled:      muted,
		},
		Rail: func(layout.Context, theme.TimelineRailRequest) {},
	}
}
