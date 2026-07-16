package colorpicker

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type ColorChannel uint8

const (
	ColorChannelHue ColorChannel = iota
	ColorChannelAlpha
)

type ColorSwatchSize uint8

const (
	ColorSwatchExtraSmall ColorSwatchSize = iota
	ColorSwatchSmall
	ColorSwatchMedium
	ColorSwatchLarge
	ColorSwatchExtraLarge
)

type ColorSwatchShape uint8

const (
	ColorSwatchCircle ColorSwatchShape = iota
	ColorSwatchSquare
)

type ColorSwatchPickerLayout uint8

const (
	ColorSwatchPickerGrid ColorSwatchPickerLayout = iota
	ColorSwatchPickerStack
)

type colorValueState struct {
	syncedColor color.NRGBA
	ready       bool
	retainedHue float64
	hueReady    bool
}

func (state *colorValueState) sync(value color.NRGBA) {
	if state.ready && state.syncedColor == value {
		return
	}
	resolved := nrgbaToHSV(value)
	rgbChanged := !state.ready || state.syncedColor.R != value.R || state.syncedColor.G != value.G || state.syncedColor.B != value.B
	if !state.hueReady || rgbChanged && resolved.s > 0 && resolved.v > 0 {
		state.retainedHue = resolved.h
		state.hueReady = true
	}
	state.syncedColor = value
	state.ready = true
}

func (state *colorValueState) accept(value color.NRGBA, hue float64) {
	state.syncedColor = value
	state.ready = true
	state.retainedHue = clampUnit(hue)
	state.hueReady = true
}

func (state *colorValueState) hsv() hsvColor {
	resolved := nrgbaToHSV(state.syncedColor)
	if state.hueReady {
		resolved.h = state.retainedHue
	}
	return resolved
}

func colorSwatchPixels(ctx *frame.Context, gtx layout.Context, size ColorSwatchSize) int {
	tokens := frame.ActiveTheme(ctx).Components.ColorSwatch
	switch size {
	case ColorSwatchExtraSmall:
		return gtx.Dp(tokens.ExtraSmallSize)
	case ColorSwatchSmall:
		return gtx.Dp(tokens.SmallSize)
	case ColorSwatchLarge:
		return gtx.Dp(tokens.LargeSize)
	case ColorSwatchExtraLarge:
		return gtx.Dp(tokens.ExtraLargeSize)
	default:
		return gtx.Dp(tokens.MediumSize)
	}
}

func colorSwatchRadius(ctx *frame.Context, gtx layout.Context, size int, shape ColorSwatchShape) int {
	if shape == ColorSwatchCircle {
		return size / 2
	}
	return min(max(gtx.Dp(frame.ActiveTheme(ctx).Components.ColorSwatch.SquareRadius), 1), size/2)
}

func colorSwatchPickerBorderWidth(ctx *frame.Context, gtx layout.Context, size ColorSwatchSize) int {
	tokens := frame.ActiveTheme(ctx).Components.ColorSwatchPicker
	switch size {
	case ColorSwatchExtraSmall:
		return max(gtx.Dp(tokens.ExtraSmallBorderWidth), 1)
	case ColorSwatchLarge, ColorSwatchExtraLarge:
		return max(gtx.Dp(tokens.LargeBorderWidth), 1)
	default:
		return max(gtx.Dp(tokens.BorderWidth), 1)
	}
}

func colorSwatchPickerItemRadius(ctx *frame.Context, gtx layout.Context, side int, size ColorSwatchSize, shape ColorSwatchShape) int {
	if shape == ColorSwatchCircle {
		return side / 2
	}
	tokens := frame.ActiveTheme(ctx).Components.ColorSwatchPicker
	radius := tokens.SquareItemRadius
	if size == ColorSwatchExtraSmall {
		radius = tokens.SquareItemRadiusExtraSmall
	} else if size == ColorSwatchSmall {
		radius = tokens.SquareItemRadiusSmall
	}
	return min(max(gtx.Dp(radius), 1), side/2)
}

func colorSwatchPickerItemRadiusDp(ctx *frame.Context, size ColorSwatchSize) unit.Dp {
	tokens := frame.ActiveTheme(ctx).Components.ColorSwatchPicker
	if size == ColorSwatchExtraSmall {
		return tokens.SquareItemRadiusExtraSmall
	}
	if size == ColorSwatchSmall {
		return tokens.SquareItemRadiusSmall
	}
	return tokens.SquareItemRadius
}

func colorSwatchPickerSwatchRadius(ctx *frame.Context, gtx layout.Context, side int, size ColorSwatchSize, shape ColorSwatchShape, selection float32) int {
	if shape == ColorSwatchCircle {
		return side / 2
	}
	tokens := frame.ActiveTheme(ctx).Components.ColorSwatchPicker
	radius := tokens.SquareSwatchRadius
	switch size {
	case ColorSwatchExtraSmall:
		radius = tokens.SquareSwatchRadiusExtraSmall
	case ColorSwatchSmall:
		radius = unit.Dp(float32(radius) + (float32(tokens.SquareSelectedSmallRadius)-float32(radius))*selection)
	}
	return min(max(gtx.Dp(radius), 1), side/2)
}
