package style

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	giotext "gioui.org/text"
	"gioui.org/unit"
)

// Style is an immutable style declaration. Its fluent methods return an
// independent value; only a resolver exposes its computed properties.
type Style struct {
	box   *BoxStyle
	paint *PaintStyle
	text  *TextStyle
	trans *TransformStyle

	transitions []Transition

	conditions []condition
	parts      map[Part]Style
	tokens     []StyleToken
}

// ResolvedStyle is a computed snapshot used for layout and drawing. It is
// intentionally mutable because transitions resolve into a per-frame copy.
type ResolvedStyle struct {
	Box         *BoxStyle
	Paint       *PaintStyle
	Text        *TextStyle
	Trans       *TransformStyle
	Transitions []Transition
}

// StyleToken maps a theme metric to its semantic style property.
type StyleToken uint8

const (
	TokenBodyFontSize StyleToken = iota
	TokenControlFontSize
	TokenSmallFontSize
	TokenControlRadius
	TokenPopoverRadius
	TokenItemRadius
	TokenCheckboxRadius
	TokenControlHeight
	TokenSmallControlHeight
	TokenLargeControlHeight
	TokenControlPaddingX
	TokenSmallControlPaddingX
	TokenLargeControlPaddingX
	TokenIconButtonSize
	TokenPanelPadding
	TokenItemHeight
)

// Part identifies a named element inside a compound component. Applications
// may define additional names for custom components.
type Part string

const (
	PartRoot        Part = ""
	PartContent     Part = "content"
	PartLabel       Part = "label"
	PartDescription Part = "description"
	PartIcon        Part = "icon"
	PartTrack       Part = "track"
	PartFill        Part = "fill"
	PartThumb       Part = "thumb"
	PartIndicator   Part = "indicator"
	PartPanel       Part = "panel"
	PartItem        Part = "item"
	PartBackdrop    Part = "backdrop"
	PartPlaceholder Part = "placeholder"
	PartSelection   Part = "selection"
	PartPrefix      Part = "prefix"
	PartSuffix      Part = "suffix"
)

type BoxStyle struct {
	Width       *unit.Dp
	Height      *unit.Dp
	MinWidth    *unit.Dp
	MaxWidth    *unit.Dp
	MinHeight   *unit.Dp
	MaxHeight   *unit.Dp
	FillWidth   *bool
	FillHeight  *bool
	AspectRatio *float32
	Padding     *Insets
	Margin      *Insets
	Overflow    *Overflow
	Cursor      *pointer.Cursor

	paddingMask uint8
	marginMask  uint8
}

type Insets struct {
	Top, Right, Bottom, Left unit.Dp
}

type Overflow uint8

const (
	OverflowVisible Overflow = iota
	OverflowHidden
)

type PaintSource interface {
	isPaintSource()
}

// ColorSource is a solid color or a theme color token. The same value can be
// used by backgrounds, text, and borders.
type ColorSource interface {
	PaintSource
	isColorSource()
}

type SolidColor struct {
	Color color.NRGBA
}

func (SolidColor) isPaintSource() {}
func (SolidColor) isColorSource() {}

type ColorToken uint8

const (
	ColorBackground ColorToken = iota
	ColorSurface
	ColorSurfaceForeground
	ColorSurfaceSecondary
	ColorSurfaceSecondaryForeground
	ColorSurfaceTertiary
	ColorSurfaceTertiaryForeground
	ColorSurfaceHover
	ColorSurfacePressed
	ColorSurfaceRaised
	ColorOverlay
	ColorOverlayForeground
	ColorForeground
	ColorMutedForeground
	ColorBorder
	ColorSeparator
	ColorDefault
	ColorDefaultForeground
	ColorDefaultHover
	ColorFieldBackground
	ColorFieldHover
	ColorFieldForeground
	ColorFieldPlaceholder
	ColorFieldFocus
	ColorSegment
	ColorSegmentForeground
	ColorAccent
	ColorAccentHover
	ColorAccentPressed
	ColorAccentForeground
	ColorAccentSoft
	ColorAccentSoftHover
	ColorAccentSoftForeground
	ColorSuccess
	ColorSuccessForeground
	ColorSuccessSoft
	ColorSuccessSoftForeground
	ColorWarning
	ColorWarningForeground
	ColorWarningSoft
	ColorWarningSoftForeground
	ColorDanger
	ColorDangerHover
	ColorDangerPressed
	ColorDangerForeground
	ColorDangerSoft
	ColorDangerSoftHover
	ColorDangerSoftForeground
	ColorFocus
	ColorSelection
	ColorSurfaceShadow
	ColorOverlayShadow
)

// ThemeColor defers color lookup until the active theme is known.
type ThemeColor struct {
	Token ColorToken
}

func (ThemeColor) isPaintSource() {}
func (ThemeColor) isColorSource() {}

// AlphaColor replaces the alpha channel after its source is resolved. It
// keeps theme colors dynamic until the active theme is known.
type AlphaColor struct {
	Source ColorSource
	Alpha  uint8
}

func (AlphaColor) isPaintSource() {}
func (AlphaColor) isColorSource() {}

var (
	TokenBackground                 = ThemeColor{Token: ColorBackground}
	TokenSurface                    = ThemeColor{Token: ColorSurface}
	TokenSurfaceForeground          = ThemeColor{Token: ColorSurfaceForeground}
	TokenSurfaceSecondary           = ThemeColor{Token: ColorSurfaceSecondary}
	TokenSurfaceSecondaryForeground = ThemeColor{Token: ColorSurfaceSecondaryForeground}
	TokenSurfaceTertiary            = ThemeColor{Token: ColorSurfaceTertiary}
	TokenSurfaceTertiaryForeground  = ThemeColor{Token: ColorSurfaceTertiaryForeground}
	TokenSurfaceHover               = ThemeColor{Token: ColorSurfaceHover}
	TokenSurfacePressed             = ThemeColor{Token: ColorSurfacePressed}
	TokenSurfaceRaised              = ThemeColor{Token: ColorSurfaceRaised}
	TokenOverlay                    = ThemeColor{Token: ColorOverlay}
	TokenOverlayForeground          = ThemeColor{Token: ColorOverlayForeground}
	TokenForeground                 = ThemeColor{Token: ColorForeground}
	TokenMutedForeground            = ThemeColor{Token: ColorMutedForeground}
	TokenBorder                     = ThemeColor{Token: ColorBorder}
	TokenSeparator                  = ThemeColor{Token: ColorSeparator}
	TokenDefault                    = ThemeColor{Token: ColorDefault}
	TokenDefaultForeground          = ThemeColor{Token: ColorDefaultForeground}
	TokenDefaultHover               = ThemeColor{Token: ColorDefaultHover}
	TokenFieldBackground            = ThemeColor{Token: ColorFieldBackground}
	TokenFieldHover                 = ThemeColor{Token: ColorFieldHover}
	TokenFieldForeground            = ThemeColor{Token: ColorFieldForeground}
	TokenFieldPlaceholder           = ThemeColor{Token: ColorFieldPlaceholder}
	TokenFieldFocus                 = ThemeColor{Token: ColorFieldFocus}
	TokenSegment                    = ThemeColor{Token: ColorSegment}
	TokenSegmentForeground          = ThemeColor{Token: ColorSegmentForeground}
	TokenAccent                     = ThemeColor{Token: ColorAccent}
	TokenAccentHover                = ThemeColor{Token: ColorAccentHover}
	TokenAccentPressed              = ThemeColor{Token: ColorAccentPressed}
	TokenAccentForeground           = ThemeColor{Token: ColorAccentForeground}
	TokenAccentSoft                 = ThemeColor{Token: ColorAccentSoft}
	TokenAccentSoftHover            = ThemeColor{Token: ColorAccentSoftHover}
	TokenAccentSoftForeground       = ThemeColor{Token: ColorAccentSoftForeground}
	TokenSuccess                    = ThemeColor{Token: ColorSuccess}
	TokenSuccessForeground          = ThemeColor{Token: ColorSuccessForeground}
	TokenSuccessSoft                = ThemeColor{Token: ColorSuccessSoft}
	TokenSuccessSoftForeground      = ThemeColor{Token: ColorSuccessSoftForeground}
	TokenWarning                    = ThemeColor{Token: ColorWarning}
	TokenWarningForeground          = ThemeColor{Token: ColorWarningForeground}
	TokenWarningSoft                = ThemeColor{Token: ColorWarningSoft}
	TokenWarningSoftForeground      = ThemeColor{Token: ColorWarningSoftForeground}
	TokenDanger                     = ThemeColor{Token: ColorDanger}
	TokenDangerHover                = ThemeColor{Token: ColorDangerHover}
	TokenDangerPressed              = ThemeColor{Token: ColorDangerPressed}
	TokenDangerForeground           = ThemeColor{Token: ColorDangerForeground}
	TokenDangerSoft                 = ThemeColor{Token: ColorDangerSoft}
	TokenDangerSoftHover            = ThemeColor{Token: ColorDangerSoftHover}
	TokenDangerSoftForeground       = ThemeColor{Token: ColorDangerSoftForeground}
	TokenFocus                      = ThemeColor{Token: ColorFocus}
	TokenSelection                  = ThemeColor{Token: ColorSelection}
	TokenSurfaceShadow              = ThemeColor{Token: ColorSurfaceShadow}
	TokenOverlayShadow              = ThemeColor{Token: ColorOverlayShadow}
)

type StyleGradientStop struct {
	Offset float32
	Color  ColorSource
}

type StyleGradient struct {
	AngleDegrees float32
	Stops        []StyleGradientStop
}

func (StyleGradient) isPaintSource() {}

type PaintStyle struct {
	Background PaintSource
	Border     *BorderStyle
	Radius     *unit.Dp
	Radii      *CornerRadii
	Shadows    []Shadow
	Outline    *OutlineStyle
	Opacity    *float32

	radiusMask    uint8
	backgroundSet bool
	shadowsSet    bool
}

type CornerRadii struct {
	TopLeft, TopRight, BottomRight, BottomLeft unit.Dp
}

type Shadow struct {
	OffsetX, OffsetY unit.Dp
	Blur, Spread     unit.Dp
	Color            ColorSource
	Profile          *ShadowProfile
}

type ShadowProfile uint8

const (
	ShadowSurface ShadowProfile = iota
	ShadowOverlay
	ShadowMenu
)

type OutlineStyle struct {
	Width  unit.Dp
	Offset unit.Dp
	Color  ColorSource
}

type BorderStyle struct {
	Color ColorSource
	Width *unit.Dp
}

type TextStyle struct {
	Color           ColorSource
	FontSize        *unit.Sp
	FontWeight      *int
	Typeface        *font.Typeface
	FontStyle       *font.Style
	LineHeight      *unit.Sp
	LineHeightScale *float32
	MaxLines        *int
	Align           *TextAlign
	Wrap            *giotext.WrapPolicy
	Truncator       *string
}

type TextAlign uint8

const (
	TextAlignStart TextAlign = iota
	TextAlignCenter
	TextAlignEnd
)

type TransformStyle struct {
	TranslateX *unit.Dp
	TranslateY *unit.Dp
	ScaleX     *float32
	ScaleY     *float32
	Rotate     *float32
}

type PropertyID uint8

const (
	PropBackgroundColor PropertyID = iota
	PropTextColor
	PropBorderColor
	PropOutlineColor
	PropOpacity
	PropRadius
	PropTransform
)

type EaseFunc func(float32) float32

type Transition struct {
	Property PropertyID
	Duration time.Duration
	Delay    time.Duration
	Ease     EaseFunc
}

type TransitionOption func(*Transition)

type StyleState struct {
	Hovered       bool
	Pressed       bool
	Focused       bool
	FocusVisible  bool
	Disabled      bool
	Selected      bool
	Checked       bool
	Indeterminate bool
	ReadOnly      bool
	Invalid       bool
	Loading       bool
	Open          bool
	Expanded      bool
	Dragging      bool
	DropTarget    bool
}

type Condition func(StyleState) bool

func All(conditions ...Condition) Condition {
	return func(state StyleState) bool {
		for _, condition := range conditions {
			if condition == nil || !condition(state) {
				return false
			}
		}
		return true
	}
}

func Any(conditions ...Condition) Condition {
	return func(state StyleState) bool {
		for _, condition := range conditions {
			if condition != nil && condition(state) {
				return true
			}
		}
		return false
	}
}

func Not(condition Condition) Condition {
	if condition == nil {
		return nil
	}
	return func(state StyleState) bool { return !condition(state) }
}

// If adapts an application model value to a style condition. The declaration
// should be rebuilt by the MVU View when the value changes.
func If(value bool) Condition {
	return func(StyleState) bool { return value }
}

var (
	Hovered       Condition = func(state StyleState) bool { return state.Hovered }
	Pressed       Condition = func(state StyleState) bool { return state.Pressed }
	Focused       Condition = func(state StyleState) bool { return state.Focused }
	FocusVisible  Condition = func(state StyleState) bool { return state.FocusVisible }
	Disabled      Condition = func(state StyleState) bool { return state.Disabled }
	Selected      Condition = func(state StyleState) bool { return state.Selected }
	Checked       Condition = func(state StyleState) bool { return state.Checked }
	Indeterminate Condition = func(state StyleState) bool { return state.Indeterminate }
	ReadOnly      Condition = func(state StyleState) bool { return state.ReadOnly }
	Invalid       Condition = func(state StyleState) bool { return state.Invalid }
	Loading       Condition = func(state StyleState) bool { return state.Loading }
	Open          Condition = func(state StyleState) bool { return state.Open }
	ExpandedState Condition = func(state StyleState) bool { return state.Expanded }
	Dragging      Condition = func(state StyleState) bool { return state.Dragging }
	DropTarget    Condition = func(state StyleState) bool { return state.DropTarget }
)

type condition struct {
	predicate Condition
	override  Style
}
