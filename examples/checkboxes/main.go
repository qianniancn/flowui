package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/checkbox"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Checkbox").Size(24),
				layout.Divider(),
				section("Variants",
					layout.Column(
						controlled("primary", fieldPrimary, model.Primary, "Primary checkbox", send).
							Description("Standard field styling with a subtle shadow"),
						controlled("secondary", fieldSecondary, model.Secondary, "Secondary checkbox", send).
							Variant(checkbox.Secondary).
							Description("Lower emphasis styling for use on surfaces"),
					).Gap(16),
				),
				section("Indeterminate",
					layout.Column(
						controlled("select-all", fieldAll, all, "Select all notifications", send).
							Indeterminate(indeterminate).
							Description("Select or clear every notification channel"),
						layout.Box(
							layout.Column(
								controlled("email", fieldEmail, model.Email, "Email notifications", send),
								controlled("reports", fieldReports, model.Reports, "Weekly reports", send),
								controlled("security", fieldSecurity, model.Security, "Security alerts", send),
							).Gap(12),
						).Style(style.New().PaddingLeft(28)),
					).Gap(14),
				),
				section("Validation and states",
					layout.Column(
						controlled("terms", fieldTerms, model.Terms, "Accept terms", send).
							Required(true).
							Invalid(!model.Terms).
							ErrorMessage("Accept the terms before continuing"),
						checkbox.New("Read-only selection").Key("readonly").Checked(true).
							ReadOnly(true).
							Description("The value is visible but cannot be changed"),
						checkbox.New("Disabled checked").Key("disabled-checked").Checked(true).Disabled(true),
						checkbox.New("Disabled").Key("disabled").Checked(false).Disabled(true),
					).Gap(16),
				),
				section("Custom indicators",
					layout.Row(
						controlled("heart", fieldHeart, model.Heart, "Favorite", send).
							Indicator(customIndicator(lucide.Heart)),
						controlled("plus", fieldPlus, model.Plus, "Add item", send).
							Indicator(customIndicator(lucide.Plus)),
					).Gap(28).AlignMiddle(),
				),
			).Gap(22),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(680).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(18), child).Gap(10)
}

func controlled(key string, field Field, checked bool, label string, send ui.Send[Msg]) checkbox.Widget {
	return checkbox.New(label).Key(key).Checked(checked).OnChange(func(checked bool) {
		send(Msg{Field: field, Checked: checked})
	})
}

func customIndicator(data lucide.Data) func(checkbox.IndicatorState) ui.Widget {
	return func(state checkbox.IndicatorState) ui.Widget {
		if !state.Checked && !state.Indeterminate {
			return nil
		}
		return icon.New(data).Size(10)
	}
}

func main() {
	ui.Run(ui.NewProgram(Model{Primary: true, Email: true, Heart: true, Plus: true},
		Update, View), window.Title("FlowUI Checkbox"),
		window.Size(900, 760),
		window.Theme(base.Light()),
	)
}
