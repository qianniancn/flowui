package main

import (
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	Primary   bool
	Secondary bool
	Email     bool
	Reports   bool
	Security  bool
	Terms     bool
	Heart     bool
	Plus      bool
}

type Field string

const (
	fieldPrimary   Field = "primary"
	fieldSecondary Field = "secondary"
	fieldAll       Field = "all"
	fieldEmail     Field = "email"
	fieldReports   Field = "reports"
	fieldSecurity  Field = "security"
	fieldTerms     Field = "terms"
	fieldHeart     Field = "heart"
	fieldPlus      Field = "plus"
)

type Msg struct {
	Field   Field
	Checked bool
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	if msg.Field == fieldAll {
		model.Email = msg.Checked
		model.Reports = msg.Checked
		model.Security = msg.Checked
		return nil
	}
	switch msg.Field {
	case fieldPrimary:
		model.Primary = msg.Checked
	case fieldSecondary:
		model.Secondary = msg.Checked
	case fieldEmail:
		model.Email = msg.Checked
	case fieldReports:
		model.Reports = msg.Checked
	case fieldSecurity:
		model.Security = msg.Checked
	case fieldTerms:
		model.Terms = msg.Checked
	case fieldHeart:
		model.Heart = msg.Checked
	case fieldPlus:
		model.Plus = msg.Checked
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	selected := 0
	for _, value := range []bool{model.Email, model.Reports, model.Security} {
		if value {
			selected++
		}
	}
	all := selected == 3
	indeterminate := selected > 0 && selected < 3

	return ui.Center(
		ui.Box(
			ui.Scroll("checkboxes",
				ui.Column(
					ui.Text("FlowUI Checkbox").Size(24),
					ui.Divider(),
					section("Variants",
						ui.Column(
							controlled("primary", fieldPrimary, model.Primary, "Primary checkbox", send).
								Description("Standard field styling with a subtle shadow"),
							controlled("secondary", fieldSecondary, model.Secondary, "Secondary checkbox", send).
								Variant(ui.CheckboxSecondary).
								Description("Lower emphasis styling for use on surfaces"),
						).Gap(16),
					),
					section("Indeterminate",
						ui.Column(
							controlled("select-all", fieldAll, all, "Select all notifications", send).
								Indeterminate(indeterminate).
								Description("Select or clear every notification channel"),
							ui.Box(
								ui.Column(
									controlled("email", fieldEmail, model.Email, "Email notifications", send),
									controlled("reports", fieldReports, model.Reports, "Weekly reports", send),
									controlled("security", fieldSecurity, model.Security, "Security alerts", send),
								).Gap(12),
							).Style(ui.PaddingLeft(28)),
						).Gap(14),
					),
					section("Validation and states",
						ui.Column(
							controlled("terms", fieldTerms, model.Terms, "Accept terms", send).
								Required(true).
								Invalid(!model.Terms).
								ErrorMessage("Accept the terms before continuing"),
							ui.Checkbox("readonly", true, "Read-only selection").
								ReadOnly(true).
								Description("The value is visible but cannot be changed"),
							ui.Checkbox("disabled-checked", true, "Disabled checked").Disabled(true),
							ui.Checkbox("disabled", false, "Disabled").Disabled(true),
						).Gap(16),
					),
					section("Custom indicators",
						ui.Row(
							controlled("heart", fieldHeart, model.Heart, "Favorite", send).
								Indicator(customIndicator(lucide.Heart)),
							controlled("plus", fieldPlus, model.Plus, "Add item", send).
								Indicator(customIndicator(lucide.Plus)),
						).Gap(28).AlignMiddle(),
					),
				).Gap(22),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(680)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func controlled(key string, field Field, checked bool, label string, send ui.Send[Msg]) ui.CheckboxWidget {
	return ui.Checkbox(key, checked, label).OnChange(func(checked bool) {
		send(Msg{Field: field, Checked: checked})
	})
}

func customIndicator(data lucide.Data) func(ui.CheckboxIndicatorState) ui.Widget {
	return func(state ui.CheckboxIndicatorState) ui.Widget {
		if !state.Checked && !state.Indeterminate {
			return nil
		}
		return ui.Icon(data).Size(10)
	}
}

func main() {
	ui.Run(ui.NewProgram(Model{Primary: true, Email: true, Heart: true, Plus: true},
		Update, View), ui.Title("FlowUI Checkbox"),
		ui.Size(900, 760),
	)
}
