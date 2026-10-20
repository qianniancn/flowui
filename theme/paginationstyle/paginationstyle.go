// Package paginationstyle contains the theme-owned visual recipe for
// pagination behavior. Page calculation, semantics and interaction remain in
// widget/pagination; this package owns the chrome and its drawing strategy.
package paginationstyle

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
)

// BuildPagination builds the complete provider-owned pagination visual plan.
func BuildPagination(active *theme.Theme, request theme.PaginationRequest) theme.PaginationVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Pagination
	metrics, renderMetrics := paginationMetrics(tokens, request.Size)
	sem := active.Semantic()
	return theme.PaginationVisual{
		Defaults: style.Style{},
		Summary: style.Style{}.
			TextColor(style.Solid(sem.MutedForeground)).
			FontSize(renderMetrics.textSize),
		Override: theme.PaginationOverrideStyle(request),
		Metrics:  metrics,
		Motion:   active.Motion,
		Button: func(gtx layout.Context, button theme.PaginationButtonRequest) layout.Dimensions {
			return renderButton(active, gtx, renderMetrics, button)
		},
	}
}

type paginationRenderMetrics struct {
	textSize       unit.Sp
	paddingX       unit.Dp
	iconSize       unit.Dp
	navGap         unit.Dp
	focusRingWidth unit.Dp
	radius         unit.Dp
	pressedScale   float32
}

func paginationMetrics(tokens theme.PaginationTheme, size theme.PaginationSize) (theme.PaginationMetrics, paginationRenderMetrics) {
	metrics := theme.PaginationMetrics{
		Size: tokens.MediumSize, ItemGap: tokens.ItemGap, ContentGap: tokens.ContentGap,
		CompactWidth: tokens.CompactWidth,
	}
	renderMetrics := paginationRenderMetrics{
		textSize: tokens.MediumTextSize, paddingX: tokens.MediumPaddingX,
		iconSize: tokens.IconSize, navGap: tokens.NavGap, focusRingWidth: tokens.FocusRingWidth,
		radius: tokens.Radius, pressedScale: .97,
	}
	switch size {
	case theme.PaginationSmall:
		metrics.Size = tokens.SmallSize
		renderMetrics.textSize, renderMetrics.paddingX = tokens.SmallTextSize, tokens.SmallPaddingX
	case theme.PaginationLarge:
		metrics.Size = tokens.LargeSize
		renderMetrics.textSize, renderMetrics.paddingX = tokens.LargeTextSize, tokens.LargePaddingX
	}
	return metrics, renderMetrics
}

func renderButton(active *theme.Theme, gtx layout.Context, metrics paginationRenderMetrics, request theme.PaginationButtonRequest) layout.Dimensions {
	if request.Kind == theme.PaginationEllipsis {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return paginationLabel(active, gtx, request.Label, metrics.textSize, active.Semantic().MutedForeground)
		})
	}
	sem := active.Semantic()
	background := color.NRGBA{}
	foreground := sem.Foreground
	if request.Active {
		background = sem.Tone(theme.ToneDefault).Fill
		foreground = sem.Tone(theme.ToneDefault).OnFill
	}
	if request.Hovered || request.Pressed {
		background = sem.Tone(theme.ToneDefault).Hover
	}
	if request.Disabled {
		background = active.DisabledColor(background)
		foreground = active.DisabledColor(foreground)
	}
	macro := op.Record(gtx.Ops)
	dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if request.Kind == theme.PaginationPage {
			return paginationLabel(active, gtx, request.Label, metrics.textSize, foreground)
		}
		return layout.Inset{Left: metrics.paddingX, Right: metrics.paddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			iconChild := func(gtx layout.Context) layout.Dimensions {
				diameter := min(gtx.Dp(metrics.iconSize), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
				iconGtx := gtx
				iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
				data := lucide.ChevronLeft
				if request.Kind == theme.PaginationNext {
					data = lucide.ChevronRight
				}
				return icon.Layout(data, iconGtx, foreground)
			}
			labelChild := func(gtx layout.Context) layout.Dimensions {
				return paginationLabel(active, gtx, request.Label, metrics.textSize, foreground)
			}
			children := []layout.FlexChild{layout.Rigid(iconChild), layout.Rigid(labelChild)}
			if request.Kind == theme.PaginationNext {
				children[0], children[1] = children[1], children[0]
			}
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: gtx.Dp(metrics.navGap)}.Layout(gtx, children...)
		})
	})
	call := macro.Stop()
	scale := float32(1)
	if request.Pressed && !request.Disabled && metrics.pressedScale > 0 && metrics.pressedScale <= 1 {
		scale = metrics.pressedScale
	}
	transform := draw.Scale(dims.Size, scale).Push(gtx.Ops)
	drawPaginationButton(gtx, dims.Size, background, sem.Ring, request.Focus, metrics.focusRingWidth, metrics.radius)
	call.Add(gtx.Ops)
	transform.Pop()
	return dims
}

func paginationLabel(active *theme.Theme, gtx layout.Context, value string, size unit.Sp, foreground color.NRGBA) layout.Dimensions {
	label := material.Label(theme.MaterialOf(active), size, value)
	label.Color = foreground
	label.Font.Weight = font.Medium
	return label.Layout(gtx)
}

func drawPaginationButton(gtx layout.Context, size image.Point, background, focusColor color.NRGBA, focus float32, focusWidth, radiusDp unit.Dp) {
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(radiusDp), 0), min(size.X, size.Y)/2)
	if background.A != 0 {
		paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(focusWidth), 1)
	focusRect := rect.Inset(width + 1)
	if focusRect.Empty() {
		return
	}
	focusColor.A = byte(float32(focusColor.A)*focus + .5)
	stroke := clip.Stroke{Path: clip.UniformRRect(focusRect, max(radius-width-1, 0)).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, focusColor)
	stroke.Pop()
}
