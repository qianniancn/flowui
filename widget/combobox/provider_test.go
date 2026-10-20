package combobox

import (
	"errors"
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type testComboVisualProvider struct {
	theme.ComponentVisualProvider
}

func (p testComboVisualProvider) Combo(active *theme.Theme, request theme.ComboRequest) theme.ComboVisual {
	visual := p.ComponentVisualProvider.Combo(active, request)
	visual.Defaults = style.Join(visual.Defaults, style.Style{}.Height(unit.Dp(79)))
	visual.Override = theme.ComboOverrideStyle(request)
	return visual
}

func newTestComboVisualProvider() testComboVisualProvider {
	return testComboVisualProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
}

type incompleteComboVisualProvider struct {
	theme.ComponentVisualProvider
	visual theme.ComboVisual
}

func (p incompleteComboVisualProvider) Combo(*theme.Theme, theme.ComboRequest) theme.ComboVisual {
	return p.visual
}

func TestComboUsesActiveVisualProvider(t *testing.T) {
	active := theme.NeutralTheme()
	active.VisualProvider = newTestComboVisualProvider()

	bundle := comboVisualBundle(&active, field.Primary, nil)
	resolved := bundle.visual.Defaults.Resolve(style.State{})
	if resolved.Box == nil || resolved.Box.Height == nil || *resolved.Box.Height != 79 {
		t.Fatalf("provider height = %#v, want 79", resolved.Box)
	}
}

func TestComboIncompleteVisualProviderPanicsBeforeCacheCommit(t *testing.T) {
	active := theme.NeutralTheme()
	base := newTestComboVisualProvider()
	visual := base.ComponentVisualProvider.Combo(&active, theme.ComboRequest{Variant: theme.InputPrimary})
	visual.Check = nil
	active.VisualProvider = incompleteComboVisualProvider{
		ComponentVisualProvider: base.ComponentVisualProvider,
		visual:                  visual,
	}
	cache := computed.ProviderCache[field.Variant, comboDeclarationBundle]{}
	defer func() {
		recovered := recover()
		if err, ok := recovered.(error); !ok || !errors.Is(err, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete combo visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Misses != 0 || stats.Hits != 0 {
			t.Fatalf("incomplete combo visual changed cache stats = %#v", stats)
		}
	}()
	comboStyleDeclarationsCached(&active, field.Primary, &cache)
}

func TestComboMissingColorDoesNotUseFallback(t *testing.T) {
	active := theme.NeutralTheme()
	visual := newTestComboVisualProvider().Combo(&active, theme.ComboRequest{})
	visual.Text = nil
	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		if !ok || !errors.Is(err, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("missing combo color panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
	}()
	theme.ValidateComboVisual(visual)
}

type recordingComboVisualProvider struct {
	testComboVisualProvider
	requests []theme.ComboRequest
}

func (p *recordingComboVisualProvider) Combo(active *theme.Theme, request theme.ComboRequest) theme.ComboVisual {
	p.requests = append(p.requests, request)
	return p.testComboVisualProvider.Combo(active, request)
}

func TestComboInstanceOverrideUsesProviderLayer(t *testing.T) {
	provider := &recordingComboVisualProvider{testComboVisualProvider: newTestComboVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	declaration := style.Style{}.BorderWidth(2)
	override := theme.ComboOverride{Declaration: declaration}
	resolved := comboVisualBundle(&active, field.Primary, &override).visual.Override.Resolve(style.State{})
	if resolved.Paint == nil || resolved.Paint.Border == nil || resolved.Paint.Border.Width == nil || *resolved.Paint.Border.Width != 2 {
		t.Fatalf("provider override = %#v, want border width 2", resolved.Paint)
	}
	if len(provider.requests) != 1 || provider.requests[0].Override == nil {
		t.Fatal("combo provider did not receive the instance override")
	}
}

func TestComboStyledWrapperKeepsOverrideOutsideBehaviorValue(t *testing.T) {
	provider := &recordingComboVisualProvider{testComboVisualProvider: newTestComboVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)
	base := New(comboBoxTestItems()).Key("wrapped")
	wrapped := (base).Style(style.Style{}.BorderWidth(2))

	wrapped.Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(180, 48)},
		Ops:         new(op.Ops),
	})

	for _, request := range provider.requests {
		if request.Override != nil {
			return
		}
	}
	t.Fatal("styled wrapper did not send an instance override to the provider")
}

func TestComboWithoutVisualProviderPanics(t *testing.T) {
	active := theme.NeutralTheme()
	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		if !ok || !errors.Is(err, theme.ErrMissingVisualProvider) {
			t.Fatalf("missing provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
		}
	}()
	comboVisualBundle(&active, field.Primary, nil)
}

func TestComboOverlayUsesResolvedBundleOnce(t *testing.T) {
	provider := &recordingComboVisualProvider{testComboVisualProvider: newTestComboVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)
	state := new(comboBoxState)
	bundle := comboStyleDeclarationsCached(&active, field.Primary, &state.styleCache)
	combo := New([]Item{{Key: "one", Label: "One"}}).Key("overlay-combo").SelectedKey("one")
	combo.layoutPanelOverlayWithBundle(
		ctx,
		testLayoutContext(),
		testLayoutContext(),
		state,
		&state.editor,
		[]int{0},
		image.Rect(0, 0, 180, 36),
		1,
		true,
		style.Style{},
		bundle,
	)
	if got := len(provider.requests); got != 1 {
		t.Fatalf("combo overlay provider requests = %d, want one resolved bundle", got)
	}
}

func TestStyledComboUsesOneProviderBundlePerFrame(t *testing.T) {
	provider := &recordingComboVisualProvider{testComboVisualProvider: newTestComboVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)
	(New(comboBoxTestItems()).Key("styled-combo")).Style(style.Style{}.BorderWidth(2)).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(180, 48)},
		Ops:         new(op.Ops),
	})
	if got := len(provider.requests); got != 1 {
		t.Fatalf("styled combo provider requests = %d, want one cached visual bundle", got)
	}
}

func TestComboOverrideBypassesBaseProviderCache(t *testing.T) {
	provider := &recordingComboVisualProvider{testComboVisualProvider: newTestComboVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	cache := computed.ProviderCache[field.Variant, comboDeclarationBundle]{}

	comboStyleDeclarationsCached(&active, field.Primary, &cache)
	comboStyleDeclarationsCached(&active, field.Primary, &cache)
	if len(provider.requests) != 1 {
		t.Fatalf("base combo bundle requests = %d, want 1", len(provider.requests))
	}
	emptyOverride := &theme.ComboOverride{}
	comboVisualBundle(&active, field.Primary, emptyOverride)
	if len(provider.requests) != 2 {
		t.Fatalf("empty combo override requests = %d, want 2", len(provider.requests))
	}
	comboVisualBundle(&active, field.Primary, emptyOverride)
	if len(provider.requests) != 3 {
		t.Fatalf("repeated empty combo override requests = %d, want 3", len(provider.requests))
	}
	comboStyleDeclarationsCached(&active, field.Primary, &cache)
	if got := cache.Stats(); got.Hits != 2 || got.Misses != 1 {
		t.Fatalf("base combo cache stats = %#v, want 2 hits and 1 miss", got)
	}
}
