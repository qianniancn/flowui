// Package modalstyle contains the theme-owned visual recipe for modal
// behavior. The widget supplies size/backdrop choices and executes the
// returned declarations; this package owns all visual values and drawing.
package modalstyle

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func BuildModal(active *theme.Theme, request theme.ModalRequest) theme.ModalVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Modal
	sem := active.Semantic()
	backdrop := tokens.Backdrop
	switch request.Backdrop {
	case theme.ModalBackdropBlur:
		backdrop = tokens.BlurBackdrop
	case theme.ModalBackdropTransparent:
		backdrop = color.NRGBA{}
	}
	defaults := style.Style{}.
		Background(style.Solid(sem.Overlay)).
		TextColor(style.Solid(sem.OverlayForeground)).
		Padding(tokens.Padding).
		Overflow(style.OverflowHidden).
		Part(style.PartBackdrop, style.Style{}.Background(style.Solid(backdrop)))
	if request.Size != theme.ModalSizeFull {
		defaults = defaults.Radius(tokens.Radius).Shadow(style.ShadowOverlay)
	}
	return theme.ModalVisual{
		Defaults: defaults,
		Header: style.Style{}.
			FontSize(tokens.TitleSize).
			FontWeight(500).
			TextColor(style.Solid(sem.OverlayForeground)),
		Body: style.Style{}.
			FontSize(tokens.BodyTextSize).
			TextColor(style.Solid(sem.MutedForeground)),
		Override: theme.ModalOverrideStyle(request),
		Metrics: theme.ModalMetrics{
			XSmallWidth: tokens.XSmallWidth, SmallWidth: tokens.SmallWidth,
			MediumWidth: tokens.MediumWidth, LargeWidth: tokens.LargeWidth,
			Margin: tokens.Margin, DesktopMargin: tokens.DesktopMargin,
			DesktopBreakpoint: tokens.DesktopBreakpoint, Radius: tokens.Radius,
			Padding: tokens.Padding, HeaderGap: tokens.HeaderGap, BodyGap: tokens.BodyGap,
			FooterGap: tokens.FooterGap, SectionGap: tokens.SectionGap,
			IconSize: tokens.IconSize, CloseInset: tokens.CloseInset,
			CloseButtonSize: active.Components.CloseButton.Size,
			TitleSize:       tokens.TitleSize, BodyTextSize: tokens.BodyTextSize,
			AnimationScale: tokens.AnimationScale, AnimationDuration: tokens.AnimationDuration,
			AnimationDistance:          tokens.AnimationDistance,
			AnimationBounceScale:       tokens.AnimationBounceScale,
			AnimationBounceSettleScale: tokens.AnimationBounceSettleScale,
			AnimationZoomScale:         tokens.AnimationZoomScale,
			AnimationPopStartScale:     tokens.AnimationPopStartScale,
			AnimationPopOvershootScale: tokens.AnimationPopOvershootScale,
			AnimationPopSettleScale:    tokens.AnimationPopSettleScale,
		},
		Motion:    active.Motion,
		IconFrame: iconRenderer(sem.Tone(theme.TonePrimary).Soft),
		Padding:   paddingResolver,
	}
}

func paddingResolver(resolved style.ResolvedStyle) theme.ModalPadding {
	if resolved.Box == nil || resolved.Box.Padding == nil {
		return theme.ModalPadding{}
	}
	padding := resolved.Box.Padding
	return theme.ModalPadding{
		Top: padding.Top, Right: padding.Right, Bottom: padding.Bottom, Left: padding.Left,
	}
}

func iconRenderer(colorValue color.NRGBA) theme.ModalIconFrameRenderer {
	return func(gtx layout.Context, request theme.ModalIconFrameRequest) {
		if request.Size.X <= 0 || request.Size.Y <= 0 {
			return
		}
		radius := min(request.Size.X, request.Size.Y) / 2
		paint.FillShape(gtx.Ops, colorValue, clip.UniformRRect(image.Rectangle{Max: request.Size}, radius).Op(gtx.Ops))
	}
}
