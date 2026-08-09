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
	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

type facadeModel struct {
	selected      string
	expanded      []string
	tableSelected []string
	tableSort     ui.TableSortDescriptor
	open          bool
}

var _ ui.ToolbarTheme = ui.DefaultTheme().Components.Toolbar

func TestStyleBackedWidgetsExposeStyleMethod(t *testing.T) {
	styleType := reflect.TypeFor[ui.Style]()
	styleBacked := map[string]reflect.Type{
		"Alert":             reflect.TypeFor[ui.AlertWidget](),
		"AlertDialog":       reflect.TypeFor[ui.AlertDialogWidget](),
		"Avatar":            reflect.TypeFor[ui.AvatarWidget](),
		"Badge":             reflect.TypeFor[ui.BadgeWidget](),
		"BarChart":          reflect.TypeFor[ui.BarChartWidget](),
		"Box":               reflect.TypeFor[ui.BoxWidget](),
		"Button":            reflect.TypeFor[ui.ButtonWidget](),
		"ButtonGroup":       reflect.TypeFor[ui.ButtonGroupWidget](),
		"Card":              reflect.TypeFor[ui.CardWidget](),
		"CandlestickChart":  reflect.TypeFor[ui.CandlestickChartWidget](),
		"Checkbox":          reflect.TypeFor[ui.CheckboxWidget](),
		"Chip":              reflect.TypeFor[ui.ChipWidget](),
		"ColorArea":         reflect.TypeFor[ui.ColorAreaWidget](),
		"ColorWheel":        reflect.TypeFor[ui.ColorWheelWidget](),
		"ColorField":        reflect.TypeFor[ui.ColorFieldWidget](),
		"ColorPicker":       reflect.TypeFor[ui.ColorPickerWidget](),
		"ColorSlider":       reflect.TypeFor[ui.ColorSliderWidget](),
		"ColorSwatch":       reflect.TypeFor[ui.ColorSwatchWidget](),
		"ColorSwatchPicker": reflect.TypeFor[ui.ColorSwatchPickerWidget](),
		"CloseButton":       reflect.TypeFor[ui.CloseButtonWidget](),
		"Collapsible":       reflect.TypeFor[ui.CollapsibleWidget](),
		"CollapsibleGroup":  reflect.TypeFor[ui.CollapsibleGroupWidget](),
		"ComboBox":          reflect.TypeFor[ui.ComboBoxWidget](),
		"ContextMenu":       reflect.TypeFor[ui.ContextMenuWidget](),
		"DateField":         reflect.TypeFor[ui.DateFieldWidget](),
		"DatePicker":        reflect.TypeFor[ui.DatePickerWidget](),
		"DateRangePicker":   reflect.TypeFor[ui.DateRangePickerWidget](),
		"Description":       reflect.TypeFor[ui.DescriptionWidget](),
		"DockLayout":        reflect.TypeFor[ui.DockLayoutWidget](),
		"Dropdown":          reflect.TypeFor[ui.DropdownWidget](),
		"Input":             reflect.TypeFor[ui.InputWidget](),
		"InputGroup":        reflect.TypeFor[ui.InputGroupWidget](),
		"Label":             reflect.TypeFor[ui.LabelWidget](),
		"LineChart":         reflect.TypeFor[ui.LineChartWidget](),
		"ListBox":           reflect.TypeFor[ui.ListBoxWidget](),
		"Menu":              reflect.TypeFor[ui.MenuWidget](),
		"Menubar":           reflect.TypeFor[ui.MenubarWidget](),
		"Meter":             reflect.TypeFor[ui.MeterWidget](),
		"Modal":             reflect.TypeFor[ui.ModalWidget](),
		"Pagination":        reflect.TypeFor[ui.PaginationWidget](),
		"PanelHost":         reflect.TypeFor[ui.PanelHostWidget](),
		"PieChart":          reflect.TypeFor[ui.PieChartWidget](),
		"Popover":           reflect.TypeFor[ui.PopoverWidget](),
		"ProgressBar":       reflect.TypeFor[ui.ProgressBarWidget](),
		"ProgressCircle":    reflect.TypeFor[ui.ProgressCircleWidget](),
		"RadioGroup":        reflect.TypeFor[ui.RadioGroupWidget](),
		"Scrollbar":         reflect.TypeFor[ui.ScrollbarWidget](),
		"Select":            reflect.TypeFor[ui.SelectWidget](),
		"Sidebar":           reflect.TypeFor[ui.SidebarWidget](),
		"Slider":            reflect.TypeFor[ui.SliderWidget](),
		"Spinner":           reflect.TypeFor[ui.SpinnerWidget](),
		"SplitPane":         reflect.TypeFor[ui.SplitPaneWidget](),
		"Surface":           reflect.TypeFor[ui.SurfaceWidget](),
		"Switch":            reflect.TypeFor[ui.SwitchWidget](),
		"SwitchGroup":       reflect.TypeFor[ui.SwitchGroupWidget](),
		"Table":             reflect.TypeFor[ui.TableWidget](),
		"Tabs":              reflect.TypeFor[ui.TabsWidget](),
		"Text":              reflect.TypeFor[ui.TextWidget](),
		"TextArea":          reflect.TypeFor[ui.TextAreaWidget](),
		"ToastProvider":     reflect.TypeFor[ui.ToastProviderWidget](),
		"ToggleButton":      reflect.TypeFor[ui.ToggleButtonWidget](),
		"Toolbar":           reflect.TypeFor[ui.ToolbarWidget](),
		"ToolbarSeparator":  reflect.TypeFor[ui.ToolbarSeparatorWidget](),
		"TitleBar":          reflect.TypeFor[ui.WindowTitleBarWidget](),
		"Tooltip":           reflect.TypeFor[ui.TooltipWidget](),
		"Tree":              reflect.TypeFor[ui.TreeWidget](),
	}
	for name, widgetType := range styleBacked {
		method, ok := widgetType.MethodByName("Style")
		if !ok {
			t.Errorf("%sWidget has no Style method", name)
			continue
		}
		if method.Type.NumIn() != 2 || method.Type.In(1) != styleType {
			t.Errorf("%sWidget Style signature = %v, want Style", name, method.Type)
		}
		if _, ok := widgetType.MethodByName("Theme"); ok {
			t.Errorf("%sWidget still exposes legacy Theme method", name)
		}
	}
}

func TestDropdownPublicItemKinds(t *testing.T) {
	var kind ui.DropdownItemKind = ui.DropdownItemCheckbox
	if kind != ui.MenuItemCheckbox {
		t.Fatalf("dropdown checkbox kind = %v, want menu checkbox kind", kind)
	}
	if item := ui.DropdownGroupLabel("More"); item.Kind != ui.DropdownItemGroupLabel || item.Label != "More" {
		t.Fatalf("dropdown group label = %#v", item)
	}
}

func TestContextExposesOnlySupportedMethods(t *testing.T) {
	contextType := reflect.TypeFor[*ui.Context]()
	want := map[string]struct{}{
		"BackgroundColor":     {},
		"BoolState":           {},
		"Clickable":           {},
		"Draggable":           {},
		"Editor":              {},
		"FocusVisible":        {},
		"ForegroundColor":     {},
		"Invalidate":          {},
		"Language":            {},
		"ListState":           {},
		"PreserveFocus":       {},
		"RequestFocus":        {},
		"RequestFocusVisible": {},
		"ScrollState":         {},
		"Theme":               {},
		"WindowState":         {},
	}
	if contextType.NumMethod() != len(want) {
		t.Errorf("ui.Context method count = %d, want %d", contextType.NumMethod(), len(want))
	}
	for method := range contextType.Methods() {
		name := method.Name
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
	if col, ok := ui.ResolveColor(ctx, ui.WithAlpha(ui.TokenAccent, .5)); ok {
		_ = col
	}
	if brush, ok := ui.ResolveBrush(ctx, ui.LinearGradient(
		ui.ColorStop(0, ui.TokenAccent),
		ui.ColorStop(1, ui.TokenDanger),
	)); ok {
		bounds := image.Rectangle{Max: image.Pt(20, 20)}
		ui.DrawBrush(gtx, bounds, 4, brush)
		ui.DrawBrushRRect(gtx, clip.UniformRRect(bounds, 4), brush)
	}
	_ = ui.MeasureText(ctx, gtx, ui.Text("Measure me").Size(14).MaxLines(1))
	var pointerTag struct{}
	ui.AddPointerArea(gtx, &pointerTag, image.Rectangle{Max: image.Pt(20, 20)}, pointer.CursorPointer)
	if eventValue, ok := ui.NextPointerEvent(gtx, &pointerTag, pointer.Press|pointer.Drag|pointer.Release|pointer.Cancel); ok {
		if ui.IsPrimaryPointerPress(eventValue) {
			ui.GrabPointer(gtx, &pointerTag, eventValue)
		}
	}
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

func facadeUpdate(model *facadeModel, msg facadeMsg) ui.Cmd[facadeMsg] {
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
	return nil
}

func facadeView(ctx *ui.Context, model facadeModel, send ui.Send[facadeMsg]) ui.Widget {
	if ctx != nil {
		var _ ui.Theme = ctx.Theme()
		var _ ui.Language = ctx.Language()
	}
	items := []ui.SelectItem{{Key: "one", Label: "One"}}
	treeItems := []ui.TreeItem{{
		Key: "folder", Label: "Folder",
		Leading: ui.Icon(lucide.Folder), ExpandedLeading: ui.Icon(lucide.FolderOpen),
		AcceptsChildren: true, Renamable: true, ChildrenState: ui.TreeChildrenLoaded,
		Children: []ui.TreeItem{{Key: "file", Label: "File"}},
	}}
	tableColumns := []ui.TableColumn{
		{Key: "name", Label: "Name", Sortable: true, Resizable: true, RowHeader: true, MinWidth: 120, MaxWidth: 320, Weight: 2},
		{Key: "status", Label: "Status", Width: 120, Align: ui.TableAlignEnd},
	}
	tableRows := []ui.TableRow{{
		Key: "member", Label: "Member",
		Cells: []ui.TableCell{{Text: "Member"}, {Content: ui.Input("member-status", "Active"), Interactive: true}},
	}}
	tabs := []ui.TabItem{{Key: "general", Label: "General", Icon: lucide.Settings, Content: ui.Text("General"), Panel: ui.Text("Panel")}}
	saveCommand := ui.NewCommand("save-command", "Save").
		Icon(ui.Icon(lucide.Save).Size(16)).
		Shortcut(ui.KeyShortcut("S", ui.ShortcutPrimary)).
		OnExecute(func() {})
	boldCommand := ui.NewCommand("bold-command", "Bold").
		Icon(ui.Icon(lucide.Bold).Size(16)).
		Shortcut(ui.KeyShortcut("B", ui.ShortcutPrimary|ui.ShortcutShift)).
		Toggle(model.open).
		OnExecute(func() {})

	return ui.Column(
		ui.Text("Formatted text").
			Font(font.Font{Typeface: "serif", Style: font.Italic, Weight: font.Medium}).
			Align(ui.TextAlignCenter).
			MaxLines(2).
			Truncator("...").
			Wrap(ui.TextWrapWords).
			LineHeight(20).
			LineHeightScale(1.1),
		ui.SelectableText("selectable-text", "Selectable text").
			Typeface("monospace").
			FontStyle(font.Italic).
			Weight(font.SemiBold).
			Align(ui.TextAlignEnd).
			MaxLines(1).
			Wrap(ui.TextWrapGraphemes),
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
				Stack("traffic").
				StackStrategy(ui.LineStackSameSign).
				StackOrder(ui.LineStackSeriesDescending).
				Width(2).
				Hidden(false),
			ui.LineXYSeries("latency", "Latency", []ui.LineChartPoint{{X: 0, Y: 8}, {X: 1, Y: 11}}).
				Smoothness(0.35),
		}).
			DataVersion(1).
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
			DataVersion(1).
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
		ui.PieChart("sources", []ui.PieChartData{
			ui.PieData("search", "Search", 1048).
				Color(color.NRGBA{R: 1, A: 0xff}).
				Hidden(false),
		}).
			DataVersion(1).
			Height(320).
			InnerRadius(.35).
			OuterRadius(.7).
			Clockwise(true).
			StartAngle(90).
			PadAngle(2).
			MinAngle(1).
			RoseType(ui.PieRoseRadius).
			StillShowZeroSum(true).
			Labels(true).
			Legend(true).
			Tooltip(true).
			Animation(true).
			AnimationDuration(time.Second).
			AnimationEasing(ui.EaseCubicInOut).
			UpdateAnimationDuration(500*time.Millisecond).
			UpdateAnimationEasing(ui.EaseCubicInOut).
			OnLegendChange(func(string, bool) {}).
			OnDataClick(func(ui.ChartSelection) {}).
			TooltipContent(func(selection ui.ChartSelection) ui.Widget {
				_ = selection.Items[0].Percent
				return ui.Text(selection.Label)
			}).
			Label("Traffic sources").
			EmptyText("No sources").
			Disabled(false),
		ui.CandlestickChart("market", []ui.CandlestickChartData{
			ui.Candle(100, 105, 98, 108),
		}).
			DataVersion(1).
			Times([]time.Time{time.Unix(1, 0)}).
			FormatTime(func(value time.Time) string { return value.Format("15:04") }).
			Height(320).
			Grid(true).
			Tooltip(true).
			Crosshair(true).
			YRange(90, 110).
			YTicks(5).
			FormatY(func(value float64) string { return fmt.Sprint(value) }).
			XAxis("Date").
			YAxis("Price").
			Width(10).
			MaxWidth(18).
			MinWidth(2).
			UpColor(color.NRGBA{R: 1, A: 0xff}).
			DownColor(color.NRGBA{G: 1, A: 0xff}).
			DojiColor(color.NRGBA{B: 1, A: 0xff}).
			Animation(true).
			AnimationDuration(time.Second).
			AnimationEasing(ui.EaseCubicOut).
			UpdateAnimationDuration(500*time.Millisecond).
			UpdateAnimationEasing(ui.EaseCubicInOut).
			OnDataClick(func(selection ui.ChartSelection) {
				_ = selection.Items[0].Open
				_ = selection.Items[0].Close
				_ = selection.Items[0].Low
				_ = selection.Items[0].High
			}).
			TooltipContent(func(selection ui.ChartSelection) ui.Widget { return ui.Text(selection.Label) }).
			DataWindow(0.1, 0.9).
			OnDataWindowChange(func(ui.ChartDataWindow) {}).
			MarkLines([]ui.ChartMarkLine{ui.MarkLine(ui.ChartAxisY, 100).Text("Target")}).
			MarkAreas([]ui.ChartMarkArea{ui.MarkArea(ui.ChartAxisX, 0, 1).Text("Session")}).
			MarkPoints([]ui.ChartMarkPoint{
				ui.MarkPoint(0, 105).Text("Buy").Content(ui.Icon(lucide.ArrowUp).Size(14)),
			}).
			Label("Market").
			EmptyText("No candles").
			Disabled(false),
		ui.Menubar("application-menu", []ui.MenubarItem{
			ui.MenubarMenu("file", "File", []ui.MenuItem{{Key: "new", Label: "New"}}).
				OnActionEvent(func(ui.MenuActionEvent) {}).
				Width(220),
			ui.MenubarMenuContent(
				"edit",
				"Edit",
				ui.Menu("edit-menu", []ui.MenuItem{{Key: "copy", Label: "Copy"}}),
			).Disabled(false),
		}).
			Alt("Application menu").
			Compact(true).
			Orientation(ui.MenubarHorizontal).
			LoopFocus(true).
			Modal(true).
			DefaultOpenKey("").
			OpenKey("").
			OnOpenChange(func(string) {}).
			Disabled(false),
		ui.WindowTitleBar(
			"workspace-title-bar",
			"FlowUI Workspace",
			ui.Menubar("workspace-title-menu", []ui.MenubarItem{
				ui.MenubarMenu("workspace-file", "File", []ui.MenuItem{{Key: "open", Label: "Open"}}),
			}),
		).
			Leading(ui.Icon(lucide.AppWindow).Size(16)).
			Menu(ui.Text("Menu")).
			Center(ui.Input("title-search", "").Placeholder("Search")).
			Trailing(ui.Button("title-settings", ui.Icon(lucide.Settings).Size(16))).
			ShowMinimize(true).
			ShowMaximize(true).
			ShowClose(true).
			OnClose(func() {}),
		ui.Toolbar(
			ui.Button("save", ui.Icon(lucide.Save).Size(16)).
				Style(ui.Radius(8).
					PaddingX(20).
					Background(ui.RGB(0x177245)),
				).
				IconOnly(),
			ui.ToolbarSeparator(),
			ui.ToggleButton("bold", true, ui.Icon(lucide.Bold).Size(16)).IconOnly(),
		).
			Orientation(ui.ToolbarHorizontal).
			Attached(true).
			LoopFocus(true).
			Alt("Editor tools").
			Disabled(false),
		ui.ButtonGroup(
			ui.Button("group-first", ui.Text("First")),
			ui.Button("group-second", ui.Text("Second")),
		).
			Variant(ui.ButtonSecondary).
			Size(ui.ButtonSmall).
			Style(ui.Radius(6)).
			Orientation(ui.ButtonGroupHorizontal).
			Separators(true).
			FullWidth().
			Disabled(false),
		ui.CommandScope(
			[]ui.Command{saveCommand, boldCommand},
			ui.Toolbar(
				ui.CommandButton("command-save", saveCommand),
				ui.CommandButton("command-bold", boldCommand),
			),
		).DisableWhenFieldFocused(),
		ui.Menu("command-menu", []ui.MenuItem{
			ui.CommandMenuItem(saveCommand),
			ui.CommandMenuItem(boldCommand),
		}),
		ui.Dropdown(
			"command-dropdown",
			ui.Button("command-dropdown-trigger", ui.Text("Commands")),
			[]ui.DropdownItem{ui.CommandMenuItem(saveCommand)},
		),
		ui.Alert("Update available", "Refresh to get the latest features.").
			Status(ui.AlertAccent).
			Indicator(ui.Icon(lucide.Info).Size(16)).
			Content(ui.Text("Custom alert content")).
			Action(ui.CloseButton("dismiss-alert")),
		ui.Card(
			ui.Column(
				ui.Text("Settings"),
				ui.Text("Public facade contract"),
			),
			externalWidget{},
			ui.Row(ui.Text("Footer")),
		).Variant(ui.CardSecondary),
		ui.Surface(externalWidget{}).Style(ui.BorderWidth(1).
			BorderColor(ui.Color(color.NRGBA{A: 0xff})),
		),
		ui.Surface(
			ui.Tabs("settings", "general", tabs).
				Variant(ui.TabsSecondary).
				Size(ui.TabsLarge).
				Leading(ui.Icon(lucide.LayoutDashboard).Size(16)).
				Trailing(ui.Icon(lucide.Plus).Size(16)).
				KeepAlive(true).
				ForceRender(true).
				DestroyOnHidden(false).
				Animated(true).
				AnimationDuration(time.Millisecond).
				IndicatorAnimationDuration(2*time.Millisecond).
				PanelTransition(ui.TabsPanelFade).
				PanelAnimationDuration(3*time.Millisecond).
				Activation(ui.TabsActivationManual).
				IndicatorWidth(unit.Dp(24)).
				IndicatorAlign(ui.TabsIndicatorCenter).
				OverflowTrigger(ui.Icon(lucide.Ellipsis).Size(16)).
				MoreLabel("More settings").
				Centered(true).
				Placement(ui.TabsEnd).
				OnClose(func(string) {}),
		).Variant(ui.SurfaceSecondary).Style(ui.Radius(8)),
		ui.Collapsible("details", model.open, "Details", ui.Text("More information")).
			Leading(ui.Icon(lucide.Info).Size(16)).
			Trailing(ui.Chip("New").Size(ui.ChipSmall)).
			Disabled(false).
			OnExpandedChange(func(expanded bool) { send(facadeMsg{open: &expanded}) }),
		ui.CollapsibleGroup("sections", model.expanded, []ui.CollapsibleItem{
			{Key: "general", Label: "General", Content: ui.Text("General settings")},
			{Key: "advanced", Label: "Advanced", Content: ui.Text("Advanced settings"), Disabled: true},
		}).
			AllowMultipleExpanded(true).
			Disabled(false).
			OnExpandedChange(func(keys []string) { send(facadeMsg{expanded: keys}) }),
		ui.Select("choice", model.selected, items).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }),
		ui.Tree("files", model.selected, treeItems).
			DataVersion(1).
			SelectedKeys([]string{"file"}).
			ExpandedKeys(model.expanded).
			ContextMenu(ui.Menu("tree-actions", []ui.MenuItem{{Key: "open", Label: "Open"}})).
			OnContextMenu(func(string) {}).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }).
			OnExpandedChange(func(keys []string) { send(facadeMsg{expanded: keys}) }).
			OnAction(func(string) {}).
			OnDrop(func(ui.TreeDropEvent) {}).
			CanDrop(func(ui.TreeDropEvent) bool { return true }).
			OnDragStart(func(ui.TreeDragEvent) {}).
			OnDragEnter(func(ui.TreeDragEvent) {}).
			OnDragLeave(func(ui.TreeDragEvent) {}).
			OnDragOver(func(ui.TreeDragEvent) {}).
			OnDragEnd(func(ui.TreeDragEvent) {}).
			OnLoadChildren(func(string) {}).
			OnLoadChildrenEvent(func(ui.TreeLoadEvent) {}).
			FilterFunc(func(ui.TreeItem) bool { return true }).
			FilterVersion(1).
			OnRename(func(string, string) {}).
			RequestRename("folder", 1).
			Variant(ui.TreeSurface).
			Size(ui.TreeSmall).
			SelectionMode(ui.TreeSelectionMultiple).
			OnSelectionChange(func([]string) {}).
			DisabledKeys([]string{"file"}).
			AllowEmptySelection().
			Guides(true).
			GuideConnectors(true).
			GuideStyle(ui.TreeGuideDashed).
			ExpandOnRowClick(true).
			EmptyText("No files").
			MaxHeight(240),
		ui.SidebarSections("primary-navigation", model.selected, []ui.SidebarSection{
			{Title: "Workspace", Items: []ui.SidebarItem{
				{Key: "overview", Label: "Overview", Leading: ui.Icon(lucide.LayoutDashboard).Size(18)},
				{Key: "projects", Label: "Projects", Leading: ui.Icon(lucide.FolderKanban).Size(18), Trailing: ui.Text("8")},
			}},
			{Title: "Account", Items: []ui.SidebarItem{
				{Key: "settings", Label: "Settings", Leading: ui.Icon(lucide.Settings).Size(18)},
			}},
		}).
			Header(ui.Text("FlowUI")).
			Footer(ui.Text("Signed in")).
			Collapsed(false).
			Width(248).
			CollapsedWidth(64).
			Alt("Primary navigation").
			EmptyText("No destinations").
			DisabledKeys([]string{"settings"}).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }).
			OnAction(func(string) {}).
			Disabled(false),
		ui.Scrollbar("facade-scrollbar", ui.Spacer(640, 480)).
			Horizontal().
			AlignEnd().
			StickToEnd().
			ScrollAnyAxis().
			Overlay().
			Disabled(false),
		ui.SplitPane("facade-split-pane", ui.Text("Primary"), ui.Text("Secondary")).
			Vertical().
			DefaultRatio(.6).
			MinFirst(120).
			MinSecond(80).
			Label("Resize panes").
			OnRatioChange(func(float32) {}).
			Disabled(false),
		ui.PanelHost("facade-panel-host", "editor", []ui.PanelItem{
			{Key: "editor", Content: ui.Text("Editor")},
			{Key: "output", Content: ui.Text("Output")},
		}).
			KeepAlive(true).
			ForceRender(true).
			DestroyOnHidden(false),
		ui.DockLayout("facade-dock", ui.DockSplit("facade-root", ui.DockHorizontal,
			ui.DockPanel("facade-left", ui.Text("Left")),
			ui.DockPanel("facade-right", ui.Text("Right")),
		).Ratio(.3)).
			DefaultSnapshot(ui.DockLayoutSnapshot{Ratios: map[string]float32{"facade-root": .3}}).
			MaximizedKey("").
			KeepAlive(true),
		ui.StatusBar(ui.Text("Ready"), ui.Text("Ln 1, Col 1")),
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
			HeaderHeight(36).
			RowHeight(44).
			GridLines(true).
			Border(true).
			LoadMore(true, false, func() {}).
			LoadMoreContent(ui.Spinner()).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }).
			OnSelectionChange(func(keys []string) { send(facadeMsg{tableSelected: keys}) }).
			OnSortChange(func(sort ui.TableSortDescriptor) { send(facadeMsg{tableSort: &sort}) }).
			OnAction(func(string) {}).
			OnColumnResize(func(string, int) {}).
			RowContextMenu(func(ui.TableRow) ui.MenuWidget {
				return ui.Menu("row-actions", []ui.MenuItem{{Key: "open", Label: "Open"}})
			}).
			Disabled(false),
		ui.VirtualTable("virtual-members", tableColumns, 1000, func(int) ui.TableRow { return tableRows[0] }).
			MaxHeight(280),
		ui.Pagination("members-pages", 2, 12).
			Size(ui.PaginationSmall).
			Summary(ui.Text("11 to 20 of 120 results")).
			Siblings(1).
			Boundaries(1).
			ShowControls(true).
			Labels("Previous", "Next").
			OnChange(func(int) {}).
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
				OnActionEvent(func(ui.MenuActionEvent) {}).
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
			OnActionEvent(func(ui.DropdownActionEvent) {}).
			OnChange(func(key string) { send(facadeMsg{selected: key}) }).
			OnSelectionChange(func(keys []string) { send(facadeMsg{tableSelected: keys}) }).
			Open(model.open).
			DefaultOpen(false).
			OnOpenChangeEvent(func(event ui.DropdownOpenChangeEvent) { send(facadeMsg{open: &event.Open}) }).
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
		ui.InputGroup(
			ui.Input("search-action", "flowui"),
		).
			Style(ui.Part(ui.PartContent, ui.Cursor(ui.CursorText))).
			SuffixAction(
				ui.InputGroupAction("clear-search-action", "Clear search", ui.Icon(lucide.X).Size(16)).
					OnClick(func() {}),
			).
			FocusOnActionPress(false).
			FullWidth(),
		ui.ProgressBar("progress", 50).ShowValue(),
		ui.ProgressCircle("progress-circle", 60).
			Label("Upload").
			ValueText("60 percent").
			Range(0, 100).
			Indeterminate().
			Color(ui.ProgressCircleSuccess).
			Size(ui.ProgressCircleLarge).
			Disabled(false),
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
		ui.ColorPicker("brand-color", color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}).
			Label("Brand color").
			Alpha(true).
			ShowField().
			Presets([]color.NRGBA{{R: 0xf4, G: 0x3f, B: 0x5e, A: 255}}).
			Disabled(false).
			OnChange(func(color.NRGBA) {}),
		ui.DateField("due-date", time.Time{}).
			Label("Due date").
			Description("Choose a due date").
			Locale(ui.DatePickerEnglish()).
			OnChange(func(time.Time) {}),
		ui.DatePicker("appointment", time.Time{}).
			Label("Appointment").
			Description("Choose an appointment date").
			ErrorMessage("Invalid date").
			Required(true).
			OnChange(func(time.Time) {}),
		ui.DateRangePicker("trip-dates", ui.DateRange{}).
			Label("Trip dates").
			Description("Choose a date range").
			OnChange(func(ui.DateRange) {}),
		ui.RangeSlider("price", 10, 80).Range(0, 100).Step(5),
		ui.Tooltip("save-help", ui.Button("save", ui.Text("Save")), ui.Text("Save changes")).
			Placement(ui.TooltipTop).
			Arrow(true).
			Delay(0),
		ui.ToastProvider("toasts", []ui.ToastItem{
			ui.Toast("saved", "Saved").Variant(ui.ToastSuccess).Description("Changes saved"),
		}).Offset(24).OnClose(func(string) {}),
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
	graph := ui.NodeGraph("workflow", ui.NodeGraphData{
		Nodes: []ui.NodeGraphNode{
			ui.NewNodeGraphNode("source", "Source", ui.NodeGraphPoint{}).
				Outputs(ui.NewNodeGraphPort("output", "Output")),
			ui.NewNodeGraphNode("target", "Target", ui.NodeGraphPoint{X: 240}).
				Inputs(ui.NewNodeGraphPort("input", "Input")),
		},
		Edges: []ui.NodeGraphEdge{
			ui.NewNodeGraphEdge("source-target", ui.NewNodeGraphEndpoint("source", "output"), ui.NewNodeGraphEndpoint("target", "input")),
		},
	}).Height(280).
		Viewport(ui.NodeGraphViewport{Zoom: 1}).
		DefaultViewport(ui.NodeGraphViewport{Zoom: 1}).
		FitView(true).
		FitViewPadding(.1).
		Grid(true).
		GridPattern(ui.NodeGraphGridDots).
		GridColor(color.NRGBA{R: 0x32, G: 0x78, B: 0xc8, A: 0xff}).
		GridOpacity(.45).
		GridSize(16).
		ZoomRange(.5, 2).
		SelectedKeys([]string{"source"}).
		SelectedEdgeKeys([]string{"source-target"}).
		SelectionMode(ui.NodeGraphSelectionMultiple).
		SelectionOnDrag(true).
		SelectionBoxMode(ui.NodeGraphSelectionBoxPartial).
		NodesDraggable(true).
		NodesSelectable(true).
		NodesConnectable(true).
		NodesDeletable(true).
		EdgesSelectable(true).
		EdgesDeletable(true).
		EdgesReconnectable(true).
		NodeDragThreshold(3).
		SnapToGrid(true).
		SnapGrid(16, 16).
		DropTypes("application/x-flowui-node", "text/plain").
		OnDrop(func(ui.NodeGraphDropEvent) {}).
		Disabled(false).
		OnViewportChange(func(ui.NodeGraphViewport) {}).
		OnCanvasClick(func(ui.NodeGraphCanvasEvent) {}).
		OnCanvasDoubleClick(func(ui.NodeGraphCanvasEvent) {}).
		OnCanvasContextMenu(func(ui.NodeGraphCanvasEvent) {}).
		OnNodeClick(func(ui.NodeGraphNodeEvent) {}).
		OnNodeDoubleClick(func(ui.NodeGraphNodeEvent) {}).
		OnNodeContextMenu(func(ui.NodeGraphNodeEvent) {}).
		OnNodeHover(func(ui.NodeGraphNodeEvent) {}).
		OnNodeLeave(func(ui.NodeGraphNodeEvent) {}).
		OnEdgeClick(func(ui.NodeGraphEdgeEvent) {}).
		OnEdgeDoubleClick(func(ui.NodeGraphEdgeEvent) {}).
		OnEdgeContextMenu(func(ui.NodeGraphEdgeEvent) {}).
		OnEdgeHover(func(ui.NodeGraphEdgeEvent) {}).
		OnEdgeLeave(func(ui.NodeGraphEdgeEvent) {}).
		OnNodesChange(func([]ui.NodeGraphNodeChange) {}).
		OnEdgesChange(func([]ui.NodeGraphEdgeChange) {}).
		OnReconnect(func(ui.NodeGraphEdge, ui.NodeGraphConnection) {}).
		IsValidConnection(func(ui.NodeGraphConnection) bool { return true }).
		OnConnect(func(ui.NodeGraphConnection) {})
	var _ ui.Widget = graph
	_ = ui.ApplyNodeGraphEdgeChanges(nil, []ui.NodeGraphEdgeChange{{Kind: ui.NodeGraphEdgeChangeRemove}})
	_ = ui.ReconnectNodeGraphEdge(ui.NodeGraphEdge{}, ui.NodeGraphConnection{}, nil)
	_ = ui.NodeGraphReconnectBoth
	_ = ui.WindowTitleBarSupported()
	workbench := ui.NewWorkbenchController(ui.NewWorkbenchState([]ui.WorkbenchGroup{{
		Key:  "editor",
		Tabs: []ui.WorkbenchTab{{Key: "main.go", Title: "main.go", Closable: true}},
	}}))
	_ = workbench.Commands()
	_ = workbench.CommandScope(ui.Text("Editor"))
	_ = workbench.BindTabs("editor", ui.Tabs("editor-tabs", "main.go", nil))
	_ = workbench.BindPanel("editor", ui.PanelHost("editor-panels", "main.go", nil))
	_ = workbench.BindDock(ui.DockLayout("editor-dock", ui.DockPanel("editor", ui.Text("Editor"))))
	encodedWorkbench, err := ui.MarshalWorkbenchSnapshot(workbench.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ui.UnmarshalWorkbenchSnapshot(encodedWorkbench); err != nil {
		t.Fatal(err)
	}
	encodedDock, err := ui.MarshalDockLayoutSnapshot(ui.DockLayoutSnapshot{Version: ui.DockSnapshotVersion})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ui.UnmarshalDockLayoutSnapshot(encodedDock); err != nil {
		t.Fatal(err)
	}
	program := ui.Program[facadeModel, facadeMsg]{
		Init:   func() (facadeModel, ui.Cmd[facadeMsg]) { return facadeModel{}, nil },
		Update: func(*facadeModel, facadeMsg) ui.Cmd[facadeMsg] { return nil },
		View:   facadeView,
		WindowStateMessage: func(ui.WindowState) facadeMsg {
			return facadeMsg{}
		},
	}
	var _ func(ui.Program[facadeModel, facadeMsg], ...ui.Option) = ui.Run[facadeModel, facadeMsg]
	_ = ui.NewProgram(facadeModel{}, facadeUpdate, facadeView)
	_ = ui.NewWindow("program", program)
	window := ui.NewWindow("secondary", ui.NewProgram(facadeModel{}, facadeUpdate, facadeView), ui.Title("Secondary"))
	application := ui.NewApplication()
	_ = application.Open
	_ = application.Close
	_ = application.RequestClose
	_ = application.CloseAll
	_ = application.Quit
	_ = application.SetKeepAlive
	_ = application.IsOpen
	_ = application.Configure
	_ = application.SetTheme
	_ = application.SetLanguage
	_ = application.Perform
	_ = application.WindowState
	if window.Key() != "secondary" {
		t.Fatalf("window key = %q", window.Key())
	}
	var _ ui.Widget = externalWidget{}
	var _ ui.Widget = ui.WidgetFunc(func(*ui.Context, layout.Context) layout.Dimensions { return layout.Dimensions{} })
	_ = ui.LayoutVisualOverflow(nil, layout.Context{}, nil, func(*ui.Context, layout.Context, image.Rectangle) {})
	_ = ui.LayoutVisualOutset(nil, layout.Context{}, nil, 1, 2, 3, 4)
	_ = ui.VisualOutset(1, 2, 3, 4)
	var _ ui.Widget = ui.Box(ui.Text("Save")).Key("save").Label("Save").Disabled(false).OnClick(func() {})
	var _ func(*ui.Context, string) *int = ui.UseState[int]
	var _ func(*ui.Context, string, func() *int) *int = ui.UseStateWith[int]
	var _ ui.PortalContent = func(image.Rectangle, bool) ui.Widget { return ui.Text("Portal") }
	var _ ui.PortalWidget = ui.Portal("external-portal", false, nil, nil).Layer(ui.PortalLayerPopup).Passive(false).Disabled(false)
	var _ func(*ui.Context, func() layout.Dimensions) (layout.Dimensions, ui.OverlayPlacement) = ui.TrackOverlayPlacement
	var _ ui.Update[facadeModel, facadeMsg] = facadeUpdate
	var _ ui.View[facadeModel, facadeMsg] = facadeView
	var _ ui.Cmd[facadeMsg] = ui.Do(func(ui.Send[facadeMsg]) {})
	var _ ui.Cmd[facadeMsg] = ui.DoContext(func(context.Context, ui.Send[facadeMsg]) error { return nil })
	var _ ui.Cmd[facadeMsg] = ui.LatestCmd("load", ui.Do(func(ui.Send[facadeMsg]) {}))
	var _ ui.Cmd[facadeMsg] = ui.CancelLatestCmd[facadeMsg]("load")
	var _ ui.Subscriptions[facadeModel, facadeMsg] = func(facadeModel) []ui.Subscription[facadeMsg] {
		return []ui.Subscription[facadeMsg]{
			ui.Subscribe("events", func(context.Context, ui.Send[facadeMsg]) error { return nil }),
		}
	}
	var _ ui.Option = ui.Title("FlowUI")
	var _ ui.WindowOption = ui.MinSize(320, 240)
	var _ ui.WindowOption = ui.MaxSize(1280, 960)
	var _ ui.WindowOption = ui.TopMost(true)
	var _ ui.WindowOption = ui.Decorated(false)
	var _ ui.WindowAction = ui.WindowActionMaximize
	var _ ui.WindowMode = ui.WindowModeFullscreen
	var _ ui.WindowState
	var _ ui.Option = ui.OnError(func(error) {})
	var _ ui.WindowCloseDecision = ui.WindowCloseProceed
	var _ ui.WindowCloseDecision = ui.WindowCloseCancel
	var _ ui.WindowCloseDecision = ui.WindowCloseKeepAlive
	var _ ui.Option = ui.OnWindowCloseRequest(func() ui.WindowCloseDecision { return ui.WindowCloseCancel })
	var _ ui.Option = ui.RetainModelOnClose()
	var _ error = ui.ErrEffectShutdownTimeout
	var _ ui.Option = ui.Locale(ui.LanguageEnglish)
	var _ ui.DatePickerLocale = ui.DatePickerEnglish()
	var _ ui.DatePart = ui.DatePartYear
	var _ = ui.DatePickerLocale{
		DateOrder:    [3]ui.DatePart{ui.DatePartMonth, ui.DatePartDay, ui.DatePartYear},
		DateLiterals: [4]string{"", "/", "/", ""},
	}
	var _ ui.DateFieldWidget = ui.DateField("date", time.Time{}).Required(true).FullWidth().MinDate(time.Time{}).MaxDate(time.Time{})
	var _ ui.DateRangePickerWidget = ui.DateRangePicker("range", ui.DateRange{}).Required(true).FullWidth().MinDate(time.Time{}).MaxDate(time.Time{})
	var _ ui.ColorPickerTheme
	var _ ui.ColorWheelTheme
	var _ ui.ButtonGroupTheme
	var _ ui.ColorAreaTheme
	var _ ui.ColorFieldTheme
	var _ ui.ColorSliderTheme
	var _ ui.ColorSwatchTheme
	var _ ui.ColorSwatchPickerTheme
	var _ ui.SurfaceTheme
	var _ ui.ShadowLayerTheme
	var _ ui.ShadowTheme
	var _ ui.ShadowsTheme
	var _ ui.ShadowsTheme = ui.DefaultShadows()
	var _ = ui.ThemeShadow(ui.ShadowTheme{}, color.NRGBA{}, 1)
	var _ ui.ColorWheelWidget = ui.ColorWheel("wheel", color.NRGBA{}).Size(180).Label("Color wheel").Disabled(false).OnChange(func(color.NRGBA) {})
	var _ ui.ColorAreaWidget = ui.ColorArea("area", color.NRGBA{}).ShowDots(true).Disabled(false).OnChange(func(color.NRGBA) {})
	var _ ui.ColorFieldWidget = ui.ColorField("field", color.NRGBA{}).Swatch(true).Alpha(true).Variant(ui.InputSecondary).FullWidth().OnChange(func(color.NRGBA) {})
	var _ ui.ColorSliderWidget = ui.ColorSlider("hue", color.NRGBA{}, ui.ColorChannelHue).HideLabel().ShowOutput(false).OnChange(func(color.NRGBA) {})
	var _ ui.ColorSwatchWidget = ui.ColorSwatch(color.NRGBA{}).Size(ui.ColorSwatchLarge).Shape(ui.ColorSwatchSquare).Alt("Color")
	var _ ui.ColorSwatchPickerWidget = ui.ColorSwatchPicker("swatches", color.NRGBA{}, nil).Size(ui.ColorSwatchSmall).Shape(ui.ColorSwatchCircle).Arrangement(ui.ColorSwatchPickerGrid).DisabledColors(nil).OnChange(func(color.NRGBA) {})
	var _ ui.SplitPaneTheme
	var _ ui.TitleBarTheme
	var _ ui.PaginationTheme
	var _ ui.ProgressCircleTheme
	var _ ui.NodeGraphTheme

	if root := facadeView(nil, facadeModel{}, func(facadeMsg) {}); root == nil {
		t.Fatal("public facade returned a nil widget tree")
	}
}
