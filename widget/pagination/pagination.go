package pagination

import (
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Size uint8

const (
	SizeMedium Size = iota
	SizeSmall
	SizeLarge
)

type Widget struct {
	override         *theme.PaginationOverride
	key              string
	page             int
	total            int
	size             Size
	summary          core.Widget
	siblings         int
	boundaries       int
	showControls     bool
	disabled         bool
	previousLabel    string
	nextLabel        string
	hasPreviousLabel bool
	hasNextLabel     bool
	onChange         func(int)
	visualCache      *paginationVisualCache
}

func New(key string, page, total int) Widget {
	total = max(total, 1)
	return Widget{
		key:          key,
		page:         min(max(page, 1), total),
		total:        total,
		siblings:     1,
		boundaries:   1,
		showControls: true,
		visualCache:  &paginationVisualCache{},
	}
}

func (p Widget) Size(size Size) Widget {
	p.size = size
	return p
}

func (p Widget) Summary(summary core.Widget) Widget {
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
	p.hasPreviousLabel = true
	p.hasNextLabel = true
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
	// The page set is the union of at most five sorted ranges. Merging those
	// ranges directly avoids a map, a sort, and a second page slice on every
	// layout frame.
	type pageRange struct{ start, end int }
	ranges := [5]pageRange{}
	rangeCount := 0
	addRange := func(start, end int) {
		start, end = max(start, 1), min(end, total)
		if start <= end {
			ranges[rangeCount] = pageRange{start: start, end: end}
			rangeCount++
		}
	}
	addRange(1, boundaries)
	addRange(page-siblings, page+siblings)
	addRange(total-boundaries+1, total)
	addRange(1, 1)
	addRange(total, total)
	for index := 1; index < rangeCount; index++ {
		current := ranges[index]
		position := index
		for position > 0 && current.start < ranges[position-1].start {
			ranges[position] = ranges[position-1]
			position--
		}
		ranges[position] = current
	}

	// Keep the common compact range in one allocation without trusting the
	// caller-provided boundary counts for capacity arithmetic.
	capacity := min(total, 7) + 2
	items := make([]int, 0, capacity)
	previous := 0
	for index := 0; index < rangeCount; index++ {
		current := ranges[index]
		if current.start <= previous {
			if current.end <= previous {
				continue
			}
			current.start = previous + 1
		}
		if previous > 0 {
			gap := current.start - previous
			if gap == 2 {
				items = append(items, previous+1)
			} else if gap > 2 {
				items = append(items, 0)
			}
		}
		for value := current.start; value <= current.end; value++ {
			items = append(items, value)
		}
		previous = current.end
	}
	return items
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.PaginationOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
