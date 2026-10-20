package input

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

type inputBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
}

func newInputBenchmarkFrame(maximum image.Point) *inputBenchmarkFrame {
	activeTheme := *testtheme.Behavior()
	frame := &inputBenchmarkFrame{
		ctx: core.New(nil, &activeTheme, locale.LanguageEnglish),
	}
	frame.gtx = layout.Context{
		Constraints: layout.Constraints{Max: maximum},
		Source:      frame.router.Source(),
		Ops:         &frame.ops,
		Now:         time.Unix(1, 0),
	}
	return frame
}

func (f *inputBenchmarkFrame) layout(value interface {
	Layout(*core.Context, layout.Context) layout.Dimensions
}) layout.Dimensions {
	f.ops.Reset()
	core.BeginFrame(f.ctx)
	dimensions := value.Layout(f.ctx, f.gtx)
	core.ApplyFrameCommands(f.ctx, f.gtx)
	core.EndFrame(f.ctx)
	f.router.Frame(&f.ops)
	return dimensions
}

func BenchmarkInputLayoutSteadyState(b *testing.B) {
	frame := newInputBenchmarkFrame(image.Pt(300, 80))
	value := New("Ada").Key("name").Hint("Name")
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkTextAreaLayoutSteadyState(b *testing.B) {
	frame := newInputBenchmarkFrame(image.Pt(300, 160))
	value := NewTextArea("Ada\nLovelace\nAnalytical Engine").Key("notes").Rows(4).Hint("Notes")
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

// BenchmarkInputGroupLayoutSteadyState measures a composed field shell after
// its provider declarations and editor state have been warmed.
func BenchmarkInputGroupLayoutSteadyState(b *testing.B) {
	frame := newInputBenchmarkFrame(image.Pt(360, 80))
	value := Group(New("Ada").Key("query").Hint("Search")).
		Prefix(inputGroupFixedWidget{size: image.Pt(20, 20)}).
		Suffix(inputGroupFixedWidget{size: image.Pt(24, 20)})
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkInputGroupTextAreaLayoutSteadyState(b *testing.B) {
	frame := newInputBenchmarkFrame(image.Pt(360, 160))
	value := GroupTextArea(NewTextArea("Ada\nLovelace\nEngine").Key("notes").Rows(4).Hint("Notes")).
		Prefix(inputGroupFixedWidget{size: image.Pt(20, 20)}).
		Suffix(inputGroupFixedWidget{size: image.Pt(24, 20)})
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

// BenchmarkInputGroupActionLayoutSteadyState measures the interactive action
// slots, including their chrome and input-focus routing setup.
func BenchmarkInputGroupActionLayoutSteadyState(b *testing.B) {
	frame := newInputBenchmarkFrame(image.Pt(360, 80))
	value := Group(New("Ada").Key("query").Hint("Search")).
		PrefixAction(GroupAction("filter", "Filter values").Content(inputGroupFixedWidget{size: image.Pt(16, 16)})).
		SuffixAction(GroupAction("clear", "Clear value").Content(inputGroupFixedWidget{size: image.Pt(16, 16)})).
		FocusOnActionPress(true)
	frame.layout(value)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.layout(value)
	}
}

func BenchmarkInputVisualDeclarations(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	value := New("Ada").Key("name")

	b.Run("cache-hit", func(b *testing.B) {
		state := &inputState{}
		state.visualDeclarations(value, &activeTheme, nil)
		baseline := state.styleCache.Stats()
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			state.visualDeclarations(value, &activeTheme, nil)
		}
		b.StopTimer()
		reportInputDeclarationCacheStats(b, "input-declaration-cache", state.styleCache.Stats(), baseline)
	})

	b.Run("cache-miss", func(b *testing.B) {
		state := &inputState{}
		b.ReportAllocs()
		b.ResetTimer()
		for index := 0; b.Loop(); index++ {
			variant := Primary
			if index&1 != 0 {
				variant = Secondary
			}
			state.visualDeclarations(value.Variant(variant), &activeTheme, nil)
		}
		b.StopTimer()
		reportInputDeclarationCacheStats(b, "input-declaration-cache", state.styleCache.Stats(), computed.ProviderCacheStats{})
	})
}

func BenchmarkTextAreaVisualDeclarations(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	value := NewTextArea("Ada").Key("notes").Rows(4)

	b.Run("cache-hit", func(b *testing.B) {
		state := &textAreaState{}
		state.visualDeclarations(value, &activeTheme, unit.Dp(80), nil)
		baseline := state.styleCache.Stats()
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			state.visualDeclarations(value, &activeTheme, unit.Dp(80), nil)
		}
		b.StopTimer()
		reportInputDeclarationCacheStats(b, "textarea-declaration-cache", state.styleCache.Stats(), baseline)
	})

	b.Run("cache-miss", func(b *testing.B) {
		state := &textAreaState{}
		b.ReportAllocs()
		b.ResetTimer()
		for index := 0; b.Loop(); index++ {
			height := unit.Dp(40 + index%9)
			state.visualDeclarations(value, &activeTheme, height, nil)
		}
		b.StopTimer()
		reportInputDeclarationCacheStats(b, "textarea-declaration-cache", state.styleCache.Stats(), computed.ProviderCacheStats{})
	})
}

func BenchmarkGroupVisualDeclarations(b *testing.B) {
	activeTheme := *testtheme.Behavior()
	value := Group(New("Ada").Key("query"))

	b.Run("cache-hit", func(b *testing.B) {
		state := &inputGroupState{}
		state.visualDeclarations(value, &activeTheme, unit.Dp(36), nil)
		baseline := state.styleCache.Stats()
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			state.visualDeclarations(value, &activeTheme, unit.Dp(36), nil)
		}
		b.StopTimer()
		reportInputDeclarationCacheStats(b, "input-group-declaration-cache", state.styleCache.Stats(), baseline)
	})

	b.Run("cache-miss", func(b *testing.B) {
		state := &inputGroupState{}
		b.ReportAllocs()
		b.ResetTimer()
		for index := 0; b.Loop(); index++ {
			minHeight := unit.Dp(36 + index%9)
			state.visualDeclarations(value, &activeTheme, minHeight, nil)
		}
		b.StopTimer()
		reportInputDeclarationCacheStats(b, "input-group-declaration-cache", state.styleCache.Stats(), computed.ProviderCacheStats{})
	})
}

func reportInputDeclarationCacheStats(b *testing.B, prefix string, after, before computed.ProviderCacheStats) {
	b.ReportMetric(float64(after.Hits-before.Hits)/float64(b.N), prefix+"-hits/op")
	b.ReportMetric(float64(after.Misses-before.Misses)/float64(b.N), prefix+"-misses/op")
	b.ReportMetric(float64(after.ThemeInvalidations-before.ThemeInvalidations)/float64(b.N), prefix+"-theme-invalidations/op")
	b.ReportMetric(float64(after.KeyInvalidations-before.KeyInvalidations)/float64(b.N), prefix+"-key-invalidations/op")
}

func BenchmarkInputDeclarationCacheInvalidation(b *testing.B) {
	first := *testtheme.Behavior()
	second := *testtheme.Behavior()

	b.Run("input", func(b *testing.B) {
		state := &inputState{}
		value := New("Ada").Key("name")
		state.visualDeclarations(value, &first, nil)
		baseline := state.styleCache.Stats()
		b.ReportAllocs()
		b.ResetTimer()
		for index := 0; b.Loop(); index++ {
			switch index & 3 {
			case 0:
				state.visualDeclarations(value, &first, nil)
			case 1:
				state.visualDeclarations(value.Variant(Secondary), &first, nil)
			case 2:
				state.visualDeclarations(value.Variant(Secondary), &second, nil)
			default:
				state.visualDeclarations(value, &second, nil)
			}
		}
		b.StopTimer()
		reportInputDeclarationCacheStats(b, "input-declaration-cache", state.styleCache.Stats(), baseline)
	})

	b.Run("textarea", func(b *testing.B) {
		state := &textAreaState{}
		value := NewTextArea("Ada").Key("notes").Rows(4)
		state.visualDeclarations(value, &first, unit.Dp(80), nil)
		baseline := state.styleCache.Stats()
		b.ReportAllocs()
		b.ResetTimer()
		for index := 0; b.Loop(); index++ {
			switch index & 3 {
			case 0:
				state.visualDeclarations(value, &first, unit.Dp(80), nil)
			case 1:
				state.visualDeclarations(value, &first, unit.Dp(81), nil)
			case 2:
				state.visualDeclarations(value, &second, unit.Dp(81), nil)
			default:
				state.visualDeclarations(value, &second, unit.Dp(80), nil)
			}
		}
		b.StopTimer()
		reportInputDeclarationCacheStats(b, "textarea-declaration-cache", state.styleCache.Stats(), baseline)
	})

	b.Run("group", func(b *testing.B) {
		state := &inputGroupState{}
		value := Group(New("Ada").Key("query"))
		state.visualDeclarations(value, &first, unit.Dp(36), nil)
		baseline := state.styleCache.Stats()
		b.ReportAllocs()
		b.ResetTimer()
		for index := 0; b.Loop(); index++ {
			switch index & 3 {
			case 0:
				state.visualDeclarations(value, &first, unit.Dp(36), nil)
			case 1:
				state.visualDeclarations(value, &first, unit.Dp(37), nil)
			case 2:
				state.visualDeclarations(value, &second, unit.Dp(37), nil)
			default:
				state.visualDeclarations(value, &second, unit.Dp(36), nil)
			}
		}
		b.StopTimer()
		reportInputDeclarationCacheStats(b, "input-group-declaration-cache", state.styleCache.Stats(), baseline)
	})
}

func BenchmarkEditorUpdateSteadyState(b *testing.B) {
	var editor widget.Editor
	editor.SingleLine = true
	editor.Submit = true
	editor.SetText("Ada Lovelace")
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Source: router.Source(),
		Ops:    &ops,
		Now:    time.Unix(1, 0),
	}
	for {
		if _, ok := editor.Update(gtx); !ok {
			break
		}
	}
	router.Frame(&ops)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ops.Reset()
		for {
			if _, ok := editor.Update(gtx); !ok {
				break
			}
		}
	}
}

func BenchmarkEditorUpdateEditEvent(b *testing.B) {
	var editor widget.Editor
	editor.SingleLine = true
	editor.MaxLen = 1
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Source: router.Source(),
		Ops:    &ops,
		Now:    time.Unix(1, 0),
	}
	for {
		if _, ok := editor.Update(gtx); !ok {
			break
		}
	}
	router.Frame(&ops)
	router.Source().Execute(key.FocusCmd{Tag: &editor})
	for {
		if _, ok := editor.Update(gtx); !ok {
			break
		}
	}
	router.Frame(&ops)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ops.Reset()
		router.Queue(key.EditEvent{Text: "x"})
		for {
			if _, ok := editor.Update(gtx); !ok {
				break
			}
		}
		editor.SetText("")
	}
}
