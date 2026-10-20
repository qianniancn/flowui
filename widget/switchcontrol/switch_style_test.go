package switchcontrol

import (
	"testing"
	"time"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestSwitchVisualCacheTracksSizeAndThemeInvalidation(t *testing.T) {
	activeTheme := testtheme.Behavior()
	var cache VisualCache

	cache.Resolve(activeTheme, SwitchMedium)
	cache.Resolve(activeTheme, SwitchMedium)
	cache.Resolve(activeTheme, SwitchLarge)

	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.ThemeInvalidations != 0 {
		t.Fatalf("warm cache stats = %#v", stats)
	}

	replaced := *activeTheme
	cache.Resolve(&replaced, SwitchMedium)
	stats = cache.Stats()
	if stats.Hits != 1 || stats.Misses != 3 || stats.ThemeInvalidations != 1 {
		t.Fatalf("theme invalidation stats = %#v", stats)
	}
}

func TestSwitchVisualCachePanicDoesNotDiscardPreviousValue(t *testing.T) {
	active := testtheme.Behavior()
	missing := theme.NeutralTheme()
	var cache VisualCache

	cache.Resolve(active, SwitchMedium)
	func() {
		defer func() {
			if recovered := recover(); recovered != theme.ErrMissingVisualProvider {
				t.Fatalf("missing provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
			}
		}()
		cache.Resolve(&missing, SwitchMedium)
	}()

	cache.Resolve(active, SwitchMedium)
	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 || stats.ThemeInvalidations != 1 {
		t.Fatalf("panic cache stats = %#v, want one hit, one miss, one theme invalidation", stats)
	}
}

func TestSwitchInstanceOverrideBypassesVisualCache(t *testing.T) {
	activeTheme := testtheme.Behavior()
	var cache VisualCache
	cache.Resolve(activeTheme, SwitchMedium)
	before := cache.Stats()

	_ = switchVisualCached(
		activeTheme,
		SwitchMedium,
		&theme.SwitchOverride{Declaration: style.Style{}.Width(30)},
		[]*VisualCache{&cache},
	)

	if got := cache.Stats(); got != before {
		t.Fatalf("override changed visual cache stats: before=%#v after=%#v", before, got)
	}
}

func TestSwitchResolvedPartCacheTracksStateScopeSizeAndTheme(t *testing.T) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageEnglish)
	var cache VisualCache
	visual := cache.Resolve(active, SwitchMedium)
	state := style.State{}
	value := style.ResolvedStyle{}
	cache.prepareResolved(ctx, active, SwitchMedium, visual)

	if _, ok := cache.resolvedPart(style.PartTrack, state); ok {
		t.Fatal("first resolved part lookup unexpectedly hit")
	}
	cache.storeResolvedPart(style.PartTrack, state, value)
	if _, ok := cache.resolvedPart(style.PartTrack, state); !ok {
		t.Fatal("settled resolved part did not hit")
	}
	if _, ok := cache.resolvedPart(style.PartTrack, style.State{Checked: true}); ok {
		t.Fatal("checked state reused unchecked resolved part")
	}

	cache.storeResolvedPart(style.PartTrack, style.State{Checked: true}, value)
	cache.prepareResolved(ctx, active, SwitchLarge, visual)
	if _, ok := cache.resolvedPart(style.PartTrack, state); ok {
		t.Fatal("size change preserved resolved part")
	}
	if got := cache.Stats().ResolvedSizeInvalidations; got != 1 {
		t.Fatalf("size invalidations = %d, want 1", got)
	}

	restore := core.PushStyle(ctx, style.Style{}.TextColor(style.RGB(0x102030)))
	cache.prepareResolved(ctx, active, SwitchLarge, visual)
	if _, ok := cache.resolvedPart(style.PartTrack, state); ok {
		t.Fatal("style scope change preserved resolved part")
	}
	restore()

	other := testtheme.Behavior()
	core.ReplaceTheme(ctx, *other)
	cache.prepareResolved(ctx, other, SwitchLarge, visual)
	if _, ok := cache.resolvedPart(style.PartTrack, state); ok {
		t.Fatal("theme replacement preserved resolved part")
	}
	stats := cache.Stats()
	if stats.ResolvedScopeInvalidations == 0 || stats.ResolvedThemeInvalidations == 0 {
		t.Fatalf("resolved invalidation stats = %#v", stats)
	}
}

func TestSwitchResolvedPartCacheBypassesTransitions(t *testing.T) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageEnglish)
	var cache VisualCache
	visual := cache.Resolve(active, SwitchMedium)
	state := style.State{}
	cache.prepareResolved(ctx, active, SwitchMedium, visual)

	transition := style.ResolvedStyle{Transitions: []style.Transition{{Property: style.PropOpacity, Duration: time.Millisecond}}}
	cache.resolvedPart(style.PartTrack, state)
	cache.storeResolvedPart(style.PartTrack, state, transition)
	if _, ok := cache.resolvedPart(style.PartTrack, state); ok {
		t.Fatal("transition-bearing resolved part was cached")
	}

	restore := core.PushStyle(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond, style.TransitionEase(func(progress float32) float32 {
		return progress * progress
	})))
	cache.prepareResolved(ctx, active, SwitchMedium, visual)
	if _, ok := cache.resolvedPart(style.PartTrack, state); ok {
		t.Fatal("custom easing scope unexpectedly enabled resolved cache")
	}
	restore()
}
