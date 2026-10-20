package button

import (
	"errors"
	"image"
	"image/color"
	"math"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type testVisualProvider struct {
	theme.ComponentVisualProvider
}

type nonComparableImmutableVisualProvider struct {
	theme.ComponentVisualProvider
	marker []byte
}

func (nonComparableImmutableVisualProvider) VisualsAreImmutable() {}

type dynamicImmutableVisualProvider struct {
	theme.ComponentVisualProvider
	marker any
}

func (dynamicImmutableVisualProvider) VisualsAreImmutable() {}

func (provider testVisualProvider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	visual := provider.ComponentVisualProvider.Button(active, request)
	visual.Defaults = style.Join(visual.Defaults, style.Style{}.Radius(3))
	visual.Variant = style.Join(visual.Variant, style.Style{}.Background(style.Solid(themeColor(0x112233ff))))
	visual.Size = style.Join(visual.Size, style.Style{}.Height(73))
	return visual
}

func newTestVisualProvider() testVisualProvider {
	return testVisualProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
}

func themeColor(value uint32) color.NRGBA {
	return color.NRGBA{R: uint8(value >> 24), G: uint8(value >> 16), B: uint8(value >> 8), A: uint8(value)}
}

func TestButtonValidationCacheOnlyKeysComparableImmutableProviders(t *testing.T) {
	request := theme.ButtonRequest{Variant: theme.ButtonPrimary, Size: theme.ButtonMedium}
	if _, ok := buttonVisualValidationKeyFor(testtheme.BehaviorProvider(), request); !ok {
		t.Fatal("comparable immutable provider was rejected by validation cache key")
	}
	nonComparable := nonComparableImmutableVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		marker:                  []byte{1},
	}
	if _, ok := buttonVisualValidationKeyFor(nonComparable, request); ok {
		t.Fatal("non-comparable immutable provider entered validation cache key")
	}
	dynamic := dynamicImmutableVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		marker:                  []byte{1},
	}
	if _, ok := buttonVisualValidationKeyFor(dynamic, request); ok {
		t.Fatal("provider with non-comparable interface field entered validation cache key")
	}
	active := *testtheme.Behavior()
	active.VisualProvider = dynamic
	buttonVisual(&active, request, nil)
}

func TestButtonUsesActiveVisualProvider(t *testing.T) {
	active := theme.NeutralTheme()
	active.VisualProvider = newTestVisualProvider()
	ctx := core.New(nil, &active, locale.LanguageEnglish)

	resolved := New("").Content(nil).staticStyle(ctx, style.State{})
	if resolved.root.Box == nil || resolved.root.Box.Height == nil || *resolved.root.Box.Height != 73 {
		t.Fatalf("provider height = %#v, want 73", resolved.root.Box)
	}
	if resolved.root.Paint == nil || resolved.root.Paint.Radius == nil || *resolved.root.Paint.Radius != 3 {
		t.Fatalf("provider radius = %#v, want 3", resolved.root.Paint)
	}
}

func TestButtonLayoutRejectsProviderMutationInStrictContext(t *testing.T) {
	active := theme.NeutralTheme()
	active.VisualProvider = newTestVisualProvider()
	ctx, err := core.NewWithCatalogStrict(nil, &active, locale.LanguageEnglish, i18n.Catalog{})
	if err != nil {
		t.Fatalf("strict context construction failed: %v", err)
	}
	active.VisualProvider = nil
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(100, 40)}, Ops: &ops}
	defer func() {
		if recover() == nil {
			t.Fatal("button layout accepted a missing provider in a strict context")
		}
	}()
	New("").Content(nil).Layout(ctx, gtx)
}

func TestButtonMissingProviderPanicsInsteadOfUsingEmptyVisual(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := core.New(nil, &active, locale.LanguageAuto)

	defer func() {
		if recovered := recover(); recovered != theme.ErrMissingVisualProvider {
			t.Fatalf("missing provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
		}
	}()

	New("").Content(nil).staticStyle(ctx, style.State{})
}

type incompleteVisualProvider struct {
	theme.ComponentVisualProvider
	mutate func(*theme.ButtonVisual)
}

// mutableValidationProvider deliberately returns the same closure code for a
// valid and invalid visual while changing the captured provider state. It
// proves that custom providers bypass the shared immutable validation facts.
type mutableValidationProvider struct {
	theme.ComponentVisualProvider
	invalid bool
}

func (provider mutableValidationProvider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	visual := provider.ComponentVisualProvider.Button(active, request)
	if provider.invalid {
		visual.Size = style.Join(visual.Size,
			style.Style{}.Part(style.PartLabel, style.Style{}.FontSize(0)))
	}
	return visual
}

func (provider incompleteVisualProvider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	visual := provider.ComponentVisualProvider.Button(active, request)
	provider.mutate(&visual)
	return visual
}

func TestButtonIncompleteVisualPanicsBeforeCaching(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.Spinner = nil
		},
	}
	value := New("").Content(nil)
	cache := buttonDeclarationCache{}

	func() {
		defer func() {
			recovered := recover()
			if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
				t.Fatalf("incomplete provider panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
			}
		}()
		value.declarationsForOverride(&active, &cache, nil)
	}()

	active.VisualProvider = testtheme.BehaviorProvider()
	if value.declarationsForOverride(&active, &cache, nil).visual.Spinner == nil {
		t.Fatal("complete provider did not resolve after incomplete result")
	}
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 1 {
		t.Fatalf("incomplete provider cache stats = %#v, want one successful miss and no hit", stats)
	}
}

func TestButtonMissingForegroundResolverPanicsBeforeCaching(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.ForegroundColor = nil
		},
	}
	cache := buttonDeclarationCache{}

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("missing foreground resolver panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete foreground resolver polluted cache = %#v", stats)
		}
	}()
	New("").Content(nil).declarationsForOverride(&active, &cache, nil)
}

func TestButtonForegroundResolverRejectsInvalidResultBeforeCaching(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.ForegroundColor = func(style.ResolvedStyle) (color.NRGBA, bool) {
				return color.NRGBA{}, false
			}
		},
	}
	cache := buttonDeclarationCache{}

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("invalid foreground resolver panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("invalid foreground resolver polluted cache = %#v", stats)
		}
	}()
	New("").Content(nil).declarationsForOverride(&active, &cache, nil)
}

func TestButtonMissingLoadingInsetResolverPanicsBeforeCaching(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.LoadingInset = nil
		},
	}
	cache := buttonDeclarationCache{}

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("missing loading inset resolver panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete loading inset resolver polluted cache = %#v", stats)
		}
	}()
	New("").Content(nil).declarationsForOverride(&active, &cache, nil)
}

func TestButtonMissingSpinnerSizeResolverPanicsBeforeCaching(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.SpinnerSize = nil
		},
	}
	cache := buttonDeclarationCache{}

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("missing spinner size resolver panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete spinner size resolver polluted cache = %#v", stats)
		}
	}()
	New("").Content(nil).declarationsForOverride(&active, &cache, nil)
}

func TestButtonMissingSpinnerGeometryResolversPanicsBeforeCaching(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*theme.ButtonVisual)
	}{{
		name: "stroke width",
		mutate: func(visual *theme.ButtonVisual) {
			visual.SpinnerStrokeWidth = nil
		},
	}, {
		name: "period",
		mutate: func(visual *theme.ButtonVisual) {
			visual.SpinnerPeriod = nil
		},
	}} {
		t.Run(test.name, func(t *testing.T) {
			active := *testtheme.Behavior()
			active.VisualProvider = incompleteVisualProvider{
				ComponentVisualProvider: testtheme.BehaviorProvider(),
				mutate:                  test.mutate,
			}
			cache := buttonDeclarationCache{}

			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("missing spinner geometry resolver panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
				if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
					t.Fatalf("incomplete spinner geometry resolver polluted cache = %#v", stats)
				}
			}()
			New("").Content(nil).declarationsForOverride(&active, &cache, nil)
		})
	}
}

func TestButtonMissingContentGapResolverPanicsBeforeCaching(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.ContentGap = nil
		},
	}
	cache := buttonDeclarationCache{}

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("missing content gap resolver panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete content gap resolver polluted cache = %#v", stats)
		}
	}()
	New("").Content(nil).declarationsForOverride(&active, &cache, nil)
}

func TestButtonResolverMetricContractsRejectContextualInvalidValuesBeforeCaching(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*theme.ButtonVisual)
	}{
		{
			name: "loading content gap",
			mutate: func(visual *theme.ButtonVisual) {
				visual.ContentGap = func(request theme.ButtonContentGapRequest) unit.Dp {
					if request.Loading {
						return unit.Dp(math.NaN())
					}
					return 0
				}
			},
		},
		{
			name: "loading inset",
			mutate: func(visual *theme.ButtonVisual) {
				visual.LoadingInset = func(layout.Context, theme.ButtonLoadingRequest) layout.Inset {
					return layout.Inset{Left: -1}
				}
			},
		},
		{
			name: "spinner size",
			mutate: func(visual *theme.ButtonVisual) {
				visual.SpinnerSize = func() unit.Dp {
					return unit.Dp(math.Inf(1))
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			active := *testtheme.Behavior()
			active.VisualProvider = incompleteVisualProvider{
				ComponentVisualProvider: testtheme.BehaviorProvider(),
				mutate:                  test.mutate,
			}
			cache := buttonDeclarationCache{}
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("invalid contextual resolver panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
				if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
					t.Fatalf("invalid contextual resolver polluted cache = %#v", stats)
				}
			}()
			New("").Content(nil).declarationsForOverride(&active, &cache, nil)
		})
	}
}

func TestButtonValidationCacheTracksInPlaceProviderReplacement(t *testing.T) {
	active := *testtheme.Behavior()
	value := New("").Content(nil)
	value.declarationsForOverride(&active, &buttonDeclarationCache{}, nil)
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.Spinner = nil
		},
	}
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("in-place provider replacement panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
	}()
	value.declarationsForOverride(&active, &buttonDeclarationCache{}, nil)
}

func TestButtonMutableProviderDoesNotReuseImmutableValidationFact(t *testing.T) {
	active := *testtheme.Behavior()
	base := testtheme.BehaviorProvider()
	active.VisualProvider = mutableValidationProvider{ComponentVisualProvider: base}
	value := New("").Content(nil)
	value.declarationsForOverride(&active, &buttonDeclarationCache{}, nil)

	active.VisualProvider = mutableValidationProvider{ComponentVisualProvider: base, invalid: true}
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("mutable provider validation panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
	}()
	value.declarationsForOverride(&active, &buttonDeclarationCache{}, nil)
}

func TestButtonIncompletePartVisualPanicsBeforeCaching(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.Size = style.Join(visual.Size,
				style.Style{}.Part(style.PartLabel, style.Style{}.FontSize(0)))
		},
	}
	cache := buttonDeclarationCache{}

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete part visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete part visual polluted cache = %#v", stats)
		}
	}()
	New("").Content(nil).declarationsForOverride(&active, &cache, nil)
}

func TestButtonMissingRippleFadeDurationPanicsBeforeCaching(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.RippleFadeDuration = 0
		},
	}
	cache := buttonDeclarationCache{}

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("missing ripple fade duration panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
		if stats := cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete ripple fade duration polluted cache = %#v", stats)
		}
	}()
	New("").Content(nil).declarationsForOverride(&active, &cache, nil)
}

func TestButtonRippleDisabledDoesNotRequireRippleVisualProtocol(t *testing.T) {
	active := *testtheme.Behavior()
	active.VisualProvider = incompleteVisualProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		mutate: func(visual *theme.ButtonVisual) {
			visual.Ripple = false
			visual.RippleDuration = 0
			visual.RippleFadeDuration = 0
			visual.RippleRenderer = nil
			visual.RippleColor = nil
		},
	}

	visual := New("").Content(nil).declarationsForOverride(&active, nil, nil).visual
	if visual.Ripple {
		t.Fatal("provider ripple flag was not disabled")
	}
}

func TestButtonLoadingStyleAppliesProviderInsetWithoutMutatingDeclaration(t *testing.T) {
	padding := style.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16}
	var received theme.ButtonLoadingRequest
	resolved := buttonStyle{
		root: style.ResolvedStyle{Box: &style.BoxStyle{Padding: &padding}},
		visual: theme.ButtonVisual{
			LoadingInset: func(_ layout.Context, request theme.ButtonLoadingRequest) layout.Inset {
				received = request
				return layout.Inset{Left: 3, Right: 5}
			},
		},
	}

	loading := buttonLoadingStyle(layout.Context{}, resolved)
	if received.Inset != (layout.Inset{Top: 8, Right: 16, Bottom: 8, Left: 16}) {
		t.Fatalf("loading resolver input = %+v, want original root inset", received.Inset)
	}
	if got := buttonInset(loading.root); got != (layout.Inset{Top: 8, Right: 5, Bottom: 8, Left: 3}) {
		t.Fatalf("loading root inset = %+v, want provider inset applied", got)
	}
	if got := buttonInset(resolved.root); got != (layout.Inset{Top: 8, Right: 16, Bottom: 8, Left: 16}) {
		t.Fatalf("original root inset changed to %+v", got)
	}
}

func TestButtonVisualHelpersRejectMissingRootDeclarations(t *testing.T) {
	for _, test := range []struct {
		name string
		call func()
	}{
		{name: "inset", call: func() { buttonInset(style.ResolvedStyle{}) }},
		{name: "radius", call: func() { buttonRadius(style.ResolvedStyle{}) }},
		{name: "opacity", call: func() { buttonOpacity(style.ResolvedStyle{}) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("missing %s declaration panic = %v, want %v", test.name, recovered, theme.ErrIncompleteVisualProvider)
				}
			}()
			test.call()
		})
	}
}

func TestButtonLoadingStyleRejectsMissingRootPadding(t *testing.T) {
	resolved := buttonStyle{
		visual: theme.ButtonVisual{
			LoadingInset: func(layout.Context, theme.ButtonLoadingRequest) layout.Inset {
				return layout.Inset{}
			},
		},
	}

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("missing loading root padding panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
	}()
	buttonLoadingStyle(testLayoutContext(), resolved)
}

func TestButtonLoadingStyleRejectsContextualInvalidInset(t *testing.T) {
	padding := style.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16}
	resolved := buttonStyle{
		root: style.ResolvedStyle{Box: &style.BoxStyle{Padding: &padding}},
		visual: theme.ButtonVisual{
			LoadingInset: func(_ layout.Context, request theme.ButtonLoadingRequest) layout.Inset {
				if request.Inset.Left == 8 {
					return layout.Inset{}
				}
				return layout.Inset{Left: -1}
			},
		},
	}

	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("contextual invalid loading inset panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
	}()
	buttonLoadingStyle(testLayoutContext(), resolved)
}
