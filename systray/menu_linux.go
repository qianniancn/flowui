//go:build linux && !android

package systray

// linuxMenu implements menuImpl for Linux
type linuxMenu struct {
	parent *Menu
}

func (l *linuxMenu) update() {
	// Menu updates are handled via D-Bus signals in linuxSystemTray
}

func (l *linuxMenu) show() {
	// Menu is shown via D-Bus
}

func (l *linuxMenu) destroy() {
	// Cleanup handled by linuxSystemTray
}

// linuxMenuItem implements menuItemImpl for Linux
type linuxMenuItem struct {
	parent *MenuItem
}

func (l *linuxMenuItem) setLabel(label string) {
	// Updates are propagated via D-Bus
}

func (l *linuxMenuItem) setTooltip(tooltip string) {
	// Not supported in DBusMenu
}

func (l *linuxMenuItem) setDisabled(disabled bool) {
	// Updates are propagated via D-Bus
}

func (l *linuxMenuItem) setChecked(checked bool) {
	// Updates are propagated via D-Bus
}

func (l *linuxMenuItem) setHidden(hidden bool) {
	// Updates are propagated via D-Bus
}

func (l *linuxMenuItem) destroy() {
	// Cleanup handled by linuxSystemTray
}
