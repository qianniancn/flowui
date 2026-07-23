package label

import (
	"image/color"
	"testing"

	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestOptions(t *testing.T) {
	custom := flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: color.NRGBA{R: 1, A: 255}})
	label := Label("Email").For("email").Required(true).Disabled(true).Invalid(true).Style(custom)
	if label.text != "Email" || label.forKey != "email" || !label.required || !label.disabled || !label.invalid {
		t.Fatal("label options were not retained")
	}
	if label.customStyle.Resolve(flowstyle.StyleState{}).Text == nil {
		t.Fatal("label style was not retained")
	}
}

func TestStyleStates(t *testing.T) {
	theme := theme.DefaultTheme()
	foreground := color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	base := labelStyleFor(&theme, foreground, false, false)
	invalid := labelStyleFor(&theme, foreground, false, true)
	disabled := labelStyleFor(&theme, foreground, true, true)
	if base.text != foreground || invalid.text != theme.Palette.Danger {
		t.Fatal("label state colors are incorrect")
	}
	if disabled.text != theme.DisabledColor(theme.Palette.Danger) {
		t.Fatal("disabled label color is incorrect")
	}
}
