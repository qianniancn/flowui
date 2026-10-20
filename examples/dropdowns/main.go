package main

import (
	"strings"

	"gioui.org/font"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/dropdown"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case SetAction:
		model.LastAction = string(msg)
	case SetSort:
		model.Sort = string(msg)
	case SetVisible:
		model.Visible = append([]string(nil), msg...)
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := "Choose an action"
	if model.LastAction != "" {
		status = "Last action: " + model.LastAction
	}
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Dropdown").Size(24),
				text.New(status).Size(13),
				layout.Divider(),
				section("Actions",
					layout.Wrap(
						basicDropdown(send),
						richDropdown(send),
						longPressDropdown(send),
					).Gap(12).LineGap(12),
				),
				section("Selection",
					layout.Wrap(
						sortDropdown(model, send),
						columnsDropdown(model, send),
					).Gap(12).LineGap(12),
				),
			).Gap(18),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func basicDropdown(send ui.Send[Msg]) dropdown.Widget {
	return dropdown.New("basic-dropdown", button.New("Actions").Variant(button.Secondary), []dropdown.Item{
		{Key: "new-file", Label: "New file", Leading: icon.New(lucide.FilePlus).Size(16), Trailing: keyboardShortcut("N", false)},
		{Key: "open-file", Label: "Open file", Leading: icon.New(lucide.FolderOpen).Size(16), Trailing: keyboardShortcut("O", false)},
		{Key: "save-file", Label: "Save file", Leading: icon.New(lucide.Save).Size(16), Trailing: keyboardShortcut("S", false)},
		{Key: "delete-file", Label: "Delete file", Variant: dropdown.ItemDanger, Leading: icon.New(lucide.Trash).Size(16), Trailing: keyboardShortcut("D", true)},
	}).OnActionEvent(func(event dropdown.ActionEvent) { send(SetAction(event.Key)) })
}

func keyboardShortcut(key string, shift bool) ui.Widget {
	children := []ui.Widget{icon.New(lucide.Command).Size(14)}
	if shift {
		children = append(children, icon.New(lucide.ArrowUp).Size(14))
	}
	children = append(children, text.New(key).Size(14).Weight(font.Medium))
	return layout.Box(layout.Row(children...).Gap(2).AlignMiddle()).
		Style(style.New().Height(24).PaddingLeft(8).PaddingRight(8)).
		Align(layout.AlignCenter)
}

func richDropdown(send ui.Send[Msg]) dropdown.Widget {
	return dropdown.New("account-dropdown", trigger("account-trigger", "Account"), []dropdown.Item{
		{Key: "profile", Label: "Profile", Description: "View account details", Leading: icon.New(lucide.UserRound).Size(16)},
		{Key: "settings", Label: "Settings", Shortcut: "Ctrl+,", Leading: icon.New(lucide.Settings).Size(16)},
		{
			Key: "invite", Label: "Invite people", Leading: icon.New(lucide.UserPlus).Size(16),
			Children: []dropdown.Item{
				{Key: "invite-email", Label: "Invite by email", Leading: icon.New(lucide.Mail).Size(16)},
				{Key: "invite-link", Label: "Share invite link", Leading: icon.New(lucide.Link).Size(16)},
			},
		},
	}).BeforeContent(
		layout.Box(
			layout.Column(
				text.New("Alex Morgan").Size(13),
				text.New("alex@example.com").Size(12),
			).Gap(2),
		).Style(style.New().Padding(12)),
	).OnActionEvent(func(event dropdown.ActionEvent) { send(SetAction(event.Key)) })
}

func longPressDropdown(send ui.Send[Msg]) dropdown.Widget {
	return dropdown.New("long-press-dropdown", trigger("long-press-trigger", "Long press"), []dropdown.Item{
		{Key: "preview", Label: "Preview", Leading: icon.New(lucide.Eye).Size(16)},
		{Key: "download", Label: "Download", Leading: icon.New(lucide.Download).Size(16)},
	}).TriggerMode(dropdown.TriggerLongPress).
		OnActionEvent(func(event dropdown.ActionEvent) { send(SetAction(event.Key)) })
}

func sortDropdown(model Model, send ui.Send[Msg]) dropdown.Widget {
	return dropdown.New("sort-dropdown", trigger("sort-trigger", selectedLabel("Sort", model.Sort)), []dropdown.Item{
		{Key: "name", Label: "Name", IndicatorType: dropdown.IndicatorDot},
		{Key: "modified", Label: "Last modified", IndicatorType: dropdown.IndicatorDot},
		{Key: "size", Label: "Size", IndicatorType: dropdown.IndicatorDot},
	}).SelectionMode(dropdown.SelectionSingle).
		SelectedKey(model.Sort).
		OnChange(func(key string) { send(SetSort(key)) })
}

func columnsDropdown(model Model, send ui.Send[Msg]) dropdown.Widget {
	return dropdown.New("columns-dropdown", trigger("columns-trigger", visibleLabel(model.Visible)), []dropdown.Item{
		{Key: "name", Label: "Name", IndicatorType: dropdown.IndicatorCheckmark},
		{Key: "owner", Label: "Owner", IndicatorType: dropdown.IndicatorCheckmark},
		{Key: "modified", Label: "Last modified", IndicatorType: dropdown.IndicatorCheckmark},
	}).SelectionMode(dropdown.SelectionMultiple).
		SelectedKeys(model.Visible).
		CloseOnSelect(false).
		OnSelectionChange(func(keys []string) { send(SetVisible(keys)) })
}

func trigger(key, label string) ui.Widget {
	return button.New("").Key(key).Content(layout.Row(
		text.New(label),
		icon.New(lucide.ChevronDown).Size(14),
	).Gap(8).AlignMiddle()).Variant(button.Secondary)
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
	return layout.Column(text.New(title).Size(18), child).Gap(10)
}

func main() {
	ui.Run(ui.NewProgram(Model{Sort: "modified", Visible: []string{"name", "modified"}},
		Update, View), window.Title("FlowUI Dropdown"),
		window.Size(900, 620),
		window.Theme(base.Light()),
	)
}
