package field

import (
	"testing"

	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestDefaultDeclarationUsesTargetPartAndConditions(t *testing.T) {
	activeTheme := defaultTheme()
	options := DeclarationOptions{
		TargetPart:          flowstyle.PartContent,
		Radius:              activeTheme.Components.Input.Radius,
		FocusRingWidth:      activeTheme.Components.Input.FocusRingWidth,
		InvalidOutlineWidth: activeTheme.Components.Input.InvalidOutlineWidth,
		ShadowOpacity:       activeTheme.Components.Input.ShadowOpacity,
		ShadowStrength:      activeTheme.Components.Input.ShadowStrength,
	}
	declaration := DefaultDeclaration(activeTheme, Primary, options)
	base := flowstyle.CascadePart(flowstyle.StyleState{}, flowstyle.PartContent, declaration)
	hovered := flowstyle.CascadePart(flowstyle.StyleState{Hovered: true}, flowstyle.PartContent, declaration)
	focused := flowstyle.CascadePart(flowstyle.StyleState{Focused: true}, flowstyle.PartContent, declaration)
	invalid := flowstyle.CascadePart(flowstyle.StyleState{Invalid: true}, flowstyle.PartContent, declaration)

	if base.Paint == nil || base.Paint.Background != flowstyle.TokenFieldBackground || len(base.Paint.Shadows) == 0 {
		t.Fatalf("base field declaration = %#v", base.Paint)
	}
	if hovered.Paint.Background != flowstyle.TokenFieldHover {
		t.Fatalf("hover background = %#v", hovered.Paint.Background)
	}
	if focused.Paint.Outline == nil || focused.Paint.Outline.Color != flowstyle.TokenFocus {
		t.Fatalf("focused outline = %#v", focused.Paint.Outline)
	}
	if invalid.Paint.Outline == nil || invalid.Paint.Outline.Color != flowstyle.TokenDanger {
		t.Fatalf("invalid outline = %#v", invalid.Paint.Outline)
	}
}

func TestSecondaryDeclarationRemovesControlShadow(t *testing.T) {
	activeTheme := defaultTheme()
	declaration := DefaultDeclaration(activeTheme, Secondary, DeclarationOptions{
		TargetPart:     flowstyle.PartContent,
		ShadowOpacity:  1,
		ShadowStrength: 1,
	})
	resolved := flowstyle.CascadePart(flowstyle.StyleState{}, flowstyle.PartContent, declaration)
	if resolved.Paint == nil || resolved.Paint.Background != flowstyle.TokenDefault || len(resolved.Paint.Shadows) != 0 {
		t.Fatalf("secondary declaration = %#v", resolved.Paint)
	}
}

func TestRootTargetSupportsTextInputs(t *testing.T) {
	declaration := DefaultDeclaration(defaultTheme(), Primary, DeclarationOptions{})
	root := declaration.Resolve(flowstyle.StyleState{})
	if root.Paint == nil || root.Paint.Background != flowstyle.TokenFieldBackground {
		t.Fatalf("root field declaration = %#v", root.Paint)
	}
}

func defaultTheme() *theme.Theme {
	value := theme.DefaultTheme()
	return &value
}
