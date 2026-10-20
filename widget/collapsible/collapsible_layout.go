package collapsible

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/button"
	componenttext "github.com/qianniancn/flowui/widget/text"
)

func layoutItem(ctx *core.Context, gtx layout.Context, state *collapsibleItemState, item Item, expanded, disabled bool, visual theme.CollapsibleVisual) layout.Dimensions {
	target := float32(0)
	if expanded {
		target = 1
	}
	heightProgress := animation.Tween("content-height", target).
		Duration(visual.Metrics.ContentDuration).
		Easing(animation.EaseQuadraticOut).
		Value(ctx, gtx)
	opacityProgress := animation.Tween("content-opacity", target).
		Duration(visual.Metrics.ContentDuration).
		Easing(animation.EaseCubicOut).
		Value(ctx, gtx)
	indicatorProgress := animation.Tween("indicator", target).
		Duration(visual.Metrics.IndicatorDuration).
		Easing(animation.EaseCubicOut).
		Value(ctx, gtx)

	variant := button.Ghost
	if expanded {
		variant = button.Secondary
	}
	trigger := (button.New("").Content(triggerContent{item: item, indicatorProgress: indicatorProgress, expanded: expanded, disabled: disabled, visual: visual}).
		Label(item.Label).
		Variant(variant).
		Disabled(disabled).
		FullWidth()).Style(visual.Trigger)

	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	var triggerDims, contentDims layout.Dimensions
	var triggerPlacement, contentPlacement core.OverlayPlacement
	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			triggerDims, triggerPlacement = core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return button.LayoutWithClickableNoEvents(trigger, ctx, gtx, &state.clickable)
			})
			return triggerDims
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			contentDims, contentPlacement = core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
				return layoutContent(ctx, gtx, item.Content, heightProgress, opacityProgress, visual.Metrics.BodyPadding)
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
	expanded          bool
	disabled          bool
	visual            theme.CollapsibleVisual
}

func (c triggerContent) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	children := make([]core.Widget, 0, 4)
	if c.item.Leading != nil {
		children = append(children, c.item.Leading)
	}
	children = append(children, layoutui.Expanded(collapsibleLabel{value: c.item.Label, expanded: c.expanded, disabled: c.disabled, declaration: c.visual.Label}))
	if c.item.Trailing != nil {
		children = append(children, c.item.Trailing)
	}
	children = append(children, indicatorWidget{progress: c.indicatorProgress, visual: c.visual})
	return layoutui.TrackedFlex(
		ctx,
		gtx,
		layout.Horizontal,
		c.visual.Metrics.ContentGap,
		layout.Middle,
		children...,
	)
}

type collapsibleLabel struct {
	value       string
	expanded    bool
	disabled    bool
	declaration style.Style
}

func (value collapsibleLabel) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	disabled := value.disabled || !gtx.Enabled()
	state := style.State{Expanded: value.expanded, Disabled: disabled}
	resolved := computed.ResolveStatic(ctx, state, value.declaration, style.Style{}, style.Style{}, style.Style{})
	return componenttext.New(value.value).
		MaxLines(1).
		Truncator("...").
		Style(style.TextDeclaration(resolved.Text)).
		Layout(ctx, gtx)
}

type indicatorWidget struct {
	progress float32
	visual   theme.CollapsibleVisual
}

func (w indicatorWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	if w.visual.Indicator == nil {
		panic("flowui: collapsible visual has no indicator renderer")
	}
	return w.visual.Indicator(gtx, theme.CollapsibleIndicatorRequest{
		Progress: w.progress,
		Disabled: !gtx.Enabled(),
	})
}

func layoutContent(ctx *core.Context, gtx layout.Context, content core.Widget, heightProgress, opacityProgress float32, bodyPadding unit.Dp) layout.Dimensions {
	if content == nil || (heightProgress <= 0 && opacityProgress <= 0) {
		return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
	}
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	gtx.Constraints.Min.Y = 0
	macro := op.Record(gtx.Ops)
	fullDims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return layout.UniformInset(bodyPadding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
