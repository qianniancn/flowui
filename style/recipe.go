package style

import "slices"

// PartRecipe is one part's entry in a Recipe: the unconditional base plus the
// variant rows that specialize it.
type PartRecipe struct {
	// Base is the part's unconditional declaration.
	Base Style

	// Rules are the part's variant table rows. Order matters only as the
	// tiebreak between rows of equal Rank and equal specificity.
	Rules []Rule
}

// Recipe is a component style authored as "parts, each owning a base
// declaration and a variant table".
//
// Recipe is a compile-time shape rather than a runtime concept: Compile lowers
// it to the same Style the fluent API builds, so the resolver, cache, and
// renderer never observe it. It exists so the authoring surface can name parts
// as keys instead of hiding them inside a chain, which is what makes "which
// part does this declaration belong to" answerable by reading the code.
//
// A map is the right container because parts are a set, not a sequence: merging
// PartRoot and PartText are independent operations, so map iteration order
// cannot change the result. Compile still walks the keys in sorted order so
// that the determinism is a property of the code rather than an argument the
// reader has to reconstruct.
type Recipe map[Part]PartRecipe

// Compile lowers the recipe into a single immutable declaration.
func (r Recipe) Compile() Style {
	if len(r) == 0 {
		return Style{}
	}
	parts := make([]Part, 0, len(r))
	for part := range r {
		parts = append(parts, part)
	}
	slices.Sort(parts)

	var result Style
	for _, part := range parts {
		entry := r[part]
		result = result.Part(part, entry.Base.Rules(entry.Rules...))
	}
	return result
}
