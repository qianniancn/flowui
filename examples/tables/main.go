package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

type member struct {
	key    string
	name   string
	role   string
	status string
	email  string
}

type Model struct {
	Selected     []string
	Sort         ui.TableSortDescriptor
	Last         string
	Page         int
	Loaded       int
	Loading      bool
	EditableName string
	EditableRole string
}

type Msg struct {
	Selected  []string
	Sort      *ui.TableSortDescriptor
	Action    string
	Page      int
	LoadMore  bool
	Loaded    int
	EditCell  string
	EditValue string
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	if msg.Selected != nil {
		model.Selected = append([]string(nil), msg.Selected...)
	}
	if msg.Sort != nil {
		model.Sort = *msg.Sort
	}
	if msg.Action != "" {
		model.Last = msg.Action
	}
	if msg.Page > 0 {
		model.Page = msg.Page
	}
	if msg.LoadMore && !model.Loading && model.Loaded < asyncRowCount {
		model.Loading = true
		nextLoaded := min(model.Loaded+asyncBatchSize, asyncRowCount)
		return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
			timer := time.NewTimer(700 * time.Millisecond)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				send(Msg{Loaded: nextLoaded})
				return nil
			}
		})
	}
	if msg.Loaded > 0 {
		model.Loaded = msg.Loaded
		model.Loading = false
	}
	switch msg.EditCell {
	case "name":
		model.EditableName = msg.EditValue
	case "role":
		model.EditableRole = msg.EditValue
	}
	return nil
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
					section("Excel-style grid and border", basicTable("grid", ui.TableSecondary).GridLines(true).Border(true)),
					section("Interactive Input and Select cells", editableTable(model, send)),
					section("Selection, sorting, resizing, and custom cells", membersTable(model, send)),
					section("Pagination", paginatedTable(model, send)),
					section("Async loading", asyncTable(model, send)),
					section("Virtualized 10,000 rows", virtualTable()),
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
		).Style(ui.FillWidth().MaxWidth(980).Padding(24)),
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
		OnColumnResize(func(key string, width int) { send(Msg{Action: fmt.Sprintf("Resized %s to %d dp", key, width)}) }).
		Footer(ui.Text(fmt.Sprintf("%d members, %d selected", len(rows), len(model.Selected))).Size(12))
}

func memberColumns() []ui.TableColumn {
	return []ui.TableColumn{
		{Key: "name", Label: "Member", Sortable: true, Resizable: true, RowHeader: true, MinWidth: 180, MaxWidth: 320, Weight: 1.8},
		{Key: "role", Label: "Role", Sortable: true, Resizable: true, MinWidth: 150, MaxWidth: 320, Weight: 1.2},
		{Key: "status", Label: "Status", Sortable: true, Resizable: true, Width: 124, MinWidth: 100, MaxWidth: 180},
		{Key: "email", Label: "Email", Sortable: true, MinWidth: 210, Weight: 1.5},
	}
}

func editableTable(model Model, send ui.Send[Msg]) ui.TableWidget {
	roles := []ui.SelectItem{
		{Key: "designer", Label: "Designer"},
		{Key: "engineer", Label: "Engineer"},
		{Key: "manager", Label: "Manager"},
	}
	rows := []ui.TableRow{{
		Key: "editable-member", Label: model.EditableName,
		Cells: []ui.TableCell{
			{
				Content: ui.Input("editable-member-name", model.EditableName).
					OnChange(func(value string) { send(Msg{EditCell: "name", EditValue: value}) }).
					FullWidth(),
				Interactive: true,
			},
			{
				Content: ui.Select("editable-member-role", model.EditableRole, roles).
					OnChange(func(value string) { send(Msg{EditCell: "role", EditValue: value}) }).
					FullWidth(),
				Interactive: true,
			},
			{Content: statusChip("Active")},
		},
	}}
	columns := []ui.TableColumn{
		{Key: "name", Label: "Name", MinWidth: 220, Weight: 1.4},
		{Key: "role", Label: "Role", MinWidth: 180, Weight: 1},
		{Key: "status", Label: "Status", Width: 120},
	}
	return ui.Table("editable-members", columns, rows).
		Variant(ui.TableSecondary).
		GridLines(true).
		Border(true).
		MinWidth(620).
		RowHeight(64)
}

func paginatedTable(model Model, send ui.Send[Msg]) ui.TableWidget {
	const rowsPerPage = 3
	totalPages := (len(members) + rowsPerPage - 1) / rowsPerPage
	page := min(max(model.Page, 1), totalPages)
	start := (page - 1) * rowsPerPage
	end := min(start+rowsPerPage, len(members))
	rows := make([]ui.TableRow, 0, end-start)
	for _, value := range members[start:end] {
		rows = append(rows, customMemberRow(value))
	}
	pagination := ui.Pagination("members-pages", page, totalPages).
		Size(ui.PaginationSmall).
		Summary(ui.Text(fmt.Sprintf("%d to %d of %d results", start+1, end, len(members))).Size(12)).
		OnChange(func(page int) { send(Msg{Page: page}) })
	return ui.Table("paginated-members", memberColumns(), rows).
		MinWidth(760).
		Footer(pagination)
}

const (
	asyncRowCount  = 24
	asyncBatchSize = 6
)

func asyncTable(model Model, send ui.Send[Msg]) ui.TableWidget {
	rows := make([]ui.TableRow, model.Loaded)
	for index := range rows {
		rows[index] = generatedRow(index)
	}
	return ui.Table("async-members", memberColumns(), rows).
		Variant(ui.TableSecondary).
		MinWidth(760).
		MaxHeight(260).
		LoadMore(model.Loaded < asyncRowCount, model.Loading, func() { send(Msg{LoadMore: true}) })
}

func virtualTable() ui.TableWidget {
	return ui.VirtualTable("virtual-members", memberColumns(), 10_000, generatedRow).
		Variant(ui.TableSecondary).
		MinWidth(760).
		MaxHeight(300).
		RowHeight(44)
}

func generatedRow(index int) ui.TableRow {
	firstNames := [...]string{"Emma", "Liam", "Olivia", "Noah", "Ava", "James", "Sophia", "Oliver"}
	lastNames := [...]string{"Smith", "Johnson", "Brown", "Davis", "Wilson", "Taylor", "Thomas", "Martin"}
	roles := [...]string{"Software Engineer", "Product Manager", "Designer", "Data Analyst"}
	statuses := [...]string{"Active", "On Leave", "Inactive"}
	first := firstNames[index%len(firstNames)]
	last := lastNames[index/len(firstNames)%len(lastNames)]
	status := statuses[index%len(statuses)]
	return ui.TableRow{
		Key: fmt.Sprintf("generated-%d", index), Label: first + " " + last,
		Cells: []ui.TableCell{
			{Text: first + " " + last},
			{Text: roles[index%len(roles)]},
			{Content: statusChip(status)},
			{Text: fmt.Sprintf("%s.%s%04d@flowui.dev", strings.ToLower(first), strings.ToLower(last), index+1)},
		},
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
	var initials strings.Builder
	for part := range strings.FieldsSeq(value.name) {
		initials.WriteString(part[:1])
	}
	return ui.Row(
		ui.Avatar(initials.String()).Size(ui.AvatarSmall).Variant(ui.AvatarSoft),
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
	ui.RunCmd(
		Model{
			Selected:     []string{"kate", "sara"},
			Sort:         ui.TableSortDescriptor{Column: "name", Direction: ui.TableSortAscending},
			Page:         1,
			Loaded:       asyncBatchSize,
			EditableName: "Ada Lovelace",
			EditableRole: "engineer",
		},
		Update,
		View,
		ui.Title("FlowUI Table"),
		ui.Size(1100, 820),
	)
}
