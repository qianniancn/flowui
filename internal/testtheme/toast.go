package testtheme

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Toast(active *theme.Theme, request theme.ToastRequest) theme.ToastVisual {
	tokens := active.Components.Toast
	return theme.ToastVisual{
		Defaults: style.Style{}.
			Background(style.Solid(background)).
			TextColor(style.Solid(foreground)).
			PaddingX(tokens.PaddingX).
			PaddingY(tokens.PaddingY).
			Radius(tokens.Radius).
			Overflow(style.OverflowHidden).
			Opacity(1),
		Override: theme.ToastOverrideStyle(request),
		Metrics: theme.ToastMetrics{
			Width: tokens.Width, Inset: tokens.Inset, Gap: tokens.Gap,
			ContentGap: tokens.ContentGap, ActionInsetTop: tokens.ActionInsetTop,
			IndicatorSize: tokens.IndicatorSize, CloseInset: tokens.CloseInset,
			CloseButtonSize: active.Components.CloseButton.Size, MaxVisible: tokens.MaxVisible,
			ScaleFactor: tokens.ScaleFactor, AnimationDuration: tokens.AnimationDuration,
			DefaultTimeout: tokens.DefaultTimeout,
		},
		Motion:    active.Motion,
		Text:      neutralToastTextRenderer(active, tokens),
		Indicator: neutralToastIndicatorRenderer,
		Action:    neutralToastActionRenderer,
		Close:     neutralToastCloseRenderer,
	}
}

func neutralToastTextRenderer(active *theme.Theme, tokens theme.ToastTheme) theme.ToastTextRenderer {
	return func(gtx layout.Context, request theme.ToastTextRenderRequest) layout.Dimensions {
		size := tokens.DescriptionSize
		weight := font.Normal
		if request.Role == theme.ToastTextTitle {
			size = tokens.TitleSize
			weight = font.Medium
		}
		label := material.Label(theme.MaterialOf(active), size, request.Text)
		label.Color = foreground
		label.Font.Weight = weight
		return label.Layout(gtx)
	}
}

func neutralToastIndicatorRenderer(gtx layout.Context, request theme.ToastIndicatorRenderRequest) layout.Dimensions {
	size := max(gtx.Dp(request.Size), 1)
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	if request.Custom != nil {
		return request.Custom(gtx)
	}
	return layout.Dimensions{Size: image.Pt(size, size)}
}

func neutralToastActionRenderer(gtx layout.Context, request theme.ToastActionRenderRequest) layout.Dimensions {
	if request.Clickable == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	width := max(8*len([]rune(request.Label))+16, 24)
	size := gtx.Constraints.Constrain(image.Pt(width, 24))
	return request.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(true).Add(gtx.Ops)
		semantic.LabelOp(request.Label).Add(gtx.Ops)
		return layout.Dimensions{Size: size}
	})
}

func neutralToastCloseRenderer(gtx layout.Context, request theme.ToastCloseRenderRequest) layout.Dimensions {
	if request.Clickable == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	size := max(gtx.Dp(request.Size), 1)
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	return request.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.EnabledOp(true).Add(gtx.Ops)
		semantic.LabelOp(request.Label).Add(gtx.Ops)
		return layout.Dimensions{Size: image.Pt(size, size)}
	})
}
