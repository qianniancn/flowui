package systray

import (
	"sync"
	"sync/atomic"
)

var (
	menuItemIDCounter uint32
	menuItemMap       = make(map[uint]*MenuItem)
	menuItemMapMu     sync.RWMutex
)

// MenuItem represents a menu item
type MenuItem struct {
	id       uint
	label    string
	tooltip  string
	disabled bool
	checked  bool
	hidden   bool
	itemType MenuItemType
	submenu  *Menu
	callback func()

	impl              menuItemImpl
	radioGroupMembers []*MenuItem
	mu                sync.Mutex
}

// menuItemImpl is the platform-specific menu item interface
type menuItemImpl interface {
	setLabel(label string)
	setTooltip(tooltip string)
	setDisabled(disabled bool)
	setChecked(checked bool)
	setHidden(hidden bool)
	destroy()
}

// newMenuItem creates a new menu item
func newMenuItem(label string, itemType MenuItemType) *MenuItem {
	id := atomic.AddUint32(&menuItemIDCounter, 1)
	item := &MenuItem{
		id:       uint(id),
		label:    label,
		itemType: itemType,
	}

	menuItemMapMu.Lock()
	menuItemMap[item.id] = item
	menuItemMapMu.Unlock()

	return item
}

// GetMenuItemByID retrieves a menu item by ID (internal use)
func GetMenuItemByID(id uint) *MenuItem {
	menuItemMapMu.RLock()
	defer menuItemMapMu.RUnlock()
	return menuItemMap[id]
}

// OnClick sets the click handler
func (mi *MenuItem) OnClick(handler func()) *MenuItem {
	mi.mu.Lock()
	mi.callback = handler
	mi.mu.Unlock()
	return mi
}

// SetLabel sets the label
func (mi *MenuItem) SetLabel(label string) *MenuItem {
	mi.mu.Lock()
	mi.label = label
	impl := mi.impl
	mi.mu.Unlock()

	if impl != nil {
		impl.setLabel(label)
	}
	return mi
}

// SetTooltip sets the tooltip
func (mi *MenuItem) SetTooltip(tooltip string) *MenuItem {
	mi.mu.Lock()
	mi.tooltip = tooltip
	impl := mi.impl
	mi.mu.Unlock()

	if impl != nil {
		impl.setTooltip(tooltip)
	}
	return mi
}

// SetDisabled sets the disabled state
func (mi *MenuItem) SetDisabled(disabled bool) *MenuItem {
	mi.mu.Lock()
	mi.disabled = disabled
	impl := mi.impl
	mi.mu.Unlock()

	if impl != nil {
		impl.setDisabled(disabled)
	}
	return mi
}

// SetEnabled sets the enabled state
func (mi *MenuItem) SetEnabled(enabled bool) *MenuItem {
	return mi.SetDisabled(!enabled)
}

// SetChecked sets the checked state
func (mi *MenuItem) SetChecked(checked bool) *MenuItem {
	mi.mu.Lock()
	mi.checked = checked
	itemType := mi.itemType
	radioGroupMembers := mi.radioGroupMembers
	impl := mi.impl
	mi.mu.Unlock()

	// Handle radio group logic
	if itemType == MenuItemRadio && checked && len(radioGroupMembers) > 0 {
		for _, member := range radioGroupMembers {
			if member != mi {
				member.SetChecked(false)
			}
		}
	}

	if impl != nil {
		impl.setChecked(checked)
	}
	return mi
}

// SetHidden sets the hidden state
func (mi *MenuItem) SetHidden(hidden bool) *MenuItem {
	mi.mu.Lock()
	mi.hidden = hidden
	impl := mi.impl
	mi.mu.Unlock()

	if impl != nil {
		impl.setHidden(hidden)
	}
	return mi
}

// Label returns the label
func (mi *MenuItem) Label() string {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	return mi.label
}

// Tooltip returns the tooltip
func (mi *MenuItem) Tooltip() string {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	return mi.tooltip
}

// IsDisabled returns the disabled state
func (mi *MenuItem) IsDisabled() bool {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	return mi.disabled
}

// IsEnabled returns the enabled state
func (mi *MenuItem) IsEnabled() bool {
	return !mi.IsDisabled()
}

// IsChecked returns the checked state
func (mi *MenuItem) IsChecked() bool {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	return mi.checked
}

// IsHidden returns the hidden state
func (mi *MenuItem) IsHidden() bool {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	return mi.hidden
}

// Type returns the item type
func (mi *MenuItem) Type() MenuItemType {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	return mi.itemType
}

// Submenu returns the submenu (if any)
func (mi *MenuItem) Submenu() *Menu {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	return mi.submenu
}

// ID returns the item ID
func (mi *MenuItem) ID() uint {
	return mi.id
}

// Click triggers the callback
func (mi *MenuItem) Click() {
	mi.mu.Lock()
	callback := mi.callback
	itemType := mi.itemType
	mi.mu.Unlock()

	// Handle checkbox toggle
	if itemType == MenuItemCheckbox {
		mi.SetChecked(!mi.IsChecked())
	}

	// Handle radio selection
	if itemType == MenuItemRadio {
		mi.SetChecked(true)
	}

	if callback != nil {
		callback()
	}
}

// destroy cleans up the menu item
func (mi *MenuItem) destroy() {
	mi.mu.Lock()
	impl := mi.impl
	submenu := mi.submenu
	mi.mu.Unlock()

	if impl != nil {
		impl.destroy()
	}

	if submenu != nil {
		submenu.Destroy()
	}

	// Remove from map
	menuItemMapMu.Lock()
	delete(menuItemMap, mi.id)
	menuItemMapMu.Unlock()
}
