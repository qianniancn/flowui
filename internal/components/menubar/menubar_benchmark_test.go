package menubar

import (
	"testing"
	"time"

	"gioui.org/io/input"
)

func BenchmarkMenubarLayout(b *testing.B) {
	ctx := menubarTestContext()
	router := new(input.Router)
	widget := menubarTestWidget()
	now := time.Unix(1, 0)
	layoutMenubarFrame(ctx, router, widget, now)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		layoutMenubarFrame(ctx, router, widget, now)
	}
}
