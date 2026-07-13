package core

import (
	"math"
	"time"
)

// Easing maps normalized time to animation progress. Custom functions may
// overshoot, but must return a finite value.
type Easing func(float32) float32

func EaseLinear(progress float32) float32 {
	return clampProgress(progress)
}

func EaseQuadraticIn(progress float32) float32 {
	progress = clampProgress(progress)
	return progress * progress
}

func EaseQuadraticOut(progress float32) float32 {
	progress = clampProgress(progress)
	return progress * (2 - progress)
}

func EaseQuadraticInOut(progress float32) float32 {
	progress = clampProgress(progress) * 2
	if progress < 1 {
		return 0.5 * progress * progress
	}
	progress--
	return -0.5 * (progress*(progress-2) - 1)
}

func EaseCubicIn(progress float32) float32 {
	progress = clampProgress(progress)
	return progress * progress * progress
}

func EaseCubicOut(progress float32) float32 {
	progress = clampProgress(progress) - 1
	return progress*progress*progress + 1
}

func EaseCubicInOut(progress float32) float32 {
	progress = clampProgress(progress) * 2
	if progress < 1 {
		return 0.5 * progress * progress * progress
	}
	progress -= 2
	return 0.5 * (progress*progress*progress + 2)
}

func EaseQuarticIn(progress float32) float32 {
	progress = clampProgress(progress)
	return progress * progress * progress * progress
}

func EaseQuarticOut(progress float32) float32 {
	progress = clampProgress(progress) - 1
	return 1 - progress*progress*progress*progress
}

func EaseQuarticInOut(progress float32) float32 {
	progress = clampProgress(progress) * 2
	if progress < 1 {
		return 0.5 * progress * progress * progress * progress
	}
	progress -= 2
	return -0.5 * (progress*progress*progress*progress - 2)
}

func EaseQuinticIn(progress float32) float32 {
	progress = clampProgress(progress)
	return progress * progress * progress * progress * progress
}

func EaseQuinticOut(progress float32) float32 {
	progress = clampProgress(progress) - 1
	return progress*progress*progress*progress*progress + 1
}

func EaseQuinticInOut(progress float32) float32 {
	progress = clampProgress(progress) * 2
	if progress < 1 {
		return 0.5 * progress * progress * progress * progress * progress
	}
	progress -= 2
	return 0.5 * (progress*progress*progress*progress*progress + 2)
}

func EaseSinusoidalIn(progress float32) float32 {
	progress = clampProgress(progress)
	return float32(1 - math.Cos(float64(progress)*math.Pi/2))
}

func EaseSinusoidalOut(progress float32) float32 {
	progress = clampProgress(progress)
	return float32(math.Sin(float64(progress) * math.Pi / 2))
}

func EaseSinusoidalInOut(progress float32) float32 {
	progress = clampProgress(progress)
	return float32(0.5 * (1 - math.Cos(math.Pi*float64(progress))))
}

func EaseExponentialIn(progress float32) float32 {
	progress = clampProgress(progress)
	if progress == 0 {
		return 0
	}
	return float32(math.Pow(1024, float64(progress-1)))
}

func EaseExponentialOut(progress float32) float32 {
	progress = clampProgress(progress)
	if progress == 1 {
		return 1
	}
	return float32(1 - math.Pow(2, float64(-10*progress)))
}

func EaseExponentialInOut(progress float32) float32 {
	progress = clampProgress(progress)
	if progress == 0 || progress == 1 {
		return progress
	}
	progress *= 2
	if progress < 1 {
		return float32(0.5 * math.Pow(1024, float64(progress-1)))
	}
	return float32(0.5 * (-math.Pow(2, float64(-10*(progress-1))) + 2))
}

func EaseCircularIn(progress float32) float32 {
	progress = clampProgress(progress)
	return float32(1 - math.Sqrt(float64(1-progress*progress)))
}

func EaseCircularOut(progress float32) float32 {
	progress = clampProgress(progress) - 1
	return float32(math.Sqrt(float64(1 - progress*progress)))
}

func EaseCircularInOut(progress float32) float32 {
	progress = clampProgress(progress) * 2
	if progress < 1 {
		return float32(-0.5 * (math.Sqrt(float64(1-progress*progress)) - 1))
	}
	progress -= 2
	return float32(0.5 * (math.Sqrt(float64(1-progress*progress)) + 1))
}

func EaseElasticIn(progress float32) float32 {
	progress = clampProgress(progress)
	if progress == 0 || progress == 1 {
		return progress
	}
	const period = 0.4
	const shift = period / 4
	progress--
	return float32(-math.Pow(2, float64(10*progress)) *
		math.Sin(float64(progress-shift)*(2*math.Pi)/period))
}

func EaseElasticOut(progress float32) float32 {
	progress = clampProgress(progress)
	if progress == 0 || progress == 1 {
		return progress
	}
	const period = 0.4
	const shift = period / 4
	return float32(math.Pow(2, float64(-10*progress))*
		math.Sin(float64(progress-shift)*(2*math.Pi)/period) + 1)
}

func EaseElasticInOut(progress float32) float32 {
	progress = clampProgress(progress)
	if progress == 0 || progress == 1 {
		return progress
	}
	const period = 0.4
	const shift = period / 4
	progress *= 2
	if progress < 1 {
		progress--
		return float32(-0.5 * math.Pow(2, float64(10*progress)) *
			math.Sin(float64(progress-shift)*(2*math.Pi)/period))
	}
	progress--
	return float32(math.Pow(2, float64(-10*progress))*
		math.Sin(float64(progress-shift)*(2*math.Pi)/period)*0.5 + 1)
}

func EaseBackIn(progress float32) float32 {
	const overshoot = float32(1.70158)
	progress = clampProgress(progress)
	return progress * progress * ((overshoot+1)*progress - overshoot)
}

func EaseBackOut(progress float32) float32 {
	const overshoot = float32(1.70158)
	progress = clampProgress(progress) - 1
	return 1 + (overshoot+1)*progress*progress*progress + overshoot*progress*progress
}

func EaseBackInOut(progress float32) float32 {
	const overshoot = float32(1.70158 * 1.525)
	progress = clampProgress(progress) * 2
	if progress < 1 {
		return 0.5 * progress * progress * ((overshoot+1)*progress - overshoot)
	}
	progress -= 2
	return 0.5 * (progress*progress*((overshoot+1)*progress+overshoot) + 2)
}

func EaseBounceIn(progress float32) float32 {
	progress = clampProgress(progress)
	return 1 - easeBounceOut(1-progress)
}

func EaseBounceOut(progress float32) float32 {
	return easeBounceOut(clampProgress(progress))
}

func EaseBounceInOut(progress float32) float32 {
	progress = clampProgress(progress)
	if progress < 0.5 {
		return (1 - easeBounceOut(1-progress*2)) * 0.5
	}
	return easeBounceOut(progress*2-1)*0.5 + 0.5
}

func easeBounceOut(progress float32) float32 {
	const scale = float32(7.5625)
	if progress < 1/float32(2.75) {
		return scale * progress * progress
	}
	if progress < 2/float32(2.75) {
		progress -= 1.5 / 2.75
		return scale*progress*progress + 0.75
	}
	if progress < 2.5/float32(2.75) {
		progress -= 2.25 / 2.75
		return scale*progress*progress + 0.9375
	}
	progress -= 2.625 / 2.75
	return scale*progress*progress + 0.984375
}

func EaseSmoothstep(progress float32) float32 {
	progress = clampProgress(progress)
	return progress * progress * (3 - 2*progress)
}

func ApplyEasing(easing Easing, progress float32) float32 {
	if easing == nil {
		easing = EaseCubicOut
	}
	value := easing(clampProgress(progress))
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		panic("flowui: easing returned a non-finite value")
	}
	return value
}

func Progress(elapsed, duration time.Duration) float32 {
	if elapsed <= 0 {
		return 0
	}
	if duration <= 0 || elapsed >= duration {
		return 1
	}
	return float32(elapsed) / float32(duration)
}

func clampProgress(progress float32) float32 {
	if math.IsNaN(float64(progress)) {
		return 0
	}
	return min(max(progress, 0), 1)
}
