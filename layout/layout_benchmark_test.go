package layout

import (
	"image"
	"runtime"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
)

func BenchmarkFlexLayout(b *testing.B) {
	children := make([]core.Widget, 12)
	for index := range children {
		children[index] = Spacer(20+index, 16)
	}
	widget := Row(children...).Gap(4)
	benchmarkLayoutWidget(b, widget)
}

func BenchmarkWrapLayout(b *testing.B) {
	children := make([]core.Widget, 12)
	for index := range children {
		children[index] = Spacer(48+index%3, 16+index%2)
	}
	widget := Wrap(children...).Gap(4)
	benchmarkLayoutWidget(b, widget)
}

func BenchmarkPrepareFieldAssociations(b *testing.B) {
	ctx := newContext(nil)
	widget := Box(Box(Box(&enabledProbeWidget{})))
	b.ReportAllocs()
	for b.Loop() {
		prepareFieldAssociations(ctx, widget)
	}
}

func benchmarkLayoutWidget(b *testing.B, widget core.Widget) {
	ctx := newContext(nil)
	viewport := image.Pt(320, 200)
	b.Helper()
	b.ReportAllocs()
	for b.Loop() {
		var ops op.Ops
		gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: &ops}
		core.BeginFrameWithViewport(ctx, viewport)
		dims := widget.Layout(ctx, gtx)
		core.EndFrame(ctx)
		runtime.KeepAlive(dims)
	}
}
