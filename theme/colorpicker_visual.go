package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// ColorPickerRequest contains caller-controlled visual declarations. Selection,
// focus, history, and overlay dismissal remain widget-owned.
type ColorPickerRequest struct {
	Override *ColorPickerOverride
}

type ColorPickerOverride struct {
	Declaration style.Style
}

// ColorPickerOverrideStyle extracts the instance declaration owned by the
// visual provider boundary.
func ColorPickerOverrideStyle(request ColorPickerRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

type ColorPickerMetrics struct {
	TriggerGap              unit.Dp
	PanelGap                unit.Dp
	PanelWidth              unit.Dp
	PanelPadding            unit.Dp
	PanelRadius             unit.Dp
	PanelMaxHeight          unit.Dp
	ContentGap              unit.Dp
	CompactContentGap       unit.Dp
	ChannelGap              unit.Dp
	EnterDuration           time.Duration
	ExitDuration            time.Duration
	EnterScale              float32
	SwatchSelectionDuration time.Duration
	SwatchCheckDuration     time.Duration
	SwatchHoverDuration     time.Duration
}

// ColorPickerHSV is the neutral color sample passed to theme-owned renderers.
// The widget owns conversion and interaction state; themes only consume the
// normalized values to paint the current control.
type ColorPickerHSV struct {
	Hue        float64
	Saturation float64
	Value      float64
	Alpha      float64
}

type ColorWheelRenderRequest struct {
	Size  image.Point
	Value ColorPickerHSV
	Focus float32
}

type ColorWheelRenderer func(layout.Context, ColorWheelRenderRequest)

type ColorAreaRenderRequest struct {
	Size     image.Point
	Value    ColorPickerHSV
	Dragging bool
	Focus    float32
	ShowDots bool
}

type ColorAreaRenderer func(layout.Context, ColorAreaRenderRequest)

type ColorSliderChannel uint8

const (
	ColorSliderHue ColorSliderChannel = iota
	ColorSliderAlpha
)

type ColorSliderRenderRequest struct {
	Size    image.Point
	Channel ColorSliderChannel
	Value   ColorPickerHSV
	Color   color.NRGBA
	Focus   float32
}

type ColorSliderRenderer func(layout.Context, ColorSliderRenderRequest)

type ColorSwatchShape uint8

const (
	ColorSwatchCircle ColorSwatchShape = iota
	ColorSwatchSquare
)

type ColorSwatchRenderRequest struct {
	Size     image.Point
	Value    color.NRGBA
	Shape    ColorSwatchShape
	SizeKind uint8
}

type ColorSwatchRenderer func(layout.Context, ColorSwatchRenderRequest)

type ColorSwatchPickerItemRequest struct {
	Size       image.Point
	Value      color.NRGBA
	Background color.NRGBA
	SwatchSize uint8
	Shape      ColorSwatchShape
	Selection  float32
	Check      float32
	Hover      float32
	Focus      float32
}

type ColorSwatchPickerItemRenderer func(layout.Context, ColorSwatchPickerItemRequest)

type ColorPickerVisual struct {
	Override        style.Style
	Metrics         ColorPickerMetrics
	Wheel           ColorWheelVisual
	Area            ColorAreaVisual
	Slider          ColorSliderVisual
	Field           ColorFieldVisual
	Swatch          ColorSwatchVisual
	SwatchPicker    ColorSwatchPickerVisual
	PanelForeground color.NRGBA
	PanelBackground color.NRGBA
	Foreground      color.NRGBA
	MutedForeground color.NRGBA
	FocusColor      color.NRGBA
	DisabledOpacity float32
	Motion          MotionTheme
	TriggerLabel    func(layout.Context, string) layout.Dimensions
	TriggerFocus    func(layout.Context, image.Point, float32)
	Panel           func(layout.Context, image.Point)
}

type ColorWheelVisual struct {
	Size             unit.Dp
	ThumbSize        unit.Dp
	ThumbBorderWidth unit.Dp
	FocusRingWidth   unit.Dp
	Render           ColorWheelRenderer
}

type ColorAreaVisual struct {
	Size              unit.Dp
	Radius            unit.Dp
	ThumbSize         unit.Dp
	DraggingThumbSize unit.Dp
	ThumbBorderWidth  unit.Dp
	FocusRingWidth    unit.Dp
	DotSize           unit.Dp
	DotGap            unit.Dp
	Render            ColorAreaRenderer
}

type ColorSliderVisual struct {
	TextSize         unit.Sp
	HeaderGap        unit.Dp
	TrackHeight      unit.Dp
	ThumbSize        unit.Dp
	ThumbBorderWidth unit.Dp
	FocusRingWidth   unit.Dp
	Render           ColorSliderRenderer
}

type ColorFieldVisual struct {
	Gap             unit.Dp
	ErrorTextSize   unit.Sp
	ErrorForeground color.NRGBA
}

type ColorSwatchVisual struct {
	ExtraSmallSize unit.Dp
	SmallSize      unit.Dp
	MediumSize     unit.Dp
	LargeSize      unit.Dp
	ExtraLargeSize unit.Dp
	SquareRadius   unit.Dp
	Render         ColorSwatchRenderer
}

type ColorSwatchPickerVisual struct {
	Gap                          unit.Dp
	ExtraSmallBorderWidth        unit.Dp
	BorderWidth                  unit.Dp
	LargeBorderWidth             unit.Dp
	FocusRingWidth               unit.Dp
	FocusRingGap                 unit.Dp
	CheckStroke                  unit.Dp
	SquareItemRadiusExtraSmall   unit.Dp
	SquareItemRadiusSmall        unit.Dp
	SquareItemRadius             unit.Dp
	SquareSwatchRadiusExtraSmall unit.Dp
	SquareSwatchRadiusSmall      unit.Dp
	SquareSwatchRadius           unit.Dp
	SquareSelectedSmallRadius    unit.Dp
	ShadowOpacity                float32
	SelectedScale                float32
	Render                       ColorSwatchPickerItemRenderer
}

type ColorPickerProvider interface {
	ColorPicker(active *Theme, request ColorPickerRequest) ColorPickerVisual
}

// ValidateColorPickerVisual rejects partial visual plans before layout.
func ValidateColorPickerVisual(visual ColorPickerVisual) {
	m := visual.Metrics
	if !validNonNegativeDp(m.TriggerGap) {
		invalidVisual("ColorPicker", "Metrics.TriggerGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PanelGap) {
		invalidVisual("ColorPicker", "Metrics.PanelGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.PanelWidth) {
		invalidVisual("ColorPicker", "Metrics.PanelWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.PanelPadding) {
		invalidVisual("ColorPicker", "Metrics.PanelPadding", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PanelRadius) {
		invalidVisual("ColorPicker", "Metrics.PanelRadius", "must be finite and non-negative")
	}
	if !validPositiveDp(m.PanelMaxHeight) {
		invalidVisual("ColorPicker", "Metrics.PanelMaxHeight", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.ContentGap) {
		invalidVisual("ColorPicker", "Metrics.ContentGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.CompactContentGap) {
		invalidVisual("ColorPicker", "Metrics.CompactContentGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.ChannelGap) {
		invalidVisual("ColorPicker", "Metrics.ChannelGap", "must be finite and non-negative")
	}
	if m.EnterDuration <= 0 {
		invalidVisual("ColorPicker", "Metrics.EnterDuration", "must be greater than 0")
	}
	if m.ExitDuration <= 0 {
		invalidVisual("ColorPicker", "Metrics.ExitDuration", "must be greater than 0")
	}
	if m.EnterScale <= 0 {
		invalidVisual("ColorPicker", "Metrics.EnterScale", "must be greater than 0")
	}
	if m.EnterScale > 1 {
		invalidVisual("ColorPicker", "Metrics.EnterScale", "must be at most 1")
	}
	if m.SwatchSelectionDuration <= 0 {
		invalidVisual("ColorPicker", "Metrics.SwatchSelectionDuration", "must be greater than 0")
	}
	if m.SwatchCheckDuration <= 0 {
		invalidVisual("ColorPicker", "Metrics.SwatchCheckDuration", "must be greater than 0")
	}
	if m.SwatchHoverDuration <= 0 {
		invalidVisual("ColorPicker", "Metrics.SwatchHoverDuration", "must be greater than 0")
	}
	if !hasValidOpacityValue(visual.DisabledOpacity) {
		invalidVisual("ColorPicker", "DisabledOpacity", "must be finite and between 0 and 1")
	}
	if visual.TriggerLabel == nil {
		invalidVisual("ColorPicker", "TriggerLabel", "must be provided")
	}
	if visual.TriggerFocus == nil {
		invalidVisual("ColorPicker", "TriggerFocus", "must be provided")
	}
	if visual.Panel == nil {
		invalidVisual("ColorPicker", "Panel", "must be provided")
	}

	for _, value := range []ColorWheelVisual{visual.Wheel} {
		if !validPositiveDp(value.Size) {
			invalidVisual("ColorPicker", "Wheel.Size", "must be finite and greater than 0")
		}
		if !validPositiveDp(value.ThumbSize) {
			invalidVisual("ColorPicker", "Wheel.ThumbSize", "must be finite and greater than 0")
		}
		if !validPositiveDp(value.ThumbBorderWidth) {
			invalidVisual("ColorPicker", "Wheel.ThumbBorderWidth", "must be finite and greater than 0")
		}
		if !validPositiveDp(value.FocusRingWidth) {
			invalidVisual("ColorPicker", "Wheel.FocusRingWidth", "must be finite and greater than 0")
		}
		if value.Render == nil {
			invalidVisual("ColorPicker", "Wheel.Render", "must be provided")
		}
	}
	if !validPositiveDp(visual.Area.Size) {
		invalidVisual("ColorPicker", "Area.Size", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(visual.Area.Radius) {
		invalidVisual("ColorPicker", "Area.Radius", "must be finite and non-negative")
	}
	if !validPositiveDp(visual.Area.ThumbSize) {
		invalidVisual("ColorPicker", "Area.ThumbSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.Area.DraggingThumbSize) {
		invalidVisual("ColorPicker", "Area.DraggingThumbSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.Area.ThumbBorderWidth) {
		invalidVisual("ColorPicker", "Area.ThumbBorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.Area.FocusRingWidth) {
		invalidVisual("ColorPicker", "Area.FocusRingWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.Area.DotSize) {
		invalidVisual("ColorPicker", "Area.DotSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.Area.DotGap) {
		invalidVisual("ColorPicker", "Area.DotGap", "must be finite and greater than 0")
	}
	if visual.Area.Render == nil {
		invalidVisual("ColorPicker", "Area.Render", "must be provided")
	}
	if !validPositiveSp(visual.Slider.TextSize) {
		invalidVisual("ColorPicker", "Slider.TextSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(visual.Slider.HeaderGap) {
		invalidVisual("ColorPicker", "Slider.HeaderGap", "must be finite and non-negative")
	}
	if !validPositiveDp(visual.Slider.TrackHeight) {
		invalidVisual("ColorPicker", "Slider.TrackHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.Slider.ThumbSize) {
		invalidVisual("ColorPicker", "Slider.ThumbSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.Slider.ThumbBorderWidth) {
		invalidVisual("ColorPicker", "Slider.ThumbBorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.Slider.FocusRingWidth) {
		invalidVisual("ColorPicker", "Slider.FocusRingWidth", "must be finite and greater than 0")
	}
	if visual.Slider.Render == nil {
		invalidVisual("ColorPicker", "Slider.Render", "must be provided")
	}
	if !validNonNegativeDp(visual.Field.Gap) {
		invalidVisual("ColorPicker", "Field.Gap", "must be finite and non-negative")
	}
	if !validPositiveSp(visual.Field.ErrorTextSize) {
		invalidVisual("ColorPicker", "Field.ErrorTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.Swatch.MediumSize) {
		invalidVisual("ColorPicker", "Swatch.MediumSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(visual.Swatch.SquareRadius) {
		invalidVisual("ColorPicker", "Swatch.SquareRadius", "must be finite and non-negative")
	}
	if visual.Swatch.Render == nil {
		invalidVisual("ColorPicker", "Swatch.Render", "must be provided")
	}
	if !validNonNegativeDp(visual.SwatchPicker.Gap) {
		invalidVisual("ColorPicker", "SwatchPicker.Gap", "must be finite and non-negative")
	}
	if !validPositiveDp(visual.SwatchPicker.BorderWidth) {
		invalidVisual("ColorPicker", "SwatchPicker.BorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.SwatchPicker.FocusRingWidth) {
		invalidVisual("ColorPicker", "SwatchPicker.FocusRingWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.SwatchPicker.CheckStroke) {
		invalidVisual("ColorPicker", "SwatchPicker.CheckStroke", "must be finite and greater than 0")
	}
	if visual.SwatchPicker.Render == nil {
		invalidVisual("ColorPicker", "SwatchPicker.Render", "must be provided")
	}
}
