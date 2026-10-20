package core

import (
	"image"
	"testing"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/locale"
)

func TestTextMeasureCacheStoresAndSeparatesKeys(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageEnglish)
	key := TextMeasureKey{Text: "same", Constraints: layout.Constraints{Max: image.Pt(100, 40)}}
	want := layout.Dimensions{Size: image.Pt(42, 18), Baseline: 14}

	StoreTextMeasure(ctx, key, want)
	got, ok := LookupTextMeasure(ctx, key)
	if !ok || got != want {
		t.Fatalf("cached dimensions = %#v, hit %v; want %#v, true", got, ok, want)
	}

	key.Constraints.Max.X++
	if _, ok := LookupTextMeasure(ctx, key); ok {
		t.Fatal("measurement with different constraints unexpectedly hit cache")
	}
}

func TestTextMeasureCacheEvictsLeastRecentlyUsed(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageEnglish)
	for index := range textMeasureCacheCapacity {
		StoreTextMeasure(ctx, TextMeasureKey{Text: string(rune(index + 1))}, layout.Dimensions{Size: image.Pt(index, 1)})
	}
	if _, ok := LookupTextMeasure(ctx, TextMeasureKey{Text: string(rune(1))}); !ok {
		t.Fatal("oldest entry was not present before eviction")
	}
	StoreTextMeasure(ctx, TextMeasureKey{Text: "new"}, layout.Dimensions{})
	if _, ok := LookupTextMeasure(ctx, TextMeasureKey{Text: string(rune(1))}); !ok {
		t.Fatal("recently used entry was evicted")
	}
	if _, ok := LookupTextMeasure(ctx, TextMeasureKey{Text: string(rune(2))}); ok {
		t.Fatal("least recently used entry was not evicted")
	}
}

func TestTextMeasureCacheReset(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageEnglish)
	key := TextMeasureKey{Text: "theme-sensitive"}
	StoreTextMeasure(ctx, key, layout.Dimensions{Size: image.Pt(1, 1)})

	active := ctx.Theme()
	ReplaceTheme(ctx, active)
	if _, ok := LookupTextMeasure(ctx, key); ok {
		t.Fatal("theme replacement retained stale text measurement")
	}
}

func TestReplaceLanguageInvalidatesAppearanceCaches(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageEnglish)
	key := TextMeasureKey{Text: "locale-sensitive"}
	StoreTextMeasure(ctx, key, layout.Dimensions{Size: image.Pt(1, 1)})

	before := ThemeGeneration(ctx)
	ReplaceLanguage(ctx, locale.LanguageChinese)
	if got := ThemeGeneration(ctx); got == before {
		t.Fatal("language replacement did not advance the appearance generation")
	}
	if _, ok := LookupTextMeasure(ctx, key); ok {
		t.Fatal("language replacement retained stale text measurement")
	}
	if got := ActiveLanguage(ctx); got != locale.LanguageChinese {
		t.Fatalf("active language = %v, want %v", got, locale.LanguageChinese)
	}

	// Re-applying the resolved language must not churn component caches.
	generation := ThemeGeneration(ctx)
	ReplaceLanguage(ctx, locale.LanguageChinese)
	if got := ThemeGeneration(ctx); got != generation {
		t.Fatalf("redundant language replacement advanced generation: %d -> %d", generation, got)
	}
}
