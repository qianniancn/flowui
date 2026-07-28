package description

import (
	"testing"

	"github.com/qianniancn/FlowUI/internal/theme"
)

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
