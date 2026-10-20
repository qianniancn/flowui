package collapsible

import (
	"errors"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type recordingCollapsibleProvider struct {
	theme.ComponentVisualProvider
	calls    int
	requests []theme.CollapsibleRequest
}

type incompleteCollapsibleProvider struct {
	theme.ComponentVisualProvider
}

func (p *recordingCollapsibleProvider) Collapsible(active *theme.Theme, request theme.CollapsibleRequest) theme.CollapsibleVisual {
	p.calls++
	p.requests = append(p.requests, request)
	return p.ComponentVisualProvider.Collapsible(active, request)
}

func (p incompleteCollapsibleProvider) Collapsible(active *theme.Theme, request theme.CollapsibleRequest) theme.CollapsibleVisual {
	visual := p.ComponentVisualProvider.Collapsible(active, request)
	visual.Indicator = nil
	return visual
}

func collapsibleProviderContext(active *theme.Theme) *core.Context {
	return core.New(nil, active, locale.LanguageAuto)
}

func TestCollapsibleVisualCacheTracksThemeLifecycle(t *testing.T) {
	active := testtheme.Behavior()
	provider := &recordingCollapsibleProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	active.VisualProvider = provider
	state := new(collapsibleState)
	ctx := collapsibleProviderContext(active)

	state.visual(ctx, nil)
	state.visual(ctx, nil)
	if provider.calls != 1 {
		t.Fatalf("collapsible provider calls = %d, want 1", provider.calls)
	}
	if stats := state.visualCache.Stats(); stats.Hits != 1 || stats.Misses != 1 || stats.ThemeInvalidations != 0 {
		t.Fatalf("collapsible visual cache stats = %#v, want one hit and one miss", stats)
	}

	replacement := *active
	state.visual(collapsibleProviderContext(&replacement), nil)
	if provider.calls != 2 {
		t.Fatalf("theme replacement provider calls = %d, want 2", provider.calls)
	}
	if stats := state.visualCache.Stats(); stats.ThemeInvalidations != 1 || stats.Misses != 2 {
		t.Fatalf("theme replacement cache stats = %#v, want one invalidation and two misses", stats)
	}
}

func TestCollapsibleInstanceOverrideBypassesVisualCache(t *testing.T) {
	active := testtheme.Behavior()
	provider := &recordingCollapsibleProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	active.VisualProvider = provider
	state := new(collapsibleState)
	ctx := collapsibleProviderContext(active)
	declaration := style.Style{}.PaddingX(9)
	override := &theme.CollapsibleOverride{Declaration: declaration}

	state.visual(ctx, nil)
	state.visual(ctx, override)
	state.visual(ctx, override)
	if provider.calls != 3 {
		t.Fatalf("override provider calls = %d, want 3", provider.calls)
	}
	if len(provider.requests) != 3 || provider.requests[1].Override == nil || provider.requests[2].Override == nil {
		t.Fatal("collapsible provider did not receive both instance overrides")
	}
	if provider.requests[1].Override.Declaration.Hash64() != declaration.Hash64() {
		t.Fatalf("override hash = %x, want %x", provider.requests[1].Override.Declaration.Hash64(), declaration.Hash64())
	}
	if stats := state.visualCache.Stats(); stats.Hits != 0 || stats.Misses != 1 {
		t.Fatalf("override unexpectedly populated or hit shared cache: %#v", stats)
	}
}

func TestCollapsibleIncompleteVisualDoesNotPolluteCache(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = incompleteCollapsibleProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	state := new(collapsibleState)
	ctx := collapsibleProviderContext(active)

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete collapsible panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := state.visualCache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete collapsible polluted cache = %#v", stats)
		}
	}()
	state.visual(ctx, nil)
}

func TestCollapsibleMissingProviderFailsExplicitly(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := collapsibleProviderContext(&active)
	state := new(collapsibleState)
	defer func() {
		if recovered := recover(); recovered != theme.ErrMissingVisualProvider {
			t.Fatalf("missing collapsible provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
		}
	}()
	state.visual(ctx, nil)
}

func TestCollapsibleStyledWrapperKeepsVisualDeclarationsOutsideBehaviorValue(t *testing.T) {
	value := New("Styled", nil).Key("styled")
	root := style.Style{}.PaddingX(9)
	trigger := style.Style{}.PaddingY(4)
	wrapper := value.StyleWithTrigger(root, trigger)

	if wrapper.key != value.key || wrapper.label != value.label {
		t.Fatal("Style changed the behavior value")
	}
	if wrapper.override.Declaration.Hash64() != root.Hash64() {
		t.Fatalf("root declaration hash = %x, want %x", wrapper.override.Declaration.Hash64(), root.Hash64())
	}
	if wrapper.override.Trigger.Hash64() != trigger.Hash64() {
		t.Fatalf("trigger declaration hash = %x, want %x", wrapper.override.Trigger.Hash64(), trigger.Hash64())
	}

	group := Group("group", nil, nil).StyleWithTrigger(root, trigger)
	if group.override.Declaration.Hash64() != root.Hash64() || group.override.Trigger.Hash64() != trigger.Hash64() {
		t.Fatal("group wrapper did not retain provider-bound declarations")
	}
}

func TestCollapsibleProviderReturnsRootAndTriggerOverrides(t *testing.T) {
	active := testtheme.Behavior()
	provider := active.VisualProvider.(theme.CollapsibleProvider)
	root := style.Style{}.PaddingX(7)
	trigger := style.Style{}.PaddingY(3)
	visual := provider.Collapsible(active, theme.CollapsibleRequest{
		Override: &theme.CollapsibleOverride{Declaration: root, Trigger: trigger},
	})
	if visual.Override.Hash64() != root.Hash64() {
		t.Fatalf("root override hash = %x, want %x", visual.Override.Hash64(), root.Hash64())
	}
	if visual.Trigger.Hash64() != trigger.Hash64() {
		t.Fatalf("trigger override hash = %x, want %x", visual.Trigger.Hash64(), trigger.Hash64())
	}
}
