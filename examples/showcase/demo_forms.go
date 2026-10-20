package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/checkbox"
	"github.com/qianniancn/flowui/widget/colorpicker"
	"github.com/qianniancn/flowui/widget/combobox"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/radiogroup"
	"github.com/qianniancn/flowui/widget/selectfield"
	"github.com/qianniancn/flowui/widget/slider"
	"github.com/qianniancn/flowui/widget/switchcontrol"
	"github.com/qianniancn/flowui/widget/text"
)

func demoInput(m demoState, send ui.Send[Msg]) ui.Widget {
	nameChanged := setString(send, func(m *demoState) *string { return &m.name })
	return layout.Column(
		field("项目名称", input.New(m.name).Key("input-name").Label("项目名称输入框").Hint("输入项目名称").
			FullWidth().OnChange(nameChanged).OnSubmit(func(value string) { send(logMsg("已提交：" + value)) })),
		field("邮箱", input.New(m.email).Key("input-email").Label("邮箱").Type(input.Email).FullWidth().
			OnChange(setString(send, func(m *demoState) *string { return &m.email }))),
		field("密码", input.New(m.password).Key("input-password").Label("密码").Type(input.Password).
			Hint("输入密码").FullWidth().OnChange(setString(send, func(m *demoState) *string { return &m.password }))),
		field("多行输入", input.NewTextArea(m.notes).Key("input-notes").Label("备注").Rows(3).FullWidth().
			OnChange(setString(send, func(m *demoState) *string { return &m.notes }))),
		field("前后缀与操作", input.Group(input.New(m.name).Key("input-group").OnChange(nameChanged)).
			Prefix(text.New("项目 / ")).Suffix(text.New(".go")).FullWidth().
			SuffixAction(input.GroupAction("clear-name", "清空").OnClick(func() { nameChanged("") }))),
		input.GroupTextArea(input.NewTextArea(m.notes).Key("notes-group").Rows(2).
			OnChange(setString(send, func(m *demoState) *string { return &m.notes }))).
			Prefix(text.New("备注")).FullWidth(),
		input.New("name@example").Invalid(true).FullWidth(),
		input.New("只读内容").ReadOnly(true).FullWidth(),
		input.New("禁用内容").Disabled(true).FullWidth(),
	).Gap(16)
}

func field(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(14), child).Gap(6)
}

func demoCheckbox(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		checkbox.New("接收项目更新").Key("checkbox-main").Checked(m.checked).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.checked })),
		checkbox.New("部分选中").Key("checkbox-partial").Checked(false).Indeterminate(true),
		checkbox.New("已选中 · 禁用").Key("checkbox-disabled").Checked(true).Disabled(true),
		text.New(fmt.Sprintf("接收更新：%t", m.checked)),
	).Gap(16)
}

func demoSwitch(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		switchcontrol.New("自动同步").Key("switch-sync").Checked(m.enabled).
			OnChange(setBool(send, func(m *demoState) *bool { return &m.enabled })),
		switchcontrol.New("不可用的设置").Key("switch-disabled").Checked(false).Disabled(true),
		text.New(fmt.Sprintf("自动同步：%t", m.enabled)),
	).Gap(16)
}

func demoRadio(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []radiogroup.RadioItem{
		{Key: "desktop", Label: "桌面应用"}, {Key: "web", Label: "Web 应用"},
		{Key: "mobile", Label: "移动应用"},
	}
	changed := setString(send, func(m *demoState) *string { return &m.choice })
	return layout.Column(
		radiogroup.New(items).Key("radio-vertical").SelectedKey(m.choice).OnChange(changed),
		layout.Divider(),
		radiogroup.New(items).Key("radio-horizontal").SelectedKey(m.choice).Horizontal().OnChange(changed),
		text.New("当前选择："+m.choice),
	).Gap(20)
}

func languageItems() []selectfield.SelectItem {
	return []selectfield.SelectItem{
		{Key: "go", Label: "Go", Description: "原生桌面开发"},
		{Key: "rust", Label: "Rust", Description: "系统开发"},
		{Key: "typescript", Label: "TypeScript", Description: "Web 开发"},
		{Key: "python", Label: "Python", Description: "数据分析"},
		{Key: "legacy", Label: "旧版语言（禁用）", Disabled: true},
	}
}

func demoSelect(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := setString(send, func(m *demoState) *string { return &m.combo })
	return layout.Column(
		field("单选", selectfield.New(languageItems()).Key("select-single").SelectedKey(m.combo).SelectedKey(m.combo).
			FullWidth().OnChange(changed)),
		field("多选", selectfield.New(languageItems()).Key("select-multi").SelectedKeys(m.keys).FullWidth().
			OnSelectionChange(func(keys []string) { send(demoMsg(func(m *demoState) { m.keys = keys })) })),
		field("分组选项", selectfield.NewSections([]selectfield.SelectSection{
			{Title: "编译型", Items: languageItems()[:2]},
			{Title: "脚本与应用", Items: languageItems()[2:]},
		}).Key("select-sections").SelectedKey(m.combo).SelectedKey(m.combo).FullWidth().OnChange(changed)),
		text.New("多选值："+strings.Join(m.keys, ", ")),
	).Gap(20)
}

func demoCombo(m demoState, send ui.Send[Msg]) ui.Widget {
	items := []combobox.Item{}
	for _, item := range languageItems() {
		items = append(items, combobox.Item{
			Key: item.Key, Label: item.Label, Description: item.Description, Disabled: item.Disabled,
		})
	}
	return layout.Column(
		combobox.New(items).Key("combo-language").SelectedKey(m.combo).SelectedKey(m.combo).
			InputValue(m.comboInput).Hint("搜索语言，例如 Go").FullWidth().
			OnInputChange(setString(send, func(m *demoState) *string { return &m.comboInput })).
			OnChange(func(value string) {
				send(demoMsg(func(m *demoState) {
					m.combo = value
					for _, item := range items {
						if item.Key == value {
							m.comboInput = item.Label
							break
						}
					}
				}))
			}),
		text.New("选中值："+m.combo),
	).Gap(16)
}

func demoDate(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := func(value time.Time) { send(demoMsg(func(m *demoState) { m.date = value })) }
	return layout.Column(
		field("日历选择", datepicker.New("date-picker", m.date).Label("日期").
			Locale(datepicker.Chinese()).OnChange(changed)),
		field("可编辑日期", datepicker.DateField("date-field", m.date).Label("日期字段").OnChange(changed)),
		field("日期范围", datepicker.DateRangePicker("date-range", m.dateRange).Label("计划周期").
			Locale(datepicker.Chinese()).OnChange(func(value datepicker.DateRange) {
			send(demoMsg(func(m *demoState) { m.dateRange = value }))
		})),
		text.New("所选日期："+m.date.Format("2006-01-02")),
	).Gap(20)
}

func demoColor(m demoState, send ui.Send[Msg]) ui.Widget {
	changed := func(value color.NRGBA) { send(demoMsg(func(m *demoState) { m.color = value })) }
	colors := []color.NRGBA{
		{R: 62, G: 99, B: 221, A: 255}, {R: 48, G: 164, B: 108, A: 255},
		{R: 229, G: 72, B: 77, A: 255}, {R: 245, G: 164, B: 36, A: 255},
	}
	return layout.Column(
		layout.Wrap(
			colorpicker.New("color-picker", m.color).ShowRGB(true).ShowHistory(true).OnChange(changed),
			colorpicker.ColorSwatch(m.color),
			text.New(fmt.Sprintf("#%02X%02X%02X", m.color.R, m.color.G, m.color.B)),
		).Gap(12).AlignMiddle(),
		layout.Wrap(
			colorpicker.ColorArea("color-area", m.color).OnChange(changed),
			colorpicker.ColorWheel("color-wheel", m.color).Size(180).OnChange(changed),
		).Gap(20).LineGap(16),
		colorpicker.ColorSlider("color-hue", m.color, colorpicker.ColorChannelHue).OnChange(changed),
		colorpicker.ColorSlider("color-alpha", m.color, colorpicker.ColorChannelAlpha).OnChange(changed),
		colorpicker.ColorField("color-field", m.color).OnChange(changed),
		colorpicker.ColorSwatchPicker("color-swatches", m.color, colors).OnChange(changed),
	).Gap(20)
}

func demoSlider(m demoState, send ui.Send[Msg]) ui.Widget {
	return layout.Column(
		slider.New("slider-main", m.value).Label("音量").Range(0, 100).Step(1).ShowValue().OnChange(setNumber(send)),
		slider.RangeSlider("slider-range", m.lower, m.upper).Label("价格范围").Range(0, 100).
			OnRangeChange(func(lower, upper float64) {
				send(demoMsg(func(m *demoState) { m.lower, m.upper = lower, upper }))
			}),
		text.New(fmt.Sprintf("音量 %.0f · 范围 %.0f–%.0f", m.value, m.lower, m.upper)),
		slider.New("slider-disabled", 40).Label("不可用").Disabled(true),
	).Gap(24)
}
