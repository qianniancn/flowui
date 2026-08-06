package ui

import "github.com/qianniancn/flowui/internal/components/tabs"

// TabItem describes one tab.
type TabItem = tabs.TabItem

// TabsVariant selects the tab strip's visual treatment.
type TabsVariant = tabs.TabsVariant

// TabsOrientation controls the direction of the tab strip.
type TabsOrientation = tabs.TabsOrientation

// TabsPlacement controls which side of the panel owns the tab strip.
type TabsPlacement = tabs.TabsPlacement

// TabsSize selects the tab strip's control height.
type TabsSize = tabs.TabsSize

// TabsColor selects the tab strip's accent color.
type TabsColor = tabs.TabsColor

// TabsActivationMode controls keyboard selection behavior.
type TabsActivationMode = tabs.TabsActivationMode

// TabsPanelTransition controls selected-panel entry animation.
type TabsPanelTransition = tabs.TabsPanelTransition

// TabsIndicatorAlign controls the selected indicator alignment.
type TabsIndicatorAlign = tabs.TabsIndicatorAlign

// TabsOverflowMode controls how overflowing tab items are exposed.
type TabsOverflowMode = tabs.TabsOverflowMode

type TabsWidget = tabs.TabsWidget

const (
	TabsPrimary   = tabs.TabsPrimary
	TabsSecondary = tabs.TabsSecondary

	TabsHorizontal = tabs.TabsHorizontal
	TabsVertical   = tabs.TabsVertical

	TabsTop    = tabs.TabsTop
	TabsBottom = tabs.TabsBottom
	TabsStart  = tabs.TabsStart
	TabsEnd    = tabs.TabsEnd

	TabsMedium = tabs.TabsMedium
	TabsSmall  = tabs.TabsSmall
	TabsLarge  = tabs.TabsLarge

	TabsColorDefault = tabs.TabsColorDefault
	TabsColorAccent  = tabs.TabsColorAccent

	TabsActivationAutomatic = tabs.TabsActivationAutomatic
	TabsActivationManual    = tabs.TabsActivationManual

	TabsPanelNone = tabs.TabsPanelNone
	TabsPanelFade = tabs.TabsPanelFade

	TabsIndicatorStart  = tabs.TabsIndicatorStart
	TabsIndicatorCenter = tabs.TabsIndicatorCenter
	TabsIndicatorEnd    = tabs.TabsIndicatorEnd

	TabsOverflowScroll = tabs.TabsOverflowScroll
	TabsOverflowMenu   = tabs.TabsOverflowMenu
	TabsOverflowAuto   = tabs.TabsOverflowAuto
)

// Tabs creates a tab strip initialized with selectedKey.
func Tabs(key, selectedKey string, items []TabItem) TabsWidget {
	return tabs.Tabs(key, selectedKey, items)
}
