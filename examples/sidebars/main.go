package main

import (
	"image/color"

	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	selected  string
	collapsed bool
}

type Msg any

type Navigate string
type ToggleSidebar struct{}

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case Navigate:
		model.selected = string(msg)
	case ToggleSidebar:
		model.collapsed = !model.collapsed
	}
}

func View(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	muted := ctx.Theme().Palette.MutedForeground
	sidebar := ui.SidebarSections("workspace-navigation", model.selected, navigationSections()).
		Header(sidebarHeader(model.collapsed, muted, send)).
		Footer(sidebarFooter(model.collapsed, muted)).
		Collapsed(model.collapsed).
		Alt("Workspace navigation").
		DisabledKeys([]string{"activity"}).
		OnChange(func(key string) { send(Navigate(key)) })

	return ui.Surface(
		ui.Row(
			sidebar,
			ui.Expanded(workspace(model.selected, muted)),
		),
	).Variant(ui.SurfaceSecondary)
}

func navigationSections() []ui.SidebarSection {
	return []ui.SidebarSection{
		{
			Title: "Workspace",
			Items: []ui.SidebarItem{
				{Key: "overview", Label: "Overview", Leading: ui.Icon(lucide.LayoutDashboard).Size(18)},
				{Key: "projects", Label: "Projects", Leading: ui.Icon(lucide.FolderKanban).Size(18), Trailing: ui.Chip("8").Variant(ui.ChipSoft).Size(ui.ChipSmall)},
				{Key: "calendar", Label: "Calendar", Leading: ui.Icon(lucide.CalendarDays).Size(18)},
			},
		},
		{
			Title: "Insights",
			Items: []ui.SidebarItem{
				{Key: "reports", Label: "Reports", Leading: ui.Icon(lucide.ChartNoAxesCombined).Size(18)},
				{Key: "activity", Label: "Activity", Leading: ui.Icon(lucide.ClipboardList).Size(18)},
			},
		},
		{
			Title: "Account",
			Items: []ui.SidebarItem{
				{Key: "team", Label: "Team", Leading: ui.Icon(lucide.Users).Size(18)},
				{Key: "settings", Label: "Settings", Leading: ui.Icon(lucide.Settings).Size(18)},
			},
		},
	}
}

func sidebarHeader(collapsed bool, muted color.NRGBA, send ui.Send[Msg]) ui.Widget {
	icon := lucide.PanelLeftClose
	label := "Collapse sidebar"
	if collapsed {
		icon = lucide.PanelLeftOpen
		label = "Expand sidebar"
	}
	toggle := ui.Tooltip(
		"sidebar-toggle-tip",
		ui.Button("sidebar-toggle", ui.Icon(icon).Size(18)).
			Variant(ui.ButtonGhost).
			Size(ui.ButtonSmall).
			IconOnly().
			OnClick(func() { send(ToggleSidebar{}) }),
		ui.Text(label),
	).Placement(ui.TooltipRight)
	if collapsed {
		return ui.Center(toggle)
	}
	return ui.Row(
		ui.Icon(lucide.LayoutDashboard).Size(22),
		ui.Expanded(ui.Column(
			ui.Text("Northstar").Size(15),
			ui.Text("Product workspace").Size(12).Color(muted),
		)),
		toggle,
	).AlignMiddle().Gap(10)
}

func sidebarFooter(collapsed bool, muted color.NRGBA) ui.Widget {
	avatar := ui.Avatar("QN").Variant(ui.AvatarSoft).Size(ui.AvatarSmall)
	if collapsed {
		return ui.Center(avatar)
	}
	return ui.Row(
		avatar,
		ui.Expanded(ui.Column(
			ui.Text("Qian Nian").Size(14),
			ui.Text("Administrator").Size(12).Color(muted),
		)),
	).AlignMiddle().Gap(10)
}

func workspace(selected string, muted color.NRGBA) ui.Widget {
	title := map[string]string{
		"overview": "Overview",
		"projects": "Projects",
		"calendar": "Calendar",
		"reports":  "Reports",
		"team":     "Team",
		"settings": "Settings",
	}[selected]
	if title == "" {
		title = "Overview"
	}
	return ui.Surface(
		ui.Box(
			ui.Column(
				ui.Row(
					ui.Expanded(ui.Column(
						ui.Text(title).Size(26),
						ui.Text("Monday, July 14").Size(13).Color(muted),
					)),
					ui.Button("new-project", ui.Text("New project")),
				).AlignMiddle(),
				ui.Grid(3,
					metricCard("Active projects", "12", "Across four teams", muted),
					metricCard("Tasks due", "27", "Five due today", muted),
					metricCard("Team members", "36", "Three currently away", muted),
				).Gap(12),
				ui.Card(
					ui.Column(
						ui.Text("Upcoming milestones").Size(16),
						ui.Text("Delivery dates across active projects").Size(13).Color(muted),
					).Gap(4),
					ui.Column(
						milestone("Desktop navigation", "Today", muted),
						ui.Divider(),
						milestone("Workspace permissions", "Jul 18", muted),
						ui.Divider(),
						milestone("Quarterly review", "Jul 25", muted),
					).Gap(12),
				).Variant(ui.CardSecondary),
			).Gap(20),
		).Style(ui.FillWidth()).Style(ui.FillHeight()).Style(ui.Padding(28)),
	).Variant(ui.SurfaceSecondary)
}

func metricCard(label, value, detail string, muted color.NRGBA) ui.Widget {
	return ui.Card(
		ui.Column(
			ui.Text(label).Size(13).Color(muted),
			ui.Text(value).Size(28),
			ui.Text(detail).Size(12).Color(muted),
		).Gap(6),
	).Variant(ui.CardSecondary)
}

func milestone(name, date string, muted color.NRGBA) ui.Widget {
	return ui.Row(
		ui.Expanded(ui.Text(name).Size(14)),
		ui.Text(date).Size(13).Color(muted),
	).AlignMiddle()
}

func main() {
	ui.Run(
		Model{selected: "overview"},
		Update,
		View,
		ui.Title("FlowUI Sidebar"),
		ui.Size(1080, 700),
	)
}
