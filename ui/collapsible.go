package ui

import "github.com/qianniancn/flowui/internal/components/collapsible"

// CollapsibleItem describes one expandable section.
type CollapsibleItem = collapsible.Item

type CollapsibleWidget = collapsible.Widget

type CollapsibleGroupWidget = collapsible.GroupWidget

// Collapsible creates one expandable section with content.
func Collapsible(key string, expanded bool, label string, content Widget) CollapsibleWidget {
	return collapsible.Collapsible(key, expanded, label, content)
}

// CollapsibleGroup creates multiple expandable sections sharing one state key.
func CollapsibleGroup(key string, expandedKeys []string, items []CollapsibleItem) CollapsibleGroupWidget {
	return collapsible.CollapsibleGroup(key, expandedKeys, items)
}
