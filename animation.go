package flowui

import (
	"image/color"
	"time"

	"github.com/qianniancn/FlowUI/render"
)

func animationProgress(elapsed, duration time.Duration) float32 {
	return render.Progress(elapsed, duration)
}

func animationEase(progress float32) float32 {
	return render.Ease(progress)
}

func lerp(from, to, progress float32) float32 {
	return render.Lerp(from, to, progress)
}

func lerpColor(from, to color.NRGBA, progress float32) color.NRGBA {
	return render.LerpColor(from, to, progress)
}

func lerpByte(from, to byte, progress float32) byte {
	return render.LerpByte(from, to, progress)
}
