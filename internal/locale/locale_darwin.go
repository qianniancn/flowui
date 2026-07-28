//go:build darwin

package locale

import "github.com/qianniancn/flowui/internal/sys/darwin"

func systemLocaleName() string {
	return darwin.SystemLocaleName()
}
