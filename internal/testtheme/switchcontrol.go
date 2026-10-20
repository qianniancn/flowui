package testtheme

import (
	"time"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Switch(active *theme.Theme, request theme.SwitchRequest) theme.SwitchVisual {
	switchTokens := active.Components.Switch
	trackWidth, trackHeight := switchTokens.MediumTrackWidth, switchTokens.MediumTrackHeight
	thumbWidth, thumbHeight := switchTokens.MediumThumbWidth, switchTokens.MediumThumbHeight
	switch request.Size {
	case theme.SwitchSmall:
		trackWidth, trackHeight = switchTokens.SmallTrackWidth, switchTokens.SmallTrackHeight
		thumbWidth, thumbHeight = switchTokens.SmallThumbWidth, switchTokens.SmallThumbHeight
	case theme.SwitchLarge:
		trackWidth, trackHeight = switchTokens.LargeTrackWidth, switchTokens.LargeTrackHeight
		thumbWidth, thumbHeight = switchTokens.LargeThumbWidth, switchTokens.LargeThumbHeight
	}
	defaults := style.Style{}.
		Part(style.PartTrack, style.Style{}.
			Width(trackWidth).Height(trackHeight).
			Background(style.Solid(active.Semantic().SurfaceRaised)).
			Radius(trackHeight/2)).
		Part(style.PartThumb, style.Style{}.
			Width(thumbWidth).Height(thumbHeight).
			Background(style.Solid(active.Semantic().Surface)).
			Radius(thumbHeight/2)).
		Part(style.PartLabel, style.Style{}.
			TextColor(style.Solid(active.Semantic().Foreground)).
			FontSize(switchTokens.TextSize)).
		Part(style.PartDescription, style.Style{}.
			TextColor(style.Solid(active.Semantic().MutedForeground)).
			FontSize(active.Components.Description.TextSize))
	return theme.SwitchVisual{
		Defaults: defaults,
		Override: theme.SwitchOverrideStyle(request),
		Metrics: theme.SwitchMetrics{
			FocusSpace: switchTokens.FocusSpace, ContentGap: switchTokens.ContentGap,
			DescriptionGap: switchTokens.DescriptionGap,
		},
		Motion:            active.Motion,
		SelectionDuration: 300 * time.Millisecond,
		FocusDuration:     100 * time.Millisecond,
	}
}

func (behaviorProvider) SwitchGroup(active *theme.Theme, request theme.SwitchGroupRequest) theme.SwitchGroupVisual {
	return theme.SwitchGroupVisual{
		Metrics: theme.SwitchGroupMetrics{
			VerticalGap:   active.Components.SwitchGroup.VerticalGap,
			HorizontalGap: active.Components.SwitchGroup.HorizontalGap,
		},
		Override: theme.SwitchOverrideStyle(theme.SwitchRequest{Override: request.Override}),
	}
}
