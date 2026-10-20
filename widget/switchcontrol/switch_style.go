package switchcontrol

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type switchStyle struct {
	trackOff    style.ResolvedStyle
	trackOn     style.ResolvedStyle
	thumbOff    style.ResolvedStyle
	thumbOn     style.ResolvedStyle
	label       style.ResolvedStyle
	description style.ResolvedStyle
	// Keep the provider bundle intact so runtime geometry and motion never
	// drift into a second set of visual fields.
	visual   theme.SwitchVisual
	selected float32
	focus    float32
}

const switchResolvedPartCacheLimit = 32

type switchResolvedPartKey struct {
	part  style.Part
	state style.StateMask
}

// switchResolvedPartCache retains only settled Part snapshots. Transition
// state remains owned by the widget's derived animation keys, so a cached
// value can never carry one switch instance's animation into another.
type switchResolvedPartCache struct {
	theme           *theme.Theme
	themeGeneration uint64
	scopeHash       uint64
	size            SwitchSize
	enabled         bool
	values          map[switchResolvedPartKey]style.ResolvedStyle
	hits            uint64
	misses          uint64
	evictions       uint64
	themeInvalid    uint64
	scopeInvalid    uint64
	sizeInvalid     uint64
}

// VisualCache retains provider-resolved switch visuals for repeated layouts.
// Instance overrides deliberately bypass this cache because they are scoped to
// one widget instance rather than the size/theme declaration.
type VisualCache struct {
	providerCache computed.ProviderCache[SwitchSize, theme.SwitchVisual]
	resolved      switchResolvedPartCache
}

// VisualCacheStats exposes cache behavior for compound-widget benchmarks and
// regression tests without exposing declaration-cache internals.
type VisualCacheStats struct {
	Hits                       uint64
	Misses                     uint64
	ThemeInvalidations         uint64
	ResolvedHits               uint64
	ResolvedMisses             uint64
	ResolvedEvictions          uint64
	ResolvedThemeInvalidations uint64
	ResolvedScopeInvalidations uint64
	ResolvedSizeInvalidations  uint64
}

// Stats returns a snapshot of cache activity.
func (c *VisualCache) Stats() VisualCacheStats {
	if c == nil {
		return VisualCacheStats{}
	}
	provider := c.providerCache.Stats()
	resolved := c.resolved
	return VisualCacheStats{
		Hits:                       provider.Hits,
		Misses:                     provider.Misses,
		ThemeInvalidations:         provider.ThemeInvalidations,
		ResolvedHits:               resolved.hits,
		ResolvedMisses:             resolved.misses,
		ResolvedEvictions:          resolved.evictions,
		ResolvedThemeInvalidations: resolved.themeInvalid,
		ResolvedScopeInvalidations: resolved.scopeInvalid,
		ResolvedSizeInvalidations:  resolved.sizeInvalid,
	}
}

// Resolve returns the provider visual for size, reusing the active theme's
// declaration when possible.
func (c *VisualCache) Resolve(activeTheme *theme.Theme, size SwitchSize) theme.SwitchVisual {
	if c == nil {
		return resolveSwitchVisual(activeTheme, size, nil)
	}
	return c.providerCache.Resolve(activeTheme, size, func() theme.SwitchVisual {
		return resolveSwitchVisual(activeTheme, size, nil)
	})
}

func (c *VisualCache) prepareResolved(
	ctx *core.Context,
	activeTheme *theme.Theme,
	size SwitchSize,
	visual theme.SwitchVisual,
) {
	if c == nil {
		return
	}
	cache := &c.resolved
	enabled := switchStyleLayersCacheSafe(ctx, visual)
	scopeHash := uint64(0)
	if enabled {
		scopeHash = switchStyleScopeHash(ctx, visual)
	}
	generation := core.ThemeGeneration(ctx)
	if cache.theme != activeTheme || cache.themeGeneration != generation ||
		cache.scopeHash != scopeHash || cache.size != size || cache.enabled != enabled {
		if cache.theme != nil {
			switch {
			case cache.theme != activeTheme:
				cache.themeInvalid++
			case cache.size != size:
				cache.sizeInvalid++
			default:
				cache.scopeInvalid++
			}
		}
		cache.theme = activeTheme
		cache.themeGeneration = generation
		cache.scopeHash = scopeHash
		cache.size = size
		cache.enabled = enabled
		clear(cache.values)
	}
}

func (c *VisualCache) resolvedPart(part style.Part, state style.State) (style.ResolvedStyle, bool) {
	if c == nil {
		return style.ResolvedStyle{}, false
	}
	cache := &c.resolved
	if !cache.enabled {
		return style.ResolvedStyle{}, false
	}
	key := switchResolvedPartKey{part: part, state: state.Bits()}
	value, ok := cache.values[key]
	if ok {
		cache.hits++
		return value, true
	}
	cache.misses++
	return style.ResolvedStyle{}, false
}

func (c *VisualCache) storeResolvedPart(part style.Part, state style.State, value style.ResolvedStyle) {
	if c == nil || !c.resolved.enabled || len(value.Transitions) != 0 {
		return
	}
	cache := &c.resolved
	key := switchResolvedPartKey{part: part, state: state.Bits()}
	if _, exists := cache.values[key]; exists {
		cache.values[key] = value
		return
	}
	if len(cache.values) >= switchResolvedPartCacheLimit {
		clear(cache.values)
		cache.evictions++
	}
	if cache.values == nil {
		cache.values = make(map[switchResolvedPartKey]style.ResolvedStyle, switchResolvedPartCacheLimit)
	}
	cache.values[key] = value
}

func switchStyleLayersCacheSafe(ctx *core.Context, visual theme.SwitchVisual) bool {
	if visual.Defaults.CacheUnsafe() || visual.Override.CacheUnsafe() {
		return false
	}
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	return true
}

func switchStyleScopeHash(ctx *core.Context, visual theme.SwitchVisual) uint64 {
	const (
		offset64 = uint64(14695981039346656037)
		prime64  = uint64(1099511628211)
	)
	hash := offset64
	mix := func(value uint64) {
		for range 8 {
			hash ^= value & 0xff
			hash *= prime64
			value >>= 8
		}
	}
	mix(visual.Defaults.Hash64())
	mix(visual.Override.Hash64())
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		mix(value.Hash64())
	}
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		mix(value.Hash64())
	}
	return hash
}

func (s Widget) resolveStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, override *theme.SwitchOverride, caches ...*VisualCache) switchStyle {
	activeTheme := core.ActiveTheme(ctx)
	visual := switchVisualCached(activeTheme, s.size, override, caches)
	if override == nil && len(caches) > 0 && caches[0] != nil {
		caches[0].prepareResolved(ctx, activeTheme, s.size, visual)
	}
	defaults := visual.Defaults
	offState := state
	offState.Checked, offState.Selected = false, false
	onState := state
	onState.Checked, onState.Selected = true, true

	resolvePart := func(part style.Part, partState style.State) style.ResolvedStyle {
		if override == nil && len(caches) > 0 && caches[0] != nil {
			if value, ok := caches[0].resolvedPart(part, partState); ok {
				return value
			}
			value := computed.ResolvePartStatic(ctx, part, partState, defaults, style.Style{}, style.Style{}, visual.Override)
			caches[0].storeResolvedPart(part, partState, value)
			return value
		}
		return computed.ResolvePartStatic(ctx, part, partState, defaults, style.Style{}, style.Style{}, visual.Override)
	}
	trackOff := resolvePart(style.PartTrack, offState)
	trackOn := resolvePart(style.PartTrack, onState)
	thumbOff := resolvePart(style.PartThumb, offState)
	thumbOn := resolvePart(style.PartThumb, onState)
	var label, description style.ResolvedStyle
	if s.label != "" {
		label = resolvePart(style.PartLabel, state)
	}
	if s.description != "" {
		description = resolvePart(style.PartDescription, state)
	}
	offOwner := core.DerivedKey(ctx, key, "switch-off")
	onOwner := core.DerivedKey(ctx, key, "switch-on")
	trackOff = computed.ApplyPartTransitions(ctx, gtx, offOwner, style.PartTrack, trackOff)
	trackOn = computed.ApplyPartTransitions(ctx, gtx, onOwner, style.PartTrack, trackOn)
	thumbOff = computed.ApplyPartTransitions(ctx, gtx, offOwner, style.PartThumb, thumbOff)
	thumbOn = computed.ApplyPartTransitions(ctx, gtx, onOwner, style.PartThumb, thumbOn)
	label = computed.ApplyPartTransitions(ctx, gtx, key, style.PartLabel, label)
	description = computed.ApplyPartTransitions(ctx, gtx, key, style.PartDescription, description)

	resolved := switchStyle{
		trackOff: trackOff, trackOn: trackOn,
		thumbOff: thumbOff, thumbOn: thumbOn,
		label: label, description: description, visual: visual,
	}
	return resolved
}

func switchVisualCached(activeTheme *theme.Theme, size SwitchSize, override *theme.SwitchOverride, caches []*VisualCache) theme.SwitchVisual {
	if override == nil && len(caches) > 0 && caches[0] != nil {
		return caches[0].Resolve(activeTheme, size)
	}
	return resolveSwitchVisual(activeTheme, size, override)
}

func resolveSwitchVisual(activeTheme *theme.Theme, size SwitchSize, override *theme.SwitchOverride) theme.SwitchVisual {
	return theme.ResolveVisual(activeTheme, theme.SwitchRequest{Size: size, Override: override},
		theme.ComponentVisualProvider.Switch, theme.ValidateSwitchVisual)
}
