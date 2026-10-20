package main

import (
	"image/color"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	avatarpkg "github.com/qianniancn/flowui/widget/avatar"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/chip"
	iconpkg "github.com/qianniancn/flowui/widget/icon"
	sidebarpkg "github.com/qianniancn/flowui/widget/sidebar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/tooltip"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	selected  string
	collapsed bool
	openKeys  []string
}

type Msg any

type Navigate string
type ToggleSidebar struct{}
type SetSidebarOpenKeys []string

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Navigate:
		model.selected = string(msg)
	case ToggleSidebar:
		model.collapsed = !model.collapsed
	case SetSidebarOpenKeys:
		model.openKeys = append(model.openKeys[:0], msg...)
	}
	return nil
}

func View(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	muted := ctx.Theme().Palette.MutedForeground
	sidebar := sidebarpkg.NewSections("workspace-navigation", model.selected, navigationSections()).
		Header(sidebarHeader(model.collapsed, muted)).
		Footer(sidebarFooter(model.collapsed, muted)).
		Collapsed(model.collapsed).
		OpenKeys(model.openKeys).
		Alt("Workspace navigation").
		DisabledKeys([]string{"activity"}).
		OnChange(func(key string) { send(Navigate(key)) }).
		OnOpenChange(func(keys []string) { send(SetSidebarOpenKeys(keys)) })

	content := layout.Stack(
		layout.Stacked(workspace(model.selected, muted)),
		layout.Overlay(
			layout.Stack(
				layout.Overlay(sidebarBoundaryToggle(model.collapsed, send)).Align(layout.AlignTopStart),
			),
		).Expanded().Offset(-16, 10),
	)
	return surface.New(layout.Row(sidebar, layout.Expanded(content))).Variant(surface.Secondary)
}

func navigationSections() []sidebarpkg.Section {
	return []sidebarpkg.Section{
		{
			Title: "Workspace",
			Items: []sidebarpkg.Item{
				{Key: "overview", Label: "Overview", Leading: iconpkg.New(lucide.LayoutDashboard).Size(18)},
				{
					Key:      "projects",
					Label:    "Projects",
					Leading:  iconpkg.New(lucide.FolderKanban).Size(18),
					Trailing: chip.New("8").Variant(chip.VariantSoft).Size(chip.SizeSmall),
					Children: []sidebarpkg.Item{
						{Key: "roadmap", Label: "Roadmap", Leading: iconpkg.New(lucide.Map).Size(16)},
						{Key: "releases", Label: "Releases", Leading: iconpkg.New(lucide.Rocket).Size(16)},
					},
				},
				{Key: "calendar", Label: "Calendar", Leading: iconpkg.New(lucide.CalendarDays).Size(18)},
			},
		},
		{
			Title: "Insights",
			Items: []sidebarpkg.Item{
				{Key: "reports", Label: "Reports", Leading: iconpkg.New(lucide.ChartNoAxesCombined).Size(18)},
				{Key: "activity", Label: "Activity", Leading: iconpkg.New(lucide.ClipboardList).Size(18)},
			},
		},
		{
			Title: "Account",
			Items: []sidebarpkg.Item{
				{Key: "team", Label: "Team", Leading: iconpkg.New(lucide.Users).Size(18)},
				{Key: "settings", Label: "Settings", Leading: iconpkg.New(lucide.Settings).Size(18)},
			},
		},
	}
}

func sidebarHeader(collapsed bool, muted color.NRGBA) ui.Widget {
	if collapsed {
		return layout.Center(iconpkg.New(lucide.LayoutDashboard).Size(22))
	}
	return layout.Row(
		iconpkg.New(lucide.LayoutDashboard).Size(22),
		layout.Expanded(layout.Column(
			text.New("Northstar").Size(15),
			text.New("Product workspace").Size(12).Color(muted),
		)),
	).AlignMiddle().Gap(10)
}

func sidebarBoundaryToggle(collapsed bool, send ui.Send[Msg]) ui.Widget {
	icon := lucide.CircleChevronLeft
	label := "Collapse sidebar"
	if collapsed {
		icon = lucide.CircleChevronRight
		label = "Expand sidebar"
	}
	return tooltip.New("sidebar-toggle-tip", button.New("").Content(iconpkg.New(icon).Size(18)).
		Variant(button.Ghost).
		Size(button.Small).
		IconOnly().
		OnClick(func() { send(ToggleSidebar{}) }), label).Placement(tooltip.Right)
}

func sidebarFooter(collapsed bool, muted color.NRGBA) ui.Widget {
	avatar := avatarpkg.New("QN").Variant(avatarpkg.VariantSoft).Size(avatarpkg.SizeSmall)
	if collapsed {
		return layout.Center(avatar)
	}
	return layout.Row(
		avatar,
		layout.Expanded(layout.Column(
			text.New("Qian Nian").Size(14),
			text.New("Administrator").Size(12).Color(muted),
		)),
	).AlignMiddle().Gap(10)
}

func workspace(selected string, muted color.NRGBA) ui.Widget {
	title := map[string]string{
		"overview": "Overview",
		"projects": "Projects",
		"roadmap":  "Roadmap",
		"releases": "Releases",
		"calendar": "Calendar",
		"reports":  "Reports",
		"team":     "Team",
		"settings": "Settings",
	}[selected]
	if title == "" {
		title = "Overview"
	}
	return surface.New(
		layout.Box(
			layout.Column(
				layout.Row(
					layout.Expanded(layout.Column(
						text.New(title).Size(26),
						text.New("Monday, July 14").Size(13).Color(muted),
					)),
					button.New("New project"),
				).AlignMiddle(),
				layout.Grid(3,
					metricCard("Active projects", "12", "Across four teams", muted),
					metricCard("Tasks due", "27", "Five due today", muted),
					metricCard("Team members", "36", "Three currently away", muted),
				).Gap(12),
				card.New(
					layout.Column(
						text.New("Upcoming milestones").Size(16),
						text.New("Delivery dates across active projects").Size(13).Color(muted),
					).Gap(4),
					layout.Column(
						milestone("Desktop navigation", "Today", muted),
						layout.Divider(),
						milestone("Workspace permissions", "Jul 18", muted),
						layout.Divider(),
						milestone("Quarterly review", "Jul 25", muted),
					).Gap(12),
				).Variant(card.Secondary),
			).Gap(20),
		).Style(style.New().FillWidth().FillHeight().Padding(28)),
	).Variant(surface.Secondary)
}

func metricCard(label, value, detail string, muted color.NRGBA) ui.Widget {
	return card.New(
		layout.Column(
			text.New(label).Size(13).Color(muted),
			text.New(value).Size(28),
			text.New(detail).Size(12).Color(muted),
		).Gap(6),
	).Variant(card.Secondary)
}

func milestone(name, date string, muted color.NRGBA) ui.Widget {
	return layout.Row(
		layout.Expanded(text.New(name).Size(14)),
		text.New(date).Size(13).Color(muted),
	).AlignMiddle()
}

func main() {
	ui.Run(ui.NewProgram(Model{selected: "overview", openKeys: []string{"projects"}},
		Update, View), window.Title("FlowUI Sidebar"),
		window.Size(1080, 700),
		window.Theme(base.Light()),
	)
}
