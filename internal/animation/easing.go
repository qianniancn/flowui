package animation

import (
	"time"

	animationcore "github.com/qianniancn/FlowUI/internal/animation/core"
)

type Easing = animationcore.Easing

func EaseLinear(progress float32) float32 {
	return animationcore.EaseLinear(progress)
}

func EaseQuadraticIn(progress float32) float32 {
	return animationcore.EaseQuadraticIn(progress)
}

func EaseQuadraticOut(progress float32) float32 {
	return animationcore.EaseQuadraticOut(progress)
}

func EaseQuadraticInOut(progress float32) float32 {
	return animationcore.EaseQuadraticInOut(progress)
}

func EaseCubicIn(progress float32) float32 {
	return animationcore.EaseCubicIn(progress)
}

func EaseCubicOut(progress float32) float32 {
	return animationcore.EaseCubicOut(progress)
}

func EaseCubicInOut(progress float32) float32 {
	return animationcore.EaseCubicInOut(progress)
}

func EaseQuarticIn(progress float32) float32 {
	return animationcore.EaseQuarticIn(progress)
}

func EaseQuarticOut(progress float32) float32 {
	return animationcore.EaseQuarticOut(progress)
}

func EaseQuarticInOut(progress float32) float32 {
	return animationcore.EaseQuarticInOut(progress)
}

func EaseQuinticIn(progress float32) float32 {
	return animationcore.EaseQuinticIn(progress)
}

func EaseQuinticOut(progress float32) float32 {
	return animationcore.EaseQuinticOut(progress)
}

func EaseQuinticInOut(progress float32) float32 {
	return animationcore.EaseQuinticInOut(progress)
}

func EaseSinusoidalIn(progress float32) float32 {
	return animationcore.EaseSinusoidalIn(progress)
}

func EaseSinusoidalOut(progress float32) float32 {
	return animationcore.EaseSinusoidalOut(progress)
}

func EaseSinusoidalInOut(progress float32) float32 {
	return animationcore.EaseSinusoidalInOut(progress)
}

func EaseExponentialIn(progress float32) float32 {
	return animationcore.EaseExponentialIn(progress)
}

func EaseExponentialOut(progress float32) float32 {
	return animationcore.EaseExponentialOut(progress)
}

func EaseExponentialInOut(progress float32) float32 {
	return animationcore.EaseExponentialInOut(progress)
}

func EaseCircularIn(progress float32) float32 {
	return animationcore.EaseCircularIn(progress)
}

func EaseCircularOut(progress float32) float32 {
	return animationcore.EaseCircularOut(progress)
}

func EaseCircularInOut(progress float32) float32 {
	return animationcore.EaseCircularInOut(progress)
}

func EaseElasticIn(progress float32) float32 {
	return animationcore.EaseElasticIn(progress)
}

func EaseElasticOut(progress float32) float32 {
	return animationcore.EaseElasticOut(progress)
}

func EaseElasticInOut(progress float32) float32 {
	return animationcore.EaseElasticInOut(progress)
}

func EaseBackIn(progress float32) float32 {
	return animationcore.EaseBackIn(progress)
}

func EaseBackOut(progress float32) float32 {
	return animationcore.EaseBackOut(progress)
}

func EaseBackInOut(progress float32) float32 {
	return animationcore.EaseBackInOut(progress)
}

func EaseBounceIn(progress float32) float32 {
	return animationcore.EaseBounceIn(progress)
}

func EaseBounceOut(progress float32) float32 {
	return animationcore.EaseBounceOut(progress)
}

func EaseBounceInOut(progress float32) float32 {
	return animationcore.EaseBounceInOut(progress)
}

func EaseSmoothstep(progress float32) float32 {
	return animationcore.EaseSmoothstep(progress)
}

func Progress(elapsed, duration time.Duration) float32 {
	return animationcore.Progress(elapsed, duration)
}

func applyEasing(easing Easing, progress float32) float32 {
	return animationcore.ApplyEasing(easing, progress)
}
