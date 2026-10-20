package theme

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ErrMissingVisualProvider reports a theme that cannot supply component
// visuals. Applications must select a concrete style provider instead of
// relying on an implicit framework appearance.
var ErrMissingVisualProvider = errors.New("flowui: theme has no visual provider")

// ErrIncompleteVisualProvider reports a provider that implements only part of
// the component visual contract. A running window needs one coherent provider
// so each migrated component cannot silently fall back to a different recipe.
var ErrIncompleteVisualProvider = errors.New("flowui: theme visual provider is incomplete")

// ComponentVisualProvider is the complete visual protocol currently required
// by the migrated component family. It composes component-specific methods so
// a theme owns one provider without forcing widgets to know concrete themes.
type ComponentVisualProvider interface {
	VisualProvider
	CloseButtonProvider
	ChipProvider
	AvatarProvider
	BadgeProvider
	DescriptionProvider
	LabelProvider
	TextProvider
	IconProvider
	SurfaceProvider
	CardProvider
	StatusBarProvider
	TitleBarProvider
	ToolbarProvider
	InputProvider
	DatePickerProvider
	InputActionProvider
	SelectProvider
	ComboProvider
	ListBoxProvider
	SidebarProvider
	SwitchProvider
	ToggleProvider
	CheckboxProvider
	RadioGroupProvider
	SliderProvider
	MenuProvider
	DropdownProvider
	PopoverProvider
	TooltipProvider
	ToastProvider
	TabsProvider
	PaginationProvider
	ModalProvider
	AlertProvider
	AlertDialogProvider
	TableProvider
	TreeProvider
	ColorPickerProvider
	HeatmapProvider
	LineChartProvider
	PieChartProvider
	CandlestickProvider
	BarChartProvider
	GanttProvider
	TimelineProvider
	NodeGraphProvider
	MenubarProvider
	ProgressProvider
	SpinnerProvider
	CollapsibleProvider
	ImageProvider
	CodeProvider
	SkeletonProvider
	SegmentedProvider
}

// ValidateVisualProvider checks the complete provider contract required by a
// running window for all components currently in the migration boundary.
func ValidateVisualProvider(active Theme) error {
	if active.VisualProvider == nil {
		return ErrMissingVisualProvider
	}
	if _, ok := active.VisualProvider.(ComponentVisualProvider); !ok {
		return incompleteVisualProviderError(active.VisualProvider)
	}
	return nil
}

// ComponentProvider returns the complete visual provider for an active theme.
// Keeping the assertion here gives callers one diagnostic instead of each
// widget independently deciding which optional capability to accept.
func ComponentProvider(active *Theme) (ComponentVisualProvider, error) {
	if active == nil || active.VisualProvider == nil {
		return nil, ErrMissingVisualProvider
	}
	provider, ok := active.VisualProvider.(ComponentVisualProvider)
	if !ok {
		return nil, incompleteVisualProviderError(active.VisualProvider)
	}
	return provider, nil
}

// VisualValidationError identifies the first failed component contract. Field
// is a visual field or resolved declaration path. It unwraps to
// ErrIncompleteVisualProvider so callers can keep using errors.Is.
type VisualValidationError struct {
	Component string
	Field     string
	Reason    string
}

func (err *VisualValidationError) Error() string {
	return fmt.Sprintf("%v: %s.%s: %s", ErrIncompleteVisualProvider, err.Component, err.Field, err.Reason)
}

func (err *VisualValidationError) Unwrap() error { return ErrIncompleteVisualProvider }

func invalidVisual(component, field, reason string) {
	panic(&VisualValidationError{Component: component, Field: field, Reason: reason})
}

func incompleteVisualProviderError(provider VisualProvider) error {
	// Reflection is confined to this failure path. Successful lookups need only
	// the complete interface assertion above.
	contract := reflect.TypeFor[ComponentVisualProvider]()
	actual := reflect.TypeOf(provider)
	var missing []string
	for expected := range contract.Methods() {
		expected := expected
		method, ok := actual.MethodByName(expected.Name)
		if !ok || !matchesProviderMethod(method.Type, expected.Type) {
			missing = append(missing, expected.Name)
		}
	}
	return &VisualValidationError{
		Component: "Provider", Field: strings.Join(missing, ", "),
		Reason: "missing or incompatible methods",
	}
}

func matchesProviderMethod(actual, expected reflect.Type) bool {
	if actual.NumIn() != expected.NumIn()+1 || actual.NumOut() != expected.NumOut() || actual.IsVariadic() != expected.IsVariadic() {
		return false
	}
	for i := 0; i < expected.NumIn(); i++ {
		if actual.In(i+1) != expected.In(i) {
			return false
		}
	}
	for i := 0; i < expected.NumOut(); i++ {
		if actual.Out(i) != expected.Out(i) {
			return false
		}
	}
	return true
}

// VisualProvider supplies component visuals without owning component behavior.
// The active Theme is passed on every call so copied themes and runtime theme
// replacement cannot leave a provider holding a stale theme pointer.
type VisualProvider interface {
	Button(active *Theme, request ButtonRequest) ButtonVisual
}

// ImmutableVisualProvider is an optional marker for providers whose visual
// resolver results are deterministic for a theme pointer and request. It
// allows component runtimes to share validation facts across instances. A
// provider whose resolver closures can change in place must omit this marker;
// runtimes will validate each returned visual instead of reusing a shared fact.
type ImmutableVisualProvider interface {
	VisualsAreImmutable()
}
