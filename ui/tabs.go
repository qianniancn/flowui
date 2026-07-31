package ui

import "github.com/qianniancn/flowui/internal/components/tabs"

// TabItem describes one tab.
type TabItem = tabs.TabItem

// TabsVariant selects the tab strip's visual treatment.
type TabsVariant = tabs.TabsVariant

// TabsOrientation controls the direction of the tab strip.
type TabsOrientation = tabs.TabsOrientation

// TabsSize selects the tab strip's control height.
type TabsSize = tabs.TabsSize

// TabsColor selects the tab strip's accent color.
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

// Tabs creates a tab strip initialized with selectedKey.
func Tabs(key, selectedKey string, items []TabItem) TabsWidget {
	return tabs.Tabs(key, selectedKey, items)
}
