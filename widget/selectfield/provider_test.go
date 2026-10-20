package selectfield

import (
	"errors"
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type testSelectVisualProvider struct {
	theme.ComponentVisualProvider
}

func (p testSelectVisualProvider) Select(active *theme.Theme, request theme.SelectRequest) theme.SelectVisual {
	visual := p.ComponentVisualProvider.Select(active, request)
	visual.Defaults = style.Join(visual.Defaults, style.Style{}.MinHeight(unit.Dp(77)))
	visual.Override = theme.SelectOverrideStyle(request)
	return visual
}

func newTestSelectVisualProvider() testSelectVisualProvider {
	return testSelectVisualProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
}

func TestSelectUsesActiveVisualProvider(t *testing.T) {
	active := theme.NeutralTheme()
	active.VisualProvider = newTestSelectVisualProvider()

	bundle := selectVisualBundle(&active, SelectPrimary, nil)
	resolved := bundle.visual.Defaults.Resolve(style.State{})
	if resolved.Box == nil || resolved.Box.MinHeight == nil || *resolved.Box.MinHeight != 77 {
		t.Fatalf("provider min height = %#v, want 77", resolved.Box)
	}
}

type recordingSelectVisualProvider struct {
	testSelectVisualProvider
	requests []theme.SelectRequest
}

type incompleteSelectVisualProvider struct {
	theme.ComponentVisualProvider
	visual theme.SelectVisual
}

func (p incompleteSelectVisualProvider) Select(*theme.Theme, theme.SelectRequest) theme.SelectVisual {
	return p.visual
}

func (p *recordingSelectVisualProvider) Select(active *theme.Theme, request theme.SelectRequest) theme.SelectVisual {
	p.requests = append(p.requests, request)
	return p.testSelectVisualProvider.Select(active, request)
}

func TestSelectInstanceOverrideUsesProviderLayer(t *testing.T) {
	provider := &recordingSelectVisualProvider{testSelectVisualProvider: newTestSelectVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	declaration := style.Style{}.BorderWidth(2)
	override := theme.SelectOverride{Declaration: declaration}
	resolved := selectVisualBundle(&active, SelectPrimary, &override).visual.Override.Resolve(style.State{})
	if resolved.Paint == nil || resolved.Paint.Border == nil || resolved.Paint.Border.Width == nil || *resolved.Paint.Border.Width != 2 {
		t.Fatalf("provider override = %#v, want border width 2", resolved.Paint)
	}
	if len(provider.requests) != 1 || provider.requests[0].Override == nil {
		t.Fatal("select provider did not receive the instance override")
	}
}

func TestSelectWithoutVisualProviderPanics(t *testing.T) {
	active := theme.NeutralTheme()
	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		if !ok || !errors.Is(err, theme.ErrMissingVisualProvider) {
			t.Fatalf("missing provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
		}
	}()
	selectVisualBundle(&active, SelectPrimary, nil)
}

func TestSelectIncompleteVisualProviderPanicsBeforeCacheCommit(t *testing.T) {
	active := theme.NeutralTheme()
	base := newTestSelectVisualProvider()
	visual := base.ComponentVisualProvider.Select(&active, theme.SelectRequest{Variant: theme.InputPrimary})
	visual.Indicator = nil
	active.VisualProvider = incompleteSelectVisualProvider{
		ComponentVisualProvider: base.ComponentVisualProvider,
		visual:                  visual,
	}
	cache := computed.ProviderCache[SelectVariant, selectDeclarationBundle]{}
	defer func() {
		recovered := recover()
		if !errors.Is(asError(recovered), theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete select visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Misses != 0 || stats.Hits != 0 {
			t.Fatalf("incomplete select visual changed cache stats = %#v", stats)
		}
	}()
	selectStyleDeclarationsCached(&active, SelectPrimary, &cache)
}

func TestSelectMissingColorDoesNotUseFallback(t *testing.T) {
	active := theme.NeutralTheme()
	visual := newTestSelectVisualProvider().Select(&active, theme.SelectRequest{})
	visual.Text = nil
	defer func() {
		if recovered := recover(); !errors.Is(asError(recovered), theme.ErrIncompleteVisualProvider) {
			t.Fatalf("missing select color panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
	}()
	theme.ValidateSelectVisual(visual)
}

func asError(value any) error {
	if err, ok := value.(error); ok {
		return err
	}
	return nil
}

func TestSelectOverlayUsesResolvedBundleOnce(t *testing.T) {
	provider := &recordingSelectVisualProvider{testSelectVisualProvider: newTestSelectVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)
	state := new(selectState)
	bundle := selectStyleDeclarationsCached(&active, SelectPrimary, &state.styleCache)
	widget := New([]SelectItem{{Key: "one", Label: "One"}}).Key("overlay-select").SelectedKey("one")
	widget.layoutOverlayWithBundle(ctx, testLayoutContext(), state, image.Rect(0, 0, 180, 36), true, 1, true, bundle)
	if got := len(provider.requests); got != 1 {
		t.Fatalf("select overlay provider requests = %d, want one resolved bundle", got)
	}
}

func TestStyledSelectUsesOneProviderBundlePerFrame(t *testing.T) {
	provider := &recordingSelectVisualProvider{testSelectVisualProvider: newTestSelectVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)
	(New([]SelectItem{{Key: "one", Label: "One"}}).Key("styled-select")).Style(style.Style{}.BorderWidth(2)).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(180, 48)},
		Ops:         new(op.Ops),
	})
	if got := len(provider.requests); got != 1 {
		t.Fatalf("styled select provider requests = %d, want one cached visual bundle", got)
	}
}

func TestSelectOverrideBypassesBaseProviderCache(t *testing.T) {
	provider := &recordingSelectVisualProvider{testSelectVisualProvider: newTestSelectVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	cache := computed.ProviderCache[SelectVariant, selectDeclarationBundle]{}

	selectStyleDeclarationsCached(&active, SelectPrimary, &cache)
	selectStyleDeclarationsCached(&active, SelectPrimary, &cache)
	if len(provider.requests) != 1 {
		t.Fatalf("base select bundle requests = %d, want 1", len(provider.requests))
	}
	emptyOverride := &theme.SelectOverride{}
	selectVisualBundle(&active, SelectPrimary, emptyOverride)
	if len(provider.requests) != 2 {
		t.Fatalf("empty select override requests = %d, want 2", len(provider.requests))
	}
	selectVisualBundle(&active, SelectPrimary, emptyOverride)
	if len(provider.requests) != 3 {
		t.Fatalf("repeated empty select override requests = %d, want 3", len(provider.requests))
	}
	selectStyleDeclarationsCached(&active, SelectPrimary, &cache)
	if got := cache.Stats(); got.Hits != 2 || got.Misses != 1 {
		t.Fatalf("base select cache stats = %#v, want 2 hits and 1 miss", got)
	}
}
