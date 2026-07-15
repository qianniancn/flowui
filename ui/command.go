package ui

import "github.com/qianniancn/FlowUI/internal/components/command"

type Command = command.Command
type Shortcut = command.Shortcut
type ShortcutModifier = command.ShortcutModifier
type CommandScopeWidget = command.ScopeWidget

const (
	ShortcutPrimary = command.ShortcutPrimary
	ShortcutControl = command.ShortcutControl
	ShortcutCommand = command.ShortcutCommand
	ShortcutShift   = command.ShortcutShift
	ShortcutAlt     = command.ShortcutAlt
	ShortcutSuper   = command.ShortcutSuper
)

func NewCommand(key, label string) Command {
	return command.New(key, label)
}

func KeyShortcut(key string, modifiers ShortcutModifier) Shortcut {
	return command.KeyShortcut(key, modifiers)
}

func CommandScope(commands []Command, child Widget) CommandScopeWidget {
	return command.Scope(commands, child)
}

func CommandMenuItem(value Command) MenuItem {
	return command.MenuItem(value)
}

func CommandButton(key string, value Command) Widget {
	return command.Button(key, value)
}
