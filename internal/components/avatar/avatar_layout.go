package avatar

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	imageview "github.com/qianniancn/FlowUI/internal/components/image"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/flowui-icons-lucide"
)

func (a Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	style := avatarStyleFor(frame.ActiveTheme(ctx), a.color, a.variant, a.size)
	size := avatarSize(gtx, style)
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}
	gtx.Constraints = layout.Exact(size)
	radius := min(max(gtx.Dp(style.radius), 0), min(size.X, size.Y)/2)
	rect := image.Rectangle{Max: size}
	root := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, style.background, clip.Rect(rect).Op())
	if a.image.Size() != (image.Point{}) {
		a.layoutImage(ctx, gtx)
	} else {
		if label := a.semanticLabel(); label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		a.layoutFallback(ctx, gtx, style)
	}
	root.Pop()
	return layout.Dimensions{Size: size}
}

func avatarSize(gtx layout.Context, style avatarStyle) image.Point {
	maximum := min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
	if maximum <= 0 {
		return image.Point{}
	}
	minimum := min(max(gtx.Constraints.Min.X, gtx.Constraints.Min.Y), maximum)
	diameter := min(max(gtx.Dp(style.diameter), minimum), maximum)
	return image.Pt(diameter, diameter)
}

func (a Widget) layoutImage(ctx *frame.Context, gtx layout.Context) {
	imageview.New(a.image).
		Fit(imageview.FitCover).
		Alt(a.semanticLabel()).
		Layout(ctx, gtx)
}

func (a Widget) layoutFallback(ctx *frame.Context, gtx layout.Context, style avatarStyle) {
	restore := frame.PushColors(ctx, style.foreground, style.background)
	defer restore()
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		switch {
		case a.fallback != nil:
			return a.fallback.Layout(ctx, gtx)
		case a.fallbackText != "":
			label := material.Label(frame.ActiveTheme(ctx).Material, style.textSize, a.fallbackText)
			label.Color = style.foreground
			label.Font.Weight = font.Medium
			label.MaxLines = 1
			return label.Layout(gtx)
		default:
			return icon.New(lucide.UserRound).Size(float32(style.iconSize)).Color(style.foreground).Layout(ctx, gtx)
		}
	})
}

func (a Widget) semanticLabel() string {
	if a.alt != "" {
		return a.alt
	}
	return a.fallbackText
}
