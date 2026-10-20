package input

import (
	"reflect"
	"testing"

	"gioui.org/io/pointer"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/rulesetest"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestInputProviderDeclarationsReachResolvedParts(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	tests := []struct {
		name string
		mode theme.InputStyleMode
		part style.Part
	}{
		{name: "input", mode: theme.InputModeSingleLine, part: style.PartRoot},
		{name: "textarea", mode: theme.InputModeTextArea, part: style.PartRoot},
		{name: "group", mode: theme.InputModeGroup, part: style.PartContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defaults := inputVisualDeclarations(&activeTheme, Primary, false, test.mode, Medium, nil, 48).Defaults
			if !inputDeclarationAnswersToState(defaults, test.part) {
				t.Fatalf("disabled state does not affect %s", rulesetest.PartName(test.part))
			}
		})
	}

	if declaration := inputGroupActionVisual(&activeTheme, theme.InputActionRequest{}).Declaration.Resolve(style.State{Disabled: true}).Box; declaration == nil {
		t.Fatal("group action provider declaration was not consumed")
	}
}

func TestInputBehaviorStateReachesProviderCascade(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageAuto)
	visual := (&inputState{}).visualDeclarations(New("").Key("name"), &activeTheme, nil)
	defaults := visual.Defaults

	enabled := computed.ResolveStatic(ctx, style.State{}, defaults, style.Style{}, style.Style{}, style.Style{})
	focused := computed.ResolveStatic(ctx, style.State{Focused: true}, defaults, style.Style{}, style.Style{}, style.Style{})
	disabled := computed.ResolveStatic(ctx, style.State{Disabled: true}, defaults, style.Style{}, style.Style{}, style.Style{})
	if enabled.Paint == nil || enabled.Paint.Opacity == nil || *enabled.Paint.Opacity != 1 {
		t.Fatalf("enabled opacity is missing: %s", rulesetest.Describe(enabled))
	}
	if disabled.Paint == nil || disabled.Paint.Opacity == nil || *disabled.Paint.Opacity != activeTheme.DisabledOpacityValue() {
		t.Fatalf("disabled style = %s", rulesetest.Describe(disabled))
	}
	if rulesetest.Describe(enabled) == rulesetest.Describe(focused) {
		t.Fatal("focus state did not reach the cascade")
	}

	groupVisual := (&inputGroupState{}).visualDeclarations(
		Group(New("").Key("query")),
		&activeTheme,
		36,
		nil,
	)
	groupDefaults := groupVisual.Defaults
	groupEnabled := computed.ResolvePartStatic(
		ctx,
		style.PartContent,
		style.State{},
		groupDefaults,
		style.Style{},
		style.Style{},
		style.Style{},
	)
	groupDisabled := computed.ResolvePartStatic(
		ctx,
		style.PartContent,
		style.State{Disabled: true},
		groupDefaults,
		style.Style{},
		style.Style{},
		style.Style{},
	)
	if groupEnabled.Box == nil || groupEnabled.Box.Cursor == nil || *groupEnabled.Box.Cursor != pointer.CursorText {
		t.Fatalf("enabled group cursor = %#v", groupEnabled.Box)
	}
	if groupDisabled.Box == nil || groupDisabled.Box.Cursor == nil || *groupDisabled.Box.Cursor != pointer.CursorNotAllowed {
		t.Fatalf("disabled group cursor = %#v", groupDisabled.Box)
	}
}

func inputDeclarationAnswersToState(declaration style.Style, part style.Part) bool {
	enabled := rulesetest.Normalize(style.CascadePart(style.State{}, part, declaration))
	disabled := rulesetest.Normalize(style.CascadePart(style.State{Disabled: true}, part, declaration))
	return !reflect.DeepEqual(enabled, disabled)
}
