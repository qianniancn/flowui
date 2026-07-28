//go:build windows

package windows

import "golang.org/x/sys/windows"

func SystemLocaleName() string {
	languages, err := windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	if err != nil || len(languages) == 0 {
		return ""
	}
	return languages[0]
}
