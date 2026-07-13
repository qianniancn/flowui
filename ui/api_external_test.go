package ui_test

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"reflect"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type facadeModel struct {
	selected      string
	expanded      []string
	tableSelected []string
	tableSort     ui.TableSortDescriptor
	open          bool
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
	selected      string
	expanded      []string
	tableSelected []string
	tableSort     *ui.TableSortDescriptor
	open          *bool
}

type externalWidget struct{}

func (externalWidget) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	progress, _ := ui.Tween("external-tween", 1).
		Initial(0).
		Revision(1).
		Duration(time.Millisecond).
		Easing(ui.EaseCubicOut).
		Disabled(false).
		Sample(ctx, gtx)
	easings := []ui.Easing{
		ui.EaseLinear,
		ui.EaseQuadraticIn, ui.EaseQuadraticOut, ui.EaseQuadraticInOut,
		ui.EaseCubicIn, ui.EaseCubicOut, ui.EaseCubicInOut,
		ui.EaseQuarticIn, ui.EaseQuarticOut, ui.EaseQuarticInOut,
		ui.EaseQuinticIn, ui.EaseQuinticOut, ui.EaseQuinticInOut,
		ui.EaseSinusoidalIn, ui.EaseSinusoidalOut, ui.EaseSinusoidalInOut,
		ui.EaseExponentialIn, ui.EaseExponentialOut, ui.EaseExponentialInOut,
		ui.EaseCircularIn, ui.EaseCircularOut, ui.EaseCircularInOut,
		ui.EaseElasticIn, ui.EaseElasticOut, ui.EaseElasticInOut,
		ui.EaseBackIn, ui.EaseBackOut, ui.EaseBackInOut,
		ui.EaseBounceIn, ui.EaseBounceOut, ui.EaseBounceInOut,
	}
	for _, easing := range easings {
		_ = easing(progress)
	}
	_ = ui.LerpFloat(0, 1, progress)
	_ = ui.LerpFloat64(0, 1, progress)
	_ = ui.LerpColor(color.NRGBA{}, color.NRGBA{A: 0xff}, progress)
	_ = ui.LerpPoint(f32.Point{}, f32.Pt(1, 1), progress)
	_ = ui.LerpRect(image.Rectangle{}, image.Rect(1, 1, 2, 2), progress)
	return layout.Dimensions{}
}

func facadeUpdate(model *facadeModel, msg facadeMsg) {
	if msg.selected != "" {
		model.selected = msg.selected
	}
	if msg.expanded != nil {
		model.expanded = append([]string(nil), msg.expanded...)
	}
	if msg.tableSelected != nil {
		model.tableSelected = append([]string(nil), msg.tableSelected...)
	}
	if msg.tableSort != nil {
		model.tableSort = *msg.tableSort
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
	treeItems := []ui.TreeItem{{Key: "folder", Label: "Folder", Children: []ui.TreeItem{{Key: "file", Label: "File"}}}}
	tableColumns := []ui.TableColumn{
		{Key: "name", Label: "Name", Sortable: true, RowHeader: true, Weight: 2},
		{Key: "status", Label: "Status", Width: 120, Align: ui.TableAlignEnd},
	}
	tableRows := []ui.TableRow{{
		Key: "member", Label: "Member",
		Cells: []ui.TableCell{{Text: "Member"}, {Content: ui.Chip("Active").Color(ui.ChipSuccess)}},
	}}
	tabs := []ui.TabItem{{Key: "general", Label: "General", Panel: ui.Text("Panel")}}

	return ui.Column(
		ui.Image(paint.ImageOp{}).
			Fit(ui.ImageCover).
			Position(ui.AlignCenter).
			Width(160).
			Height(90).
			Radius(16).
			Opacity(0.8).
			Alt("Cover image"),
		ui.Avatar("AM").
			Image(paint.ImageOp{}).
			Alt("Alex Morgan").
			Fallback(ui.Icon(lucide.UserRound).Size(20)).
			Color(ui.AvatarAccent).
			Variant(ui.AvatarSoft).
			Size(ui.AvatarLarge),
		ui.Badge(ui.Avatar("AM"), "5").
			Content(ui.Icon(lucide.Bell).Size(10)).
			Alt("Notifications").
			Color(ui.BadgeDanger).
			Variant(ui.BadgeSoft).
			Size(ui.BadgeSmall).
			Placement(ui.BadgeBottomRight),
		ui.Meter("storage-meter", 60).
			Label("Storage").
			Alt("Storage usage").
			ShowValue().
			ValueText("60 GB").
			ValueFormatter(func(value float64) string { return "formatted" }).
			ValueContent(ui.Text("60 GB")).
			Range(0, 100).
			Color(ui.MeterSuccess).
			Size(ui.MeterLarge).
			Disabled(false),
		ui.LineChart("traffic", []ui.LineChartSeries{
			ui.LineSeries("requests", "Requests", []float64{12, 18, 16}).
				Smooth(true).
				ShowPoints(true).
				PointSize(8).
				ConnectNulls(false).
				Step(ui.LineStepMiddle).
				LineStyle(ui.LineDashed).
				Area(true).
				AreaColor(color.NRGBA{R: 1, A: 0x40}).
				Sampling(ui.LineSamplingMinMax).
				Width(2).
				Hidden(false),
			ui.LineXYSeries("latency", "Latency", []ui.LineChartPoint{{X: 0, Y: 8}, {X: 1, Y: 11}}).
				Smoothness(0.35),
		}).
			Categories([]string{"Mon", "Tue", "Wed"}).
			Height(280).
			Grid(true).
			Legend(true).
			Tooltip(true).
			IncludeZero(false).
			XAxis("Day").
			YAxis("Requests").
			XTicks(4).
			YTicks(5).
			FormatX(func(value float64) string { return fmt.Sprint(value) }).
			FormatY(func(value float64) string { return fmt.Sprint(value) }).
			Animation(true).
			AnimationDuration(time.Second).
			AnimationEasing(ui.EaseCubicOut).
			UpdateAnimationDuration(500*time.Millisecond).
			UpdateAnimationEasing(ui.EaseCubicInOut).
			OnLegendChange(func(string, bool) {}).
			OnDataClick(func(ui.ChartSelection) {}).
			TooltipContent(func(selection ui.ChartSelection) ui.Widget {
				var _ ui.ChartDatum = selection.Items[0]
				return ui.Text(selection.Label)
			}).
			DataWindow(0.1, 0.9).
			OnDataWindowChange(func(ui.ChartDataWindow) {}).
			MarkLines([]ui.ChartMarkLine{ui.MarkLine(ui.ChartAxisY, 10).Text("Target").Width(2)}).
			MarkAreas([]ui.ChartMarkArea{ui.MarkArea(ui.ChartAxisX, 0, 1).Text("Window")}).
			MarkPoints([]ui.ChartMarkPoint{ui.MarkPoint(1, 10).Text("Peak").Size(10)}).
			Label("Traffic").
			EmptyText("No samples").
			Disabled(false),
		ui.BarChart("sales", []ui.BarChartSeries{
			ui.BarSeries("online", "Online", []float64{12, 18, 16}).
				Color(color.NRGBA{R: 1, A: 0xff}).
				ItemColors([]color.NRGBA{{R: 2, A: 0xff}}).
				Stack("total").
				Width(18).
				MaxWidth(24).
				MinHeight(1).
				Radius(3).
				Background(true).
				ShowLabels(true).
				LabelPosition(ui.BarLabelInside).
				FormatLabel(func(value float64) string { return fmt.Sprint(value) }).
				Hidden(false),
		}).
			Categories([]string{"Mon", "Tue", "Wed"}).
			Height(280).
			Grid(true).
			Legend(true).
			Tooltip(true).
			IncludeZero(false).
			YRange(0, 20).
			XAxis("Day").
			YAxis("Sales").
			CategoryAxis("Category").
			ValueAxis("Value").
			ValueRange(0, 20).
			ValueTicks(5).
			FormatValue(func(value float64) string { return fmt.Sprint(value) }).
			YTicks(5).
			BarGap(0.1).
			CategoryGap(0.25).
			FormatY(func(value float64) string { return fmt.Sprint(value) }).
			Animation(true).
			AnimationDuration(time.Second).
			AnimationEasing(ui.EaseCubicOut).
			UpdateAnimationDuration(500*time.Millisecond).
			UpdateAnimationEasing(ui.EaseCubicInOut).
			OnLegendChange(func(string, bool) {}).
			OnDataClick(func(ui.ChartSelection) {}).
			TooltipContent(func(selection ui.ChartSelection) ui.Widget { return ui.Text(selection.Label) }).
			DataWindow(0.1, 0.9).
			OnDataWindowChange(func(ui.ChartDataWindow) {}).
			MarkLines([]ui.ChartMarkLine{ui.MarkLine(ui.ChartAxisY, 10)}).
			MarkAreas([]ui.ChartMarkArea{ui.MarkArea(ui.ChartAxisX, 0, 1)}).
			MarkPoints([]ui.ChartMarkPoint{ui.MarkPoint(1, 10)}).
			Orientation(ui.BarHorizontal).
			Label("Sales").
			EmptyText("No samples").
			Disabled(false),
		ui.Menubar("application-menu", []ui.MenubarItem{
			ui.MenubarMenu("file", "File", []ui.MenuItem{{Key: "new", Label: "New"}}).
				OnAction(func(string) {}).
				Width(220),
			ui.MenubarMenuContent(
				"edit",
				"Edit",
				ui.Menu("edit-menu", []ui.MenuItem{{Key: "copy", Label: "Copy"}}),
			).Disabled(false),
		}).
			Alt("Application menu").
			Orientation(ui.MenubarHorizontal).
			LoopFocus(true).
			Modal(true).
			DefaultOpenKey("").
			OpenKey("").
			OnOpenChange(func(string) {}).
			Disabled(false),
		ui.Alert("Update available", "Refresh to get the latest features.").
			Status(ui.AlertAccent).
			Indicator(ui.Icon(lucide.Info).Size(16)).
			Content(ui.Text("Custom alert content")).
			Action(ui.CloseButton("dismiss-alert")),
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
		ui.Tree("files", model.selected, treeItems).
			ExpandedKeys(model.expanded).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }).
			OnExpandedChange(func(keys []string) { send(facadeMsg{expanded: keys}) }).
			OnAction(func(string) {}).
			Variant(ui.TreeSurface).
			SelectionMode(ui.TreeSelectionSingle).
			DisabledKeys([]string{"file"}).
			AllowEmptySelection().
			EmptyText("No files").
			MaxHeight(240),
		ui.Table("members", tableColumns, tableRows).
			Variant(ui.TableSecondary).
			SelectionMode(ui.TableSelectionMultiple).
			SelectedKey(model.selected).
			SelectedKeys(model.tableSelected).
			SortDescriptor(model.tableSort).
			DisabledKeys([]string{"archived"}).
			ShowSelectionIndicator().
			AllowEmptySelection().
			EmptyText("No members").
			EmptyContent(ui.Text("Empty")).
			Footer(ui.Text("Footer")).
			MinWidth(640).
			MaxHeight(280).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }).
			OnSelectionChange(func(keys []string) { send(facadeMsg{tableSelected: keys}) }).
			OnSortChange(func(sort ui.TableSortDescriptor) { send(facadeMsg{tableSort: &sort}) }).
			OnAction(func(string) {}).
			Disabled(false),
		ui.ContextMenu(
			"member-actions",
			ui.Text("Member"),
			ui.Menu("actions", []ui.MenuItem{
				{Key: "open", Label: "Open", Leading: ui.Icon(lucide.UserRound).Size(16)},
				{Key: "favorite", Label: "Favorite", Kind: ui.MenuItemCheckbox, Checked: model.open, KeepOpen: true},
				{Key: "compact", Label: "Compact", Kind: ui.MenuItemRadio, RadioGroup: "density", Value: "compact", KeepOpen: true},
				ui.MenuSeparator(),
				ui.MenuGroupLabel("More"),
				{Key: "share", Label: "Share", Kind: ui.MenuItemSubmenu, Children: []ui.MenuItem{{Key: "copy-link", Label: "Copy link"}}},
				{Key: "delete", Label: "Delete", Variant: ui.MenuItemDanger},
			}).
				Sections([]ui.MenuSection{{Title: "Actions", Items: []ui.MenuItem{{Key: "inspect", Label: "Inspect"}}}}).
				AutoSeparateSections(true).
				BeforeContent(ui.Text("Menu header")).
				AfterContent(ui.Text("Menu footer")).
				EmptyText("No actions").
				SelectionMode(ui.MenuSelectionMultiple).
				SelectedKey(model.selected).
				SelectedKeys(model.tableSelected).
				DisabledKeys([]string{"delete"}).
				OnAction(func(string) {}).
				OnChange(func(string) {}).
				OnSelectionChange(func([]string) {}).
				OnCheckedChange(func(string, bool) {}).
				OnRadioChange(func(string, string) {}).
				CloseOnSelect(false).
				Disabled(false).
				Width(240),
		).
			Open(model.open).
			DefaultOpen(false).
			OnOpenChange(func(open bool) { send(facadeMsg{open: &open}) }).
			LongPressDisabled(false).
			Disabled(false),
		ui.Dropdown("account-dropdown", ui.Button("account-trigger", ui.Text("Account")), []ui.DropdownItem{
			{
				Key: "profile", Label: "Profile", Description: "View profile", Shortcut: "Ctrl+P", Leading: ui.Icon(lucide.UserRound).Size(16),
				Indicator: func(bool) ui.Widget { return ui.Icon(lucide.Check).Size(12) },
			},
			ui.DropdownSeparator(),
			{
				Key: "theme", Label: "Theme", IndicatorType: ui.DropdownIndicatorCheckmark,
				SubmenuIndicator: ui.Icon(lucide.ChevronRight).Size(14), Children: []ui.DropdownItem{{Key: "dark", Label: "Dark"}},
			},
			{Key: "delete", Label: "Delete", Variant: ui.DropdownItemDanger, Disabled: true},
		}).
			Sections([]ui.DropdownSection{{Title: "Account", Items: []ui.DropdownItem{{Key: "settings", Label: "Settings"}}}}).
			BeforeContent(ui.Text("Signed in")).
			AfterContent(ui.Text("Footer")).
			EmptyText("No actions").
			SelectionMode(ui.DropdownSelectionMultiple).
			SelectedKey(model.selected).
			SelectedKeys(model.tableSelected).
			DisabledKeys([]string{"delete"}).
			OnAction(func(string) {}).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }).
			OnSelectionChange(func(keys []string) { send(facadeMsg{tableSelected: keys}) }).
			Open(model.open).
			DefaultOpen(false).
			OnOpenChange(func(open bool) { send(facadeMsg{open: &open}) }).
			TriggerMode(ui.DropdownTriggerPress).
			Placement(ui.PopoverBottomStart).
			Offset(4).
			ShouldFlip(true).
			AvoidOverflow(true).
			CloseOnSelect(false).
			Disabled(false).
			Width(240),
		ui.Input("email", "").
			Placeholder("name@example.com").
			Type(ui.InputEmail).
			ReadOnly(false).
			MaxLength(120).
			Label("Email address"),
		ui.TextArea("biography", "").
			Placeholder("Tell us about yourself").
			Variant(ui.TextAreaSecondary).
			Rows(4).
			ReadOnly(false).
			MaxLength(500).
			Label("Biography").
			Invalid(false).
			Disabled(false).
			FullWidth().
			OnChange(func(string) {}),
		ui.InputGroupTextArea(
			ui.TextArea("notes", "").Rows(5),
		).
			Prefix(ui.Icon(lucide.MessageSquare).Size(16)).
			Suffix(ui.Icon(lucide.SendHorizontal).Size(16)).
			FullWidth(),
		ui.Checkbox("agreement", model.open, "Agreement").
			Variant(ui.CheckboxSecondary).
			Indeterminate(false).
			ReadOnly(false).
			Required(true).
			Description("Accept the agreement").
			ErrorMessage("Agreement is required").
			Indicator(func(state ui.CheckboxIndicatorState) ui.Widget {
				if !state.Checked && !state.Indeterminate {
					return nil
				}
				return ui.Icon(lucide.Check).Size(10)
			}).
			Invalid(false).
			Disabled(false).
			OnChange(func(checked bool) { send(facadeMsg{open: &checked}) }),
		ui.InputGroup(
			ui.Input("website", "flowui").Label("Website"),
		).
			Prefix(ui.Icon(lucide.Globe).Size(16)).
			Suffix(ui.Text(".com")).
			SuffixPadding(12, 0).
			Variant(ui.InputSecondary).
			Invalid(false).
			Disabled(false).
			FullWidth(),
		ui.ProgressBar("progress", 50).ShowValue(),
		ui.CloseButton("close").Label("Dismiss"),
		ui.Chip("Completed").
			Color(ui.ChipSuccess).
			Variant(ui.ChipSoft).
			Size(ui.ChipSmall).
			StartContent(ui.Icon(lucide.CircleCheck).Size(12)).
			EndContent(ui.Icon(lucide.X).Size(12)),
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
		ui.AlertDialog("confirm", model.open, "Delete project?", "This action cannot be undone.").
			Status(ui.AlertDialogDanger).
			Icon(ui.Icon(lucide.CircleAlert).Size(20)).
			Body(ui.Text("Custom body")).
			Header(ui.Text("Custom header")).
			Footer(ui.Button("cancel", ui.Text("Cancel"))).
			Size(ui.AlertDialogSmall).
			Placement(ui.AlertDialogCenter).
			Backdrop(ui.AlertDialogBackdropBlur).
			Dismissable(false).
			KeyboardDismissDisabled(true).
			CloseButton(true).
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
