//go:build linux && !android

/*
Portions of this code are derived from:
- https://github.com/fyne-io/systray
- https://github.com/wailsapp/wails/v3
*/

package systray

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"sync"
	"sync/atomic"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	"github.com/qianniancn/flowui/internal/sys/linux/dbus/menu"
	"github.com/qianniancn/flowui/internal/sys/linux/dbus/notifier"
)

const (
	itemPath = "/StatusNotifierItem"
	menuPath = "/StatusNotifierMenu"
)

type linuxSystemTray struct {
	parent  *SystemTray
	stateMu sync.RWMutex

	id      uint
	label   string
	tooltip string
	icon    []byte
	menu    *Menu

	quitChan  chan struct{}
	conn      *dbus.Conn
	props     *prop.Properties
	menuProps *prop.Properties

	menuVersion atomic.Uint32

	// itemMapLock guards itemMap and the per-item dbusItem.V1 maps
	itemMapLock sync.RWMutex
	itemMap     map[int32]*dbusMenuItem

	lastClickX, lastClickY int32
}

type statusNotifierPixmap struct {
	Width  int32
	Height int32
	Data   []byte
}

type dbusMenuItem struct {
	item     *MenuItem
	children []int32
}

func newSystemTrayImpl(parent *SystemTray) systemTrayImpl {
	return &linuxSystemTray{
		parent:   parent,
		id:       parent.id,
		quitChan: make(chan struct{}),
		itemMap:  make(map[int32]*dbusMenuItem),
	}
}

func (s *linuxSystemTray) setLabel(label string) {
	s.stateMu.Lock()
	s.label = label
	s.stateMu.Unlock()
	if err := s.setProperty("Title", label); err != nil {
		s.parent.reportError(fmt.Errorf("systray: update Linux title: %w", err))
	}
}

func (s *linuxSystemTray) setTooltip(tooltip string) {
	s.stateMu.Lock()
	s.tooltip = tooltip
	s.stateMu.Unlock()
	tooltipData := struct {
		IconName string
		Icons    []struct {
			Width  int32
			Height int32
			Data   []byte
		}
		Title       string
		Description string
	}{
		IconName: "",
		Icons: []struct {
			Width, Height int32
			Data          []byte
		}{},
		Title:       tooltip,
		Description: "",
	}
	if err := s.setProperty("ToolTip", tooltipData); err != nil {
		s.parent.reportError(fmt.Errorf("systray: update Linux tooltip: %w", err))
	}
}

func (s *linuxSystemTray) setIcon(iconData []byte) {
	pixmaps, err := statusNotifierPixmapsFromPNG(iconData)
	if err != nil {
		s.parent.reportError(fmt.Errorf("systray: decode Linux icon: %w", err))
		return
	}
	s.stateMu.Lock()
	s.icon = append(s.icon[:0], iconData...)
	s.stateMu.Unlock()
	if err := s.setProperty("IconPixmap", pixmaps); err != nil {
		s.parent.reportError(fmt.Errorf("systray: update Linux IconPixmap: %w", err))
		return
	}
	s.emitSignal("NewIcon")
}

func (s *linuxSystemTray) setMenu(m *Menu) {
	if m != nil {
		m.processRadioGroups()
	}
	s.stateMu.Lock()
	s.menu = m
	s.stateMu.Unlock()
	version := s.menuVersion.Add(1)
	s.emitMenuSignal("LayoutUpdated", version, int32(0))
}

func (s *linuxSystemTray) loadInitialState() error {
	s.parent.mu.Lock()
	s.stateMu.Lock()
	s.label = s.parent.label
	s.tooltip = s.parent.tooltip
	s.icon = append(s.icon[:0], s.parent.icon...)
	s.menu = s.parent.menu
	icon := append([]byte(nil), s.icon...)
	configuredMenu := s.menu
	s.stateMu.Unlock()
	s.parent.mu.Unlock()
	if configuredMenu != nil {
		configuredMenu.processRadioGroups()
		s.menuVersion.Store(1)
	}
	if _, err := statusNotifierPixmapsFromPNG(icon); err != nil {
		return fmt.Errorf("decode initial Linux icon: %w", err)
	}
	return nil
}

func (s *linuxSystemTray) run() {
	if err := s.loadInitialState(); err != nil {
		s.parent.fail(fmt.Errorf("systray: %w", err))
		return
	}
	var err error
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		s.parent.fail(fmt.Errorf("systray: connect to Linux session bus: %w", err))
		return
	}
	s.stateMu.Lock()
	s.conn = conn
	s.stateMu.Unlock()
	defer conn.Close()

	// Register StatusNotifierItem
	if err := s.registerStatusNotifierItem(); err != nil {
		s.parent.fail(fmt.Errorf("systray: export Linux StatusNotifierItem: %w", err))
		return
	}

	// Register with StatusNotifierWatcher
	if err := s.registerWithWatcher(); err != nil {
		s.parent.fail(fmt.Errorf("systray: register with Linux StatusNotifierWatcher: %w", err))
		return
	}
	s.parent.markReady()

	select {
	case <-s.quitChan:
		s.parent.finish()
	case <-conn.Context().Done():
		s.parent.fail(fmt.Errorf("systray: Linux session bus connection closed: %w", conn.Context().Err()))
	}
}

func (s *linuxSystemTray) destroy() {
	close(s.quitChan)
	if conn := s.connection(); conn != nil {
		conn.Close()
	}
}

func (s *linuxSystemTray) show() {
	if err := s.setProperty("Status", "Active"); err != nil {
		s.parent.reportError(fmt.Errorf("systray: show Linux tray icon: %w", err))
		return
	}
	s.emitSignal("NewStatus", "Active")
}

func (s *linuxSystemTray) hide() {
	if err := s.setProperty("Status", "Passive"); err != nil {
		s.parent.reportError(fmt.Errorf("systray: hide Linux tray icon: %w", err))
		return
	}
	s.emitSignal("NewStatus", "Passive")
}

func (s *linuxSystemTray) setProperty(name string, value interface{}) (err error) {
	s.stateMu.RLock()
	properties := s.props
	s.stateMu.RUnlock()
	if properties == nil {
		return nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()
	properties.SetMust("org.kde.StatusNotifierItem", name, value)
	return nil
}

// D-Bus interface implementations

func (s *linuxSystemTray) ContextMenu(x, y int32) *dbus.Error {
	s.stateMu.Lock()
	s.lastClickX, s.lastClickY = x, y
	s.stateMu.Unlock()
	// Context menu is handled via DBusMenu
	return nil
}

func (s *linuxSystemTray) Activate(x, y int32) *dbus.Error {
	s.stateMu.Lock()
	s.lastClickX, s.lastClickY = x, y
	s.stateMu.Unlock()
	s.parent.mu.Lock()
	handler := s.parent.clickHandler
	s.parent.mu.Unlock()

	if handler != nil {
		go handler()
	}
	return nil
}

func (s *linuxSystemTray) SecondaryActivate(x, y int32) *dbus.Error {
	s.stateMu.Lock()
	s.lastClickX, s.lastClickY = x, y
	s.stateMu.Unlock()
	s.parent.mu.Lock()
	handler := s.parent.rightClickHandler
	s.parent.mu.Unlock()

	if handler != nil {
		go handler()
	}
	return nil
}

func (s *linuxSystemTray) Scroll(delta int32, orientation string) *dbus.Error {
	// Scroll not implemented yet
	return nil
}

// DBusMenu interface implementations

func (s *linuxSystemTray) GetLayout(parentId int32, recursionDepth int32, propertyNames []string) (
	revision uint32,
	layout struct {
		V0 int32
		V1 map[string]dbus.Variant
		V2 []dbus.Variant
	},
	err *dbus.Error,
) {
	revision = s.menuVersion.Load()

	if s.menuSnapshot() == nil {
		layout.V0 = 0
		layout.V1 = make(map[string]dbus.Variant)
		layout.V2 = []dbus.Variant{}
		return
	}

	layout = s.buildMenuLayout(parentId, recursionDepth)
	return
}

func (s *linuxSystemTray) GetGroupProperties(ids []int32, propertyNames []string) (
	properties []struct {
		V0 int32
		V1 map[string]dbus.Variant
	},
	err *dbus.Error,
) {
	properties = make([]struct {
		V0 int32
		V1 map[string]dbus.Variant
	}, len(ids))

	for i, id := range ids {
		properties[i].V0 = id
		properties[i].V1 = s.getMenuItemProperties(id)
	}
	return
}

func (s *linuxSystemTray) GetProperty(id int32, name string) (value dbus.Variant, err *dbus.Error) {
	props := s.getMenuItemProperties(id)
	if v, ok := props[name]; ok {
		value = v
	}
	return
}

func (s *linuxSystemTray) Event(id int32, eventId string, data dbus.Variant, timestamp uint32) *dbus.Error {
	if eventId == "clicked" {
		item := GetMenuItemByID(uint(id))
		if item != nil {
			go item.Click()
		}
	}
	return nil
}

func (s *linuxSystemTray) EventGroup(events []struct {
	V0 int32
	V1 string
	V2 dbus.Variant
	V3 uint32
}) (idErrors []int32, err *dbus.Error) {
	for _, event := range events {
		s.Event(event.V0, event.V1, event.V2, event.V3)
	}
	return
}

func (s *linuxSystemTray) AboutToShow(id int32) (needUpdate bool, err *dbus.Error) {
	needUpdate = false
	return
}

func (s *linuxSystemTray) AboutToShowGroup(ids []int32) (updatesNeeded []int32, idErrors []int32, err *dbus.Error) {
	return
}

// Helper functions

func (s *linuxSystemTray) registerStatusNotifierItem() error {
	conn := s.connection()
	if conn == nil {
		return fmt.Errorf("Linux session bus is not connected")
	}
	// Export StatusNotifierItem interface
	if err := notifier.ExportStatusNotifierItem(conn, itemPath, s); err != nil {
		return err
	}

	// Export properties
	properties, err := prop.Export(conn, itemPath, s.createPropSpec())
	if err != nil {
		return err
	}
	s.stateMu.Lock()
	s.props = properties
	s.stateMu.Unlock()

	// Export introspection
	node := &introspect.Node{
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			notifier.IntrospectDataStatusNotifierItem,
		},
	}
	if err := conn.Export(introspect.NewIntrospectable(node), itemPath, "org.freedesktop.DBus.Introspectable"); err != nil {
		return err
	}

	// Export DBusMenu interface
	if err := menu.ExportDbusmenu(conn, menuPath, s); err != nil {
		return err
	}

	// Export menu properties
	menuProperties, err := prop.Export(conn, menuPath, s.createMenuPropSpec())
	if err != nil {
		return err
	}
	s.stateMu.Lock()
	s.menuProps = menuProperties
	s.stateMu.Unlock()

	// Export menu introspection
	menuNode := &introspect.Node{
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			menu.IntrospectDataDbusmenu,
		},
	}
	if err := conn.Export(introspect.NewIntrospectable(menuNode), menuPath, "org.freedesktop.DBus.Introspectable"); err != nil {
		return err
	}

	return nil
}

func (s *linuxSystemTray) registerWithWatcher() error {
	conn := s.connection()
	if conn == nil {
		return fmt.Errorf("Linux session bus is not connected")
	}
	names := conn.Names()
	if len(names) == 0 {
		return fmt.Errorf("Linux session bus has no unique name")
	}
	busName := names[0]
	obj := conn.Object("org.kde.StatusNotifierWatcher", "/StatusNotifierWatcher")
	call := obj.Call("org.kde.StatusNotifierWatcher.RegisterStatusNotifierItem", 0, busName)
	return call.Err
}

func (s *linuxSystemTray) emitSignal(name string, args ...interface{}) {
	if conn := s.connection(); conn != nil {
		_ = conn.Emit(itemPath, "org.kde.StatusNotifierItem."+name, args...)
	}
}

func (s *linuxSystemTray) emitMenuSignal(name string, args ...interface{}) {
	if conn := s.connection(); conn != nil {
		_ = conn.Emit(menuPath, "com.canonical.dbusmenu."+name, args...)
	}
}

func (s *linuxSystemTray) createPropSpec() map[string]map[string]*prop.Prop {
	s.stateMu.RLock()
	icon := append([]byte(nil), s.icon...)
	label := s.label
	tooltip := s.tooltip
	s.stateMu.RUnlock()
	iconPixmaps, _ := statusNotifierPixmapsFromPNG(icon)
	return map[string]map[string]*prop.Prop{
		"org.kde.StatusNotifierItem": {
			"Status": {
				Value:    "Active",
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"Title": {
				Value:    label,
				Writable: true,
				Emit:     prop.EmitTrue,
			},
			"Id": {
				Value:    fmt.Sprintf("flowui-%d", s.id),
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"Category": {
				Value:    "ApplicationStatus",
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"IconPixmap": {
				Value:    iconPixmaps,
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"ToolTip": {
				Value: struct {
					IconName string
					Icons    []struct {
						Width  int32
						Height int32
						Data   []byte
					}
					Title       string
					Description string
				}{Title: tooltip},
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"Menu": {
				Value:    dbus.ObjectPath(menuPath),
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"ItemIsMenu": {
				Value:    true,
				Writable: false,
				Emit:     prop.EmitTrue,
			},
		},
	}
}

func statusNotifierPixmapsFromPNG(data []byte) ([]statusNotifierPixmap, error) {
	if len(data) == 0 {
		return []statusNotifierPixmap{}, nil
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	pixels := make([]byte, width*height*4)
	offset := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			pixels[offset] = pixel.A
			pixels[offset+1] = pixel.R
			pixels[offset+2] = pixel.G
			pixels[offset+3] = pixel.B
			offset += 4
		}
	}
	return []statusNotifierPixmap{{
		Width:  int32(width),
		Height: int32(height),
		Data:   pixels,
	}}, nil
}

func (s *linuxSystemTray) createMenuPropSpec() map[string]map[string]*prop.Prop {
	return map[string]map[string]*prop.Prop{
		"com.canonical.dbusmenu": {
			"Version": {
				Value:    uint32(3),
				Writable: false,
				Emit:     prop.EmitTrue,
			},
			"Status": {
				Value:    "normal",
				Writable: false,
				Emit:     prop.EmitTrue,
			},
		},
	}
}

func (s *linuxSystemTray) buildMenuLayout(parentId int32, depth int32) struct {
	V0 int32
	V1 map[string]dbus.Variant
	V2 []dbus.Variant
} {
	result := struct {
		V0 int32
		V1 map[string]dbus.Variant
		V2 []dbus.Variant
	}{
		V0: parentId,
		V1: make(map[string]dbus.Variant),
		V2: []dbus.Variant{},
	}

	activeMenu := s.menuSnapshot()
	if activeMenu == nil || depth == 0 {
		return result
	}

	if parentId == 0 {
		// Root menu
		result.V1["children-display"] = dbus.MakeVariant("submenu")

		items := activeMenu.Items()
		for _, item := range items {
			itemLayout := s.buildMenuItemLayout(item, depth-1)
			result.V2 = append(result.V2, dbus.MakeVariant(itemLayout))
		}
	}

	return result
}

func (s *linuxSystemTray) connection() *dbus.Conn {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.conn
}

func (s *linuxSystemTray) menuSnapshot() *Menu {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.menu
}

func (s *linuxSystemTray) buildMenuItemLayout(item *MenuItem, depth int32) struct {
	V0 int32
	V1 map[string]dbus.Variant
	V2 []dbus.Variant
} {
	dbusItem := &dbusMenuItem{item: item}
	s.itemMapLock.Lock()
	s.itemMap[int32(item.ID())] = dbusItem
	s.itemMapLock.Unlock()

	result := struct {
		V0 int32
		V1 map[string]dbus.Variant
		V2 []dbus.Variant
	}{
		V0: int32(item.ID()),
		V1: s.getMenuItemProperties(int32(item.ID())),
		V2: []dbus.Variant{},
	}

	if depth > 0 && item.Type() == MenuItemSubmenu && item.Submenu() != nil {
		submenu := item.Submenu()
		items := submenu.Items()
		childIDs := make([]int32, 0, len(items))

		for _, child := range items {
			childLayout := s.buildMenuItemLayout(child, depth-1)
			result.V2 = append(result.V2, dbus.MakeVariant(childLayout))
			childIDs = append(childIDs, int32(child.ID()))
		}

		dbusItem.children = childIDs
	}

	return result
}

func (s *linuxSystemTray) getMenuItemProperties(id int32) map[string]dbus.Variant {
	props := make(map[string]dbus.Variant)

	s.itemMapLock.RLock()
	dbusItem, exists := s.itemMap[id]
	s.itemMapLock.RUnlock()

	if !exists {
		return props
	}

	item := dbusItem.item
	props["label"] = dbus.MakeVariant(item.Label())
	props["enabled"] = dbus.MakeVariant(!item.IsDisabled())
	props["visible"] = dbus.MakeVariant(!item.IsHidden())

	switch item.Type() {
	case MenuItemSeparator:
		props["type"] = dbus.MakeVariant("separator")
	case MenuItemCheckbox:
		props["toggle-type"] = dbus.MakeVariant("checkmark")
		if item.IsChecked() {
			props["toggle-state"] = dbus.MakeVariant(int32(1))
		} else {
			props["toggle-state"] = dbus.MakeVariant(int32(0))
		}
	case MenuItemRadio:
		props["toggle-type"] = dbus.MakeVariant("radio")
		if item.IsChecked() {
			props["toggle-state"] = dbus.MakeVariant(int32(1))
		} else {
			props["toggle-state"] = dbus.MakeVariant(int32(0))
		}
	case MenuItemSubmenu:
		props["children-display"] = dbus.MakeVariant("submenu")
	}

	return props
}
