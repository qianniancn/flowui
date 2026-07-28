//go:build windows

package locale

import "github.com/qianniancn/FlowUI/internal/sys/windows"

func systemLocaleName() string {
	return windows.SystemLocaleName()
}
