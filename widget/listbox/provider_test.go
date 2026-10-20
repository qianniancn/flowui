package listbox

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type recordingListBoxVisualProvider struct {
	theme.ComponentVisualProvider
	requests  []theme.ListBoxRequest
	itemParts []style.ResolvedStyle
}

func (p *recordingListBoxVisualProvider) ListBox(active *theme.Theme, request theme.ListBoxRequest) theme.ListBoxVisual {
	p.requests = append(p.requests, request)
	visual := p.ComponentVisualProvider.ListBox(active, request)
	visual.Override = theme.ListBoxOverrideStyle(request)
	item := visual.Item
	visual.Item = func(request theme.ListBoxItemRequest) theme.ListBoxItemVisual {
		result := item(request)
		colors := result.Colors
		result.Colors = func(resolved style.ResolvedStyle) theme.ListBoxItemColors {
			p.itemParts = append(p.itemParts, resolved)
			return colors(resolved)
		}
		return result
	}
	return visual
}

func TestListBoxInstanceOverrideUsesProviderLayer(t *testing.T) {
	active := testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	provider := &recordingListBoxVisualProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, active, locale.LanguageEnglish)

	(New([]Item{{Key: "one", Label: "One"}}).Key("wrapped").SelectedKey("one")).Style(style.Style{}.
		Part(style.PartItem, style.Style{}.Height(48))).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(180, 120)},
		Ops:         new(op.Ops),
	})

	for _, request := range provider.requests {
		if request.Override != nil {
			resolved := style.CascadePart(style.State{}, style.PartItem, theme.ListBoxOverrideStyle(request))
			if resolved.Box == nil || resolved.Box.Height == nil || *resolved.Box.Height != 48 {
				t.Fatalf("provider override = %#v, want item height 48", resolved.Box)
			}
			return
		}
	}
	t.Fatal("styled listbox did not send an instance override to the provider")
}

func TestListBoxStyledResolvesProviderOncePerLayout(t *testing.T) {
	active := testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	provider := &recordingListBoxVisualProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, active, locale.LanguageEnglish)

	(New([]Item{{Key: "one", Label: "One"}}).Key("once").SelectedKey("one")).Style(style.Style{}.
		Part(style.PartItem, style.Style{}.Height(48))).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(180, 120)},
		Ops:         new(op.Ops),
	})

	if got := len(provider.requests); got != 1 {
		t.Fatalf("ListBox Provider requests = %d, want exactly one", got)
	}
}

func TestListBoxStyledPartsEmptyDeclarationCachesProviderVisualAcrossFrames(t *testing.T) {
	active := testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	provider := &recordingListBoxVisualProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, active, locale.LanguageEnglish)
	styled := (New([]Item{{Key: "one", Label: "One"}}).Key("parts-only").SelectedKey("one")).StyleParts(style.Style{})

	for range 2 {
		var ops op.Ops
		gtx := layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(180, 120)},
			Ops:         &ops,
		}
		core.BeginFrame(ctx)
		styled.Layout(ctx, gtx)
		core.EndFrame(ctx)
	}

	if got := len(provider.requests); got != 1 {
		t.Fatalf("ListBox Provider requests for empty parts declaration = %d, want 1 across frames", got)
	}
	if provider.requests[0].Override != nil {
		t.Fatalf("empty parts declaration sent override %#v, want no override", provider.requests[0].Override)
	}
}

func TestListBoxStyledWrapperKeepsOverrideOutsideBehaviorValue(t *testing.T) {
	base := New([]Item{{Key: "one", Label: "One"}}).Key("wrapped").SelectedKey("one")
	wrapped := (base).Style(style.Style{}.Height(48))
	if wrapped.key != base.key || wrapped.selectedKey != base.selectedKey {
		t.Fatalf("styled wrapper changed behavior value: %#v", wrapped)
	}
	if wrapped.partsOnly {
		t.Fatal("direct Styled wrapper unexpectedly uses parts-only mode")
	}
}

func TestListBoxPartsOverrideDoesNotStyleRoot(t *testing.T) {
	base := New([]Item{{Key: "one", Label: "One"}}).Key("embedded").SelectedKey("one")

	measure := func(declaration style.Style) (image.Point, []style.ResolvedStyle) {
		active := testtheme.Behavior()
		baseProvider := active.VisualProvider.(theme.ComponentVisualProvider)
		provider := &recordingListBoxVisualProvider{ComponentVisualProvider: baseProvider}
		active.VisualProvider = provider
		ctx := core.New(nil, active, locale.LanguageEnglish)
		gtx := layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(180, 240)},
			Ops:         new(op.Ops),
		}
		size := (base).StyleParts(declaration).Layout(ctx, gtx).Size
		return size, provider.itemParts
	}

	baseline, _ := measure(style.Style{})
	rootOverride, _ := measure(style.Style{}.Height(160))
	if rootOverride != baseline {
		t.Fatalf("parts-only root override changed ListBox size: got %v, want natural size %v", rootOverride, baseline)
	}

	itemSize, itemParts := measure(style.Style{}.Part(style.PartItem, style.Style{}.MinHeight(88)))
	if len(itemParts) == 0 || itemParts[len(itemParts)-1].Box == nil || itemParts[len(itemParts)-1].Box.MinHeight == nil {
		t.Fatalf("ListBox layout did not resolve an Item Part: %#v", itemParts)
	}
	if got := *itemParts[len(itemParts)-1].Box.MinHeight; got != 88 {
		t.Fatalf("resolved Item Part min height = %v, want 88 (layout size %v, baseline %v)", got, itemSize, baseline)
	}
	if itemSize.Y <= baseline.Y {
		t.Fatalf("item Part override did not affect ListBox layout height: got %v, baseline %v", itemSize, baseline)
	}
}
