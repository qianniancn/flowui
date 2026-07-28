//go:build linux

package locale

import "github.com/qianniancn/flowui/internal/sys/linux"

func systemLocaleName() string {
	return linux.SystemLocaleName()
}
