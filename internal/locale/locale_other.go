//go:build !darwin && !linux && !windows

package locale

func systemLocaleName() string {
	return ""
}
