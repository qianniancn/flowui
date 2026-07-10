package main

import (
	"fmt"
	"strings"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Member string
	Tool   string
	City   string
	Roles  []string
	Last   string
}

type Field string

const (
	fieldMember Field = "member"
	fieldTool   Field = "tool"
	fieldCity   Field = "city"
	fieldRoles  Field = "roles"
)

type Msg struct {
	Field Field
	Key   string
	Keys  []string
}

func Update(m *Model, msg Msg) {
	if msg.Keys != nil {
		switch msg.Field {
		case fieldRoles:
			m.Roles = append([]string(nil), msg.Keys...)
		}
		if len(msg.Keys) == 0 {
			m.Last = fmt.Sprintf("%s cleared", msg.Field)
			return
		}
		m.Last = fmt.Sprintf("%s selected %s", msg.Field, strings.Join(msg.Keys, ", "))
		return
	}

	switch msg.Field {
	case fieldMember:
		m.Member = msg.Key
	case fieldTool:
		m.Tool = msg.Key
	case fieldCity:
		m.City = msg.Key
	}
	if msg.Key == "" {
		m.Last = fmt.Sprintf("%s cleared", msg.Field)
		return
	}
	m.Last = fmt.Sprintf("%s selected %s", msg.Field, msg.Key)
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No selection"
	if m.Last != "" {
		status = m.Last
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("list-boxes",
				ui.Column(
					ui.Text("FlowUI ListBox").Size(24),
					ui.Text(status).Size(16),
					ui.Divider(),
					section("Members",
						ui.Box(listBox("members", fieldMember, m.Member, members, send).FullWidth()).
							Width(420),
					),
					section("Actions",
						ui.Box(actionList(send)).
							Width(420),
					),
					section("Sections",
						ui.Box(sectionedActionList(send)).
							Width(420),
					),
					section("States",
						ui.Column(
							ui.Box(listBox("tools", fieldTool, m.Tool, tools, send).
								DisabledKeys([]string{"eraser"}).
								FullWidth()).
								Width(420),
							ui.Box(ui.ListBox("disabled-tools", "brush", tools).
								DisabledKeys([]string{"eraser"}).
								FullWidth().
								Disabled(true)).
								Width(420),
						).Gap(12),
					),
					section("Multiple",
						ui.Box(multiListBox("roles", fieldRoles, m.Roles, roles, send).
							FullWidth()).
							Width(420),
					),
					section("Scrollable",
						ui.Box(listBox("cities", fieldCity, m.City, cities, send).
							FullWidth().
							MaxHeight(160)).
							Width(420),
					),
					section("Empty",
						ui.Box(ui.ListBox("empty", "", nil).
							EmptyText("No matching items").
							FullWidth()).
							Width(420),
					),
				).Gap(18),
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

func listBox(key string, field Field, selected string, items []ui.ListBoxItem, send ui.Send[Msg]) ui.ListBoxWidget {
	return ui.ListBox(key, selected, items).
		AllowEmptySelection().
		OnChange(func(selected string) {
			send(Msg{
				Field: field,
				Key:   selected,
			})
		})
}

func actionList(send ui.Send[Msg]) ui.ListBoxWidget {
	return ui.ListBox("actions", "", actions).
		SelectionMode(ui.ListBoxSelectionNone).
		HideIndicator().
		OnAction(func(action string) {
			send(Msg{
				Field: "action",
				Key:   action,
			})
		})
}

func sectionedActionList(send ui.Send[Msg]) ui.ListBoxWidget {
	return ui.ListBoxSections("sectioned-actions", "", actionSections).
		SelectionMode(ui.ListBoxSelectionNone).
		HideIndicator().
		OnAction(func(action string) {
			send(Msg{
				Field: "section-action",
				Key:   action,
			})
		})
}

func multiListBox(key string, field Field, selected []string, items []ui.ListBoxItem, send ui.Send[Msg]) ui.ListBoxWidget {
	return ui.ListBoxMultiple(key, selected, items).
		OnSelectionChange(func(selected []string) {
			send(Msg{
				Field: field,
				Keys:  selected,
			})
		})
}

func avatar(letter string) ui.Widget {
	return ui.Box(ui.Text(letter).Size(14)).
		Width(28).
		Height(28).
		Align(ui.AlignCenter)
}

func statusIndicator(selected bool) ui.Widget {
	if !selected {
		return nil
	}
	return ui.Text("on").Size(11)
}

var members = []ui.ListBoxItem{
	{Key: "bob", Label: "Bob", Description: "bob@flowui.dev", Leading: avatar("B")},
	{Key: "fred", Label: "Fred", Description: "fred@flowui.dev", Leading: avatar("F")},
	{Key: "martha", Label: "Martha", Description: "martha@flowui.dev", Leading: avatar("M")},
}

var actions = []ui.ListBoxItem{
	{Key: "new-file", Label: "New file", Description: "Create a new file", Trailing: ui.Text("Ctrl+N").Size(12)},
	{Key: "edit-file", Label: "Edit file", Description: "Make changes", Trailing: ui.Text("Ctrl+E").Size(12)},
	{Key: "delete-file", Label: "Delete file", Description: "Move to trash", Trailing: ui.Text("Ctrl+Shift+D").Size(12), Variant: ui.ListBoxItemDanger},
}

var actionSections = []ui.ListBoxSection{
	{
		Title: "Actions",
		Items: []ui.ListBoxItem{
			{Key: "new-file", Label: "New file", Description: "Create a new file", Trailing: ui.Text("Ctrl+N").Size(12)},
			{Key: "edit-file", Label: "Edit file", Description: "Make changes", Trailing: ui.Text("Ctrl+E").Size(12)},
		},
	},
	{
		Title: "Danger zone",
		Items: []ui.ListBoxItem{
			{Key: "delete-file", Label: "Delete file", Description: "Move to trash", Trailing: ui.Text("Ctrl+Shift+D").Size(12), Variant: ui.ListBoxItemDanger},
		},
	},
}

var tools = []ui.ListBoxItem{
	{Key: "brush", Label: "Brush"},
	{Key: "eraser", Label: "Eraser"},
	{Key: "legacy", Label: "Legacy tool", Description: "Unavailable for this document", Disabled: true},
}

var roles = []ui.ListBoxItem{
	{Key: "read", Label: "Read", Description: "View project content", Indicator: statusIndicator},
	{Key: "comment", Label: "Comment", Description: "Discuss changes", Indicator: statusIndicator},
	{Key: "write", Label: "Write", Description: "Create and edit content", Indicator: statusIndicator},
	{Key: "admin", Label: "Admin", Description: "Manage access", Variant: ui.ListBoxItemDanger, Indicator: statusIndicator},
}

var cities = []ui.ListBoxItem{
	{Key: "beijing", Label: "Beijing", Description: "China north region"},
	{Key: "shanghai", Label: "Shanghai", Description: "China east region"},
	{Key: "shenzhen", Label: "Shenzhen", Description: "China south region"},
	{Key: "hangzhou", Label: "Hangzhou", Description: "Design and commerce"},
	{Key: "chengdu", Label: "Chengdu", Description: "Southwest operations"},
	{Key: "wuhan", Label: "Wuhan", Description: "Central region"},
	{Key: "xian", Label: "Xi'an", Description: "Northwest region"},
	{Key: "nanjing", Label: "Nanjing", Description: "Jiangsu hub"},
}

func main() {
	ui.Run(Model{
		Member: "fred",
		Tool:   "brush",
		City:   "shanghai",
		Roles:  []string{"read", "comment"},
	}, Update, View,
		ui.Title("FlowUI ListBox"),
		ui.Size(900, 720),
	)
}
