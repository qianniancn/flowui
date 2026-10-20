package main

import (
	"image/color"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/alertdialog"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/colorpicker"
	"github.com/qianniancn/flowui/widget/datepicker"
	iconpkg "github.com/qianniancn/flowui/widget/icon"
	menupkg "github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/menubar"
	"github.com/qianniancn/flowui/widget/modal"
	"github.com/qianniancn/flowui/widget/nodegraph"
	"github.com/qianniancn/flowui/widget/popover"
	sidebarpkg "github.com/qianniancn/flowui/widget/sidebar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/titlebar"
	"github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/widget/tree"
	"github.com/qianniancn/flowui/window"
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
	DateRange           datepicker.DateRange
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
	TreeDragItems    []tree.Item
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
	Toasts          []toast.Item
	ToastSequence   int

	TableSelected []string
	TableSort     table.SortDescriptor
	LastAction    string
	GraphViewport nodegraph.Viewport
	GraphNodes    []nodegraph.Node
	GraphEdges    []nodegraph.Edge
}

func initialModel() Model {
	now := time.Now()
	defaultTheme := base.Light()
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
		DateRange:           datepicker.DateRange{Start: now, End: now.AddDate(0, 0, 5)},
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
		TreeDragItems:       append([]tree.Item(nil), catalogDragTreeItems...),
		TreeDragExpanded:    []string{"workspace", "workspace-src"},
		TreeDropMessage:     "Drag a row to move it before, inside, or after another node.",
		MotionForward:       true,
		MotionPlaying:       true,
		GraphNodes:          catalogNodeGraphNodes(),
		GraphEdges:          catalogNodeGraphEdges(),
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
		sidebar := sidebarpkg.NewSections("component-catalog-navigation", page, catalogNavigationSections).
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

		content := layout.Scrollbar(
			layout.Box(componentPage(ctx, page, model, send)).Style(style.New().FillWidth().MaxWidth(1080).Padding(28)),
		).Key("component-page-" + page).Overlay()

		body := layout.Stack(
			layout.Stacked(layout.Row(sidebar, layout.Expanded(content))),
			layout.Overlay(globalOverlays(model, send)).Expanded(),
		)
		return layout.Column(
			catalogTitleBar(application, model, send),
			layout.Expanded(body),
		)
	}
}

func catalogTitleBar(application *ui.Application, model Model, send ui.Send[Msg]) ui.Widget {
	navigate := func(page string) {
		send(func(model *Model) { model.Page = page })
	}
	menu := menubar.New("catalog-window-menu", []menubar.Item{
		menubar.NewMenuContent("catalog-window-file", "File", menupkg.New("catalog-window-file:menu", []menupkg.Item{
			{Key: "overview", Label: "Overview", Leading: iconpkg.New(lucide.LayoutGrid).Size(16)},
			menupkg.Separator(),
			{Key: "close", Label: "Close", Shortcut: "Alt+F4", Leading: iconpkg.New(lucide.X).Size(16)},
		}).Width(190).OnActionEvent(func(event menupkg.ActionEvent) {
			if event.Key == "close" {
				application.Close(catalogWindowKey)
				return
			}
			navigate(defaultPage)
		})),
		menubar.NewMenuContent("catalog-window-go", "Go", menupkg.New("catalog-window-go:menu", []menupkg.Item{
			{Key: "buttons", Label: "Buttons", Leading: iconpkg.New(lucide.MousePointerClick).Size(16)},
			{Key: "text-fields", Label: "Text fields", Leading: iconpkg.New(lucide.TextCursorInput).Size(16)},
			{Key: "motion-style", Label: "Motion & style", Leading: iconpkg.New(lucide.RefreshCw).Size(16)},
			{Key: "colors", Label: "Colors", Leading: iconpkg.New(lucide.Palette).Size(16)},
			{Key: "sidebar-tree", Label: "Sidebar & tree", Leading: iconpkg.New(lucide.PanelLeft).Size(16)},
			{Key: "overlays", Label: "Overlays", Leading: iconpkg.New(lucide.MessageSquareMore).Size(16)},
			{Key: "tables", Label: "Tables", Leading: iconpkg.New(lucide.Table2).Size(16)},
			{Key: "charts", Label: "Charts", Leading: iconpkg.New(lucide.ChartNoAxesCombined).Size(16)},
		}).Width(210).OnActionEvent(func(event menupkg.ActionEvent) { navigate(event.Key) })),
		menubar.NewMenuContent("catalog-window-help", "Help", menupkg.New("catalog-window-help:menu", []menupkg.Item{
			{Key: "app-shell", Label: "Application shell", Leading: iconpkg.New(lucide.AppWindow).Size(16)},
			{Key: "about", Label: "About FlowUI", Leading: iconpkg.New(lucide.Info).Size(16)},
		}).Width(210).OnActionEvent(func(event menupkg.ActionEvent) {
			if event.Key == "about" {
				send(func(model *Model) { model.AboutOpen = true })
				return
			}
			navigate(event.Key)
		})),
	}).Compact(true).Alt("Application menu")

	icon := lucide.Moon
	label := "Use dark theme"
	if model.Dark {
		icon = lucide.Sun
		label = "Use light theme"
	}
	themeToggle := button.New("").Content(iconpkg.New(icon).Size(16)).
		Variant(button.Ghost).
		Size(button.Small).
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
	navigationToggle := button.New("").Content(iconpkg.New(navigationIcon).Size(16)).
		Variant(button.Ghost).
		Size(button.Small).
		IconOnly().
		Label(navigationLabel).
		OnClick(func() {
			send(func(model *Model) { model.SidebarCollapsed = !model.SidebarCollapsed })
		})

	return titlebar.New("component-catalog-title-bar", "FlowUI Components", menu).
		Leading(navigationToggle).
		Trailing(layout.Row(themeColor, themeToggle).Gap(4).AlignMiddle())
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
	trigger := button.New("").Content(iconpkg.New(lucide.Palette).Size(16)).
		Variant(button.Ghost).
		Size(button.Small).
		IconOnly().
		Label("Adjust theme color").
		OnClick(func() {
			send(func(model *Model) { model.ThemeColorOpen = !model.ThemeColorOpen })
		})
	wheel := colorpicker.ColorWheel("catalog-theme-color-wheel", model.ThemeAccent).
		Size(168).
		Label("Theme color").
		OnChange(applyColor)
	presets := colorpicker.ColorSwatchPicker("catalog-theme-color-presets", model.ThemeAccent, catalogColors).
		Size(colorpicker.ColorSwatchExtraSmall).
		OnChange(applyColor)
	return popover.New(trigger, layout.Column(wheel, presets).Gap(10).AlignMiddle()).Key("catalog-theme-color").Open(model.ThemeColorOpen).
		Open(model.ThemeColorOpen).
		Heading("Theme color").
		Placement(popover.BottomEnd).
		Arrow(true).
		OnOpenChange(func(open bool) {
			send(func(model *Model) { model.ThemeColorOpen = open })
		})
}

func catalogTheme(model Model) theme.Theme {
	activeTheme := base.Light()
	if model.Dark {
		activeTheme = base.Dark()
	}
	applyCatalogAccent(&activeTheme, model.ThemeAccent)
	return activeTheme
}

func applyCatalogAccent(activeTheme *theme.Theme, accent color.NRGBA) {
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
	theme.SyncMaterialTheme(activeTheme)
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
	brandIcon := (surface.New(layout.Center(iconpkg.New(lucide.LayoutGrid).Size(18)))).Style(style.New().Radius(8).
		Background(style.Color(color.NRGBA{R: 0x3d, G: 0x63, B: 0xdd, A: 0xff})).
		TextColor(style.Color(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})))
	if collapsed {
		return layout.Center(brandIcon)
	}
	return layout.Row(
		brandIcon,
		layout.Column(
			text.New("FlowUI").Size(15).MaxLines(1),
			text.New("Components").Size(12).MaxLines(1),
		).Gap(1),
	).AlignMiddle().Gap(10)
}

func navigationFooter(collapsed bool) ui.Widget {
	if collapsed {
		return layout.Center(iconpkg.New(lucide.Package).Size(16))
	}
	return layout.Row(
		iconpkg.New(lucide.Package).Size(16),
		layout.Expanded(text.New("Component catalog").Size(12)),
		chip.New("105").Size(chip.SizeSmall).Variant(chip.VariantSoft),
	).AlignMiddle().Gap(8)
}

func navigationSections() []sidebarpkg.Section {
	return []sidebarpkg.Section{
		{Title: "Content", Items: []sidebarpkg.Item{
			{Key: "typography", Label: "Typography & media", Leading: iconpkg.New(lucide.Type).Size(17)},
			{Key: "surfaces", Label: "Surfaces & display", Leading: iconpkg.New(lucide.Layers).Size(17)},
			{Key: "motion-style", Label: "Motion & style", Leading: iconpkg.New(lucide.RefreshCw).Size(17)},
		}},
		{Title: "Actions", Items: []sidebarpkg.Item{
			{Key: "buttons", Label: "Buttons", Leading: iconpkg.New(lucide.MousePointerClick).Size(17)},
			{Key: "toolbars", Label: "Toolbars", Leading: iconpkg.New(lucide.Wrench).Size(17)},
		}},
		{Title: "Forms", Items: []sidebarpkg.Item{
			{Key: "text-fields", Label: "Text fields", Leading: iconpkg.New(lucide.TextCursorInput).Size(17)},
			{Key: "selection", Label: "Selection controls", Leading: iconpkg.New(lucide.ListChecks).Size(17)},
			{Key: "choice-fields", Label: "Choice fields", Leading: iconpkg.New(lucide.ChevronsUpDown).Size(17)},
			{Key: "dates", Label: "Date controls", Leading: iconpkg.New(lucide.CalendarDays).Size(17)},
			{Key: "colors", Label: "Color controls", Leading: iconpkg.New(lucide.Palette).Size(17)},
			{Key: "sliders", Label: "Sliders", Leading: iconpkg.New(lucide.SlidersHorizontal).Size(17)},
		}},
		{Title: "Navigation", Items: []sidebarpkg.Item{
			{Key: "tabs-pagination", Label: "Tabs & pagination", Leading: iconpkg.New(lucide.PanelsTopLeft).Size(17)},
			{Key: "sidebar-tree", Label: "Sidebar & tree", Leading: iconpkg.New(lucide.PanelLeft).Size(17)},
			{Key: "menus", Label: "Menus", Leading: iconpkg.New(lucide.Menu).Size(17)},
		}},
		{Title: "Feedback", Items: []sidebarpkg.Item{
			{Key: "status", Label: "Status & progress", Leading: iconpkg.New(lucide.CircleAlert).Size(17)},
			{Key: "disclosure", Label: "Disclosure", Leading: iconpkg.New(lucide.ChevronsUpDown).Size(17)},
			{Key: "overlays", Label: "Overlays", Leading: iconpkg.New(lucide.MessageSquareMore).Size(17)},
		}},
		{Title: "Data", Items: []sidebarpkg.Item{
			{Key: "tables", Label: "Tables", Leading: iconpkg.New(lucide.Table2).Size(17)},
			{Key: "charts", Label: "Charts", Leading: iconpkg.New(lucide.ChartNoAxesCombined).Size(17)},
		}},
		{Title: "Layout", Items: []sidebarpkg.Item{
			{Key: "layout", Label: "Layout primitives", Leading: iconpkg.New(lucide.LayoutGrid).Size(17)},
			{Key: "scrolling", Label: "Scrolling", Leading: iconpkg.New(lucide.ScrollText).Size(17)},
			{Key: "split-pane", Label: "Split pane", Leading: iconpkg.New(lucide.Columns2).Size(17)},
		}},
		{Title: "Application", Items: []sidebarpkg.Item{
			{Key: "app-shell", Label: "Application shell", Leading: iconpkg.New(lucide.AppWindow).Size(17)},
		}},
	}
}

func globalOverlays(model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		modal.New("About FlowUI", layout.Column(
			text.New("FlowUI").Size(22),
			text.New("A Go UI framework and component library for cross-platform desktop applications, built on Gio."),
			text.New("github.com/qianniancn/flowui").Size(12),
		).Gap(10)).Key("catalog-about").Open(model.AboutOpen).
			Icon(iconpkg.New(lucide.Layers).Size(28)).
			Size(modal.Small).
			Footer(button.New("Close").OnClick(func() {
				send(func(model *Model) { model.AboutOpen = false })
			})).
			OnOpenChange(func(open bool) { send(func(model *Model) { model.AboutOpen = open }) }),
		modal.New("Modal", text.New("Modal content rendered from the component catalog.")).Key("catalog-modal").Open(model.ModalOpen).
			Footer(layout.Row(
				button.New("Cancel").Variant(button.Secondary).OnClick(func() {
					send(func(model *Model) { model.ModalOpen = false })
				}),
				button.New("Confirm").OnClick(func() {
					send(func(model *Model) { model.ModalOpen = false })
				}),
			).Gap(8)).
			OnOpenChange(func(open bool) { send(func(model *Model) { model.ModalOpen = open }) }),
		alertdialog.New("Delete component preset?", "This action cannot be undone.").Key("catalog-alert-dialog").Open(model.AlertDialogOpen).
			Status(alertdialog.StatusDanger).
			Footer(layout.Row(
				button.New("Cancel").Variant(button.Tertiary).OnClick(func() {
					send(func(model *Model) { model.AlertDialogOpen = false })
				}),
				button.New("Delete").Variant(button.Danger).OnClick(func() {
					send(func(model *Model) { model.AlertDialogOpen = false })
				}),
			).Gap(8)).
			OnOpenChange(func(open bool) { send(func(model *Model) { model.AlertDialogOpen = open }) }),
		toast.New("catalog-toasts", model.Toasts).
			Placement(toast.BottomEnd).
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
	}, window.Title("FlowUI Components"),
		window.Size(1120, 760),
		window.MinSize(880, 560),
		window.Theme(catalogTheme(initialModel())),
		window.Decorated(false),
		window.CenterOnStart(),
	)
	application.Run(window)
}
