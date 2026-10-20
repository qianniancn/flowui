package togglebutton

import (
	"errors"
	"image/color"
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type countingToggleProvider struct {
	theme.ComponentVisualProvider
	calls *int
}

func (p countingToggleProvider) ToggleButton(active *theme.Theme, request theme.ToggleRequest) theme.ToggleVisual {
	(*p.calls)++
	return p.ComponentVisualProvider.ToggleButton(active, request)
}

type incompleteToggleProvider struct {
	theme.ComponentVisualProvider
}

func (p incompleteToggleProvider) ToggleButton(*theme.Theme, theme.ToggleRequest) theme.ToggleVisual {
	return theme.ToggleVisual{}
}

func TestVisualCacheTracksRequestKeyAndThemeInvalidation(t *testing.T) {
	active := testtheme.Behavior()
	calls := 0
	active.VisualProvider = countingToggleProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}
	var cache VisualCache
	request := theme.ToggleRequest{
		Variant: theme.ToggleButtonDefault,
		Size:    theme.ToggleButtonMedium,
		Foreground: color.NRGBA{
			R: 0x10, G: 0x20, B: 0x30, A: 0xff,
		},
	}

	cache.Resolve(active, request)
	cache.Resolve(active, request)
	request.Size = theme.ToggleButtonLarge
	cache.Resolve(active, request)
	request.IconOnly = true
	cache.Resolve(active, request)
	request.Foreground.R++
	cache.Resolve(active, request)

	if calls != 4 {
		t.Fatalf("provider calls = %d, want one per distinct request", calls)
	}
	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 4 || stats.ThemeInvalidations != 0 || stats.KeyInvalidations != 3 {
		t.Fatalf("warm cache stats = %#v", stats)
	}

	replacement := *active
	cache.Resolve(&replacement, request)
	stats = cache.Stats()
	if calls != 5 || stats.Misses != 5 || stats.ThemeInvalidations != 1 || stats.KeyInvalidations != 3 {
		t.Fatalf("theme invalidation stats = %#v, calls=%d", stats, calls)
	}
}

func TestToggleInstanceOverrideBypassesVisualCache(t *testing.T) {
	active := testtheme.Behavior()
	calls := 0
	active.VisualProvider = countingToggleProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}
	var cache VisualCache
	request := theme.ToggleRequest{Variant: theme.ToggleButtonDefault, Size: theme.ToggleButtonMedium}
	cache.Resolve(active, request)
	before := cache.Stats()
	override := &theme.ToggleOverride{Declaration: style.Style{}.Height(91)}
	visual := toggleVisualCached(active, Widget{}, color.NRGBA{}, override, []*VisualCache{&cache})
	resolved := style.Cascade(style.State{}, visual.Override)
	if resolved.Box == nil || resolved.Box.Height == nil || *resolved.Box.Height != 91 {
		t.Fatalf("provider did not receive instance override: %#v", resolved)
	}
	if got := cache.Stats(); got != before {
		t.Fatalf("override changed visual cache stats: before=%#v after=%#v", before, got)
	}
	if calls != 2 {
		t.Fatalf("provider calls = %d, want cached request plus override request", calls)
	}
}

func TestIncompleteToggleVisualDoesNotPolluteCache(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = incompleteToggleProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
	}
	var cache VisualCache
	request := theme.ToggleRequest{Variant: theme.ToggleButtonDefault, Size: theme.ToggleButtonMedium}

	func() {
		defer func() {
			recovered := recover()
			if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
				t.Fatalf("incomplete toggle visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
			}
		}()
		cache.Resolve(active, request)
	}()
	if stats := cache.Stats(); stats.Misses != 1 || stats.Size != 0 {
		t.Fatalf("failed toggle visual changed cache state: %#v", stats)
	}

	active.VisualProvider = testtheme.BehaviorProvider()
	cache.Resolve(active, request)
	if stats := cache.Stats(); stats.Misses != 2 || stats.Size != 1 {
		t.Fatalf("valid toggle visual did not recover cache: %#v", stats)
	}
}

func TestToggleLayoutResolvesProviderOncePerFrame(t *testing.T) {
	active := testtheme.Behavior()
	calls := 0
	active.VisualProvider = countingToggleProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	router := new(input.Router)
	value := New("").Key("once").Selected(false).Content(nil)
	layoutToggleButtonFrame(ctx, router, value, time.Unix(1, 0))
	if calls != 1 {
		t.Fatalf("provider calls during one layout = %d, want 1", calls)
	}
	layoutToggleButtonFrame(ctx, router, value, time.Unix(1, 0).Add(time.Millisecond))
	if calls != 1 {
		t.Fatalf("provider calls after warm layout = %d, want 1", calls)
	}
}
