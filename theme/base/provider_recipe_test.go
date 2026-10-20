package base_test

import (
	"testing"
	"time"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
)

// TestImageCodeSkeletonRecipesReadTokens pins that the newly protocol-backed
// components take their geometry from the theme instead of the widget.
func TestImageCodeSkeletonRecipesReadTokens(t *testing.T) {
	active := base.Light()
	active.Components.Image.Radius = 10
	active.Components.Code.Padding = 20
	active.Components.Code.Radius = 9
	active.Components.Code.HeaderGap = 14
	active.Components.Code.LineNumberGap = 10
	active.Components.Code.TextSize = 16
	active.Components.Skeleton.Height = 24
	active.Components.Skeleton.Radius = 8
	active.Components.Skeleton.Period = 2 * time.Second

	provider := base.Provider{}
	image := provider.Image(&active, theme.ImageRequest{})
	if image.Metrics.Radius != 10 {
		t.Fatalf("image radius = %v, want 10", image.Metrics.Radius)
	}
	code := provider.Code(&active, theme.CodeRequest{})
	if code.Metrics.Padding != 20 || code.Metrics.Radius != 9 || code.Metrics.HeaderGap != 14 ||
		code.Metrics.LineNumberGap != 10 || code.Metrics.TextSize != 16 {
		t.Fatalf("code metrics = %+v", code.Metrics)
	}
	skeleton := provider.Skeleton(&active, theme.SkeletonRequest{})
	if skeleton.Metrics.Height != 24 || skeleton.Metrics.Radius != 8 || skeleton.Metrics.Period != 2*time.Second {
		t.Fatalf("skeleton metrics = %+v", skeleton.Metrics)
	}
	theme.ValidateImageVisual(image)
	theme.ValidateCodeVisual(code)
	theme.ValidateSkeletonVisual(skeleton)
}

// TestChartRecipesExposeAnimationDurations pins that chart animation timing is
// provider-owned instead of a widget constant.
func TestChartRecipesExposeAnimationDurations(t *testing.T) {
	active := base.Light()
	provider := base.Provider{}
	bar := provider.BarChart(&active, theme.BarChartRequest{})
	if bar.Metrics.AnimationDuration <= 0 || bar.Metrics.UpdateDuration <= 0 {
		t.Fatalf("bar chart durations = %v/%v", bar.Metrics.AnimationDuration, bar.Metrics.UpdateDuration)
	}
	line := provider.LineChart(&active, theme.LineChartRequest{})
	if line.Metrics.AnimationDuration <= 0 || line.Metrics.UpdateDuration <= 0 {
		t.Fatalf("line chart durations = %v/%v", line.Metrics.AnimationDuration, line.Metrics.UpdateDuration)
	}
	candle := provider.Candlestick(&active, theme.CandlestickRequest{})
	if candle.Metrics.AnimationDuration <= 0 || candle.Metrics.UpdateDuration <= 0 {
		t.Fatalf("candlestick durations = %v/%v", candle.Metrics.AnimationDuration, candle.Metrics.UpdateDuration)
	}
	gantt := provider.Gantt(&active, theme.GanttRequest{})
	if gantt.Metrics.AnimationDuration <= 0 || gantt.Metrics.UpdateDuration <= 0 {
		t.Fatalf("gantt durations = %v/%v", gantt.Metrics.AnimationDuration, gantt.Metrics.UpdateDuration)
	}
}
