package theme

import (
	"math"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

func hasPositiveHeightOrAspect(box *style.BoxStyle) bool {
	if hasPositiveHeight(box.Height, box.MinHeight) {
		return true
	}
	if box.AspectRatio == nil {
		return false
	}
	value := float64(*box.AspectRatio)
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func validOptionalPadding(padding *style.Insets) bool {
	if padding == nil {
		return true
	}
	return validNonNegativeDp(padding.Top) && validNonNegativeDp(padding.Right) &&
		validNonNegativeDp(padding.Bottom) && validNonNegativeDp(padding.Left)
}

func hasPositiveHeight(height, minHeight *unit.Dp) bool {
	return (height != nil && validPositiveDp(*height)) || (minHeight != nil && validPositiveDp(*minHeight))
}

func hasPositiveWidth(width, minWidth *unit.Dp) bool {
	return (width != nil && validPositiveDp(*width)) || (minWidth != nil && validPositiveDp(*minWidth))
}

func validPositiveDp(value unit.Dp) bool {
	return validFiniteDp(value) && value > 0
}

func validPositiveSp(value unit.Sp) bool {
	converted := float64(value)
	return !math.IsNaN(converted) && !math.IsInf(converted, 0) && value > 0
}

func validNonNegativeDp(value unit.Dp) bool {
	return validFiniteDp(value) && value >= 0
}

func validFiniteDp(value unit.Dp) bool {
	converted := float64(value)
	return !math.IsNaN(converted) && !math.IsInf(converted, 0)
}

func validFiniteFloat32(value float32) bool {
	converted := float64(value)
	return !math.IsNaN(converted) && !math.IsInf(converted, 0)
}

func hasValidScale(transform *style.TransformStyle) bool {
	if transform == nil || transform.ScaleX == nil || transform.ScaleY == nil {
		return false
	}
	return *transform.ScaleX > 0 && *transform.ScaleY > 0 &&
		!math.IsNaN(float64(*transform.ScaleX)) && !math.IsNaN(float64(*transform.ScaleY)) &&
		!math.IsInf(float64(*transform.ScaleX), 0) && !math.IsInf(float64(*transform.ScaleY), 0)
}

func hasValidOpacity(value *float32) bool {
	return value != nil && !math.IsNaN(float64(*value)) && !math.IsInf(float64(*value), 0) &&
		*value >= 0 && *value <= 1
}

func hasValidOpacityValue(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0) && value >= 0 && value <= 1
}

func hasOptionalScale(transform *style.TransformStyle) bool {
	if transform == nil {
		return true
	}
	if transform.ScaleX == nil && transform.ScaleY == nil {
		return true
	}
	return hasValidScale(transform)
}

func validateMotion(component, field string, motion MotionTheme) {
	if motion.DefaultDuration <= 0 {
		invalidVisual(component, field+".DefaultDuration", "must be greater than 0")
	}
	if motion.ControlDuration < 0 {
		invalidVisual(component, field+".ControlDuration", "must be at least 0")
	}
	if motion.DurationScale < 0 {
		invalidVisual(component, field+".DurationScale", "must be at least 0")
	}
	if math.IsNaN(float64(motion.DurationScale)) {
		invalidVisual(component, field+".DurationScale", "must be finite")
	}
	if math.IsInf(float64(motion.DurationScale), 0) {
		invalidVisual(component, field+".DurationScale", "must be finite")
	}
	if motion.Ease > EaseBack {
		invalidVisual(component, field+".Ease", "must be at most EaseBack")
	}
}

func hasColor(value style.ColorSource) bool {
	switch value := value.(type) {
	case style.SolidColor:
		return true
	case *style.SolidColor:
		return value != nil
	default:
		return false
	}
}

func hasPaint(value style.PaintSource) bool {
	switch value := value.(type) {
	case style.SolidColor:
		return true
	case *style.SolidColor:
		return value != nil
	case style.Gradient:
		return len(value.Stops) > 0
	case *style.Gradient:
		return value != nil && len(value.Stops) > 0
	default:
		return false
	}
}
