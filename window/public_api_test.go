package window_test

import (
	"errors"
	"testing"

	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/window"
)

func TestWindowOptionsResolveInDeclarationOrder(t *testing.T) {
	var gotError error
	closeHandler := func() window.CloseDecision { return window.CloseCancel }
	cfg := window.Resolve(
		nil,
		window.Title("FlowUI"),
		window.Size(900, 600),
		window.MinSize(320, 200),
		window.MaxSize(1600, 1200),
		window.CenterOnStart(),
		window.Locale(window.LanguageChinese),
		window.Translations(i18n.NewCatalog(i18n.Translation{Language: "fr", Key: i18n.ActionClose, Value: "Fermer"})),
		window.OnError(func(err error) { gotError = err }),
		window.OnWindowCloseRequest(closeHandler),
		window.RetainModelOnClose(),
		window.Theme(theme.NeutralDarkTheme()),
		window.CustomizeTheme(func(theme *theme.Theme) { theme.Palette.Background = theme.Palette.Surface }),
	)

	if len(cfg.Native) != 4 {
		t.Fatalf("native options = %d, want 4", len(cfg.Native))
	}
	if cfg.InitialActions&window.ActionCenter == 0 {
		t.Fatal("center-on-start action was not resolved")
	}
	if cfg.Language != window.LanguageChinese || !cfg.RetainModel {
		t.Fatalf("resolved config = %#v", cfg)
	}
	if got := cfg.Translations.Text("fr-CA", i18n.ActionClose); got != "Fermer" {
		t.Fatalf("window translation = %q, want Fermer", got)
	}
	if cfg.ErrorHandler == nil || cfg.CloseRequestHandler == nil || len(cfg.ThemeOps) != 2 {
		t.Fatal("window callbacks or theme options were not retained")
	}

	resolvedTheme := theme.NeutralTheme()
	for _, apply := range cfg.ThemeOps {
		apply(&resolvedTheme)
	}
	if resolvedTheme.Palette.Background != resolvedTheme.Palette.Surface {
		t.Fatal("theme customization was not applied")
	}

	want := errors.New("failure")
	cfg.ErrorHandler(want)
	if !errors.Is(gotError, want) {
		t.Fatal("error callback did not receive its error")
	}
	if got := cfg.CloseRequestHandler(); got != window.CloseCancel {
		t.Fatalf("close decision = %v, want cancel", got)
	}
}
