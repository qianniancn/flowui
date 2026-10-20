package uitest

import (
	"image"

	"gioui.org/io/input"
	"gioui.org/io/semantic"
	"github.com/qianniancn/flowui/ui"
)

// Semantic is one node of the Gio semantic tree produced by a frame. The tree
// is the platform-facing accessibility output: it is what an assistive
// technology reads, so assertions against it are the only evidence that a
// component reports the role, label, description, selected, and disabled state
// it claims to.
type Semantic = input.SemanticNode

// Semantics returns the top-level semantic nodes produced by the most recent
// frame.
//
// Walk the returned slice directly and recurse with WalkSemantics. Gio's raw
// AppendSemantics result is a flat list whose element 0 is a synthetic root
// whose Children point back into that same list, so iterating the flat list and
// recursing visits nodes more than once. This helper starts below that root.
//
// Call it after Frame; a later Frame replaces the result.
func (h *Harness) Semantics() []Semantic {
	if h == nil {
		return nil
	}
	tree := h.Router().AppendSemantics(nil)
	if len(tree) == 0 {
		return nil
	}
	return tree[0].Children
}

// Relations returns FlowUI's additive semantic relationship registry for the
// most recent frame.
//
// Gio's tree does not model every relationship FlowUI exposes, so components
// register the remainder here: currently Tab to its list and Tab to its panel.
// Prefer Semantics for role, label, description, and state assertions; use
// Relations only for relationships the Gio tree cannot carry.
func (h *Harness) Relations() []ui.SemanticNode {
	if h == nil {
		return nil
	}
	return ui.Semantics(h.Context())
}

// WalkSemantics visits every node in depth-first tree order. Returning false
// from visit stops the walk, including recursion into the current node.
func WalkSemantics(nodes []Semantic, visit func(Semantic) bool) {
	if visit == nil {
		return
	}
	var walk func([]Semantic) bool
	walk = func(list []Semantic) bool {
		for _, node := range list {
			if !visit(node) {
				return false
			}
			if !walk(node.Children) {
				return false
			}
		}
		return true
	}
	walk(nodes)
}

// FindSemantic returns the first node in tree order that satisfies match.
func FindSemantic(nodes []Semantic, match func(Semantic) bool) (Semantic, bool) {
	if match == nil {
		return Semantic{}, false
	}
	var found Semantic
	var ok bool
	WalkSemantics(nodes, func(node Semantic) bool {
		if match(node) {
			found, ok = node, true
			return false
		}
		return true
	})
	return found, ok
}

// SemanticByLabel returns the first node whose label equals label.
func SemanticByLabel(nodes []Semantic, label string) (Semantic, bool) {
	return FindSemantic(nodes, func(node Semantic) bool {
		return node.Desc.Label == label
	})
}

// SemanticByDescription returns the first node whose description equals
// description.
func SemanticByDescription(nodes []Semantic, description string) (Semantic, bool) {
	return FindSemantic(nodes, func(node Semantic) bool {
		return node.Desc.Description == description
	})
}

// HasSemanticLabel reports whether any node carries label.
func HasSemanticLabel(nodes []Semantic, label string) bool {
	_, ok := SemanticByLabel(nodes, label)
	return ok
}

// HasSemanticDescription reports whether any node carries description.
func HasSemanticDescription(nodes []Semantic, description string) bool {
	_, ok := SemanticByDescription(nodes, description)
	return ok
}

// SemanticLabels returns every non-empty label in tree order, including
// duplicates, so a test can assert on ordering or on a repeated label.
func SemanticLabels(nodes []Semantic) []string {
	var labels []string
	WalkSemantics(nodes, func(node Semantic) bool {
		if node.Desc.Label != "" {
			labels = append(labels, node.Desc.Label)
		}
		return true
	})
	return labels
}

// SemanticDescriptions returns every non-empty description in tree order.
func SemanticDescriptions(nodes []Semantic) []string {
	var descriptions []string
	WalkSemantics(nodes, func(node Semantic) bool {
		if node.Desc.Description != "" {
			descriptions = append(descriptions, node.Desc.Description)
		}
		return true
	})
	return descriptions
}

// SelectedSemanticLabels returns the labels of selected nodes in tree order.
func SelectedSemanticLabels(nodes []Semantic) []string {
	var labels []string
	WalkSemantics(nodes, func(node Semantic) bool {
		if node.Desc.Selected && node.Desc.Label != "" {
			labels = append(labels, node.Desc.Label)
		}
		return true
	})
	return labels
}

// DisabledSemanticLabels returns the labels of disabled nodes in tree order.
func DisabledSemanticLabels(nodes []Semantic) []string {
	var labels []string
	WalkSemantics(nodes, func(node Semantic) bool {
		if node.Desc.Disabled && node.Desc.Label != "" {
			labels = append(labels, node.Desc.Label)
		}
		return true
	})
	return labels
}

// SemanticBoundsForLabel returns the reported bounds of the first node whose
// label equals label. Tests use it to assert hit-target geometry instead of
// duplicating layout arithmetic.
func SemanticBoundsForLabel(nodes []Semantic, label string) (image.Rectangle, bool) {
	node, ok := SemanticByLabel(nodes, label)
	if !ok {
		return image.Rectangle{}, false
	}
	return node.Desc.Bounds, true
}

// SemanticClassForLabel returns the semantic class reported for the first node
// whose label equals label.
func SemanticClassForLabel(nodes []Semantic, label string) (semantic.ClassOp, bool) {
	node, ok := SemanticByLabel(nodes, label)
	if !ok {
		return semantic.ClassOp(0), false
	}
	return node.Desc.Class, true
}
