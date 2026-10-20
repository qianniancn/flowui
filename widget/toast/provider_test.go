package toast

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestToastVisualCacheHitsAndInvalidatesByOverride(t *testing.T) {
	active := *testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	provider := &recordingToastProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	value := New("toasts", nil)
	value.visualWithCache(ctx, Success, nil)
	value.visualWithCache(ctx, Success, nil)
	if provider.calls != 1 {
		t.Fatalf("provider calls after cache hit = %d, want 1", provider.calls)
	}
	value.visualWithCache(ctx, Success, &theme.ToastOverride{Declaration: style.Style{}.Background(style.Solid(color.NRGBA{A: 0xff}))})
	if provider.calls != 2 {
		t.Fatalf("provider calls after override = %d, want 2", provider.calls)
	}
	stats := value.visualCache.stats()
	if stats.Hits != 1 || stats.Misses != 2 || stats.KeyInvalidations == 0 {
		t.Fatalf("cache stats = %#v", stats)
	}
}

func TestToastProviderMissingFailsBeforeLayout(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	defer func() {
		if recover() == nil {
			t.Fatal("missing toast provider did not panic")
		}
	}()
	New("toasts", nil).visualWithCache(ctx, Default, nil)
}

func TestToastLayoutResolvesEachVariantOnce(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	provider := New("toasts", []Item{
		NewItem("default-1", "Default 1").Timeout(0),
		NewItem("default-2", "Default 2").Timeout(0),
		NewItem("success-1", "Success 1").Variant(Success).Timeout(0),
		NewItem("success-2", "Success 2").Variant(Success).Timeout(0),
	})
	layoutToastFrame(ctx, router, provider, time.Unix(1, 0), image.Pt(900, 500))

	stats := provider.visualCache.stats()
	if stats.Misses != 2 || stats.Hits != 0 {
		t.Fatalf("toast visual cache stats = %#v, want one miss per variant and no same-frame hits", stats)
	}
}

type recordingToastProvider struct {
	theme.ComponentVisualProvider
	calls int
}

func (p *recordingToastProvider) Toast(active *theme.Theme, request theme.ToastRequest) theme.ToastVisual {
	p.calls++
	return p.ComponentVisualProvider.Toast(active, request)
}
