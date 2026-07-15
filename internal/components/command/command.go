package command

import (
	"fmt"
	"runtime"
	"strings"
	"unicode/utf8"

	"gioui.org/io/key"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// ShortcutModifier describes modifiers required by a keyboard shortcut.
type ShortcutModifier uint8

const (
	ShortcutPrimary ShortcutModifier = 1 << iota
	ShortcutControl
	ShortcutCommand
	ShortcutShift
	ShortcutAlt
	ShortcutSuper
)

const allShortcutModifiers = ShortcutPrimary | ShortcutControl | ShortcutCommand | ShortcutShift | ShortcutAlt | ShortcutSuper

// Shortcut is a normalized, platform-aware keyboard shortcut.
type Shortcut struct {
	name      key.Name
	modifiers key.Modifiers
	text      string
}

// KeyShortcut constructs a shortcut. Printable keys require Ctrl, Command,
// Alt, Super, or the platform primary modifier so normal typing is unaffected.
func KeyShortcut(name string, modifiers ShortcutModifier) Shortcut {
	if modifiers&^allShortcutModifiers != 0 {
		panic("flowui: invalid shortcut modifiers")
	}
	resolvedName, display, printable := shortcutName(name)
	resolvedModifiers, labels := shortcutModifiers(modifiers)
	commandModifier := key.ModCtrl | key.ModCommand | key.ModAlt | key.ModSuper
	if printable && resolvedModifiers&commandModifier == 0 {
		panic("flowui: printable shortcut requires Ctrl, Command, Alt, Super, or Primary")
	}
	if display == "+" && resolvedModifiers.Contain(key.ModShift) {
		labels = removeShortcutLabel(labels, "Shift")
	}
	labels = append(labels, display)
	return Shortcut{name: resolvedName, modifiers: resolvedModifiers, text: strings.Join(labels, "+")}
}

func (s Shortcut) String() string {
	return s.text
}

func (s Shortcut) empty() bool {
	return s.name == ""
}

func (s Shortcut) filter() key.Filter {
	return key.Filter{Name: s.name, Required: s.modifiers}
}

func (s Shortcut) matches(event key.Event) bool {
	return event.Name == s.name && event.Modifiers == s.modifiers
}

func shortcutName(value string) (key.Name, string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		panic("flowui: empty shortcut key")
	}
	switch strings.ToLower(value) {
	case "left", "leftarrow":
		return key.NameLeftArrow, "Left", false
	case "right", "rightarrow":
		return key.NameRightArrow, "Right", false
	case "up", "uparrow":
		return key.NameUpArrow, "Up", false
	case "down", "downarrow":
		return key.NameDownArrow, "Down", false
	case "return":
		return key.NameReturn, "Return", false
	case "enter":
		return key.NameEnter, "Enter", false
	case "escape", "esc":
		return key.NameEscape, "Esc", false
	case "home":
		return key.NameHome, "Home", false
	case "end":
		return key.NameEnd, "End", false
	case "backspace":
		return key.NameDeleteBackward, "Backspace", false
	case "delete":
		return key.NameDeleteForward, "Delete", false
	case "pageup":
		return key.NamePageUp, "PageUp", false
	case "pagedown":
		return key.NamePageDown, "PageDown", false
	case "tab":
		return key.NameTab, "Tab", false
	case "space":
		return key.NameSpace, "Space", false
	case "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12":
		name := strings.ToUpper(value)
		return key.Name(name), name, false
	}
	if utf8.RuneCountInString(value) != 1 {
		panic(fmt.Sprintf("flowui: unsupported shortcut key %q", value))
	}
	name := strings.ToUpper(value)
	return key.Name(name), name, true
}

func shortcutModifiers(modifiers ShortcutModifier) (key.Modifiers, []string) {
	resolved := key.Modifiers(0)
	labels := make([]string, 0, 5)
	add := func(modifier key.Modifiers, label string) {
		if resolved.Contain(modifier) {
			return
		}
		resolved |= modifier
		labels = append(labels, label)
	}
	if modifiers&ShortcutPrimary != 0 {
		label := "Ctrl"
		if runtime.GOOS == "darwin" {
			label = "Command"
		}
		add(key.ModShortcut, label)
	}
	if modifiers&ShortcutControl != 0 {
		add(key.ModCtrl, "Ctrl")
	}
	if modifiers&ShortcutCommand != 0 {
		add(key.ModCommand, "Command")
	}
	if modifiers&ShortcutAlt != 0 {
		add(key.ModAlt, "Alt")
	}
	if modifiers&ShortcutShift != 0 {
		add(key.ModShift, "Shift")
	}
	if modifiers&ShortcutSuper != 0 {
		add(key.ModSuper, "Super")
	}
	return resolved, labels
}

func removeShortcutLabel(labels []string, value string) []string {
	for index, label := range labels {
		if label == value {
			return append(labels[:index], labels[index+1:]...)
		}
	}
	return labels
}

// Command describes one reusable application action.
type Command struct {
	key         string
	label       string
	description string
	icon        frame.Widget
	shortcut    Shortcut
	disabled    bool
	toggle      bool
	checked     bool
	danger      bool
	keepOpen    bool
	onExecute   func()
}

func New(keyValue, label string) Command {
	if keyValue == "" {
		panic("flowui: empty command key")
	}
	if label == "" {
		panic(fmt.Sprintf("flowui: empty command label for key %q", keyValue))
	}
	return Command{key: keyValue, label: label}
}

func (c Command) Description(value string) Command {
	c.description = value
	return c
}

func (c Command) Icon(value frame.Widget) Command {
	c.icon = value
	return c
}

func (c Command) Shortcut(value Shortcut) Command {
	c.shortcut = value
	return c
}

func (c Command) Disabled(disabled bool) Command {
	c.disabled = disabled
	return c
}

func (c Command) Toggle(checked bool) Command {
	c.toggle = true
	c.checked = checked
	return c
}

func (c Command) Danger(danger bool) Command {
	c.danger = danger
	return c
}

func (c Command) KeepOpen(keep bool) Command {
	c.keepOpen = keep
	return c
}

func (c Command) OnExecute(fn func()) Command {
	c.onExecute = fn
	return c
}

func (c Command) validate() {
	if c.key == "" {
		panic("flowui: empty command key")
	}
	if c.label == "" {
		panic(fmt.Sprintf("flowui: empty command label for key %q", c.key))
	}
	if !c.disabled && c.onExecute == nil {
		panic(fmt.Sprintf("flowui: enabled command %q requires OnExecute", c.key))
	}
}

func (c Command) execute() {
	if !c.disabled && c.onExecute != nil {
		c.onExecute()
	}
}
