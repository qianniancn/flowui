package main

import "github.com/qianniancn/FlowUI/ui"

type Model struct {
	Primary           string
	Secondary         string
	Vertical          string
	SecondaryVertical string
	Custom            string
	Overflow          string
}

type TabGroup int

const (
	primaryTabs TabGroup = iota
	secondaryTabs
	verticalTabs
	secondaryVerticalTabs
	customTabs
	overflowTabs
)

type SelectTab struct {
	Group TabGroup
	Key   string
}

func Update(model *Model, msg SelectTab) {
	switch msg.Group {
	case primaryTabs:
		model.Primary = msg.Key
	case secondaryTabs:
		model.Secondary = msg.Key
	case verticalTabs:
		model.Vertical = msg.Key
	case secondaryVerticalTabs:
		model.SecondaryVertical = msg.Key
	case customTabs:
		model.Custom = msg.Key
	case overflowTabs:
		model.Overflow = msg.Key
	}
}

func View(_ *ui.Context, model Model, send ui.Send[SelectTab]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Scroll("tabs-example",
				ui.Column(
					ui.Text("FlowUI Tabs").Size(24),
					ui.Divider(),
					section("Primary",
						ui.Box(
							ui.Tabs("primary", model.Primary, projectTabs()).
								Separators(true).
								OnChange(func(key string) { send(SelectTab{Group: primaryTabs, Key: key}) }),
						).Width(620),
					),
					section("Secondary",
						ui.Box(
							ui.Tabs("secondary", model.Secondary, projectTabs()).
								Variant(ui.TabsSecondary).
								OnChange(func(key string) { send(SelectTab{Group: secondaryTabs, Key: key}) }),
						).Width(620),
					),
					section("Vertical and disabled",
						ui.Box(
							ui.Tabs("vertical", model.Vertical, settingsTabs()).
								Vertical().
								Separators(true).
								OnChange(func(key string) { send(SelectTab{Group: verticalTabs, Key: key}) }),
						).Width(620).Height(240),
					),
					section("Secondary vertical",
						ui.Box(
							ui.Tabs("secondary-vertical", model.SecondaryVertical, secondaryVerticalItems()).
								Variant(ui.TabsSecondary).
								Vertical().
								OnChange(func(key string) { send(SelectTab{Group: secondaryVerticalTabs, Key: key}) }),
						).Width(620).Height(240),
					),
					section("Custom styles",
						ui.Box(
							ui.Row(
								ui.Tabs("custom", model.Custom, customStyleItems()).
									Size(ui.TabsSmall).
									Color(ui.TabsColorAccent).
									Fit().
									OnChange(func(key string) { send(SelectTab{Group: customTabs, Key: key}) }),
							),
						).Width(380),
					),
					section("Overflow",
						ui.Box(
							ui.Tabs("overflow", model.Overflow, overflowTabsItems()).
								OnChange(func(key string) { send(SelectTab{Group: overflowTabs, Key: key}) }),
						).Width(460),
					),
				).Gap(24),
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

func projectTabs() []ui.TabItem {
	return []ui.TabItem{
		{Key: "overview", Label: "Overview", Panel: panel("Project overview", "View recent activity and current delivery status.")},
		{Key: "analytics", Label: "Analytics", Panel: panel("Analytics", "Track engagement, conversion, and retention metrics.")},
		{Key: "reports", Label: "Reports", Panel: panel("Reports", "Generate and review detailed project reports.")},
	}
}

func settingsTabs() []ui.TabItem {
	return []ui.TabItem{
		{Key: "account", Label: "Account", Panel: panel("Account", "Manage profile information and account preferences.")},
		{Key: "security", Label: "Security", Panel: panel("Security", "Review authentication and active sessions.")},
		{Key: "notifications", Label: "Notifications", Panel: panel("Notifications", "Choose which product updates you receive.")},
		{Key: "billing", Label: "Billing", Panel: panel("Billing", "This section is unavailable for the current workspace."), Disabled: true},
	}
}

func secondaryVerticalItems() []ui.TabItem {
	return []ui.TabItem{
		{Key: "account", Label: "Account", Panel: panel("Account settings", "Manage account information and preferences.")},
		{Key: "security", Label: "Security", Panel: panel("Security settings", "Configure authentication and password settings.")},
		{Key: "notifications", Label: "Notifications", Panel: panel("Notification preferences", "Choose how and when notifications arrive.")},
		{Key: "billing", Label: "Billing", Panel: panel("Billing information", "View and manage subscription details.")},
	}
}

func customStyleItems() []ui.TabItem {
	return []ui.TabItem{
		{Key: "daily", Label: "Daily"},
		{Key: "weekly", Label: "Weekly"},
		{Key: "bi-weekly", Label: "Bi-Weekly"},
		{Key: "monthly", Label: "Monthly"},
	}
}

func overflowTabsItems() []ui.TabItem {
	labels := []string{"Overview", "Analytics", "Reports", "Performance", "Engagement", "Conversions", "Revenue", "Retention"}
	items := make([]ui.TabItem, 0, len(labels))
	for _, label := range labels {
		items = append(items, ui.TabItem{
			Key:   label,
			Label: label,
			Panel: panel(label, "Detailed information for the selected category."),
		})
	}
	return items
}

func panel(title, body string) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(16),
		ui.Description(body),
	).Gap(6)
}

func main() {
	ui.Run(Model{
		Primary:           "overview",
		Secondary:         "analytics",
		Vertical:          "account",
		SecondaryVertical: "security",
		Custom:            "weekly",
		Overflow:          "overview",
	}, Update, View,
		ui.Title("FlowUI Tabs"),
		ui.Size(940, 860),
	)
}
