package ui_test

import (
	"context"
	"reflect"
	"testing"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide/lucide"
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
		ui.Card(
			ui.CardHeader(
				ui.CardTitle("Settings"),
				ui.CardDescription("Public facade contract"),
			),
			ui.CardContent(externalWidget{}),
			ui.CardFooter(ui.Text("Footer")),
		).Variant(ui.CardSecondary),
		ui.Surface(externalWidget{}),
		ui.Surface(
			ui.Tabs("settings", "general", tabs).Variant(ui.TabsSecondary),
		).Variant(ui.SurfaceSecondary).Radius(8),
		ui.Select("choice", model.selected, items).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }),
		ui.Input("email", "").
			Placeholder("name@example.com").
			Type(ui.InputEmail).
			ReadOnly(false).
			MaxLength(120).
			Label("Email address"),
		ui.ProgressBar("progress", 50).ShowValue(),
		ui.CloseButton("close").Label("Dismiss"),
		ui.ToggleButton("pin", model.open, ui.Text("Pin")).
			OnChange(func(selected bool) { send(facadeMsg{open: &selected}) }),
		ui.Icon(lucide.Search).Size(18),
		ui.Spinner().Color(ui.SpinnerSuccess).Size(ui.SpinnerSmall).Label("Saving"),
		ui.Slider("volume", 30).Label("Volume").ShowValue(),
		ui.RangeSlider("price", 10, 80).Range(0, 100).Step(5),
		ui.Tooltip("save-help", ui.Button("save", ui.Text("Save")), ui.Text("Save changes")).
			Placement(ui.TooltipTop).
			Arrow(true).
			Delay(0),
		ui.ToastProvider("toasts", []ui.ToastItem{
			ui.Toast("saved", "Saved").Variant(ui.ToastSuccess).Description("Changes saved"),
		}).OnClose(func(string) {}),
		ui.Modal("dialog", model.open, "Settings", ui.Text("Body")).
			OnOpenChange(func(open bool) { send(facadeMsg{open: &open}) }),
	)
}

func TestPublicFacadeImportContract(t *testing.T) {
	_ = ui.RunWithSubscriptions[facadeModel, facadeMsg]
	var _ ui.Widget = externalWidget{}
	var _ ui.Update[facadeModel, facadeMsg] = facadeUpdate
	var _ ui.View[facadeModel, facadeMsg] = facadeView
	var _ ui.Cmd[facadeMsg] = ui.Do(func(ui.Send[facadeMsg]) {})
	var _ ui.Cmd[facadeMsg] = ui.DoContext(func(context.Context, ui.Send[facadeMsg]) error { return nil })
	var _ ui.Subscriptions[facadeModel, facadeMsg] = func(facadeModel) []ui.Subscription[facadeMsg] {
		return []ui.Subscription[facadeMsg]{
			ui.Subscribe("events", func(context.Context, ui.Send[facadeMsg]) error { return nil }),
		}
	}
	var _ ui.Option = ui.Title("FlowUI")
	var _ ui.Option = ui.OnError(func(error) {})
	var _ error = ui.ErrEffectShutdownTimeout
	var _ ui.Option = ui.Locale(ui.LanguageEnglish)
	var _ ui.DatePickerLocale = ui.DatePickerEnglish()

	if root := facadeView(nil, facadeModel{}, func(facadeMsg) {}); root == nil {
		t.Fatal("public facade returned a nil widget tree")
	}
}
