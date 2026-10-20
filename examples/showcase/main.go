// Showcase is an interactive catalog of FlowUI's public components.
package main

import (
	"image/color"
	"time"

	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/nodegraph"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/widget/tree"
	"github.com/qianniancn/flowui/widget/workbench"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	selected string
	query    string
	dark     bool
	navOpen  bool
	status   string
	demo     demoState
}

type demoState struct {
	name, email, password, notes string
	choice, combo, comboInput    string
	checked, enabled, bold       bool
	expanded                     bool
	value, lower, upper          float64
	page, clicks                 int
	date                         time.Time
	dateRange                    datepicker.DateRange
	color                        color.NRGBA
	selected, tab, panel         string
	keys, expandedKeys           []string
	dialog, confirm, popup       bool
	portal                       bool
	toasts                       []toast.Item
	toastID                      int
	graph                        nodegraph.Graph
	edgeID                       int
	workbench                    workbench.State
	amount                       float64
	segment, location            string
	loading, palette             bool
	virtualSelected              string
	virtualSort                  table.SortDescriptor
	treeItems                    []tree.SimpleItem
	treeSelected, treeFilter     string
	treeExpanded, treeChecked    []string
	treeHalfChecked              []string
	treeVersion                  uint64
	animationExpanded            bool
}

func initialDemo() demoState {
	day := time.Date(2026, time.October, 12, 0, 0, 0, 0, time.Local)
	return demoState{
		name: "FlowUI", email: "hello@flowui.dev", notes: "用 Go 构建桌面应用。",
		choice: "desktop", combo: "go", comboInput: "Go", checked: true, enabled: true,
		expanded: true, value: 64, lower: 20, upper: 80, page: 1,
		date: day, dateRange: datepicker.DateRange{Start: day, End: day.AddDate(0, 0, 6)},
		color:    color.NRGBA{R: 62, G: 99, B: 221, A: 255},
		selected: "flowui", tab: "overview", panel: "preview",
		keys: []string{"go"}, expandedKeys: []string{"src", "widgets"},
		toasts: []toast.Item{}, graph: initialGraph(), workbench: initialWorkbench(),
		amount: 12.5, segment: "day", location: "components", loading: true,
		treeItems: initialExplorer(), treeExpanded: []string{"src"},
		treeChecked: []string{}, treeHalfChecked: []string{},
	}
}

func initialModel() Model {
	return Model{selected: "button", status: "选择组件，试试它的交互。", demo: initialDemo()}
}

type Msg interface{ msg() }
type searchMsg string
type selectMsg string
type logMsg string
type themeMsg struct{}
type navMsg bool
type resetMsg struct{}
type demoMsg func(*demoState)

func (searchMsg) msg() {}
func (selectMsg) msg() {}
func (logMsg) msg()    {}
func (themeMsg) msg()  {}
func (navMsg) msg()    {}
func (resetMsg) msg()  {}
func (demoMsg) msg()   {}

func update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case searchMsg:
		m.query = string(msg)
	case selectMsg:
		if _, ok := findEntry(string(msg)); !ok {
			return nil
		}
		m.selected = string(msg)
		m.navOpen = false
		m.demo.dialog, m.demo.confirm, m.demo.popup, m.demo.portal = false, false, false, false
		m.demo.palette = false
		m.status = "正在展示 " + m.selected
	case themeMsg:
		m.dark = !m.dark
	case navMsg:
		m.navOpen = bool(msg)
	case logMsg:
		m.status = string(msg)
	case demoMsg:
		msg(&m.demo)
	case resetMsg:
		m.demo = initialDemo()
		m.status = "演示状态已重置"
	}
	return nil
}

func main() {
	app := ui.NewApplication()
	program := ui.NewProgram(initialModel(), func(m *Model, msg Msg) ui.Cmd[Msg] {
		cmd := update(m, msg)
		if _, ok := msg.(themeMsg); ok {
			active := base.Light()
			if m.dark {
				active = base.Dark()
			}
			app.SetTheme("main", active)
		}
		return cmd
	}, View)
	app.Run(ui.NewWindow("main", program,
		window.Title("FlowUI · 组件展示"),
		window.Size(1240, 860),
		window.Theme(base.Light()),
		window.Locale(window.LanguageChinese),
	))
}

func report(send ui.Send[Msg], action string) func() {
	return func() { send(logMsg(action)) }
}

func setString(send ui.Send[Msg], target func(*demoState) *string) func(string) {
	return func(value string) {
		send(demoMsg(func(m *demoState) { *target(m) = value }))
	}
}

func setBool(send ui.Send[Msg], target func(*demoState) *bool) func(bool) {
	return func(value bool) {
		send(demoMsg(func(m *demoState) { *target(m) = value }))
	}
}

func setNumber(send ui.Send[Msg]) func(float64) {
	return func(value float64) { send(demoMsg(func(m *demoState) { m.value = value })) }
}

func selection(send ui.Send[Msg]) func(string) {
	return func(value string) {
		send(demoMsg(func(m *demoState) { m.selected = value }))
		send(logMsg("已选择：" + value))
	}
}
