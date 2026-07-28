package ui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"github.com/qianniancn/flowui/internal/animation"
)

// Easing maps normalized time to animation progress.
type Easing = animation.Easing

// TweenValue is a value-semantic keyed float transition.
type TweenValue = animation.TweenValue

// EaseLinear advances at a constant speed.
func EaseLinear(progress float32) float32 {
	return animation.EaseLinear(progress)
}

// EaseQuadraticIn accelerates with a quadratic curve.
func EaseQuadraticIn(progress float32) float32 {
	return animation.EaseQuadraticIn(progress)
}

// EaseQuadraticOut decelerates with a quadratic curve.
func EaseQuadraticOut(progress float32) float32 {
	return animation.EaseQuadraticOut(progress)
}

// EaseQuadraticInOut accelerates and decelerates with a quadratic curve.
func EaseQuadraticInOut(progress float32) float32 {
	return animation.EaseQuadraticInOut(progress)
}

// EaseCubicIn accelerates with a cubic curve.
func EaseCubicIn(progress float32) float32 {
	return animation.EaseCubicIn(progress)
}

// EaseCubicOut starts quickly and settles smoothly.
func EaseCubicOut(progress float32) float32 {
	return animation.EaseCubicOut(progress)
}

// EaseCubicInOut accelerates and decelerates symmetrically.
func EaseCubicInOut(progress float32) float32 {
	return animation.EaseCubicInOut(progress)
}

// EaseQuarticIn accelerates with a quartic curve.
func EaseQuarticIn(progress float32) float32 {
	return animation.EaseQuarticIn(progress)
}

// EaseQuarticOut decelerates with a quartic curve.
func EaseQuarticOut(progress float32) float32 {
	return animation.EaseQuarticOut(progress)
}

// EaseQuarticInOut accelerates and decelerates with a quartic curve.
func EaseQuarticInOut(progress float32) float32 {
	return animation.EaseQuarticInOut(progress)
}

// EaseQuinticIn accelerates with a quintic curve.
func EaseQuinticIn(progress float32) float32 {
	return animation.EaseQuinticIn(progress)
}

// EaseQuinticOut decelerates with a quintic curve.
func EaseQuinticOut(progress float32) float32 {
	return animation.EaseQuinticOut(progress)
}

// EaseQuinticInOut accelerates and decelerates with a quintic curve.
func EaseQuinticInOut(progress float32) float32 {
	return animation.EaseQuinticInOut(progress)
}

// EaseSinusoidalIn accelerates with a sine curve.
func EaseSinusoidalIn(progress float32) float32 {
	return animation.EaseSinusoidalIn(progress)
}

// EaseSinusoidalOut decelerates with a sine curve.
func EaseSinusoidalOut(progress float32) float32 {
	return animation.EaseSinusoidalOut(progress)
}

// EaseSinusoidalInOut accelerates and decelerates with a sine curve.
func EaseSinusoidalInOut(progress float32) float32 {
	return animation.EaseSinusoidalInOut(progress)
}

// EaseExponentialIn accelerates exponentially.
func EaseExponentialIn(progress float32) float32 {
	return animation.EaseExponentialIn(progress)
}

// EaseExponentialOut decelerates exponentially.
func EaseExponentialOut(progress float32) float32 {
	return animation.EaseExponentialOut(progress)
}

// EaseExponentialInOut accelerates and decelerates exponentially.
func EaseExponentialInOut(progress float32) float32 {
	return animation.EaseExponentialInOut(progress)
}

// EaseCircularIn accelerates along a circular curve.
func EaseCircularIn(progress float32) float32 {
	return animation.EaseCircularIn(progress)
}

// EaseCircularOut decelerates along a circular curve.
func EaseCircularOut(progress float32) float32 {
	return animation.EaseCircularOut(progress)
}

// EaseCircularInOut accelerates and decelerates along a circular curve.
func EaseCircularInOut(progress float32) float32 {
	return animation.EaseCircularInOut(progress)
}

// EaseElasticIn winds back and oscillates while accelerating.
func EaseElasticIn(progress float32) float32 {
	return animation.EaseElasticIn(progress)
}

// EaseElasticOut overshoots and oscillates while settling.
func EaseElasticOut(progress float32) float32 {
	return animation.EaseElasticOut(progress)
}

// EaseElasticInOut combines elastic acceleration and deceleration.
func EaseElasticInOut(progress float32) float32 {
	return animation.EaseElasticInOut(progress)
}

// EaseBackIn briefly moves backward before accelerating.
func EaseBackIn(progress float32) float32 {
	return animation.EaseBackIn(progress)
}

// EaseBackOut briefly overshoots before settling.
func EaseBackOut(progress float32) float32 {
	return animation.EaseBackOut(progress)
}

// EaseBackInOut combines backward anticipation and overshoot.
func EaseBackInOut(progress float32) float32 {
	return animation.EaseBackInOut(progress)
}

// EaseBounceIn accelerates with a bouncing motion.
func EaseBounceIn(progress float32) float32 {
	return animation.EaseBounceIn(progress)
}

// EaseBounceOut settles with a bouncing motion.
func EaseBounceOut(progress float32) float32 {
	return animation.EaseBounceOut(progress)
}

// EaseBounceInOut combines bouncing acceleration and deceleration.
func EaseBounceInOut(progress float32) float32 {
	return animation.EaseBounceInOut(progress)
}

// LerpFloat interpolates float32 values and permits easing overshoot.
func LerpFloat(from, to, progress float32) float32 {
	return animation.LerpFloat(from, to, progress)
}

// LerpFloat64 interpolates float64 values and permits easing overshoot.
func LerpFloat64(from, to float64, progress float32) float64 {
	return animation.LerpFloat64(from, to, progress)
}

// LerpColor interpolates non-premultiplied colors and clamps channels.
func LerpColor(from, to color.NRGBA, progress float32) color.NRGBA {
	return animation.LerpColor(from, to, progress)
}

// LerpPoint interpolates Gio points.
func LerpPoint(from, to f32.Point, progress float32) f32.Point {
	return animation.LerpPoint(from, to, progress)
}

// LerpRect interpolates integer rectangle edges.
func LerpRect(from, to image.Rectangle, progress float32) image.Rectangle {
	return animation.LerpRect(from, to, progress)
}

// Tween creates a keyed frame-persistent float transition.
func Tween(key string, target float32) TweenValue {
	return animation.Tween(key, target)
}

// SpringConfig configures a damped spring animation.
type SpringConfig = animation.SpringConfig

// DefaultSpring returns a balanced desktop spring.
func DefaultSpring() SpringConfig {
	return animation.DefaultSpring()
}

// SpringSnappy is a quick, low-overshoot spring for small UI motions.
func SpringSnappy() SpringConfig {
	return animation.SpringSnappy()
}

// SpringGentle is a soft spring for larger layout moves.
func SpringGentle() SpringConfig {
	return animation.SpringGentle()
}

// SpringBouncy overshoots noticeably before settling.
func SpringBouncy() SpringConfig {
	return animation.SpringBouncy()
}

// TimelineValue is a keyed multi-keyframe animation.
type TimelineValue = animation.TimelineValue

// Timeline creates a keyed multi-keyframe animation.
func Timeline(key string) TimelineValue {
	return animation.Timeline(key)
}

// AnimateLayoutWidget animates layout dimension changes.
type AnimateLayoutWidget = animation.AnimateLayoutWidget

// AnimateLayout wraps child and animates reported size changes.
func AnimateLayout(key string, child Widget) AnimateLayoutWidget {
	return animation.AnimateLayout(key, child)
}

// RectValue animates rectangle edges.
type RectValue = animation.RectValue

// AnimateRect starts a keyed rectangle animation toward target.
func AnimateRect(key string, target image.Rectangle) RectValue {
	return animation.AnimateRect(key, target)
}
