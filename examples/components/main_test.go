package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/qianniancn/flowui/ui"
)

func TestNavigationRoutes(t *testing.T) {
	model := initialModel()
	send := ui.Send[Msg](func(Msg) {})
	seen := make(map[string]bool)
	count := 0
	for _, section := range catalogNavigationSections {
		for _, item := range section.Items {
			if item.Key == "" {
				t.Fatalf("%s contains an empty page key", section.Title)
			}
			if seen[item.Key] {
				t.Fatalf("duplicate page key %q", item.Key)
			}
			seen[item.Key] = true
			if pageTitle(item.Key) == "" {
				t.Errorf("page %q has no title", item.Key)
			}
			if componentPage(nil, item.Key, model, send) == nil {
				t.Errorf("page %q resolved to nil", item.Key)
			}
			count++
		}
	}
	if count != 23 {
		t.Fatalf("navigation has %d pages, want 23", count)
	}
}

func TestCatalogCoversPublicWidgetConstructors(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate component catalog sources")
	}
	catalogDir := filepath.Dir(filename)
	constructors := widgetConstructors(t, filepath.Join(catalogDir, "..", "..", "ui"))
	used := uiCalls(t, catalogDir)

	missing := make([]string, 0)
	for name := range constructors {
		if !used[name] {
			missing = append(missing, name)
		}
	}
	slices.Sort(missing)
	if len(missing) > 0 {
		t.Fatalf("component catalog is missing public widget constructors: %s", strings.Join(missing, ", "))
	}
}

func TestCatalogCoversAnimationAPIs(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate component catalog sources")
	}
	used := uiReferences(t, filepath.Dir(filename))
	expected := []string{
		"EaseLinear", "EaseQuadraticIn", "EaseQuadraticOut", "EaseQuadraticInOut",
		"EaseCubicIn", "EaseCubicOut", "EaseCubicInOut",
		"EaseQuarticIn", "EaseQuarticOut", "EaseQuarticInOut",
		"EaseQuinticIn", "EaseQuinticOut", "EaseQuinticInOut",
		"EaseSinusoidalIn", "EaseSinusoidalOut", "EaseSinusoidalInOut",
		"EaseExponentialIn", "EaseExponentialOut", "EaseExponentialInOut",
		"EaseCircularIn", "EaseCircularOut", "EaseCircularInOut",
		"EaseElasticIn", "EaseElasticOut", "EaseElasticInOut",
		"EaseBackIn", "EaseBackOut", "EaseBackInOut",
		"EaseBounceIn", "EaseBounceOut", "EaseBounceInOut",
		"LerpFloat", "LerpFloat64", "LerpColor", "LerpPoint", "LerpRect",
		"Tween", "DefaultSpring", "SpringSnappy", "SpringGentle", "SpringBouncy",
		"Timeline", "AnimateLayout", "AnimateRect",
	}
	missing := make([]string, 0)
	for _, name := range expected {
		if !used[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("component catalog is missing animation APIs: %s", strings.Join(missing, ", "))
	}
}

func widgetConstructors(t *testing.T, dir string) map[string]bool {
	t.Helper()
	constructors := make(map[string]bool)
	visitGoFiles(t, dir, func(file *ast.File) {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !function.Name.IsExported() || !returnsWidget(function.Type.Results) {
				continue
			}
			constructors[function.Name.Name] = true
		}
	})
	return constructors
}

func uiCalls(t *testing.T, dir string) map[string]bool {
	t.Helper()
	calls := make(map[string]bool)
	visitGoFiles(t, dir, func(file *ast.File) {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			packageName, ok := selector.X.(*ast.Ident)
			if ok && packageName.Name == "ui" {
				calls[selector.Sel.Name] = true
			}
			return true
		})
	})
	return calls
}

func uiReferences(t *testing.T, dir string) map[string]bool {
	t.Helper()
	references := make(map[string]bool)
	visitGoFiles(t, dir, func(file *ast.File) {
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			packageName, ok := selector.X.(*ast.Ident)
			if ok && packageName.Name == "ui" {
				references[selector.Sel.Name] = true
			}
			return true
		})
	})
	return references
}

func visitGoFiles(t *testing.T, dir string, visit func(*ast.File)) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, filepath.Join(dir, name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		visit(file)
	}
}

func returnsWidget(results *ast.FieldList) bool {
	if results == nil {
		return false
	}
	for _, result := range results.List {
		name, ok := result.Type.(*ast.Ident)
		if ok && (name.Name == "Widget" || strings.HasSuffix(name.Name, "Widget")) {
			return true
		}
	}
	return false
}

func TestThemeToggleMessage(t *testing.T) {
	model := initialModel()
	Update(&model, toggleCatalogTheme(nil))
	if !model.Dark {
		t.Fatal("first theme toggle did not enable dark mode")
	}
	Update(&model, toggleCatalogTheme(nil))
	if model.Dark {
		t.Fatal("second theme toggle did not restore light mode")
	}
}

func TestCatalogThemeScopesTransparentMenuSurfaceToTitlebar(t *testing.T) {
	model := initialModel()
	activeTheme := catalogTheme(model)
	if activeTheme.Components.Menu.BackgroundColor.A != 0xff {
		t.Fatalf("catalog menu background alpha = %d, want opaque", activeTheme.Components.Menu.BackgroundColor.A)
	}
	if activeTheme.Components.Menu.HoverColor.A != 0 {
		t.Fatalf("catalog menu hover alpha = %d, want unset fallback", activeTheme.Components.Menu.HoverColor.A)
	}

	style := catalogTitlebarMenuStyle(model)
	root := ui.Cascade(ui.StyleState{}, style)
	rootBackground, ok := root.Paint.Background.(ui.SolidColor)
	if !ok || rootBackground.Color.A != 0x99 {
		t.Fatalf("titlebar menu background = %#v, want alpha 0x99", root.Paint.Background)
	}
	item := ui.CascadePart(ui.StyleState{Hovered: true}, ui.PartItem, style)
	itemBackground, ok := item.Paint.Background.(ui.SolidColor)
	if !ok || itemBackground.Color.A != 0x99 {
		t.Fatalf("titlebar menu hover = %#v, want alpha 0x99", item.Paint.Background)
	}
	if rootBackground.Color == (color.NRGBA{}) || itemBackground.Color == (color.NRGBA{}) {
		t.Fatal("titlebar menu colors were not configured")
	}
}
