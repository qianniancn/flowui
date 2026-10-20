package main

import (
	"fmt"
	"strings"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/listbox"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	if msg.Keys != nil {
		switch msg.Field {
		case fieldRoles:
			m.Roles = append([]string(nil), msg.Keys...)
		}
		if len(msg.Keys) == 0 {
			m.Last = fmt.Sprintf("%s cleared", msg.Field)
			return nil
		}
		m.Last = fmt.Sprintf("%s selected %s", msg.Field, strings.Join(msg.Keys, ", "))
		return nil
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
		return nil
	}
	m.Last = fmt.Sprintf("%s selected %s", msg.Field, msg.Key)
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No selection"
	if m.Last != "" {
		status = m.Last
	}

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI ListBox").Size(24),
				text.New(status).Size(16),
				layout.Divider(),
				section("Members",
					layout.Box(listBox("members", fieldMember, m.Member, members, send).FullWidth()).
						Style(style.New().Width(420)),
				),
				section("Actions",
					layout.Box(actionList(send)).
						Style(style.New().Width(420)),
				),
				section("Sections",
					layout.Box(sectionedActionList(send)).
						Style(style.New().Width(420)),
				),
				section("States",
					layout.Column(
						layout.Box(listBox("tools", fieldTool, m.Tool, tools, send).
							DisabledKeys([]string{"eraser"}).
							FullWidth()).
							Style(style.New().Width(420)),
						layout.Box(listbox.New(tools).Key("disabled-tools").SelectedKey("brush").
							DisabledKeys([]string{"eraser"}).
							FullWidth().
							Disabled(true)).
							Style(style.New().Width(420)),
					).Gap(12),
				),
				section("Multiple",
					layout.Box(multiListBox("roles", fieldRoles, m.Roles, roles, send).
						FullWidth()).
						Style(style.New().Width(420)),
				),
				section("Scrollable",
					layout.Box(listBox("cities", fieldCity, m.City, cities, send).
						FullWidth().
						MaxHeight(160)).
						Style(style.New().Width(420)),
				),
				section("Empty",
					layout.Box(listbox.New(nil).Key("empty").
						EmptyText("No matching items").
						FullWidth()).
						Style(style.New().Width(420)),
				),
			).Gap(18),
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

func listBox(key string, field Field, selected string, items []listbox.Item, send ui.Send[Msg]) listbox.Widget {
	return listbox.New(items).Key(key).SelectedKey(selected).
		AllowEmptySelection().
		OnChange(func(selected string) {
			send(Msg{
				Field: field,
				Key:   selected,
			})
		})
}

func actionList(send ui.Send[Msg]) listbox.Widget {
	return listbox.New(actions).Key("actions").
		SelectionMode(listbox.SelectionNone).
		HideIndicator().
		OnAction(func(action string) {
			send(Msg{
				Field: "action",
				Key:   action,
			})
		})
}

func sectionedActionList(send ui.Send[Msg]) listbox.Widget {
	return listbox.Sections("sectioned-actions", "", actionSections).
		SelectionMode(listbox.SelectionNone).
		HideIndicator().
		OnAction(func(action string) {
			send(Msg{
				Field: "section-action",
				Key:   action,
			})
		})
}

func multiListBox(key string, field Field, selected []string, items []listbox.Item, send ui.Send[Msg]) listbox.Widget {
	return listbox.New(items).Key(key).SelectedKeys(selected).
		OnSelectionChange(func(selected []string) {
			send(Msg{
				Field: field,
				Keys:  selected,
			})
		})
}

func avatar(letter string) ui.Widget {
	return layout.Box(text.New(letter).Size(14)).
		Style(style.New().Width(28).Height(28)).
		Align(layout.AlignCenter)
}

func statusIndicator(selected bool) ui.Widget {
	if !selected {
		return nil
	}
	return text.New("on").Size(11)
}

var members = []listbox.Item{
	{Key: "bob", Label: "Bob", Description: "bob@flowui.dev", Leading: avatar("B")},
	{Key: "fred", Label: "Fred", Description: "fred@flowui.dev", Leading: avatar("F")},
	{Key: "martha", Label: "Martha", Description: "martha@flowui.dev", Leading: avatar("M")},
}

var actions = []listbox.Item{
	{Key: "new-file", Label: "New file", Description: "Create a new file", Trailing: text.New("Ctrl+N").Size(12)},
	{Key: "edit-file", Label: "Edit file", Description: "Make changes", Trailing: text.New("Ctrl+E").Size(12)},
	{Key: "delete-file", Label: "Delete file", Description: "Move to trash", Trailing: text.New("Ctrl+Shift+D").Size(12), Variant: listbox.ItemDanger},
}

var actionSections = []listbox.Section{
	{
		Title: "Actions",
		Items: []listbox.Item{
			{Key: "new-file", Label: "New file", Description: "Create a new file", Trailing: text.New("Ctrl+N").Size(12)},
			{Key: "edit-file", Label: "Edit file", Description: "Make changes", Trailing: text.New("Ctrl+E").Size(12)},
		},
	},
	{
		Title: "Danger zone",
		Items: []listbox.Item{
			{Key: "delete-file", Label: "Delete file", Description: "Move to trash", Trailing: text.New("Ctrl+Shift+D").Size(12), Variant: listbox.ItemDanger},
		},
	},
}

var tools = []listbox.Item{
	{Key: "brush", Label: "Brush"},
	{Key: "eraser", Label: "Eraser"},
	{Key: "legacy", Label: "Legacy tool", Description: "Unavailable for this document", Disabled: true},
}

var roles = []listbox.Item{
	{Key: "read", Label: "Read", Description: "View project content", Indicator: statusIndicator},
	{Key: "comment", Label: "Comment", Description: "Discuss changes", Indicator: statusIndicator},
	{Key: "write", Label: "Write", Description: "Create and edit content", Indicator: statusIndicator},
	{Key: "admin", Label: "Admin", Description: "Manage access", Variant: listbox.ItemDanger, Indicator: statusIndicator},
}

var cities = []listbox.Item{
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
	ui.Run(ui.NewProgram(Model{
		Member: "fred",
		Tool:   "brush",
		City:   "shanghai",
		Roles:  []string{"read", "comment"},
	},
		Update, View), window.Title("FlowUI ListBox"),
		window.Size(900, 720),
		window.Theme(base.Light()),
	)
}
