package computed

import (
	"testing"

	"github.com/qianniancn/flowui/theme"
)

func TestDeclarationCacheInvalidatesThemeAndKey(t *testing.T) {
	firstTheme := theme.NeutralTheme()
	secondTheme := theme.NeutralTheme()
	cache := DeclarationCache[string, int]{}

	if _, ok := cache.Load(&firstTheme, "primary"); ok {
		t.Fatal("empty cache reported a hit")
	}

	cache.Store(&firstTheme, "primary", 42)
	if got, ok := cache.Load(&firstTheme, "primary"); !ok || got != 42 {
		t.Fatalf("cache hit = %v, %v; want 42, true", got, ok)
	}
	if _, ok := cache.Load(&firstTheme, "secondary"); ok {
		t.Fatal("key change reported a hit")
	}
	if _, ok := cache.Load(&secondTheme, "primary"); ok {
		t.Fatal("theme change reported a hit")
	}
}
