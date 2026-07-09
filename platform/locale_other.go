//go:build !windows

package platform

func systemLocaleName() string {
	return ""
}
