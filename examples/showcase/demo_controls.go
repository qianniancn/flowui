package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/breadcrumb"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/commandpalette"
	"github.com/qianniancn/flowui/widget/numericinput"
	"github.com/qianniancn/flowui/widget/segmented"
	"github.com/qianniancn/flowui/widget/skeleton"
	"github.com/qianniancn/flowui/widget/switchcontrol"
	"github.com/qianniancn/flowui/widget/text"
)

func demoNumericInput(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		muted("金额 · -100 至 100，每步 0.25，保留两位小数"),
		numericinput.New("amount", m.amount).Range(-100, 100).Step(0.25).Precision(2).Label("金额").
			OnChange(func(value float64) { send(demoMsg(func(m *demoState) { m.amount = value })) }),
		text.New(fmt.Sprintf("已提交：%.2f", m.amount)),
		muted("输入时保留草稿；回车或离开输入框时提交，加减按钮立即生效。"),
		numericinput.New("amount-disabled", 42).Label("禁用金额").Disabled(true),
	).Gap(12)
}

func demoSegmented(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []segmented.Item{
		{Key: "day", Label: "日"}, {Key: "week", Label: "周"}, {Key: "month", Label: "月"},
		{Key: "year", Label: "年", Disabled: true},
	}
	changed := setString(send, func(m *demoState) *string { return &m.segment })
	return layout.Column(
		segmented.New(items).Key("period").SelectedKey(m.segment).OnChange(changed),
		muted("占满容器宽度"),
		segmented.New(items).Key("period-full").SelectedKey(m.segment).FullWidth().OnChange(changed),
		text.New("当前周期："+m.segment),
	).Gap(16)
}

func demoBreadcrumb(m demoState, send ui.Send[Msg]) ui.Widget {
	path := []breadcrumb.Item{
		{Key: "home", Label: "首页"}, {Key: "docs", Label: "文档"}, {Key: "components", Label: "组件"},
	}
	for index, item := range path {
		if item.Key == m.location {
			path = path[:index+1]
			break
		}
	}
	return layout.Column(
		breadcrumb.New("location", path).OnChange(setString(send, func(m *demoState) *string { return &m.location })),
		text.New("当前位置："+m.location),
		button.New("回到组件页").Variant(button.Secondary).
			OnClick(func() { send(demoMsg(func(m *demoState) { m.location = "components" })) }),
	).Gap(16)
}

func demoSkeleton(m demoState, send ui.Send[Msg]) ui.Widget {
	var content ui.Widget = layout.Column(text.New("FlowUI 项目"), muted("内容已经加载，可以开始使用。")).Gap(8)
	if m.loading {
		content = layout.Row(
			skeleton.New().Circle().Height(48),
			layout.Expanded(layout.Column(skeleton.New().Width(160), skeleton.New(), skeleton.New().Width(220)).Gap(10)),
		).Gap(16).AlignMiddle()
	}
	return layout.Column(
		switchcontrol.New("模拟加载中").Key("loading").Checked(m.loading).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.loading })),
		content,
		muted("静态占位（关闭动效时，所有骨架屏都会停止刷新）"),
		skeleton.New().Height(80).Animated(false),
	).Gap(20)
}

func demoCommandPalette(m demoState, send ui.Send[Msg]) ui.Widget {
	commands := []command.Command{
		command.New("save", "保存项目").Description("保存当前演示内容").
			Shortcut(command.KeyShortcut("S", command.ShortcutPrimary)).OnExecute(report(send, "已保存演示项目")),
		command.New("theme", "切换明暗主题").Description("切换默认主题").OnExecute(func() { send(themeMsg{}) }),
		command.New("reset", "重置演示").Description("恢复所有示例的初始数据").OnExecute(func() { send(resetMsg{}) }),
		command.New("disabled", "发布项目（未启用）").Disabled(true),
	}
	return command.Scope(commands, layout.Column(
		button.New("打开命令面板").OnClick(func() {
			send(demoMsg(func(m *demoState) { m.palette = true }))
		}),
		muted("搜索名称或说明；↑ / ↓ 选择，Enter 执行，Esc 关闭。Ctrl+S 与菜单共用保存命令。"),
		commandpalette.New("palette", m.palette, commands).
			OnOpenChange(setBool(send, func(m *demoState) *bool { return &m.palette })),
	).Gap(16))
}
