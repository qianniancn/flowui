//go:build linux

package locale

import "github.com/qianniancn/FlowUI/internal/sys/linux"

func systemLocaleName() string {
	return linux.SystemLocaleName()
}
