package main

import (
	"fmt"
	"time"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/togglebutton"
	"github.com/qianniancn/flowui/window"
)

type CounterModel struct {
	Value int
}

type CounterMsg int

func counterUpdate(model *CounterModel, msg CounterMsg) ui.Cmd[CounterMsg] {
	model.Value += int(msg)
	return nil
}

func counterView(application *ui.Application) ui.View[CounterModel, CounterMsg] {
	return func(_ *ui.Context, model CounterModel, send ui.Send[CounterMsg]) ui.Widget {
		return layout.Center(
			layout.Column(
				text.New("Independent window").Size(22),
				text.New(fmt.Sprintf("Count: %d", model.Value)).Size(18),
				layout.Row(
					button.New("").Content(icon.New(lucide.Minus).Size(16)).OnClick(func() { send(-1) }),
					button.New("").Content(icon.New(lucide.Plus).Size(16)).OnClick(func() { send(1) }),
					button.New("Close").Variant(button.Secondary).OnClick(func() { application.Close("counter") }),
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
	return func(model *MainModel, msg MainMsg) ui.Cmd[MainMsg] {
		switch msg.kind {
		case setDark:
			model.Dark = msg.enabled
			activeTheme := base.Light()
			if model.Dark {
				activeTheme = base.Dark()
			}
			application.SetTheme("main", activeTheme)
		case setChinese:
			model.Chinese = msg.enabled
			language := window.LanguageEnglish
			if model.Chinese {
				language = window.LanguageChinese
			}
			application.SetLanguage("main", language)
		case setDate:
			model.Date = msg.date
		}
		return nil
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
		return layout.Center(
			layout.Box(
				layout.Column(
					text.New("FlowUI Multi-window").Size(24),
					text.New(status).Size(13),
					layout.Divider(),
					layout.Wrap(
						windowAction(application, "minimize", "Minimize", window.ActionMinimize),
						windowAction(application, "maximize", "Maximize", window.ActionMaximize),
						windowAction(application, "restore", "Restore", window.ActionRestore),
						windowAction(application, "fullscreen", "Fullscreen", window.ActionFullscreen),
						windowAction(application, "center", "Center", window.ActionCenter),
						windowAction(application, "raise", "Raise", window.ActionRaise),
					).Gap(8).LineGap(8),
					layout.Wrap(
						button.New("Set title").
							Variant(button.Secondary).
							OnClick(func() { application.Configure("main", window.Title("FlowUI Workspace")) }),
						button.New("Resize").
							Variant(button.Secondary).
							OnClick(func() { application.Configure("main", window.Size(900, 620)) }),
						togglebutton.New("").Key("top-most").Selected(state.TopMost).Content(text.New("Top most")).
							OnChange(func(enabled bool) { application.Configure("main", window.TopMost(enabled)) }),
						togglebutton.New("").Key("decorated").Selected(state.Decorated).Content(text.New("Decorated")).
							OnChange(func(enabled bool) { application.Configure("main", window.Decorated(enabled)) }),
					).Gap(8).LineGap(8),
					layout.Wrap(
						togglebutton.New("").Key("dark-theme").Selected(model.Dark).Content(text.New("Dark theme")).
							OnChange(func(enabled bool) { send(MainMsg{kind: setDark, enabled: enabled}) }),
						togglebutton.New("").Key("chinese-language").Selected(model.Chinese).Content(text.New("Chinese locale")).
							OnChange(func(enabled bool) { send(MainMsg{kind: setChinese, enabled: enabled}) }),
						datepicker.New("locale-preview", model.Date).
							OnChange(func(value time.Time) { send(MainMsg{kind: setDate, date: value}) }),
					).Gap(8).LineGap(8),
					layout.Divider(),
					layout.Row(
						button.New("Open counter").OnClick(func() { application.Open(counter) }),
						button.New("Close all").Variant(button.Secondary).OnClick(application.CloseAll),
					).Gap(8).AlignMiddle(),
				).Gap(16),
			).Style(style.New().FillWidth().MaxWidth(720).Padding(24)),
		)
	}
}

func windowAction(application *ui.Application, key, label string, action window.Action) ui.Widget {
	return button.New(label).Key(key).
		Variant(button.Secondary).
		OnClick(func() { application.Perform("main", action) })
}

func main() {
	application := ui.NewApplication()
	counter := ui.NewWindow(
		"counter", ui.NewProgram(CounterModel{}, counterUpdate, counterView(application)), window.Title("Counter"),
		window.Size(420, 260),
		window.Theme(base.Light()),
	)
	mainWindow := ui.NewWindow(
		"main", ui.NewProgram(MainModel{}, mainUpdate(application), mainView(application, counter)), window.Title("Multi-window"),
		window.Size(760, 520),
		window.MinSize(520, 360),
		window.MaxSize(1400, 1000),
		window.Theme(base.Light()),
	)
	application.Run(mainWindow)
}
