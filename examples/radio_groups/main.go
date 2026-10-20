package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/radiogroup"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Plan     string
	Billing  string
	Delivery string
	Support  string
	Last     string
}

type Field string

const (
	fieldPlan     Field = "plan"
	fieldBilling  Field = "billing"
	fieldDelivery Field = "delivery"
	fieldSupport  Field = "support"
)

type Msg struct {
	Field Field
	Key   string
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.Field {
	case fieldPlan:
		m.Plan = msg.Key
	case fieldBilling:
		m.Billing = msg.Key
	case fieldDelivery:
		m.Delivery = msg.Key
	case fieldSupport:
		m.Support = msg.Key
	}
	m.Last = fmt.Sprintf("%s selected %s", msg.Field, msg.Key)
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No option selected"
	if m.Last != "" {
		status = m.Last
	}

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI RadioGroup").Size(24),
				text.New(status).Size(16),
				layout.Divider(),
				section("Plan",
					layout.Box(radioGroup("plan", fieldPlan, m.Plan, plans, send)).
						Style(style.New().Width(420)),
				),
				section("Variants",
					layout.Column(
						layout.Box(radioGroup("billing", fieldBilling, m.Billing, billing, send).
							Variant(radiogroup.RadioSecondary)).
							Style(style.New().Width(420)),
						layout.Box(radiogroup.New(plans).Key("disabled").SelectedKey("pro").
							Disabled(true)).
							Style(style.New().Width(420)),
					).Gap(18),
				),
				section("Horizontal",
					layout.Box(radioGroup("delivery", fieldDelivery, m.Delivery, delivery, send).
						Horizontal()).
						Style(style.New().Width(520)),
				),
				section("Item states",
					layout.Box(radioGroup("support", fieldSupport, m.Support, support, send)).
						Style(style.New().Width(420)),
				),
			).Gap(18),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(18),
		child,
	).Gap(10)
}

func radioGroup(key string, field Field, selected string, items []radiogroup.RadioItem, send ui.Send[Msg]) radiogroup.Widget {
	return radiogroup.New(items).Key(key).SelectedKey(selected).
		OnChange(func(selected string) {
			send(Msg{
				Field: field,
				Key:   selected,
			})
		})
}

var plans = []radiogroup.RadioItem{
	{Key: "basic", Label: "Basic", Description: "For experiments and small utilities"},
	{Key: "pro", Label: "Pro", Description: "For production apps and active projects"},
	{Key: "team", Label: "Team", Description: "Shared workflows for multiple contributors"},
}

var billing = []radiogroup.RadioItem{
	{Key: "monthly", Label: "Monthly", Description: "Flexible billing with no long commitment"},
	{Key: "yearly", Label: "Yearly", Description: "Lower total cost for long-running projects"},
}

var delivery = []radiogroup.RadioItem{
	{Key: "standard", Label: "Standard"},
	{Key: "express", Label: "Express"},
	{Key: "scheduled", Label: "Scheduled"},
}

var support = []radiogroup.RadioItem{
	{Key: "community", Label: "Community", Description: "Public discussion and issue triage"},
	{Key: "priority", Label: "Priority", Description: "Faster response for launch windows"},
	{Key: "legacy", Label: "Legacy", Description: "Unavailable for new subscriptions", Disabled: true},
	{Key: "enterprise", Label: "Enterprise", Description: "Requires account approval", Invalid: true},
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI RadioGroup"),
		window.Size(900, 720),
		window.Theme(base.Light()),
	)
}
