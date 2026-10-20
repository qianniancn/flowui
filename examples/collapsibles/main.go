package main

import (
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/collapsible"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
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
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Collapsible").Size(24),
				section("Standalone",
					collapsible.New("What is FlowUI?", description.New("FlowUI is a desktop UI framework built on Gio with controlled MVU components.")).Key("standalone").Expanded(model.Standalone).
						Leading(icon.New(lucide.Info).Size(16)).
						OnExpandedChange(func(expanded bool) {
							send(Msg{Target: "standalone", Expanded: expanded})
						}),
				),
				section("Single expansion",
					collapsible.Group("single", model.Single, accountItems()).
						OnExpandedChange(func(keys []string) {
							send(Msg{Target: "single", Keys: keys})
						}),
				),
				section("Multiple expansion",
					collapsible.Group("multiple", model.Multiple, notificationItems()).
						AllowMultipleExpanded(true).
						OnExpandedChange(func(keys []string) {
							send(Msg{Target: "multiple", Keys: keys})
						}),
				),
				section("Desktop panels",
					(surface.New(
						collapsible.Group("panels", model.Panels, desktopPanelItems()).
							AllowMultipleExpanded(true).
							OnExpandedChange(func(keys []string) {
								send(Msg{Target: "panels", Keys: keys})
							}),
					).Variant(surface.Secondary)).Style(style.New().Radius(12)),
				),
			).Gap(22),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(640).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(18), child).Gap(8)
}

func accountItems() []collapsible.Item {
	return []collapsible.Item{
		{
			Key: "profile", Label: "Profile", Leading: icon.New(lucide.User).Size(16),
			Content: description.New("Update your display name, avatar, and public profile information."),
		},
		{
			Key: "security", Label: "Security", Leading: icon.New(lucide.ShieldCheck).Size(16),
			Content: description.New("Manage your password, passkeys, and active sessions."),
		},
		{
			Key: "billing", Label: "Billing", Leading: icon.New(lucide.CreditCard).Size(16),
			Content: description.New("View invoices and manage the subscription for this workspace."),
		},
	}
}

func notificationItems() []collapsible.Item {
	return []collapsible.Item{
		{
			Key: "email", Label: "Email notifications", Leading: icon.New(lucide.Mail).Size(16),
			Trailing: chip.New("On").Color(chip.ColorSuccess).Size(chip.SizeSmall),
			Content:  description.New("Receive product updates and weekly reports by email."),
		},
		{
			Key: "push", Label: "Desktop notifications", Leading: icon.New(lucide.Bell).Size(16),
			Content: description.New("Show notifications for mentions and assigned work."),
		},
		{
			Key: "digest", Label: "Daily digest", Leading: icon.New(lucide.Newspaper).Size(16), Disabled: true,
			Content: description.New("Daily digests are unavailable for this workspace."),
		},
	}
}

func desktopPanelItems() []collapsible.Item {
	return []collapsible.Item{
		{
			Key: "explorer", Label: "Explorer", Leading: icon.New(lucide.FolderTree).Size(16),
			Content: layout.Column(
				text.New("FLOWUI").Size(12),
				text.New("  internal").Size(13),
				text.New("  ui").Size(13),
				text.New("  examples").Size(13),
			).Gap(4),
		},
		{
			Key: "outline", Label: "Outline", Leading: icon.New(lucide.ListTree).Size(16),
			Content: layout.Column(
				text.New("Model").Size(13),
				text.New("Update").Size(13),
				text.New("View").Size(13),
			).Gap(4),
		},
		{
			Key: "timeline", Label: "Timeline", Leading: icon.New(lucide.RotateCcwClock).Size(16),
			Content: description.New("No timeline information is available for the active file."),
		},
	}
}

func main() {
	ui.Run(ui.NewProgram(Model{
		Standalone: true,
		Single:     []string{"security"},
		Multiple:   []string{"email", "push"},
		Panels:     []string{"explorer"},
	},
		Update, View), window.Title("FlowUI Collapsible"),
		window.Size(860, 820),
		window.Theme(base.Light()),
	)
}
