//go:build windows

package locale

import "github.com/qianniancn/flowui/internal/sys/windows"

func systemLocaleName() string {
	return windows.SystemLocaleName()
}
