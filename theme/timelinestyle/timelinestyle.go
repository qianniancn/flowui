// Package timelinestyle contains the visual recipe for timeline providers.
// Timeline behavior stays in widget/timeline; this package owns geometry,
// palette, and the instance declaration boundary.
package timelinestyle

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildTimeline returns the complete visual plan for the active theme.
func BuildTimeline(active *theme.Theme, request theme.TimelineRequest) theme.TimelineVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Timeline
	semantic := active.Semantic()
	return theme.TimelineVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
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
			Background:    semantic.Background,
			Tail:          tokens.TailColor,
			DotBackground: tokens.DotBackground,
			Title:         tokens.TitleColor,
			Content:       tokens.ContentColor,
			Muted:         tokens.MutedColor,
			Primary:       tokens.PrimaryColor,
			Error:         tokens.ErrorColor,
			Success:       tokens.SuccessColor,
			Disabled:      tokens.DisabledColor,
		},
		Rail: renderRail,
	}
}

func renderRail(gtx layout.Context, request theme.TimelineRailRequest) {
	dotSize := max(request.DotSize, 2)
	width := max(request.TailWidth, 1)
	if request.Tail {
		lineStart := request.Start + dotSize/2
		if request.Process {
			if request.Vertical {
				paintVerticalDashes(gtx, request.Center, lineStart, request.End, width, request.TailColor, max(request.DashLength, 2), max(request.DashGap, 2))
			} else {
				paintHorizontalDashes(gtx, lineStart, request.End, request.Center, width, request.TailColor, max(request.DashLength, 2), max(request.DashGap, 2))
			}
		} else if request.Vertical {
			paint.FillShape(gtx.Ops, request.TailColor, clip.Rect(image.Rect(
				request.Center-width/2,
				lineStart,
				request.Center+(width+1)/2,
				request.End,
			)).Op())
		} else {
			paint.FillShape(gtx.Ops, request.TailColor, clip.Rect(image.Rect(
				lineStart,
				request.Center-width/2,
				request.End,
				request.Center+(width+1)/2,
			)).Op())
		}
	}
	if request.CustomIcon {
		return
	}
	drawMarker(gtx, request, dotSize)
}

func paintVerticalDashes(gtx layout.Context, center, start, end, width int, value color.NRGBA, dashLength, dashGap int) {
	for y := start; y < end; y += dashLength + dashGap {
		dashEnd := min(y+dashLength, end)
		paint.FillShape(gtx.Ops, value, clip.Rect(image.Rect(
			center-width/2,
			y,
			center+(width+1)/2,
			dashEnd,
		)).Op())
	}
}

func paintHorizontalDashes(gtx layout.Context, start, end, center, width int, value color.NRGBA, dashLength, dashGap int) {
	for x := start; x < end; x += dashLength + dashGap {
		dashEnd := min(x+dashLength, end)
		paint.FillShape(gtx.Ops, value, clip.Rect(image.Rect(
			x,
			center-width/2,
			dashEnd,
			center+(width+1)/2,
		)).Op())
	}
}

func drawMarker(gtx layout.Context, request theme.TimelineRailRequest, size int) {
	outer := image.Rect(
		request.MarkerX-size/2,
		request.MarkerY-size/2,
		request.MarkerX+(size+1)/2,
		request.MarkerY+(size+1)/2,
	)
	paint.FillShape(gtx.Ops, request.MarkerColor, clip.Ellipse{Min: outer.Min, Max: outer.Max}.Op(gtx.Ops))
	if request.Filled {
		return
	}
	inner := size - max(request.BorderWidth*2, 2)
	if inner <= 0 {
		return
	}
	innerRect := image.Rect(
		request.MarkerX-inner/2,
		request.MarkerY-inner/2,
		request.MarkerX+(inner+1)/2,
		request.MarkerY+(inner+1)/2,
	)
	paint.FillShape(gtx.Ops, request.DotBackground, clip.Ellipse{Min: innerRect.Min, Max: innerRect.Max}.Op(gtx.Ops))
}
