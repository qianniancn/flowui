package input

import (
	"image"
	"reflect"
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

// TestInputInstanceStyleAddsFieldBorder pins the styling path where chrome
// enters through the provider-bound Styled wrapper (the
// cascade's instance layer), not through a design callback slot.
func TestInputInstanceStyleAddsFieldBorder(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	defaultCtx := core.New(nil, &activeTheme, locale.LanguageAuto)
	visual := inputVisualDeclarations(
		&activeTheme,
		Primary,
		false,
		theme.InputModeSingleLine,
		Medium,
		nil,
	)
	declaration := visual.Defaults

	defaultStyle := computed.ResolveStatic(defaultCtx, style.State{}, declaration, style.Style{}, style.Style{}, style.Style{})
	instanceChrome := style.Style{}.
		BorderColor(style.Solid(activeTheme.Palette.Border)).
		BorderWidth(1)
	styledStyle := computed.ResolveStatic(defaultCtx, style.State{}, declaration, style.Style{}, style.Style{}, instanceChrome)

	if defaultStyle.Paint == nil || defaultStyle.Paint.Border != nil {
		t.Fatal("default input unexpectedly has external style border")
	}
	if styledStyle.Paint == nil || styledStyle.Paint.Border == nil {
		t.Fatal("instance input style did not add border chrome")
	}
}

func TestInputEditorStyleResolverProvidesDefaults(t *testing.T) {
	active := *testtheme.Behavior()
	visual := inputVisualDeclarations(&active, Primary, false, theme.InputModeSingleLine, Medium, nil)
	if visual.EditorStyle == nil {
		t.Fatal("input provider returned no editor style resolver")
	}
	resolved := visual.EditorStyle(style.ResolvedStyle{}, style.ResolvedStyle{}, style.ResolvedStyle{})
	if resolved.Foreground.A == 0 || resolved.Placeholder.A == 0 || resolved.Selection.A == 0 || resolved.TextSize <= 0 || resolved.LineHeight <= 0 {
		t.Fatalf("provider editor defaults = %#v", resolved)
	}
}

func TestInputStyledWrapperKeepsOverrideOutsideBehaviorValue(t *testing.T) {
	provider := &recordingInputVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := *testtheme.Behavior()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)
	declaration := style.Style{}.BorderWidth(2)
	base := New("").Key("wrapped")
	wrapped := (base).Style(declaration)

	wrapped.Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(180, 48)},
		Ops:         new(op.Ops),
	})
	if len(provider.requests) == 0 || provider.requests[len(provider.requests)-1].Override == nil {
		t.Fatal("styled wrapper did not send an instance override to the provider")
	}
}

func TestStyledTextAreaAndGroupSendOverridesToProvider(t *testing.T) {
	tests := []struct {
		name   string
		layout func(*core.Context, layout.Context)
	}{
		{
			name: "textarea",
			layout: func(ctx *core.Context, gtx layout.Context) {
				(NewTextArea("").Key("wrapped-textarea").Rows(2)).Style(style.Style{}.BorderWidth(2)).Layout(ctx, gtx)
			},
		},
		{
			name: "group",
			layout: func(ctx *core.Context, gtx layout.Context) {
				(Group(New("").Key("wrapped-group"))).Style(style.Style{}.BorderWidth(2)).Layout(ctx, gtx)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &recordingInputVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
			active := *testtheme.Behavior()
			active.VisualProvider = provider
			ctx := core.New(nil, &active, locale.LanguageAuto)
			test.layout(ctx, layout.Context{
				Constraints: layout.Constraints{Max: image.Pt(180, 96)},
				Ops:         new(op.Ops),
			})
			if len(provider.requests) == 0 {
				t.Fatal("styled wrapper did not call the provider")
			}
			request := provider.requests[len(provider.requests)-1]
			if request.Override == nil || request.Override.Declaration.Hash64() == (style.Style{}).Hash64() {
				t.Fatal("styled wrapper did not send its declaration to the provider")
			}
		})
	}
}

func TestStyledInputUsesOneProviderBundlePerFrame(t *testing.T) {
	provider := &recordingInputVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := *testtheme.Behavior()
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageAuto)

	(New("Ada").Key("single-bundle")).Style(style.Style{}.BorderWidth(2)).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(180, 48)},
		Ops:         new(op.Ops),
	})
	if got := len(provider.requests); got != 1 {
		t.Fatalf("styled input provider requests = %d, want one cached visual bundle", got)
	}
}

func TestInputVisualDeclarationCacheTracksOverride(t *testing.T) {
	provider := &recordingInputVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := *testtheme.Behavior()
	active.VisualProvider = provider
	state := &inputState{}
	value := New("Ada").Key("override-cache")

	state.visualDeclarations(value, &active, nil)
	state.visualDeclarations(value, &active, nil)
	if got := len(provider.requests); got != 1 {
		t.Fatalf("unstyled visual bundle requests = %d, want one cache hit", got)
	}
	first := style.Style{}.BorderWidth(1)
	second := style.Style{}.BorderWidth(2)
	state.visualDeclarations(value, &active, &theme.InputOverride{Declaration: first})
	state.visualDeclarations(value, &active, &theme.InputOverride{Declaration: second})
	if got := len(provider.requests); got != 3 {
		t.Fatalf("override visual bundle requests = %d, want one request per distinct override", got)
	}
}

func TestStyleDeclarationCacheLoadAndInvalidation(t *testing.T) {
	firstTheme := *testtheme.Behavior()
	secondTheme := *testtheme.Behavior()
	declaration := style.Style{}.Height(unit.Dp(24))
	cache := computed.ProviderCache[inputVisualKey, theme.InputVisual]{}
	key := inputVisualKey{variant: field.Primary, sizeTier: Medium}
	cache.Resolve(&firstTheme, key, func() theme.InputVisual {
		return theme.InputVisual{Defaults: declaration}
	})

	tests := []struct {
		name          string
		theme         *theme.Theme
		variant       field.Variant
		sizeTier      Size
		fullWidth     bool
		dynamicHeight unit.Dp
		wantHit       bool
	}{
		{name: "same key", theme: &firstTheme, variant: field.Primary, sizeTier: Medium, wantHit: true},
		{name: "theme pointer", theme: &secondTheme, variant: field.Primary},
		{name: "variant", theme: &firstTheme, variant: field.Secondary},
		{name: "size tier", theme: &firstTheme, variant: field.Primary, sizeTier: Small},
		{name: "full width", theme: &firstTheme, variant: field.Primary, fullWidth: true},
		{name: "dynamic height", theme: &firstTheme, variant: field.Primary, dynamicHeight: 24},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before := cache.Stats()
			cache.Resolve(test.theme, inputVisualKey{
				variant: test.variant, sizeTier: test.sizeTier,
				fullWidth: test.fullWidth, dynamicHeight: test.dynamicHeight,
			}, func() theme.InputVisual {
				return theme.InputVisual{Defaults: declaration}
			})
			after := cache.Stats()
			gotHit := after.Misses == before.Misses
			if gotHit != test.wantHit {
				t.Fatalf("cache hit = %v, want %v (stats before=%#v after=%#v)", gotHit, test.wantHit, before, after)
			}
		})
	}
}

func TestInputStateStyleDeclarationsCache(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	otherTheme := *testtheme.Behavior()
	state := &inputState{}
	widget := New("").Key("name").Variant(Secondary)

	firstVisual := state.visualDeclarations(widget, &activeTheme, nil)
	secondVisual := state.visualDeclarations(widget, &activeTheme, nil)
	if firstVisual.Defaults.Hash64() != secondVisual.Defaults.Hash64() {
		t.Fatal("repeated input declaration lookup changed the cached layers")
	}
	stats := state.styleCache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("input cache stats after repeated lookup = %#v", stats)
	}
	state.visualDeclarations(widget.Size(Small), &activeTheme, nil)
	stats = state.styleCache.Stats()
	if stats.Misses != 2 || stats.KeyInvalidations == 0 {
		t.Fatalf("input size change did not invalidate cache: %#v", stats)
	}

	state.visualDeclarations(widget.FullWidth(), &activeTheme, nil)
	stats = state.styleCache.Stats()
	if stats.Misses != 3 || stats.KeyInvalidations < 2 {
		t.Fatalf("input full-width change did not invalidate cache: %#v", stats)
	}
	state.visualDeclarations(widget, &otherTheme, nil)
	stats = state.styleCache.Stats()
	if stats.Misses != 4 || stats.ThemeInvalidations == 0 {
		t.Fatalf("input theme change did not invalidate cache: %#v", stats)
	}
}

func TestInputEditorPartsBatchMatchesIndividualResolution(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testLayoutContext()
	state := style.State{Hovered: true, Focused: true}
	visual := inputVisualDeclarations(
		&activeTheme,
		Primary,
		false,
		theme.InputModeSingleLine,
		Medium,
		nil,
	)
	defaults := visual.Defaults
	custom := style.Style{}.
		TextColor(style.RGB(0x123456)).
		Part(style.PartPlaceholder, style.Style{}.TextColor(style.RGB(0x654321)))

	root, placeholder, selection := resolveInputEditorStyles(inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      "batch-input",
		state:    state,
		defaults: defaults,
		custom:   custom,
	})
	wantRoot := computed.Resolve(ctx, gtx, "individual-input", state, defaults, style.Style{}, style.Style{}, custom)
	wantPlaceholder := computed.ResolvePart(ctx, gtx, "individual-input", style.PartPlaceholder, state, defaults, style.Style{}, style.Style{}, custom)
	wantSelection := computed.ResolvePart(ctx, gtx, "individual-input", style.PartSelection, state, defaults, style.Style{}, style.Style{}, custom)
	if !reflect.DeepEqual(root, wantRoot) || !reflect.DeepEqual(placeholder, wantPlaceholder) || !reflect.DeepEqual(selection, wantSelection) {
		t.Fatalf("batch editor parts differ from individual resolution: got %#v/%#v/%#v want %#v/%#v/%#v", root, placeholder, selection, wantRoot, wantPlaceholder, wantSelection)
	}
}

func TestInputEditorRootResolutionMatchesBatchRoot(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	gtx := testLayoutContext()
	request := inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      "root-input",
		state:    style.State{Hovered: true, Focused: true},
		defaults: style.Style{}.TextColor(style.RGB(0x123456)),
		custom:   style.Style{}.Part(style.PartPlaceholder, style.Style{}.TextColor(style.RGB(0x654321))),
	}
	want, _, _ := resolveInputEditorStyles(request)
	if got := resolveInputEditorRootStyle(request); !reflect.DeepEqual(got, want) {
		t.Fatalf("root-only editor resolution differs from batch root: got %#v want %#v", got, want)
	}
}

func TestTextAreaStyleDeclarationsCacheTracksHeight(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	state := &textAreaState{}
	widget := NewTextArea("").Key("notes").Variant(TextAreaPrimary)

	state.visualDeclarations(widget, &activeTheme, 40, nil)
	stats := state.styleCache.Stats()
	if stats.Misses != 1 {
		t.Fatalf("textarea first height cache stats = %#v", stats)
	}
	state.visualDeclarations(widget, &activeTheme, 80, nil)
	stats = state.styleCache.Stats()
	if stats.Misses != 2 || stats.KeyInvalidations == 0 {
		t.Fatalf("textarea height change did not invalidate cache: %#v", stats)
	}
}

func TestTextAreaVisualCacheRetainsProvisionalAndResolvedHeights(t *testing.T) {
	provider := &recordingInputVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := *testtheme.Behavior()
	active.VisualProvider = provider
	state := &textAreaState{}
	widget := NewTextArea("").Key("notes-cache").Rows(4)

	state.visualDeclarations(widget, &active, 0, nil)
	state.visualDeclarations(widget, &active, 76, nil)
	state.visualDeclarations(widget, &active, 0, nil)
	state.visualDeclarations(widget, &active, 76, nil)

	if len(provider.requests) != 2 {
		t.Fatalf("textarea provisional/resolved provider requests = %d, want 2", len(provider.requests))
	}
	stats := state.styleCache.Stats()
	if stats.Misses != 2 || stats.Hits != 2 || stats.Size != 2 {
		t.Fatalf("textarea multi-height cache stats = %#v, want two misses/two hits/two entries", stats)
	}
}

func TestGroupStyleDeclarationsCacheTracksHeight(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	state := &inputGroupState{}
	widget := Group(New("").Key("query")).Variant(Primary)

	state.visualDeclarations(widget, &activeTheme, 36, nil)
	stats := state.styleCache.Stats()
	if stats.Misses != 1 {
		t.Fatalf("input-group first height cache stats = %#v", stats)
	}
	state.visualDeclarations(widget, &activeTheme, 76, nil)
	stats = state.styleCache.Stats()
	if stats.Misses != 2 || stats.KeyInvalidations == 0 {
		t.Fatalf("input-group height change did not invalidate cache: %#v", stats)
	}
}

func TestGroupVisualCacheRetainsProvisionalAndResolvedHeights(t *testing.T) {
	provider := &recordingInputVisualProvider{testInputVisualProvider: newTestInputVisualProvider()}
	active := *testtheme.Behavior()
	active.VisualProvider = provider
	state := &inputGroupState{}
	value := GroupTextArea(NewTextArea("").Key("group-cache").Rows(4))

	state.visualDeclarations(value, &active, 0, nil)
	state.visualDeclarations(value, &active, 76, nil)
	state.visualDeclarations(value, &active, 0, nil)
	state.visualDeclarations(value, &active, 76, nil)

	if len(provider.requests) != 2 {
		t.Fatalf("group provisional/resolved provider requests = %d, want 2", len(provider.requests))
	}
	stats := state.styleCache.Stats()
	if stats.Misses != 2 || stats.Hits != 2 || stats.Size != 2 {
		t.Fatalf("group multi-height cache stats = %#v, want two misses/two hits/two entries", stats)
	}
}
