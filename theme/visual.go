package theme

// ResolveVisual builds and validates one component visual from the active
// provider. Pass a ComponentVisualProvider method expression and its matching
// Validate*Visual function. Like the validators, it panics on an incomplete
// contract; callers may cache the returned visual after successful validation.
func ResolveVisual[Request, Visual any](
	active *Theme,
	request Request,
	build func(ComponentVisualProvider, *Theme, Request) Visual,
	validate func(Visual),
) Visual {
	provider, err := ComponentProvider(active)
	if err != nil {
		panic(err)
	}
	visual := build(provider, active, request)
	validate(visual)
	return visual
}
