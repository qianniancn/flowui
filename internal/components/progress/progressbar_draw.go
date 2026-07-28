package progress

import (
	"time"

	"github.com/qianniancn/flowui/internal/render"
)

func progressBarIndeterminateOffset(now time.Time, fillWidth int, period time.Duration) int {
	if fillWidth <= 0 || now.IsZero() || period <= 0 {
		return -fillWidth
	}
	elapsed := now.UnixNano() % int64(period)
	if elapsed < 0 {
		elapsed += int64(period)
	}
	progress := render.Ease(float32(elapsed) / float32(period))
	return int(render.Lerp(float32(-fillWidth), float32(fillWidth)*3.5, progress) + 0.5)
}

func progressBarIndeterminatePosition(now time.Time, fillWidth int, period time.Duration) int {
	if period <= 0 {
		return 0
	}
	return progressBarIndeterminateOffset(now, fillWidth, period)
}
