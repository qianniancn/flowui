package description

import (
	"image/color"
	"testing"

	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestOptions(t *testing.T) {
	custom := flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: color.NRGBA{R: 1, A: 255}})
	description := Description("Supporting text").For("field").Disabled(true).Style(custom)
	if description.text != "Supporting text" || description.forKey != "field" || !description.disabled {
		t.Fatal("description options were not retained")
	}
	if description.customStyle.Resolve(flowstyle.StyleState{}).Text == nil {
		t.Fatal("description style was not retained")
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
