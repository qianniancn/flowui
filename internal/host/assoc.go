package host

import "github.com/qianniancn/flowui/internal/frame"

// AssociationPreparer walks composite children to register field semantics.
// layout package registers the full walker at init to avoid import cycles.
type AssociationPreparer func(ctx *frame.Context, widgets ...frame.Widget)

var associationPreparer AssociationPreparer

// SetAssociationPreparer installs the field-association walker (called from layout).
func SetAssociationPreparer(fn AssociationPreparer) {
	associationPreparer = fn
}

func runAssociationPreparer(ctx *frame.Context, widgets ...frame.Widget) {
	if associationPreparer != nil {
		associationPreparer(ctx, widgets...)
	}
}
