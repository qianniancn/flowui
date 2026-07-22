package layoutui

import "testing"

func BenchmarkPrepareFieldAssociations(b *testing.B) {
	ctx := newContext(nil)
	widget := Box(Box(Box(&enabledProbeWidget{})))
	b.ReportAllocs()
	for b.Loop() {
		prepareFieldAssociations(ctx, widget)
	}
}
