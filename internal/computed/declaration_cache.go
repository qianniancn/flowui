package computed

import "github.com/qianniancn/flowui/theme"

// DeclarationCache memoizes a component-owned style declaration for one
// widget state slot. Rule tables are immutable after construction, so retaining
// the compiled value avoids rebuilding them on every frame while the theme and
// declaration inputs stay unchanged.
//
// The theme pointer is deliberately part of the key. ReplaceTheme installs a
// new theme instance, and declarations must never outlive that instance.
type DeclarationCache[K comparable, V any] struct {
	theme *theme.Theme
	key   K
	value V
	valid bool
}

func (c *DeclarationCache[K, V]) Load(activeTheme *theme.Theme, key K) (V, bool) {
	if c.valid && c.theme == activeTheme && c.key == key {
		return c.value, true
	}
	var zero V
	return zero, false
}

func (c *DeclarationCache[K, V]) Store(activeTheme *theme.Theme, key K, value V) {
	c.theme = activeTheme
	c.key = key
	c.value = value
	c.valid = true
}
