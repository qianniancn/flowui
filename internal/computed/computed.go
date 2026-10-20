// Package computed turns style declarations into final values: it cascades
// component declarations, inherited scopes, and the instance style in order,
// then applies transitions and caches the result.
//
// Three neighbouring packages are easy to confuse with it, so the split is
// deliberate:
//
//   - style: the pure declaration mechanism. It imports nothing from the
//     framework, which is what lets theme depend on it.
//   - style/resolve: the public, author-facing entry points, plus the Context,
//     Widget, and Brush aliases that make this engine reachable from an
//     application.
//   - computed: this package — the implementation behind those entry points.
//
// The API here is expressed in internal types (`*core.Context`), so this
// package cannot itself be an author-facing entry point.
package computed

import (
	"image/color"
	"math"
	"time"
	"unsafe"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const stateSlot = "computed"

// Resolve cascades component declarations, inherited scopes, and the
// instance style in that order, then applies transitions.
func Resolve(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	styleState style.State,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	resolved := resolveStatic(ctx, styleState, defaults, variant, size, custom, true)
	return ApplyTransitions(ctx, gtx, key, resolved)
}

// ResolveParts resolves several parts that share the same component cascade.
// The returned snapshots follow the order of parts. It is intended for
// compound widgets that need a root and multiple named parts in one frame.
func ResolveParts(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	parts []style.Part,
	defaults, variant, size, custom style.Style,
) []style.ResolvedStyle {
	resolved := make([]style.ResolvedStyle, len(parts))
	ResolvePartsInto(ctx, gtx, key, state, parts, resolved, defaults, variant, size, custom)
	return resolved
}

// ResolvePartsInto is the allocation-conscious form of ResolveParts. The
// destination must have room for every requested part; extra capacity is
// ignored. A shared prepared cascade keeps layer collection and cache-key
// inputs out of the per-part loop.
func ResolvePartsInto(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	parts []style.Part,
	destination []style.ResolvedStyle,
	defaults, variant, size, custom style.Style,
) {
	if len(destination) < len(parts) {
		return
	}
	prepared := prepareResolveLayers(ctx, state, defaults, variant, size, custom)
	for index, part := range parts {
		resolved := resolvePreparedPart(ctx, &prepared, part, cacheCloneAnimation)
		if part == style.PartRoot {
			destination[index] = ApplyTransitions(ctx, gtx, key, resolved)
		} else {
			destination[index] = ApplyPartTransitions(ctx, gtx, key, part, resolved)
		}
	}
}

// ApplyTransitions animates a style that has already been resolved. Styles
// without transitions remain stateless.
//
// An empty key soft-fails: transitions are not applied (values stay at their
// resolved targets) so callers without identity do not panic on state slots.
func ApplyTransitions(ctx *core.Context, gtx layout.Context, key string, resolved style.ResolvedStyle) style.ResolvedStyle {
	if len(resolved.Transitions) == 0 {
		return resolved
	}
	if core.ActiveTheme(ctx) == nil {
		return resolved
	}
	if key == "" {
		return resolved
	}
	// Isolate transition slots from interact/component kinds under a derived
	// identity so two unrelated claims on the user key cannot collide, while
	// the same component key still shares one animation state.
	animate(ctx, gtx, core.DerivedKey(ctx, key, "style-transition"), &resolved)
	return resolved
}

// ApplyTransitionsWithKey applies transitions with an already-derived
// animation identity. The identity must be retained by the owning component.
func ApplyTransitionsWithKey(ctx *core.Context, gtx layout.Context, animationKey string, resolved style.ResolvedStyle) style.ResolvedStyle {
	if len(resolved.Transitions) == 0 || core.ActiveTheme(ctx) == nil || animationKey == "" {
		return resolved
	}
	// UseState is intentional here: keyed animation state is frame-owned and
	// must be marked as used even when the settled fast path skips mutation.
	state := core.UseState[engineState](ctx, animationKey, stateSlot)
	if animationSettled(*state, resolved) {
		return resolved
	}
	animateState(ctx, gtx, state, &resolved)
	return resolved
}

// ApplyOpacity applies a resolved Paint opacity to a concrete color.
func ApplyOpacity(value color.NRGBA, opacity *float32) color.NRGBA {
	if opacity == nil {
		return value
	}
	value.A = byte(float32(value.A)*min(max(*opacity, 0), 1) + .5)
	return value
}

// ApplyOutlineOpacity returns a copy whose outline alpha is multiplied by opacity.
func ApplyOutlineOpacity(value style.ResolvedStyle, opacity float32) style.ResolvedStyle {
	if value.Paint == nil || value.Paint.Outline == nil {
		return value
	}
	paintStyle := *value.Paint
	outline := *paintStyle.Outline
	if col, ok := Color(outline.Color); ok {
		col.A = byte(float32(col.A)*min(max(opacity, 0), 1) + .5)
		outline.Color = style.SolidColor{Color: col}
	}
	paintStyle.Outline = &outline
	value.Paint = &paintStyle
	return value
}

// HasTransitions reports whether any resolved style needs runtime state.
func HasTransitions(values ...style.ResolvedStyle) bool {
	for _, value := range values {
		if len(value.Transitions) != 0 {
			return true
		}
	}
	return false
}

// ResolveStatic performs the same cascade without retaining transition
// state. It is used for measurement passes.
func ResolveStatic(
	ctx *core.Context,
	state style.State,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	return resolveStatic(ctx, state, defaults, variant, size, custom, false)
}

// ResolvePartsStatic resolves several parts that share the same component
// cascade without retaining transition state. The returned snapshots follow
// the order of parts.
func ResolvePartsStatic(
	ctx *core.Context,
	state style.State,
	parts []style.Part,
	defaults, variant, size, custom style.Style,
) []style.ResolvedStyle {
	resolved := make([]style.ResolvedStyle, len(parts))
	ResolvePartsStaticInto(ctx, state, parts, resolved, defaults, variant, size, custom)
	return resolved
}

// ResolvePartsStaticInto is the allocation-conscious form of
// ResolvePartsStatic. The destination must have room for every requested
// part; extra capacity is ignored. A shared prepared cascade keeps layer
// collection and cache-key inputs out of the per-part loop.
func ResolvePartsStaticInto(
	ctx *core.Context,
	state style.State,
	parts []style.Part,
	destination []style.ResolvedStyle,
	defaults, variant, size, custom style.Style,
) {
	if len(destination) < len(parts) {
		return
	}
	prepared := prepareResolveLayers(ctx, state, defaults, variant, size, custom)
	for index, part := range parts {
		destination[index] = resolvePreparedPart(ctx, &prepared, part, cacheCloneStatic)
	}
}

func resolveStatic(
	ctx *core.Context,
	state style.State,
	defaults, variant, size, custom style.Style,
	forAnimation bool,
) style.ResolvedStyle {
	prepared := prepareResolveLayers(ctx, state, defaults, variant, size, custom)
	return resolvePreparedPart(ctx, &prepared, style.PartRoot, ownershipForAnimation(forAnimation))
}

func ownershipForAnimation(forAnimation bool) cacheOwnership {
	if forAnimation {
		return cacheCloneAnimation
	}
	return cacheCloneStatic
}

type preparedResolveLayers struct {
	state      style.State
	defaults   style.Style
	variant    style.Style
	size       style.Style
	custom     style.Style
	layers     []style.Style
	cacheSafe  bool
	themeHash  uint64
	stateHash  uint16
	layersHash uint64
}

func prepareResolveLayers(
	ctx *core.Context,
	state style.State,
	defaults, variant, size, custom style.Style,
) preparedResolveLayers {
	prepared := preparedResolveLayers{
		state:     state,
		defaults:  defaults,
		variant:   variant,
		size:      size,
		custom:    custom,
		cacheSafe: styleLayersCacheSafe(ctx, defaults, variant, size, custom),
	}
	if !DisableCascadeCache && prepared.cacheSafe {
		activeTheme := core.ActiveTheme(ctx)
		prepared.themeHash = core.CachedThemeHash(ctx, activeTheme, func() uint64 {
			return computeThemeHash(activeTheme)
		})
		prepared.stateHash = computeStyleStateHash(state)
		prepared.layersHash = hashStyleLayers(ctx, state, defaults, variant, size, custom)
	}
	return prepared
}

func resolvePreparedPart(
	ctx *core.Context,
	prepared *preparedResolveLayers,
	part style.Part,
	ownership cacheOwnership,
) style.ResolvedStyle {
	if !DisableCascadeCache && prepared.cacheSafe {
		cacheKey := cascadeCacheKey{
			contextPtr:   uintptr(unsafe.Pointer(ctx)),
			componentKey: string(part),
			stateHash:    prepared.stateHash,
			themeHash:    prepared.themeHash,
			customHash:   prepared.layersHash,
		}
		if cached, found := globalCascadeCache.getShared(cacheKey); found {
			if len(cached.Transitions) == 0 {
				return cached
			}
			switch ownership {
			case cacheShare:
				return cached
			case cacheCloneAnimation:
				return cloneResolvedStyleForAnimation(cached)
			default:
				return cloneResolvedStyle(cached)
			}
		}

		resolved := cascadePreparedState(ctx, prepared.state, prepared.layersFor(ctx), part)
		globalCascadeCache.put(cacheKey, resolved)
		return resolved
	}

	return cascadePreparedState(ctx, prepared.state, prepared.layersFor(ctx), part)
}

func (prepared *preparedResolveLayers) layersFor(ctx *core.Context) []style.Style {
	if prepared.layers == nil {
		prepared.layers = collectStyleLayers(ctx, prepared.defaults, prepared.variant, prepared.size, prepared.custom)
	}
	return prepared.layers
}

func cascadePreparedState(ctx *core.Context, state style.State, layers []style.Style, part style.Part) style.ResolvedStyle {
	var resolved style.ResolvedStyle
	if part == style.PartRoot {
		resolved = style.Cascade(state, layers...)
	} else {
		resolved = style.CascadePart(state, part, layers...)
	}
	if resolved.Paint != nil {
		resolved.Paint.Shadows = resolveShadows(resolved.Paint.Shadows, core.ActiveTheme(ctx))
	}
	return resolved
}

// ResolvePart cascades and animates a named element inside a compound component.
func ResolvePart(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	part style.Part,
	state style.State,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	if part == style.PartRoot {
		return Resolve(ctx, gtx, key, state, defaults, variant, size, custom)
	}
	resolved := resolvePartStatic(ctx, part, state, defaults, variant, size, custom, true)
	return ApplyPartTransitions(ctx, gtx, key, part, resolved)
}

// ApplyPartTransitions animates an already-resolved named part under its
// component-owned identity.
func ApplyPartTransitions(ctx *core.Context, gtx layout.Context, key string, part style.Part, resolved style.ResolvedStyle) style.ResolvedStyle {
	if part == style.PartRoot {
		return ApplyTransitions(ctx, gtx, key, resolved)
	}
	// Soft-fail like root ApplyTransitions: no identity means snap, no panic.
	if key == "" {
		return resolved
	}
	return ApplyTransitions(ctx, gtx, core.DerivedKey(ctx, key, "style-part:"+string(part)), resolved)
}

// ResolvePartStatic resolves a named part without retaining transition state.
func ResolvePartStatic(
	ctx *core.Context,
	part style.Part,
	state style.State,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	return resolvePartStatic(ctx, part, state, defaults, variant, size, custom, false)
}

// ResolvePartStaticShared resolves a named part and may return immutable
// cache-owned branches. Callers must treat the result as read-only and use
// ApplyTransitionsWithSharedKey before any animation is applied.
func ResolvePartStaticShared(
	ctx *core.Context,
	part style.Part,
	state style.State,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	return resolvePartStaticShared(ctx, part, state, defaults, variant, size, custom)
}

// ApplyTransitionsWithSharedKey applies transitions to a static result that
// may be shared by a local component cache. It clones only while animation is
// active, so a cache-owned snapshot is never mutated by the animation engine.
func ApplyTransitionsWithSharedKey(
	ctx *core.Context,
	gtx layout.Context,
	animationKey string,
	resolved style.ResolvedStyle,
) style.ResolvedStyle {
	if len(resolved.Transitions) == 0 || core.ActiveTheme(ctx) == nil || animationKey == "" {
		return resolved
	}
	state := core.UseState[engineState](ctx, animationKey, stateSlot)
	if animationSettled(*state, resolved) {
		return resolved
	}
	resolved = cloneResolvedStyleForAnimation(resolved)
	animateState(ctx, gtx, state, &resolved)
	return resolved
}

// ResolvePartWithKey resolves a named part for animation and applies
// transitions using a caller-retained identity. It is the allocation-conscious
// counterpart of ResolvePart for repeated keyed children.
func ResolvePartWithKey(
	ctx *core.Context,
	gtx layout.Context,
	transitionKey string,
	part style.Part,
	state style.State,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	resolved := resolvePartStaticShared(ctx, part, state, defaults, variant, size, custom)
	if len(resolved.Transitions) != 0 && core.ActiveTheme(ctx) != nil && transitionKey != "" {
		animation, ok := core.PeekState[engineState](ctx, transitionKey, stateSlot)
		if !ok || !animationSettled(*animation, resolved) {
			// Shared cascade entries are immutable. Copy only when animation is
			// still active and animate must mutate the resolved branches.
			resolved = cloneResolvedStyleForAnimation(resolved)
		}
	}
	return ApplyTransitionsWithKey(ctx, gtx, transitionKey, resolved)
}

func resolvePartStaticShared(
	ctx *core.Context,
	part style.Part,
	state style.State,
	defaults, variant, size, custom style.Style,
) style.ResolvedStyle {
	return resolvePartStaticWithOwnership(ctx, part, state, defaults, variant, size, custom, cacheShare)
}

func resolvePartStatic(
	ctx *core.Context,
	part style.Part,
	state style.State,
	defaults, variant, size, custom style.Style,
	forAnimation bool,
) style.ResolvedStyle {
	ownership := cacheCloneStatic
	if forAnimation {
		ownership = cacheCloneAnimation
	}
	return resolvePartStaticWithOwnership(ctx, part, state, defaults, variant, size, custom, ownership)
}

type cacheOwnership uint8

const (
	cacheCloneStatic cacheOwnership = iota
	cacheCloneAnimation
	cacheShare
)

func resolvePartStaticWithOwnership(
	ctx *core.Context,
	part style.Part,
	state style.State,
	defaults, variant, size, custom style.Style,
	ownership cacheOwnership,
) style.ResolvedStyle {
	prepared := prepareResolveLayers(ctx, state, defaults, variant, size, custom)
	return resolvePreparedPart(ctx, &prepared, part, ownership)
}

// computeThemeHash hashes every theme value resolution still reads: the
// shadow profiles and the menu shadow override. Metric and color values now
// travel inside declarations, so this hash stays conservative.
func computeThemeHash(t *theme.Theme) uint64 {
	const offset64 = 14695981039346656037
	const prime64 = 1099511628211
	hash := uint64(offset64)
	if t == nil {
		return hash
	}
	mix := func(value uint64) {
		hash ^= value
		hash *= prime64
	}
	mixMetric := func(value float32) { mix(uint64(math.Float32bits(value))) }
	mixColor := func(value color.NRGBA) {
		mix(uint64(value.R)<<24 | uint64(value.G)<<16 | uint64(value.B)<<8 | uint64(value.A))
	}

	for _, value := range []float32{
		float32(t.Typography.BodySize), float32(t.Typography.ControlSize), float32(t.Typography.SmallSize),
		float32(t.Components.Checkbox.Radius),
	} {
		mixMetric(value)
	}

	sem := t.Semantic()
	for _, value := range []color.NRGBA{
		sem.Background, sem.Foreground, sem.Surface, sem.SurfaceForeground, sem.SurfaceSecondary,
		sem.SurfaceSecondaryForeground, sem.SurfaceTertiary, sem.SurfaceTertiaryForeground,
		sem.SurfaceHover, sem.SurfacePressed, sem.SurfaceRaised, sem.Overlay, sem.OverlayForeground,
		sem.Muted, sem.MutedForeground, sem.Border, sem.Separator,
		sem.FieldBackground, sem.FieldHover, sem.FieldForeground, sem.FieldPlaceholder, sem.FieldFocus,
		sem.Segment, sem.SegmentForeground, sem.Ring, sem.Selection,
		sem.DisabledBackground, sem.DisabledForeground, sem.DisabledBorder,
		sem.SurfaceShadow, sem.OverlayShadow,
	} {
		mixColor(value)
	}
	for i := range sem.Tones {
		tone := sem.Tones[i]
		for _, value := range []color.NRGBA{tone.Fill, tone.OnFill, tone.Hover, tone.Pressed, tone.Soft, tone.SoftHover, tone.OnSoft} {
			mixColor(value)
		}
	}

	// A gradient is a semantic role (brand surface decoration). Nil means "no
	// gradient", so themes without one keep their previous hash.
	if gradient := sem.PrimaryGradient; gradient != nil {
		mixMetric(gradient.AngleDegrees)
		mix(uint64(len(gradient.Stops)))
		for _, stop := range gradient.Stops {
			mixMetric(stop.Offset)
			mixColor(stop.Color)
		}
	}

	for _, profile := range []theme.ShadowTheme{t.Shadows.Surface, t.Shadows.Overlay, t.Shadows.Menu} {
		for _, layer := range profile.Layers {
			mixMetric(float32(layer.OffsetX))
			mixMetric(float32(layer.OffsetY))
			mixMetric(float32(layer.Blur))
			mixMetric(float32(layer.Spread))
			mixMetric(layer.Opacity)
		}
	}
	mixMetric(t.Components.Menu.ShadowOpacity)
	mixColor(t.Components.Menu.ShadowColor)
	return hash
}

func resolveShadows(shadows []style.Shadow, activeTheme *theme.Theme) []style.Shadow {
	if len(shadows) == 0 {
		return nil
	}
	result := make([]style.Shadow, 0, len(shadows))
	for _, shadow := range shadows {
		if shadow.Profile == nil {
			shadow.Color = resolveColor(shadow.Color)
			result = append(result, shadow)
			continue
		}
		profile, col := themeShadow(*shadow.Profile, activeTheme)
		profileOpacity := float32(1)
		if *shadow.Profile == style.ShadowMenu && activeTheme != nil {
			profileOpacity = clampShadowOpacity(activeTheme.Components.Menu.ShadowOpacity)
		}
		for _, layer := range profile.Layers {
			layerColor := col
			layerColor.A = scaledAlpha(layerColor.A, layer.Opacity*profileOpacity)
			if layer.Blur < 0 || layerColor.A == 0 {
				continue
			}
			result = append(result, style.Shadow{
				OffsetX: layer.OffsetX,
				OffsetY: layer.OffsetY,
				Blur:    layer.Blur,
				Spread:  layer.Spread,
				Color:   style.SolidColor{Color: layerColor},
			})
		}
	}
	return result
}

func clampShadowOpacity(value float32) float32 {
	if value <= 0 || math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return 0
	}
	return min(value, 1)
}

func themeShadow(profile style.ShadowProfile, activeTheme *theme.Theme) (theme.ShadowTheme, color.NRGBA) {
	if activeTheme == nil {
		return theme.ShadowTheme{}, color.NRGBA{}
	}
	sem := activeTheme.Semantic()
	if profile == style.ShadowOverlay {
		return activeTheme.Shadows.Overlay, sem.OverlayShadow
	}
	if profile == style.ShadowMenu {
		return activeTheme.Shadows.Menu, theme.ColorOr(activeTheme.Components.Menu.ShadowColor, sem.OverlayShadow)
	}
	return activeTheme.Shadows.Surface, sem.SurfaceShadow
}

func scaledAlpha(alpha uint8, opacity float32) uint8 {
	if opacity <= 0 {
		return 0
	}
	if opacity >= 1 {
		return alpha
	}
	return uint8(float32(alpha)*opacity + .5)
}

func resolveColor(source style.ColorSource) style.SolidColor {
	if source == nil {
		return style.SolidColor{}
	}
	switch value := source.(type) {
	case style.SolidColor:
		return value
	case *style.SolidColor:
		if value == nil {
			return style.SolidColor{}
		}
		return *value
	default:
		return style.SolidColor{}
	}
}

type engineState struct {
	background   colorAnimation
	text         colorAnimation
	border       colorAnimation
	borderBottom colorAnimation
	outline      colorAnimation
	opacity      floatAnimation
	radius       floatAnimation
	radiusNW     floatAnimation
	radiusNE     floatAnimation
	radiusSE     floatAnimation
	radiusSW     floatAnimation
	translateX   floatAnimation
	translateY   floatAnimation
	scaleX       floatAnimation
	scaleY       floatAnimation
	rotate       floatAnimation
}

func animationSettled(state engineState, resolved style.ResolvedStyle) bool {
	if resolved.Paint != nil {
		if hasAnimationTransition(resolved, style.PropBackgroundColor) {
			value, ok := solidColor(resolved.Paint.Background)
			if ok && !colorSettled(state.background, value) {
				return false
			}
		}
		if resolved.Paint.Border != nil && hasAnimationTransition(resolved, style.PropBorderColor) {
			value, ok := solidColor(resolved.Paint.Border.Color)
			if ok && !colorSettled(state.border, value) {
				return false
			}
		}
		if resolved.Paint.BorderBottom != nil && hasAnimationTransition(resolved, style.PropBorderColor) {
			value, ok := solidColor(resolved.Paint.BorderBottom.Color)
			if ok && !colorSettled(state.borderBottom, value) {
				return false
			}
		}
		if resolved.Paint.Outline != nil && hasAnimationTransition(resolved, style.PropOutlineColor) {
			value, ok := solidColor(resolved.Paint.Outline.Color)
			if ok && !colorSettled(state.outline, value) {
				return false
			}
		}
		if resolved.Paint.Opacity != nil && hasAnimationTransition(resolved, style.PropOpacity) && !floatSettled(state.opacity, *resolved.Paint.Opacity) {
			return false
		}
		if resolved.Paint.Radius != nil && hasAnimationTransition(resolved, style.PropRadius) && !floatSettled(state.radius, float32(*resolved.Paint.Radius)) {
			return false
		}
		if resolved.Paint.Radii != nil {
			if hasAnimationTransition(resolved, style.PropRadius) &&
				(!floatSettled(state.radiusNW, float32(resolved.Paint.Radii.TopLeft)) ||
					!floatSettled(state.radiusNE, float32(resolved.Paint.Radii.TopRight)) ||
					!floatSettled(state.radiusSE, float32(resolved.Paint.Radii.BottomRight)) ||
					!floatSettled(state.radiusSW, float32(resolved.Paint.Radii.BottomLeft))) {
				return false
			}
		}
	}
	if resolved.Text != nil {
		if hasAnimationTransition(resolved, style.PropTextColor) {
			value, ok := solidColor(resolved.Text.Color)
			if ok && !colorSettled(state.text, value) {
				return false
			}
		}
	}
	if resolved.Trans != nil && hasAnimationTransition(resolved, style.PropTransform) {
		if resolved.Trans.TranslateX != nil && !floatSettled(state.translateX, float32(*resolved.Trans.TranslateX)) {
			return false
		}
		if resolved.Trans.TranslateY != nil && !floatSettled(state.translateY, float32(*resolved.Trans.TranslateY)) {
			return false
		}
		if resolved.Trans.ScaleX != nil && !floatSettled(state.scaleX, *resolved.Trans.ScaleX) {
			return false
		}
		if resolved.Trans.ScaleY != nil && !floatSettled(state.scaleY, *resolved.Trans.ScaleY) {
			return false
		}
		if resolved.Trans.Rotate != nil && !floatSettled(state.rotate, *resolved.Trans.Rotate) {
			return false
		}
	}
	return true
}

func hasAnimationTransition(resolved style.ResolvedStyle, property style.PropertyID) bool {
	return transitionFor(resolved, property).ok
}

func colorSettled(value colorAnimation, target color.NRGBA) bool {
	return value.transition.Ready() && value.transition.Current() == target && value.transition.Target() == target
}

func floatSettled(value floatAnimation, target float32) bool {
	return value.transition.Ready() && value.transition.Current() == target && value.transition.Target() == target
}

func animate(ctx *core.Context, gtx layout.Context, key string, resolved *style.ResolvedStyle) {
	state := core.UseState[engineState](ctx, key, stateSlot)
	animateState(ctx, gtx, state, resolved)
}

func animateState(ctx *core.Context, gtx layout.Context, state *engineState, resolved *style.ResolvedStyle) {
	motion := core.ActiveTheme(ctx).Motion

	if resolved.Paint != nil {
		if value, ok := solidColor(resolved.Paint.Background); ok {
			value = state.background.value(gtx, value, transitionFor(*resolved, style.PropBackgroundColor), motion)
			resolved.Paint.Background = style.SolidColor{Color: value}
		}
		if resolved.Paint.Border != nil {
			if value, ok := solidColor(resolved.Paint.Border.Color); ok {
				value = state.border.value(gtx, value, transitionFor(*resolved, style.PropBorderColor), motion)
				resolved.Paint.Border.Color = style.SolidColor{Color: value}
			}
		}
		if resolved.Paint.BorderBottom != nil {
			if value, ok := solidColor(resolved.Paint.BorderBottom.Color); ok {
				value = state.borderBottom.value(gtx, value, transitionFor(*resolved, style.PropBorderColor), motion)
				resolved.Paint.BorderBottom.Color = style.SolidColor{Color: value}
			}
		}
		if resolved.Paint.Outline != nil {
			if value, ok := solidColor(resolved.Paint.Outline.Color); ok {
				value = state.outline.value(gtx, value, transitionFor(*resolved, style.PropOutlineColor), motion)
				resolved.Paint.Outline.Color = style.SolidColor{Color: value}
			}
		}
		if resolved.Paint.Opacity != nil {
			value := state.opacity.value(gtx, *resolved.Paint.Opacity, transitionFor(*resolved, style.PropOpacity), motion)
			resolved.Paint.Opacity = &value
		}
		if resolved.Paint.Radius != nil {
			value := state.radius.value(gtx, float32(*resolved.Paint.Radius), transitionFor(*resolved, style.PropRadius), motion)
			radius := unit.Dp(value)
			resolved.Paint.Radius = &radius
		}
		if resolved.Paint.Radii != nil {
			transition := transitionFor(*resolved, style.PropRadius)
			resolved.Paint.Radii.TopLeft = unit.Dp(state.radiusNW.value(gtx, float32(resolved.Paint.Radii.TopLeft), transition, motion))
			resolved.Paint.Radii.TopRight = unit.Dp(state.radiusNE.value(gtx, float32(resolved.Paint.Radii.TopRight), transition, motion))
			resolved.Paint.Radii.BottomRight = unit.Dp(state.radiusSE.value(gtx, float32(resolved.Paint.Radii.BottomRight), transition, motion))
			resolved.Paint.Radii.BottomLeft = unit.Dp(state.radiusSW.value(gtx, float32(resolved.Paint.Radii.BottomLeft), transition, motion))
		}
	}
	if resolved.Text != nil {
		if value, ok := solidColor(resolved.Text.Color); ok {
			value = state.text.value(gtx, value, transitionFor(*resolved, style.PropTextColor), motion)
			resolved.Text.Color = style.SolidColor{Color: value}
		}
	}
	if resolved.Trans != nil {
		transition := transitionFor(*resolved, style.PropTransform)
		if resolved.Trans.TranslateX != nil {
			value := unit.Dp(state.translateX.value(gtx, float32(*resolved.Trans.TranslateX), transition, motion))
			resolved.Trans.TranslateX = &value
		}
		if resolved.Trans.TranslateY != nil {
			value := unit.Dp(state.translateY.value(gtx, float32(*resolved.Trans.TranslateY), transition, motion))
			resolved.Trans.TranslateY = &value
		}
		if resolved.Trans.ScaleX != nil {
			value := state.scaleX.value(gtx, *resolved.Trans.ScaleX, transition, motion)
			resolved.Trans.ScaleX = &value
		}
		if resolved.Trans.ScaleY != nil {
			value := state.scaleY.value(gtx, *resolved.Trans.ScaleY, transition, motion)
			resolved.Trans.ScaleY = &value
		}
		if resolved.Trans.Rotate != nil {
			value := state.rotate.value(gtx, *resolved.Trans.Rotate, transition, motion)
			resolved.Trans.Rotate = &value
		}
	}
}

func solidColor(source any) (color.NRGBA, bool) {
	switch value := source.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		if value == nil {
			return color.NRGBA{}, false
		}
		return value.Color, true
	default:
		return color.NRGBA{}, false
	}
}

type transitionSpec struct {
	value style.Transition
	ok    bool
}

func transitionFor(resolved style.ResolvedStyle, property style.PropertyID) transitionSpec {
	for index := len(resolved.Transitions) - 1; index >= 0; index-- {
		transition := resolved.Transitions[index]
		if transition.Property == property {
			return transitionSpec{value: transition, ok: true}
		}
	}
	return transitionSpec{}
}

type colorAnimation struct {
	transition animation.ColorTransition
	target     color.NRGBA
	changedAt  time.Time
	ready      bool
}

func (a *colorAnimation) value(gtx layout.Context, target color.NRGBA, spec transitionSpec, motion theme.MotionTheme) color.NRGBA {
	if !a.ready {
		a.ready = true
		a.target = target
		return a.transition.Value(gtx, target, 0, nil)
	}
	if target != a.target {
		a.target = target
		a.changedAt = gtx.Now
		if delay := transitionDelay(spec, motion); delay > 0 {
			a.transition.Set(a.transition.Current(), target, gtx.Now.Add(delay))
		}
	}
	if !spec.ok {
		return a.transition.Value(gtx, target, 0, nil)
	}
	if waitForDelay(gtx, a.changedAt, spec.value.Delay, motion) {
		return a.transition.Current()
	}
	return a.transition.Value(gtx, target, spec.value.Duration, specEasing(spec, motion), motion)
}

type floatAnimation struct {
	transition animation.FloatTransition
	target     float32
	changedAt  time.Time
	ready      bool
}

func (a *floatAnimation) value(gtx layout.Context, target float32, spec transitionSpec, motion theme.MotionTheme) float32 {
	if !a.ready {
		a.ready = true
		a.target = target
		return a.transition.Value(gtx, target, 0, nil)
	}
	if target != a.target {
		a.target = target
		a.changedAt = gtx.Now
		if delay := transitionDelay(spec, motion); delay > 0 {
			a.transition.Set(a.transition.Current(), target, gtx.Now.Add(delay))
		}
	}
	if !spec.ok {
		return a.transition.Value(gtx, target, 0, nil)
	}
	if waitForDelay(gtx, a.changedAt, spec.value.Delay, motion) {
		return a.transition.Current()
	}
	return a.transition.Value(gtx, target, spec.value.Duration, specEasing(spec, motion), motion)
}

// specEasing picks the transition curve: a declaration-level ease wins, then
// the style family's named curve, then the framework default.
func specEasing(spec transitionSpec, motion theme.MotionTheme) animation.Easing {
	if spec.value.Ease != nil {
		return animation.Easing(spec.value.Ease)
	}
	return animation.EasingFor(motion.Ease)
}

func waitForDelay(gtx layout.Context, changedAt time.Time, delay time.Duration, motion theme.MotionTheme) bool {
	delay = theme.ResolveMotionDuration(motion, delay)
	if delay <= 0 || changedAt.IsZero() || gtx.Now.Sub(changedAt) >= delay {
		return false
	}
	gtx.Execute(op.InvalidateCmd{})
	return true
}

func transitionDelay(spec transitionSpec, motion theme.MotionTheme) time.Duration {
	if !spec.ok {
		return 0
	}
	return theme.ResolveMotionDuration(motion, spec.value.Delay)
}
