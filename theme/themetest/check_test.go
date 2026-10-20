package themetest

import (
	"reflect"
	"testing"

	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/theme/buttonstyle"
)

func TestThemes(t *testing.T) {
	seed := theme.DefaultSeed()
	for _, test := range []struct {
		name   string
		active theme.Theme
	}{
		{"default/light", base.Light()}, {"default/dark", base.Dark()},
		{"base/light", base.New(seed)}, {"base/dark", base.NewDark(seed)},
		{"custom/light", customTheme(false)}, {"custom/dark", customTheme(true)},
	} {
		t.Run(test.name, func(t *testing.T) { Check(t, test.active) })
	}
}

type customProvider struct{ base.Provider }

func (customProvider) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	config := buttonstyle.DefaultConfig()
	config.Radius = 0
	config.ContentGap = 0
	return buttonstyle.BuildButton(active, request, config)
}

func customTheme(dark bool) theme.Theme {
	active := base.New(theme.DefaultSeed())
	if dark {
		active = base.NewDark(theme.DefaultSeed())
	}
	active.Components.Toolbar.Gap = 11
	active.VisualProvider = customProvider{}
	return active
}

// This guard forces a new component capability to gain an acceptance case.
func TestCasesCoverCompleteProtocol(t *testing.T) {
	active := base.New(theme.DefaultSeed())
	methods := reflect.TypeFor[theme.ComponentVisualProvider]()
	cases := componentCases(&active, base.Provider{})
	seen := make(map[string]bool)
	for _, test := range cases {
		if seen[test.name] {
			t.Errorf("duplicate case %s", test.name)
		}
		seen[test.name] = true
		if _, ok := methods.MethodByName(test.name); !ok {
			t.Errorf("case %s has no provider method", test.name)
		}
	}
	for method := range methods.Methods() {
		if name := method.Name; !seen[name] {
			t.Errorf("missing acceptance case for %s", name)
		}
	}
}
