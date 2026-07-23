package style

import (
	"testing"
	"time"
)

var benchmarkStyle Style

func BenchmarkStyleFluent(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkStyle = Style{}.
			Width(190).
			Height(36).
			PaddingX(12).
			Radius(8).
			Background(TokenAccent).
			TextColor(TokenAccentForeground).
			Cursor(0).
			Transition(PropBackgroundColor, 150*time.Millisecond)
	}
}

func BenchmarkStyleFluentRules(b *testing.B) {
	hovered := Style{}.Background(TokenAccentHover)
	label := Style{}.FontWeight(600)
	b.ReportAllocs()
	for b.Loop() {
		benchmarkStyle = Style{}.Background(TokenAccent).
			When(Hovered, hovered).
			Part(PartLabel, label)
	}
}
