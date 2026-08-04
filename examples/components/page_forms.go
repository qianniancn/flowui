package main

import (
	"image/color"
	"time"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

func textFieldsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Text fields",
		demoSection{Title: "Input", Content: demoPanel(ui.Wrap(
			ui.Box(ui.Column(
				ui.Input("catalog-name", model.InputValue).
					Label("Project name").FullWidth().
					OnChange(func(value string) { send(func(model *Model) { model.InputValue = value }) }),
				ui.Description("Public display name").For("catalog-name"),
			).Gap(6)).Style(ui.Width(320)),
			ui.Box(ui.Input("catalog-email", model.EmailValue).
				Label("Email").Type(ui.InputEmail).Variant(ui.InputSecondary).FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.EmailValue = value }) })).Style(ui.Width(320)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "TextArea", Content: demoPanel(
			ui.Box(ui.TextArea("catalog-textarea", model.TextAreaValue).
				Label("Notes").Rows(5).MaxLength(500).FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.TextAreaValue = value }) })).Style(ui.Width(520)),
		)},
		demoSection{Title: "InputGroup", Content: demoPanel(ui.Column(
			ui.Box(ui.InputGroup(
				ui.Input("catalog-search", model.InputValue).
					OnChange(func(value string) { send(func(model *Model) { model.InputValue = value }) }),
			).Prefix(ui.Icon(lucide.Search).Size(16)).
				SuffixAction(ui.InputGroupAction(
					"catalog-search-action", "Submit search", ui.Icon(lucide.ArrowRight).Size(15),
				).OnClick(func() {
					send(func(model *Model) { model.LastAction = "Search submitted" })
				})).FullWidth()).Style(ui.Width(420)),
			ui.Text(model.LastAction).Size(12),
			ui.Box(ui.InputGroupTextArea(
				ui.TextArea("catalog-prompt", model.TextAreaValue).Rows(3).
					OnChange(func(value string) { send(func(model *Model) { model.TextAreaValue = value }) }),
			).Prefix(ui.Icon(lucide.MessageSquare).Size(16)).FullWidth()).Style(ui.Width(520)),
		).Gap(14))},
	)
}

func selectionPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Selection controls",
		demoSection{Title: "Checkbox", Content: demoPanel(ui.Column(
			ui.Checkbox("catalog-checkbox", model.Checked, "Enable automatic updates").
				Description("Install verified updates when available").
				OnChange(func(value bool) { send(func(model *Model) { model.Checked = value }) }),
			ui.Checkbox("catalog-checkbox-secondary", true, "Secondary checkbox").Variant(ui.CheckboxSecondary),
			ui.Checkbox("catalog-checkbox-disabled", false, "Disabled checkbox").Disabled(true),
		).Gap(14))},
		demoSection{Title: "Switch & SwitchGroup", Content: demoPanel(ui.Column(
			ui.Switch("catalog-switch", model.SwitchOn, "Notifications").
				Description("Product and security updates").
				OnChange(func(value bool) { send(func(model *Model) { model.SwitchOn = value }) }),
			ui.SwitchGroup(
				ui.Switch("catalog-switch-small", true, "Small").Size(ui.SwitchSmall),
				ui.Switch("catalog-switch-medium", false, "Medium"),
				ui.Switch("catalog-switch-large", true, "Large").Size(ui.SwitchLarge),
			).Horizontal(),
		).Gap(16))},
		demoSection{Title: "RadioGroup", Content: demoPanel(
			ui.RadioGroup("catalog-radio", model.RadioValue, []ui.RadioItem{
				{Key: "starter", Label: "Starter", Description: "Small projects"},
				{Key: "pro", Label: "Pro", Description: "Production applications"},
				{Key: "team", Label: "Team", Description: "Shared workspaces"},
			}).OnChange(func(value string) { send(func(model *Model) { model.RadioValue = value }) }),
		)},
	)
}

func choiceFieldsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Choice fields",
		demoSection{Title: "Select & SelectMultiple", Content: demoPanel(ui.Wrap(
			ui.Box(ui.Select("catalog-select", model.SelectValue, catalogSelectItems).
				Label("Discipline").Description("Primary project discipline").FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.SelectValue = value }) })).Style(ui.Width(320)),
			ui.Box(ui.SelectMultiple("catalog-select-multiple", model.SelectValues, catalogSelectItems).
				Label("Disciplines").Description("Choose one or more disciplines").FullWidth().
				OnSelectionChange(func(values []string) {
					send(func(model *Model) { model.SelectValues = append([]string(nil), values...) })
				})).Style(ui.Width(360)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "SelectSections & SelectMultipleSections", Content: demoPanel(ui.Wrap(
			ui.Box(ui.SelectSections("catalog-select-sections", model.SectionSelectValue, catalogSelectSections).
				Label("Owner").FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.SectionSelectValue = value }) })).Style(ui.Width(320)),
			ui.Box(ui.SelectMultipleSections("catalog-select-multiple-sections", model.SectionSelectValues, catalogSelectSections).
				Label("Reviewers").FullWidth().
				OnSelectionChange(func(values []string) {
					send(func(model *Model) { model.SectionSelectValues = append([]string(nil), values...) })
				})).Style(ui.Width(360)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "ComboBox", Content: demoPanel(
			ui.Box(ui.ComboBox("catalog-combobox", model.ComboValue, catalogComboItems).
				Hint("Search city").FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.ComboValue = value }) })).Style(ui.Width(320)),
		)},
		demoSection{Title: "ListBox & ListBoxMultiple", Content: demoPanel(ui.Wrap(
			ui.Box(ui.ListBox("catalog-listbox", model.ListValue, catalogListItems).
				AllowEmptySelection().FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.ListValue = value }) })).Style(ui.Width(360)),
			ui.Box(ui.ListBoxMultiple("catalog-listbox-multiple", model.ListValues, catalogListItems).
				FullWidth().
				OnSelectionChange(func(values []string) {
					send(func(model *Model) { model.ListValues = append([]string(nil), values...) })
				})).Style(ui.Width(400)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "ListBoxSections & ListBoxMultipleSections", Content: demoPanel(ui.Wrap(
			ui.Box(ui.ListBoxSections("catalog-listbox-sections", model.SectionListValue, catalogListSections).
				FullWidth().
				OnChange(func(value string) { send(func(model *Model) { model.SectionListValue = value }) })).Style(ui.Width(360)),
			ui.Box(ui.ListBoxMultipleSections("catalog-listbox-multiple-sections", model.SectionListValues, catalogListSections).
				FullWidth().
				OnSelectionChange(func(values []string) {
					send(func(model *Model) { model.SectionListValues = append([]string(nil), values...) })
				})).Style(ui.Width(400)),
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
		demoSection{Title: "DatePicker", Content: demoPanel(ui.Wrap(
			ui.Box(ui.DatePicker("catalog-date-picker", model.DateValue).
				Label("Release date").
				Description("Editable date field with a calendar popup").
				FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DateValue = value }))).Style(ui.Width(320)),
			ui.Box(ui.DatePicker("catalog-date-picker-only", model.DatePickerOnly).
				Label("Calendar only").
				Description("Disable text editing and choose from the calendar").
				Editable(false).FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DatePickerOnly = value }))).Style(ui.Width(320)),
			ui.Box(ui.DatePicker("catalog-date-picker-chinese", model.DatePickerChinese).
				Label("Chinese locale").
				Locale(ui.DatePickerChinese()).
				FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DatePickerChinese = value }))).Style(ui.Width(320)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "DatePicker constraints", Content: demoPanel(ui.Wrap(
			ui.Box(ui.DatePicker("catalog-date-picker-limited", model.DatePickerLimited).
				Label("Next 30 days").
				Description("Only dates from today through the next 30 days are available").
				MinDate(today).MaxDate(today.AddDate(0, 0, 30)).FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DatePickerLimited = value }))).Style(ui.Width(320)),
			ui.Box(ui.DatePicker("catalog-date-picker-default", time.Time{}).
				Label("Default value").
				Description("Seeds an uncontrolled field with today's date").
				DefaultValue(today).FullWidth()).Style(ui.Width(320)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "DatePicker states", Content: demoPanel(ui.Wrap(
			ui.Box(ui.DatePicker("catalog-date-picker-disabled", today).
				Label("Disabled date").Disabled(true).FullWidth()).Style(ui.Width(320)),
			ui.Box(ui.DatePicker("catalog-date-picker-invalid", model.DatePickerInvalid).
				Label("Required date").
				Description("An empty value is marked invalid until selected").
				Required(true).Invalid(model.DatePickerInvalid.IsZero()).
				ErrorMessage("Choose a valid date").FullWidth().
				OnChange(datePickerChange(func(model *Model, value time.Time) { model.DatePickerInvalid = value }))).Style(ui.Width(320)),
		).Gap(18).LineGap(18))},
		demoSection{Title: "DateField", Content: demoPanel(
			ui.Box(ui.DateField("catalog-date-field", model.DateField).
				Label("Review date").FullWidth().
				OnChange(func(value time.Time) { send(func(model *Model) { model.DateField = value }) })).Style(ui.Width(320)),
		)},
		demoSection{Title: "DateRangePicker", Content: demoPanel(
			ui.Box(ui.DateRangePicker("catalog-date-range", model.DateRange).
				Label("Project window").FullWidth().
				OnChange(func(value ui.DateRange) { send(func(model *Model) { model.DateRange = value }) })).Style(ui.Width(380)),
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
			ui.ColorPicker("catalog-color-picker", model.ColorValue).Label("Accent color").Presets(catalogColors).ShowField().OnChange(changed),
		)},
		demoSection{Title: "ColorWheel", Content: demoPanel(
			ui.Box(ui.Row(
				ui.Box(ui.ColorWheel("catalog-color-wheel", model.ColorValue).Size(190).Label("Accent color wheel").OnChange(changed)).
					Style(ui.Width(190).Height(190)),
				ui.Box(ui.Column(
					ui.ColorSwatch(model.ColorValue).Size(ui.ColorSwatchLarge).Shape(ui.ColorSwatchSquare),
					ui.Text(swatchLabel(model.ColorValue)).Size(12),
				).Gap(8)).Style(ui.Width(64)).Align(ui.AlignCenter),
			).AlignMiddle().Gap(18)).
				Style(ui.FillWidth().Height(190)).Align(ui.AlignCenter),
		)},
		demoSection{Title: "ColorArea", Content: demoPanel(
			ui.Box(ui.Row(
				ui.Box(ui.ColorArea("catalog-color-area", model.ColorValue).ShowDots(true).OnChange(changed)).
					Style(ui.Width(224).Height(224)),
				ui.Box(ui.Column(ui.ColorSwatch(model.ColorValue).Size(ui.ColorSwatchLarge), ui.Text(swatchLabel(model.ColorValue)).Size(12)).
					Gap(8)).Style(ui.Width(64)).Align(ui.AlignCenter),
			).AlignMiddle().Gap(18)).
				Style(ui.FillWidth().Height(224)).Align(ui.AlignCenter),
		)},
		demoSection{Title: "ColorField & ColorSlider", Content: demoPanel(ui.Column(
			ui.Box(ui.ColorField("catalog-color-field", model.ColorValue).Label("Hex color").Swatch(true).FullWidth().OnChange(changed)).Style(ui.Width(300)),
			ui.Box(ui.ColorSlider("catalog-color-hue", model.ColorValue, ui.ColorChannelHue).OnChange(changed)).Style(ui.Width(360)),
			ui.Box(ui.ColorSlider("catalog-color-alpha", model.ColorValue, ui.ColorChannelAlpha).OnChange(changed)).Style(ui.Width(360)),
		).Gap(12))},
		demoSection{Title: "ColorSwatch & ColorSwatchPicker", Content: demoPanel(ui.Column(
			demoRow(
				ui.ColorSwatch(model.ColorValue).Size(ui.ColorSwatchSmall),
				ui.ColorSwatch(model.ColorValue),
				ui.ColorSwatch(model.ColorValue).Size(ui.ColorSwatchLarge).Shape(ui.ColorSwatchSquare),
			),
			ui.ColorSwatchPicker("catalog-swatches", model.ColorValue, catalogColors).OnChange(changed),
		).Gap(14))},
	)
}

func slidersPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return demoPage("Sliders",
		demoSection{Title: "Slider", Content: demoPanel(
			ui.Box(ui.Slider("catalog-slider", model.SliderValue).Label("Volume").Range(0, 100).ShowValue().
				OnChange(func(value float64) { send(func(model *Model) { model.SliderValue = value }) })).Style(ui.Width(520)),
		)},
		demoSection{Title: "RangeSlider", Content: demoPanel(
			ui.Box(ui.RangeSlider("catalog-range-slider", model.RangeLower, model.RangeUpper).Label("Price range").Range(0, 100).ShowValue().
				OnRangeChange(func(lower, upper float64) {
					send(func(model *Model) { model.RangeLower, model.RangeUpper = lower, upper })
				})).Style(ui.Width(520)),
		)},
		demoSection{Title: "Vertical & disabled", Content: demoPanel(demoRow(
			ui.Box(ui.Slider("catalog-vertical-slider", model.SliderValue).Vertical().ShowValue().
				OnChange(func(value float64) { send(func(model *Model) { model.SliderValue = value }) })).Style(ui.Width(84).Height(220)),
			ui.Box(ui.Slider("catalog-disabled-slider", 68).Label("Storage").ShowValue().Disabled(true)).Style(ui.Width(360)),
		))},
	)
}

var catalogSelectItems = []ui.SelectItem{
	{Key: "design", Label: "Design"},
	{Key: "engineering", Label: "Engineering"},
	{Key: "product", Label: "Product"},
}

var catalogSelectSections = []ui.SelectSection{
	{Title: "Product", Items: []ui.SelectItem{
		{Key: "design", Label: "Design"},
		{Key: "product", Label: "Product"},
	}},
	{Title: "Engineering", Items: []ui.SelectItem{
		{Key: "engineering", Label: "Engineering"},
		{Key: "quality", Label: "Quality assurance"},
	}},
}

var catalogComboItems = []ui.ComboBoxItem{
	{Key: "tokyo", Label: "Tokyo"},
	{Key: "shanghai", Label: "Shanghai"},
	{Key: "singapore", Label: "Singapore"},
	{Key: "berlin", Label: "Berlin"},
}

var catalogListItems = []ui.ListBoxItem{
	{Key: "alpha", Label: "Alpha", Description: "Early access"},
	{Key: "beta", Label: "Beta", Description: "Feature complete"},
	{Key: "stable", Label: "Stable", Description: "Production ready"},
}

var catalogListSections = []ui.ListBoxSection{
	{Title: "Release channels", Items: []ui.ListBoxItem{
		{Key: "alpha", Label: "Alpha", Description: "Early access"},
		{Key: "beta", Label: "Beta", Description: "Feature complete"},
	}},
	{Title: "Production", Items: []ui.ListBoxItem{
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
