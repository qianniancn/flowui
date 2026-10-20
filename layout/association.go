package layout

// AssociationPreparer walks composite children to register field semantics.
type AssociationPreparer func(ctx *Context, widgets ...Widget)

var associationPreparer AssociationPreparer

func setAssociationPreparer(fn AssociationPreparer) {
	associationPreparer = fn
}

func runAssociationPreparer(ctx *Context, widgets ...Widget) {
	if associationPreparer != nil {
		associationPreparer(ctx, widgets...)
	}
}

// runSingleAssociationPreparer avoids creating a variadic argument slice for
// the common one-child layout path. The installed preparer is the field
// association walker below; associationPreparer remains the general hook for
// multi-child callers.
func runSingleAssociationPreparer(ctx *Context, widget Widget) {
	if associationPreparer != nil {
		prepareFieldAssociation(ctx, widget)
	}
}

func init() {
	setAssociationPreparer(prepareFieldAssociations)
}
