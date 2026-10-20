package main

import (
	"bytes"
	"fmt"
	"image/jpeg"

	"gioui.org/font"
	"gioui.org/op/paint"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/assets/images"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/avatar"
	"github.com/qianniancn/flowui/widget/badge"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/code"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/image"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
)

func demoText(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		text.New("用 Go 构建桌面应用").Size(28).Weight(font.Bold),
		text.New("正文 · 默认字号与主题前景色"),
		muted("辅助信息 · 随明暗主题调整对比度"),
		text.Selectable("selectable-text", "这段文字可以用鼠标选择并复制。"),
		layout.Divider(),
		label.New("项目名称").For("label-project"),
		input.New(m.name).Key("label-project").FullWidth().
			OnChange(setString(send, func(m *demoState) *string { return &m.name })),
		description.New("使用易于识别的名称，点击标签可聚焦输入框。").For("label-project"),
	).Gap(14)
}

func demoIcon(_ demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		layout.Wrap(
			icon.New(lucide.Search).Size(16), icon.New(lucide.Plus).Size(20),
			icon.New(lucide.Settings).Size(24), icon.New(lucide.Heart).Size(32),
		).Gap(24).AlignMiddle(),
		layout.Wrap(
			button.New("添加").Content(icon.New(lucide.Plus)).IconOnly().OnClick(report(send, "添加")),
			button.New("保存").Content(layout.Row(icon.New(lucide.Save).Size(16), text.New("保存")).Gap(8)).
				OnClick(report(send, "已保存")),
		).Gap(12),
	).Gap(24)
}

var desertImage = func() paint.ImageOp {
	img, err := jpeg.Decode(bytes.NewReader(images.BGDesertJPG))
	if err != nil {
		panic(fmt.Errorf("decode bundled showcase image: %w", err))
	}
	return paint.NewImageOp(img)
}()

func demoImage(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		muted("Cover · 填充并裁剪"),
		image.New(desertImage).Width(640).Height(210).Fit(image.FitCover).Radius(8).Alt("沙漠沙丘"),
		muted("Contain · 保留完整图片"),
		image.New(desertImage).Width(360).Height(200).Fit(image.FitContain).Radius(8).Alt("完整沙丘图片"),
	).Gap(12)
}

func demoAvatar(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Wrap(
		avatar.New("FL").Size(avatar.SizeSmall), avatar.New("UI").Size(avatar.SizeMedium),
		avatar.New("GO").Size(avatar.SizeLarge),
	).Gap(20).AlignMiddle()
}

func demoBadge(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Wrap(
		badge.New(button.New("通知").Variant(button.Secondary).OnClick(func() {
			send(demoMsg(func(m *demoState) { m.clicks++ }))
		}), fmt.Sprint(m.clicks+3)),
		badge.New(avatar.New("FL"), "8"),
	).Gap(32).AlignMiddle()
}

func demoChip(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Wrap(chip.New("已发布"), chip.New("Go"), chip.New("桌面应用"), chip.New("v1.0")).Gap(10)
}

func demoCard(_ demoState, send ui.Send[Msg]) ui.Widget {
	return card.New(layout.Column(
		layout.Row(avatar.New("FL"), layout.Column(
			text.New("FlowUI").Size(20).Weight(font.Bold), muted("原生桌面组件库"),
		).Gap(4)).Gap(12).AlignMiddle(),
		text.New("一套默认风格，组合组件来构建你的应用。"),
		button.New("打开项目").OnClick(report(send, "已打开 FlowUI 项目")),
	).Gap(20))
}

func demoSurface(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		surface.New(text.New("Default · 默认表面")),
		surface.New(text.New("Secondary · 次级表面")).Variant(surface.Secondary),
		surface.New(text.New("Tertiary · 三级表面")).Variant(surface.Tertiary),
		surface.New(text.New("Transparent · 透明表面")).Variant(surface.Transparent),
	).Gap(12)
}

func demoCode(_ demoState, _ ui.Send[Msg]) ui.Widget {
	return layout.Column(
		text.New("语法高亮 · 可拖拽选中").Size(13).Weight(font.Bold),
		code.New("demo-code", "go", `// Save writes the model to disk.
func (m *Model) Save() error {
	if !m.Dirty {
		return nil
	}
	return os.WriteFile(m.Path, m.Bytes(), 0o644)
}`),
		muted("右上角按钮复制全部代码；也可拖拽选中后按 Ctrl/Cmd+C 复制。长行按容器宽度换行。"),
	).Gap(12)
}
