package command

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/components/button"
	"github.com/qianniancn/flowui/internal/components/menu"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/components/togglebutton"
	"github.com/qianniancn/flowui/internal/components/tooltip"
	"github.com/qianniancn/flowui/internal/frame"
)

func MenuItem(command Command) menu.Item {
	command.validate()
	item := menu.Item{
		Key:         command.key,
		Label:       command.label,
		Description: command.description,
		Shortcut:    command.shortcut.String(),
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

func Button(keyValue string, command Command) frame.Widget {
	if keyValue == "" {
		panic("flowui: empty command button key")
	}
	command.validate()
	child := command.icon
	iconOnly := child != nil
	if child == nil {
		child = text.New(command.label)
	}

	var trigger frame.Widget
	if command.toggle {
		widget := togglebutton.ToggleButton(keyValue, command.checked, child).
			Label(command.label).
			Disabled(command.disabled).
			OnChange(func(bool) { command.execute() })
		if iconOnly {
			widget = widget.IconOnly()
		}
		trigger = widget
	} else {
		if iconOnly {
			child = semanticLabel{label: command.label, child: child}
		}
		variant := button.ButtonSecondary
		if command.danger {
			variant = button.ButtonDangerSoft
		}
		widget := button.Button(keyValue, child).
			Variant(variant).
			Disabled(command.disabled).
			OnClick(command.execute)
		if iconOnly {
			widget = widget.IconOnly()
		}
		trigger = widget
	}
	if !iconOnly {
		return trigger
	}
	return frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		owner := frame.FullKey(ctx, keyValue)
		key := frame.DerivedKey(ctx, owner, "tooltip")
		return tooltip.Tooltip(key, trigger, text.New(command.label)).Delay(0).Layout(ctx, gtx)
	})
}

type semanticLabel struct {
	label string
	child frame.Widget
}

func (w semanticLabel) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	semantic.LabelOp(w.label).Add(gtx.Ops)
	return w.child.Layout(ctx, gtx)
}
