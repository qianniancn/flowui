package ui

import (
	"testing"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/window"
)

func themeContractProgram() Program[int, struct{}] {
	return NewProgram(
		0,
		func(*int, struct{}) Cmd[struct{}] { return nil },
		func(*Context, int, Send[struct{}]) Widget { return nil },
	)
}

func TestNewWindowRequiresExplicitTheme(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("NewWindow accepted a missing explicit theme")
		}
	}()

	NewWindow("missing-theme", themeContractProgram())
}

func TestNewWindowRejectsThemeWithoutVisualProvider(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("NewWindow accepted a theme without a visual provider")
		}
	}()

	NewWindow(
		"missing-provider",
		themeContractProgram(),
		window.Theme(theme.New(theme.Seed{})),
	)
}

func TestNewWindowAcceptsShippedVisualProvider(t *testing.T) {
	spec := NewWindow(
		"default-theme",
		themeContractProgram(),
		window.Theme(base.Light()),
	)
	if spec.Key() != "default-theme" {
		t.Fatalf("window key = %q, want default-theme", spec.Key())
	}
}

func TestNewWindowRejectsProviderRemovedByThemeCustomization(t *testing.T) {
	cfg := window.Resolve(
		window.Theme(base.Light()),
		window.CustomizeTheme(func(value *theme.Theme) {
			value.VisualProvider = nil
		}),
	)
	active := cfg.InitialTheme
	for _, apply := range cfg.ThemeOps {
		apply(&active)
	}
	if err := theme.ValidateVisualProvider(active); err == nil {
		t.Fatal("customization removed the visual provider without an error")
	}
}
