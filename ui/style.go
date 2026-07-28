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

type Style = style.Style
type ResolvedStyle = style.ResolvedStyle
type StyleState = style.StyleState
type Condition = style.Condition
type StylePart = style.Part
type StyleToken = style.StyleToken
type PaintSource = style.PaintSource
type ColorSource = style.ColorSource
type SolidColor = style.SolidColor
type ThemeColor = style.ThemeColor
type AlphaColor = style.AlphaColor
type ColorToken = style.ColorToken
type Gradient = style.StyleGradient
type GradientStop = style.StyleGradientStop
type Insets = style.Insets
type BoxStyle = style.BoxStyle
type PaintStyle = style.PaintStyle
type CornerRadii = style.CornerRadii
type StyleShadow = style.Shadow
type ShadowProfile = style.ShadowProfile
type OutlineStyle = style.OutlineStyle
type BorderStyle = style.BorderStyle
type TextStyle = style.TextStyle
type TransformStyle = style.TransformStyle
type StyleTransition = style.Transition
type PropertyID = style.PropertyID
type EaseFunc = style.EaseFunc
type TransitionOption = style.TransitionOption
type StyleOverflow = style.Overflow
type StyleTextAlign = style.TextAlign
type StyleCursor = pointer.Cursor

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

const (
	StyleOverflowVisible = style.OverflowVisible
	StyleOverflowHidden  = style.OverflowHidden

	StyleTextAlignStart  = style.TextAlignStart
	StyleTextAlignCenter = style.TextAlignCenter
	StyleTextAlignEnd    = style.TextAlignEnd
)

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

func RGB(value uint32) SolidColor {
	return style.RGB(value)
}

func RGBA(value uint32) SolidColor {
	return style.RGBA(value)
}

func Color(value color.Color) SolidColor {
	return style.Color(value)
}

func WithAlpha(value ColorSource, alpha float32) AlphaColor {
	return style.WithAlpha(value, alpha)
}

func All(conditions ...Condition) Condition {
	return style.All(conditions...)
}

func Any(conditions ...Condition) Condition {
	return style.Any(conditions...)
}

func Not(condition Condition) Condition {
	return style.Not(condition)
}

// If adapts an MVU model value for use with When.
func If(value bool) Condition {
	return style.If(value)
}

func TransitionDelay(value time.Duration) TransitionOption {
	return style.TransitionDelay(value)
}

func TransitionEase(value EaseFunc) TransitionOption {
	return style.TransitionEase(value)
}

func Cascade(state StyleState, layers ...Style) ResolvedStyle {
	return style.Cascade(state, layers...)
}

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
