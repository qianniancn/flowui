package ui_test

import (
	"reflect"
	"testing"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/ui"
)

type facadeModel struct {
	selected string
	open     bool
}

func TestContextExposesOnlySupportedMethods(t *testing.T) {
	contextType := reflect.TypeOf((*ui.Context)(nil))
	want := map[string]struct{}{
		"BackgroundColor": {},
		"BoolState":       {},
		"Clickable":       {},
		"Editor":          {},
		"ForegroundColor": {},
		"Invalidate":      {},
		"Language":        {},
		"ListState":       {},
		"ScrollState":     {},
		"Theme":           {},
	}
	if contextType.NumMethod() != len(want) {
		t.Errorf("ui.Context method count = %d, want %d", contextType.NumMethod(), len(want))
	}
	for index := 0; index < contextType.NumMethod(); index++ {
		name := contextType.Method(index).Name
		if _, ok := want[name]; !ok {
			t.Errorf("ui.Context unexpectedly exposes method %s", name)
		}
	}
}

type facadeMsg struct {
	selected string
	open     *bool
}

type externalWidget struct{}

func (externalWidget) Layout(*ui.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{}
}

func facadeUpdate(model *facadeModel, msg facadeMsg) {
	if msg.selected != "" {
		model.selected = msg.selected
	}
	if msg.open != nil {
		model.open = *msg.open
	}
}

func facadeView(ctx *ui.Context, model facadeModel, send ui.Send[facadeMsg]) ui.Widget {
	if ctx != nil {
		var _ ui.Theme = ctx.Theme()
		var _ ui.Language = ctx.Language()
	}
	items := []ui.SelectItem{{Key: "one", Label: "One"}}
	tabs := []ui.TabItem{{Key: "general", Label: "General", Panel: ui.Text("Panel")}}

	return ui.Column(
		ui.Surface(externalWidget{}),
		ui.Surface(
			ui.Tabs("settings", "general", tabs).Variant(ui.TabsSecondary),
		).Variant(ui.SurfaceSecondary).Radius(8),
		ui.Select("choice", model.selected, items).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }),
		ui.ProgressBar("progress", 50).ShowValue(),
		ui.Modal("dialog", model.open, "Settings", ui.Text("Body")).
			OnOpenChange(func(open bool) { send(facadeMsg{open: &open}) }),
	)
}

func TestPublicFacadeImportContract(t *testing.T) {
	var _ ui.Widget = externalWidget{}
	var _ ui.Update[facadeModel, facadeMsg] = facadeUpdate
	var _ ui.View[facadeModel, facadeMsg] = facadeView
	var _ ui.Cmd[facadeMsg] = ui.Do(func(ui.Send[facadeMsg]) {})
	var _ ui.Option = ui.Title("FlowUI")
	var _ ui.Option = ui.Locale(ui.LanguageEnglish)
	var _ ui.DatePickerLocale = ui.DatePickerEnglish()

	if root := facadeView(nil, facadeModel{}, func(facadeMsg) {}); root == nil {
		t.Fatal("public facade returned a nil widget tree")
	}
}
