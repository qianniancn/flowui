package command

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/togglebutton"
	"github.com/qianniancn/flowui/widget/tooltip"
)

func MenuItem(command Command) menu.Item {
	return MenuItemFor(nil, command)
}

// MenuItemFor adapts a command to a menu item using the supplied window
// context for localized framework-owned shortcut names.
func MenuItemFor(ctx *core.Context, command Command) menu.Item {
	command.validate()
	item := menu.Item{
		Key:         command.key,
		Label:       command.label,
		Description: command.description,
		Shortcut:    command.shortcut.StringFor(ctx),
		Leading:     command.icon,
		Disabled:    command.disabled,
		Checked:     command.checked,
		KeepOpen:    command.keepOpen,
		OnAction:    command.execute,
	}
	if command.toggle {
		item.Kind = menu.ItemCheckbox
	}
	if command.danger {
		item.Variant = menu.ItemDanger
	}
	return item
}

func Button(keyValue string, command Command) core.Widget {
	if keyValue == "" {
		panic("flowui: empty command button key")
	}
	command.validate()
	iconOnly := command.icon != nil

	var trigger core.Widget
	if command.toggle {
		widget := togglebutton.New(command.label).Key(keyValue).
			Selected(command.checked).Disabled(command.disabled).
			OnChange(func(bool) { command.execute() })
		if iconOnly {
			widget = widget.Content(command.icon).IconOnly()
		}
		trigger = widget
	} else {
		variant := button.Secondary
		if command.danger {
			variant = button.DangerSoft
		}
		widget := button.New(command.label).Key(keyValue).
			Variant(variant).
			Disabled(command.disabled).
			OnClick(command.execute)
		if iconOnly {
			widget = widget.Content(command.icon).IconOnly()
		}
		trigger = widget
	}
	if !iconOnly {
		return trigger
	}
	return core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		owner := core.FullKey(ctx, keyValue)
		key := core.DerivedKey(ctx, owner, "tooltip")
		return tooltip.New(key, trigger, command.label).Delay(0).Layout(ctx, gtx)
	})
}
