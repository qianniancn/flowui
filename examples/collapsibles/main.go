package main

import (
	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	Standalone bool
	Single     []string
	Multiple   []string
	Panels     []string
}

type Msg struct {
	Target   string
	Expanded bool
	Keys     []string
}

func Update(model *Model, msg Msg) {
	switch msg.Target {
	case "standalone":
		model.Standalone = msg.Expanded
	case "single":
		model.Single = append([]string(nil), msg.Keys...)
	case "multiple":
		model.Multiple = append([]string(nil), msg.Keys...)
	case "panels":
		model.Panels = append([]string(nil), msg.Keys...)
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Scroll("collapsibles-scroll",
				ui.Column(
					ui.Text("FlowUI Collapsible").Size(24),
					section("Standalone",
						ui.Collapsible("standalone", model.Standalone, "What is FlowUI?",
							ui.Description("FlowUI is a desktop UI framework built on Gio with controlled MVU components."),
						).
							Leading(ui.Icon(lucide.Info).Size(16)).
							OnExpandedChange(func(expanded bool) {
								send(Msg{Target: "standalone", Expanded: expanded})
							}),
					),
					section("Single expansion",
						ui.CollapsibleGroup("single", model.Single, accountItems()).
							OnExpandedChange(func(keys []string) {
								send(Msg{Target: "single", Keys: keys})
							}),
					),
					section("Multiple expansion",
						ui.CollapsibleGroup("multiple", model.Multiple, notificationItems()).
							AllowMultipleExpanded(true).
							OnExpandedChange(func(keys []string) {
								send(Msg{Target: "multiple", Keys: keys})
							}),
					),
					section("Desktop panels",
						ui.Surface(
							ui.CollapsibleGroup("panels", model.Panels, desktopPanelItems()).
								AllowMultipleExpanded(true).
								OnExpandedChange(func(keys []string) {
									send(Msg{Target: "panels", Keys: keys})
								}),
						).Variant(ui.SurfaceSecondary).Style(ui.Radius(12)),
					),
				).Gap(22),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(640)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(8)
}

func accountItems() []ui.CollapsibleItem {
	return []ui.CollapsibleItem{
		{
			Key: "profile", Label: "Profile", Leading: ui.Icon(lucide.User).Size(16),
			Content: ui.Description("Update your display name, avatar, and public profile information."),
		},
		{
			Key: "security", Label: "Security", Leading: ui.Icon(lucide.ShieldCheck).Size(16),
			Content: ui.Description("Manage your password, passkeys, and active sessions."),
		},
		{
			Key: "billing", Label: "Billing", Leading: ui.Icon(lucide.CreditCard).Size(16),
			Content: ui.Description("View invoices and manage the subscription for this workspace."),
		},
	}
}

func notificationItems() []ui.CollapsibleItem {
	return []ui.CollapsibleItem{
		{
			Key: "email", Label: "Email notifications", Leading: ui.Icon(lucide.Mail).Size(16),
			Trailing: ui.Chip("On").Color(ui.ChipSuccess).Size(ui.ChipSmall),
			Content:  ui.Description("Receive product updates and weekly reports by email."),
		},
		{
			Key: "push", Label: "Desktop notifications", Leading: ui.Icon(lucide.Bell).Size(16),
			Content: ui.Description("Show notifications for mentions and assigned work."),
		},
		{
			Key: "digest", Label: "Daily digest", Leading: ui.Icon(lucide.Newspaper).Size(16), Disabled: true,
			Content: ui.Description("Daily digests are unavailable for this workspace."),
		},
	}
}

func desktopPanelItems() []ui.CollapsibleItem {
	return []ui.CollapsibleItem{
		{
			Key: "explorer", Label: "Explorer", Leading: ui.Icon(lucide.FolderTree).Size(16),
			Content: ui.Column(
				ui.Text("FLOWUI").Size(12),
				ui.Text("  internal").Size(13),
				ui.Text("  ui").Size(13),
				ui.Text("  examples").Size(13),
			).Gap(4),
		},
		{
			Key: "outline", Label: "Outline", Leading: ui.Icon(lucide.ListTree).Size(16),
			Content: ui.Column(
				ui.Text("Model").Size(13),
				ui.Text("Update").Size(13),
				ui.Text("View").Size(13),
			).Gap(4),
		},
		{
			Key: "timeline", Label: "Timeline", Leading: ui.Icon(lucide.History).Size(16),
			Content: ui.Description("No timeline information is available for the active file."),
		},
	}
}

func main() {
	ui.Run(
		Model{
			Standalone: true,
			Single:     []string{"security"},
			Multiple:   []string{"email", "push"},
			Panels:     []string{"explorer"},
		},
		Update,
		View,
		ui.Title("FlowUI Collapsible"),
		ui.Size(860, 820),
	)
}
