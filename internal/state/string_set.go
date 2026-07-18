package state

import "slices"

const stringSetThreshold = 64

type StringSet map[string]struct{}

type StringSetCache struct {
	keys []string
	set  StringSet
}

// Resolve returns a set for large controlled key lists. Short lists remain
// linear to avoid a map and snapshot that cost more than their lookups.
func (c *StringSetCache) Resolve(keys []string) StringSet {
	if len(keys) < stringSetThreshold {
		c.keys = nil
		c.set = nil
		return nil
	}
	if slices.Equal(c.keys, keys) {
		return c.set
	}
	c.keys = append(c.keys[:0], keys...)
	if c.set == nil {
		c.set = make(StringSet, len(keys))
	} else {
		clear(c.set)
	}
	for _, key := range keys {
		if key != "" {
			c.set[key] = struct{}{}
		}
	}
	return c.set
}

func StringSetContains(keys []string, set StringSet, key string) bool {
	if set != nil {
		_, ok := set[key]
		return ok
	}
	return slices.Contains(keys, key)
}
