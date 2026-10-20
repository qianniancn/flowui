package button

import (
	"image/color"
	"reflect"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type buttonStyle struct {
	root          style.ResolvedStyle
	content       style.ResolvedStyle
	indicatorPart style.ResolvedStyle
	visual        theme.ButtonVisual
}

const stateSlotStyle = "style"

type buttonDeclarationKey struct {
	request         theme.ButtonRequest
	overrideHash    uint64
	overridePresent bool
}

type buttonDeclarationBundle struct {
	visual theme.ButtonVisual
}

// buttonVisualValidationCache avoids repeating the complete provider
// contract walk for every instance that shares a theme and request when the
// provider opts into ImmutableVisualProvider. Mutable/custom providers are
// validated on every returned visual instead of sharing this fact. The
// provider declaration cache remains instance-owned; a theme replacement
// invalidates this map through ProviderCacheMap's normal lifecycle.
type buttonVisualValidationKey struct {
	provider theme.ComponentVisualProvider
	request  theme.ButtonRequest
}

var buttonVisualValidationCache = func() computed.ProviderCacheMap[buttonVisualValidationKey, struct{}] {
	var cache computed.ProviderCacheMap[buttonVisualValidationKey, struct{}]
	cache.SetCapacity(128)
	return cache
}()

// buttonDeclarationCache keeps the provider declaration bundle for one button
// state slot. The shared ProviderCache owns theme/key invalidation and
// successful-load semantics; this wrapper only keeps the button-specific
// cache type and test-facing stats name. Immutable providers may also cache a
// caller override by structural hash; mutable providers continue to resolve
// overrides on every call because a theme pointer alone is not a stability
// guarantee for their returned closures.
type buttonDeclarationCache struct {
	cache computed.ProviderCache[buttonDeclarationKey, buttonDeclarationBundle]
}

// ButtonDeclarationCacheStats is a snapshot of provider declaration cache
// activity for one Button instance.
type ButtonDeclarationCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64
	KeyInvalidations   uint64
}

// Stats returns a snapshot of cache activity.
func (c *buttonDeclarationCache) Stats() ButtonDeclarationCacheStats {
	if c == nil {
		return ButtonDeclarationCacheStats{}
	}
	stats := c.cache.Stats()
	return ButtonDeclarationCacheStats{
		Hits:               stats.Hits,
		Misses:             stats.Misses,
		ThemeInvalidations: stats.ThemeInvalidations,
		KeyInvalidations:   stats.KeyInvalidations,
	}
}

func (c *buttonDeclarationCache) resolve(activeTheme *theme.Theme, key buttonDeclarationKey, load func() buttonDeclarationBundle) buttonDeclarationBundle {
	if c == nil {
		return load()
	}
	return c.cache.Resolve(activeTheme, key, load)
}

// resolveStyle is the reference composite-widget style assembly:
// defaults -> inherited styles -> variant -> size -> active scope styles ->
// instance override. State and context layers are resolved by computed.
// Other controls should follow these four provider declaration slots.
func (b Widget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, overrides ...*theme.ButtonOverride) buttonStyle {
	activeTheme := core.ActiveTheme(ctx)
	declarations := b.declarationsForOverride(activeTheme, core.UseState[buttonDeclarationCache](ctx, key, stateSlotStyle), b.buttonOverride(overrides))
	return b.resolveStyleWithDeclarations(ctx, gtx, key, state, declarations)
}

func (b Widget) resolveStyleWithDeclarations(ctx *core.Context, gtx layout.Context, key string, state style.State, declarations buttonDeclarationBundle) buttonStyle {
	return b.resolveStyleWithDeclarationsParts(ctx, gtx, key, state, declarations, b.loading)
}

// resolvePreparedStyleWithDeclarations resolves only the parts that remain
// live during a prepared layout pass. Prepared content already contains the
// static indicator operation, so resolving Indicator again would be redundant.
func (b Widget) resolvePreparedStyleWithDeclarations(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	declarations buttonDeclarationBundle,
) buttonStyle {
	return b.resolveStyleWithDeclarationsParts(ctx, gtx, key, state, declarations, false)
}

func (b Widget) resolveStyleWithDeclarationsParts(
	ctx *core.Context,
	gtx layout.Context,
	key string,
	state style.State,
	declarations buttonDeclarationBundle,
	includeIndicator bool,
) buttonStyle {
	contentPart := style.PartLabel
	if b.iconOnly {
		contentPart = style.PartIcon
	}
	parts := [...]style.Part{style.PartRoot, contentPart, style.PartIndicator}
	partCount := 2
	if includeIndicator {
		partCount++
	}
	var resolved [len(parts)]style.ResolvedStyle
	computed.ResolvePartsInto(
		ctx,
		gtx,
		key,
		state,
		parts[:partCount],
		resolved[:partCount],
		declarations.visual.Defaults,
		declarations.visual.Variant,
		declarations.visual.Size,
		declarations.visual.Override,
	)
	return buttonStyleFrom(resolved[0], resolved[1], resolved[2], declarations.visual)
}

func (b Widget) staticStyle(ctx *core.Context, state style.State, overrides ...*theme.ButtonOverride) buttonStyle {
	activeTheme := core.ActiveTheme(ctx)
	key := core.FullKey(ctx, b.key)
	declarations := b.declarationsForOverride(activeTheme, core.UseState[buttonDeclarationCache](ctx, key, stateSlotStyle), b.buttonOverride(overrides))
	return b.staticStyleWithDeclarations(ctx, state, declarations)
}

func (b Widget) staticStyleWithDeclarations(ctx *core.Context, state style.State, declarations buttonDeclarationBundle) buttonStyle {
	return b.staticStyleWithDeclarationsParts(ctx, state, declarations, true)
}

func (b Widget) staticContentStyleWithDeclarations(
	ctx *core.Context,
	state style.State,
	declarations buttonDeclarationBundle,
) buttonStyle {
	return b.staticStyleWithDeclarationsParts(ctx, state, declarations, b.loading)
}

func (b Widget) staticStyleWithDeclarationsParts(
	ctx *core.Context,
	state style.State,
	declarations buttonDeclarationBundle,
	includeIndicator bool,
) buttonStyle {
	contentPart := style.PartLabel
	if b.iconOnly {
		contentPart = style.PartIcon
	}
	parts := [...]style.Part{style.PartRoot, contentPart, style.PartIndicator}
	partCount := 2
	if includeIndicator {
		partCount++
	}
	var resolved [len(parts)]style.ResolvedStyle
	computed.ResolvePartsStaticInto(
		ctx,
		state,
		parts[:partCount],
		resolved[:partCount],
		declarations.visual.Defaults,
		declarations.visual.Variant,
		declarations.visual.Size,
		declarations.visual.Override,
	)
	return buttonStyleFrom(resolved[0], resolved[1], resolved[2], declarations.visual)
}

func (b Widget) declarationKey() buttonDeclarationKey {
	return buttonDeclarationKey{
		request: b.buttonRequest(),
	}
}

func (b Widget) declarationsForOverride(activeTheme *theme.Theme, cache *buttonDeclarationCache, override *theme.ButtonOverride) buttonDeclarationBundle {
	key := b.declarationKey()
	var provider theme.ComponentVisualProvider
	if override != nil {
		var err error
		provider, err = theme.ComponentProvider(activeTheme)
		if err != nil {
			panic(err)
		}
	}
	load := func() buttonDeclarationBundle {
		var visual theme.ButtonVisual
		if provider != nil {
			visual = buttonVisualFromProvider(activeTheme, provider, key.request, override)
		} else {
			visual = buttonVisual(activeTheme, key.request, override)
		}
		return buttonDeclarationBundle{visual: visual}
	}
	var declarations buttonDeclarationBundle
	if override == nil {
		declarations = cache.resolve(activeTheme, key, load)
		return declarations
	}

	// Instance declarations are caller-owned. Cache them only when the
	// provider explicitly promises deterministic visuals and the declaration
	// contains no function-valued transitions that a hash cannot represent.
	if cache != nil {
		if _, stable := provider.(theme.ImmutableVisualProvider); stable && !override.Declaration.CacheUnsafe() {
			key.overridePresent = true
			key.overrideHash = override.Declaration.Hash64()
			return cache.resolve(activeTheme, key, load)
		}
	}
	return load()
}

func (b Widget) buttonRequest() theme.ButtonRequest {
	return theme.ButtonRequest{
		Variant:      theme.ButtonVariant(b.variant),
		Size:         theme.ButtonSize(b.size),
		IconOnly:     b.iconOnly,
		Grouped:      b.group.grouped,
		GroupCorners: buttonGroupCorners(b.group),
		Override:     nil,
	}
}

func buttonVisual(activeTheme *theme.Theme, request theme.ButtonRequest, override *theme.ButtonOverride) theme.ButtonVisual {
	provider, err := theme.ComponentProvider(activeTheme)
	if err != nil {
		panic(err)
	}
	return buttonVisualFromProvider(activeTheme, provider, request, override)
}

func buttonVisualFromProvider(activeTheme *theme.Theme, provider theme.ComponentVisualProvider, request theme.ButtonRequest, override *theme.ButtonOverride) theme.ButtonVisual {
	request.Override = override
	visual := provider.Button(activeTheme, request)
	if override != nil {
		// Instance declarations are intentionally uncached and must be validated
		// independently because their Part layers are caller-owned.
		theme.ValidateButtonVisual(visual)
	} else if _, stable := provider.(theme.ImmutableVisualProvider); !stable {
		// A custom provider may return closures whose captured state changes
		// while the theme pointer stays the same. Do not let the shared
		// validation-fact cache turn that mutable result into a stale contract.
		theme.ValidateButtonVisual(visual)
	} else {
		if key, cacheable := buttonVisualValidationKeyFor(provider, request); cacheable {
			buttonVisualValidationCache.Resolve(activeTheme, key, func() struct{} {
				theme.ValidateButtonVisual(visual)
				return struct{}{}
			})
		} else {
			// Immutable providers with non-comparable dynamic values cannot be
			// used safely as a map key. Keep their validation local rather than
			// risking a false cache hit or an interface-comparison panic.
			theme.ValidateButtonVisual(visual)
		}
	}
	return visual
}

func buttonVisualValidationKeyFor(provider theme.ComponentVisualProvider, request theme.ButtonRequest) (buttonVisualValidationKey, bool) {
	if provider == nil {
		return buttonVisualValidationKey{}, false
	}
	if !reflect.ValueOf(provider).Comparable() {
		return buttonVisualValidationKey{}, false
	}
	return buttonVisualValidationKey{provider: provider, request: request}, true
}

func firstButtonOverride(overrides []*theme.ButtonOverride) *theme.ButtonOverride {
	if len(overrides) == 0 {
		return nil
	}
	return overrides[0]
}

// buttonOverride resolves the override an internal call should use: the one
// passed explicitly, or the widget's own instance declaration.
func (b Widget) buttonOverride(overrides []*theme.ButtonOverride) *theme.ButtonOverride {
	if override := firstButtonOverride(overrides); override != nil {
		return override
	}
	return b.override
}

func buttonStyleFrom(resolved, content, indicator style.ResolvedStyle, visual theme.ButtonVisual) buttonStyle {
	return buttonStyle{
		root:          resolved,
		content:       content,
		indicatorPart: indicator,
		visual:        visual,
	}
}

func buttonInset(resolved style.ResolvedStyle) layout.Inset {
	if resolved.Box == nil || resolved.Box.Padding == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	padding := resolved.Box.Padding
	return layout.Inset{Top: padding.Top, Right: padding.Right, Bottom: padding.Bottom, Left: padding.Left}
}

func buttonRadius(resolved style.ResolvedStyle) unit.Dp {
	if resolved.Paint == nil || resolved.Paint.Radius == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return *resolved.Paint.Radius
}

func buttonOpacity(resolved style.ResolvedStyle) float32 {
	if resolved.Paint == nil || resolved.Paint.Opacity == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return *resolved.Paint.Opacity
}

func buttonForeground(visual theme.ButtonVisual, resolved style.ResolvedStyle) color.NRGBA {
	if visual.ForegroundColor == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	foreground, ok := visual.ForegroundColor(resolved)
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return foreground
}

func buttonRippleColor(visual theme.ButtonVisual, resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if visual.RippleColor == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return visual.RippleColor(resolved)
}

func buttonGroupCorners(group buttonGroupItemStyle) theme.ButtonGroupRequest {
	all := theme.ButtonGroupRequest{NorthWest: true, NorthEast: true, SouthEast: true, SouthWest: true}
	if !group.grouped || group.position == buttonGroupSingle {
		return all
	}
	if group.orientation == GroupVertical {
		switch group.position {
		case buttonGroupStart:
			return theme.ButtonGroupRequest{NorthWest: true, NorthEast: true}
		case buttonGroupEnd:
			return theme.ButtonGroupRequest{SouthEast: true, SouthWest: true}
		default:
			return theme.ButtonGroupRequest{}
		}
	}
	switch group.position {
	case buttonGroupStart:
		return theme.ButtonGroupRequest{NorthWest: true, SouthWest: true}
	case buttonGroupEnd:
		return theme.ButtonGroupRequest{NorthEast: true, SouthEast: true}
	default:
		return theme.ButtonGroupRequest{}
	}
}

func buttonLoadingStyle(gtx layout.Context, resolved buttonStyle) buttonStyle {
	if resolved.visual.LoadingInset == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	if resolved.root.Box == nil || resolved.root.Box.Padding == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	loadingInset := resolved.visual.LoadingInset(gtx, theme.ButtonLoadingRequest{
		Inset: buttonInset(resolved.root),
	})
	theme.ValidateButtonLoadingInset(loadingInset)
	// ResolveStatic may return cache-owned immutable data. Only the box is
	// changed here, so a shallow copy is sufficient for copy-on-write.
	boxCopy := *resolved.root.Box
	resolved.root.Box = &boxCopy
	padding := *boxCopy.Padding
	padding.Left = loadingInset.Left
	padding.Right = loadingInset.Right
	resolved.root.Box.Padding = &padding
	return resolved
}
