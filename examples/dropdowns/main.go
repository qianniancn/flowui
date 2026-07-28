package main

import (
	"strings"

	"gioui.org/font"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	LastAction string
	Sort       string
	Visible    []string
}

type Msg any

type SetAction string
type SetSort string
type SetVisible []string

func Update(model *Model, msg Msg) {
	switch msg := msg.(type) {
	case SetAction:
		model.LastAction = string(msg)
	case SetSort:
		model.Sort = string(msg)
	case SetVisible:
		model.Visible = append([]string(nil), msg...)
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "Choose an action"
	if model.LastAction != "" {
		status = "Last action: " + model.LastAction
	}
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Dropdown").Size(24),
				ui.Text(status).Size(13),
				ui.Divider(),
				section("Actions",
					ui.Wrap(
						basicDropdown(send),
						richDropdown(send),
						longPressDropdown(send),
					).Gap(12).LineGap(12),
				),
				section("Selection",
					ui.Wrap(
						sortDropdown(model, send),
						columnsDropdown(model, send),
					).Gap(12).LineGap(12),
				),
			).Gap(18),
		).Style(ui.FillWidth().MaxWidth(760).Padding(24)),
	)
}

func basicDropdown(send ui.Send[Msg]) ui.DropdownWidget {
	return ui.Dropdown("basic-dropdown", ui.Button("basic-trigger", ui.Text("Actions")).Variant(ui.ButtonSecondary), []ui.DropdownItem{
		{Key: "new-file", Label: "New file", Leading: ui.Icon(lucide.FilePlus).Size(16), Trailing: keyboardShortcut("N", false)},
		{Key: "open-file", Label: "Open file", Leading: ui.Icon(lucide.FolderOpen).Size(16), Trailing: keyboardShortcut("O", false)},
		{Key: "save-file", Label: "Save file", Leading: ui.Icon(lucide.Save).Size(16), Trailing: keyboardShortcut("S", false)},
		{Key: "delete-file", Label: "Delete file", Variant: ui.DropdownItemDanger, Leading: ui.Icon(lucide.Trash2).Size(16), Trailing: keyboardShortcut("D", true)},
	}).OnAction(func(key string) { send(SetAction(key)) })
}

func keyboardShortcut(key string, shift bool) ui.Widget {
	children := []ui.Widget{ui.Icon(lucide.Command).Size(14)}
	if shift {
		children = append(children, ui.Icon(lucide.ArrowUp).Size(14))
	}
	children = append(children, ui.Text(key).Size(14).Weight(font.Medium))
	return ui.Box(ui.Row(children...).Gap(2).AlignMiddle()).
		Style(ui.Height(24).PaddingLeft(8).PaddingRight(8)).
		Align(ui.AlignCenter)
}

func richDropdown(send ui.Send[Msg]) ui.DropdownWidget {
	return ui.Dropdown("account-dropdown", trigger("account-trigger", "Account"), []ui.DropdownItem{
		{Key: "profile", Label: "Profile", Description: "View account details", Leading: ui.Icon(lucide.UserRound).Size(16)},
		{Key: "settings", Label: "Settings", Shortcut: "Ctrl+,", Leading: ui.Icon(lucide.Settings).Size(16)},
		{
			Key: "invite", Label: "Invite people", Leading: ui.Icon(lucide.UserPlus).Size(16),
			Children: []ui.DropdownItem{
				{Key: "invite-email", Label: "Invite by email", Leading: ui.Icon(lucide.Mail).Size(16)},
				{Key: "invite-link", Label: "Share invite link", Leading: ui.Icon(lucide.Link).Size(16)},
			},
		},
	}).BeforeContent(
		ui.Box(
			ui.Column(
				ui.Text("Alex Morgan").Size(13),
				ui.Text("alex@example.com").Size(12),
			).Gap(2),
		).Style(ui.Padding(12)),
	).OnAction(func(key string) { send(SetAction(key)) })
}

func longPressDropdown(send ui.Send[Msg]) ui.DropdownWidget {
	return ui.Dropdown("long-press-dropdown", trigger("long-press-trigger", "Long press"), []ui.DropdownItem{
		{Key: "preview", Label: "Preview", Leading: ui.Icon(lucide.Eye).Size(16)},
		{Key: "download", Label: "Download", Leading: ui.Icon(lucide.Download).Size(16)},
	}).TriggerMode(ui.DropdownTriggerLongPress).
		OnAction(func(key string) { send(SetAction(key)) })
}

func sortDropdown(model Model, send ui.Send[Msg]) ui.DropdownWidget {
	return ui.Dropdown("sort-dropdown", trigger("sort-trigger", selectedLabel("Sort", model.Sort)), []ui.DropdownItem{
		{Key: "name", Label: "Name", IndicatorType: ui.DropdownIndicatorDot},
		{Key: "modified", Label: "Last modified", IndicatorType: ui.DropdownIndicatorDot},
		{Key: "size", Label: "Size", IndicatorType: ui.DropdownIndicatorDot},
	}).SelectionMode(ui.DropdownSelectionSingle).
		SelectedKey(model.Sort).
		OnChange(func(key string) { send(SetSort(key)) })
}

func columnsDropdown(model Model, send ui.Send[Msg]) ui.DropdownWidget {
	return ui.Dropdown("columns-dropdown", trigger("columns-trigger", visibleLabel(model.Visible)), []ui.DropdownItem{
		{Key: "name", Label: "Name", IndicatorType: ui.DropdownIndicatorCheckmark},
		{Key: "owner", Label: "Owner", IndicatorType: ui.DropdownIndicatorCheckmark},
		{Key: "modified", Label: "Last modified", IndicatorType: ui.DropdownIndicatorCheckmark},
	}).SelectionMode(ui.DropdownSelectionMultiple).
		SelectedKeys(model.Visible).
		CloseOnSelect(false).
		OnSelectionChange(func(keys []string) { send(SetVisible(keys)) })
}

func trigger(key, label string) ui.Widget {
	return ui.Button(key,
		ui.Row(
			ui.Text(label),
			ui.Icon(lucide.ChevronDown).Size(14),
		).Gap(8).AlignMiddle(),
	).Variant(ui.ButtonSecondary)
}

func selectedLabel(prefix, selected string) string {
	if selected == "" {
		return prefix
	}
	return prefix + ": " + selected
}

func visibleLabel(keys []string) string {
	if len(keys) == 0 {
		return "Visible columns"
	}
	return "Columns: " + strings.Join(keys, ", ")
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func main() {
	ui.Run(
		Model{Sort: "modified", Visible: []string{"name", "modified"}},
		Update,
		View,
		ui.Title("FlowUI Dropdown"),
		ui.Size(900, 620),
	)
}
