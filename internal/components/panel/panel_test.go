package panel

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/state"
)

func TestHostLaysOutOnlySelectedPanel(t *testing.T) {
	ctx := panelTestContext()
	router := new(input.Router)
	first := new(panelProbe)
	second := new(panelProbe)
	host := Host("settings", "account", []Item{
		{Key: "account", Content: first},
		{Key: "security", Content: second},
	})

	layoutHostFrame(ctx, router, host, time.Unix(1, 0))
	if first.layouts != 1 || second.layouts != 0 {
		t.Fatalf("layouts = %d/%d, want 1/0", first.layouts, second.layouts)
	}
	if first.key == second.key || first.key == "" {
		t.Fatalf("panel keys = %q/%q, want distinct scoped keys", first.key, second.key)
	}
}

func TestHostDefaultSelectionAndFallback(t *testing.T) {
	ctx := panelTestContext()
	router := new(input.Router)
	first := new(panelProbe)
	second := new(panelProbe)
	changed := ""
	host := Host("settings", "", []Item{
		{Key: "account", Content: first, Disabled: true},
		{Key: "security", Content: second},
	}).
		DefaultSelectedKey("account").
		OnChange(func(key string) { changed = key })

	layoutHostFrame(ctx, router, host, time.Unix(1, 0))
	if first.layouts != 0 || second.layouts != 1 {
		t.Fatalf("fallback layouts = %d/%d, want 0/1", first.layouts, second.layouts)
	}
	if changed != "security" {
		t.Fatalf("fallback change = %q, want security", changed)
	}
}

func TestHostForceRenderInitializesHiddenPanels(t *testing.T) {
	ctx := panelTestContext()
	router := new(input.Router)
	selected := "account"
	account := new(panelProbe)
	security := new(panelProbe)
	host := func() HostWidget {
		return Host("settings", selected, []Item{
			{Key: "account", Content: account},
			{Key: "security", Content: security},
		}).ForceRender(true)
	}

	layoutHostFrame(ctx, router, host(), time.Unix(1, 0))
	if account.layouts != 1 || security.layouts != 1 {
		t.Fatalf("force-render layouts = %d/%d, want 1/1", account.layouts, security.layouts)
	}
	if _, ok := frame.PeekState[widget.Clickable](ctx, security.key, "probe"); !ok {
		t.Fatal("hidden panel state was not retained")
	}

	selected = "security"
	layoutHostFrame(ctx, router, host(), time.Unix(1, int64(time.Millisecond)))
	if account.layouts != 1 || security.layouts != 2 {
		t.Fatalf("after selection layouts = %d/%d, want 1/2", account.layouts, security.layouts)
	}
}

func TestHostKeepAliveAndDestroyOnHidden(t *testing.T) {
	t.Run("retains", func(t *testing.T) {
		ctx := panelTestContext()
		router := new(input.Router)
		selected := "account"
		account := new(panelProbe)
		security := new(panelProbe)
		host := func() HostWidget {
			return Host("settings", selected, []Item{
				{Key: "account", Content: account},
				{Key: "security", Content: security},
			}).KeepAlive(true)
		}
		layoutHostFrame(ctx, router, host(), time.Unix(1, 0))
		first := account.state
		selected = "security"
		layoutHostFrame(ctx, router, host(), time.Unix(1, int64(time.Millisecond)))
		retained, ok := frame.PeekState[widget.Clickable](ctx, account.key, "probe")
		if !ok || retained != first {
			t.Fatal("hidden panel state was not retained")
		}
	})

	t.Run("destroys", func(t *testing.T) {
		ctx := panelTestContext()
		router := new(input.Router)
		selected := "account"
		account := new(panelProbe)
		security := new(panelProbe)
		host := func() HostWidget {
			return Host("settings", selected, []Item{
				{Key: "account", Content: account},
				{Key: "security", Content: security},
			}).KeepAlive(true).DestroyOnHidden(true)
		}
		layoutHostFrame(ctx, router, host(), time.Unix(1, 0))
		selected = "security"
		layoutHostFrame(ctx, router, host(), time.Unix(1, int64(time.Millisecond)))
		if _, ok := frame.PeekState[widget.Clickable](ctx, account.key, "probe"); ok {
			t.Fatal("destroy-on-hidden retained panel state")
		}
	})
}

func TestHostRemovedPanelReleasesRetainedState(t *testing.T) {
	ctx := panelTestContext()
	router := new(input.Router)
	selected := "account"
	account := new(panelProbe)
	security := new(panelProbe)
	items := []Item{
		{Key: "account", Content: account},
		{Key: "security", Content: security},
	}
	host := func() HostWidget { return Host("settings", selected, items).KeepAlive(true) }

	layoutHostFrame(ctx, router, host(), time.Unix(1, 0))
	selected = "security"
	layoutHostFrame(ctx, router, host(), time.Unix(1, int64(time.Millisecond)))
	items = items[:1]
	selected = "account"
	layoutHostFrame(ctx, router, host(), time.Unix(1, int64(2*time.Millisecond)))
	if _, ok := frame.PeekState[widget.Clickable](ctx, security.key, "probe"); ok {
		t.Fatal("removed panel state was retained")
	}
}

type panelProbe struct {
	layouts int
	key     string
	state   *widget.Clickable
}

func (p *panelProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.key = frame.ClaimKey(ctx, state.KindCustom, "control")
	p.state = frame.UseState[widget.Clickable](ctx, p.key, "probe")
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(20, 20))}
}

func panelTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func layoutHostFrame(ctx *frame.Context, router *input.Router, widget HostWidget, now time.Time) layout.Dimensions {
	var ops op.Ops
	size := image.Pt(300, 200)
	gtx := layout.Context{Constraints: layout.Exact(size), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, size)
	dims := widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
