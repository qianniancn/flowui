package ui

import (
	"image/color"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

// Style is an immutable declaration of layout, paint, text, and transition
// properties. Fluent helpers return new values that can be composed.
type Style = style.Style

// ResolvedStyle is the per-frame style snapshot produced by a resolver.
type ResolvedStyle = style.ResolvedStyle

// StyleState contains the interaction state used by conditional styles.
type StyleState = style.StyleState

// Condition reports whether a StyleState matches a conditional style rule.
type Condition = style.Condition

// StylePart names a visual part of a compound component.
type StylePart = style.Part

// StyleToken identifies a metric supplied by the active theme.
type StyleToken = style.StyleToken

// PaintSource is a background or other paint value accepted by Style.
type PaintSource = style.PaintSource

// ColorSource is a solid or theme-resolved color accepted by Style.
type ColorSource = style.ColorSource

// SolidColor stores an explicit non-premultiplied color.
type SolidColor = style.SolidColor

// ThemeColor resolves a semantic color token from the active theme.
type ThemeColor = style.ThemeColor

// AlphaColor replaces the alpha channel of another ColorSource.
type AlphaColor = style.AlphaColor

// ColorToken identifies a semantic color in the active theme.
type ColorToken = style.ColorToken

// Gradient describes a linear gradient paint source.
type Gradient = style.StyleGradient

// GradientStop associates a normalized offset with a color source.
type GradientStop = style.StyleGradientStop

// Insets stores top, right, bottom, and left spacing values.
type Insets = style.Insets

// BoxStyle contains layout constraints and spacing properties.
type BoxStyle = style.BoxStyle

// PaintStyle contains background, border, radius, shadow, and opacity values.
type PaintStyle = style.PaintStyle

// CornerRadii stores the radius of each box corner.
type CornerRadii = style.CornerRadii

// StyleShadow describes one custom shadow layer.
type StyleShadow = style.Shadow

// ShadowProfile selects a shadow profile from the active theme.
type ShadowProfile = style.ShadowProfile

// OutlineStyle describes an outline's width, offset, and color.
type OutlineStyle = style.OutlineStyle

// BorderStyle describes a border color and width.
type BorderStyle = style.BorderStyle

// TextStyle contains font, color, alignment, wrapping, and truncation values.
type TextStyle = style.TextStyle

// TransformStyle contains translation, scale, and rotation values.
type TransformStyle = style.TransformStyle

// StyleTransition describes an animated style-property change.
type StyleTransition = style.Transition

// PropertyID identifies a style property that can be animated.
type PropertyID = style.PropertyID

// EaseFunc maps normalized transition progress to eased progress.
type EaseFunc = style.EaseFunc

// TransitionOption configures one Style transition.
type TransitionOption = style.TransitionOption

// StyleOverflow controls whether child paint may leave its box.
type StyleOverflow = style.Overflow

// StyleTextAlign controls text alignment inside its layout width.
type StyleTextAlign = style.TextAlign

// StyleCursor identifies the pointer cursor shown over a widget.
type StyleCursor = pointer.Cursor

// Interaction states used by conditional styles.
var (
	Hovered       = style.Hovered
	Pressed       = style.Pressed
	Focused       = style.Focused
	FocusVisible  = style.FocusVisible
	Disabled      = style.Disabled
	Selected      = style.Selected
	Checked       = style.Checked
	Indeterminate = style.Indeterminate
	ReadOnly      = style.ReadOnly
	Invalid       = style.Invalid
	Loading       = style.Loading
	Open          = style.Open
	ExpandedState = style.ExpandedState
	Dragging      = style.Dragging
	DropTarget    = style.DropTarget
)

// Component parts, theme metrics, and animatable properties.
const (
	PartRoot        = style.PartRoot
	PartContent     = style.PartContent
	PartLabel       = style.PartLabel
	PartDescription = style.PartDescription
	PartIcon        = style.PartIcon
	PartTrack       = style.PartTrack
	PartFill        = style.PartFill
	PartThumb       = style.PartThumb
	PartIndicator   = style.PartIndicator
	PartPanel       = style.PartPanel
	PartItem        = style.PartItem
	PartBackdrop    = style.PartBackdrop
	PartPlaceholder = style.PartPlaceholder
	PartSelection   = style.PartSelection
	PartPrefix      = style.PartPrefix
	PartSuffix      = style.PartSuffix
	ShadowSurface   = style.ShadowSurface
	ShadowOverlay   = style.ShadowOverlay
	ShadowMenu      = style.ShadowMenu

	TokenBodyFontSize         = style.TokenBodyFontSize
	TokenControlFontSize      = style.TokenControlFontSize
	TokenSmallFontSize        = style.TokenSmallFontSize
	TokenControlRadius        = style.TokenControlRadius
	TokenPopoverRadius        = style.TokenPopoverRadius
	TokenItemRadius           = style.TokenItemRadius
	TokenCheckboxRadius       = style.TokenCheckboxRadius
	TokenControlHeight        = style.TokenControlHeight
	TokenSmallControlHeight   = style.TokenSmallControlHeight
	TokenLargeControlHeight   = style.TokenLargeControlHeight
	TokenControlPaddingX      = style.TokenControlPaddingX
	TokenSmallControlPaddingX = style.TokenSmallControlPaddingX
	TokenLargeControlPaddingX = style.TokenLargeControlPaddingX
	TokenIconButtonSize       = style.TokenIconButtonSize
	TokenPanelPadding         = style.TokenPanelPadding
	TokenItemHeight           = style.TokenItemHeight

	PropBackgroundColor = style.PropBackgroundColor
	PropTextColor       = style.PropTextColor
	PropBorderColor     = style.PropBorderColor
	PropOutlineColor    = style.PropOutlineColor
	PropOpacity         = style.PropOpacity
	PropRadius          = style.PropRadius
	PropTransform       = style.PropTransform
)

// Pointer cursor values understood by Style.
const (
	CursorDefault                  = pointer.CursorDefault
	CursorNone                     = pointer.CursorNone
	CursorText                     = pointer.CursorText
	CursorVerticalText             = pointer.CursorVerticalText
	CursorPointer                  = pointer.CursorPointer
	CursorCrosshair                = pointer.CursorCrosshair
	CursorAllScroll                = pointer.CursorAllScroll
	CursorColResize                = pointer.CursorColResize
	CursorRowResize                = pointer.CursorRowResize
	CursorGrab                     = pointer.CursorGrab
	CursorGrabbing                 = pointer.CursorGrabbing
	CursorNotAllowed               = pointer.CursorNotAllowed
	CursorWait                     = pointer.CursorWait
	CursorProgress                 = pointer.CursorProgress
	CursorNorthWestResize          = pointer.CursorNorthWestResize
	CursorNorthEastResize          = pointer.CursorNorthEastResize
	CursorSouthWestResize          = pointer.CursorSouthWestResize
	CursorSouthEastResize          = pointer.CursorSouthEastResize
	CursorNorthSouthResize         = pointer.CursorNorthSouthResize
	CursorEastWestResize           = pointer.CursorEastWestResize
	CursorWestResize               = pointer.CursorWestResize
	CursorEastResize               = pointer.CursorEastResize
	CursorNorthResize              = pointer.CursorNorthResize
	CursorSouthResize              = pointer.CursorSouthResize
	CursorNorthEastSouthWestResize = pointer.CursorNorthEastSouthWestResize
	CursorNorthWestSouthEastResize = pointer.CursorNorthWestSouthEastResize
)

// Text overflow and alignment values.
const (
	StyleOverflowVisible = style.OverflowVisible
	StyleOverflowHidden  = style.OverflowHidden

	StyleTextAlignStart  = style.TextAlignStart
	StyleTextAlignCenter = style.TextAlignCenter
	StyleTextAlignEnd    = style.TextAlignEnd
)

// Semantic theme colors exposed as ColorSource values.
var (
	TokenBackground                 = style.TokenBackground
	TokenSurface                    = style.TokenSurface
	TokenSurfaceForeground          = style.TokenSurfaceForeground
	TokenSurfaceSecondary           = style.TokenSurfaceSecondary
	TokenSurfaceSecondaryForeground = style.TokenSurfaceSecondaryForeground
	TokenSurfaceTertiary            = style.TokenSurfaceTertiary
	TokenSurfaceTertiaryForeground  = style.TokenSurfaceTertiaryForeground
	TokenSurfaceHover               = style.TokenSurfaceHover
	TokenSurfacePressed             = style.TokenSurfacePressed
	TokenSurfaceRaised              = style.TokenSurfaceRaised
	TokenOverlay                    = style.TokenOverlay
	TokenOverlayForeground          = style.TokenOverlayForeground
	TokenForeground                 = style.TokenForeground
	TokenMutedForeground            = style.TokenMutedForeground
	TokenBorder                     = style.TokenBorder
	TokenSeparator                  = style.TokenSeparator
	TokenDefault                    = style.TokenDefault
	TokenDefaultForeground          = style.TokenDefaultForeground
	TokenDefaultHover               = style.TokenDefaultHover
	TokenFieldBackground            = style.TokenFieldBackground
	TokenFieldHover                 = style.TokenFieldHover
	TokenFieldForeground            = style.TokenFieldForeground
	TokenFieldPlaceholder           = style.TokenFieldPlaceholder
	TokenFieldFocus                 = style.TokenFieldFocus
	TokenSegment                    = style.TokenSegment
	TokenSegmentForeground          = style.TokenSegmentForeground
	TokenAccent                     = style.TokenAccent
	TokenAccentHover                = style.TokenAccentHover
	TokenAccentPressed              = style.TokenAccentPressed
	TokenAccentForeground           = style.TokenAccentForeground
	TokenAccentSoft                 = style.TokenAccentSoft
	TokenAccentSoftHover            = style.TokenAccentSoftHover
	TokenAccentSoftForeground       = style.TokenAccentSoftForeground
	TokenSuccess                    = style.TokenSuccess
	TokenSuccessForeground          = style.TokenSuccessForeground
	TokenSuccessSoft                = style.TokenSuccessSoft
	TokenSuccessSoftForeground      = style.TokenSuccessSoftForeground
	TokenWarning                    = style.TokenWarning
	TokenWarningForeground          = style.TokenWarningForeground
	TokenWarningSoft                = style.TokenWarningSoft
	TokenWarningSoftForeground      = style.TokenWarningSoftForeground
	TokenDanger                     = style.TokenDanger
	TokenDangerHover                = style.TokenDangerHover
	TokenDangerPressed              = style.TokenDangerPressed
	TokenDangerForeground           = style.TokenDangerForeground
	TokenDangerSoft                 = style.TokenDangerSoft
	TokenDangerSoftHover            = style.TokenDangerSoftHover
	TokenDangerSoftForeground       = style.TokenDangerSoftForeground
	TokenFocus                      = style.TokenFocus
	TokenSelection                  = style.TokenSelection
	TokenSurfaceShadow              = style.TokenSurfaceShadow
	TokenOverlayShadow              = style.TokenOverlayShadow
)

// RGB creates a solid color from a 0xRRGGBB value.
func RGB(value uint32) SolidColor {
	return style.RGB(value)
}

// RGBA creates a solid color from a 0xRRGGBBAA value.
func RGBA(value uint32) SolidColor {
	return style.RGBA(value)
}

// Color converts a standard library color to a SolidColor.
func Color(value color.Color) SolidColor {
	return style.Color(value)
}

// WithAlpha creates a color source with a replacement alpha in [0, 1].
func WithAlpha(value ColorSource, alpha float32) AlphaColor {
	return style.WithAlpha(value, alpha)
}

// All matches when every supplied condition matches.
func All(conditions ...Condition) Condition {
	return style.All(conditions...)
}

// Any matches when at least one supplied condition matches.
func Any(conditions ...Condition) Condition {
	return style.Any(conditions...)
}

// Not negates a style condition.
func Not(condition Condition) Condition {
	return style.Not(condition)
}

// If adapts an MVU model value for use with When.
func If(value bool) Condition {
	return style.If(value)
}

// TransitionDelay sets the delay before a transition starts.
func TransitionDelay(value time.Duration) TransitionOption {
	return style.TransitionDelay(value)
}

// TransitionEase sets the easing function used by a transition.
func TransitionEase(value EaseFunc) TransitionOption {
	return style.TransitionEase(value)
}

// Cascade resolves style layers for a component state.
func Cascade(state StyleState, layers ...Style) ResolvedStyle {
	return style.Cascade(state, layers...)
}

// CascadePart resolves a named component part from style layers.
func CascadePart(state StyleState, part StylePart, layers ...Style) ResolvedStyle {
	return style.CascadePart(state, part, layers...)
}

// ResolveStyle resolves a custom component's base style, inherited scopes,
// and instance style in that order, including theme tokens and transitions.
func ResolveStyle(
	ctx *Context,
	gtx layout.Context,
	key string,
	state StyleState,
	base, instance Style,
) ResolvedStyle {
	return styleruntime.Resolve(ctx, gtx, frame.FullKey(ctx, key), state, base, Style{}, Style{}, instance)
}

// ResolveStyleStatic resolves the same cascade without retaining transition
// state. Custom components can use it for measurement passes.
func ResolveStyleStatic(ctx *Context, state StyleState, base, instance Style) ResolvedStyle {
	return styleruntime.ResolveStatic(ctx, state, base, Style{}, Style{}, instance)
}

// ResolveStylePart resolves a named part of a custom component, including
// inherited scopes, theme tokens, conditions, and transitions.
func ResolveStylePart(
	ctx *Context,
	gtx layout.Context,
	key string,
	part StylePart,
	state StyleState,
	base, instance Style,
) ResolvedStyle {
	return styleruntime.ResolvePart(ctx, gtx, frame.FullKey(ctx, key), part, state, base, Style{}, Style{}, instance)
}

// ResolveStylePartStatic resolves the same part without transition state.
func ResolveStylePartStatic(ctx *Context, part StylePart, state StyleState, base, instance Style) ResolvedStyle {
	return styleruntime.ResolvePartStatic(ctx, part, state, base, Style{}, Style{}, instance)
}

// StyleScope applies value to all descendant components that support the
// common style runtime. An instance Style call still has the final say.
func StyleScope(value Style, child Widget) Widget {
	return WidgetFunc(func(ctx *Context, gtx layout.Context) layout.Dimensions {
		if child == nil {
			return layout.Dimensions{}
		}
		restore := frame.PushStyle(ctx, value)
		defer restore()
		return child.Layout(ctx, gtx)
	})
}
