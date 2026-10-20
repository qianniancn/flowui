package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/dock"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/panel"
	"github.com/qianniancn/flowui/widget/statusbar"
	"github.com/qianniancn/flowui/widget/tabs"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/titlebar"
	"github.com/qianniancn/flowui/widget/togglebutton"
	"github.com/qianniancn/flowui/widget/toolbar"
	"github.com/qianniancn/flowui/widget/workbench"
)

func demoToolbar(m demoState, send ui.Send[Msg]) ui.Widget {
	return toolbar.New(
		button.New("保存").OnClick(report(send, "工具栏：已保存")),
		button.New("撤销").Variant(button.Ghost).OnClick(report(send, "工具栏：已撤销")),
		toolbar.Separator(),
		togglebutton.New("加粗").Key("tool-bold").Selected(m.bold).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.bold })),
	)
}

func demoStatusbar(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return statusbar.New(text.New("● 已同步 · 3 个文件"), text.New("UTF-8 · Go · 100%"))
}

func demoTitlebar(m demoState, send ui.Send[Msg]) ui.Widget {
	return titlebar.New("titlebar-demo", "FlowUI Studio", demoMenubar(m, send)).
		ShowMinimize(false).ShowMaximize(false).ShowClose(false)
}

func demoPanel(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		layout.Wrap(
			button.New("预览").OnClick(func() {
				send(demoMsg(func(m *demoState) { m.panel = "preview" }))
			}),
			button.New("编辑").Variant(button.Secondary).OnClick(func() {
				send(demoMsg(func(m *demoState) { m.panel = "edit" }))
			}),
		).Gap(8),
		panel.New("panel-demo", m.panel, []panel.Item{
			{Key: "preview", Content: text.New("当前预览：" + m.name)},
			{Key: "edit", Content: input.New(m.name).Key("panel-name").FullWidth().
				OnChange(setString(send, func(m *demoState) *string { return &m.name }))},
		}).KeepAlive(true),
	).Gap(20)
}

func demoDock(_ demoState, _ ui.Send[Msg]) ui.Widget {
	root := dock.Split("dock-main", dock.Horizontal,
		dock.Panel("files", pane("资源", "main.go\nview.go\nmodel.go")),
		dock.Split("dock-right", dock.Vertical,
			dock.Panel("editor", pane("编辑区", "拖动分隔条调整各区域大小。")),
			dock.Panel("output", pane("输出", "Build completed successfully.")),
		).Ratio(0.7),
	).Ratio(0.28)
	return fixedHeight(340, dock.New("dock-demo", root))
}

func pane(title, content string) ui.Widget {
	return card.New(layout.Column(text.New(title).Size(16), text.New(content)).Gap(16))
}

func initialWorkbench() workbench.State {
	return workbench.NewState([]workbench.Group{{
		Key: "editor", ActiveKey: "main.go", Tabs: []workbench.Tab{
			{Key: "main.go", Title: "main.go", Closable: true},
			{Key: "README.md", Title: "README.md", Closable: true},
		},
	}})
}

func demoWorkbench(m demoState, send ui.Send[Msg]) ui.Widget {
	controller := workbench.NewController(m.workbench)
	controller.OnEvent(func(event workbench.Event) {
		state := controller.State()
		send(demoMsg(func(m *demoState) { m.workbench = state }))
		send(logMsg(fmt.Sprintf("工作台：%s", event.TabKey)))
	})
	items := []tabs.TabItem{}
	panels := []panel.Item{}
	for _, group := range m.workbench.Groups {
		for _, tab := range group.Tabs {
			items = append(items, tabs.TabItem{Key: tab.Key, Label: tab.Title, Closable: true})
			panels = append(panels, panel.Item{Key: tab.Key, Content: pane(tab.Title, "此内容由工作台状态选择。")})
		}
	}
	tabstrip := tabs.New("workbench-tabs", "", items).OnClose(func(key string) {
		controller.CloseTab("editor", key)
	})
	return controller.CommandScope(layout.Column(
		button.New("重新打开文档").Variant(button.Secondary).OnClick(func() {
			send(demoMsg(func(m *demoState) { m.workbench = initialWorkbench() }))
		}),
		controller.BindTabs("editor", tabstrip),
		controller.BindPanel("editor", panel.New("workbench-panel", "", panels).KeepAlive(true)),
	).Gap(16))
}

func demoLayout(_ demoState, _ ui.Send[Msg]) ui.Widget {
	blocks := []ui.Widget{}
	for _, title := range []string{"项目", "组件", "主题", "窗口", "状态", "布局"} {
		blocks = append(blocks, card.New(text.New(title)))
	}
	return layout.Column(
		muted("Wrap · 宽度不足时自动换行"), layout.Wrap(blocks...).Gap(8).LineGap(8),
		muted("Grid · 两列网格"), layout.Grid(2, blocks...).Gap(8),
		muted("Stack · 叠放内容"),
		layout.Stack(layout.Stacked(pane("基础层", "叠放布局")),
			layout.Overlay(text.New("浮层标注")).Offset(180, 12)),
		muted("SplitPane · 可拖动分割条"),
		fixedHeight(180, layout.SplitPane(pane("左侧", "拖动分隔条"), pane("右侧", "比例自动更新")).
			DefaultRatio(0.4)),
		muted("Scroll · 独立滚动区域"),
		layout.Box(layout.Scroll(layout.Column(blocks...).Gap(8)).Vertical()).
			Style(style.New().Height(160).FillWidth()),
	).Gap(16)
}

func demoBuildingBlocks(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		text.New("chart → 折线、柱状、饼图等图表的选择和坐标协议"),
		text.New("disclosure → 折叠面板、弹层与内容面板的展开状态"),
		text.New("navigation → 侧栏、树与分页的导航协议"),
		text.New("optionrow → 菜单和选择列表共用的选项行布局"),
		muted("这些包服务于组件组合；日常使用优先从具体 widget 包开始。"),
	).Gap(16)
}
