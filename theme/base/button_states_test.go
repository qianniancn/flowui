package base_test

import (
	"testing"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
)

// TestSecondaryButtonPressedStepsPastHover guards the tinted secondary button:
// its hover and pressed rows used to be the same colour, so a press looked
// exactly like a hover.
func TestSecondaryButtonPressedStepsPastHover(t *testing.T) {
	active := base.Light()
	visual := base.Provider{}.Button(&active, theme.ButtonRequest{Variant: theme.ButtonSecondary})

	rest := visual.Variant.Resolve(style.State{})
	hover := visual.Variant.Resolve(style.State{Hovered: true})
	pressed := visual.Variant.Resolve(style.State{Hovered: true, Pressed: true})
	if hover.Paint.Background == pressed.Paint.Background {
		t.Fatalf("pressed background = %v, same as hover", hover.Paint.Background)
	}
	if rest.Paint.Background == hover.Paint.Background {
		t.Fatalf("hover background = %v, same as resting", hover.Paint.Background)
	}
}
