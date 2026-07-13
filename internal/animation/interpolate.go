package animation

import (
	"image"
	"image/color"

	"gioui.org/f32"
	animationcore "github.com/qianniancn/FlowUI/internal/animation/core"
)

func LerpFloat(from, to, progress float32) float32 {
	return animationcore.LerpFloat(from, to, progress)
}

func LerpFloat64(from, to float64, progress float32) float64 {
	return animationcore.LerpFloat64(from, to, progress)
}

func LerpColor(from, to color.NRGBA, progress float32) color.NRGBA {
	return animationcore.LerpColor(from, to, progress)
}

func LerpPoint(from, to f32.Point, progress float32) f32.Point {
	return animationcore.LerpPoint(from, to, progress)
}

func LerpRect(from, to image.Rectangle, progress float32) image.Rectangle {
	return animationcore.LerpRect(from, to, progress)
}
