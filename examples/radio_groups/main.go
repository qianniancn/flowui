package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
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

func Update(m *Model, msg Msg) {
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
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No option selected"
	if m.Last != "" {
		status = m.Last
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("radio-groups",
				ui.Column(
					ui.Text("FlowUI RadioGroup").Size(24),
					ui.Text(status).Size(16),
					ui.Divider(),
					section("Plan",
						ui.Box(radioGroup("plan", fieldPlan, m.Plan, plans, send)).
							Width(420),
					),
					section("Variants",
						ui.Column(
							ui.Box(radioGroup("billing", fieldBilling, m.Billing, billing, send).
								Variant(ui.RadioSecondary)).
								Width(420),
							ui.Box(ui.RadioGroup("disabled", "pro", plans).
								Disabled(true)).
								Width(420),
						).Gap(18),
					),
					section("Horizontal",
						ui.Box(radioGroup("delivery", fieldDelivery, m.Delivery, delivery, send).
							Horizontal()).
							Width(520),
					),
					section("Item states",
						ui.Box(radioGroup("support", fieldSupport, m.Support, support, send)).
							Width(420),
					),
				).Gap(18),
			).Vertical(),
		).FillWidth().MaxWidth(760).Padding(24),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func radioGroup(key string, field Field, selected string, items []ui.RadioItem, send ui.Send[Msg]) ui.RadioGroupWidget {
	return ui.RadioGroup(key, selected, items).
		OnChange(func(selected string) {
			send(Msg{
				Field: field,
				Key:   selected,
			})
		})
}

var plans = []ui.RadioItem{
	{Key: "basic", Label: "Basic", Description: "For experiments and small utilities"},
	{Key: "pro", Label: "Pro", Description: "For production apps and active projects"},
	{Key: "team", Label: "Team", Description: "Shared workflows for multiple contributors"},
}

var billing = []ui.RadioItem{
	{Key: "monthly", Label: "Monthly", Description: "Flexible billing with no long commitment"},
	{Key: "yearly", Label: "Yearly", Description: "Lower total cost for long-running projects"},
}

var delivery = []ui.RadioItem{
	{Key: "standard", Label: "Standard"},
	{Key: "express", Label: "Express"},
	{Key: "scheduled", Label: "Scheduled"},
}

var support = []ui.RadioItem{
	{Key: "community", Label: "Community", Description: "Public discussion and issue triage"},
	{Key: "priority", Label: "Priority", Description: "Faster response for launch windows"},
	{Key: "legacy", Label: "Legacy", Description: "Unavailable for new subscriptions", Disabled: true},
	{Key: "enterprise", Label: "Enterprise", Description: "Requires account approval", Invalid: true},
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI RadioGroup"),
		ui.Size(900, 720),
	)
}
