package theme_test

import (
	"errors"
	"testing"

	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

func TestVisualValidatorsRejectIncompleteDeclarations(t *testing.T) {
	tests := []struct {
		name  string
		check func()
	}{
		{name: "alert", check: func() { theme.ValidateAlertVisual(theme.AlertVisual{}) }},
		{name: "alert dialog", check: func() { theme.ValidateAlertDialogVisual(theme.AlertDialogVisual{}) }},
		{name: "button", check: func() { theme.ValidateButtonVisual(theme.ButtonVisual{}) }},
		{name: "checkbox", check: func() { theme.ValidateCheckboxVisual(theme.CheckboxVisual{}) }},
		{name: "collapsible", check: func() { theme.ValidateCollapsibleVisual(theme.CollapsibleVisual{}) }},
		{name: "date picker", check: func() { theme.ValidateDatePickerVisual(theme.DatePickerVisual{}) }},
		{name: "list box", check: func() { theme.ValidateListBoxVisual(theme.ListBoxVisual{}) }},
		{name: "menu", check: func() { theme.ValidateMenuVisual(theme.MenuVisual{}) }},
		{name: "menubar", check: func() { theme.ValidateMenubarVisual(theme.MenubarVisual{}) }},
		{name: "radio group", check: func() { theme.ValidateRadioGroupVisual(theme.RadioGroupVisual{}) }},
		{name: "sidebar", check: func() { theme.ValidateSidebarVisual(theme.SidebarVisual{}) }},
		{name: "slider", check: func() { theme.ValidateSliderVisual(theme.SliderVisual{}) }},
		{name: "spinner", check: func() { theme.ValidateSpinnerVisual(theme.SpinnerVisual{}) }},
		{name: "switch", check: func() { theme.ValidateSwitchVisual(theme.SwitchVisual{}) }},
		{name: "text", check: func() { theme.ValidateTextVisual(theme.TextVisual{}) }},
		{name: "icon", check: func() { theme.ValidateIconVisual(theme.IconVisual{}) }},
		{name: "toggle", check: func() { theme.ValidateToggleVisual(theme.ToggleVisual{}) }},
		{name: "title bar", check: func() { theme.ValidateTitleBarVisual(theme.TitleBarVisual{}) }},
		{name: "node graph", check: func() { theme.ValidateNodeGraphVisual(theme.NodeGraphVisual{}) }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("validator panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
			}()
			test.check()
		})
	}
}

func TestProgressValidatorsRejectIncompleteDeclarations(t *testing.T) {
	tests := []struct {
		name  string
		check func()
	}{
		{name: "bar", check: func() { theme.ValidateProgressBarVisual(theme.ProgressBarVisual{}) }},
		{name: "circle", check: func() { theme.ValidateProgressCircleVisual(theme.ProgressCircleVisual{}) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("validator panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
			}()
			test.check()
		})
	}
}

func TestSwitchGroupValidatorRejectsInvalidMetrics(t *testing.T) {
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("validator panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
	}()
	theme.ValidateSwitchGroupVisual(theme.SwitchGroupVisual{
		Metrics: theme.SwitchGroupMetrics{VerticalGap: -1},
	})
}

func TestSidebarValidatorRejectsInvalidEasing(t *testing.T) {
	active := testtheme.Behavior()
	provider, err := theme.ComponentProvider(active)
	if err != nil {
		t.Fatal(err)
	}
	base := provider.Sidebar(active, theme.SidebarRequest{})
	for _, test := range []struct {
		name   string
		mutate func(*theme.SidebarVisual)
	}{
		{name: "expansion", mutate: func(visual *theme.SidebarVisual) {
			visual.ExpansionEasing = theme.EaseKind(255)
		}},
		{name: "collapse", mutate: func(visual *theme.SidebarVisual) {
			visual.CollapseEasing = theme.EaseKind(255)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			visual := base
			test.mutate(&visual)
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("validator panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
			}()
			theme.ValidateSidebarVisual(visual)
		})
	}
}

func TestNodeGraphValidatorRejectsMissingDefaultNodeSize(t *testing.T) {
	active := testtheme.Behavior()
	provider, err := theme.ComponentProvider(active)
	if err != nil {
		t.Fatal(err)
	}
	base := provider.NodeGraph(active, theme.NodeGraphRequest{})
	for _, test := range []struct {
		name   string
		mutate func(*theme.NodeGraphMetrics)
	}{
		{name: "width", mutate: func(metrics *theme.NodeGraphMetrics) { metrics.NodeDefaultWidth = 0 }},
		{name: "height", mutate: func(metrics *theme.NodeGraphMetrics) { metrics.NodeDefaultHeight = 0 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			visual := base
			test.mutate(&visual.Metrics)
			defer func() {
				recovered := recover()
				if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("validator panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
				}
			}()
			theme.ValidateNodeGraphVisual(visual)
		})
	}
}

func TestValidateVisualProviderRequiresCompleteProtocol(t *testing.T) {
	if err := theme.ValidateVisualProvider(theme.NeutralTheme()); !errors.Is(err, theme.ErrMissingVisualProvider) {
		t.Fatalf("neutral theme validation error = %v, want missing provider", err)
	}

	partial := theme.NeutralTheme()
	partial.VisualProvider = buttonOnlyProvider{}
	if err := theme.ValidateVisualProvider(partial); !errors.Is(err, theme.ErrIncompleteVisualProvider) {
		t.Fatalf("partial provider validation error = %v, want incomplete provider", err)
	}

	behavior := testtheme.Behavior()
	if err := theme.ValidateVisualProvider(*behavior); err != nil {
		t.Fatalf("behavior provider validation error = %v", err)
	}
	if _, err := theme.ComponentProvider(behavior); err != nil {
		t.Fatalf("behavior component provider error = %v", err)
	}
}

type buttonOnlyProvider struct{}

func (buttonOnlyProvider) Button(*theme.Theme, theme.ButtonRequest) theme.ButtonVisual {
	return theme.ButtonVisual{}
}
