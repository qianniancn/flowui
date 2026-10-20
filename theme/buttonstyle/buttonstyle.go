// Package buttonstyle contains the shared button visual recipe used by
// concrete style packages. It knows how a theme's semantic values map to the
// button states, but it does not know anything about button behavior or Gio
// event handling.
package buttonstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Config contains the visual decisions consumed by the shared button recipe.
// A concrete style owns this value and supplies it on every provider call.
type Config struct {
	Radius             unit.Dp
	Height             unit.Dp
	SmallHeight        unit.Dp
	LargeHeight        unit.Dp
	PaddingX           unit.Dp
	SmallPaddingX      unit.Dp
	LargePaddingX      unit.Dp
	BorderWidth        unit.Dp
	ContentGap         unit.Dp
	FocusRingWidth     unit.Dp
	FocusGlowBlur      unit.Dp
	SpinnerStrokeWidth unit.Dp
	SpinnerSmall       unit.Dp
	SpinnerMedium      unit.Dp
	SpinnerLarge       unit.Dp
	SpinnerPeriod      time.Duration
	PressedScaleSmall  float32
	PressedScaleMedium float32
	PressedScaleLarge  float32
	Ripple             bool
	RippleDuration     time.Duration
	RippleFadeDuration time.Duration
	FontWeight         int
	HoverBorderColor   color.NRGBA
	UnderShadow        unit.Dp

	GroupSeparatorWidth   unit.Dp
	GroupSeparatorLength  float32
	GroupSeparatorOpacity float32
}

// DefaultConfig returns a complete, independent starting point for a custom
// button recipe. Change only the fields your design needs, then pass the result
// to BuildButton. BuildButton uses values as supplied, including valid zeros.
func DefaultConfig() Config {
	return Config{
		Radius: 8, BorderWidth: 1, ContentGap: 8,
		Height: 36, SmallHeight: 32, LargeHeight: 40,
		PaddingX: 14, SmallPaddingX: 10, LargePaddingX: 18,
		FocusRingWidth: 2, SpinnerStrokeWidth: 2,
		SpinnerSmall: 14, SpinnerMedium: 16, SpinnerLarge: 18,
		SpinnerPeriod:     900 * time.Millisecond,
		PressedScaleSmall: .98, PressedScaleMedium: .97, PressedScaleLarge: .96,
		GroupSeparatorWidth: 1, GroupSeparatorLength: .5, GroupSeparatorOpacity: .15,
	}
}

// BuildButton builds the shared button declaration from the supplied theme.
// Style packages own config and may wrap the recipe for family-specific rules.
func BuildButton(active *theme.Theme, request theme.ButtonRequest, config Config) theme.ButtonVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	semanticSize := request.Size
	spinnerSize := func() unit.Dp {
		return buttonSpinnerSize(config, semanticSize)
	}
	spinnerStrokeWidth := func() unit.Dp {
		return config.SpinnerStrokeWidth
	}
	spinnerPeriod := func() time.Duration {
		return config.SpinnerPeriod
	}
	contentGap := func(theme.ButtonContentGapRequest) unit.Dp {
		return config.ContentGap
	}
	size := sizeRules(active, request.Size, request.IconOnly, request.Grouped, config)
	if request.Grouped {
		size = style.Join(size, groupDeclaration(request.GroupCorners, config))
	}
	return theme.ButtonVisual{
		Defaults: defaultRules(active, config),
		Variant:  variantRules(active, request.Variant, config),
		Size:     size,
		Override: func() style.Style {
			if request.Override == nil {
				return style.Style{}
			}
			return request.Override.Declaration
		}(),
		Ripple:             config.Ripple,
		RippleDuration:     config.RippleDuration,
		RippleFadeDuration: config.RippleFadeDuration,
		Spinner:            spinnerRenderer,
		SpinnerSize:        spinnerSize,
		SpinnerStrokeWidth: spinnerStrokeWidth,
		SpinnerPeriod:      spinnerPeriod,
		ContentGap:         contentGap,
		LoadingInset: func(gtx layout.Context, loading theme.ButtonLoadingRequest) layout.Inset {
			return loadingInset(
				gtx,
				loading,
				spinnerSize(),
				contentGap(theme.ButtonContentGapRequest{Loading: true}),
			)
		},
		ForegroundColor: foregroundColor,
		RippleRenderer:  rippleRenderer,
		RippleColor:     rippleColor,
		Motion:          active.Motion,
		Group: theme.ButtonGroupVisual{
			SeparatorRenderer: func(gtx layout.Context, separator theme.ButtonGroupSeparatorRequest) {
				groupSeparatorRenderer(gtx, separator, config.GroupSeparatorWidth, config.GroupSeparatorLength, config.GroupSeparatorOpacity)
			},
		},
	}
}

func buttonSpinnerSize(config Config, size theme.ButtonSize) unit.Dp {
	switch size {
	case theme.ButtonSmall:
		return config.SpinnerSmall
	case theme.ButtonLarge:
		return config.SpinnerLarge
	default:
		return config.SpinnerMedium
	}
}

func loadingInset(
	gtx layout.Context,
	request theme.ButtonLoadingRequest,
	spinner unit.Dp,
	contentGap unit.Dp,
) layout.Inset {
	padding := gtx.Dp(request.Inset.Left) + gtx.Dp(request.Inset.Right)
	extra := gtx.Dp(spinner) + gtx.Dp(contentGap)
	remaining := max(padding-extra, 0)
	left := remaining / 2
	request.Inset.Left = gtx.Metric.PxToDp(left)
	request.Inset.Right = gtx.Metric.PxToDp(remaining - left)
	return request.Inset
}

func groupDeclaration(group theme.ButtonGroupRequest, config Config) style.Style {
	result := style.Style{}.Radius(0)
	radius := config.Radius
	if group.NorthWest {
		result = result.RadiusTopLeft(radius)
	}
	if group.NorthEast {
		result = result.RadiusTopRight(radius)
	}
	if group.SouthEast {
		result = result.RadiusBottomRight(radius)
	}
	if group.SouthWest {
		result = result.RadiusBottomLeft(radius)
	}
	return result
}

func motionDuration(active *theme.Theme, fallback time.Duration) time.Duration {
	if active.Motion.ControlDuration > 0 {
		return active.Motion.ControlDuration
	}
	return fallback
}

func defaultRules(active *theme.Theme, config Config) style.Style {
	return style.Style{}.
		Radius(config.Radius).
		FontWeight(buttonFontWeight(config)).
		Outline(config.FocusRingWidth, 1, style.WithAlpha(active.Semantic().Ring, 0)).
		Opacity(1).
		Scale(1).
		Transition(style.PropBackgroundColor, motionDuration(active, 100*time.Millisecond)).
		Transition(style.PropTextColor, motionDuration(active, 100*time.Millisecond)).
		Transition(style.PropBorderColor, motionDuration(active, 100*time.Millisecond)).
		Transition(style.PropOpacity, motionDuration(active, 100*time.Millisecond)).
		Transition(style.PropOutlineColor, motionDuration(active, 100*time.Millisecond)).
		Transition(style.PropTransform, motionDuration(active, 160*time.Millisecond)).
		On(style.Disabled, disabledRow(active)).
		OnNot(style.FocusVisible, style.Disabled, focusRow(active, config))
}

func focusRow(active *theme.Theme, config Config) style.Style {
	ring := active.Semantic().Ring
	row := style.Style{}.
		Outline(config.FocusRingWidth, 1, style.Solid(ring))
	if blur := config.FocusGlowBlur; blur > 0 {
		row = row.BoxShadow(0, 0, blur, 0, style.Solid(ring))
	}
	return row
}

func variantRules(active *theme.Theme, variant theme.ButtonVariant, config Config) style.Style {
	colors := colors(active, variant)
	declaration := style.Style{}.
		Background(style.Solid(colors.Bg)).
		TextColor(style.Solid(colors.Fg))
	if gradient := active.Semantic().PrimaryGradient; variant == theme.ButtonPrimary && gradient != nil && len(gradient.Stops) > 0 {
		declaration = declaration.Background(toStyleGradient(gradient))
	}
	if colors.HasBorder {
		declaration = declaration.BorderColor(style.Solid(colors.Border)).BorderWidth(config.BorderWidth)
	}
	if config.UnderShadow > 0 && colors.Shadow.A != 0 {
		declaration = declaration.BoxShadow(0, config.UnderShadow, 0, 0, style.Solid(colors.Shadow))
	}
	hover := style.Style{}.Background(style.Solid(colors.Hover))
	if colors.HoverFg != (color.NRGBA{}) {
		hover = hover.TextColor(style.Solid(colors.HoverFg))
	}
	if colors.HasBorder {
		switch {
		case colors.HoverBorder != (color.NRGBA{}):
			hover = hover.BorderColor(style.Solid(colors.HoverBorder))
		case config.HoverBorderColor.A != 0:
			hover = hover.BorderColor(style.Solid(config.HoverBorderColor))
		}
	}
	pressed := style.Style{}.Background(style.Solid(colors.Pressed))
	if colors.PressedFg != (color.NRGBA{}) {
		pressed = pressed.TextColor(style.Solid(colors.PressedFg))
	}
	if colors.HasBorder && colors.PressedBorder != (color.NRGBA{}) {
		pressed = pressed.BorderColor(style.Solid(colors.PressedBorder))
	}
	// Disabled is repeated at the variant layer so it remains the final visual
	// authority after the defaults/variant cascade. Keeping the shared row in
	// Defaults preserves the neutral focus and transition contract, while this
	// row prevents a filled variant from repainting a disabled control.
	return declaration.
		On(style.Hovered, hover).
		On(style.Pressed, pressed).
		On(style.Disabled, disabledRow(active))
}

func sizeRules(active *theme.Theme, size theme.ButtonSize, iconOnly, grouped bool, config Config) style.Style {
	height := config.Height
	padding := config.PaddingX
	textSize := active.Typography.ControlSize
	switch size {
	case theme.ButtonSmall:
		height, padding = config.SmallHeight, config.SmallPaddingX
	case theme.ButtonLarge:
		height, padding = config.LargeHeight, config.LargePaddingX
		textSize += 2
	}
	if iconOnly {
		padding = 0
	}
	declaration := style.Style{}.Height(height).PaddingX(padding).FontSize(textSize)
	if iconOnly {
		declaration = declaration.Width(height)
	}
	if grouped {
		return declaration
	}
	return declaration.Rules(style.Rule{
		When: style.Pressed,
		Set:  style.Style{}.Scale(pressedScale(config, size)),
		Tween: &style.Tween{
			Property: style.PropTransform,
			Duration: motionDuration(active, 90*time.Millisecond),
		},
	})
}

type palette struct {
	Bg, Hover, Pressed, Fg, Border                 color.NRGBA
	HasBorder                                      bool
	HoverFg, HoverBorder, PressedFg, PressedBorder color.NRGBA
	Shadow                                         color.NRGBA
}

// colors resolves a button variant into its static semantic palette.
//
// This is an implementation detail of the shared theme recipe. Consumers use
// BuildButton and the neutral ButtonVisual protocol; exposing the intermediate
// palette would recreate the old theme facade as a second public API.
func colors(active *theme.Theme, variant theme.ButtonVariant) palette {
	design := appearance.Design{}
	switch variant {
	case theme.ButtonPrimary:
		design = appearance.Design{Tone: theme.TonePrimary, Variant: theme.VariantSolid, Accent: theme.TonePrimary, UnderShadow: .10}
	case theme.ButtonSecondary:
		design = appearance.Design{Tone: theme.TonePrimary, Variant: theme.VariantSoft, Accent: theme.TonePrimary, NeutralSurface: true}
	case theme.ButtonTertiary:
		design = appearance.Design{Tone: theme.ToneDefault, Variant: theme.VariantSoft, Accent: theme.TonePrimary, NeutralSurface: true}
	case theme.ButtonGhost:
		design = appearance.Design{Tone: theme.ToneDefault, Variant: theme.VariantGhost, Accent: theme.TonePrimary}
	case theme.ButtonOutline:
		design = appearance.Design{Tone: theme.ToneDefault, Variant: theme.VariantOutline, Accent: theme.TonePrimary, NeutralSurface: true, NeutralBorder: true, UnderShadow: .02, ShadowBlack: true}
	case theme.ButtonDanger:
		design = appearance.Design{Tone: theme.ToneDanger, Variant: theme.VariantSolid, Accent: theme.TonePrimary, UnderShadow: .06}
	case theme.ButtonDangerSoft:
		design = appearance.Design{Tone: theme.ToneDanger, Variant: theme.VariantSoft, Accent: theme.TonePrimary}
	case theme.ButtonDangerOutline:
		design = appearance.Design{Tone: theme.ToneDanger, Variant: theme.VariantOutline, Accent: theme.ToneDanger, UnderShadow: .06}
	}
	resolved := appearance.Resolve(active, design)
	result := palette{
		Bg: resolved.Bg, Hover: resolved.Hover, Pressed: resolved.Pressed,
		Fg: resolved.Fg, Border: resolved.Border, HasBorder: resolved.HasBorder,
		HoverFg: resolved.HoverFg, HoverBorder: resolved.HoverBorder,
		PressedFg: resolved.PressedFg, PressedBorder: resolved.PressedBorder,
		Shadow: resolved.Shadow,
	}
	return result
}

func disabledRow(active *theme.Theme) style.Style {
	sem := active.Semantic()
	if sem.DisabledBackground.A != 0 {
		return style.Style{}.
			Background(style.Solid(sem.DisabledBackground)).
			TextColor(style.Solid(sem.DisabledForeground)).
			BorderColor(style.Solid(sem.DisabledBorder)).
			BoxShadowNone().
			Opacity(1)
	}
	return style.Style{}.Opacity(active.DisabledOpacityValue())
}

func buttonFontWeight(config Config) int {
	if value := config.FontWeight; value != 0 {
		return value - 400
	}
	return int(font.Medium)
}

func pressedScale(config Config, size theme.ButtonSize) float32 {
	switch size {
	case theme.ButtonSmall:
		return config.PressedScaleSmall
	case theme.ButtonLarge:
		return config.PressedScaleLarge
	default:
		return config.PressedScaleMedium
	}
}

func toStyleGradient(value *theme.Gradient) style.Gradient {
	stops := make([]style.GradientStop, len(value.Stops))
	for index, stop := range value.Stops {
		stops[index] = style.GradientStop{Offset: stop.Offset, Color: style.Solid(stop.Color)}
	}
	return style.Gradient{AngleDegrees: value.AngleDegrees, Stops: stops}
}
