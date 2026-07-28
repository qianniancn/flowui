//go:build darwin

package locale

import "github.com/qianniancn/FlowUI/internal/sys/darwin"

func systemLocaleName() string {
	return darwin.SystemLocaleName()
}
