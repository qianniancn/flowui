package style

import "maps"

func (s Style) resolve(state State) Style {
	result := clone(s)
	result.rules = nil
	// State rule rows first: they are a table, so their precedence comes from
	// Rule.Rank and matched-bit count rather than from nesting depth.
	for _, rule := range orderedRules(s.rules, state.Bits()) {
		override := rule.Set.resolve(state)
		if rule.Tween != nil {
			override = override.Transition(
				rule.Tween.Property,
				rule.Tween.Duration,
				TransitionDelay(rule.Tween.Delay),
			)
		}
		result = merge(result, override)
	}
	return result
}

// Resolve applies this declaration's matching conditional rules.
func (s Style) Resolve(state State) ResolvedStyle {
	return resolvedStyle(s.resolve(state))
}

func Cascade(state State, layers ...Style) ResolvedStyle {
	var result Style
	for _, layer := range layers {
		result = merge(result, layer.resolve(state))
	}
	return resolvedStyle(result)
}

// Join combines declarations without resolving conditions.
// It is used internally when component defaults are assembled in stages.
func Join(styles ...Style) Style {
	var result Style
	for _, declaration := range styles {
		result = merge(result, declaration)
	}
	return result
}

// JoinPart combines a base declaration with an override that only targets one
// named part. Unlike Join, it keeps unrelated part declarations shared and
// clones only the selected part. This is useful for per-cell recipes in
// compound widgets where copying every part would dominate frame allocations.
func JoinPart(base, override Style, part Part) Style {
	if len(override.parts) == 0 {
		return base
	}
	overridePart, ok := override.parts[part]
	if !ok {
		return base
	}
	result := base
	result.parts = make(map[Part]Style, len(base.parts))
	maps.Copy(result.parts, base.parts)
	result.parts[part] = merge(base.parts[part], overridePart)
	return result
}

// CascadePart resolves a named component part. Text properties inherit from
// each root layer; box, paint, and transform properties do not.
func CascadePart(state State, part Part, layers ...Style) ResolvedStyle {
	if part == PartRoot {
		return Cascade(state, layers...)
	}
	var result Style
	for _, layer := range layers {
		resolved := layer.resolve(state)
		if resolved.text != nil {
			inherited := Style{text: cloneText(resolved.text)}
			// Pre-allocate for text color transitions if any exist
			for _, transition := range resolved.transitions {
				if transition.Property == PropTextColor {
					if inherited.transitions == nil {
						inherited.transitions = make([]Transition, 0, 1)
					}
					inherited.transitions = append(inherited.transitions, transition)
				}
			}
			result = merge(result, inherited)
		}
		if declaration, ok := resolved.parts[part]; ok {
			result = merge(result, declaration.resolve(state))
		}
	}
	result.parts = nil
	return resolvedStyle(result)
}

// TextDeclaration copies resolved text properties into an inheritable
// declaration without exposing Style internals.
func TextDeclaration(value *TextStyle) Style {
	return Style{text: cloneText(value)}
}

func resolvedStyle(source Style) ResolvedStyle {
	// Pre-allocate transitions slice with exact capacity
	var transitions []Transition
	if len(source.transitions) > 0 {
		transitions = make([]Transition, len(source.transitions))
		copy(transitions, source.transitions)
	}
	result := ResolvedStyle{
		Box:         cloneBox(source.box),
		Paint:       clonePaint(source.paint),
		Text:        cloneText(source.text),
		Trans:       cloneTransform(source.trans),
		Transitions: transitions,
	}
	// Dim is applied here rather than inside resolve because this is the
	// resolution boundary: resolve also runs for a rule row's Set, and a row
	// that asks to dim must dim the merged result, not only its own fields.
	if source.dim != nil {
		dimResolved(&result, *source.dim)
	}
	return result
}

// dimResolved scales the alpha of every color in a resolved snapshot. Shadow
// layers are skipped on purpose -- see Style.Dim.
func dimResolved(resolved *ResolvedStyle, factor float32) {
	if paint := resolved.Paint; paint != nil {
		paint.Background = dimPaintSource(paint.Background, factor)
		if paint.Border != nil {
			paint.Border.Color = dimColorSource(paint.Border.Color, factor)
		}
		if paint.BorderBottom != nil {
			paint.BorderBottom.Color = dimColorSource(paint.BorderBottom.Color, factor)
		}
		if paint.Outline != nil {
			paint.Outline.Color = dimColorSource(paint.Outline.Color, factor)
		}
	}
	if text := resolved.Text; text != nil {
		text.Color = dimColorSource(text.Color, factor)
	}
}

func dimPaintSource(source PaintSource, factor float32) PaintSource {
	switch value := source.(type) {
	case nil:
		return nil
	case Gradient:
		value.Stops = dimGradientStops(value.Stops, factor)
		return value
	case *Gradient:
		if value == nil {
			return nil
		}
		copied := *value
		copied.Stops = dimGradientStops(value.Stops, factor)
		return &copied
	case ColorSource:
		return dimColorSource(value, factor)
	default:
		return source
	}
}

func dimGradientStops(stops []GradientStop, factor float32) []GradientStop {
	if len(stops) == 0 {
		return stops
	}
	result := make([]GradientStop, len(stops))
	for index, stop := range stops {
		result[index] = GradientStop{Offset: stop.Offset, Color: dimColorSource(stop.Color, factor)}
	}
	return result
}

// dimColorSource multiplies a color source's alpha.
//
// A concrete color is scaled in place and stays concrete. That is not an
// optimization: it is what makes Dim(DisabledOpacity) on a solid color
// indistinguishable from the DisabledColor arithmetic the built-in components
// already perform, so the two can be compared field by field.
//
// ColorSource has exactly one implementation (SolidColor); the default branch
// is unreachable and only satisfies the compiler.
func dimColorSource(source ColorSource, factor float32) ColorSource {
	switch value := source.(type) {
	case nil:
		return nil
	case SolidColor:
		value.Color.A = ScaleAlpha(value.Color.A, factor)
		return value
	case *SolidColor:
		if value == nil {
			return nil
		}
		copied := *value
		copied.Color.A = ScaleAlpha(copied.Color.A, factor)
		return &copied
	default:
		return source
	}
}

func merge(first, second Style) Style {
	// Fast path: second is empty
	if second.box == nil && second.paint == nil && second.text == nil && second.trans == nil &&
		len(second.transitions) == 0 && len(second.parts) == 0 &&
		len(second.rules) == 0 && second.dim == nil {
		return clone(first)
	}

	result := first

	// One slot-merge function per storage slot: box, paint, text, transform,
	// transitions. Non-slot members follow below.
	result.box = mergeBoxSlot(first.box, second.box)
	result.paint = mergePaintSlot(first.paint, second.paint)
	result.text = mergeTextSlot(first.text, second.text)
	result.trans = mergeTransSlot(first.trans, second.trans)
	result.transitions = mergeMotionSlot(first.transitions, second.transitions)

	if len(second.rules) != 0 {
		// Concatenation, not precedence merging: a later layer's rules must be
		// able to override an earlier layer's rules of equal specificity, which
		// is exactly what the declaration-order tiebreak in orderedRules does.
		result.rules = make([]Rule, len(first.rules), len(first.rules)+len(second.rules))
		copy(result.rules, first.rules)
		result.rules = append(result.rules, second.rules...)
	}
	if len(second.parts) != 0 {
		result.parts = mergeParts(first.parts, second.parts)
	}
	if second.dim != nil {
		// A later layer's dim replaces rather than composes. Two multipliers
		// would be a product, and the declaration that says "dim by a half"
		// should mean a half, not "a half of whatever the layer below asked for".
		result.dim = clonePtr(second.dim)
	}
	return result
}

// The five slot-merge functions below share one shape: a nil incoming side
// keeps the current slot untouched, an empty current slot adopts a clone of
// the incoming side, and otherwise the merged result is a fresh clone so the
// two input declarations are never aliased.

func mergeBoxSlot(first, second *BoxStyle) *BoxStyle {
	switch {
	case second == nil:
		return first
	case first == nil:
		return cloneBox(second)
	}
	merged := cloneBox(first)
	mergeBox(merged, second)
	return merged
}

func mergePaintSlot(first, second *PaintStyle) *PaintStyle {
	switch {
	case second == nil:
		return first
	case first == nil:
		return clonePaint(second)
	}
	merged := clonePaint(first)
	mergePaint(merged, second)
	return merged
}

func mergeTextSlot(first, second *TextStyle) *TextStyle {
	switch {
	case second == nil:
		return first
	case first == nil:
		return cloneText(second)
	}
	merged := cloneText(first)
	mergeText(merged, second)
	return merged
}

func mergeTransSlot(first, second *TransformStyle) *TransformStyle {
	switch {
	case second == nil:
		return first
	case first == nil:
		return cloneTransform(second)
	}
	merged := cloneTransform(first)
	mergeTransform(merged, second)
	return merged
}

func mergeMotionSlot(first, second []Transition) []Transition {
	if len(second) == 0 {
		return first
	}
	return mergeTransitions(first, second)
}

func mergeParts(first, second map[Part]Style) map[Part]Style {
	if len(first) == 0 && len(second) == 0 {
		return nil
	}
	result := cloneParts(first)
	if result == nil {
		result = make(map[Part]Style, len(second))
	}
	for part, declaration := range second {
		result[part] = merge(result[part], declaration)
	}
	return result
}

func mergeBox(dst, src *BoxStyle) {
	dst.Width = pick(dst.Width, src.Width)
	dst.Height = pick(dst.Height, src.Height)
	dst.MinWidth = pick(dst.MinWidth, src.MinWidth)
	dst.MaxWidth = pick(dst.MaxWidth, src.MaxWidth)
	dst.MinHeight = pick(dst.MinHeight, src.MinHeight)
	dst.MaxHeight = pick(dst.MaxHeight, src.MaxHeight)
	dst.FillWidth = pick(dst.FillWidth, src.FillWidth)
	dst.FillHeight = pick(dst.FillHeight, src.FillHeight)
	dst.AspectRatio = pick(dst.AspectRatio, src.AspectRatio)
	mergeInsets(&dst.Padding, &dst.paddingMask, src.Padding, src.paddingMask)
	mergeInsets(&dst.Margin, &dst.marginMask, src.Margin, src.marginMask)
	dst.Overflow = pick(dst.Overflow, src.Overflow)
	dst.Cursor = pick(dst.Cursor, src.Cursor)
}

func mergeInsets(dst **Insets, dstMask *uint8, src *Insets, srcMask uint8) {
	if src == nil {
		return
	}
	if *dst == nil {
		*dst = &Insets{}
	}
	if srcMask == 0 {
		srcMask = sidesAll
	}
	if srcMask&sideTop != 0 {
		(*dst).Top = src.Top
	}
	if srcMask&sideRight != 0 {
		(*dst).Right = src.Right
	}
	if srcMask&sideBottom != 0 {
		(*dst).Bottom = src.Bottom
	}
	if srcMask&sideLeft != 0 {
		(*dst).Left = src.Left
	}
	*dstMask |= srcMask
}

func mergePaint(dst, src *PaintStyle) {
	if src.backgroundSet || src.Background != nil {
		dst.Background = clonePaintSource(src.Background)
		dst.backgroundSet = true
	}
	if src.Border != nil {
		if dst.Border == nil {
			dst.Border = &BorderStyle{}
		}
		if src.Border.Color != nil {
			dst.Border.Color = cloneColorSource(src.Border.Color)
		}
		dst.Border.Width = pick(dst.Border.Width, src.Border.Width)
		dst.Border.LineStyle = pick(dst.Border.LineStyle, src.Border.LineStyle)
	}
	if src.BorderBottom != nil {
		if dst.BorderBottom == nil {
			dst.BorderBottom = &BorderStyle{}
		}
		if src.BorderBottom.Color != nil {
			dst.BorderBottom.Color = cloneColorSource(src.BorderBottom.Color)
		}
		dst.BorderBottom.Width = pick(dst.BorderBottom.Width, src.BorderBottom.Width)
		dst.BorderBottom.LineStyle = pick(dst.BorderBottom.LineStyle, src.BorderBottom.LineStyle)
	}
	if src.Radius != nil {
		dst.Radius = clonePtr(src.Radius)
		dst.Radii = nil
		dst.radiusMask = 0
	}
	mergeRadii(dst, src)
	if src.shadowsSet {
		dst.Shadows = cloneShadows(src.Shadows)
		dst.shadowsSet = true
	}
	if src.Outline != nil {
		outline := *src.Outline
		outline.Color = cloneColorSource(src.Outline.Color)
		dst.Outline = &outline
	}
	if src.visualOutsetSet {
		dst.VisualOutset = clonePtr(src.VisualOutset)
		dst.visualOutsetSet = true
	}
	dst.Opacity = pick(dst.Opacity, src.Opacity)
}

func mergeRadii(dst, src *PaintStyle) {
	if src.Radii == nil {
		return
	}
	base := CornerRadii{}
	if dst.Radii != nil {
		base = *dst.Radii
	} else if dst.Radius != nil {
		base = uniformRadii(*dst.Radius)
	}
	mask := src.radiusMask
	if mask == 0 {
		mask = 0x0f
	}
	if mask&(1<<0) != 0 {
		base.TopLeft = src.Radii.TopLeft
	}
	if mask&(1<<1) != 0 {
		base.TopRight = src.Radii.TopRight
	}
	if mask&(1<<2) != 0 {
		base.BottomRight = src.Radii.BottomRight
	}
	if mask&(1<<3) != 0 {
		base.BottomLeft = src.Radii.BottomLeft
	}
	dst.Radii = &base
	dst.radiusMask |= mask
}

func mergeText(dst, src *TextStyle) {
	if src.Color != nil {
		dst.Color = cloneColorSource(src.Color)
	}
	dst.FontSize = pick(dst.FontSize, src.FontSize)
	dst.FontWeight = pick(dst.FontWeight, src.FontWeight)
	dst.Typeface = pick(dst.Typeface, src.Typeface)
	dst.FontStyle = pick(dst.FontStyle, src.FontStyle)
	dst.LineHeight = pick(dst.LineHeight, src.LineHeight)
	dst.LineHeightScale = pick(dst.LineHeightScale, src.LineHeightScale)
	dst.MaxLines = pick(dst.MaxLines, src.MaxLines)
	dst.Align = pick(dst.Align, src.Align)
	dst.Wrap = pick(dst.Wrap, src.Wrap)
	dst.Truncator = pick(dst.Truncator, src.Truncator)
	dst.Underline = pick(dst.Underline, src.Underline)
}

func mergeTransform(dst, src *TransformStyle) {
	dst.TranslateX = pick(dst.TranslateX, src.TranslateX)
	dst.TranslateY = pick(dst.TranslateY, src.TranslateY)
	dst.ScaleX = pick(dst.ScaleX, src.ScaleX)
	dst.ScaleY = pick(dst.ScaleY, src.ScaleY)
	dst.Rotate = pick(dst.Rotate, src.Rotate)
}

func mergeTransitions(first, second []Transition) []Transition {
	// Pre-allocate with capacity for worst case (no duplicates)
	result := make([]Transition, len(first), len(first)+len(second))
	copy(result, first)
	for _, incoming := range second {
		found := false
		for index := range result {
			if result[index].Property == incoming.Property {
				result[index] = incoming
				found = true
				break
			}
		}
		if !found {
			result = append(result, incoming)
		}
	}
	return result
}

func clone(source Style) Style {
	result := source
	result.box = cloneBox(source.box)
	result.paint = clonePaint(source.paint)
	result.text = cloneText(source.text)
	result.trans = cloneTransform(source.trans)
	// Pre-allocate slices with exact capacity
	if len(source.transitions) > 0 {
		result.transitions = make([]Transition, len(source.transitions))
		copy(result.transitions, source.transitions)
	} else {
		result.transitions = nil
	}
	result.dim = clonePtr(source.dim)
	result.parts = cloneParts(source.parts)
	if len(source.rules) > 0 {
		result.rules = make([]Rule, len(source.rules))
		copy(result.rules, source.rules)
	} else {
		result.rules = nil
	}
	return result
}

func cloneParts(source map[Part]Style) map[Part]Style {
	if len(source) == 0 {
		return nil
	}
	result := make(map[Part]Style, len(source))
	for part, declaration := range source {
		result[part] = clone(declaration)
	}
	return result
}

func cloneBox(source *BoxStyle) *BoxStyle {
	if source == nil {
		return nil
	}
	result := *source
	result.Width = clonePtr(source.Width)
	result.Height = clonePtr(source.Height)
	result.MinWidth = clonePtr(source.MinWidth)
	result.MaxWidth = clonePtr(source.MaxWidth)
	result.MinHeight = clonePtr(source.MinHeight)
	result.MaxHeight = clonePtr(source.MaxHeight)
	result.FillWidth = clonePtr(source.FillWidth)
	result.FillHeight = clonePtr(source.FillHeight)
	result.AspectRatio = clonePtr(source.AspectRatio)
	result.Padding = clonePtr(source.Padding)
	result.Margin = clonePtr(source.Margin)
	result.Overflow = clonePtr(source.Overflow)
	result.Cursor = clonePtr(source.Cursor)
	result.paddingMask = source.paddingMask
	result.marginMask = source.marginMask
	return &result
}

func clonePaint(source *PaintStyle) *PaintStyle {
	if source == nil {
		return nil
	}
	result := *source
	result.Background = clonePaintSource(source.Background)
	result.backgroundSet = source.backgroundSet
	if source.Border != nil {
		border := *source.Border
		border.Color = cloneColorSource(source.Border.Color)
		border.Width = clonePtr(source.Border.Width)
		border.LineStyle = clonePtr(source.Border.LineStyle)
		result.Border = &border
	}
	if source.BorderBottom != nil {
		border := *source.BorderBottom
		border.Color = cloneColorSource(source.BorderBottom.Color)
		border.Width = clonePtr(source.BorderBottom.Width)
		border.LineStyle = clonePtr(source.BorderBottom.LineStyle)
		result.BorderBottom = &border
	}
	result.Radius = clonePtr(source.Radius)
	result.Radii = clonePtr(source.Radii)
	result.radiusMask = source.radiusMask
	result.Shadows = cloneShadows(source.Shadows)
	result.shadowsSet = source.shadowsSet
	if source.Outline != nil {
		outline := *source.Outline
		outline.Color = cloneColorSource(source.Outline.Color)
		result.Outline = &outline
	}
	result.VisualOutset = clonePtr(source.VisualOutset)
	result.visualOutsetSet = source.visualOutsetSet
	result.Opacity = clonePtr(source.Opacity)
	return &result
}

//go:inline
func cloneShadows(source []Shadow) []Shadow {
	if len(source) == 0 {
		return nil
	}
	result := make([]Shadow, len(source))
	for index, shadow := range source {
		result[index] = shadow
		result[index].Color = cloneColorSource(shadow.Color)
		result[index].Profile = clonePtr(shadow.Profile)
	}
	return result
}

func cloneText(source *TextStyle) *TextStyle {
	if source == nil {
		return nil
	}
	result := *source
	result.Color = cloneColorSource(source.Color)
	result.FontSize = clonePtr(source.FontSize)
	result.FontWeight = clonePtr(source.FontWeight)
	result.Typeface = clonePtr(source.Typeface)
	result.FontStyle = clonePtr(source.FontStyle)
	result.LineHeight = clonePtr(source.LineHeight)
	result.LineHeightScale = clonePtr(source.LineHeightScale)
	result.MaxLines = clonePtr(source.MaxLines)
	result.Align = clonePtr(source.Align)
	result.Wrap = clonePtr(source.Wrap)
	result.Truncator = clonePtr(source.Truncator)
	result.Underline = clonePtr(source.Underline)
	return &result
}

func cloneTransform(source *TransformStyle) *TransformStyle {
	if source == nil {
		return nil
	}
	result := *source
	result.TranslateX = clonePtr(source.TranslateX)
	result.TranslateY = clonePtr(source.TranslateY)
	result.ScaleX = clonePtr(source.ScaleX)
	result.ScaleY = clonePtr(source.ScaleY)
	result.Rotate = clonePtr(source.Rotate)
	return &result
}

func clonePaintSource(source PaintSource) PaintSource {
	switch value := source.(type) {
	case Gradient:
		value.Stops = cloneGradientStops(value.Stops)
		return value
	case *Gradient:
		if value == nil {
			return nil
		}
		copy := *value
		copy.Stops = cloneGradientStops(value.Stops)
		return &copy
	case *SolidColor:
		if value == nil {
			return nil
		}
		copy := *value
		return &copy
	default:
		return source
	}
}

//go:inline
func cloneGradientStops(source []GradientStop) []GradientStop {
	if len(source) == 0 {
		return nil
	}
	result := make([]GradientStop, len(source))
	for index, stop := range source {
		result[index] = GradientStop{Offset: stop.Offset, Color: cloneColorSource(stop.Color)}
	}
	return result
}

func cloneColorSource(source ColorSource) ColorSource {
	if source == nil {
		return nil
	}
	cloned := clonePaintSource(source)
	if cloned == nil {
		return nil
	}
	return cloned.(ColorSource)
}

//go:inline
func clonePtr[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

//go:inline
func pick[T any](current, incoming *T) *T {
	if incoming == nil {
		return current
	}
	return clonePtr(incoming)
}
