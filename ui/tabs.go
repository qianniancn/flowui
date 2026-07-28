package ui

import "github.com/qianniancn/flowui/internal/components/tabs"

type TabItem = tabs.TabItem
type TabsVariant = tabs.TabsVariant
type TabsOrientation = tabs.TabsOrientation
type TabsSize = tabs.TabsSize
type TabsColor = tabs.TabsColor
type TabsWidget = tabs.TabsWidget

const (
	TabsPrimary   = tabs.TabsPrimary
	TabsSecondary = tabs.TabsSecondary

	TabsHorizontal = tabs.TabsHorizontal
	TabsVertical   = tabs.TabsVertical

	TabsMedium = tabs.TabsMedium
	TabsSmall  = tabs.TabsSmall

	TabsColorDefault = tabs.TabsColorDefault
	TabsColorAccent  = tabs.TabsColorAccent
)

func Tabs(key, selectedKey string, items []TabItem) TabsWidget {
	return tabs.Tabs(key, selectedKey, items)
}
