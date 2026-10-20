package main

import (
	"fmt"
	"strings"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/collapsible"
	"github.com/qianniancn/flowui/widget/listbox"
	"github.com/qianniancn/flowui/widget/pagination"
	"github.com/qianniancn/flowui/widget/sidebar"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/tabs"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/timeline"
	"github.com/qianniancn/flowui/widget/tree"
)

func demoList(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		muted("单选"),
		listbox.New(languageItems()).Key("list-single").SelectedKey(m.combo).FullWidth().MaxHeight(220).
			OnChange(setString(send, func(m *demoState) *string { return &m.combo })),
		muted("多选"),
		listbox.New(languageItems()).Key("list-multiple").SelectedKeys(m.keys).FullWidth().MaxHeight(220).
			OnSelectionChange(func(keys []string) { send(demoMsg(func(m *demoState) { m.keys = keys })) }),
		text.New("已选："+strings.Join(m.keys, ", ")),
	).Gap(12)
}

func demoTable(m demoState, send ui.Send[Msg]) ui.Widget {
	columns := []table.Column{
		{Key: "name", Label: "项目", Weight: 1}, {Key: "type", Label: "类型", Width: 120},
		{Key: "status", Label: "状态", Width: 100},
	}
	rows := []table.Row{
		{Key: "flowui", Cells: []table.Cell{{Text: "FlowUI"}, {Text: "组件库"}, {Text: "进行中"}}},
		{Key: "studio", Cells: []table.Cell{{Text: "Studio"}, {Text: "编辑器"}, {Text: "计划中"}}},
		{Key: "docs", Cells: []table.Cell{{Text: "Documentation"}, {Text: "文档"}, {Text: "已发布"}}},
	}
	return layout.Column(
		table.New("project-table", columns, rows).Border(true).
			SelectionMode(table.SelectionSingle).SelectedKey(m.selected).OnChange(selection(send)),
		text.New("选中项目："+m.selected),
	).Gap(16)
}

func demoTree(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []tree.Item{
		{Key: "src", Label: "src", Children: []tree.Item{
			{Key: "main.go", Label: "main.go"},
			{Key: "widgets", Label: "widgets", Children: []tree.Item{
				{Key: "button.go", Label: "button.go"}, {Key: "input.go", Label: "input.go"},
			}},
		}},
		{Key: "readme", Label: "README.md"},
	}
	return layout.Column(
		fixedHeight(270, tree.New("file-tree", m.selected, items).ExpandedKeys(m.expandedKeys).
			OnExpandedChange(func(keys []string) { send(demoMsg(func(m *demoState) { m.expandedKeys = keys })) }).
			OnChange(selection(send))),
		text.New("当前文件："+m.selected),
	).Gap(16)
}

func demoTabs(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []tabs.TabItem{
		{Key: "overview", Label: "概览", Panel: text.New("概览：项目运行正常。")},
		{Key: "activity", Label: "动态", Panel: text.New("最近动态：更新了默认组件主题。")},
		{Key: "settings", Label: "设置", Panel: text.New("项目设置：Go / Gio。")},
	}
	changed := setString(send, func(m *demoState) *string { return &m.tab })
	return layout.Column(
		tabs.New("tabs-primary", m.tab, items).OnChange(changed),
		layout.Divider(),
		tabs.New("tabs-secondary", m.tab, items).Variant(tabs.Secondary).OnChange(changed),
	).Gap(24)
}

func demoPagination(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		pagination.New("pagination", m.page, 12).Labels("上一页", "下一页").OnChange(func(page int) {
			send(demoMsg(func(m *demoState) { m.page = page }))
		}),
		text.New(fmt.Sprintf("第 %d 页 / 共 12 页", m.page)),
	).Gap(20)
}

func demoSidebar(m demoState, send ui.Send[Msg]) ui.Widget {
	return fixedHeight(340, sidebar.NewSections("sidebar-demo", m.selected, []sidebar.Section{
		{Title: "项目", Items: []sidebar.Item{{Key: "flowui", Label: "FlowUI"}, {Key: "studio", Label: "Studio"}}},
		{Title: "资源", Items: []sidebar.Item{{Key: "docs", Label: "文档"}, {Key: "settings", Label: "设置"}}},
	}).Header(text.New("我的工作区")).Footer(muted("4 个入口")).OnChange(selection(send)))
}

func demoTimeline(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return timeline.New([]timeline.Item{
		{Title: text.New("设计"), Content: text.New("确定默认主题与组件规范"), Color: timeline.ColorGreen},
		{Title: text.New("开发"), Content: text.New("构建全部组件展示")},
		{Title: text.New("验证"), Content: text.New("检查交互与明暗模式")},
	}).Mode(timeline.ModeAlternate)
}

func demoCollapsible(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		collapsible.New("项目详情", text.New("FlowUI 是基于 Gio 的 Go 桌面 UI 框架。")).Key("collapse-single").Expanded(m.expanded).Expanded(m.expanded).OnExpandedChange(setBool(send, func(m *demoState) *bool { return &m.expanded })),
		collapsible.Group("collapse-group", m.keys, []collapsible.Item{
			{Key: "go", Label: "如何运行？", Content: text.New("在仓库根目录运行 go run ./examples/showcase")},
			{Key: "theme", Label: "如何切换主题？", Content: text.New("使用侧栏底部或页面顶部的明暗切换按钮。")},
		}).OnExpandedChange(func(keys []string) { send(demoMsg(func(m *demoState) { m.keys = keys })) }),
	).Gap(20)
}
