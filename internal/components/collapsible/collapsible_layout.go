package collapsible

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/button"
	"github.com/qianniancn/flowui/internal/components/icon"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui-icons-lucide"
)

func layoutItem(ctx *frame.Context, gtx layout.Context, state *collapsibleItemState, item Item, expanded, disabled bool) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Collapsible
	target := float32(0)
	if expanded {
		target = 1
	}
	heightProgress := animation.Tween("content-height", target).
		Duration(tokens.ContentDuration).
		Easing(animation.EaseQuadraticOut).
		Value(ctx, gtx)
	opacityProgress := animation.Tween("content-opacity", target).
		Duration(tokens.ContentDuration).
		Easing(animation.EaseCubicOut).
		Value(ctx, gtx)
	indicatorProgress := animation.Tween("indicator", target).
		Duration(tokens.IndicatorDuration).
		Easing(animation.EaseCubicOut).
		Value(ctx, gtx)

	variant := button.ButtonGhost
	if expanded {
		variant = button.ButtonSecondary
	}
	trigger := button.Button("trigger", triggerContent{item: item, indicatorProgress: indicatorProgress}).
		Label(item.Label).
		Variant(variant).
		Disabled(disabled).
		FullWidth()

	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	var triggerDims, contentDims layout.Dimensions
	var triggerPlacement, contentPlacement frame.OverlayPlacement
	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			triggerDims, triggerPlacement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return button.LayoutWithClickableNoEvents(trigger, ctx, gtx, &state.clickable)
			})
			return triggerDims
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			contentDims, contentPlacement = frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return layoutContent(ctx, gtx, item.Content, heightProgress, opacityProgress)
			})
			return contentDims
		}),
	)
	triggerPlacement.PlaceOffset(image.Point{})
	contentPlacement.PlaceOffset(image.Pt(0, triggerDims.Size.Y))
	return dims
}

type triggerContent struct {
	item              Item
	indicatorProgress float32
}

func (c triggerContent) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	children := make([]frame.Widget, 0, 4)
	if c.item.Leading != nil {
		children = append(children, c.item.Leading)
	}
	children = append(children, layoutui.Expanded(collapsibleLabel(c.item.Label)))
	if c.item.Trailing != nil {
		children = append(children, c.item.Trailing)
	}
	children = append(children, indicatorWidget{progress: c.indicatorProgress})
	return layoutui.LayoutTrackedFlex(
		ctx,
		gtx,
		layout.Horizontal,
		frame.ActiveTheme(ctx).Components.Button.ContentGap,
		layout.Middle,
		children...,
	)
}

type collapsibleLabel string

func (value collapsibleLabel) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	label := material.Label(frame.ActiveMaterial(ctx), frame.ActiveTheme(ctx).Typography.ControlSize, string(value))
	label.Color = ctx.ForegroundColor()
	label.Font.Weight = font.Medium
	label.MaxLines = 1
	label.Truncator = "..."
	return label.Layout(gtx)
}

type indicatorWidget struct {
	progress float32
}

func (w indicatorWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.Collapsible
	diameter := max(gtx.Dp(tokens.IndicatorSize), icon.LucideSizeForStroke(gtx, tokens.IndicatorStroke))
	diameter = min(diameter, min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	if diameter <= 0 {
		return layout.Dimensions{}
	}
	size := image.Pt(diameter, diameter)
	center := f32.Pt(float32(diameter)/2, float32(diameter)/2)
	rotation := op.Affine(f32.AffineId().Rotate(center, -w.progress*float32(math.Pi))).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(size)
	color := activeTheme.Palette.MutedForeground
	if !gtx.Enabled() {
		color = activeTheme.DisabledColor(color)
	}
	icon.Layout(lucide.ChevronDown, iconGtx, color)
	rotation.Pop()
	return layout.Dimensions{Size: size}
}

func layoutContent(ctx *frame.Context, gtx layout.Context, content frame.Widget, heightProgress, opacityProgress float32) layout.Dimensions {
	if content == nil || (heightProgress <= 0 && opacityProgress <= 0) {
		return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
	}
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	gtx.Constraints.Min.Y = 0
	macro := op.Record(gtx.Ops)
	fullDims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return layout.UniformInset(frame.ActiveTheme(ctx).Components.Collapsible.BodyPadding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return content.Layout(ctx, gtx)
		})
	})
	call := macro.Stop()
	visibleHeight := min(max(int(float32(fullDims.Size.Y)*heightProgress+.5), 0), fullDims.Size.Y)
	visible := image.Rect(0, 0, fullDims.Size.X, visibleHeight)
	placement.ClipTo(visible)
	placement.SetOpacity(opacityProgress)
	if visibleHeight > 0 && opacityProgress > 0 {
		area := clip.Rect(visible).Push(gtx.Ops)
		fade := paint.PushOpacity(gtx.Ops, opacityProgress)
		call.Add(gtx.Ops)
		fade.Pop()
		area.Pop()
	}
	return layout.Dimensions{Size: image.Pt(fullDims.Size.X, visibleHeight)}
}
