package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/tabs"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Primary           string
	Secondary         string
	Vertical          string
	SecondaryVertical string
	Custom            string
	Overflow          string
	Lifecycle         string
}

type TabGroup int

const (
	primaryTabs TabGroup = iota
	secondaryTabs
	verticalTabs
	secondaryVerticalTabs
	customTabs
	overflowTabs
	lifecycleTabs
)

type SelectTab struct {
	Group TabGroup
	Key   string
}

func Update(model *Model, msg SelectTab) ui.Cmd[SelectTab] {
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
	case lifecycleTabs:
		model.Lifecycle = msg.Key
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[SelectTab]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Tabs").Size(24),
				layout.Divider(),
				section("Primary",
					layout.Box(
						tabs.New("primary", model.Primary, projectTabs()).
							Separators(true).
							OnChange(func(key string) { send(SelectTab{Group: primaryTabs, Key: key}) }),
					).Style(style.New().Width(620)),
				),
				section("Secondary",
					layout.Box(
						tabs.New("secondary", model.Secondary, projectTabs()).
							Variant(tabs.Secondary).
							IndicatorWidth(48).
							IndicatorAlign(tabs.IndicatorCenter).
							OnChange(func(key string) { send(SelectTab{Group: secondaryTabs, Key: key}) }),
					).Style(style.New().Width(620)),
				),
				section("Vertical and disabled",
					layout.Box(
						tabs.New("vertical", model.Vertical, settingsTabs()).
							Vertical().
							Separators(true).
							OnChange(func(key string) { send(SelectTab{Group: verticalTabs, Key: key}) }),
					).Style(style.New().Width(620).Height(240)),
				),
				section("Secondary vertical",
					layout.Box(
						tabs.New("secondary-vertical", model.SecondaryVertical, secondaryVerticalItems()).
							Variant(tabs.Secondary).
							Vertical().
							OnChange(func(key string) { send(SelectTab{Group: secondaryVerticalTabs, Key: key}) }),
					).Style(style.New().Width(620).Height(240)),
				),
				section("Custom styles",
					layout.Box(
						layout.Row(
							tabs.New("custom", model.Custom, customStyleItems()).
								Size(tabs.Small).
								Color(tabs.ColorAccent).
								Fit().
								Activation(tabs.ActivationManual).
								OnChange(func(key string) { send(SelectTab{Group: customTabs, Key: key}) }),
						),
					).Style(style.New().Width(380)),
				),
				section("Overflow",
					layout.Box(
						tabs.New("overflow", model.Overflow, overflowTabsItems()).
							Overflow(tabs.OverflowMenu).
							OverflowTrigger(icon.New(lucide.Ellipsis).Size(16)).
							MoreLabel("More tabs").
							OnChange(func(key string) { send(SelectTab{Group: overflowTabs, Key: key}) }),
					).Style(style.New().Width(460)),
				),
				section("Slots and panel lifecycle",
					layout.Box(
						tabs.New("lifecycle", model.Lifecycle, lifecycleTabsItems()).
							Variant(tabs.Secondary).
							Size(tabs.Large).
							Leading(icon.New(lucide.LayoutDashboard).Size(16)).
							Trailing(text.New("3 tabs").Size(12)).
							KeepAlive(true).
							ForceRender(true).
							PanelTransition(tabs.PanelFade).
							OnChange(func(key string) { send(SelectTab{Group: lifecycleTabs, Key: key}) }),
					).Style(style.New().Width(620)),
				),
			).Gap(24),
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

func projectTabs() []tabs.TabItem {
	return []tabs.TabItem{
		{Key: "overview", Label: "Overview", Panel: panel("Project overview", "View recent activity and current delivery status.")},
		{Key: "analytics", Label: "Analytics", Panel: panel("Analytics", "Track engagement, conversion, and retention metrics.")},
		{Key: "reports", Label: "Reports", Panel: panel("Reports", "Generate and review detailed project reports.")},
	}
}

func settingsTabs() []tabs.TabItem {
	return []tabs.TabItem{
		{Key: "account", Label: "Account", Panel: panel("Account", "Manage profile information and account preferences.")},
		{Key: "security", Label: "Security", Panel: panel("Security", "Review authentication and active sessions.")},
		{Key: "notifications", Label: "Notifications", Panel: panel("Notifications", "Choose which product updates you receive.")},
		{Key: "billing", Label: "Billing", Panel: panel("Billing", "This section is unavailable for the current workspace."), Disabled: true},
	}
}

func secondaryVerticalItems() []tabs.TabItem {
	return []tabs.TabItem{
		{Key: "account", Label: "Account", Panel: panel("Account settings", "Manage account information and preferences.")},
		{Key: "security", Label: "Security", Panel: panel("Security settings", "Configure authentication and password settings.")},
		{Key: "notifications", Label: "Notifications", Panel: panel("Notification preferences", "Choose how and when notifications arrive.")},
		{Key: "billing", Label: "Billing", Panel: panel("Billing information", "View and manage subscription details.")},
	}
}

func customStyleItems() []tabs.TabItem {
	return []tabs.TabItem{
		{Key: "daily", Label: "Daily"},
		{Key: "weekly", Label: "Weekly"},
		{Key: "bi-weekly", Label: "Bi-Weekly"},
		{Key: "monthly", Label: "Monthly"},
	}
}

func overflowTabsItems() []tabs.TabItem {
	labels := []string{"Overview", "Analytics", "Reports", "Performance", "Engagement", "Conversions", "Revenue", "Retention"}
	items := make([]tabs.TabItem, 0, len(labels))
	for _, label := range labels {
		items = append(items, tabs.TabItem{
			Key:   label,
			Label: label,
			Panel: panel(label, "Detailed information for the selected category."),
		})
	}
	return items
}

func panel(title, body string) ui.Widget {
	return layout.Column(
		text.New(title).Size(16),
		description.New(body),
	).Gap(6)
}

func main() {
	ui.Run(ui.NewProgram(Model{
		Primary:           "overview",
		Secondary:         "analytics",
		Vertical:          "account",
		SecondaryVertical: "security",
		Custom:            "weekly",
		Overflow:          "overview",
		Lifecycle:         "account",
	},
		Update, View), window.Title("FlowUI Tabs"),
		window.Size(940, 860),
		window.Theme(base.Light()),
	)
}

func lifecycleTabsItems() []tabs.TabItem {
	return []tabs.TabItem{
		{Key: "account", Label: "Account", Leading: icon.New(lucide.UserRound).Size(16), Panel: panel("Account", "This panel keeps its widget state while hidden.")},
		{Key: "security", Label: "Security", Leading: icon.New(lucide.ShieldCheck).Size(16), Panel: panel("Security", "Authentication settings live in this panel.")},
		{Key: "billing", Label: "Billing", Leading: icon.New(lucide.CreditCard).Size(16), Panel: panel("Billing", "Billing history and invoices.")},
	}
}
