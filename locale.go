package flowui

import (
	"strings"

	"github.com/qianniancn/FlowUI/platform"
)

type Language string

const (
	LanguageAuto    Language = ""
	LanguageEnglish Language = "en"
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
