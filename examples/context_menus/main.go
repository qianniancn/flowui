package main

import (
	"fmt"
	"strings"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/table"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	if msg.Action != "" {
		model.Last = msg.Action
	}
	if msg.ShowEmail != nil {
		model.ShowEmail = *msg.ShowEmail
	}
	if msg.Density != "" {
		model.Density = msg.Density
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "No action selected"
	if model.Last != "" {
		status = model.Last
	}
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Team members").Size(24),
				text.New(status).Size(13),
				membersTable(model, send),
			).Gap(14),
		).Style(style.New().FillWidth().MaxWidth(920).Padding(24)),
	)
}

func membersTable(model Model, send ui.Send[Msg]) table.Widget {
	columns := []table.Column{
		{Key: "name", Label: "Member", RowHeader: true, MinWidth: 210, Weight: 1.5},
		{Key: "role", Label: "Role", MinWidth: 180, Weight: 1.2},
		{Key: "status", Label: "Status", Width: 120},
	}
	if model.ShowEmail {
		columns = append(columns, table.Column{Key: "email", Label: "Email", MinWidth: 220, Weight: 1.4})
	}

	rows := make([]table.Row, 0, len(members))
	rowsByKey := make(map[string]member, len(members))
	for _, value := range members {
		rowsByKey[value.key] = value
		cells := []table.Cell{
			{Content: memberIdentity(value)},
			{Text: value.role},
			{Content: statusChip(value.status)},
		}
		if model.ShowEmail {
			cells = append(cells, table.Cell{Text: value.email})
		}
		rows = append(rows, table.Row{Key: value.key, Label: value.name, Cells: cells})
	}

	minWidth := 620
	if model.ShowEmail {
		minWidth = 820
	}
	return table.New("context-menu-members", columns, rows).
		Variant(table.VariantSecondary).
		MinWidth(minWidth).
		RowContextMenu(func(row table.Row) menu.Widget {
			return memberMenu(rowsByKey[row.Key], model, send)
		})
}

func memberMenu(value member, model Model, send ui.Send[Msg]) menu.Widget {
	return menu.New("member-actions", []menu.Item{
		{Key: "open", Label: "Open profile", Shortcut: "Enter", Leading: icon.New(lucide.UserRound).Size(16)},
		{Key: "copy-email", Label: "Copy email", Shortcut: "Ctrl+C", Leading: icon.New(lucide.Copy).Size(16)},
		menu.Separator(),
		menu.GroupLabel("View"),
		{Key: "show-email", Label: "Show email column", Kind: menu.ItemCheckbox, Checked: model.ShowEmail, KeepOpen: true},
		{Key: "comfortable", Label: "Comfortable density", Kind: menu.ItemRadio, RadioGroup: "density", Value: "comfortable", Checked: model.Density == "comfortable", KeepOpen: true},
		{Key: "compact", Label: "Compact density", Kind: menu.ItemRadio, RadioGroup: "density", Value: "compact", Checked: model.Density == "compact", KeepOpen: true},
		menu.Separator(),
		{
			Key: "move", Label: "Move to team", Kind: menu.ItemSubmenu, Leading: icon.New(lucide.FolderInput).Size(16),
			Children: []menu.Item{
				{Key: "move-design", Label: "Design"},
				{Key: "move-engineering", Label: "Engineering"},
				{Key: "move-marketing", Label: "Marketing"},
			},
		},
		{Key: "archive", Label: "Archive member", Disabled: true, Leading: icon.New(lucide.Archive).Size(16)},
		{Key: "delete", Label: "Delete member", Variant: menu.ItemDanger, Leading: icon.New(lucide.Trash).Size(16)},
	}).
		OnActionEvent(func(event menu.ActionEvent) {
			send(Msg{Action: fmt.Sprintf("%s: %s", value.name, event.Key)})
		}).
		OnCheckedChange(func(_ string, checked bool) {
			send(Msg{ShowEmail: &checked})
		}).
		OnRadioChange(func(_ string, density string) {
			send(Msg{Density: density})
		})
}

func memberIdentity(value member) ui.Widget {
	var initials strings.Builder
	for part := range strings.FieldsSeq(value.name) {
		initials.WriteString(part[:1])
	}
	return layout.Row(
		layout.Box(text.New(initials.String()).Size(12)).Style(style.New().Width(30).Height(30)).Align(layout.AlignCenter),
		layout.Column(
			text.New(value.name).Size(13),
			text.New(strings.ToUpper(value.key)).Size(11),
		).Gap(1),
	).Gap(10).AlignMiddle()
}

func statusChip(status string) ui.Widget {
	color := chip.ColorSuccess
	if status == "On leave" {
		color = chip.ColorWarning
	}
	return chip.New(status).Color(color).Variant(chip.VariantSoft).Size(chip.SizeSmall)
}

var members = []member{
	{key: "olivia", name: "Olivia Martin", role: "Product designer", status: "Active", email: "olivia@example.com"},
	{key: "jackson", name: "Jackson Lee", role: "Engineering manager", status: "Active", email: "jackson@example.com"},
	{key: "sophia", name: "Sophia Brown", role: "Marketing lead", status: "On leave", email: "sophia@example.com"},
	{key: "liam", name: "Liam Wilson", role: "Software engineer", status: "Active", email: "liam@example.com"},
}

func main() {
	ui.Run(ui.NewProgram(Model{ShowEmail: true, Density: "comfortable"},
		Update, View), window.Title("FlowUI Context Menu"),
		window.Size(1000, 620),
		window.Theme(base.Light()),
	)
}
