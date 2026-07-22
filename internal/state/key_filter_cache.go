package state

import (
	"slices"

	"gioui.org/io/event"
	"gioui.org/io/key"
)

// KeyFilterCache reuses boxed key filters for a stable event target.
type KeyFilterCache struct {
	names   []key.Name
	filters []event.Filter
}

func (c *KeyFilterCache) Resolve(target event.Tag, names ...key.Name) []event.Filter {
	if slices.Equal(c.names, names) {
		return c.filters
	}
	c.names = append(c.names[:0], names...)
	if cap(c.filters) < len(names) {
		c.filters = make([]event.Filter, len(names))
	} else {
		c.filters = c.filters[:len(names)]
	}
	for index, name := range names {
		c.filters[index] = key.Filter{Focus: target, Name: name}
	}
	return c.filters
}
