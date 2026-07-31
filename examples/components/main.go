package main

import (
	"image/color"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

const (
	defaultPage      = "typography"
	catalogWindowKey = "components"
)

var catalogNavigationSections = navigationSections()

type Msg func(*Model)

type Model struct {
	Page           string
	Dark           bool
	ThemeAccent    color.NRGBA
	ThemeColorOpen bool

	InputValue          string
	EmailValue          string
	TextAreaValue       string
	Checked             bool
	SwitchOn            bool
	RadioValue          string
	SelectValue         string
	SelectValues        []string
	SectionSelectValue  string
	SectionSelectValues []string
	ComboValue          string
	ListValue           string
	ListValues          []string
	SectionListValue    string
	SectionListValues   []string
	DateValue           time.Time
	DatePickerOnly      time.Time
	DatePickerChinese   time.Time
	DatePickerLimited   time.Time
	DatePickerInvalid   time.Time
	DateField           time.Time
	DateRange           ui.DateRange
	ColorValue          color.NRGBA
	SliderValue         float64
	RangeLower          float64
	RangeUpper          float64

	TabValue         string
	PaginationPage   int
	SidebarValue     string
	SidebarCollapsed bool
	TreeValue        string
	TreeExpanded     []string
	TreeConnectors   bool
	TreeDashed       bool
	TreeDragValue    string
	TreeDragExpanded []string
	TreeDragItems    []ui.TreeItem
	TreeDropMessage  string
	CollapsibleOpen  bool
	ExpandedGroups   []string
	MotionExpanded   bool
	MotionForward    bool
	MotionPlaying    bool
	MotionRun        uint64
	MotionRectAlt    bool

	PopoverOpen     bool
	PortalOpen      bool
	ModalOpen       bool
	AboutOpen       bool
	AlertDialogOpen bool
	Toasts          []ui.ToastItem
	ToastSequence   int

	TableSelected []string
	TableSort     ui.TableSortDescriptor
	LastAction    string
}

func initialModel() Model {
	now := time.Now()
	defaultTheme := ui.DefaultTheme()
	return Model{
		Page:                defaultPage,
		ThemeAccent:         defaultTheme.Palette.Accent,
		InputValue:          "FlowUI",
		EmailValue:          "hello@flowui.dev",
		TextAreaValue:       "A compact component collection.",
		RadioValue:          "starter",
		SelectValue:         "design",
		SelectValues:        []string{"design", "engineering"},
		SectionSelectValue:  "product",
		SectionSelectValues: []string{"design", "quality"},
		ComboValue:          "tokyo",
		ListValue:           "alpha",
		ListValues:          []string{"alpha", "stable"},
		SectionListValue:    "beta",
		SectionListValues:   []string{"alpha", "lts"},
		DateValue:           now,
		DatePickerOnly:      now,
		DatePickerChinese:   now,
		DatePickerLimited:   now.AddDate(0, 0, 3),
		DateField:           now,
		DateRange:           ui.DateRange{Start: now, End: now.AddDate(0, 0, 5)},
		ColorValue:          color.NRGBA{R: 0x3d, G: 0x63, B: 0xdd, A: 0xff},
		SliderValue:         42,
		RangeLower:          24,
		RangeUpper:          72,
		TabValue:            "preview",
		PaginationPage:      3,
		SidebarValue:        "overview",
		TreeValue:           "button",
		TreeExpanded:        []string{"components"},
		TreeConnectors:      true,
		TreeDragItems:       append([]ui.TreeItem(nil), catalogDragTreeItems...),
		TreeDragExpanded:    []string{"workspace", "workspace-src"},
		TreeDropMessage:     "Drag a row to move it before, inside, or after another node.",
		MotionForward:       true,
		MotionPlaying:       true,
	}
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	if msg != nil {
		msg(model)
	}
	return nil
}

func catalogView(application *ui.Application) ui.View[Model, Msg] {
	return func(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
		page := model.Page
		if pageTitle(page) == "" {
			page = defaultPage
		}
		sidebar := ui.SidebarSections("component-catalog-navigation", page, catalogNavigationSections).
			DataVersion(1).
			Header(navigationHeader(model.SidebarCollapsed)).
			Footer(navigationFooter(model.SidebarCollapsed)).
			Width(252).
			CollapsedWidth(64).
			Collapsed(model.SidebarCollapsed).
			ItemHeight(32).
			Alt("Component categories").
			OnChange(func(key string) {
				send(func(model *Model) { model.Page = key })
			})

		content := ui.Scrollbar("component-page-"+page,
			ui.Box(componentPage(ctx, page, model, send)).Style(ui.FillWidth().MaxWidth(1080).Padding(28)),
		).Overlay()

		body := ui.Stack(
			ui.Stacked(ui.Row(sidebar, ui.Expanded(content))),
			ui.Overlay(globalOverlays(model, send)).Expanded(),
		)
		return ui.Column(
			catalogTitleBar(application, model, send),
			ui.Expanded(body),
		)
	}
}

func catalogTitleBar(application *ui.Application, model Model, send ui.Send[Msg]) ui.Widget {
	navigate := func(page string) {
		send(func(model *Model) { model.Page = page })
	}
	titlebarMenuStyle := catalogTitlebarMenuStyle(model)
	menu := ui.Menubar("catalog-window-menu", []ui.MenubarItem{
		ui.MenubarMenuContent("catalog-window-file", "File", ui.Menu("catalog-window-file:menu", []ui.MenuItem{
			{Key: "overview", Label: "Overview", Leading: ui.Icon(lucide.LayoutGrid).Size(16)},
			ui.MenuSeparator(),
			{Key: "close", Label: "Close", Shortcut: "Alt+F4", Leading: ui.Icon(lucide.X).Size(16)},
		}).Style(titlebarMenuStyle).Width(190).OnAction(func(key string) {
			if key == "close" {
				application.Close(catalogWindowKey)
				return
			}
			navigate(defaultPage)
		})),
		ui.MenubarMenuContent("catalog-window-go", "Go", ui.Menu("catalog-window-go:menu", []ui.MenuItem{
			{Key: "buttons", Label: "Buttons", Leading: ui.Icon(lucide.MousePointerClick).Size(16)},
			{Key: "text-fields", Label: "Text fields", Leading: ui.Icon(lucide.TextCursorInput).Size(16)},
			{Key: "motion-style", Label: "Motion & style", Leading: ui.Icon(lucide.RefreshCw).Size(16)},
			{Key: "colors", Label: "Colors", Leading: ui.Icon(lucide.Palette).Size(16)},
			{Key: "sidebar-tree", Label: "Sidebar & tree", Leading: ui.Icon(lucide.PanelLeft).Size(16)},
			{Key: "overlays", Label: "Overlays", Leading: ui.Icon(lucide.MessageSquareMore).Size(16)},
			{Key: "tables", Label: "Tables", Leading: ui.Icon(lucide.Table2).Size(16)},
			{Key: "charts", Label: "Charts", Leading: ui.Icon(lucide.ChartNoAxesCombined).Size(16)},
		}).Style(titlebarMenuStyle).Width(210).OnAction(navigate)),
		ui.MenubarMenuContent("catalog-window-help", "Help", ui.Menu("catalog-window-help:menu", []ui.MenuItem{
			{Key: "app-shell", Label: "Application shell", Leading: ui.Icon(lucide.AppWindow).Size(16)},
			{Key: "about", Label: "About FlowUI", Leading: ui.Icon(lucide.Info).Size(16)},
		}).Style(titlebarMenuStyle).Width(210).OnAction(func(key string) {
			if key == "about" {
				send(func(model *Model) { model.AboutOpen = true })
				return
			}
			navigate(key)
		})),
	}).Compact(true).Alt("Application menu")

	icon := lucide.Moon
	label := "Use dark theme"
	if model.Dark {
		icon = lucide.Sun
		label = "Use light theme"
	}
	themeToggle := ui.Button("catalog-theme-toggle", ui.Icon(icon).Size(16)).
		Variant(ui.ButtonGhost).
		Size(ui.ButtonSmall).
		IconOnly().
		Label(label).
		OnClick(func() {
			send(toggleCatalogTheme(application))
		})
	themeColor := catalogThemeColorPicker(application, model, send)

	navigationIcon := lucide.PanelLeft
	navigationLabel := "Collapse navigation"
	if model.SidebarCollapsed {
		navigationIcon = lucide.PanelRight
		navigationLabel = "Expand navigation"
	}
	navigationToggle := ui.Button("catalog-navigation-toggle", ui.Icon(navigationIcon).Size(16)).
		Variant(ui.ButtonGhost).
		Size(ui.ButtonSmall).
		IconOnly().
		Label(navigationLabel).
		OnClick(func() {
			send(func(model *Model) { model.SidebarCollapsed = !model.SidebarCollapsed })
		})

	return ui.WindowTitleBar("component-catalog-title-bar", "FlowUI Components", menu).
		Leading(navigationToggle).
		Trailing(ui.Row(themeColor, themeToggle).Gap(4).AlignMiddle())
}

func catalogTitlebarMenuStyle(model Model) ui.Style {
	activeTheme := catalogTheme(model)
	background := activeTheme.Palette.Overlay
	background.A = 0x99
	hover := activeTheme.Palette.DefaultHover
	hover.A = 0x99
	return ui.Background(ui.Color(background)).Part(
		ui.PartItem,
		ui.When(ui.Hovered, ui.Background(ui.Color(hover))),
	)
}

func toggleCatalogTheme(application *ui.Application) Msg {
	return func(model *Model) {
		model.Dark = !model.Dark
		application.SetTheme(catalogWindowKey, catalogTheme(*model))
	}
}

func catalogThemeColorPicker(application *ui.Application, model Model, send ui.Send[Msg]) ui.Widget {
	applyColor := func(value color.NRGBA) {
		send(func(model *Model) {
			model.ThemeAccent = value
			application.SetTheme(catalogWindowKey, catalogTheme(*model))
		})
	}
	trigger := ui.Button(
		"catalog-theme-color-trigger",
		ui.Icon(lucide.Palette).Size(16),
	).
		Variant(ui.ButtonGhost).
		Size(ui.ButtonSmall).
		IconOnly().
		Label("Adjust theme color").
		OnClick(func() {
			send(func(model *Model) { model.ThemeColorOpen = !model.ThemeColorOpen })
		})
	wheel := ui.ColorWheel("catalog-theme-color-wheel", model.ThemeAccent).
		Size(168).
		Label("Theme color").
		OnChange(applyColor)
	presets := ui.ColorSwatchPicker("catalog-theme-color-presets", model.ThemeAccent, catalogColors).
		Size(ui.ColorSwatchExtraSmall).
		OnChange(applyColor)
	return ui.Popover("catalog-theme-color", model.ThemeColorOpen, trigger,
		ui.Column(wheel, presets).Gap(10).AlignMiddle(),
	).
		Open(model.ThemeColorOpen).
		Heading("Theme color").
		Placement(ui.PopoverBottomEnd).
		Arrow(true).
		OnOpenChange(func(open bool) {
			send(func(model *Model) { model.ThemeColorOpen = open })
		})
}

func catalogTheme(model Model) ui.Theme {
	activeTheme := ui.DefaultTheme()
	if model.Dark {
		activeTheme = ui.DarkTheme()
	}
	applyCatalogAccent(&activeTheme, model.ThemeAccent)
	return activeTheme
}

func applyCatalogAccent(activeTheme *ui.Theme, accent color.NRGBA) {
	if activeTheme == nil {
		return
	}
	if accent.A == 0 {
		accent.A = 0xff
	}
	activeTheme.Palette.Accent = accent
	activeTheme.Palette.AccentHover = scaleCatalogColor(accent, 110, 100)
	activeTheme.Palette.AccentPressed = scaleCatalogColor(accent, 82, 100)
	activeTheme.Palette.AccentForeground = catalogAccentForeground(accent)
	activeTheme.Palette.AccentSoft = accent
	activeTheme.Palette.AccentSoft.A = 0x22
	activeTheme.Palette.AccentSoftHover = accent
	activeTheme.Palette.AccentSoftHover.A = 0x33
	activeTheme.Palette.Focus = accent
	activeTheme.Palette.Selection = accent
	activeTheme.Palette.Selection.A = 0x50
	if activeTheme.Palette.Background.R < 0x80 {
		activeTheme.Palette.AccentSoftForeground = scaleCatalogColor(accent, 135, 100)
	} else {
		activeTheme.Palette.AccentSoftForeground = scaleCatalogColor(accent, 72, 100)
	}
	activeTheme.Components.Menu.IndicatorColor = accent
	activeTheme.Components.Menu.FocusColor = accent
	activeTheme.Components.Dropdown.FocusColor = accent
	ui.SyncMaterialTheme(activeTheme)
}

func scaleCatalogColor(value color.NRGBA, numerator, denominator uint16) color.NRGBA {
	value.R = scaleCatalogChannel(value.R, numerator, denominator)
	value.G = scaleCatalogChannel(value.G, numerator, denominator)
	value.B = scaleCatalogChannel(value.B, numerator, denominator)
	return value
}

func scaleCatalogChannel(value uint8, numerator, denominator uint16) uint8 {
	scaled := (uint32(value)*uint32(numerator) + uint32(denominator)/2) / uint32(denominator)
	if scaled > 0xff {
		return 0xff
	}
	return uint8(scaled)
}

func catalogAccentForeground(value color.NRGBA) color.NRGBA {
	brightness := 299*int(value.R) + 587*int(value.G) + 114*int(value.B)
	if brightness > 150000 {
		return color.NRGBA{R: 0x2f, G: 0x2f, B: 0x36, A: 0xff}
	}
	return color.NRGBA{R: 0xfc, G: 0xfc, B: 0xfc, A: 0xff}
}

func navigationHeader(collapsed bool) ui.Widget {
	brandIcon := ui.Surface(ui.Center(ui.Icon(lucide.LayoutGrid).Size(18))).
		Style(ui.Radius(8).
			Background(ui.Color(color.NRGBA{R: 0x3d, G: 0x63, B: 0xdd, A: 0xff})).
			TextColor(ui.Color(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})))
	if collapsed {
		return ui.Center(brandIcon)
	}
	return ui.Row(
		brandIcon,
		ui.Column(
			ui.Text("FlowUI").Size(15).MaxLines(1),
			ui.Text("Components").Size(12).MaxLines(1),
		).Gap(1),
	).AlignMiddle().Gap(10)
}

func navigationFooter(collapsed bool) ui.Widget {
	if collapsed {
		return ui.Center(ui.Icon(lucide.Package).Size(16))
	}
	return ui.Row(
		ui.Icon(lucide.Package).Size(16),
		ui.Expanded(ui.Text("Component catalog").Size(12)),
		ui.Chip("105").Size(ui.ChipSmall).Variant(ui.ChipSoft),
	).AlignMiddle().Gap(8)
}

func navigationSections() []ui.SidebarSection {
	return []ui.SidebarSection{
		{Title: "Content", Items: []ui.SidebarItem{
			{Key: "typography", Label: "Typography & media", Leading: ui.Icon(lucide.Type).Size(17)},
			{Key: "surfaces", Label: "Surfaces & display", Leading: ui.Icon(lucide.Layers).Size(17)},
			{Key: "motion-style", Label: "Motion & style", Leading: ui.Icon(lucide.RefreshCw).Size(17)},
		}},
		{Title: "Actions", Items: []ui.SidebarItem{
			{Key: "buttons", Label: "Buttons", Leading: ui.Icon(lucide.MousePointerClick).Size(17)},
			{Key: "toolbars", Label: "Toolbars", Leading: ui.Icon(lucide.Wrench).Size(17)},
		}},
		{Title: "Forms", Items: []ui.SidebarItem{
			{Key: "text-fields", Label: "Text fields", Leading: ui.Icon(lucide.TextCursorInput).Size(17)},
			{Key: "selection", Label: "Selection controls", Leading: ui.Icon(lucide.ListChecks).Size(17)},
			{Key: "choice-fields", Label: "Choice fields", Leading: ui.Icon(lucide.ChevronsUpDown).Size(17)},
			{Key: "dates", Label: "Date controls", Leading: ui.Icon(lucide.CalendarDays).Size(17)},
			{Key: "colors", Label: "Color controls", Leading: ui.Icon(lucide.Palette).Size(17)},
			{Key: "sliders", Label: "Sliders", Leading: ui.Icon(lucide.SlidersHorizontal).Size(17)},
		}},
		{Title: "Navigation", Items: []ui.SidebarItem{
			{Key: "tabs-pagination", Label: "Tabs & pagination", Leading: ui.Icon(lucide.PanelsTopLeft).Size(17)},
			{Key: "sidebar-tree", Label: "Sidebar & tree", Leading: ui.Icon(lucide.PanelLeft).Size(17)},
			{Key: "menus", Label: "Menus", Leading: ui.Icon(lucide.Menu).Size(17)},
		}},
		{Title: "Feedback", Items: []ui.SidebarItem{
			{Key: "status", Label: "Status & progress", Leading: ui.Icon(lucide.CircleAlert).Size(17)},
			{Key: "disclosure", Label: "Disclosure", Leading: ui.Icon(lucide.ChevronsUpDown).Size(17)},
			{Key: "overlays", Label: "Overlays", Leading: ui.Icon(lucide.MessageSquareMore).Size(17)},
		}},
		{Title: "Data", Items: []ui.SidebarItem{
			{Key: "tables", Label: "Tables", Leading: ui.Icon(lucide.Table2).Size(17)},
			{Key: "charts", Label: "Charts", Leading: ui.Icon(lucide.ChartNoAxesCombined).Size(17)},
		}},
		{Title: "Layout", Items: []ui.SidebarItem{
			{Key: "layout", Label: "Layout primitives", Leading: ui.Icon(lucide.LayoutGrid).Size(17)},
			{Key: "scrolling", Label: "Scrolling", Leading: ui.Icon(lucide.ScrollText).Size(17)},
			{Key: "split-pane", Label: "Split pane", Leading: ui.Icon(lucide.Columns2).Size(17)},
		}},
		{Title: "Application", Items: []ui.SidebarItem{
			{Key: "app-shell", Label: "Application shell", Leading: ui.Icon(lucide.AppWindow).Size(17)},
		}},
	}
}

func globalOverlays(model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Column(
		ui.Modal("catalog-about", model.AboutOpen, "About FlowUI", ui.Column(
			ui.Text("FlowUI").Size(22),
			ui.Text("A Go UI framework and component library for cross-platform desktop applications, built on Gio."),
			ui.Text("github.com/qianniancn/flowui").Size(12),
		).Gap(10)).
			Icon(ui.Icon(lucide.Layers).Size(28)).
			Size(ui.ModalSmall).
			Footer(ui.Button("catalog-about-close", ui.Text("Close")).OnClick(func() {
				send(func(model *Model) { model.AboutOpen = false })
			})).
			OnOpenChange(func(open bool) { send(func(model *Model) { model.AboutOpen = open }) }),
		ui.Modal("catalog-modal", model.ModalOpen, "Modal", ui.Text("Modal content rendered from the component catalog.")).
			Footer(ui.Row(
				ui.Button("catalog-modal-cancel", ui.Text("Cancel")).Variant(ui.ButtonSecondary).OnClick(func() {
					send(func(model *Model) { model.ModalOpen = false })
				}),
				ui.Button("catalog-modal-confirm", ui.Text("Confirm")).OnClick(func() {
					send(func(model *Model) { model.ModalOpen = false })
				}),
			).Gap(8)).
			OnOpenChange(func(open bool) { send(func(model *Model) { model.ModalOpen = open }) }),
		ui.AlertDialog("catalog-alert-dialog", model.AlertDialogOpen, "Delete component preset?", "This action cannot be undone.").
			Status(ui.AlertDialogDanger).
			Footer(ui.Row(
				ui.Button("catalog-alert-cancel", ui.Text("Cancel")).Variant(ui.ButtonTertiary).OnClick(func() {
					send(func(model *Model) { model.AlertDialogOpen = false })
				}),
				ui.Button("catalog-alert-confirm", ui.Text("Delete")).Variant(ui.ButtonDanger).OnClick(func() {
					send(func(model *Model) { model.AlertDialogOpen = false })
				}),
			).Gap(8)).
			OnOpenChange(func(open bool) { send(func(model *Model) { model.AlertDialogOpen = open }) }),
		ui.ToastProvider("catalog-toasts", model.Toasts).
			Placement(ui.ToastBottomEnd).
			OnClose(func(key string) { send(removeToastMessage(key)) }).
			OnAction(func(key string) { send(removeToastMessage(key)) }),
	)
}

func removeToastMessage(key string) Msg {
	return func(model *Model) {
		items := model.Toasts[:0]
		for _, item := range model.Toasts {
			if item.Key() != key {
				items = append(items, item)
			}
		}
		model.Toasts = items
	}
}

func startPprof() {
	if os.Getenv("FLOWUI_PPROF") != "1" {
		return
	}
	go func() {
		const address = "127.0.0.1:6060"
		log.Printf("pprof: http://%s/debug/pprof/", address)
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Printf("pprof: %v", err)
		}
	}()
}

func main() {
	startPprof()
	application := ui.NewApplication()
	window := ui.NewWindow(catalogWindowKey, ui.Program[Model, Msg]{
		Init:   func() (Model, ui.Cmd[Msg]) { return initialModel(), nil },
		Update: Update,
		View:   catalogView(application),
	}, ui.Title("FlowUI Components"),
		ui.Size(1120, 760),
		ui.MinSize(880, 560),
		ui.WithTheme(catalogTheme(initialModel())),
		ui.Decorated(false),
		ui.CenterOnStart(),
	)
	application.Run(window)
}
