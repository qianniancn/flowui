package main

import (
	"fmt"
	"time"

	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type CounterModel struct {
	Value int
}

type CounterMsg int

func counterUpdate(model *CounterModel, msg CounterMsg) {
	model.Value += int(msg)
}

func counterView(application *ui.Application) ui.View[CounterModel, CounterMsg] {
	return func(_ *ui.Context, model CounterModel, send ui.Send[CounterMsg]) ui.Widget {
		return ui.Center(
			ui.Column(
				ui.Text("Independent window").Size(22),
				ui.Text(fmt.Sprintf("Count: %d", model.Value)).Size(18),
				ui.Row(
					ui.Button("decrease", ui.Icon(lucide.Minus).Size(16)).OnClick(func() { send(-1) }),
					ui.Button("increase", ui.Icon(lucide.Plus).Size(16)).OnClick(func() { send(1) }),
					ui.Button("close", ui.Text("Close")).Variant(ui.ButtonSecondary).OnClick(func() { application.Close("counter") }),
				).Gap(8).AlignMiddle(),
			).Gap(16).AlignMiddle(),
		)
	}
}

type MainModel struct {
	Dark    bool
	Chinese bool
	Date    time.Time
}

type mainMsgKind uint8

const (
	setDark mainMsgKind = iota
	setChinese
	setDate
)

type MainMsg struct {
	kind    mainMsgKind
	enabled bool
	date    time.Time
}

func mainUpdate(application *ui.Application) ui.Update[MainModel, MainMsg] {
	return func(model *MainModel, msg MainMsg) {
		switch msg.kind {
		case setDark:
			model.Dark = msg.enabled
			activeTheme := ui.DefaultTheme()
			if model.Dark {
				activeTheme = ui.DarkTheme()
			}
			application.SetTheme("main", activeTheme)
		case setChinese:
			model.Chinese = msg.enabled
			language := ui.LanguageEnglish
			if model.Chinese {
				language = ui.LanguageChinese
			}
			application.SetLanguage("main", language)
		case setDate:
			model.Date = msg.date
		}
	}
}

func mainView(application *ui.Application, counter ui.WindowSpec) ui.View[MainModel, MainMsg] {
	return func(ctx *ui.Context, model MainModel, send ui.Send[MainMsg]) ui.Widget {
		state := ctx.WindowState()
		status := fmt.Sprintf(
			"%d x %d px | %s | focused: %t",
			state.Size.X,
			state.Size.Y,
			state.Mode.String(),
			state.Focused,
		)
		return ui.Center(
			ui.Box(
				ui.Column(
					ui.Text("FlowUI Multi-window").Size(24),
					ui.Text(status).Size(13),
					ui.Divider(),
					ui.Wrap(
						windowAction(application, "minimize", "Minimize", ui.WindowActionMinimize),
						windowAction(application, "maximize", "Maximize", ui.WindowActionMaximize),
						windowAction(application, "restore", "Restore", ui.WindowActionRestore),
						windowAction(application, "fullscreen", "Fullscreen", ui.WindowActionFullscreen),
						windowAction(application, "center", "Center", ui.WindowActionCenter),
						windowAction(application, "raise", "Raise", ui.WindowActionRaise),
					).Gap(8).LineGap(8),
					ui.Wrap(
						ui.Button("set-title", ui.Text("Set title")).
							Variant(ui.ButtonSecondary).
							OnClick(func() { application.Configure("main", ui.Title("FlowUI Workspace")) }),
						ui.Button("resize", ui.Text("Resize")).
							Variant(ui.ButtonSecondary).
							OnClick(func() { application.Configure("main", ui.Size(900, 620)) }),
						ui.ToggleButton("top-most", state.TopMost, ui.Text("Top most")).
							OnChange(func(enabled bool) { application.Configure("main", ui.TopMost(enabled)) }),
						ui.ToggleButton("decorated", state.Decorated, ui.Text("Decorated")).
							OnChange(func(enabled bool) { application.Configure("main", ui.Decorated(enabled)) }),
					).Gap(8).LineGap(8),
					ui.Wrap(
						ui.ToggleButton("dark-theme", model.Dark, ui.Text("Dark theme")).
							OnChange(func(enabled bool) { send(MainMsg{kind: setDark, enabled: enabled}) }),
						ui.ToggleButton("chinese-language", model.Chinese, ui.Text("Chinese locale")).
							OnChange(func(enabled bool) { send(MainMsg{kind: setChinese, enabled: enabled}) }),
						ui.DatePicker("locale-preview", model.Date).
							OnChange(func(value time.Time) { send(MainMsg{kind: setDate, date: value}) }),
					).Gap(8).LineGap(8),
					ui.Divider(),
					ui.Row(
						ui.Button("open-counter", ui.Text("Open counter")).OnClick(func() { application.Open(counter) }),
						ui.Button("close-all", ui.Text("Close all")).Variant(ui.ButtonSecondary).OnClick(application.CloseAll),
					).Gap(8).AlignMiddle(),
				).Gap(16),
			).Style(ui.FillWidth()).Style(ui.MaxWidth(720)).Style(ui.Padding(24)),
		)
	}
}

func windowAction(application *ui.Application, key, label string, action ui.WindowAction) ui.Widget {
	return ui.Button(key, ui.Text(label)).
		Variant(ui.ButtonSecondary).
		OnClick(func() { application.Perform("main", action) })
}

func main() {
	application := ui.NewApplication()
	counter := ui.NewWindow(
		"counter",
		func() CounterModel { return CounterModel{} },
		counterUpdate,
		counterView(application),
		ui.Title("Counter"),
		ui.Size(420, 260),
	)
	mainWindow := ui.NewWindow(
		"main",
		func() MainModel { return MainModel{} },
		mainUpdate(application),
		mainView(application, counter),
		ui.Title("Multi-window"),
		ui.Size(760, 520),
		ui.MinSize(520, 360),
		ui.MaxSize(1400, 1000),
	)
	application.Run(mainWindow)
}
