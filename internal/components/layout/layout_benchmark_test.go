package layoutui

import (
	"image"
	"runtime"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func BenchmarkFlexLayout(b *testing.B) {
	children := make([]frame.Widget, 12)
	for index := range children {
		children[index] = Spacer(20+index, 16)
	}
	widget := Row(children...).Gap(4)
	benchmarkLayoutWidget(b, widget)
}

func BenchmarkWrapLayout(b *testing.B) {
	children := make([]frame.Widget, 12)
	for index := range children {
		children[index] = Spacer(48+index%3, 16+index%2)
	}
	widget := Wrap(children...).Gap(4)
	benchmarkLayoutWidget(b, widget)
}

func benchmarkLayoutWidget(b *testing.B, widget frame.Widget) {
	ctx := newContext(nil)
	viewport := image.Pt(320, 200)
	b.Helper()
	b.ReportAllocs()
	for b.Loop() {
		var ops op.Ops
		gtx := layout.Context{Constraints: layout.Constraints{Max: viewport}, Ops: &ops}
		frame.BeginFrameWithViewport(ctx, viewport)
		dims := widget.Layout(ctx, gtx)
		frame.EndFrame(ctx)
		runtime.KeepAlive(dims)
	}
}
