package ui

import "github.com/qianniancn/flowui/internal/components/panel"

// PanelItem describes one view hosted by PanelHost or ViewStack.
type PanelItem = panel.Item

// PanelHostWidget presents one selected panel and controls hidden-panel
// lifecycle without rendering a navigation strip.
type PanelHostWidget = panel.HostWidget

// PanelHost creates a lifecycle-aware host for mutually exclusive views.
func PanelHost(key, selectedKey string, items []PanelItem) PanelHostWidget {
	return panel.Host(key, selectedKey, items)
}

// ViewStack is an alias-style constructor for PanelHost.
func ViewStack(key, selectedKey string, items []PanelItem) PanelHostWidget {
	return panel.ViewStack(key, selectedKey, items)
}
