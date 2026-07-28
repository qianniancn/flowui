package animation

import (
	"image"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/theme"
)

func BenchmarkTweenValue(b *testing.B) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	var ops op.Ops
	startTime := time.Now()

	b.Run("Initial", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ops.Reset()
			frame.BeginFrame(ctx)
			gtx := layout.Context{
				Ops:         &ops,
				Now:         startTime,
				Constraints: layout.Constraints{Max: image.Pt(300, 200)},
			}
			tween := Tween("bench-tween", 1.0).Duration(200 * time.Millisecond)
			_ = tween.Value(ctx, gtx)
			frame.EndFrame(ctx)
		}
	})

	b.Run("Animating", func(b *testing.B) {
		// Prime the tween
		frame.BeginFrame(ctx)
		gtx := layout.Context{
			Ops:         &ops,
			Now:         startTime,
			Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		}
		tween := Tween("bench-animate", 0.0).Duration(200 * time.Millisecond)
		_ = tween.Value(ctx, gtx)
		frame.EndFrame(ctx)

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ops.Reset()
			frame.BeginFrame(ctx)
			gtx := layout.Context{
				Ops:         &ops,
				Now:         startTime.Add(time.Duration(i%200) * time.Millisecond),
				Constraints: layout.Constraints{Max: image.Pt(300, 200)},
			}
			tween := Tween("bench-animate", 1.0).Duration(200 * time.Millisecond)
			_ = tween.Value(ctx, gtx)
			frame.EndFrame(ctx)
		}
	})

	b.Run("Complete", func(b *testing.B) {
		// Complete the tween
		frame.BeginFrame(ctx)
		gtx := layout.Context{
			Ops:         &ops,
			Now:         startTime.Add(300 * time.Millisecond),
			Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		}
		tween := Tween("bench-complete", 1.0).Duration(200 * time.Millisecond)
		_ = tween.Value(ctx, gtx)
		frame.EndFrame(ctx)

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ops.Reset()
			frame.BeginFrame(ctx)
			gtx := layout.Context{
				Ops:         &ops,
				Now:         startTime.Add(time.Duration(300+i) * time.Millisecond),
				Constraints: layout.Constraints{Max: image.Pt(300, 200)},
			}
			tween := Tween("bench-complete", 1.0).Duration(200 * time.Millisecond)
			_ = tween.Value(ctx, gtx)
			frame.EndFrame(ctx)
		}
	})
}

func BenchmarkTweenSpring(b *testing.B) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	var ops op.Ops
	startTime := time.Now()

	config := SpringConfig{Stiffness: 180, Damping: 18}

	b.Run("SpringAnimating", func(b *testing.B) {
		// Prime the spring
		frame.BeginFrame(ctx)
		gtx := layout.Context{
			Ops:         &ops,
			Now:         startTime,
			Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		}
		tween := Tween("bench-spring", 0.0).Spring(config).Initial(1.0)
		_ = tween.Value(ctx, gtx)
		frame.EndFrame(ctx)

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ops.Reset()
			frame.BeginFrame(ctx)
			gtx := layout.Context{
				Ops:         &ops,
				Now:         startTime.Add(time.Duration(i) * 16 * time.Millisecond),
				Constraints: layout.Constraints{Max: image.Pt(300, 200)},
			}
			tween := Tween("bench-spring", 0.0).Spring(config)
			_ = tween.Value(ctx, gtx)
			frame.EndFrame(ctx)
		}
	})
}

func BenchmarkEasingFunctions(b *testing.B) {
	progress := float32(0.5)

	b.Run("EaseLinear", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = EaseLinear(progress)
		}
	})

	b.Run("EaseCubicOut", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = EaseCubicOut(progress)
		}
	})

	b.Run("EaseCubicInOut", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = EaseCubicInOut(progress)
		}
	})
}

func BenchmarkLerpFloat(b *testing.B) {
	from, to := float32(0.0), float32(1.0)
	progress := float32(0.5)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = LerpFloat(from, to, progress)
	}
}

func BenchmarkProgress(b *testing.B) {
	elapsed := 100 * time.Millisecond
	duration := 200 * time.Millisecond

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Progress(elapsed, duration)
	}
}
