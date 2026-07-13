package render

import (
	"image/color"
	"time"

	animationcore "github.com/qianniancn/FlowUI/internal/animation/core"
)

func Progress(elapsed, duration time.Duration) float32 {
	return animationcore.Progress(elapsed, duration)
}

func Ease(progress float32) float32 {
	return animationcore.EaseSmoothstep(progress)
}

func Lerp(from, to, progress float32) float32 {
	return animationcore.LerpFloat(from, to, progress)
}

func LerpColor(from, to color.NRGBA, progress float32) color.NRGBA {
	return animationcore.LerpColor(from, to, progress)
}

func LerpByte(from, to byte, progress float32) byte {
	return animationcore.LerpByte(from, to, progress)
}
