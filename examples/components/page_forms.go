package main

import (
	"image/color"
	"time"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/checkbox"
	"github.com/qianniancn/flowui/widget/colorpicker"
	"github.com/qianniancn/flowui/widget/combobox"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/listbox"
	"github.com/qianniancn/flowui/widget/radiogroup"
	"github.com/qianniancn/flowui/widget/selectfield"
	"github.com/qianniancn/flowui/widget/slider"
	"github.com/qianniancn/flowui/widget/switchcontrol"
	"github.com/qianniancn/flowui/widget/text"
)

func textFieldsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Text fields",
		demoSection{Title: "Input", Content: demoPanel(layout.Wrap(
			layout.Box(layout.Column(
				input.New(model.InputValue).Key("catalog-name").
					Label("Project name").FullWidth().
					OnChange(func(value string) { send(func(model *Model) { model.InputValue = value }) }),
				description.New("Public display name").For("catalog-name"),
			).Gap(6)).Style(style.New().Width(320)),
			layout.Box(input.New(model.EmailValue).Key("catalog-email").
				Label("Email").Type(input.Email).Variant(input.Secondary).FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.EmailValue = value }) })).Style(style.New().Width(320)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "TextArea", Content: demoPanel(
			layout.Box(input.NewTextArea(model.TextAreaValue).Key("catalog-textarea").
				Label("Notes").Rows(5).MaxLength(500).FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.TextAreaValue = value }) })).Style(style.New().Width(520)),
		)},
		demoSection{Title: "InputGroup", Content: demoPanel(layout.Column(
			layout.Box(input.Group(
				input.New(model.InputValue).Key("catalog-search").
					OnChange(func(value string) { send(func(model *Model) { model.InputValue = value }) }),
			).Prefix(icon.New(lucide.Search).Size(16)).
				SuffixAction(input.GroupAction("catalog-search-action", "Submit search").Content(icon.New(lucide.ArrowRight).Size(15)).OnClick(func() {
					send(func(model *Model) { model.LastAction = "Search submitted" })
				})).FullWidth()).Style(style.New().Width(420)),
			text.New(model.LastAction).Size(12),
			layout.Box(input.GroupTextArea(
				input.NewTextArea(model.TextAreaValue).Key("catalog-prompt").Rows(3).
					OnChange(func(value string) { send(func(model *Model) { model.TextAreaValue = value }) }),
			).Prefix(icon.New(lucide.MessageSquare).Size(16)).FullWidth()).Style(style.New().Width(520)),
		).Gap(14))},
	)
}

func selectionPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Selection controls",
		demoSection{Title: "Checkbox", Content: demoPanel(layout.Column(
			checkbox.New("Enable automatic updates").Key("catalog-checkbox").Checked(model.Checked).
				Description("Install verified updates when available").
				OnChange(func(value bool) { send(func(model *Model) { model.Checked = value }) }),
			checkbox.New("Secondary checkbox").Key("catalog-checkbox-secondary").Checked(true).Variant(checkbox.Secondary),
			checkbox.New("Disabled checkbox").Key("catalog-checkbox-disabled").Checked(false).Disabled(true),
		).Gap(14))},
		demoSection{Title: "Switch & SwitchGroup", Content: demoPanel(layout.Column(
			switchcontrol.New("Notifications").Key("catalog-switch").Checked(model.SwitchOn).
				Description("Product and security updates").
				OnChange(func(value bool) { send(func(model *Model) { model.SwitchOn = value }) }),
			switchcontrol.SwitchGroup(
				switchcontrol.New("Small").Key("catalog-switch-small").Checked(true).Size(switchcontrol.SwitchSmall),
				switchcontrol.New("Medium").Key("catalog-switch-medium").Checked(false),
				switchcontrol.New("Large").Key("catalog-switch-large").Checked(true).Size(switchcontrol.SwitchLarge),
			).Horizontal(),
		).Gap(16))},
		demoSection{Title: "RadioGroup", Content: demoPanel(
			radiogroup.New([]radiogroup.RadioItem{
				{Key: "starter", Label: "Starter", Description: "Small projects"},
				{Key: "pro", Label: "Pro", Description: "Production applications"},
				{Key: "team", Label: "Team", Description: "Shared workspaces"},
			}).Key("catalog-radio").SelectedKey(model.RadioValue).OnChange(func(value string) { send(func(model *Model) { model.RadioValue = value }) }),
		)},
	)
}

func choiceFieldsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Choice fields",
		demoSection{Title: "Select & SelectMultiple", Content: demoPanel(layout.Wrap(
			layout.Box(selectfield.New(catalogSelectItems).Key("catalog-select").SelectedKey(model.SelectValue).
				Label("Discipline").Description("Primary project discipline").FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.SelectValue = value }) })).Style(style.New().Width(320)),
			layout.Box(selectfield.New(catalogSelectItems).Key("catalog-select-multiple").SelectedKeys(model.SelectValues).
				Label("Disciplines").Description("Choose one or more disciplines").FullWidth().
				OnSelectionChange(func(values []string) {
					send(func(model *Model) { model.SelectValues = append([]string(nil), values...) })
				})).Style(style.New().Width(360)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "SelectSections & SelectMultipleSections", Content: demoPanel(layout.Wrap(
			layout.Box(selectfield.NewSections(catalogSelectSections).Key("catalog-select-sections").SelectedKey(model.SectionSelectValue).
				Label("Owner").FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.SectionSelectValue = value }) })).Style(style.New().Width(320)),
			layout.Box(selectfield.NewSections(catalogSelectSections).Key("catalog-select-multiple-sections").SelectedKeys(model.SectionSelectValues).
				Label("Reviewers").FullWidth().
				OnSelectionChange(func(values []string) {
					send(func(model *Model) { model.SectionSelectValues = append([]string(nil), values...) })
				})).Style(style.New().Width(360)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "ComboBox", Content: demoPanel(
			layout.Box(combobox.New(catalogComboItems).Key("catalog-combobox").SelectedKey(model.ComboValue).
				Hint("Search city").FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.ComboValue = value }) })).Style(style.New().Width(320)),
		)},
		demoSection{Title: "ListBox & ListBoxMultiple", Content: demoPanel(layout.Wrap(
			layout.Box(listbox.New(catalogListItems).Key("catalog-listbox").SelectedKey(model.ListValue).
				AllowEmptySelection().FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.ListValue = value }) })).Style(style.New().Width(360)),
			layout.Box(listbox.New(catalogListItems).Key("catalog-listbox-multiple").SelectedKeys(model.ListValues).
				FullWidth().
				OnSelectionChange(func(values []string) {
					send(func(model *Model) { model.ListValues = append([]string(nil), values...) })
				})).Style(style.New().Width(400)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "ListBoxSections & ListBoxMultipleSections", Content: demoPanel(layout.Wrap(
			layout.Box(listbox.Sections("catalog-listbox-sections", model.SectionListValue, catalogListSections).
				FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.SectionListValue = value }) })).Style(style.New().Width(360)),
			layout.Box(listbox.MultipleSections("catalog-listbox-multiple-sections", model.SectionListValues, catalogListSections).
				FullWidth().
				OnSelectionChange(func(values []string) {
					send(func(model *Model) { model.SectionListValues = append([]string(nil), values...) })
				})).Style(style.New().Width(400)),
		).Gap(18).LineGap(18))},
	)
}

func datesPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	today := dateOnly(time.Now())
	datePickerChange := func(update func(*Model, time.Time)) func(time.Time) {
		return func(value time.Time) {
			send(func(model *Model) { update(model, value) })
		}
	}

	return demoPage("Date controls",
		demoSection{Title: "DatePicker", Content: demoPanel(layout.Wrap(
			layout.Box(datepicker.New("catalog-date-picker", model.DateValue).
				Label("Release date").
				Description("Editable date field with a calendar popup").
				FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DateValue = value }))).Style(style.New().Width(320)),
			layout.Box(datepicker.New("catalog-date-picker-only", model.DatePickerOnly).
				Label("Calendar only").
				Description("Disable text editing and choose from the calendar").
				Editable(false).FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DatePickerOnly = value }))).Style(style.New().Width(320)),
			layout.Box(datepicker.New("catalog-date-picker-chinese", model.DatePickerChinese).
				Label("Chinese locale").
				Locale(datepicker.Chinese()).
				FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DatePickerChinese = value }))).Style(style.New().Width(320)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "DatePicker constraints", Content: demoPanel(layout.Wrap(
			layout.Box(datepicker.New("catalog-date-picker-limited", model.DatePickerLimited).
				Label("Next 30 days").
				Description("Only dates from today through the next 30 days are available").
				MinDate(today).MaxDate(today.AddDate(0, 0, 30)).FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DatePickerLimited = value }))).Style(style.New().Width(320)),
			layout.Box(datepicker.New("catalog-date-picker-default", time.Time{}).
				Label("Default value").
				Description("Seeds an uncontrolled field with today's date").
				DefaultValue(today).FullWidth()).Style(style.New().Width(320)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "DatePicker states", Content: demoPanel(layout.Wrap(
			layout.Box(datepicker.New("catalog-date-picker-disabled", today).
				Label("Disabled date").Disabled(true).FullWidth()).Style(style.New().Width(320)),
			layout.Box(datepicker.New("catalog-date-picker-invalid", model.DatePickerInvalid).
				Label("Required date").
				Description("An empty value is marked invalid until selected").
				Required(true).Invalid(model.DatePickerInvalid.IsZero()).
				ErrorMessage("Choose a valid date").FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DatePickerInvalid = value }))).Style(style.New().Width(320)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "DateField", Content: demoPanel(
			layout.Box(datepicker.DateField("catalog-date-field", model.DateField).
				Label("Review date").FullWidth().
				OnChange(func(value time.Time) { send(func(model *Model) { model.DateField = value }) })).Style(style.New().Width(320)),
		)},
		demoSection{Title: "DateRangePicker", Content: demoPanel(
			layout.Box(datepicker.DateRangePicker("catalog-date-range", model.DateRange).
				Label("Project window").FullWidth().
				OnChange(func(value datepicker.DateRange) { send(func(model *Model) { model.DateRange = value }) })).Style(style.New().Width(380)),
		)},
	)
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func colorsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	changed := func(value color.NRGBA) { send(func(model *Model) { model.ColorValue = value }) }
	return demoPage("Color controls",
		demoSection{Title: "ColorPicker", Content: demoPanel(
			colorpicker.New("catalog-color-picker", model.ColorValue).Label("Accent color").Presets(catalogColors).ShowField().OnChange(changed),
		)},
		demoSection{Title: "ColorWheel", Content: demoPanel(
			layout.Box(layout.Row(
				layout.Box(colorpicker.ColorWheel("catalog-color-wheel", model.ColorValue).Size(190).Label("Accent color wheel").OnChange(changed)).
					Style(style.New().Width(190).Height(190)),
				layout.Box(layout.Column(
					colorpicker.ColorSwatch(model.ColorValue).Size(colorpicker.ColorSwatchExtraLarge).Shape(colorpicker.ColorSwatchSquare),
					text.New(swatchLabel(model.ColorValue)).Size(12),
				).Gap(8)).Style(style.New().Width(64)).Align(layout.AlignCenter),
			).AlignMiddle().Gap(18)).
				Style(style.New().FillWidth().Height(190)).Align(layout.AlignCenter),
		)},
		demoSection{Title: "ColorArea", Content: demoPanel(
			layout.Box(layout.Row(
				layout.Box(colorpicker.ColorArea("catalog-color-area", model.ColorValue).ShowDots(true).OnChange(changed)).
					Style(style.New().Width(224).Height(224)),
				layout.Box(layout.Column(colorpicker.ColorSwatch(model.ColorValue).Size(colorpicker.ColorSwatchExtraLarge), text.New(swatchLabel(model.ColorValue)).Size(12)).
					Gap(8)).Style(style.New().Width(64)).Align(layout.AlignCenter),
			).AlignMiddle().Gap(18)).
				Style(style.New().FillWidth().Height(224)).Align(layout.AlignCenter),
		)},
		demoSection{Title: "ColorField & ColorSlider", Content: demoPanel(layout.Column(
			layout.Box(colorpicker.ColorField("catalog-color-field", model.ColorValue).Label("Hex color").Swatch(true).FullWidth().OnChange(changed)).Style(style.New().Width(300)),
			layout.Box(colorpicker.ColorSlider("catalog-color-hue", model.ColorValue, colorpicker.ColorChannelHue).OnChange(changed)).Style(style.New().Width(360)),
			layout.Box(colorpicker.ColorSlider("catalog-color-alpha", model.ColorValue, colorpicker.ColorChannelAlpha).OnChange(changed)).Style(style.New().Width(360)),
		).Gap(12))},
		demoSection{Title: "ColorSwatch & ColorSwatchPicker", Content: demoPanel(layout.Column(
			demoRow(
				colorpicker.ColorSwatch(model.ColorValue).Size(colorpicker.ColorSwatchExtraSmall),
				colorpicker.ColorSwatch(model.ColorValue),
				colorpicker.ColorSwatch(model.ColorValue).Size(colorpicker.ColorSwatchExtraLarge).Shape(colorpicker.ColorSwatchSquare),
			),
			colorpicker.ColorSwatchPicker("catalog-swatches", model.ColorValue, catalogColors).OnChange(changed),
		).Gap(14))},
	)
}

func slidersPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Sliders",
		demoSection{Title: "Slider", Content: demoPanel(
			layout.Box(slider.New("catalog-slider", model.SliderValue).Label("Volume").Range(0, 100).ShowValue().
				OnChange(func(value float64) { send(func(model *Model) { model.SliderValue = value }) })).Style(style.New().Width(520)),
		)},
		demoSection{Title: "RangeSlider", Content: demoPanel(
			layout.Box(slider.RangeSlider("catalog-range-slider", model.RangeLower, model.RangeUpper).Label("Price range").Range(0, 100).ShowValue().
				OnRangeChange(func(lower, upper float64) {
					send(func(model *Model) { model.RangeLower, model.RangeUpper = lower, upper })
				})).Style(style.New().Width(520)),
		)},
		demoSection{Title: "Vertical & disabled", Content: demoPanel(demoRow(
			layout.Box(slider.New("catalog-vertical-slider", model.SliderValue).Vertical().ShowValue().
				OnChange(func(value float64) { send(func(model *Model) { model.SliderValue = value }) })).Style(style.New().Width(84).Height(220)),
			layout.Box(slider.New("catalog-disabled-slider", 68).Label("Storage").ShowValue().Disabled(true)).Style(style.New().Width(360)),
		))},
	)
}

var catalogSelectItems = []selectfield.SelectItem{
	{Key: "design", Label: "Design"},
	{Key: "engineering", Label: "Engineering"},
	{Key: "product", Label: "Product"},
}

var catalogSelectSections = []selectfield.SelectSection{
	{Title: "Product", Items: []selectfield.SelectItem{
		{Key: "design", Label: "Design"},
		{Key: "product", Label: "Product"},
	}},
	{Title: "Engineering", Items: []selectfield.SelectItem{
		{Key: "engineering", Label: "Engineering"},
		{Key: "quality", Label: "Quality assurance"},
	}},
}

var catalogComboItems = []combobox.Item{
	{Key: "tokyo", Label: "Tokyo"},
	{Key: "shanghai", Label: "Shanghai"},
	{Key: "singapore", Label: "Singapore"},
	{Key: "berlin", Label: "Berlin"},
}

var catalogListItems = []listbox.Item{
	{Key: "alpha", Label: "Alpha", Description: "Early access"},
	{Key: "beta", Label: "Beta", Description: "Feature complete"},
	{Key: "stable", Label: "Stable", Description: "Production ready"},
}

var catalogListSections = []listbox.Section{
	{Title: "Release channels", Items: []listbox.Item{
		{Key: "alpha", Label: "Alpha", Description: "Early access"},
		{Key: "beta", Label: "Beta", Description: "Feature complete"},
	}},
	{Title: "Production", Items: []listbox.Item{
		{Key: "stable", Label: "Stable", Description: "Production ready"},
		{Key: "lts", Label: "LTS", Description: "Long-term support"},
	}},
}

var catalogColors = []color.NRGBA{
	{R: 0x3d, G: 0x63, B: 0xdd, A: 0xff},
	{R: 0x0d, G: 0x94, B: 0x88, A: 0xff},
	{R: 0x16, G: 0xa3, B: 0x4a, A: 0xff},
	{R: 0xe1, G: 0x1d, B: 0x48, A: 0xff},
	{R: 0x93, G: 0x33, B: 0xea, A: 0xff},
	{R: 0xea, G: 0x58, B: 0x0c, A: 0xff},
}
