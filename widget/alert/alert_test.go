package alert

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	gioinput "gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/uitest"
	"github.com/qianniancn/flowui/widget/button"
)

type alertProbe struct {
	size       image.Point
	foreground color.NRGBA
	background color.NRGBA
	layouts    int
}

func (p *alertProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func newAlertContext(activeTheme *theme.Theme) *core.Context {
	if activeTheme == nil {
		activeTheme = testtheme.Behavior()
	}
	return core.New(nil, activeTheme, locale.LanguageEnglish)
}

func alertTestContext() layout.Context {
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(480, 200)},
		Ops:         &ops,
	}
}

func TestAlertFillsWidthAndHeight(t *testing.T) {
	dims := New("New features available", "Check out the latest updates.").Layout(newAlertContext(nil), alertTestContext())
	if dims.Size.X != 480 {
		t.Fatalf("alert width = %d, want 480", dims.Size.X)
	}
	if dims.Size.Y != 68 {
		t.Fatalf("two-line alert height = %d, want 68", dims.Size.Y)
	}

	dims = New("Profile updated successfully", "").Layout(newAlertContext(nil), alertTestContext())
	if dims.Size.Y != 48 {
		t.Fatalf("title-only alert height = %d, want 48", dims.Size.Y)
	}
}

func TestAlertCustomSlotsReceiveScopedColors(t *testing.T) {
	indicator := &alertProbe{size: image.Pt(16, 16)}
	content := &alertProbe{size: image.Pt(120, 20)}
	action := &alertProbe{size: image.Pt(64, 32)}
	activeTheme := testtheme.Behavior()
	New("Processing", "ignored").
		Status(Accent).
		Indicator(indicator).
		Content(content).
		Action(action).
		Layout(newAlertContext(activeTheme), alertTestContext())

	if indicator.layouts != 1 || content.layouts != 1 || action.layouts != 1 {
		t.Fatalf("slot layouts = indicator %d content %d action %d", indicator.layouts, content.layouts, action.layouts)
	}
	if indicator.foreground == content.foreground || action.foreground == content.foreground {
		t.Fatalf("slot colors were not scoped: indicator %#v content %#v action %#v", indicator.foreground, content.foreground, action.foreground)
	}
	backgrounds := map[string]color.NRGBA{
		"indicator": indicator.background,
		"content":   content.background,
		"action":    action.background,
	}
	for name, got := range backgrounds {
		if got != indicator.background {
			t.Fatalf("%s background = %#v, want %#v", name, got, indicator.background)
		}
	}
}

func TestAlertVisualCacheTracksStatusOverrideAndTheme(t *testing.T) {
	active := testtheme.Behavior()
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("cache", "description")
	cache := alertVisualCache{}
	value.visualCache = &cache

	value.visualWithCache(ctx, nil)
	value.visualWithCache(ctx, nil)
	value.Status(Success).visualWithCache(ctx, nil)
	override := &theme.AlertOverride{Declaration: style.Style{}.Padding(2)}
	value.visualWithCache(ctx, override)
	replacement := *active
	core.ReplaceTheme(ctx, replacement)
	value.visualWithCache(ctx, override)

	stats := cache.stats()
	if stats.Hits != 1 || stats.Misses != 4 || stats.ThemeInvalidations != 1 || stats.KeyInvalidations != 2 {
		t.Fatalf("alert visual cache stats = %#v", stats)
	}
}

func TestAlertProviderReceivesStatusAndOverride(t *testing.T) {
	active := testtheme.Behavior()
	provider := &recordingAlertProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	active.VisualProvider = provider
	ctx := core.New(nil, active, locale.LanguageEnglish)
	override := &theme.AlertOverride{Declaration: style.Style{}.Padding(2)}
	New("status", "description").Status(Danger).visualWithCache(ctx, override)

	if len(provider.requests) != 1 {
		t.Fatalf("alert provider requests = %d, want one", len(provider.requests))
	}
	request := provider.requests[0]
	if request.Status != theme.AlertDanger || request.Override != override {
		t.Fatalf("alert provider request = %#v", request)
	}
}

type recordingAlertProvider struct {
	theme.ComponentVisualProvider
	requests []theme.AlertRequest
}

func (p *recordingAlertProvider) Alert(active *theme.Theme, request theme.AlertRequest) theme.AlertVisual {
	p.requests = append(p.requests, request)
	return p.ComponentVisualProvider.Alert(active, request)
}

func TestAlertSemanticsExposeTitleAndDescription(t *testing.T) {
	ctx := newAlertContext(nil)
	router := new(gioinput.Router)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(480, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	New("Storage almost full", "You are using 90% of your quota.").Layout(ctx, gtx)
	router.Frame(&ops)
	if !alertSemanticTreeContains(router.AppendSemantics(nil), "Storage almost full", "You are using 90% of your quota.") {
		t.Fatal("alert semantics did not expose its title and description")
	}
}

func TestAlertActionButtonRemainsInteractive(t *testing.T) {
	ctx := newAlertContext(nil)
	router := new(gioinput.Router)
	clicked := false
	action := button.New("").Content(&alertProbe{size: image.Pt(64, 16)}).
		Size(button.Small).
		OnClick(func() { clicked = true })
	alert := New("Update available", "Refresh to get the latest version.").Action(action)
	layoutAlertTestFrame(ctx, router, alert, time.Unix(1, 0))
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(432, 28)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(432, 28)},
	)
	layoutAlertTestFrame(ctx, router, alert, time.Unix(1, int64(time.Millisecond)))
	if !clicked {
		t.Fatal("alert action button did not receive the pointer click")
	}
}

func TestAlertClosableCallsOnClose(t *testing.T) {
	h := uitest.New(image.Pt(480, 200), base.Light())
	closed := false
	root := New("Update available", "Refresh to get the latest version.").
		Closable(true).
		OnClose(func() { closed = true })
	h.Frame(root)
	h.Frame(root)

	node, ok := uitest.SemanticByLabel(h.Semantics(), "Close")
	if !ok {
		t.Fatal("alert close button is missing")
	}
	bounds := node.Desc.Bounds
	h.Click(f32.Pt(float32(bounds.Min.X+bounds.Max.X)/2, float32(bounds.Min.Y+bounds.Max.Y)/2))
	h.Frame(root)
	if !closed {
		t.Fatal("alert close button did not call OnClose")
	}
}

func layoutAlertTestFrame(ctx *core.Context, router *gioinput.Router, alert Widget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(480, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	alert.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func alertSemanticTreeContains(nodes []gioinput.SemanticNode, label, description string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label && node.Desc.Description == description {
			return true
		}
		if alertSemanticTreeContains(node.Children, label, description) {
			return true
		}
	}
	return false
}
