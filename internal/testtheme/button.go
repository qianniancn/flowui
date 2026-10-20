package testtheme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	semanticSize := request.Size
	height, padding, fontSize := unit.Dp(40), unit.Dp(16), unit.Sp(14)
	switch request.Size {
	case theme.ButtonSmall:
		height, padding, fontSize = 32, 12, 14
	case theme.ButtonLarge:
		height, padding, fontSize = 48, 20, 16
	}
	if request.IconOnly {
		padding = 0
	}
	defaults := style.Style{}.
		Radius(4).
		FontWeight(500).
		Outline(2, 1, style.RGBA(0)).
		Opacity(1).
		Transition(style.PropBackgroundColor, 100*time.Millisecond).
		Transition(style.PropTextColor, 100*time.Millisecond).
		Transition(style.PropBorderColor, 100*time.Millisecond).
		Transition(style.PropTransform, 160*time.Millisecond).
		Transition(style.PropOpacity, 100*time.Millisecond).
		On(style.Disabled, style.Style{}.Opacity(active.DisabledOpacityValue())).
		OnNot(style.FocusVisible, style.Disabled, style.Style{}.Outline(2, 1, style.Solid(focus)))
	variant := style.Style{}.
		Background(style.Solid(background)).
		TextColor(style.Solid(foreground)).
		On(style.Hovered, style.Style{}.Background(style.Solid(accent))).
		On(style.Pressed, style.Style{}.Background(style.Solid(muted)))
	size := style.Style{}.Height(height).PaddingX(padding).FontSize(fontSize)
	if request.IconOnly {
		size = size.Width(height)
	}
	if request.Grouped {
		group := style.Style{}.Radius(0)
		if request.GroupCorners.NorthWest {
			group = group.RadiusTopLeft(4)
		}
		if request.GroupCorners.NorthEast {
			group = group.RadiusTopRight(4)
		}
		if request.GroupCorners.SouthEast {
			group = group.RadiusBottomRight(4)
		}
		if request.GroupCorners.SouthWest {
			group = group.RadiusBottomLeft(4)
		}
		size = style.Join(size, group)
	}
	if !request.Grouped {
		size = size.Rules(style.Rule{When: style.Pressed, Set: style.Style{}.Scale(.98), Tween: &style.Tween{Property: style.PropTransform, Duration: 90 * time.Millisecond}})
	}
	return theme.ButtonVisual{
		Defaults: defaults, Variant: variant, Size: size, Override: buttonOverride(request),
		Ripple: true, RippleDuration: 180 * time.Millisecond, RippleFadeDuration: 100 * time.Millisecond,
		Spinner:            neutralSpinner,
		SpinnerSize:        func() unit.Dp { return neutralSpinnerSize(semanticSize) },
		SpinnerStrokeWidth: func() unit.Dp { return 2 },
		SpinnerPeriod:      func() time.Duration { return 900 * time.Millisecond },
		ContentGap:         neutralButtonContentGap,
		LoadingInset: func(gtx layout.Context, loading theme.ButtonLoadingRequest) layout.Inset {
			return neutralLoadingInset(gtx, loading, semanticSize)
		},
		ForegroundColor: neutralForegroundColor,
		RippleRenderer:  neutralRipple,
		RippleColor:     neutralRippleColor,
		Motion:          active.Motion,
		Group: theme.ButtonGroupVisual{
			SeparatorRenderer: func(layout.Context, theme.ButtonGroupSeparatorRequest) {},
		},
	}
}

func neutralSpinner(gtx layout.Context, request theme.ButtonSpinnerRequest) layout.Dimensions {
	d := max(gtx.Dp(request.Size), 1)
	paint.FillShape(gtx.Ops, request.Color, clip.Ellipse{Max: image.Pt(d, d)}.Op(gtx.Ops))
	return layout.Dimensions{Size: image.Pt(d, d)}
}

func neutralSpinnerSize(size theme.ButtonSize) unit.Dp {
	switch size {
	case theme.ButtonSmall:
		return 12
	case theme.ButtonLarge:
		return 20
	default:
		return 16
	}
}

func neutralButtonContentGap(theme.ButtonContentGapRequest) unit.Dp {
	return 8
}

func neutralLoadingInset(gtx layout.Context, request theme.ButtonLoadingRequest, size theme.ButtonSize) layout.Inset {
	spinner := neutralSpinnerSize(size)
	contentGap := neutralButtonContentGap(theme.ButtonContentGapRequest{Loading: true})
	padding := gtx.Dp(request.Inset.Left) + gtx.Dp(request.Inset.Right)
	remaining := max(padding-gtx.Dp(spinner)-gtx.Dp(contentGap), 0)
	left := remaining / 2
	request.Inset.Left = gtx.Metric.PxToDp(left)
	request.Inset.Right = gtx.Metric.PxToDp(remaining - left)
	return request.Inset
}

func neutralForegroundColor(resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if resolved.Text == nil {
		return color.NRGBA{}, false
	}
	foreground, ok := resolved.Text.Color.(style.SolidColor)
	if !ok {
		if pointer, pointerOK := resolved.Text.Color.(*style.SolidColor); pointerOK && pointer != nil {
			return pointer.Color, true
		}
		return color.NRGBA{}, false
	}
	return foreground.Color, true
}

func neutralRipple(layout.Context, theme.ButtonRippleRequest) {}

func neutralRippleColor(resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if resolved.Paint == nil || resolved.Paint.Background == nil {
		return color.NRGBA{}, false
	}
	background, ok := resolved.Paint.Background.(style.SolidColor)
	return background.Color, ok && background.Color.A != 0
}

func buttonOverride(request theme.ButtonRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}
