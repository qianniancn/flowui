package computed

import (
	"encoding/binary"
	"hash/fnv"
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestHashStyleLayersPreservesFNVByteOrder(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.TextColor(style.RGB(0x000000))
	state := style.State{Hovered: true, Focused: true, Open: true}

	got := hashStyleLayers(ctx, state, defaults, variant, size, custom)
	wantHash := fnv.New64a()
	writeStyle := func(value style.Style) {
		var bytes [8]byte
		binary.LittleEndian.PutUint64(bytes[:], value.Hash64ForState(state))
		_, _ = wantHash.Write(bytes[:])
	}
	writeStyle(defaults)
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		writeStyle(value)
	}
	writeStyle(variant)
	writeStyle(size)
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		writeStyle(value)
	}
	writeStyle(custom)

	if got != wantHash.Sum64() {
		t.Fatalf("hashStyleLayers() = %x, want FNV-1a result %x", got, wantHash.Sum64())
	}
}

func BenchmarkHashStyleLayers(b *testing.B) {
	activeTheme := theme.NeutralTheme()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	defaults := style.Style{}.Padding(8).Background(style.RGB(0xffffff))
	variant := style.Style{}.PaddingX(16)
	size := style.Style{}.Height(36)
	custom := style.Style{}.TextColor(style.RGB(0x000000))

	for _, benchmark := range []struct {
		name     string
		hasScope bool
		scope    style.Style
		custom   style.Style
	}{
		{name: "NoScopes", custom: custom},
		{name: "WithScopes", hasScope: true, scope: style.Style{}.TextColor(style.RGB(0x202020)), custom: custom},
		{name: "WithParts", custom: style.Style{}.Part(style.PartLabel, style.Style{}.TextColor(style.RGB(0x202020)))},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			var restore func()
			if benchmark.hasScope {
				restore = core.PushStyle(ctx, benchmark.scope)
				defer restore()
			}
			b.ReportAllocs()
			for b.Loop() {
				_ = hashStyleLayers(ctx, style.State{Hovered: true}, defaults, variant, size, benchmark.custom)
			}
		})
	}
}
