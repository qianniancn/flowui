package sidebar

import (
	"fmt"
	"runtime"
	"testing"
)

func BenchmarkSidebarCloneSections(b *testing.B) {
	sections := make([]Section, 8)
	for sectionIndex := range sections {
		sections[sectionIndex].Title = fmt.Sprintf("Section %d", sectionIndex)
		sections[sectionIndex].Items = make([]Item, 8)
		for itemIndex := range sections[sectionIndex].Items {
			sections[sectionIndex].Items[itemIndex] = Item{
				Key:   fmt.Sprintf("item-%d-%d", sectionIndex, itemIndex),
				Label: "Item",
			}
		}
	}

	b.ReportAllocs()
	for b.Loop() {
		cloned := cloneSections(sections)
		runtime.KeepAlive(cloned)
	}
}
