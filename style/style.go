package style

import (
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

	// rules are variant table rows. They hold no closures, so they can be
	// ordered and hashed without evaluating anything.
	rules []Rule

	// dim scales the alpha of every color this declaration resolves to. It is a
	// scalar rather than a closure so the table above stays hashable, and it is
	// applied once at the resolution boundary, not per rule row, so a row that
	// asks for it dims the merged result rather than only its own fields.
	dim *float32

	parts map[Part]Style
}

// ResolvedStyle is a computed snapshot used for layout and drawing.
//
// Resolvers may return cache-owned data when the snapshot has no transitions;
// such values must be treated as read-only. Use the runtime CloneResolvedStyle
// helper before mutating a snapshot whose ownership is not otherwise known.
type ResolvedStyle struct {
	Box         *BoxStyle
	Paint       *PaintStyle
	Text        *TextStyle
	Trans       *TransformStyle
	Transitions []Transition
}

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

// ColorSource is a concrete color value. The same value can be used by
// backgrounds, text, and borders.
type ColorSource interface {
	PaintSource
	isColorSource()
}

type SolidColor struct {
	Color color.NRGBA
}

func (SolidColor) isPaintSource() {}
func (SolidColor) isColorSource() {}

type GradientStop struct {
	Offset float32
	Color  ColorSource
}

type Gradient struct {
	AngleDegrees float32
	Stops        []GradientStop
}

func (Gradient) isPaintSource() {}

type PaintStyle struct {
	Background PaintSource
	Border     *BorderStyle
	// BorderBottom is an optional bottom-only border. It is kept separate from
	// Border so a component can underline a control without painting the other
	// three sides.
	BorderBottom *BorderStyle
	Radius       *unit.Dp
	Radii        *CornerRadii
	Shadows      []Shadow
	Outline      *OutlineStyle
	// VisualOutset is an optional minimum visual overflow reservation. It is
	// combined with the shadow and outline extents when a clipping container
	// lays out this box.
	VisualOutset *Insets
	Opacity      *float32

	radiusMask      uint8
	backgroundSet   bool
	shadowsSet      bool
	visualOutsetSet bool
}

type CornerRadii struct {
	TopLeft, TopRight, BottomRight, BottomLeft unit.Dp
}

type Shadow struct {
	OffsetX, OffsetY unit.Dp
	Blur, Spread     unit.Dp
	Color            ColorSource
	Profile          *ShadowProfile
	// Transient shadows are painted but do not reserve visual overflow in
	// clipping containers. Use this for focus and validation decoration that
	// would otherwise change viewport layout when the state changes.
	Transient bool
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
	// LineStyle selects the stroke pattern. Nil and LineSolid both paint a
	// solid border; the pointer keeps "unset" distinguishable from an
	// explicit choice, as every other field in this package does.
	LineStyle *LineStyle
}

// LineStyle selects a stroke pattern for borders.
type LineStyle uint8

const (
	LineSolid LineStyle = iota
	LineDashed
	LineDotted
)

// DashPattern returns the on/off lengths in Dp for a patterned LineStyle.
// A solid line returns zero, which lets callers skip the segment split.
func (l LineStyle) DashPattern() (on, off unit.Dp) {
	switch l {
	case LineDashed:
		return 4, 3
	case LineDotted:
		return 1, 2
	default:
		return 0, 0
	}
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
	// Underline draws a rule under each text line. It is a style property
	// rather than a widget option because link buttons differ from text
	// buttons by this alone.
	Underline *bool
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

// EaseFunc maps normalized time to animation progress.
type EaseFunc func(float32) float32

type Transition struct {
	Property PropertyID
	Duration time.Duration
	Delay    time.Duration
	Ease     EaseFunc
}

type TransitionOption func(*Transition)

type State struct {
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
	Warning       bool
}

// CacheUnsafe reports whether resolving this declaration depends on values
// that cannot be represented reliably in a cache key. The surviving check is
// custom easing functions: they are function values, so a transition carrying
// one must bypass caches keyed on Hash64. State rule rows contribute no
// closures of their own; a declaration built purely from Rules stays
// cache-safe even when it has many rows. The predicate-condition slot this
// method once also inspected was removed with the retired When mechanism.
func (s Style) CacheUnsafe() bool {
	for _, transition := range s.transitions {
		if transition.Ease != nil {
			return true
		}
	}
	// A state rule is only as cacheable as the increment it carries.
	for _, rule := range s.rules {
		if rule.Set.CacheUnsafe() {
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

// IsEmpty reports whether this declaration contains no style values, rules,
// transitions, or part declarations.
func (s Style) IsEmpty() bool {
	return s.box == nil && s.paint == nil && s.text == nil && s.trans == nil &&
		len(s.transitions) == 0 && len(s.rules) == 0 && s.dim == nil && len(s.parts) == 0
}

// Hash64 computes a structural hash of this style declaration for caching.
// Two styles with identical field values will produce the same hash.
func (s Style) Hash64() uint64 {
	return s.hash64(nil)
}

// Hash64ForState computes the declaration hash for one interaction state.
// Condition results are included because they affect the resolved declaration.
func (s Style) Hash64ForState(state State) uint64 {
	return s.hash64(&state)
}

func (s Style) hash64(state *State) uint64 {
	h := styleHasher{sum: fnvOffset64}

	if s.box != nil {
		writeUint8(&h, 1)
		hashBoxStyle(&h, s.box)
	}
	if s.paint != nil {
		writeUint8(&h, 2)
		hashPaintStyle(&h, s.paint)
	}
	if s.text != nil {
		writeUint8(&h, 3)
		hashTextStyle(&h, s.text)
	}
	if s.trans != nil {
		writeUint8(&h, 4)
		hashTransformStyle(&h, s.trans)
	}

	writeUint8(&h, 5)
	writeInt(&h, len(s.transitions))
	for _, t := range s.transitions {
		writeUint8(&h, uint8(t.Property))
		writeInt64(&h, int64(t.Duration))
		writeInt64(&h, int64(t.Delay))
		if t.Ease != nil {
			writeUint64(&h, uint64(reflect.ValueOf(t.Ease).Pointer()))
		}
	}

	writeUint8(&h, 8)
	writeInt(&h, len(s.parts))
	if len(s.parts) > 0 {
		// Compound declarations usually have only a handful of Parts. Keep
		// that common path on the stack; larger declarations retain the
		// allocation-backed deterministic sort.
		const stackPartLimit = 8
		var stackParts [stackPartLimit]string
		parts := stackParts[:0]
		if len(s.parts) > stackPartLimit {
			parts = make([]string, 0, len(s.parts))
		}
		for part := range s.parts {
			parts = append(parts, string(part))
		}
		if len(parts) <= stackPartLimit {
			for i := 1; i < len(parts); i++ {
				value := parts[i]
				j := i
				for j > 0 && parts[j-1] > value {
					parts[j] = parts[j-1]
					j--
				}
				parts[j] = value
			}
		} else {
			sort.Strings(parts)
		}
		for _, part := range parts {
			writeString(&h, part)
			writeUint64(&h, s.parts[Part(part)].hash64(state))
		}
	}

	writeUint8(&h, 9)
	writeInt(&h, len(s.rules))
	for _, rule := range s.rules {
		writeInt(&h, int(rule.Rank))
		writeUint16(&h, uint16(rule.When))
		writeUint16(&h, uint16(rule.Not))
		if rule.Tween == nil {
			writeBool(&h, false)
		} else {
			writeBool(&h, true)
			writeUint8(&h, uint8(rule.Tween.Property))
			writeInt64(&h, int64(rule.Tween.Duration))
			writeInt64(&h, int64(rule.Tween.Delay))
		}
		writeUint64(&h, rule.Set.hash64(state))
	}

	if s.dim != nil {
		writeUint8(&h, 10)
		writeFloat32(&h, *s.dim)
	}

	return h.sum
}

const (
	fnvOffset64 = uint64(14695981039346656037)
	fnvPrime64  = uint64(1099511628211)
)

// styleHasher is the byte-oriented FNV-1a implementation used by Style.Hash64.
// Keeping the hasher concrete avoids interface dispatch and temporary slices in
// this hot path while preserving the hash.Hash/fnv.New64a byte order.
type styleHasher struct {
	sum uint64
}

func (h *styleHasher) writeByte(value byte) {
	h.sum ^= uint64(value)
	h.sum *= fnvPrime64
}

func hashBoxStyle(h *styleHasher, b *BoxStyle) {
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

func hashInsets(h *styleHasher, i *Insets) {
	writeFloat32(h, float32(i.Top))
	writeFloat32(h, float32(i.Right))
	writeFloat32(h, float32(i.Bottom))
	writeFloat32(h, float32(i.Left))
}

func hashPaintStyle(h *styleHasher, p *PaintStyle) {
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
		if p.Border.LineStyle != nil {
			writeUint8(h, 2)
			writeUint8(h, uint8(*p.Border.LineStyle))
		}
	}
	if p.BorderBottom != nil {
		writeUint8(h, 11)
		hashColorSource(h, p.BorderBottom.Color)
		if p.BorderBottom.Width != nil {
			writeUint8(h, 1)
			writeFloat32(h, float32(*p.BorderBottom.Width))
		}
		if p.BorderBottom.LineStyle != nil {
			writeUint8(h, 2)
			writeUint8(h, uint8(*p.BorderBottom.LineStyle))
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
		if shadow.Transient {
			writeUint8(h, 9)
		}
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
	if p.VisualOutset != nil {
		writeUint8(h, 12)
		hashInsets(h, p.VisualOutset)
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
	writeUint8(h, 13)
	writeBool(h, p.visualOutsetSet)
}

func hashPaintSource(h *styleHasher, ps PaintSource) {
	switch v := ps.(type) {
	case SolidColor:
		writeUint8(h, 1)
		hashColor(h, v.Color)
	case Gradient:
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

func hashColorSource(h *styleHasher, cs ColorSource) {
	switch v := cs.(type) {
	case SolidColor:
		writeUint8(h, 1)
		hashColor(h, v.Color)
	default:
		writeUint8(h, 0)
	}
}

func hashColor(h *styleHasher, c color.NRGBA) {
	writeUint8(h, c.R)
	writeUint8(h, c.G)
	writeUint8(h, c.B)
	writeUint8(h, c.A)
}

func hashTextStyle(h *styleHasher, t *TextStyle) {
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
	if t.Underline != nil {
		writeUint8(h, 12)
		writeBool(h, *t.Underline)
	}
}

func hashTransformStyle(h *styleHasher, t *TransformStyle) {
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

func writeUint8(h *styleHasher, v uint8) {
	h.writeByte(v)
}

func writeUint16(h *styleHasher, v uint16) {
	h.writeByte(byte(v))
	h.writeByte(byte(v >> 8))
}

func writeInt(h *styleHasher, v int) {
	writeInt64(h, int64(v))
}

func writeInt64(h *styleHasher, v int64) {
	writeUint64(h, uint64(v))
}

func writeUint64(h *styleHasher, v uint64) {
	h.writeByte(byte(v))
	h.writeByte(byte(v >> 8))
	h.writeByte(byte(v >> 16))
	h.writeByte(byte(v >> 24))
	h.writeByte(byte(v >> 32))
	h.writeByte(byte(v >> 40))
	h.writeByte(byte(v >> 48))
	h.writeByte(byte(v >> 56))
}

func writeFloat32(h *styleHasher, v float32) {
	bits := math.Float32bits(v)
	writeUint32(h, bits)
}

func writeUint32(h *styleHasher, v uint32) {
	h.writeByte(byte(v))
	h.writeByte(byte(v >> 8))
	h.writeByte(byte(v >> 16))
	h.writeByte(byte(v >> 24))
}

func writeBool(h *styleHasher, v bool) {
	if v {
		writeUint8(h, 1)
		return
	}
	writeUint8(h, 0)
}

func writeString(h *styleHasher, s string) {
	for index := 0; index < len(s); index++ {
		h.writeByte(s[index])
	}
}

// New creates an empty style declaration for fluent configuration.
//
// Style values are immutable; each fluent method returns a configured copy.
func New() Style {
	return Style{}
}
