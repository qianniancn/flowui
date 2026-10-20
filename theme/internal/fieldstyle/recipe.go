package fieldstyle

import (
	"image/color"
	"math"
	"time"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// styleContext is the internal recipe input assembled by BuildInput.
// Providers exchange the neutral theme.InputRequest; this implementation
// context is deliberately not part of the recipe package API.
type styleContext struct {
	Theme     *theme.Theme
	Variant   theme.InputVariant
	Mode      theme.InputStyleMode
	FullWidth bool
	Height    unit.Dp
	// Size selects the small/medium/large geometry tier. Zero is medium.
	Size theme.InputSize
}

// FieldDeclarationOptions describes the shared field face used by compound
// widgets such as Select, ComboBox, and DatePicker. The owning component still
// supplies its own parts and resolves the returned declaration; this helper
// keeps the common field chrome in the theme Recipe boundary.
type FieldDeclarationOptions struct {
	TargetPart          style.Part
	Height              unit.Dp
	MinHeight           unit.Dp
	Radius              unit.Dp
	PaddingX            unit.Dp
	PaddingY            unit.Dp
	TextSize            unit.Sp
	LineHeight          unit.Sp
	FocusRingWidth      unit.Dp
	InvalidOutlineWidth unit.Dp
	FocusGlowBlur       unit.Dp
	FocusHaloBlur       unit.Dp
	FocusHaloOpacity    float32
	FocusShadowSpread   unit.Dp
	BorderWidth         unit.Dp
	ShadowColor         color.NRGBA
	ShadowOpacity       float32
	ShadowStrength      float32
	FillWidth           bool
}

// FieldDefaultRules returns the shared field face rule table for a compound
// component. Callers own their component parts, state, and layout while this
// package owns the common field chrome.
func FieldDefaultRules(activeTheme *theme.Theme, variant theme.InputVariant, options FieldDeclarationOptions) style.Style {
	return fieldRules(activeTheme, variant, fieldDeclarationOptions(options))
}

// inputTierGeometry resolves the per-size radius/padding/type tiers; zero
// values fall back to the medium tier.
func inputTierGeometry(tokens theme.InputTheme, size theme.InputSize) (radius, paddingX unit.Dp, textSize, lineHeight unit.Sp) {
	radius, paddingX, textSize, lineHeight = tokens.Radius, tokens.PaddingX, tokens.TextSize, tokens.LineHeight
	switch size {
	case theme.InputSmall:
		if tokens.SmallRadius > 0 {
			radius = tokens.SmallRadius
		}
		if tokens.SmallPaddingX > 0 {
			paddingX = tokens.SmallPaddingX
		}
	case theme.InputLarge:
		if tokens.LargeRadius > 0 {
			radius = tokens.LargeRadius
		}
		if tokens.LargePaddingX > 0 {
			paddingX = tokens.LargePaddingX
		}
		if tokens.LargeTextSize > 0 {
			textSize = tokens.LargeTextSize
		}
		if tokens.LargeLineHeight > 0 {
			lineHeight = tokens.LargeLineHeight
		}
	}
	return radius, paddingX, textSize, lineHeight
}

type fieldDeclarationOptions struct {
	TargetPart                                    style.Part
	Height, MinHeight, Radius, PaddingX, PaddingY unit.Dp
	TextSize, LineHeight                          unit.Sp
	FocusRingWidth, InvalidOutlineWidth           unit.Dp
	// FocusGlowBlur adds a blurred focus-color glow behind the focus ring.
	FocusGlowBlur unit.Dp
	// FocusHaloBlur adds a soft focus halo without changing the ring path.
	FocusHaloBlur unit.Dp
	// FocusHaloOpacity scales that halo. Zero uses full opacity.
	FocusHaloOpacity float32
	// FocusShadowSpread draws focus/error/warning feedback as a spread shadow
	// ring of this width. Zero keeps the outline ring.
	FocusShadowSpread unit.Dp
	// BorderWidth draws the outlined field's box border. Zero keeps it borderless.
	BorderWidth                   unit.Dp
	ShadowColor                   color.NRGBA
	ShadowOpacity, ShadowStrength float32
	FillWidth                     bool
}

// fieldDuration is the field box's feedback baseline: the style family's
// ControlDuration when set, else the component constant.
func fieldDuration(activeTheme *theme.Theme) time.Duration {
	if activeTheme.Motion.ControlDuration > 0 {
		return activeTheme.Motion.ControlDuration
	}
	return 150 * time.Millisecond
}

func addShadows(declaration style.Style, activeTheme *theme.Theme, shadowColor color.NRGBA, opacity, strength float32) style.Style {
	sem := activeTheme.Semantic()
	shadowColor = theme.ColorOr(shadowColor, sem.SurfaceShadow)
	shadow := draw.ThemeShadow(activeTheme.Shadows.Control, shadowColor, opacity)
	if !(strength > 0) || math.IsInf(float64(strength), 0) {
		return declaration
	}
	for _, layer := range shadow.Layers {
		layer.Color.A = uint8(min(float32(layer.Color.A)*strength, 255) + .5)
		declaration = declaration.BoxShadow(unit.Dp(layer.OffsetX), unit.Dp(layer.OffsetY), unit.Dp(layer.Blur), unit.Dp(layer.Spread), style.Solid(layer.Color))
	}
	return declaration
}
