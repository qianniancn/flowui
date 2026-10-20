package core

import (
	"slices"

	"gioui.org/io/event"
	"gioui.org/io/key"
)

// KeyFilterCache reuses boxed key filters for a stable event target.
type KeyFilterCache struct {
	target   event.Tag
	optional key.Modifiers
	names    []key.Name
	filters  []event.Filter
}

func (c *KeyFilterCache) Resolve(target event.Tag, names ...key.Name) []event.Filter {
	return c.resolve(target, 0, names...)
}

// ResolveOptional reuses filters that accept the supplied optional modifiers.
// It keeps modifier-aware keyboard paths allocation-free after their first frame.
func (c *KeyFilterCache) ResolveOptional(target event.Tag, optional key.Modifiers, names ...key.Name) []event.Filter {
	return c.resolve(target, optional, names...)
}

func (c *KeyFilterCache) resolve(target event.Tag, optional key.Modifiers, names ...key.Name) []event.Filter {
	if c.target == target && c.optional == optional && slices.Equal(c.names, names) {
		return c.filters
	}
	c.target = target
	c.optional = optional
	c.names = append(c.names[:0], names...)
	if cap(c.filters) < len(names) {
		c.filters = make([]event.Filter, len(names))
	} else {
		c.filters = c.filters[:len(names)]
	}
	for index, name := range names {
		c.filters[index] = key.Filter{Focus: target, Name: name, Optional: optional}
	}
	return c.filters
}
