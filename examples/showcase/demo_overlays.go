package main

import (
	"fmt"
	"image"
	"slices"
	"time"

	giolayout "gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/alert"
	"github.com/qianniancn/flowui/widget/alertdialog"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/dropdown"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/menubar"
	"github.com/qianniancn/flowui/widget/modal"
	"github.com/qianniancn/flowui/widget/popover"
	"github.com/qianniancn/flowui/widget/portal"
	"github.com/qianniancn/flowui/widget/progress"
	"github.com/qianniancn/flowui/widget/slider"
	"github.com/qianniancn/flowui/widget/spinner"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toast"
	"github.com/qianniancn/flowui/widget/tooltip"
)

func actionItems(send ui.Send[Msg]) []menu.Item {
	return []menu.Item{
		{Key: "new", Label: "新建项目", Shortcut: "Ctrl+N", OnAction: report(send, "已新建演示项目")},
		{Key: "save", Label: "保存", Shortcut: "Ctrl+S", OnAction: report(send, "已保存演示项目")},
		{Key: "export", Label: "导出", Children: []menu.Item{
			{Key: "png", Label: "PNG 图片", OnAction: report(send, "已选择导出 PNG")},
			{Key: "json", Label: "JSON 数据", OnAction: report(send, "已选择导出 JSON")},
		}},
		menu.Separator(),
		{Key: "disabled", Label: "暂不可用", Disabled: true},
		menu.Separator(),
		{Key: "delete", Label: "删除项目", Shortcut: "Del", Variant: menu.ItemDanger,
			OnAction: report(send, "已删除演示项目")},
	}
}

func demoMenu(_ demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		menu.New("inline-menu", actionItems(send)),
		menu.ContextMenu("context-demo", button.New("在此右键打开菜单").Variant(button.Secondary),
			menu.New("context-actions", actionItems(send))),
	).Gap(24)
}

func demoDropdown(_ demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Wrap(
		dropdown.New("dropdown-demo", button.New("更多操作"), actionItems(send)),
		dropdown.Button("dropdown-split", button.New("发布"), actionItems(send)).
			OnClick(report(send, "已发布演示项目")),
	).Gap(16).LineGap(16)
}

func demoMenubar(_ demoState, send ui.Send[Msg]) ui.Widget {
	return menubar.New("menubar-demo", []menubar.Item{
		menubar.NewMenu("file", "文件", actionItems(send)),
		menubar.NewMenu("edit", "编辑", []menu.Item{
			{Key: "undo", Label: "撤销", OnAction: report(send, "撤销操作")},
			{Key: "copy", Label: "复制", OnAction: report(send, "复制操作")},
		}),
	})
}

func demoModal(m demoState, send ui.Send[Msg]) ui.Widget {
	openChanged := setBool(send, func(m *demoState) *bool { return &m.dialog })
	return layout.Column(
		button.New("编辑项目").OnClick(func() { openChanged(true) }),
		modal.New("编辑项目", layout.Column(
			text.New("修改项目名称，然后保存。"),
			input.New(m.name).Key("modal-name").Label("编辑项目名称").FullWidth().
				OnChange(setString(send, func(m *demoState) *string { return &m.name })),
		).Gap(16)).Key("demo-dialog").Open(m.dialog).Open(m.dialog).OnOpenChange(openChanged).Footer(layout.Wrap(
			button.New("取消").Variant(button.Secondary).OnClick(func() { openChanged(false) }),
			button.New("保存").OnClick(func() {
				openChanged(false)
				send(logMsg("已保存项目：" + m.name))
			}),
		).Gap(8)),
	).Gap(12)
}

func demoAlertDialog(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := setBool(send, func(m *demoState) *bool { return &m.confirm })
	return layout.Column(
		button.New("删除演示项目").Variant(button.Danger).OnClick(func() { changed(true) }),
		alertdialog.New("删除项目？", "此处仅演示确认流程，不涉及磁盘文件。这个操作需要明确确认。").Key("demo-confirm").Open(m.confirm).
			Open(m.confirm).OnOpenChange(changed).Footer(layout.Wrap(
			button.New("保留项目").Variant(button.Secondary).OnClick(func() { changed(false) }),
			button.New("确认删除").Variant(button.Danger).OnClick(func() {
				changed(false)
				send(logMsg("已确认删除演示项目"))
			}),
		).Gap(8)),
	).Gap(12)
}

func demoPopover(m demoState, send ui.Send[Msg]) ui.Widget {
	openChanged := setBool(send, func(m *demoState) *bool { return &m.popup })
	return popover.New(button.New("项目信息").
		OnClick(func() { openChanged(!m.popup) }), layout.Column(text.New("FlowUI · 本地项目"),
		button.New("查看详情").OnClick(report(send, "查看项目详情")),
	).Gap(16)).Key("popover-demo").Open(m.popup).Open(m.popup).Heading("项目信息").
		OnOpenChange(openChanged)
}

func demoTooltip(_ demoState, send ui.Send[Msg]) ui.Widget {
	return tooltip.New("tooltip-demo", button.New("悬停或聚焦我").
		OnClick(report(send, "已保存当前文档")), "保存当前文档（Ctrl+S）")
}

func demoPortal(m demoState, send ui.Send[Msg]) ui.Widget {
	trigger := button.New("切换根浮层").OnClick(func() {
		send(demoMsg(func(m *demoState) { m.portal = !m.portal }))
	})
	return portal.New("portal-demo", m.portal, trigger, func(anchor image.Rectangle, _ bool) ui.Widget {
		return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			position := image.Pt(anchor.Min.X, anchor.Max.Y+8)
			defer op.Offset(position).Push(gtx.Ops).Pop()
			gtx.Constraints.Min = image.Point{}
			gtx.Constraints.Max.X = max(1, gtx.Constraints.Max.X-position.X)
			gtx.Constraints.Max.Y = max(1, gtx.Constraints.Max.Y-position.Y)
			return card.New(text.New("这段内容绘制在窗口根浮层。")).Layout(ctx, gtx)
		})
	}).Passive(true)
}

func demoAlert(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		alert.New("项目提示", "所有演示都使用本地数据。"),
		alert.New("保存成功", "项目的最新更改已同步。").Status(alert.Success),
		alert.New("即将达到限额", "请检查项目使用量。").Status(alert.Warning),
		alert.New("无法连接", "请检查网络设置后重试。").Status(alert.Danger),
	).Gap(12)
}

func demoToast(_ demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		layout.Wrap(
			button.New("显示成功消息").OnClick(func() { addToast(send, false) }),
			button.New("显示可撤销消息").Variant(button.Secondary).
				OnClick(func() { addToast(send, true) }),
		).Gap(12).LineGap(12),
		muted("消息出现在窗口角落，4 秒后消失，也可以手动关闭。"),
	).Gap(16)
}

func addToast(send ui.Send[Msg], action bool) {
	send(demoMsg(func(m *demoState) {
		m.toastID++
		item := toast.NewItem(fmt.Sprintf("toast-%d", m.toastID), "保存成功").
			Description("演示更改已保存").Variant(toast.Success).Timeout(4 * time.Second)
		if action {
			item = item.Action("撤销")
		}
		m.toasts = append(m.toasts, item)
	}))
}

func removeToast(send ui.Send[Msg], key string) {
	send(demoMsg(func(m *demoState) {
		m.toasts = slices.DeleteFunc(slices.Clone(m.toasts), func(item toast.Item) bool { return item.Key() == key })
	}))
}

func demoProgress(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		slider.New("progress-control", m.value).Label("调整进度").ShowValue().OnChange(setNumber(send)),
		progress.Bar("progress-bar", m.value).Label("上传进度").ShowValue(),
		progress.Circle("progress-circle", m.value),
		progress.Meter("progress-meter", m.value).Label("存储使用量").ShowValue(),
	).Gap(24)
}

func demoSpinner(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Wrap(
		spinner.New().Size(spinner.Small), spinner.New(), spinner.New().Size(spinner.Large),
		spinner.New().Size(spinner.ExtraLarge), text.New("正在加载…"),
	).Gap(24).AlignMiddle()
}
