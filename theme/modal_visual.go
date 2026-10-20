package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// ModalSize and ModalBackdrop are the neutral values used by the modal
// provider protocol. Their numeric values are intentionally independent of
// widget/modal so theme packages do not import widget behavior.
type ModalSize uint8

const (
	ModalSizeMedium ModalSize = iota
	ModalSizeXSmall
	ModalSizeSmall
	ModalSizeLarge
	ModalSizeCover
	ModalSizeFull
)

type ModalBackdrop uint8

const (
	ModalBackdropOpaque ModalBackdrop = iota
	ModalBackdropBlur
	ModalBackdropTransparent
)

// ModalRequest contains only semantic choices made by modal behavior.
type ModalRequest struct {
	Size     ModalSize
	Backdrop ModalBackdrop
	Override *ModalOverride
}

// ModalOverride carries caller customization across the provider boundary.
type ModalOverride struct {
	Declaration style.Style
}

func ModalOverrideStyle(request ModalRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

type ModalMetrics struct {
	XSmallWidth, SmallWidth, MediumWidth, LargeWidth                            unit.Dp
	Margin, DesktopMargin, DesktopBreakpoint                                    unit.Dp
	Radius, Padding                                                             unit.Dp
	HeaderGap, BodyGap, FooterGap, SectionGap                                   unit.Dp
	IconSize, CloseInset, CloseButtonSize                                       unit.Dp
	TitleSize, BodyTextSize                                                     unit.Sp
	AnimationScale, AnimationBounceScale                                        float32
	AnimationDuration                                                           time.Duration
	AnimationBounceSettleScale                                                  float32
	AnimationZoomScale                                                          float32
	AnimationPopStartScale, AnimationPopOvershootScale, AnimationPopSettleScale float32
	AnimationDistance                                                           unit.Dp
}

// ModalPadding contains the resolved panel insets needed by close-button
// placement. The provider interprets the cascade; the widget only converts
// these neutral measurements to pixels for layout.
type ModalPadding struct {
	Top, Right, Bottom, Left unit.Dp
}

// ModalPaddingResolver extracts provider-owned panel insets from a resolved
// modal root declaration, including instance overrides.
type ModalPaddingResolver func(style.ResolvedStyle) ModalPadding

type ModalIconFrameRequest struct {
	Size  image.Point
	Color color.NRGBA
}

type ModalIconFrameRenderer func(layout.Context, ModalIconFrameRequest)

type ModalVisual struct {
	Defaults  style.Style
	Header    style.Style
	Body      style.Style
	Override  style.Style
	Metrics   ModalMetrics
	Motion    MotionTheme
	IconFrame ModalIconFrameRenderer
	Padding   ModalPaddingResolver
}

// ModalProvider supplies modal declarations, geometry, motion, and icon
// treatment without taking ownership of overlay behavior or focus state.
type ModalProvider interface {
	Modal(active *Theme, request ModalRequest) ModalVisual
}

// ValidateModalVisual checks the close-button inset and icon-frame bridges
// before a widget stores a provider result in its declaration cache. Modal
// behavior may apply these measurements, but it must not interpret a BoxStyle
// or recover a missing renderer locally.
func ValidateModalVisual(visual ModalVisual) {
	if visual.IconFrame == nil {
		invalidVisual("Modal", "IconFrame", "must be provided")
	}
	if visual.Padding == nil {
		invalidVisual("Modal", "Padding", "must be provided")
	}

	metrics := visual.Metrics
	if metrics.XSmallWidth <= 0 {
		invalidVisual("Modal", "Metrics.XSmallWidth", "must be greater than 0")
	}
	if metrics.SmallWidth <= 0 {
		invalidVisual("Modal", "Metrics.SmallWidth", "must be greater than 0")
	}
	if metrics.MediumWidth <= 0 {
		invalidVisual("Modal", "Metrics.MediumWidth", "must be greater than 0")
	}
	if metrics.LargeWidth <= 0 {
		invalidVisual("Modal", "Metrics.LargeWidth", "must be greater than 0")
	}
	if metrics.Margin < 0 {
		invalidVisual("Modal", "Metrics.Margin", "must be at least 0")
	}
	if metrics.DesktopMargin < 0 {
		invalidVisual("Modal", "Metrics.DesktopMargin", "must be at least 0")
	}
	if metrics.DesktopBreakpoint <= 0 {
		invalidVisual("Modal", "Metrics.DesktopBreakpoint", "must be greater than 0")
	}
	if metrics.Radius < 0 {
		invalidVisual("Modal", "Metrics.Radius", "must be at least 0")
	}
	if metrics.Padding < 0 {
		invalidVisual("Modal", "Metrics.Padding", "must be at least 0")
	}
	if metrics.HeaderGap < 0 {
		invalidVisual("Modal", "Metrics.HeaderGap", "must be at least 0")
	}
	if metrics.BodyGap < 0 {
		invalidVisual("Modal", "Metrics.BodyGap", "must be at least 0")
	}
	if metrics.FooterGap < 0 {
		invalidVisual("Modal", "Metrics.FooterGap", "must be at least 0")
	}
	if metrics.SectionGap < 0 {
		invalidVisual("Modal", "Metrics.SectionGap", "must be at least 0")
	}
	if metrics.IconSize <= 0 {
		invalidVisual("Modal", "Metrics.IconSize", "must be greater than 0")
	}
	if metrics.CloseInset < 0 {
		invalidVisual("Modal", "Metrics.CloseInset", "must be at least 0")
	}
	if metrics.CloseButtonSize <= 0 {
		invalidVisual("Modal", "Metrics.CloseButtonSize", "must be greater than 0")
	}
	if metrics.TitleSize <= 0 {
		invalidVisual("Modal", "Metrics.TitleSize", "must be greater than 0")
	}
	if metrics.BodyTextSize <= 0 {
		invalidVisual("Modal", "Metrics.BodyTextSize", "must be greater than 0")
	}
	if metrics.AnimationScale <= 0 {
		invalidVisual("Modal", "Metrics.AnimationScale", "must be greater than 0")
	}
	if metrics.AnimationScale > 1 {
		invalidVisual("Modal", "Metrics.AnimationScale", "must be at most 1")
	}
	if metrics.AnimationDuration <= 0 {
		invalidVisual("Modal", "Metrics.AnimationDuration", "must be greater than 0")
	}
	if metrics.AnimationDistance < 0 {
		invalidVisual("Modal", "Metrics.AnimationDistance", "must be at least 0")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("Modal", "Root.Box", "must be provided")
	}
	if root.Box.Padding == nil {
		invalidVisual("Modal", "Root.Box.Padding", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("Modal", "Root.Paint", "must be provided")
	}
	if root.Paint.Background == nil {
		invalidVisual("Modal", "Root.Paint.Background", "must be provided")
	}

	backdrop := style.CascadePart(style.State{}, style.PartBackdrop, visual.Defaults, visual.Override)
	if backdrop.Paint == nil {
		invalidVisual("Modal", "Backdrop.Paint", "must be provided")
	}
	if backdrop.Paint.Background == nil {
		invalidVisual("Modal", "Backdrop.Paint.Background", "must be provided")
	}
}
