package button

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type countingButtonProvider struct {
	theme.ComponentVisualProvider
	calls    *int
	requests *[]theme.ButtonRequest
}

type immutableCountingButtonProvider struct {
	theme.ComponentVisualProvider
	calls *int
}

func (provider immutableCountingButtonProvider) VisualsAreImmutable() {}

func (provider immutableCountingButtonProvider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	(*provider.calls)++
	return provider.ComponentVisualProvider.Button(active, request)
}

func (provider countingButtonProvider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	(*provider.calls)++
	if provider.requests != nil {
		*provider.requests = append(*provider.requests, request)
	}
	return provider.ComponentVisualProvider.Button(active, request)
}

func TestButtonDeclarationCacheInvalidation(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	calls := 0
	activeTheme.VisualProvider = countingButtonProvider{ComponentVisualProvider: testtheme.Behavior().VisualProvider.(theme.ComponentVisualProvider), calls: &calls}
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	value := New("").Content(nil).Variant(Primary).Size(Medium)

	value.staticStyle(ctx, style.State{})
	value.staticStyle(ctx, style.State{Hovered: true})
	if calls != 1 {
		t.Fatalf("stable request called provider %d times, want 1", calls)
	}

	value.Variant(Secondary).staticStyle(ctx, style.State{})
	value.IconOnly().staticStyle(ctx, style.State{})
	GroupStart(value).staticStyle(ctx, style.State{})
	if calls != 4 {
		t.Fatalf("request changes called provider %d times, want 4", calls)
	}

	replacement := activeTheme
	core.ReplaceTheme(ctx, replacement)
	value.staticStyle(ctx, style.State{})
	if calls != 5 {
		t.Fatalf("theme replacement called provider %d times, want 5", calls)
	}
}

func TestButtonDeclarationCacheStatsExposeInvalidationReasons(t *testing.T) {
	active := testtheme.Behavior()
	value := New("").Content(nil).Variant(Primary).Size(Medium)
	cache := buttonDeclarationCache{}

	value.declarationsForOverride(active, &cache, nil)
	value.declarationsForOverride(active, &cache, nil)
	value.Variant(Secondary).declarationsForOverride(active, &cache, nil)
	replacement := *active
	value.Variant(Secondary).declarationsForOverride(&replacement, &cache, nil)

	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 3 {
		t.Fatalf("cache stats = %#v, want one hit and three misses", stats)
	}
	if stats.KeyInvalidations != 1 || stats.ThemeInvalidations != 1 {
		t.Fatalf("cache invalidations = %#v, want one key and one theme invalidation", stats)
	}
}

func TestButtonDeclarationCachePanicDoesNotDiscardPreviousValue(t *testing.T) {
	active := *testtheme.Behavior()
	value := New("").Content(nil).Variant(Primary)
	cache := buttonDeclarationCache{}

	first := value.declarationsForOverride(&active, &cache, nil)
	provider := active.VisualProvider
	active.VisualProvider = nil

	func() {
		defer func() {
			if recovered := recover(); recovered != theme.ErrMissingVisualProvider {
				t.Fatalf("provider panic = %v, want %v", recovered, theme.ErrMissingVisualProvider)
			}
		}()
		value.Variant(Secondary).declarationsForOverride(&active, &cache, nil)
	}()

	active.VisualProvider = provider
	if got := value.declarationsForOverride(&active, &cache, nil); got.visual.Defaults.Hash64() != first.visual.Defaults.Hash64() {
		t.Fatal("cached declaration was replaced after provider panic")
	}
	stats := cache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 || stats.KeyInvalidations != 1 || stats.ThemeInvalidations != 1 {
		t.Fatalf("panic cache stats = %#v, want one hit, one miss, one key and one provider invalidation", stats)
	}
}

func TestButtonInstanceOverrideDoesNotPolluteDeclarationCache(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	calls := 0
	activeTheme.VisualProvider = countingButtonProvider{
		ComponentVisualProvider: testtheme.Behavior().VisualProvider.(theme.ComponentVisualProvider),
		calls:                   &calls,
	}
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	base := New("").Content(nil).Variant(Primary)
	base.staticStyle(ctx, style.State{})

	instanceColor := color.NRGBA{R: 0x91, G: 0x22, B: 0x44, A: 0xff}
	styled := (base).Style(style.Style{}.Background(style.Solid(instanceColor)))
	resolvedStyled := styled.staticStyle(ctx, style.State{})
	resolvedBase := base.staticStyle(ctx, style.State{})

	if got := resolvedBackground(resolvedStyled); got != instanceColor {
		t.Fatalf("provider did not apply instance override: %#v", got)
	}
	if got := resolvedBackground(resolvedBase); got == instanceColor {
		t.Fatal("instance override polluted the cached base declaration")
	}
	if calls != 2 {
		t.Fatalf("provider calls = %d, want one base and one uncached override resolution", calls)
	}
}

func TestButtonImmutableProviderCachesStableInstanceOverride(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	calls := 0
	activeTheme.VisualProvider = immutableCountingButtonProvider{
		ComponentVisualProvider: testtheme.Behavior().VisualProvider.(theme.ComponentVisualProvider),
		calls:                   &calls,
	}
	value := New("").Content(nil).Variant(Primary)
	cache := buttonDeclarationCache{}
	override := &theme.ButtonOverride{Declaration: style.Style{}.PaddingX(8)}

	first := value.declarationsForOverride(&activeTheme, &cache, override)
	second := value.declarationsForOverride(&activeTheme, &cache, override)
	if calls != 1 {
		t.Fatalf("immutable override provider calls = %d, want one", calls)
	}
	if first.visual.Override.Hash64() != second.visual.Override.Hash64() {
		t.Fatal("cached immutable override changed the visual declaration")
	}
	if stats := cache.Stats(); stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("immutable override cache stats = %#v, want one hit and one miss", stats)
	}
}

func TestButtonImmutableOverrideCacheInvalidatesOnDeclarationAndThemeChange(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	calls := 0
	activeTheme.VisualProvider = immutableCountingButtonProvider{
		ComponentVisualProvider: testtheme.Behavior().VisualProvider.(theme.ComponentVisualProvider),
		calls:                   &calls,
	}
	value := New("").Content(nil).Variant(Primary)
	cache := buttonDeclarationCache{}
	firstOverride := &theme.ButtonOverride{Declaration: style.Style{}.PaddingX(8)}
	secondOverride := &theme.ButtonOverride{Declaration: style.Style{}.PaddingX(12)}

	value.declarationsForOverride(&activeTheme, &cache, firstOverride)
	value.declarationsForOverride(&activeTheme, &cache, secondOverride)
	replacement := activeTheme
	value.declarationsForOverride(&replacement, &cache, secondOverride)

	if calls != 3 {
		t.Fatalf("immutable override invalidation provider calls = %d, want three", calls)
	}
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 3 || stats.KeyInvalidations != 1 || stats.ThemeInvalidations != 1 {
		t.Fatalf("immutable override invalidation stats = %#v", stats)
	}
}

func TestButtonStyledWrapperKeepsOverrideOutsideBehaviorValue(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	calls := 0
	requests := []theme.ButtonRequest{}
	activeTheme.VisualProvider = countingButtonProvider{
		ComponentVisualProvider: testtheme.Behavior().VisualProvider.(theme.ComponentVisualProvider),
		calls:                   &calls,
		requests:                &requests,
	}
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	declaration := style.Style{}.Background(style.Solid(themeColor(0x912244ff)))
	base := New("").Content(nil).Variant(Primary)
	wrapped := (base).Style(declaration)

	wrapped.Layout(ctx, layout.Context{Ops: new(op.Ops)})
	if calls != 1 {
		t.Fatalf("styled wrapper provider calls = %d, want 1", calls)
	}
	if len(requests) != 1 || requests[0].Override == nil || requests[0].Override.Declaration.Hash64() == (style.Style{}).Hash64() {
		t.Fatal("styled wrapper did not send its declaration to the provider")
	}
}

func TestPreparedButtonStyleResolvesOnlyLiveParts(t *testing.T) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("").Content(nil)
	declarations := value.declarationsForOverride(active, nil, nil)

	prepared := value.resolvePreparedStyleWithDeclarations(
		ctx,
		testLayoutContext(),
		value.key,
		style.State{},
		declarations,
	)
	if !resolvedStyleEmpty(prepared.indicatorPart) {
		t.Fatalf("prepared indicator part = %#v, want unresolved", prepared.indicatorPart)
	}

	static := value.staticContentStyleWithDeclarations(ctx, style.State{}, declarations)
	if !resolvedStyleEmpty(static.indicatorPart) {
		t.Fatalf("non-loading prepared indicator part = %#v, want unresolved", static.indicatorPart)
	}

	loading := value.Loading(true)
	loadingDeclarations := loading.declarationsForOverride(active, nil, nil)
	loadingStyle := loading.staticContentStyleWithDeclarations(
		ctx,
		style.State{Loading: true},
		loadingDeclarations,
	)
	if resolvedStyleEmpty(loadingStyle.indicatorPart) {
		t.Fatal("loading prepared indicator part was not resolved")
	}
}

func TestButtonStyleResolvesIndicatorOnlyWhileLoading(t *testing.T) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageEnglish)
	gtx := testLayoutContext()
	value := New("").Content(nil)
	declarations := value.declarationsForOverride(active, nil, nil)

	regular := value.resolveStyleWithDeclarations(ctx, gtx, value.key, style.State{}, declarations)
	if !resolvedStyleEmpty(regular.indicatorPart) {
		t.Fatalf("regular button indicator part = %#v, want unresolved", regular.indicatorPart)
	}

	loading := value.Loading(true)
	loadingStyle := loading.resolveStyleWithDeclarations(
		ctx,
		gtx,
		loading.key,
		style.State{Loading: true},
		declarations,
	)
	if resolvedStyleEmpty(loadingStyle.indicatorPart) {
		t.Fatal("loading button indicator part was not resolved")
	}
}

func resolvedStyleEmpty(value style.ResolvedStyle) bool {
	return value.Box == nil && value.Paint == nil && value.Text == nil && value.Trans == nil && len(value.Transitions) == 0
}

func TestButtonGroupSeparatorsReuseResolvedProviderVisual(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	calls := 0
	requests := []theme.ButtonRequest{}
	activeTheme.VisualProvider = countingButtonProvider{
		ComponentVisualProvider: testtheme.Behavior().VisualProvider.(theme.ComponentVisualProvider),
		calls:                   &calls,
		requests:                &requests,
	}
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(320, 80)}, Ops: new(op.Ops)}

	Group(
		New("").Content(nil).Variant(Primary),
		New("").Content(nil).Variant(Primary),
	).Separators(true).Layout(ctx, gtx)

	if calls != 2 {
		t.Fatalf("group provider calls = %d, want one per grouped request", calls)
	}
	if len(requests) != 2 || !requests[0].Grouped || !requests[1].Grouped {
		t.Fatalf("group provider requests = %#v, want only grouped button requests", requests)
	}
}

type separatorRendererProvider struct {
	theme.ComponentVisualProvider
	firstCalls  *int
	secondCalls *int
}

func (provider separatorRendererProvider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	visual := provider.ComponentVisualProvider.Button(active, request)
	visual.Group.SeparatorRenderer = func(layout.Context, theme.ButtonGroupSeparatorRequest) {
		if request.Variant == Primary {
			(*provider.firstCalls)++
			return
		}
		(*provider.secondCalls)++
	}
	return visual
}

func TestButtonGroupUsesBoundaryItemSeparatorRenderer(t *testing.T) {
	active := *testtheme.Behavior()
	firstCalls, secondCalls := 0, 0
	active.VisualProvider = separatorRendererProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		firstCalls:              &firstCalls,
		secondCalls:             &secondCalls,
	}
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(320, 80)}, Ops: new(op.Ops)}

	Group(
		New("").Content(nil).Variant(Primary),
		New("").Content(nil).Variant(Danger),
	).Separators(true).Layout(ctx, gtx)

	if firstCalls != 0 || secondCalls != 1 {
		t.Fatalf("separator renderer calls = first:%d second:%d, want first:0 second:1", firstCalls, secondCalls)
	}
}

func TestButtonGroupPositionIsPassedAsSemanticCorners(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	calls := 0
	requests := []theme.ButtonRequest{}
	activeTheme.VisualProvider = countingButtonProvider{
		ComponentVisualProvider: testtheme.Behavior().VisualProvider.(theme.ComponentVisualProvider),
		calls:                   &calls,
		requests:                &requests,
	}
	value := New("").Content(nil).Variant(Primary)
	cases := []struct {
		name    string
		group   buttonGroupItemStyle
		corners theme.ButtonGroupRequest
	}{
		{
			name:    "horizontal-start",
			group:   buttonGroupItemStyle{grouped: true, orientation: GroupHorizontal, position: buttonGroupStart},
			corners: theme.ButtonGroupRequest{NorthWest: true, SouthWest: true},
		},
		{
			name:    "horizontal-middle",
			group:   buttonGroupItemStyle{grouped: true, orientation: GroupHorizontal, position: buttonGroupMiddle},
			corners: theme.ButtonGroupRequest{},
		},
		{
			name:    "horizontal-end",
			group:   buttonGroupItemStyle{grouped: true, orientation: GroupHorizontal, position: buttonGroupEnd},
			corners: theme.ButtonGroupRequest{NorthEast: true, SouthEast: true},
		},
		{
			name:    "vertical-start",
			group:   buttonGroupItemStyle{grouped: true, orientation: GroupVertical, position: buttonGroupStart},
			corners: theme.ButtonGroupRequest{NorthWest: true, NorthEast: true},
		},
		{
			name:    "vertical-middle",
			group:   buttonGroupItemStyle{grouped: true, orientation: GroupVertical, position: buttonGroupMiddle},
			corners: theme.ButtonGroupRequest{},
		},
		{
			name:    "vertical-end",
			group:   buttonGroupItemStyle{grouped: true, orientation: GroupVertical, position: buttonGroupEnd},
			corners: theme.ButtonGroupRequest{SouthEast: true, SouthWest: true},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			calls = 0
			requests = requests[:0]
			value.group = test.group
			value.declarationsForOverride(&activeTheme, nil, nil)
			if calls != 1 || len(requests) != 1 {
				t.Fatalf("group position provider calls = %d, requests = %d; want one each", calls, len(requests))
			}
			request := requests[0]
			if !request.Grouped || request.GroupCorners != test.corners {
				t.Fatalf("group request = %#v, want grouped corners %#v", request, test.corners)
			}
		})
	}
}

func TestStyledVerticalGroupReusesPreparedProviderDeclarations(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	calls := 0
	activeTheme.VisualProvider = countingButtonProvider{
		ComponentVisualProvider: testtheme.Behavior().VisualProvider.(theme.ComponentVisualProvider),
		calls:                   &calls,
	}
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(320, 120)}, Ops: new(op.Ops)}

	(Group(
		New("").Content(nil),
		New("").Content(nil),
	).Orientation(GroupVertical)).Style(style.Style{}.PaddingX(8)).Layout(ctx, gtx)

	if calls != 2 {
		t.Fatalf("styled vertical group provider calls = %d, want one per button", calls)
	}
}
