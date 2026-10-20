package theme

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// ButtonRequest describes the behavior-owned inputs a visual provider needs
// to build a button appearance. It contains no callbacks or Gio state.
type ButtonRequest struct {
	Variant  ButtonVariant
	Size     ButtonSize
	IconOnly bool
	Grouped  bool
	// GroupCorners carries the resolved group position. The provider applies
	// the corner declaration while building Size so widget code never has to
	// splice a second visual layer into a cached declaration.
	GroupCorners ButtonGroupRequest
	// Override is an optional instance-level declaration supplied by the
	// caller. The provider owns how it is merged into its visual layers.
	Override *ButtonOverride
}

// ButtonOverride carries caller customization across the neutral provider
// boundary. It is deliberately a provider input, not widget-owned styling.
type ButtonOverride struct {
	Declaration style.Style
}

// ButtonGroupRequest describes the corner treatment needed by a grouped
// button. The widget owns group position; the provider owns the resulting
// visual declaration.
type ButtonGroupRequest struct {
	NorthWest bool
	NorthEast bool
	SouthEast bool
	SouthWest bool
}

// ButtonGroupVisual contains the theme-owned separator renderer used by a
// button group. The group behavior still owns ordering and separator
// placement; geometry and painting stay inside the renderer closure.
type ButtonGroupVisual struct {
	SeparatorRenderer ButtonGroupSeparatorRenderer
}

// ButtonGroupAxis identifies the neutral axis used to place a group
// separator. The widget owns item order and offsets; the provider owns the
// separator geometry, color treatment, and painting strategy.
type ButtonGroupAxis uint8

const (
	ButtonGroupHorizontal ButtonGroupAxis = iota
	ButtonGroupVertical
)

// ButtonGroupSeparatorRequest contains the layout facts needed by a theme to
// paint one separator between two already-laid-out group items.
type ButtonGroupSeparatorRequest struct {
	Size       image.Point
	MainOffset int
	Axis       ButtonGroupAxis
	Foreground color.NRGBA
	Opacity    float32
}

// ButtonGroupSeparatorRenderer is a theme-owned separator drawing strategy.
type ButtonGroupSeparatorRenderer func(layout.Context, ButtonGroupSeparatorRequest)

// ButtonSpinnerRequest contains the provider-owned geometry and motion values
// needed to paint a loading indicator. The widget decides when the indicator
// is shown; the provider decides how it is drawn.
type ButtonSpinnerRequest struct {
	Size        unit.Dp
	StrokeWidth unit.Dp
	Color       color.NRGBA
	Period      time.Duration
}

// ButtonSpinnerRenderer is a theme-owned loading-indicator strategy.
type ButtonSpinnerRenderer func(layout.Context, ButtonSpinnerRequest) layout.Dimensions

// ButtonSpinnerSizeResolver resolves the provider-owned loading indicator
// diameter captured for the visual's semantic button tier.
type ButtonSpinnerSizeResolver func() unit.Dp

// ButtonSpinnerStrokeWidthResolver resolves provider-owned loading indicator
// stroke geometry captured for the visual's semantic button tier.
type ButtonSpinnerStrokeWidthResolver func() unit.Dp

// ButtonSpinnerPeriodResolver resolves the provider-owned base loading
// indicator motion period captured for the visual's semantic button tier. The
// widget applies the shared active motion policy before passing the period to
// the renderer.
type ButtonSpinnerPeriodResolver func() time.Duration

// ButtonContentGapRequest identifies the button content context whose spacing
// is needed. Providers may vary the gap by icon-only or loading mode.
type ButtonContentGapRequest struct {
	IconOnly bool
	Loading  bool
}

// ButtonContentGapResolver resolves provider-owned spacing between button
// content parts.
type ButtonContentGapResolver func(ButtonContentGapRequest) unit.Dp

// ButtonLoadingRequest contains the resolved button inset needed to preserve
// intrinsic width while a loading indicator replaces the button content. The
// provider owns the semantic size through the visual closure that receives
// this request, so the widget does not echo ButtonRequest.Size back across
// the protocol.
type ButtonLoadingRequest struct {
	Inset layout.Inset
}

// ButtonLoadingInsetResolver derives the provider-owned inset used while a
// button is loading. Widgets apply the returned inset to their resolved root
// declaration but do not interpret spinner dimensions or content gaps.
type ButtonLoadingInsetResolver func(layout.Context, ButtonLoadingRequest) layout.Inset

// ButtonRipple is the neutral press sample retained by button behavior.
// Position is included for themes that anchor a wave to the press point.
type ButtonRipple struct {
	StartedAt time.Time
	X, Y      float32
}

// ButtonRippleRequest contains the resolved button surface and active press
// samples. The widget owns lifecycle and invalidation decisions; the provider
// owns wave geometry, alpha, and painting.
type ButtonRippleRequest struct {
	Ripples        []ButtonRipple
	SpreadDuration time.Duration
	FadeDuration   time.Duration
	Size           image.Point
	Radius         unit.Dp
	Color          color.NRGBA
	ColorValid     bool
}

// ButtonRippleRenderer is a theme-owned press-feedback strategy.
type ButtonRippleRenderer func(layout.Context, ButtonRippleRequest)

// ButtonRippleColorResolver derives the provider-owned press color from the
// resolved root declaration. Widgets do not inspect background/border color
// semantics to decide whether a ripple is visible.
type ButtonRippleColorResolver func(style.ResolvedStyle) (color.NRGBA, bool)

// ButtonForegroundColorResolver derives the provider-owned foreground color
// for a resolved semantic part. Widgets pass the resolved part to this
// resolver when a theme-owned renderer needs a concrete Gio color; the widget
// does not inspect style property representations itself.
type ButtonForegroundColorResolver func(style.ResolvedStyle) (color.NRGBA, bool)

// ButtonVisual is the complete visual input for the shared button behavior.
// The declaration layers are resolved by the component runtime; loading
// geometry and motion are requested through provider-owned resolvers because
// they are theme-owned as well.
type ButtonVisual struct {
	Defaults style.Style
	Variant  style.Style
	Size     style.Style
	// Override is the provider-owned instance layer. It is resolved after
	// inherited scope, variant, and size declarations.
	Override style.Style

	Ripple             bool
	RippleDuration     time.Duration
	RippleFadeDuration time.Duration
	Spinner            ButtonSpinnerRenderer
	SpinnerSize        ButtonSpinnerSizeResolver
	SpinnerStrokeWidth ButtonSpinnerStrokeWidthResolver
	SpinnerPeriod      ButtonSpinnerPeriodResolver
	ContentGap         ButtonContentGapResolver
	LoadingInset       ButtonLoadingInsetResolver
	ForegroundColor    ButtonForegroundColorResolver
	RippleRenderer     ButtonRippleRenderer
	RippleColor        ButtonRippleColorResolver
	Motion             MotionTheme
	Group              ButtonGroupVisual
}

// ValidateButtonVisual checks the complete renderer and metric contract
// before a widget stores a provider result in its declaration cache. Button
// behavior decides when loading, ripple, and grouping occur; it must never
// replace a missing provider strategy with a local visual fallback.
func ValidateButtonVisual(visual ButtonVisual) {
	if visual.ContentGap == nil {
		invalidVisual("Button", "ContentGap", "must be provided")
	}
	if visual.SpinnerSize == nil {
		invalidVisual("Button", "SpinnerSize", "must be provided")
	}
	if visual.SpinnerStrokeWidth == nil {
		invalidVisual("Button", "SpinnerStrokeWidth", "must be provided")
	}
	if visual.SpinnerPeriod == nil {
		invalidVisual("Button", "SpinnerPeriod", "must be provided")
	}
	if visual.Spinner == nil {
		invalidVisual("Button", "Spinner", "must be provided")
	}
	if visual.LoadingInset == nil {
		invalidVisual("Button", "LoadingInset", "must be provided")
	}
	if visual.ForegroundColor == nil {
		invalidVisual("Button", "ForegroundColor", "must be provided")
	}
	if visual.Group.SeparatorRenderer == nil {
		invalidVisual("Button", "Group.SeparatorRenderer", "must be provided")
	}

	validateMotion("Button", "Motion", visual.Motion)
	if !validPositiveDp(visual.SpinnerSize()) {
		invalidVisual("Button", "SpinnerSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(visual.SpinnerStrokeWidth()) {
		invalidVisual("Button", "SpinnerStrokeWidth", "must be finite and greater than 0")
	}
	if visual.SpinnerPeriod() <= 0 {
		invalidVisual("Button", "SpinnerPeriod", "must be greater than 0")
	}

	for _, request := range []ButtonContentGapRequest{
		{},
		{IconOnly: true},
		{Loading: true},
		{IconOnly: true, Loading: true},
	} {
		if !validNonNegativeDp(visual.ContentGap(request)) {
			invalidVisual("Button", "ContentGap", "must be finite and non-negative")
		}
	}
	loading := visual.LoadingInset(
		layout.Context{Metric: unit.Metric{}},
		ButtonLoadingRequest{
			Inset: layout.Inset{Top: 8, Right: 8, Bottom: 8, Left: 8},
		},
	)
	if !validNonNegativeDp(loading.Top) {
		invalidVisual("Button", "LoadingInset.Top", "must be finite and non-negative")
	}
	if !validNonNegativeDp(loading.Right) {
		invalidVisual("Button", "LoadingInset.Right", "must be finite and non-negative")
	}
	if !validNonNegativeDp(loading.Bottom) {
		invalidVisual("Button", "LoadingInset.Bottom", "must be finite and non-negative")
	}
	if !validNonNegativeDp(loading.Left) {
		invalidVisual("Button", "LoadingInset.Left", "must be finite and non-negative")
	}

	states := []style.State{
		{},
		{Hovered: true},
		{Pressed: true},
		{Focused: true},
		{FocusVisible: true},
		{Disabled: true},
		{Loading: true},
	}
	for _, state := range states {
		root := style.Cascade(state, visual.Defaults, visual.Variant, visual.Size, visual.Override)
		if root.Box == nil {
			invalidVisual("Button", "Root.Box", "must be provided")
		}
		if !hasPositiveHeight(root.Box.Height, root.Box.MinHeight) {
			invalidVisual("Button", "Root.Box.Height", "height or minimum height must be finite and greater than 0")
		}
		validateButtonBox(root.Box)
		if root.Paint == nil {
			invalidVisual("Button", "Root.Paint", "must be provided")
		}
		validateButtonPaint(root.Paint)
		if root.Text == nil {
			invalidVisual("Button", "Root.Text", "must be provided")
		}
		if !hasColor(root.Text.Color) {
			invalidVisual("Button", "Root.Text.Color", "must resolve to a solid color")
		}
		if !hasOptionalScale(root.Trans) {
			invalidVisual("Button", "Root.Trans", "scale axes must be finite and greater than 0")
		}

		if _, ok := visual.ForegroundColor(root); !ok {
			invalidVisual("Button", "ForegroundColor", "must resolve the declaration color")
		}

		for _, part := range []style.Part{style.PartLabel, style.PartIcon} {
			resolved := style.CascadePart(state, part, visual.Defaults, visual.Variant, visual.Size, visual.Override)
			if resolved.Text == nil {
				invalidVisual("Button", "Resolved.Text", "must be provided")
			}
			if resolved.Text.FontSize == nil {
				invalidVisual("Button", "Resolved.Text.FontSize", "must be provided")
			}
			if math.IsNaN(float64(*resolved.Text.FontSize)) {
				invalidVisual("Button", "Resolved.Text.FontSize", "must be finite")
			}
			if math.IsInf(float64(*resolved.Text.FontSize), 0) {
				invalidVisual("Button", "Resolved.Text.FontSize", "must be finite")
			}
			if *resolved.Text.FontSize <= 0 {
				invalidVisual("Button", "Resolved.Text.FontSize", "must be greater than 0")
			}
			if !hasColor(resolved.Text.Color) {
				invalidVisual("Button", "Resolved.Text.Color", "must resolve to a solid color")
			}

			if _, ok := visual.ForegroundColor(resolved); !ok {
				invalidVisual("Button", "ForegroundColor", "must resolve the declaration color")
			}
		}
		indicator := style.CascadePart(
			state,
			style.PartIndicator,
			visual.Defaults,
			visual.Variant,
			visual.Size,
			visual.Override,
		)
		if indicator.Text == nil {
			invalidVisual("Button", "Indicator.Text", "must be provided")
		}
		if !hasColor(indicator.Text.Color) {
			invalidVisual("Button", "Indicator.Text.Color", "must resolve to a solid color")
		}

		if _, ok := visual.ForegroundColor(indicator); !ok {
			invalidVisual("Button", "ForegroundColor", "must resolve the declaration color")
		}
	}
	focus := style.Cascade(style.State{FocusVisible: true}, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if focus.Paint == nil {
		invalidVisual("Button", "Focus.Paint", "must be provided")
	}
	if focus.Paint.Outline == nil {
		invalidVisual("Button", "Focus.Paint.Outline", "must be provided")
	}
	if !validPositiveDp(focus.Paint.Outline.Width) {
		invalidVisual("Button", "Focus.Paint.Outline.Width", "must be finite and greater than 0")
	}
	if !hasColor(focus.Paint.Outline.Color) {
		invalidVisual("Button", "Focus.Paint.Outline.Color", "must resolve to a solid color")
	}

	if visual.Ripple {
		if visual.RippleDuration <= 0 {
			invalidVisual("Button", "RippleDuration", "must be greater than 0 when ripple is enabled")
		}
		if visual.RippleFadeDuration <= 0 {
			invalidVisual("Button", "RippleFadeDuration", "must be greater than 0 when ripple is enabled")
		}
		if visual.RippleRenderer == nil {
			invalidVisual("Button", "RippleRenderer", "must be provided when ripple is enabled")
		}
		if visual.RippleColor == nil {
			invalidVisual("Button", "RippleColor", "must be provided when ripple is enabled")
		}
	}
}

// ValidateButtonLoadingInset checks a resolver result at the point where a
// concrete layout request is applied. The provider contract validates a
// representative request before caching; this runtime check covers resolvers
// whose output depends on the current metric or resolved root inset.
func ValidateButtonLoadingInset(inset layout.Inset) {
	if !validNonNegativeDp(inset.Top) {
		invalidVisual("Button", "LoadingInset.Top", "must be finite and non-negative")
	}
	if !validNonNegativeDp(inset.Right) {
		invalidVisual("Button", "LoadingInset.Right", "must be finite and non-negative")
	}
	if !validNonNegativeDp(inset.Bottom) {
		invalidVisual("Button", "LoadingInset.Bottom", "must be finite and non-negative")
	}
	if !validNonNegativeDp(inset.Left) {
		invalidVisual("Button", "LoadingInset.Left", "must be finite and non-negative")
	}
}

func validateButtonBox(box *style.BoxStyle) {
	for _, metric := range []struct {
		name  string
		value *unit.Dp
	}{
		{"Width", box.Width}, {"MinWidth", box.MinWidth}, {"MaxWidth", box.MaxWidth},
		{"Height", box.Height}, {"MinHeight", box.MinHeight}, {"MaxHeight", box.MaxHeight},
	} {
		if metric.value != nil && !validFiniteDp(*metric.value) {
			invalidVisual("Button", "Root.Box."+metric.name, "must be finite when provided")
		}
	}
	if box.Padding == nil {
		invalidVisual("Button", "Root.Box.Padding", "must be provided")
	}
	for _, side := range []struct {
		name  string
		value unit.Dp
	}{
		{"Top", box.Padding.Top}, {"Right", box.Padding.Right},
		{"Bottom", box.Padding.Bottom}, {"Left", box.Padding.Left},
	} {
		if !validNonNegativeDp(side.value) {
			invalidVisual("Button", "Root.Box.Padding."+side.name, "must be finite and non-negative")
		}
	}
}

func validateButtonPaint(paint *style.PaintStyle) {
	if !hasPaint(paint.Background) {
		invalidVisual("Button", "Root.Paint.Background", "must provide a solid color or nonempty gradient")
	}
	if paint.Radius == nil {
		invalidVisual("Button", "Root.Paint.Radius", "must be provided")
	}
	if !validNonNegativeDp(*paint.Radius) {
		invalidVisual("Button", "Root.Paint.Radius", "must be finite and non-negative")
	}
	if !hasValidOpacity(paint.Opacity) {
		invalidVisual("Button", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}
	for _, part := range []struct {
		name   string
		border *style.BorderStyle
	}{
		{"Border", paint.Border}, {"BorderBottom", paint.BorderBottom},
	} {
		if border := part.border; border != nil {
			if border.Width != nil && !validNonNegativeDp(*border.Width) {
				invalidVisual("Button", "Root.Paint."+part.name+".Width", "must be finite and non-negative when provided")
			}
			if border.Color != nil && !hasColor(border.Color) {
				invalidVisual("Button", "Root.Paint."+part.name+".Color", "must resolve to a solid color when provided")
			}
		}
	}
	if outline := paint.Outline; outline != nil {
		if !validNonNegativeDp(outline.Width) {
			invalidVisual("Button", "Root.Paint.Outline.Width", "must be finite and non-negative")
		}
		if !validFiniteDp(outline.Offset) {
			invalidVisual("Button", "Root.Paint.Outline.Offset", "must be finite")
		}
		if !hasColor(outline.Color) {
			invalidVisual("Button", "Root.Paint.Outline.Color", "must resolve to a solid color")
		}
	}
}
