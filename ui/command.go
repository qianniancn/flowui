package ui

import "github.com/qianniancn/flowui/internal/components/command"

// Command describes an action that can be invoked by a menu item, button, or shortcut.
type Command = command.Command

// Shortcut describes the key combination used to invoke a command.
type Shortcut = command.Shortcut

// ShortcutModifier is a bitmask of modifier keys for a Shortcut.
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

// NewCommand returns a command with the given key and label.
func NewCommand(key, label string) Command {
	return command.New(key, label)
}

// KeyShortcut creates a keyboard shortcut descriptor.
func KeyShortcut(key string, modifiers ShortcutModifier) Shortcut {
	return command.KeyShortcut(key, modifiers)
}

// CommandScope installs commands and shortcuts for child.
func CommandScope(commands []Command, child Widget) CommandScopeWidget {
	return command.Scope(commands, child)
}

// CommandMenuItem converts a command into a menu item.
func CommandMenuItem(value Command) MenuItem {
	return command.MenuItem(value)
}

// CommandButton creates a button bound to a command.
func CommandButton(key string, value Command) Widget {
	return command.Button(key, value)
}
