package description

import (
	"testing"

	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestOptions(t *testing.T) {
	description := Description("Supporting text").For("field").Disabled(true)
	if description.text != "Supporting text" || description.forKey != "field" || !description.disabled {
		t.Fatal("description options were not retained")
	}
}

func TestStyleUsesMutedForeground(t *testing.T) {
	theme := theme.DefaultTheme()
	base := descriptionStyleFor(&theme, false)
	disabled := descriptionStyleFor(&theme, true)
	if base.text != theme.Palette.MutedForeground {
		t.Fatal("description did not use muted foreground")
	}
	if disabled.text != theme.DisabledColor(theme.Palette.MutedForeground) {
		t.Fatal("disabled description color was not applied")
	}
}
