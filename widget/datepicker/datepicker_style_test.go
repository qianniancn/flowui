package datepicker

import (
	"errors"
	"testing"

	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/theme"
)

func TestDatePickerProductionRequiresVisualProvider(t *testing.T) {
	activeTheme := theme.NeutralTheme()

	assertMissingProvider := func(name string, fn func()) {
		t.Helper()
		defer func() {
			recovered := recover()
			err, ok := recovered.(error)
			if !ok || !errors.Is(err, theme.ErrMissingVisualProvider) {
				t.Fatalf("%s panic = %v, want %v", name, recovered, theme.ErrMissingVisualProvider)
			}
		}()
		fn()
	}

	assertMissingProvider("date picker variant visual", func() {
		datePickerVisualDeclarationForVariant(&activeTheme, field.Primary)
	})
}
