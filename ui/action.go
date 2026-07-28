package ui

// Action is an alias for Command, providing an alternative name for application actions.
// Use Action when you prefer action-oriented terminology, or Command for command-oriented terminology.
type Action = Command

// ActionScopeWidget is an alias for CommandScopeWidget.
type ActionScopeWidget = CommandScopeWidget

// NewAction creates a new action with the given key and label.
// It is an alias for NewCommand.
func NewAction(key, label string) Action {
	return NewCommand(key, label)
}

// ActionScope installs keyboard shortcuts for actions in the child subtree.
// It is an alias for CommandScope.
func ActionScope(actions []Action, child Widget) ActionScopeWidget {
	return CommandScope(actions, child)
}

// ActionMenuItem converts an Action to a MenuItem.
// It is an alias for CommandMenuItem.
func ActionMenuItem(action Action) MenuItem {
	return CommandMenuItem(action)
}

// ActionButton converts an Action to a Button widget.
// It is an alias for CommandButton.
func ActionButton(key string, action Action) Widget {
	return CommandButton(key, action)
}
