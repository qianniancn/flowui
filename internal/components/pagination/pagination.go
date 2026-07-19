package pagination

import (
	"sort"

	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type Size uint8

const (
	SizeMedium Size = iota
	SizeSmall
	SizeLarge
)

type Widget struct {
	theme         func(*theme.Theme)
	key           string
	page          int
	total         int
	size          Size
	summary       frame.Widget
	siblings      int
	boundaries    int
	showControls  bool
	disabled      bool
	previousLabel string
	nextLabel     string
	onChange      func(int)
}

func New(key string, page, total int) Widget {
	total = max(total, 1)
	return Widget{
		key:           key,
		page:          min(max(page, 1), total),
		total:         total,
		siblings:      1,
		boundaries:    1,
		showControls:  true,
		previousLabel: "Prev",
		nextLabel:     "Next",
	}
}

func (p Widget) Size(size Size) Widget {
	p.size = size
	return p
}

func (p Widget) Summary(summary frame.Widget) Widget {
	p.summary = summary
	return p
}

func (p Widget) Siblings(count int) Widget {
	p.siblings = max(count, 0)
	return p
}

func (p Widget) Boundaries(count int) Widget {
	p.boundaries = max(count, 0)
	return p
}

func (p Widget) ShowControls(show bool) Widget {
	p.showControls = show
	return p
}

func (p Widget) Labels(previous, next string) Widget {
	p.previousLabel = previous
	p.nextLabel = next
	return p
}

func (p Widget) Disabled(disabled bool) Widget {
	p.disabled = disabled
	return p
}

func (p Widget) OnChange(fn func(int)) Widget {
	p.onChange = fn
	return p
}

// pageItems returns page numbers with zero representing an ellipsis.
func pageItems(page, total, boundaries, siblings int) []int {
	if total <= 0 {
		return nil
	}
	pages := make(map[int]struct{})
	addRange := func(start, end int) {
		start, end = max(start, 1), min(end, total)
		for value := start; value <= end; value++ {
			pages[value] = struct{}{}
		}
	}
	addRange(1, boundaries)
	addRange(page-siblings, page+siblings)
	addRange(total-boundaries+1, total)
	pages[1] = struct{}{}
	pages[total] = struct{}{}

	ordered := make([]int, 0, len(pages))
	for value := range pages {
		ordered = append(ordered, value)
	}
	sort.Ints(ordered)
	items := make([]int, 0, len(ordered)+2)
	for index, value := range ordered {
		if index > 0 {
			gap := value - ordered[index-1]
			if gap == 2 {
				items = append(items, value-1)
			} else if gap > 2 {
				items = append(items, 0)
			}
		}
		items = append(items, value)
	}
	return items
}
