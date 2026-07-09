package flowui

import (
	"strings"

	"github.com/qianniancn/FlowUI/platform"
)

// Language identifies the language used by localized FlowUI widgets.
type Language string

const (
	// LanguageAuto detects the language from the host system.
	LanguageAuto Language = ""
	// LanguageEnglish selects English UI strings.
	LanguageEnglish Language = "en"
	// LanguageChinese selects Chinese UI strings.
	LanguageChinese Language = "zh"
)

func resolvedLanguage(language Language) Language {
	language = languageFromTag(string(language))
	if language != LanguageAuto {
		return language
	}
	return detectSystemLanguage()
}

func detectSystemLanguage() Language {
	language := languageFromTag(systemLanguageTag())
	if language != LanguageAuto {
		return language
	}
	return LanguageEnglish
}

func systemLanguageTag() string {
	return platform.SystemLanguageTag()
}

func languageFromTag(tag string) Language {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" || tag == "c" || tag == "posix" {
		return LanguageAuto
	}
	if before, _, ok := strings.Cut(tag, ":"); ok {
		tag = before
	}
	if before, _, ok := strings.Cut(tag, "."); ok {
		tag = before
	}
	tag = strings.ReplaceAll(tag, "_", "-")
	if tag == "zh" || strings.HasPrefix(tag, "zh-") {
		return LanguageChinese
	}
	if tag == "en" || strings.HasPrefix(tag, "en-") {
		return LanguageEnglish
	}
	return LanguageAuto
}

func datePickerLocaleForLanguage(language Language) DatePickerLocale {
	if resolvedLanguage(language) == LanguageChinese {
		return DatePickerChinese()
	}
	return DatePickerEnglish()
}

func languageForDatePickerLocale(locale DatePickerLocale) Language {
	locale = normalizeDatePickerLocale(locale)
	chinese := DatePickerChinese()
	if locale.Hint == chinese.Hint && locale.Weekdays == chinese.Weekdays {
		return LanguageChinese
	}
	return LanguageEnglish
}
