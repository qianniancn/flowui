package main

import (
	"fmt"
	"sort"
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
	Selected []string
	Sort     ui.TableSortDescriptor
	Last     string
}

type Msg struct {
	Selected []string
	Sort     *ui.TableSortDescriptor
	Action   string
}

func Update(model *Model, msg Msg) {
	if msg.Selected != nil {
		model.Selected = append([]string(nil), msg.Selected...)
	}
	if msg.Sort != nil {
		model.Sort = *msg.Sort
	}
	if msg.Action != "" {
		model.Last = msg.Action
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "No row activated"
	if model.Last != "" {
		status = "Activated: " + model.Last
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("tables",
				ui.Column(
					ui.Text("FlowUI Table").Size(24),
					ui.Text(status).Size(14),
					ui.Divider(),
					section("Primary", basicTable("primary", ui.TablePrimary)),
					section("Secondary", basicTable("secondary", ui.TableSecondary)),
					section("Selection, sorting, and custom cells", membersTable(model, send)),
					section("Scrollable", scrollableTable()),
					section("Empty",
						ui.Table("empty-members", memberColumns(), nil).
							MinWidth(680).
							EmptyContent(
								ui.Column(
									ui.Icon(lucide.Inbox).Size(24),
									ui.Text("No results found").Size(14),
								).Gap(8).AlignMiddle(),
							),
					),
				).Gap(20),
			).Vertical(),
		).FillWidth().MaxWidth(980).Padding(24),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func basicTable(key string, variant ui.TableVariant) ui.TableWidget {
	columns := []ui.TableColumn{
		{Key: "name", Label: "Name", RowHeader: true, Weight: 1.4},
		{Key: "role", Label: "Role"},
		{Key: "status", Label: "Status", Width: 120},
		{Key: "email", Label: "Email", Weight: 1.6},
	}
	rows := make([]ui.TableRow, 0, 4)
	for _, value := range members[:4] {
		rows = append(rows, ui.TableRow{
			Key: value.key, Label: value.name,
			Cells: []ui.TableCell{
				{Text: value.name},
				{Text: value.role},
				{Text: value.status},
				{Text: value.email},
			},
		})
	}
	return ui.Table(key, columns, rows).Variant(variant).MinWidth(680)
}

func membersTable(model Model, send ui.Send[Msg]) ui.TableWidget {
	sorted := append([]member(nil), members...)
	sort.SliceStable(sorted, func(left, right int) bool {
		first := memberValue(sorted[left], model.Sort.Column)
		second := memberValue(sorted[right], model.Sort.Column)
		less := strings.Compare(first, second) < 0
		if model.Sort.Direction == ui.TableSortDescending {
			return !less && first != second
		}
		return less
	})
	rows := make([]ui.TableRow, 0, len(sorted))
	for _, value := range sorted {
		rows = append(rows, customMemberRow(value))
	}
	return ui.Table("members", memberColumns(), rows).
		MinWidth(760).
		MaxHeight(300).
		SelectionMode(ui.TableSelectionMultiple).
		SelectedKeys(model.Selected).
		ShowSelectionIndicator().
		DisabledKeys([]string{"michael"}).
		SortDescriptor(model.Sort).
		OnSelectionChange(func(keys []string) { send(Msg{Selected: keys}) }).
		OnSortChange(func(descriptor ui.TableSortDescriptor) { send(Msg{Sort: &descriptor}) }).
		OnAction(func(key string) { send(Msg{Action: key}) }).
		Footer(ui.Text(fmt.Sprintf("%d members, %d selected", len(rows), len(model.Selected))).Size(12))
}

func memberColumns() []ui.TableColumn {
	return []ui.TableColumn{
		{Key: "name", Label: "Member", Sortable: true, RowHeader: true, MinWidth: 180, Weight: 1.8},
		{Key: "role", Label: "Role", Sortable: true, MinWidth: 150, Weight: 1.2},
		{Key: "status", Label: "Status", Sortable: true, Width: 124},
		{Key: "email", Label: "Email", Sortable: true, MinWidth: 210, Weight: 1.5},
	}
}

func customMemberRow(value member) ui.TableRow {
	return ui.TableRow{
		Key: value.key, Label: value.name,
		Cells: []ui.TableCell{
			{Content: memberIdentity(value)},
			{Text: value.role},
			{Content: statusChip(value.status)},
			{Text: value.email},
		},
	}
}

func memberIdentity(value member) ui.Widget {
	initials := ""
	for _, part := range strings.Fields(value.name) {
		initials += part[:1]
	}
	return ui.Row(
		ui.Box(ui.Text(initials).Size(12)).Width(28).Height(28).Align(ui.AlignCenter),
		ui.Column(
			ui.Text(value.name).Size(13),
			ui.Text("ID "+strings.ToUpper(value.key)).Size(11),
		).Gap(1),
	).Gap(10).AlignMiddle()
}

func statusChip(status string) ui.Widget {
	color := ui.ChipSuccess
	switch status {
	case "Inactive":
		color = ui.ChipDanger
	case "On Leave":
		color = ui.ChipWarning
	}
	return ui.Chip(status).Color(color).Variant(ui.ChipSoft).Size(ui.ChipSmall)
}

func scrollableTable() ui.TableWidget {
	rows := make([]ui.TableRow, 16)
	for index := range rows {
		rows[index] = ui.TableRow{
			Key: fmt.Sprintf("job-%02d", index+1), Label: fmt.Sprintf("Job %02d", index+1),
			Cells: []ui.TableCell{
				{Text: fmt.Sprintf("Job %02d", index+1)},
				{Text: []string{"Designer", "Engineer", "Manager"}[index%3]},
				{Content: statusChip([]string{"Active", "On Leave", "Inactive"}[index%3])},
				{Text: fmt.Sprintf("job%02d@flowui.dev", index+1)},
			},
		}
	}
	return ui.Table("scrollable-members", memberColumns(), rows).
		Variant(ui.TableSecondary).
		MinWidth(720).
		MaxHeight(260)
}

func memberValue(value member, column string) string {
	switch column {
	case "role":
		return value.role
	case "status":
		return value.status
	case "email":
		return value.email
	default:
		return value.name
	}
}

var members = []member{
	{key: "kate", name: "Kate Moore", role: "Chief Executive Officer", status: "Active", email: "kate@acme.com"},
	{key: "john", name: "John Smith", role: "Chief Technology Officer", status: "Active", email: "john@acme.com"},
	{key: "sara", name: "Sara Johnson", role: "Chief Marketing Officer", status: "On Leave", email: "sara@acme.com"},
	{key: "michael", name: "Michael Brown", role: "Chief Financial Officer", status: "Active", email: "michael@acme.com"},
	{key: "emily", name: "Emily Davis", role: "Product Manager", status: "Inactive", email: "emily@acme.com"},
	{key: "davis", name: "Davis Wilson", role: "Lead Designer", status: "Active", email: "davis@acme.com"},
}

func main() {
	ui.Run(
		Model{
			Selected: []string{"kate", "sara"},
			Sort:     ui.TableSortDescriptor{Column: "name", Direction: ui.TableSortAscending},
		},
		Update,
		View,
		ui.Title("FlowUI Table"),
		ui.Size(1100, 820),
	)
}
