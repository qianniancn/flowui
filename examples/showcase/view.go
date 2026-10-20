package main

import (
	"fmt"
	"strings"

	"gioui.org/font"
	giolayout "gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	widgetcode "github.com/qianniancn/flowui/widget/code"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/sidebar"
	"github.com/qianniancn/flowui/widget/statusbar"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/toast"
)

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		compact := gtx.Constraints.Max.X < gtx.Dp(unit.Dp(820))
		var body ui.Widget = detail(m, send, compact)
		if compact && m.navOpen {
			body = navigation(m, send, true)
		}
		if !compact {
			body = layout.Row(navigation(m, send, false), layout.Expanded(body))
		}
		root := layout.Stack(
			layout.Stacked(layout.Column(
				layout.Expanded(body),
				statusbar.New(text.New(m.status).MaxLines(1).Truncator("…"), text.New("FlowUI · Go / Gio")),
			)),
			layout.Overlay(toast.New("showcase-toasts", m.demo.toasts).
				OnClose(func(key string) { removeToast(send, key) }).
				OnAction(func(key string) {
					removeToast(send, key)
					send(logMsg("已撤销演示操作"))
				})),
		)
		return root.Layout(ctx, gtx)
	})
}

func navigation(m Model, send ui.Send[Msg], compact bool) ui.Widget {
	sections, count := filteredSections(m.query)
	themeLabel := "切换深色"
	if m.dark {
		themeLabel = "切换浅色"
	}
	header := []ui.Widget{
		text.New("FlowUI").Size(26).Weight(font.Bold),
		muted("组件展示 / Component showcase"),
		input.New(m.query).Key("catalog-search").Label("搜索组件").
			Hint("搜索名称、用途或包名").FullWidth().
			OnChange(func(value string) { send(searchMsg(value)) }),
		muted(fmt.Sprintf("%d / %d 个演示", count, len(catalog))),
	}
	if compact {
		header = append(header, button.New("返回演示").Variant(button.Ghost).
			OnClick(func() { send(navMsg(false)) }))
	}
	nav := sidebar.NewSections("catalog", m.selected, sections).
		Width(280).Padding(16).ItemHeight(34).ItemGap(2).
		Header(layout.Column(header...).Gap(12)).
		Footer(layout.Column(
			layout.Wrap(
				button.New(themeLabel).Key("theme-toggle").Variant(button.Secondary).
					OnClick(func() { send(themeMsg{}) }),
				button.New("重置演示").Variant(button.Ghost).
					OnClick(func() { send(resetMsg{}) }),
			).Gap(8),
			muted("默认主题 · 本地演示，无需联网"),
		).Gap(10)).
		EmptyText("没有匹配组件，请换个关键词").
		OnChange(func(key string) { send(selectMsg(key)) })
	if compact {
		return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
			width := int(gtx.Metric.PxToDp(gtx.Constraints.Max.X))
			return nav.Width(max(1, width)).Layout(ctx, gtx)
		})
	}
	return nav
}

func detail(m Model, send ui.Send[Msg], compact bool) ui.Widget {
	item, ok := findEntry(m.selected)
	if !ok {
		item = catalog[0]
	}
	padding := unit.Dp(28)
	if compact {
		padding = 16
	}
	path := "github.com/qianniancn/flowui/layout"
	if item.packages != "" {
		path = "github.com/qianniancn/flowui/widget/" + strings.Fields(item.packages)[0]
	}
	if item.importPath != "" {
		path = item.importPath
	}
	sections := []ui.Widget{
		layout.Column(
			muted(item.group),
			text.New(item.title).Size(28).Weight(font.Bold),
			text.New(item.summary).Size(14),
		).Gap(10),
		section("交互演示", item.demo(m.demo, send)),
	}
	if item.code != "" {
		sections = append(sections, section("基本用法", widgetcode.New("code-"+item.key, "go", item.code)))
	}
	if documentation := componentDocumentation(item.key); documentation != nil {
		sections = append(sections, documentation)
	}
	sections = append(sections, layout.Column(
		muted(path),
		muted("完整示例：examples/showcase/"+item.source),
		muted("修改控件后可在底部查看反馈；“重置演示”恢复初始值。"),
	).Gap(8))
	content := layout.Column(sections...).Gap(24)
	header := layout.Wrap(
		muted("组件 / "+item.title),
		button.New("明暗切换").Variant(button.Ghost).
			OnClick(func() { send(themeMsg{}) }),
	).Gap(12)
	var top ui.Widget = header
	if compact {
		top = layout.Wrap(
			button.New("组件目录").Variant(button.Secondary).
				OnClick(func() { send(navMsg(true)) }),
			header,
		).Gap(8)
	}
	return layout.Column(
		layout.Box(top).Style(style.New().FillWidth().PaddingX(padding).PaddingY(10)),
		layout.Divider(),
		layout.Expanded(layout.Scroll(
			layout.Box(content).Style(style.New().FillWidth().MaxWidth(1080).Padding(padding)),
		).Key("detail-scroll-"+item.key).Vertical()),
	)
}

func section(title string, body ui.Widget) ui.Widget {
	return card.New(layout.Column(
		text.New(title).Size(14).Weight(font.Bold),
		body,
	).Gap(20))
}

func muted(value string) ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		active := ctx.Theme()
		return text.New(value).Size(12).Color(active.Semantic().MutedForeground).Layout(ctx, gtx)
	})
}

func fixedHeight(height unit.Dp, child ui.Widget) ui.Widget {
	return layout.Box(child).Style(style.New().FillWidth().Height(height))
}
