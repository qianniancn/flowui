package input

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// testInputVisualProvider overrides only the input capabilities under test;
// the remaining component contract comes from the neutral behavior fixture.
type testInputVisualProvider struct {
	theme.ComponentVisualProvider
}

func (testInputVisualProvider) Input(active *theme.Theme, request theme.InputRequest) theme.InputVisual {
	base := testtheme.BehaviorProvider().Input(active, request)
	base.Defaults = style.Join(base.Defaults, style.Style{}.Height(unit.Dp(91)))
	base.Override = theme.InputOverrideStyle(request)
	return base
}

type incompleteInputVisualProvider struct {
	theme.ComponentVisualProvider
	visual theme.InputVisual
}

func (p incompleteInputVisualProvider) Input(*theme.Theme, theme.InputRequest) theme.InputVisual {
	return p.visual
}

func TestInputIncompleteVisualProviderDoesNotCommitCache(t *testing.T) {
	active := theme.NeutralTheme()
	base := newTestInputVisualProvider()
	visual := base.ComponentVisualProvider.Input(&active, theme.InputRequest{})
	visual.EditorStyle = nil
	active.VisualProvider = incompleteInputVisualProvider{
		ComponentVisualProvider: base.ComponentVisualProvider,
		visual:                  visual,
	}
	cache := computed.ProviderCache[inputVisualKey, theme.InputVisual]{}
	defer func() {
		recovered := recover()
		if err, ok := recovered.(error); !ok || !errors.Is(err, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete input visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Misses != 0 || stats.Hits != 0 {
			t.Fatalf("incomplete input visual changed cache stats = %#v", stats)
		}
	}()
	cache.Resolve(&active, inputVisualKey{}, func() theme.InputVisual {
		return inputVisualDeclarations(&active, Primary, false, theme.InputModeSingleLine, Medium, nil)
	})
}

func TestInputEditorStyleResolverDoesNotCommitCache(t *testing.T) {
	active := theme.NeutralTheme()
	base := newTestInputVisualProvider()
	visual := base.ComponentVisualProvider.Input(&active, theme.InputRequest{})
	visual.EditorStyle = func(style.ResolvedStyle, style.ResolvedStyle, style.ResolvedStyle) theme.InputEditorStyle {
		return theme.InputEditorStyle{}
	}
	active.VisualProvider = incompleteInputVisualProvider{
		ComponentVisualProvider: base.ComponentVisualProvider,
		visual:                  visual,
	}
	cache := computed.ProviderCache[inputVisualKey, theme.InputVisual]{}
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("invalid editor style panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Misses != 0 || stats.Hits != 0 {
			t.Fatalf("invalid editor style changed cache stats = %#v", stats)
		}
	}()
	cache.Resolve(&active, inputVisualKey{}, func() theme.InputVisual {
		return inputVisualDeclarations(&active, Primary, false, theme.InputModeSingleLine, Medium, nil)
	})
}

func TestInputIncompleteStateVisualDoesNotCommitCache(t *testing.T) {
	active := theme.NeutralTheme()
	base := newTestInputVisualProvider()
	visual := base.ComponentVisualProvider.Input(&active, theme.InputRequest{})
	visual.Defaults = style.Join(visual.Defaults, style.Style{}.
		On(style.Disabled, style.Style{}.BackgroundNone()))
	active.VisualProvider = incompleteInputVisualProvider{
		ComponentVisualProvider: base.ComponentVisualProvider,
		visual:                  visual,
	}
	cache := computed.ProviderCache[inputVisualKey, theme.InputVisual]{}
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("invalid disabled state panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Misses != 0 || stats.Hits != 0 {
			t.Fatalf("invalid disabled state changed cache stats = %#v", stats)
		}
	}()
	cache.Resolve(&active, inputVisualKey{}, func() theme.InputVisual {
		return inputVisualDeclarations(&active, Primary, false, theme.InputModeSingleLine, Medium, nil)
	})
}

func TestInputIncompletePartVisualsPanicBeforeCaching(t *testing.T) {
	baseTheme := *testtheme.Behavior()
	provider := baseTheme.VisualProvider.(theme.InputProvider)
	base := provider.Input(&baseTheme, theme.InputRequest{Mode: theme.InputModeSingleLine})
	root := style.Style{}.
		Part(style.PartRoot, style.Style{}.
			Height(unit.Dp(36)).
			Background(style.Solid(color.NRGBA{A: 0xff})).
			TextColor(style.Solid(color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})))
	cases := []struct {
		name string
		make func() theme.InputVisual
	}{
		{name: "root", make: func() theme.InputVisual {
			value := base
			value.Defaults = style.Style{}
			return value
		}},
		{name: "placeholder", make: func() theme.InputVisual {
			value := base
			value.Defaults = root
			return value
		}},
		{name: "selection", make: func() theme.InputVisual {
			value := base
			value.Defaults = style.Join(root, style.Style{}.
				Part(style.PartPlaceholder, style.Style{}.TextColor(style.Solid(color.NRGBA{A: 0xff}))))
			return value
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			active := baseTheme
			visual := test.make()
			active.VisualProvider = incompleteInputVisualProvider{
				ComponentVisualProvider: testtheme.BehaviorProvider(),
				visual:                  visual,
			}
			cache := computed.ProviderCache[inputVisualKey, theme.InputVisual]{}
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("incomplete %s visual panic = %v, want %v", test.name, recovered, theme.ErrIncompleteVisualProvider)
				}
				if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
					t.Fatalf("incomplete %s visual polluted cache = %#v", test.name, stats)
				}
			}()
			cache.Resolve(&active, inputVisualKey{}, func() theme.InputVisual {
				return inputVisualDeclarations(&active, Primary, false, theme.InputModeSingleLine, Medium, nil)
			})
		})
	}
}

func TestInputInstanceOverrideUsesProviderLayer(t *testing.T) {
	provider := &recordingInputVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := *testtheme.Behavior()
	active.VisualProvider = provider
	value := style.Style{}.BorderWidth(2)
	override := theme.InputOverride{Declaration: value}
	visual := inputVisualDeclarations(
		&active,
		Primary,
		false,
		theme.InputModeSingleLine,
		Medium,
		&override,
	)
	resolved := visual.Override.Resolve(style.State{})
	if resolved.Paint == nil || resolved.Paint.Border == nil || resolved.Paint.Border.Width == nil || *resolved.Paint.Border.Width != 2 {
		t.Fatalf("provider override = %#v, want border width 2", resolved.Paint)
	}
	if len(provider.requests) != 1 || provider.requests[0].Override == nil {
		t.Fatal("input provider did not receive the instance override")
	}
}

func (testInputVisualProvider) InputAction(active *theme.Theme, request theme.InputActionRequest) theme.InputActionVisual {
	base := testtheme.BehaviorProvider().InputAction(active, request)
	base.Declaration = style.Join(base.Declaration, style.Style{}.Width(unit.Dp(37)), theme.InputActionOverrideStyle(request))
	base.Metrics.SlotPadding = 9
	return base
}

type slotPaddingInputVisualProvider struct {
	theme.ComponentVisualProvider
}

func (slotPaddingInputVisualProvider) InputAction(active *theme.Theme, request theme.InputActionRequest) theme.InputActionVisual {
	base := testtheme.BehaviorProvider().InputAction(active, request)
	base.Declaration = style.Join(
		base.Declaration,
		style.Style{}.Width(unit.Dp(24)).Height(unit.Dp(24)),
		theme.InputActionOverrideStyle(request),
	)
	base.Metrics.SlotPadding = 9
	return base
}

func newTestInputVisualProvider() testInputVisualProvider {
	return testInputVisualProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
}

func TestInputUsesActiveVisualProvider(t *testing.T) {
	active := theme.NeutralTheme()
	active.VisualProvider = newTestInputVisualProvider()

	visual := inputVisualDeclarations(
		&active,
		Primary,
		false,
		theme.InputModeSingleLine,
		Medium,
		nil,
	)
	defaults := visual.Defaults
	resolved := defaults.Resolve(style.State{})
	if resolved.Box == nil || resolved.Box.Height == nil || *resolved.Box.Height != 91 {
		t.Fatalf("provider height = %#v, want 91", resolved.Box)
	}
}

func TestInputLayoutRejectsProviderMutationInStrictContext(t *testing.T) {
	active := theme.NeutralTheme()
	active.VisualProvider = newTestInputVisualProvider()
	ctx, err := core.NewWithCatalogStrict(nil, &active, locale.LanguageEnglish, i18n.Catalog{})
	if err != nil {
		t.Fatalf("strict context construction failed: %v", err)
	}
	active.VisualProvider = nil
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(180, 48)}, Ops: &ops}
	defer func() {
		if recover() == nil {
			t.Fatal("input layout accepted a missing provider in a strict context")
		}
	}()
	New("").Key("strict-provider").Layout(ctx, gtx)
}

type recordingInputVisualProvider struct {
	testInputVisualProvider
	requests []theme.InputRequest
}

func (p *recordingInputVisualProvider) Input(active *theme.Theme, request theme.InputRequest) theme.InputVisual {
	p.requests = append(p.requests, request)
	return p.testInputVisualProvider.Input(active, request)
}

func TestInputSizeIsResolvedByVisualProvider(t *testing.T) {
	provider := &recordingInputVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider

	state := &inputState{}
	state.visualDeclarations(New("").Key("name").Size(Small), &active, nil)

	if len(provider.requests) != 1 {
		t.Fatalf("provider requests = %d, want 1", len(provider.requests))
	}
	request := provider.requests[0]
	if request.Size != theme.InputSmall {
		t.Fatalf("provider size = %v, want small", request.Size)
	}
	if request.Height != 0 {
		t.Fatalf("single-line provider height = %v, want zero dynamic height", request.Height)
	}
}

func TestInputGroupActionUsesActiveVisualProvider(t *testing.T) {
	active := theme.NeutralTheme()
	active.VisualProvider = newTestInputVisualProvider()

	declaration := inputGroupActionVisual(&active, theme.InputActionRequest{}).Declaration
	resolved := declaration.Resolve(style.State{})
	if resolved.Box == nil || resolved.Box.Width == nil || *resolved.Box.Width != 37 {
		t.Fatalf("provider action width = %#v, want 37", resolved.Box)
	}
}

func TestInputGroupActionIncompleteVisualPanicsBeforeCaching(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteInputActionVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
	}
	cache := inputActionVisualCache{}
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete input action panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete input action polluted cache = %#v", stats)
		}
	}()
	activeContext := core.New(nil, &active, locale.LanguageAuto)
	cache.cache.Resolve(&active, inputActionVisualKey{}, func() theme.InputActionVisual {
		return inputGroupActionVisual(core.ActiveTheme(activeContext), theme.InputActionRequest{})
	})
}

type incompleteInputActionVisualProvider struct {
	theme.ComponentVisualProvider
}

func (p incompleteInputActionVisualProvider) InputAction(*theme.Theme, theme.InputActionRequest) theme.InputActionVisual {
	return theme.InputActionVisual{}
}

type recordingInputActionVisualProvider struct {
	testInputVisualProvider
	requests []theme.InputActionRequest
}

func (p *recordingInputActionVisualProvider) InputAction(active *theme.Theme, request theme.InputActionRequest) theme.InputActionVisual {
	p.requests = append(p.requests, request)
	return p.testInputVisualProvider.InputAction(active, request)
}

func TestInputGroupReusesPlainActionVisualWithinLayout(t *testing.T) {
	provider := &recordingInputActionVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)

	Group(New("").Key("query")).
		PrefixAction(GroupAction("filter", "Filter values").Content(inputGroupFixedWidget{size: image.Pt(16, 16)})).
		SuffixAction(GroupAction("clear", "Clear value").Content(inputGroupFixedWidget{size: image.Pt(16, 16)})).
		Layout(ctx, testLayoutContext())

	if len(provider.requests) != 1 {
		t.Fatalf("two plain group actions provider requests = %d, want 1", len(provider.requests))
	}
	if provider.requests[0].Override != nil {
		t.Fatal("plain group action unexpectedly received an instance override")
	}
}

func TestInputGroupActionVisualCacheReusesStableThemePlan(t *testing.T) {
	provider := &recordingInputActionVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)
	value := Group(New("").Key("query")).PrefixAction(
		GroupAction("filter", "Filter values").Content(inputGroupFixedWidget{size: image.Pt(16, 16)}),
	)

	layoutFrame := func() {
		core.BeginFrame(ctx)
		value.Layout(ctx, testLayoutContext())
		core.EndFrame(ctx)
	}
	layoutFrame()
	layoutFrame()
	if len(provider.requests) != 1 {
		t.Fatalf("stable group action provider requests = %d, want 1", len(provider.requests))
	}

	replacement := active
	core.ReplaceTheme(ctx, replacement)
	layoutFrame()
	if len(provider.requests) != 2 {
		t.Fatalf("theme replacement provider requests = %d, want 2", len(provider.requests))
	}
}

func TestInputGroupActionVisualCacheInvalidatesForOverrideKey(t *testing.T) {
	provider := &recordingInputActionVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)
	plain := theme.InputActionRequest{}
	firstOverride := theme.InputActionRequest{Override: &theme.InputActionOverride{
		Declaration: style.Style{}.Width(unit.Dp(41)),
	}}
	secondOverride := theme.InputActionRequest{Override: &theme.InputActionOverride{
		Declaration: style.Style{}.Width(unit.Dp(43)),
	}}

	plainVisual := inputGroupActionVisualCached(ctx, "clear", plain)
	if got := plainVisual.Declaration.Resolve(style.State{}).Box.Width; got == nil || *got != 37 {
		t.Fatalf("plain action width = %v, want 37", got)
	}
	inputGroupActionVisualCached(ctx, "clear", plain)
	firstVisual := inputGroupActionVisualCached(ctx, "clear", firstOverride)
	if got := firstVisual.Declaration.Resolve(style.State{}).Box.Width; got == nil || *got != 41 {
		t.Fatalf("first override width = %v, want 41", got)
	}
	inputGroupActionVisualCached(ctx, "clear", firstOverride)
	secondVisual := inputGroupActionVisualCached(ctx, "clear", secondOverride)
	if got := secondVisual.Declaration.Resolve(style.State{}).Box.Width; got == nil || *got != 43 {
		t.Fatalf("second override width = %v, want 43", got)
	}

	if len(provider.requests) != 3 {
		t.Fatalf("override key provider requests = %d, want 3", len(provider.requests))
	}
}

func TestInputGroupActionVisualCacheDoesNotReuseAfterProviderRemoval(t *testing.T) {
	active := *testtheme.Behavior()
	ctx, err := core.NewWithCatalogStrict(nil, &active, locale.LanguageAuto, i18n.Catalog{})
	if err != nil {
		t.Fatalf("strict context construction failed: %v", err)
	}

	inputGroupActionVisualCached(ctx, "clear", theme.InputActionRequest{})
	core.ActiveTheme(ctx).VisualProvider = nil
	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		if !ok || !errors.Is(err, theme.ErrMissingVisualProvider) {
			t.Fatalf("provider removal panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
		}
	}()
	inputGroupActionVisualCached(ctx, "clear", theme.InputActionRequest{})
}

func TestInputGroupWithoutActionDoesNotRequestActionVisual(t *testing.T) {
	provider := &recordingInputActionVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)

	Group(New("").Key("query")).Prefix(inputGroupFixedWidget{size: image.Pt(16, 16)}).Layout(ctx, testLayoutContext())

	if len(provider.requests) != 0 {
		t.Fatalf("group without action provider requests = %d, want 0", len(provider.requests))
	}
}

func TestInputGroupStyledActionKeepsInstanceOverrideRequest(t *testing.T) {
	provider := &recordingInputActionVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := theme.NeutralTheme()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)
	declaration := style.Style{}.Width(unit.Dp(41))

	Group(New("").Key("query")).SuffixAction(
		GroupAction("clear", "Clear value").Content(inputGroupFixedWidget{size: image.Pt(16, 16)}).Style(declaration),
	).Layout(ctx, testLayoutContext())

	if len(provider.requests) != 2 {
		t.Fatalf("styled group action provider requests = %d, want 2", len(provider.requests))
	}
	if provider.requests[1].Override == nil {
		t.Fatal("styled group action did not receive an instance override")
	}
}

func TestInputMissingVisualProviderPanics(t *testing.T) {
	active := theme.NeutralTheme()

	assertMissingProvider := func(name string, fn func()) {
		t.Helper()
		defer func() {
			recovered := recover()
			err, ok := recovered.(error)
			if !ok || !errors.Is(err, theme.ErrMissingVisualProvider) {
				t.Fatalf("%s panic = %v, want %v", name, recovered, theme.ErrMissingVisualProvider)
			}
		}()
		fn()
	}

	assertMissingProvider("input", func() {
		inputVisualDeclarations(&active, Primary, false, theme.InputModeSingleLine, Medium, nil)
	})
	assertMissingProvider("group action", func() {
		inputGroupActionVisual(&active, theme.InputActionRequest{})
	})
}
