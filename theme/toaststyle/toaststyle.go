// Package toaststyle contains the theme-owned visual recipe for toast
// notifications. It does not import widget/toast; stacking, timers, focus,
// and actions remain behavior-owned by the widget.
package toaststyle

import (
	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
	"image"
	"image/color"
)

// BuildToast returns the complete toast visual plan for the active theme.
func BuildToast(active *theme.Theme, request theme.ToastRequest) theme.ToastVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Toast
	sem := active.Semantic()
	componentProvider, err := theme.ComponentProvider(active)
	if err != nil {
		panic(err)
	}
	title := sem.OverlayForeground
	indicator := title
	switch request.Variant {
	case theme.ToastAccent:
		title = appearance.Tone(active, theme.TonePrimary).OnSoft
	case theme.ToastSuccess:
		title = appearance.Tone(active, theme.ToneSuccess).OnSoft
		indicator = title
	case theme.ToastWarning:
		title = appearance.Tone(active, theme.ToneWarning).OnSoft
		indicator = title
	case theme.ToastDanger:
		title = appearance.Tone(active, theme.ToneDanger).OnSoft
		indicator = title
	}
	defaults := style.Style{}.
		Background(style.Solid(sem.Surface)).
		TextColor(style.Solid(sem.OverlayForeground)).
		PaddingX(tokens.PaddingX).
		PaddingY(tokens.PaddingY).
		Radius(tokens.Radius).
		Shadow(style.ShadowOverlay).
		Overflow(style.OverflowHidden).
		Opacity(1).
		OnNot(style.FocusVisible, style.Disabled,
			style.Style{}.Outline(tokens.FocusRingWidth, 0, style.Solid(sem.Ring)))
	spinnerVisual := componentProvider.Spinner(active, theme.SpinnerRequest{Color: theme.SpinnerCurrent, Size: theme.SpinnerSmall})
	closeSize := tokens.CloseButtonSize
	if closeSize <= 0 {
		closeSize = active.Components.CloseButton.Size
	}
	return theme.ToastVisual{
		Defaults: defaults,
		Override: theme.ToastOverrideStyle(request),
		Metrics: theme.ToastMetrics{
			Width: tokens.Width, Inset: tokens.Inset, Gap: tokens.Gap,
			ContentGap: tokens.ContentGap, ActionInsetTop: tokens.ActionInsetTop,
			IndicatorSize: tokens.IndicatorSize,
			CloseInset:    tokens.CloseInset, CloseButtonSize: closeSize,
			MaxVisible:  tokens.MaxVisible,
			ScaleFactor: tokens.ScaleFactor, AnimationDuration: tokens.AnimationDuration,
			DefaultTimeout: tokens.DefaultTimeout,
		},
		Motion:    active.Motion,
		Text:      textRenderer(active, tokens, title, sem.MutedForeground),
		Indicator: indicatorRenderer(active, tokens, indicator, toastIcon(request.Variant), spinnerVisual),
		Action:    actionRenderer(active, tokens),
		Close:     closeRenderer(active, tokens),
	}
}

func textRenderer(active *theme.Theme, tokens theme.ToastTheme, title, description color.NRGBA) theme.ToastTextRenderer {
	return func(gtx layout.Context, request theme.ToastTextRenderRequest) layout.Dimensions {
		size := tokens.DescriptionSize
		weight := font.Normal
		foreground := description
		if request.Role == theme.ToastTextTitle {
			size = tokens.TitleSize
			weight = font.Medium
			foreground = title
		}
		label := materialLabel(active, size, request.Text, foreground)
		label.Font.Weight = weight
		return label.Layout(gtx)
	}
}

func indicatorRenderer(active *theme.Theme, tokens theme.ToastTheme, foreground color.NRGBA, glyph []byte, spinner theme.SpinnerVisual) theme.ToastIndicatorRenderer {
	return func(gtx layout.Context, request theme.ToastIndicatorRenderRequest) layout.Dimensions {
		outer := gtx.Dp(request.Size + 2*tokens.IndicatorPadding)
		outer = max(outer, 1)
		gtx.Constraints = layout.Exact(image.Pt(outer, outer))
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			size := max(gtx.Dp(request.Size), 1)
			gtx.Constraints = layout.Exact(image.Pt(size, size))
			if request.Loading {
				if spinner.Renderer == nil {
					panic(theme.ErrIncompleteVisualProvider)
				}
				period := theme.ResolveMotionDuration(spinner.Motion, spinner.Metrics.SpinDuration)
				dims := spinner.Renderer(gtx, theme.SpinnerRenderRequest{
					Size: request.Size, StrokeRatio: spinner.Metrics.StrokeRatio,
					InsetRatio: spinner.Metrics.InsetRatio, Color: foreground, Period: period,
				})
				if period > 0 {
					gtx.Execute(op.InvalidateCmd{})
				}
				return dims
			}
			if request.Custom != nil {
				return request.Custom(gtx)
			}
			return icon.Layout(glyph, gtx, foreground)
		})
	}
}

func actionRenderer(active *theme.Theme, tokens theme.ToastTheme) theme.ToastActionRenderer {
	return func(gtx layout.Context, request theme.ToastActionRenderRequest) layout.Dimensions {
		if request.Clickable == nil {
			panic(theme.ErrIncompleteVisualProvider)
		}
		background, foreground := toastButtonColors(active, request.Variant)
		if request.State.Hovered {
			background = active.Semantic().SurfacePressed
		}
		if request.State.Pressed {
			background = active.Semantic().SurfacePressed
		}
		border := color.NRGBA{}
		if request.Variant == theme.ButtonSecondary {
			border = appearance.Tone(active, theme.TonePrimary).Fill
		}
		label := materialLabel(active, tokens.DescriptionSize, request.Label, foreground)
		label.Font.Weight = font.Medium
		measure := gtx
		measure.Constraints.Min = image.Point{}
		content := label.Layout(measure)
		paddingX := max(gtx.Dp(tokens.PaddingX*3/4), 10)
		paddingY := max(gtx.Dp(tokens.PaddingY/2), 5)
		size := gtx.Constraints.Constrain(image.Pt(content.Size.X+2*paddingX, content.Size.Y+2*paddingY))
		return request.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.Button.Add(gtx.Ops)
			semantic.EnabledOp(true).Add(gtx.Ops)
			semantic.LabelOp(request.Label).Add(gtx.Ops)
			rect := image.Rectangle{Max: size}
			radius := max(gtx.Dp(tokens.Radius/2), 1)
			if border.A != 0 {
				paint.FillShape(gtx.Ops, border, clip.UniformRRect(rect, radius).Op(gtx.Ops))
				rect = rect.Inset(gtx.Dp(1))
				radius = max(radius-gtx.Dp(1), 1)
			}
			paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
			contentGtx := gtx
			contentGtx.Constraints = layout.Exact(image.Pt(max(size.X-2*paddingX, 0), max(size.Y-2*paddingY, 0)))
			return layout.Center.Layout(contentGtx, func(gtx layout.Context) layout.Dimensions {
				return label.Layout(gtx)
			})
		})
	}
}

func closeRenderer(active *theme.Theme, tokens theme.ToastTheme) theme.ToastCloseRenderer {
	return func(gtx layout.Context, request theme.ToastCloseRenderRequest) layout.Dimensions {
		if request.Clickable == nil {
			panic(theme.ErrIncompleteVisualProvider)
		}
		size := max(gtx.Dp(request.Size), gtx.Dp(tokens.CloseButtonSize))
		size = max(size, 1)
		gtx.Constraints = layout.Exact(image.Pt(size, size))
		sem := active.Semantic()
		background := sem.SurfacePressed
		if request.State.Hovered {
			background = sem.SurfacePressed
		}
		return request.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.Button.Add(gtx.Ops)
			semantic.EnabledOp(true).Add(gtx.Ops)
			semantic.LabelOp(request.Label).Add(gtx.Ops)
			paint.FillShape(gtx.Ops, background, clip.UniformRRect(image.Rectangle{Max: image.Pt(size, size)}, size/2).Op(gtx.Ops))
			iconSize := max(size/2, 1)
			iconGtx := gtx
			iconGtx.Constraints = layout.Exact(image.Pt(iconSize, iconSize))
			return layout.Center.Layout(iconGtx, func(gtx layout.Context) layout.Dimensions {
				return icon.Layout(lucide.X, gtx, sem.MutedForeground)
			})
		})
	}
}

func materialLabel(active *theme.Theme, size unit.Sp, value string, foreground color.NRGBA) material.LabelStyle {
	label := material.Label(theme.MaterialOf(active), size, value)
	label.Color = foreground
	return label
}

func toastButtonColors(active *theme.Theme, variant theme.ButtonVariant) (color.NRGBA, color.NRGBA) {
	switch variant {
	case theme.ButtonDanger, theme.ButtonDangerOutline, theme.ButtonDangerSoft:
		tone := appearance.Tone(active, theme.ToneDanger)
		if variant == theme.ButtonDanger {
			return tone.Fill, tone.OnFill
		}
		return tone.Soft, tone.OnSoft
	case theme.ButtonSecondary:
		tone := appearance.Tone(active, theme.TonePrimary)
		return tone.Soft, tone.OnSoft
	case theme.ButtonTertiary, theme.ButtonGhost, theme.ButtonOutline:
		return color.NRGBA{}, appearance.Tone(active, theme.TonePrimary).OnSoft
	default:
		tone := appearance.Tone(active, theme.TonePrimary)
		return tone.Fill, tone.OnFill
	}
}

func toastIcon(variant theme.ToastVariant) []byte {
	switch variant {
	case theme.ToastSuccess:
		return lucide.CircleCheck
	case theme.ToastWarning:
		return lucide.TriangleAlert
	case theme.ToastDanger:
		return lucide.CircleAlert
	default:
		return lucide.Info
	}
}
