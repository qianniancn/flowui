package style

import (
	"hash"
	"hash/fnv"
	"image/color"
	"math"
	"reflect"
	"sort"
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

// PropertyID identifies a Style property that may participate in Transition.
type PropertyID uint8

const (
	PropBackgroundColor PropertyID = iota
	PropTextColor
	PropBorderColor
	PropOutlineColor
	PropOpacity
	PropRadius
	PropTransform
	propertyIDCount
)

// Animatable reports whether property is in the Transition whitelist.
func (p PropertyID) Animatable() bool {
	return p < propertyIDCount
}

// EaseFunc maps normalized time to animation progress (same contract as ui.Easing).
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

// unsafeCondition wraps a condition that captures external state not in StyleState.
// Such conditions cannot be safely cached because their behavior depends on values
// we cannot hash.
type unsafeCondition struct {
	fn Condition
}

func (u unsafeCondition) eval(state StyleState) bool {
	if u.fn == nil {
		return false
	}
	return u.fn(state)
}

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
	unsafe    bool // true if predicate captures external state (e.g., from If())
}

// HasConditions reports whether this style uses conditional styling that
// depends on runtime state not captured by StyleState (e.g., If(value)).
// Styles with unsafe conditions should bypass caching because their predicates
// capture external values that cannot be hashed.
func (s Style) HasConditions() bool {
	// Check direct conditions
	for _, cond := range s.conditions {
		if cond.unsafe {
			return true
		}
		// Recursively check the override style
		if cond.override.HasConditions() {
			return true
		}
	}
	// Recursively check parts
	for _, partStyle := range s.parts {
		if partStyle.HasConditions() {
			return true
		}
	}
	return false
}

// CacheUnsafe reports whether resolving this declaration depends on values
// that cannot be represented reliably in a cache key.
func (s Style) CacheUnsafe() bool {
	if s.HasConditions() {
		return true
	}
	for _, transition := range s.transitions {
		if transition.Ease != nil {
			return true
		}
	}
	for _, condition := range s.conditions {
		if condition.override.CacheUnsafe() {
			return true
		}
	}
	for _, partStyle := range s.parts {
		if partStyle.CacheUnsafe() {
			return true
		}
	}
	return false
}

// Hash64 computes a structural hash of this style declaration for caching.
// Two styles with identical field values will produce the same hash.
func (s Style) Hash64() uint64 {
	return s.hash64(nil)
}

// Hash64ForState computes the declaration hash for one interaction state.
// Condition results are included because they affect the resolved declaration.
func (s Style) Hash64ForState(state StyleState) uint64 {
	return s.hash64(&state)
}

func (s Style) hash64(state *StyleState) uint64 {
	h := fnv.New64a()

	if s.box != nil {
		writeUint8(h, 1)
		hashBoxStyle(h, s.box)
	}
	if s.paint != nil {
		writeUint8(h, 2)
		hashPaintStyle(h, s.paint)
	}
	if s.text != nil {
		writeUint8(h, 3)
		hashTextStyle(h, s.text)
	}
	if s.trans != nil {
		writeUint8(h, 4)
		hashTransformStyle(h, s.trans)
	}

	writeUint8(h, 5)
	writeInt(h, len(s.transitions))
	for _, t := range s.transitions {
		writeUint8(h, uint8(t.Property))
		writeInt64(h, int64(t.Duration))
		writeInt64(h, int64(t.Delay))
		if t.Ease != nil {
			writeUint64(h, uint64(reflect.ValueOf(t.Ease).Pointer()))
		}
	}

	writeUint8(h, 6)
	writeInt(h, len(s.tokens))
	for _, token := range s.tokens {
		writeUint8(h, uint8(token))
	}

	writeUint8(h, 7)
	writeInt(h, len(s.conditions))
	for _, cond := range s.conditions {
		writeBool(h, cond.unsafe)
		if state != nil {
			writeBool(h, cond.predicate != nil && cond.predicate(*state))
		} else if cond.predicate != nil {
			writeUint64(h, uint64(reflect.ValueOf(cond.predicate).Pointer()))
		}
		h.Write(uint64ToBytes(cond.override.hash64(state)))
	}

	writeUint8(h, 8)
	writeInt(h, len(s.parts))
	parts := make([]string, 0, len(s.parts))
	for part := range s.parts {
		parts = append(parts, string(part))
	}
	sort.Strings(parts)
	for _, part := range parts {
		writeString(h, part)
		h.Write(uint64ToBytes(s.parts[Part(part)].hash64(state)))
	}

	return h.Sum64()
}

func uint64ToBytes(v uint64) []byte {
	return []byte{
		byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24),
		byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56),
	}
}

func hashBoxStyle(h hash.Hash, b *BoxStyle) {
	if b.Width != nil {
		writeUint8(h, 1)
		writeFloat32(h, float32(*b.Width))
	}
	if b.Height != nil {
		writeUint8(h, 2)
		writeFloat32(h, float32(*b.Height))
	}
	if b.MinWidth != nil {
		writeUint8(h, 3)
		writeFloat32(h, float32(*b.MinWidth))
	}
	if b.MaxWidth != nil {
		writeUint8(h, 4)
		writeFloat32(h, float32(*b.MaxWidth))
	}
	if b.MinHeight != nil {
		writeUint8(h, 5)
		writeFloat32(h, float32(*b.MinHeight))
	}
	if b.MaxHeight != nil {
		writeUint8(h, 6)
		writeFloat32(h, float32(*b.MaxHeight))
	}
	if b.FillWidth != nil {
		writeUint8(h, 7)
		writeBool(h, *b.FillWidth)
	}
	if b.FillHeight != nil {
		writeUint8(h, 8)
		writeBool(h, *b.FillHeight)
	}
	if b.AspectRatio != nil {
		writeUint8(h, 9)
		writeFloat32(h, *b.AspectRatio)
	}
	if b.Padding != nil {
		writeUint8(h, 10)
		hashInsets(h, b.Padding)
	}
	if b.Margin != nil {
		writeUint8(h, 11)
		hashInsets(h, b.Margin)
	}
	if b.Overflow != nil {
		writeUint8(h, 12)
		writeUint8(h, uint8(*b.Overflow))
	}
	if b.Cursor != nil {
		writeUint8(h, 13)
		writeUint8(h, uint8(*b.Cursor))
	}
	writeUint8(h, 14)
	writeUint8(h, b.paddingMask)
	writeUint8(h, 15)
	writeUint8(h, b.marginMask)
}

func hashInsets(h hash.Hash, i *Insets) {
	writeFloat32(h, float32(i.Top))
	writeFloat32(h, float32(i.Right))
	writeFloat32(h, float32(i.Bottom))
	writeFloat32(h, float32(i.Left))
}

func hashPaintStyle(h hash.Hash, p *PaintStyle) {
	if p.Background != nil {
		writeUint8(h, 1)
		hashPaintSource(h, p.Background)
	}
	if p.Border != nil {
		writeUint8(h, 2)
		hashColorSource(h, p.Border.Color)
		if p.Border.Width != nil {
			writeUint8(h, 1)
			writeFloat32(h, float32(*p.Border.Width))
		}
	}
	if p.Radius != nil {
		writeUint8(h, 3)
		writeFloat32(h, float32(*p.Radius))
	}
	if p.Radii != nil {
		writeUint8(h, 4)
		writeFloat32(h, float32(p.Radii.TopLeft))
		writeFloat32(h, float32(p.Radii.TopRight))
		writeFloat32(h, float32(p.Radii.BottomRight))
		writeFloat32(h, float32(p.Radii.BottomLeft))
	}
	writeUint8(h, 5)
	writeInt(h, len(p.Shadows))
	for _, shadow := range p.Shadows {
		writeFloat32(h, float32(shadow.OffsetX))
		writeFloat32(h, float32(shadow.OffsetY))
		writeFloat32(h, float32(shadow.Blur))
		writeFloat32(h, float32(shadow.Spread))
		hashColorSource(h, shadow.Color)
		if shadow.Profile != nil {
			writeUint8(h, 1)
			writeUint8(h, uint8(*shadow.Profile))
		}
	}
	if p.Outline != nil {
		writeUint8(h, 6)
		writeFloat32(h, float32(p.Outline.Width))
		writeFloat32(h, float32(p.Outline.Offset))
		hashColorSource(h, p.Outline.Color)
	}
	if p.Opacity != nil {
		writeUint8(h, 7)
		writeFloat32(h, *p.Opacity)
	}
	writeUint8(h, 8)
	writeUint8(h, p.radiusMask)
	writeUint8(h, 9)
	writeBool(h, p.backgroundSet)
	writeUint8(h, 10)
	writeBool(h, p.shadowsSet)
}

func hashPaintSource(h hash.Hash, ps PaintSource) {
	switch v := ps.(type) {
	case SolidColor:
		writeUint8(h, 1)
		hashColor(h, v.Color)
	case ThemeColor:
		writeUint8(h, 2)
		writeUint8(h, uint8(v.Token))
	case AlphaColor:
		writeUint8(h, 3)
		hashColorSource(h, v.Source)
		writeUint8(h, v.Alpha)
	case StyleGradient:
		writeUint8(h, 4)
		writeFloat32(h, v.AngleDegrees)
		writeInt(h, len(v.Stops))
		for _, stop := range v.Stops {
			writeFloat32(h, stop.Offset)
			hashColorSource(h, stop.Color)
		}
	default:
		writeUint8(h, 0)
	}
}

func hashColorSource(h hash.Hash, cs ColorSource) {
	switch v := cs.(type) {
	case SolidColor:
		writeUint8(h, 1)
		hashColor(h, v.Color)
	case ThemeColor:
		writeUint8(h, 2)
		writeUint8(h, uint8(v.Token))
	case AlphaColor:
		writeUint8(h, 3)
		hashColorSource(h, v.Source)
		writeUint8(h, v.Alpha)
	default:
		writeUint8(h, 0)
	}
}

func hashColor(h hash.Hash, c color.NRGBA) {
	writeUint8(h, c.R)
	writeUint8(h, c.G)
	writeUint8(h, c.B)
	writeUint8(h, c.A)
}

func hashTextStyle(h hash.Hash, t *TextStyle) {
	if t.Color != nil {
		writeUint8(h, 1)
		hashColorSource(h, t.Color)
	}
	if t.FontSize != nil {
		writeUint8(h, 2)
		writeFloat32(h, float32(*t.FontSize))
	}
	if t.FontWeight != nil {
		writeUint8(h, 3)
		writeInt(h, *t.FontWeight)
	}
	if t.Typeface != nil {
		writeUint8(h, 4)
		writeString(h, string(*t.Typeface))
	}
	if t.FontStyle != nil {
		writeUint8(h, 5)
		writeInt(h, int(*t.FontStyle))
	}
	if t.LineHeight != nil {
		writeUint8(h, 6)
		writeFloat32(h, float32(*t.LineHeight))
	}
	if t.LineHeightScale != nil {
		writeUint8(h, 7)
		writeFloat32(h, *t.LineHeightScale)
	}
	if t.MaxLines != nil {
		writeUint8(h, 8)
		writeInt(h, *t.MaxLines)
	}
	if t.Align != nil {
		writeUint8(h, 9)
		writeUint8(h, uint8(*t.Align))
	}
	if t.Wrap != nil {
		writeUint8(h, 10)
		writeUint8(h, uint8(*t.Wrap))
	}
	if t.Truncator != nil {
		writeUint8(h, 11)
		writeString(h, *t.Truncator)
	}
}

func hashTransformStyle(h hash.Hash, t *TransformStyle) {
	if t.TranslateX != nil {
		writeUint8(h, 1)
		writeFloat32(h, float32(*t.TranslateX))
	}
	if t.TranslateY != nil {
		writeUint8(h, 2)
		writeFloat32(h, float32(*t.TranslateY))
	}
	if t.ScaleX != nil {
		writeUint8(h, 3)
		writeFloat32(h, *t.ScaleX)
	}
	if t.ScaleY != nil {
		writeUint8(h, 4)
		writeFloat32(h, *t.ScaleY)
	}
	if t.Rotate != nil {
		writeUint8(h, 5)
		writeFloat32(h, *t.Rotate)
	}
}

// Hash helper functions

func writeUint8(h hash.Hash, v uint8) {
	h.Write([]byte{v})
}

func writeUint16(h hash.Hash, v uint16) {
	h.Write([]byte{byte(v), byte(v >> 8)})
}

func writeInt(h hash.Hash, v int) {
	writeInt64(h, int64(v))
}

func writeInt64(h hash.Hash, v int64) {
	writeUint64(h, uint64(v))
}

func writeUint64(h hash.Hash, v uint64) {
	h.Write([]byte{
		byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24),
		byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56),
	})
}

func writeFloat32(h hash.Hash, v float32) {
	bits := math.Float32bits(v)
	h.Write([]byte{byte(bits), byte(bits >> 8), byte(bits >> 16), byte(bits >> 24)})
}

func writeBool(h hash.Hash, v bool) {
	if v {
		h.Write([]byte{1})
	} else {
		h.Write([]byte{0})
	}
}

func writeString(h hash.Hash, s string) {
	h.Write([]byte(s))
}
