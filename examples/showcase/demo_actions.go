package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/closebutton"
	"github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/togglebutton"
)

func demoButton(m demoState, send ui.Send[Msg]) ui.Widget {
	variants := []struct {
		label string
		value button.Variant
	}{
		{label: "主要操作", value: button.Primary},
		{label: "次要操作", value: button.Secondary},
		{label: "低强调", value: button.Tertiary},
		{label: "描边", value: button.Outline},
		{label: "轻量", value: button.Ghost},
		{label: "危险操作", value: button.Danger},
		{label: "危险柔和", value: button.DangerSoft},
		{label: "危险描边", value: button.DangerOutline},
	}
	buttons := []ui.Widget{}
	for index, variant := range variants {
		buttons = append(buttons, button.New(variant.label).Key(fmt.Sprintf("variant-%d", index)).
			Variant(variant.value).OnClick(report(send, "点击了“"+variant.label+"”")))
	}
	return layout.Column(
		muted("语义变体"), layout.Wrap(buttons...).Gap(10).LineGap(10),
		muted("尺寸与状态"),
		layout.Wrap(
			button.New("小按钮").Size(button.Small).OnClick(report(send, "小按钮")),
			button.New("默认按钮").OnClick(report(send, "默认按钮")),
			button.New("大按钮").Size(button.Large).OnClick(report(send, "大按钮")),
			button.New("保存中").Loading(true),
			button.New("不可用").Disabled(true),
		).Gap(10).LineGap(10).AlignMiddle(),
		muted("按钮组"),
		button.Group(
			button.New("左对齐").OnClick(report(send, "左对齐")),
			button.New("居中").OnClick(report(send, "居中")),
			button.New("右对齐").OnClick(report(send, "右对齐")),
		),
		layout.Divider(),
		layout.Wrap(
			button.New("点击计数").OnClick(func() {
				send(demoMsg(func(m *demoState) { m.clicks++ }))
			}),
			text.New(fmt.Sprintf("已点击 %d 次", m.clicks)),
		).Gap(12).AlignMiddle(),
	).Gap(16)
}

func demoToggle(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Wrap(
		togglebutton.New("加粗").Key("toggle-bold").Selected(m.bold).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.bold })),
		togglebutton.New("已锁定").Key("toggle-disabled").Selected(true).Disabled(true),
		text.New(fmt.Sprintf("加粗：%t", m.bold)),
	).Gap(12).AlignMiddle()
}

func demoClose(_ demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Row(text.New("关闭操作会在底部显示反馈"),
		closebutton.New("close-demo").OnClick(report(send, "已点击关闭按钮")),
	).Gap(16).AlignMiddle()
}

func demoCommand(_ demoState, send ui.Send[Msg]) ui.Widget {
	save := command.New("demo-save", "保存文档").
		Shortcut(command.KeyShortcut("S", command.ShortcutPrimary)).
		OnExecute(report(send, "保存命令已执行（按钮与 Ctrl+S 共享操作）"))
	return command.Scope([]command.Command{save}, layout.Column(
		command.Button("command-save", save),
		muted("点击按钮后，按 Ctrl+S / macOS ⌘S 再执行一次。"),
	).Gap(12))
}
