package menu

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

type menuBenchmarkFrame struct {
	ctx    *core.Context
	router input.Router
	ops    op.Ops
	gtx    layout.Context
	value  ContextMenuWidget
}

func newMenuBenchmarkFrame() *menuBenchmarkFrame {
	active := testtheme.Behavior()
	items := []Item{
		{Key: "new", Label: "New document", Shortcut: "Ctrl+N"},
		{Key: "open", Label: "Open document", Shortcut: "Ctrl+O"},
		{Key: "save", Label: "Save document", Shortcut: "Ctrl+S"},
		{Key: "share", Label: "Share", Children: []Item{
			{Key: "copy-link", Label: "Copy link"},
			{Key: "invite", Label: "Invite collaborators"},
		}},
		Separator(),
		{Key: "archive", Label: "Archive", Variant: ItemDanger},
		{Key: "delete", Label: "Delete", Variant: ItemDanger, Disabled: true},
	}
	menuValue := New("benchmark-menu", items).
		BeforeContent(fixedMenuWidget{size: image.Pt(12, 8)}).
		AfterContent(fixedMenuWidget{size: image.Pt(12, 8)}).
		DataVersion(1)
	return &menuBenchmarkFrame{
		ctx: core.New(nil, active, locale.LanguageEnglish),
		gtx: layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(640, 480)},
			Now:         time.Unix(1, 0),
		},
		value: ContextMenu("benchmark-context", fixedMenuWidget{size: image.Pt(180, 36)}, menuValue).DefaultOpen(true),
	}
}

func (frame *menuBenchmarkFrame) layout() {
	frame.ops.Reset()
	frame.gtx.Source = frame.router.Source()
	frame.gtx.Ops = &frame.ops
	core.BeginFrameWithViewport(frame.ctx, frame.gtx.Constraints.Max)
	frame.value.Layout(frame.ctx, frame.gtx)
	core.LayoutOverlays(frame.ctx, frame.gtx)
	core.ApplyFrameCommands(frame.ctx, frame.gtx)
	core.EndFrame(frame.ctx)
	frame.router.Frame(&frame.ops)
}

func (frame *menuBenchmarkFrame) visualCacheStats() MenuVisualCacheStats {
	state, ok := core.PeekState[contextMenuState](frame.ctx, "benchmark-context", stateSlotContextMenu)
	if !ok || state == nil {
		return MenuVisualCacheStats{}
	}
	return state.visualCache.stats()
}

// BenchmarkMenuComponentTreeSteadyState covers trigger, popup panel, item
// content, submenu-capable rows, separators, and context-menu overlay input.
func BenchmarkMenuComponentTreeSteadyState(b *testing.B) {
	frame := newMenuBenchmarkFrame()
	frame.layout()
	frame.gtx.Now = frame.gtx.Now.Add(contextMenuEnterDuration)
	frame.layout()
	baseline := frame.visualCacheStats()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		frame.gtx.Now = frame.gtx.Now.Add(time.Millisecond)
		frame.layout()
	}
	b.StopTimer()
	stats := frame.visualCacheStats()
	if b.N > 0 {
		b.ReportMetric(float64(stats.Hits-baseline.Hits)/float64(b.N), "menu-visual-cache-hits/op")
		b.ReportMetric(float64(stats.Misses-baseline.Misses)/float64(b.N), "menu-visual-cache-misses/op")
		b.ReportMetric(float64(stats.ThemeInvalidations-baseline.ThemeInvalidations)/float64(b.N), "menu-visual-cache-theme-invalidations/op")
		b.ReportMetric(float64(stats.KeyInvalidations-baseline.KeyInvalidations)/float64(b.N), "menu-visual-cache-key-invalidations/op")
	}
}

// BenchmarkMenuAutoWidthMeasurement covers the text, slot, and item-style
// resolution path used by AutoWidth without the surrounding overlay tree.
func BenchmarkMenuAutoWidthMeasurement(b *testing.B) {
	frame := newMenuBenchmarkFrame()
	menu := frame.value.menu.AutoWidth()
	visual := menu.visual(frame.ctx)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = menu.preferredWidthPxWithVisual(frame.ctx, frame.gtx, visual)
	}
}
