package navigation

import (
	"strconv"
	"testing"
	"time"
)

func BenchmarkNavigationMove(b *testing.B) {
	items := navigationBenchmarkList()
	current := 0
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		current, _ = Move(items, current, 1, true)
	}
	b.StopTimer()
	if current < 0 {
		b.Fatal("benchmark navigation index became invalid")
	}
}

func BenchmarkNavigationMatch(b *testing.B) {
	items := navigationBenchmarkList()
	current := 0
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		current, _ = Match(items, current, "item-15")
	}
	b.StopTimer()
	if current < 0 {
		b.Fatal("benchmark match index became invalid")
	}
}

func BenchmarkNavigationTypeahead(b *testing.B) {
	var typeahead Typeahead
	base := time.Unix(0, 0)
	index := 0
	chars := []string{"a", "b", "c", "d"}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if index%len(chars) == 0 {
			typeahead.Reset()
		}
		index++
		typeahead.Append(base.Add(time.Duration(index)*10*time.Millisecond), chars[index%len(chars)])
	}
	b.StopTimer()
	if typeahead.text == "" {
		b.Fatal("benchmark typeahead query is empty")
	}
}

func navigationBenchmarkList() List {
	labels := make([]string, 32)
	for index := range labels {
		labels[index] = "item-" + strconv.Itoa(index)
	}
	return List{
		Count: 32,
		Disabled: func(index int) bool {
			return index%7 == 0
		},
		Label: func(index int) string {
			return labels[index]
		},
	}
}
