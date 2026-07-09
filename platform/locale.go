package platform

import "os"

var localeEnvKeys = [...]string{"FLOWUI_LANG", "LC_ALL", "LC_MESSAGES", "LANGUAGE", "LANG"}

func SystemLanguageTag() string {
	for _, key := range localeEnvKeys {
		value := os.Getenv(key)
		if value != "" {
			return value
		}
	}
	return systemLocaleName()
}
