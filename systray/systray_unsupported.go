//go:build (!windows && !darwin && !linux) || android || ios

package systray

import "errors"

var errUnsupportedPlatform = errors.New("flowui/systray: platform unsupported")

// unsupportedSystemTray keeps the public package buildable on Gio targets
// without a native tray backend. Run reports the capability error through the
// normal asynchronous tray lifecycle instead of leaving a missing symbol or
// panicking at startup.
type unsupportedSystemTray struct {
	parent *SystemTray
}

func newSystemTrayImpl(parent *SystemTray) systemTrayImpl {
	return unsupportedSystemTray{parent: parent}
}

func (s unsupportedSystemTray) setLabel(string)   {}
func (s unsupportedSystemTray) setTooltip(string) {}
func (s unsupportedSystemTray) setIcon([]byte)    {}
func (s unsupportedSystemTray) setMenu(*Menu)     {}
func (s unsupportedSystemTray) show()             {}
func (s unsupportedSystemTray) hide()             {}
func (s unsupportedSystemTray) destroy()          {}
func (s unsupportedSystemTray) run()              { s.parent.fail(errUnsupportedPlatform) }
