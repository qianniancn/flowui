package style

func (s Style) resolve(state StyleState) Style {
	result := clone(s)
	result.conditions = nil
	for _, rule := range s.conditions {
		if rule.predicate != nil && rule.predicate(state) {
			result = merge(result, rule.override.resolve(state))
		}
	}
	return result
}

// Resolve applies this declaration's matching conditional rules.
func (s Style) Resolve(state StyleState) ResolvedStyle {
	return resolvedStyle(s.resolve(state))
}

func Cascade(state StyleState, layers ...Style) ResolvedStyle {
	var result Style
	for _, layer := range layers {
		result = merge(result, layer.resolve(state))
	}
	return resolvedStyle(result)
}

// Join combines declarations without resolving conditions or theme tokens.
// It is used internally when component defaults are assembled in stages.
func Join(styles ...Style) Style {
	var result Style
	for _, declaration := range styles {
		result = merge(result, declaration)
	}
	return result
}

// CascadePart resolves a named component part. Text properties inherit from
// each root layer; box, paint, and transform properties do not.
func CascadePart(state StyleState, part Part, layers ...Style) ResolvedStyle {
	if part == PartRoot {
		return Cascade(state, layers...)
	}
	var result Style
	for _, layer := range layers {
		resolved := layer.resolve(state)
		if resolved.text != nil {
			inherited := Style{text: cloneText(resolved.text)}
			for _, transition := range resolved.transitions {
				if transition.Property == PropTextColor {
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
	return ResolvedStyle{
		Box:         cloneBox(source.box),
		Paint:       clonePaint(source.paint),
		Text:        cloneText(source.text),
		Trans:       cloneTransform(source.trans),
		Transitions: append([]Transition(nil), source.transitions...),
	}
}

func merge(first, second Style) Style {
	result := first
	if second.box != nil {
		result.box = cloneBox(first.box)
		if result.box == nil {
			result.box = &BoxStyle{}
		}
		mergeBox(result.box, second.box)
	}
	if second.paint != nil {
		result.paint = clonePaint(first.paint)
		if result.paint == nil {
			result.paint = &PaintStyle{}
		}
		mergePaint(result.paint, second.paint)
	}
	if second.text != nil {
		result.text = cloneText(first.text)
		if result.text == nil {
			result.text = &TextStyle{}
		}
		mergeText(result.text, second.text)
	}
	if second.trans != nil {
		result.trans = cloneTransform(first.trans)
		if result.trans == nil {
			result.trans = &TransformStyle{}
		}
		mergeTransform(result.trans, second.trans)
	}
	if len(second.transitions) != 0 {
		result.transitions = mergeTransitions(first.transitions, second.transitions)
	}
	if len(second.conditions) != 0 {
		result.conditions = append(append([]condition(nil), first.conditions...), cloneConditions(second.conditions)...)
	}
	if len(second.parts) != 0 {
		result.parts = mergeParts(first.parts, second.parts)
	}
	if len(second.tokens) != 0 {
		result.tokens = append(append([]StyleToken(nil), first.tokens...), second.tokens...)
	}
	return result
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
}

func mergeTransform(dst, src *TransformStyle) {
	dst.TranslateX = pick(dst.TranslateX, src.TranslateX)
	dst.TranslateY = pick(dst.TranslateY, src.TranslateY)
	dst.ScaleX = pick(dst.ScaleX, src.ScaleX)
	dst.ScaleY = pick(dst.ScaleY, src.ScaleY)
	dst.Rotate = pick(dst.Rotate, src.Rotate)
}

func mergeTransitions(first, second []Transition) []Transition {
	result := append([]Transition(nil), first...)
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
	result.transitions = append([]Transition(nil), source.transitions...)
	result.conditions = cloneConditions(source.conditions)
	result.parts = cloneParts(source.parts)
	result.tokens = append([]StyleToken(nil), source.tokens...)
	return result
}

// ExpandTokens recursively replaces semantic theme tokens with declarations.
// Direct properties are merged last and therefore remain authoritative.
func ExpandTokens(source Style, resolver func(StyleToken) Style) Style {
	if resolver == nil {
		return clone(source)
	}
	var defaults Style
	for _, token := range source.tokens {
		defaults = merge(defaults, resolver(token))
	}
	direct := clone(source)
	direct.tokens = nil
	for index := range direct.conditions {
		direct.conditions[index].override = ExpandTokens(direct.conditions[index].override, resolver)
	}
	for part, declaration := range direct.parts {
		direct.parts[part] = ExpandTokens(declaration, resolver)
	}
	return merge(defaults, direct)
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
		result.Border = &border
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
	result.Opacity = clonePtr(source.Opacity)
	return &result
}

func cloneShadows(source []Shadow) []Shadow {
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

func cloneConditions(source []condition) []condition {
	if len(source) == 0 {
		return nil
	}
	result := make([]condition, len(source))
	for index, item := range source {
		result[index] = condition{
			predicate: item.predicate,
			override:  clone(item.override),
		}
	}
	return result
}

func clonePaintSource(source PaintSource) PaintSource {
	switch value := source.(type) {
	case StyleGradient:
		value.Stops = cloneGradientStops(value.Stops)
		return value
	case *StyleGradient:
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
	case *ThemeColor:
		if value == nil {
			return nil
		}
		copy := *value
		return &copy
	case AlphaColor:
		value.Source = cloneColorSource(value.Source)
		return value
	case *AlphaColor:
		if value == nil {
			return nil
		}
		copy := *value
		copy.Source = cloneColorSource(value.Source)
		return &copy
	default:
		return source
	}
}

func cloneGradientStops(source []StyleGradientStop) []StyleGradientStop {
	result := make([]StyleGradientStop, len(source))
	for index, stop := range source {
		result[index] = StyleGradientStop{Offset: stop.Offset, Color: cloneColorSource(stop.Color)}
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

func clonePtr[T any](value *T) *T {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func pick[T any](current, incoming *T) *T {
	if incoming == nil {
		return current
	}
	return clonePtr(incoming)
}
