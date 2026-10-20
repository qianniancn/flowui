package theme_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
)

type missingDatePicker struct{ base.Provider }

// A wrong signature hides the embedded capability.
func (missingDatePicker) DatePicker() {}

func TestProviderReportsOnlyMissingCapabilities(t *testing.T) {
	active := base.New(theme.DefaultSeed())
	active.VisualProvider = missingDatePicker{}
	for _, err := range []error{
		theme.ValidateVisualProvider(active),
		func() error { _, err := theme.ComponentProvider(&active); return err }(),
	} {
		if !errors.Is(err, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("error = %v, want incomplete provider", err)
		}
		var detail *theme.VisualValidationError
		if !errors.As(err, &detail) || detail.Component != "Provider" || detail.Field != "DatePicker" {
			t.Fatalf("error = %#v, want only missing DatePicker", err)
		}
	}
}

func TestVisualValidationDiagnostics(t *testing.T) {
	active := base.New(theme.DefaultSeed())
	provider := base.Provider{}
	for _, test := range []struct {
		name, component, field, reason string
		validate                       func()
	}{
		{"missing resolver", "Button", "ContentGap", "must be provided", func() {
			visual := provider.Button(&active, theme.ButtonRequest{})
			visual.ContentGap = nil
			theme.ValidateButtonVisual(visual)
		}},
		{"invalid resolver result", "Button", "SpinnerSize", "finite and greater than 0", func() {
			visual := provider.Button(&active, theme.ButtonRequest{})
			visual.SpinnerSize = func() unit.Dp { return unit.Dp(math.NaN()) }
			theme.ValidateButtonVisual(visual)
		}},
		{"nested geometry", "Button", "Root.Box.Padding.Left", "finite and non-negative", func() {
			visual := provider.Button(&active, theme.ButtonRequest{
				Override: &theme.ButtonOverride{Declaration: style.Style{}.PaddingLeft(-1)},
			})
			theme.ValidateButtonVisual(visual)
		}},
		{"nested paint", "Button", "Root.Paint.Radius", "finite and non-negative", func() {
			visual := provider.Button(&active, theme.ButtonRequest{
				Override: &theme.ButtonOverride{Declaration: style.Style{}.Radius(-1)},
			})
			theme.ValidateButtonVisual(visual)
		}},
		{"negative gap", "Toolbar", "Metrics.Gap", "at least 0", func() {
			visual := provider.Toolbar(&active, theme.ToolbarRequest{})
			visual.Metrics.Gap = -1
			theme.ValidateToolbarVisual(visual)
		}},
		{"missing action width", "InputAction", "Resolved.Box.MinWidth", "greater than 0", func() {
			visual := theme.InputActionVisual{
				Declaration: style.Style{}.Height(24),
			}
			theme.ValidateInputActionVisual(visual)
		}},
		{"invalid fixed action width", "InputAction", "Resolved.Box.Width", "greater than 0", func() {
			visual := provider.InputAction(&active, theme.InputActionRequest{
				Override: &theme.InputActionOverride{Declaration: style.Style{}.Width(0)},
			})
			theme.ValidateInputActionVisual(visual)
		}},
		{"shared motion", "Checkbox", "Motion.DurationScale", "finite", func() {
			visual := provider.Checkbox(&active, theme.CheckboxRequest{})
			visual.Motion.DurationScale = float32(math.NaN())
			theme.ValidateCheckboxVisual(visual)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				value := recover()
				err, ok := value.(error)
				if !ok || !errors.Is(err, theme.ErrIncompleteVisualProvider) {
					t.Fatalf("panic = %v, want incomplete provider error", value)
				}
				var detail *theme.VisualValidationError
				if !errors.As(err, &detail) || detail.Component != test.component || detail.Field != test.field || !strings.Contains(detail.Reason, test.reason) {
					t.Fatalf("diagnostic = %#v, want %s.%s (%s)", err, test.component, test.field, test.reason)
				}
				if !strings.Contains(err.Error(), test.component+"."+test.field) {
					t.Fatalf("message does not identify the field: %v", err)
				}
			}()
			test.validate()
		})
	}
}
