package ui

import "github.com/qianniancn/FlowUI/internal/components/collapsible"

type CollapsibleItem = collapsible.Item
type CollapsibleWidget = collapsible.Widget
type CollapsibleGroupWidget = collapsible.GroupWidget

func Collapsible(key string, expanded bool, label string, content Widget) CollapsibleWidget {
	return collapsible.Collapsible(key, expanded, label, content)
}

func CollapsibleGroup(key string, expandedKeys []string, items []CollapsibleItem) CollapsibleGroupWidget {
	return collapsible.CollapsibleGroup(key, expandedKeys, items)
}
