package style

import (
	"math/bits"
	"sort"
	"time"
)

// StateMask is a bitmask of interaction states.
//
// StateMask covers the state families that every control shares. It is plain
// data rather than predicates, which buys three properties a predicate tree
// cannot have:
//
//   - precedence is deterministic and documentable (see Rule),
//   - a declaration can be hashed statically, without evaluating anything,
//   - a linter can prove coverage ("every component has a Disabled rule").
//
// The name deliberately avoids "variant": in this library a variant is the
// visual personality a component exposes (button.Variant), while StateMask
// describes the interaction state a component is in. Reach for the unexported
// predicate helpers only when a condition genuinely cannot be reduced to a set
// of state bits.
type StateMask uint16

const (
	Disabled StateMask = 1 << iota
	Hovered
	Pressed
	Focused
	FocusVisible
	Selected
	Checked
	Indeterminate
	ReadOnly
	Invalid
	Loading
	Open
	Expanded
	Dragging
	DropTarget
	Warning
)

// StateMaskAll is the union of every defined state bit.
const StateMaskAll = Disabled |
	Hovered |
	Pressed |
	Focused |
	FocusVisible |
	Selected |
	Checked |
	Indeterminate |
	ReadOnly |
	Invalid |
	Loading |
	Open |
	Expanded |
	Dragging |
	DropTarget |
	Warning

// Bits returns the bits a resolved interaction state contributes.
//
// The mapping is intentionally one-to-one: no bit implies another. Any
// suppression between families (for example "a disabled control is not
// hovered") belongs in the component that produces State, where it can be
// asserted directly, rather than hidden in this conversion.
func (s State) Bits() StateMask {
	var v StateMask
	if s.Disabled {
		v |= Disabled
	}
	if s.Hovered {
		v |= Hovered
	}
	if s.Pressed {
		v |= Pressed
	}
	if s.Focused {
		v |= Focused
	}
	if s.FocusVisible {
		v |= FocusVisible
	}
	if s.Selected {
		v |= Selected
	}
	if s.Checked {
		v |= Checked
	}
	if s.Indeterminate {
		v |= Indeterminate
	}
	if s.ReadOnly {
		v |= ReadOnly
	}
	if s.Invalid {
		v |= Invalid
	}
	if s.Loading {
		v |= Loading
	}
	if s.Open {
		v |= Open
	}
	if s.Expanded {
		v |= Expanded
	}
	if s.Dragging {
		v |= Dragging
	}
	if s.DropTarget {
		v |= DropTarget
	}
	if s.Warning {
		v |= Warning
	}
	return v
}

// Split decomposes a state into independent families. It exists for tests and
// for components that need to reason about one family at a time.
func (v StateMask) Split() State {
	return State{
		Disabled:      v&Disabled != 0,
		Hovered:       v&Hovered != 0,
		Pressed:       v&Pressed != 0,
		Focused:       v&Focused != 0,
		FocusVisible:  v&FocusVisible != 0,
		Selected:      v&Selected != 0,
		Checked:       v&Checked != 0,
		Indeterminate: v&Indeterminate != 0,
		ReadOnly:      v&ReadOnly != 0,
		Invalid:       v&Invalid != 0,
		Loading:       v&Loading != 0,
		Open:          v&Open != 0,
		Expanded:      v&Expanded != 0,
		Dragging:      v&Dragging != 0,
		DropTarget:    v&DropTarget != 0,
		Warning:       v&Warning != 0,
	}
}

// Tween declares the animation applied when a rule's Set takes effect.
//
// Tween deliberately carries no EaseFunc. An easing function is a closure and
// would make the rule table unhashable, which is the property the table exists
// to provide. Named easings belong in a resolver-side registry keyed by an
// enum, so that a rule stays data.
type Tween struct {
	Property PropertyID
	Duration time.Duration
	Delay    time.Duration
}

// Rule is one row of a state rule table.
//
// A rule only writes the properties it mentions; everything else keeps the
// value contributed by a lower-precedence rule or by the base declaration.
type Rule struct {
	// Rank groups rules into precedence levels. A rule with a higher Rank
	// always overrides one with a lower Rank, no matter how many bits matched.
	//
	// Rank exists because bit count alone cannot order orthogonal families.
	// "Disabled always beats Invalid and Hovered" is a fixed precedence, but
	// expressing it through specificity would need one rule per combination of
	// the losing family's bits — 2^n rules for n bits. Leave Rank at zero
	// unless two families are genuinely orthogonal yet strictly ordered.
	Rank int8

	// When requires every listed bit to be set.
	When StateMask

	// Not requires every listed bit to be clear.
	Not StateMask

	// Set is the increment this rule contributes.
	Set Style

	// Tween optionally animates the properties Set changes.
	Tween *Tween
}

// Matches reports whether the rule applies to the given state bits.
func (r Rule) Matches(v StateMask) bool {
	return v&r.When == r.When && v&r.Not == 0
}

// Rules appends state rules to this declaration.
//
// Rules and When may coexist on one declaration. Rules are applied first and
// When second, so an explicit predicate always wins over a table row. That
// ordering makes When the documented escape hatch rather than a second, silent
// source of precedence.
func (s Style) Rules(rules ...Rule) Style {
	if len(rules) == 0 {
		return s
	}
	kept := make([]Rule, len(s.rules), len(s.rules)+len(rules))
	copy(kept, s.rules)
	kept = append(kept, rules...)
	s.rules = kept
	return s
}

// RuleCount reports how many table rows this declaration carries.
func (s Style) RuleCount() int {
	return len(s.rules)
}

// RuleTable returns the declaration's state rows in declaration order.
//
// The table is an authoring surface rather than a runtime one, but it must be
// readable: "does every component declare a Disabled row?" and "are two rows
// unreachable?" are questions lint tooling has to answer, and they cannot be
// answered by walking closures.
//
// The returned slice aliases the declaration's storage. Treat it as read-only;
// the declaration itself is immutable, so a returned row cannot be edited in
// place in a way that affects any other value.
func (s Style) RuleTable() []Rule {
	return s.rules
}

// On appends one state rule row to this declaration.
//
// It is the authoring form of Rule{When: mask, Set: set}: the two constructs
// are interchangeable, and the equivalence is pinned by test. mask may be
// zero, which yields an unconditional row -- it matches every state and sorts
// first.
func (s Style) On(mask StateMask, set Style, tween ...*Tween) Style {
	return s.Rules(Rule{When: mask, Set: set, Tween: firstTween(tween)})
}

// OnAny appends one rule row per bit of mask, ordered from the least
// significant bit up. It is the authoring form of a disjunction: a row per
// bit, because a rule row's When is a mask-conjunction and cannot express OR.
//
// A tween, if given, attaches to the first expanded row (the lowest bit's)
// only. The rows are independent increments of the same set; the first row is
// the declaration's head, and attaching the tween to every row would make the
// per-row semantics depend on how many bits happened to match. An empty mask
// appends nothing.
func (s Style) OnAny(mask StateMask, set Style, tween ...*Tween) Style {
	if mask == 0 {
		return s
	}
	rules := make([]Rule, 0, bits.OnesCount16(uint16(mask)))
	for rest := mask; rest != 0; rest &= rest - 1 {
		rule := Rule{When: rest & -rest, Set: set}
		if len(rules) == 0 {
			rule.Tween = firstTween(tween)
		}
		rules = append(rules, rule)
	}
	return s.Rules(rules...)
}

// OnNot appends one state rule row that requires every bit of when to be
// set and every bit of not to be clear. It is the authoring form of
// Rule{When: when, Not: not, Set: set}; not may carry several bits at once,
// which reads as "and not that, and not the other".
func (s Style) OnNot(when, not StateMask, set Style, tween ...*Tween) Style {
	return s.Rules(Rule{When: when, Not: not, Set: set, Tween: firstTween(tween)})
}

// Quick builds a row-level tween for the On family.
//
// The property is explicit because a rule row must expand to exactly one
// tween: an immutable declaration cannot be asked which properties its Set
// changes, so an inferred property would make the expansion ambiguous.
// Transitions declared inside set itself (Transition calls on the chain) stay
// base-layer transitions; Quick is for the row that introduces the change.
func Quick(property PropertyID, duration time.Duration) *Tween {
	return &Tween{Property: property, Duration: max(duration, 0)}
}

// firstTween returns the first non-nil tween in the variadic tail, or nil.
func firstTween(tweens []*Tween) *Tween {
	for _, tween := range tweens {
		if tween != nil {
			return tween
		}
	}
	return nil
}

// orderedRules returns the rules matching v in application order: lowest
// precedence first, so that a later merge overwrites an earlier one.
//
// Ordering key, ascending: Rank, then number of matched bits, then declaration
// order. The declaration-order tiebreak preserves the meaning the old
// predicate-chain override had, where a later call overrode an earlier one.
// A rule matching zero bits in When (an unconditional default) therefore
// sorts first.
func orderedRules(rules []Rule, v StateMask) []Rule {
	if len(rules) == 0 {
		return nil
	}
	matched := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		if rule.Matches(v) {
			matched = append(matched, rule)
		}
	}
	if len(matched) < 2 {
		return matched
	}
	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Rank != matched[j].Rank {
			return matched[i].Rank < matched[j].Rank
		}
		return bits.OnesCount16(uint16(matched[i].When)) < bits.OnesCount16(uint16(matched[j].When))
	})
	return matched
}
