package testtheme

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Modal(active *theme.Theme, request theme.ModalRequest) theme.ModalVisual {
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
		IconFrame: func(layout.Context, theme.ModalIconFrameRequest) {},
		Padding: func(resolved style.ResolvedStyle) theme.ModalPadding {
			if resolved.Box == nil || resolved.Box.Padding == nil {
				return theme.ModalPadding{}
			}
			padding := resolved.Box.Padding
			return theme.ModalPadding{
				Top: padding.Top, Right: padding.Right, Bottom: padding.Bottom, Left: padding.Left,
			}
		},
	}
}
