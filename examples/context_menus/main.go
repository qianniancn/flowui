package main

import (
	"fmt"
	"strings"

	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type member struct {
	key    string
	name   string
	role   string
	status string
	email  string
}

type Model struct {
	ShowEmail bool
	Density   string
	Last      string
}

type Msg struct {
	Action    string
	ShowEmail *bool
	Density   string
}

func Update(model *Model, msg Msg) {
	if msg.Action != "" {
		model.Last = msg.Action
	}
	if msg.ShowEmail != nil {
		model.ShowEmail = *msg.ShowEmail
	}
	if msg.Density != "" {
		model.Density = msg.Density
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "No action selected"
	if model.Last != "" {
		status = model.Last
	}
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Team members").Size(24),
				ui.Text(status).Size(13),
				membersTable(model, send),
			).Gap(14),
		).FillWidth().MaxWidth(920).Padding(24),
	)
}

func membersTable(model Model, send ui.Send[Msg]) ui.TableWidget {
	columns := []ui.TableColumn{
		{Key: "name", Label: "Member", RowHeader: true, MinWidth: 210, Weight: 1.5},
		{Key: "role", Label: "Role", MinWidth: 180, Weight: 1.2},
		{Key: "status", Label: "Status", Width: 120},
	}
	if model.ShowEmail {
		columns = append(columns, ui.TableColumn{Key: "email", Label: "Email", MinWidth: 220, Weight: 1.4})
	}

	rows := make([]ui.TableRow, 0, len(members))
	for _, value := range members {
		cells := []ui.TableCell{
			{Content: memberContextMenu(value, model, send)},
			{Text: value.role},
			{Content: statusChip(value.status)},
		}
		if model.ShowEmail {
			cells = append(cells, ui.TableCell{Text: value.email})
		}
		rows = append(rows, ui.TableRow{Key: value.key, Label: value.name, Cells: cells})
	}

	minWidth := 620
	if model.ShowEmail {
		minWidth = 820
	}
	return ui.Table("context-menu-members", columns, rows).
		Variant(ui.TableSecondary).
		MinWidth(minWidth)
}

func memberContextMenu(value member, model Model, send ui.Send[Msg]) ui.ContextMenuWidget {
	menu := ui.Menu("member-actions", []ui.MenuItem{
		{Key: "open", Label: "Open profile", Shortcut: "Enter", Leading: ui.Icon(lucide.UserRound).Size(16)},
		{Key: "copy-email", Label: "Copy email", Shortcut: "Ctrl+C", Leading: ui.Icon(lucide.Copy).Size(16)},
		ui.MenuSeparator(),
		ui.MenuGroupLabel("View"),
		{Key: "show-email", Label: "Show email column", Kind: ui.MenuItemCheckbox, Checked: model.ShowEmail, KeepOpen: true},
		{Key: "comfortable", Label: "Comfortable density", Kind: ui.MenuItemRadio, RadioGroup: "density", Value: "comfortable", Checked: model.Density == "comfortable", KeepOpen: true},
		{Key: "compact", Label: "Compact density", Kind: ui.MenuItemRadio, RadioGroup: "density", Value: "compact", Checked: model.Density == "compact", KeepOpen: true},
		ui.MenuSeparator(),
		{
			Key: "move", Label: "Move to team", Kind: ui.MenuItemSubmenu, Leading: ui.Icon(lucide.FolderInput).Size(16),
			Children: []ui.MenuItem{
				{Key: "move-design", Label: "Design"},
				{Key: "move-engineering", Label: "Engineering"},
				{Key: "move-marketing", Label: "Marketing"},
			},
		},
		{Key: "archive", Label: "Archive member", Disabled: true, Leading: ui.Icon(lucide.Archive).Size(16)},
		{Key: "delete", Label: "Delete member", Variant: ui.MenuItemDanger, Leading: ui.Icon(lucide.Trash2).Size(16)},
	}).
		OnAction(func(action string) {
			send(Msg{Action: fmt.Sprintf("%s: %s", value.name, action)})
		}).
		OnCheckedChange(func(_ string, checked bool) {
			send(Msg{ShowEmail: &checked})
		}).
		OnRadioChange(func(_ string, density string) {
			send(Msg{Density: density})
		})

	return ui.ContextMenu("member-menu-"+value.key, memberIdentity(value), menu)
}

func memberIdentity(value member) ui.Widget {
	initials := ""
	for _, part := range strings.Fields(value.name) {
		initials += part[:1]
	}
	return ui.Row(
		ui.Box(ui.Text(initials).Size(12)).Width(30).Height(30).Align(ui.AlignCenter),
		ui.Column(
			ui.Text(value.name).Size(13),
			ui.Text(strings.ToUpper(value.key)).Size(11),
		).Gap(1),
	).Gap(10).AlignMiddle()
}

func statusChip(status string) ui.Widget {
	color := ui.ChipSuccess
	if status == "On leave" {
		color = ui.ChipWarning
	}
	return ui.Chip(status).Color(color).Variant(ui.ChipSoft).Size(ui.ChipSmall)
}

var members = []member{
	{key: "olivia", name: "Olivia Martin", role: "Product designer", status: "Active", email: "olivia@example.com"},
	{key: "jackson", name: "Jackson Lee", role: "Engineering manager", status: "Active", email: "jackson@example.com"},
	{key: "sophia", name: "Sophia Brown", role: "Marketing lead", status: "On leave", email: "sophia@example.com"},
	{key: "liam", name: "Liam Wilson", role: "Software engineer", status: "Active", email: "liam@example.com"},
}

func main() {
	ui.Run(
		Model{ShowEmail: true, Density: "comfortable"},
		Update,
		View,
		ui.Title("FlowUI Context Menu"),
		ui.Size(1000, 620),
	)
}
