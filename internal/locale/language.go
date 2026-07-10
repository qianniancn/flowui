package locale

import (
	"strings"

	"github.com/qianniancn/FlowUI/internal/platform"
)

// Language identifies the language used by localized FlowUI widgets.
type Language string

const (
	LanguageAuto    Language = ""
	LanguageEnglish Language = "en"
	LanguageChinese Language = "zh"
)

// Resolve normalizes language and detects the host language for LanguageAuto.
func Resolve(language Language) Language {
	language = FromTag(string(language))
	if language != LanguageAuto {
		return language
	}
	return Detect()
}

// Detect returns the supported host language or English as a fallback.
func Detect() Language {
	language := FromTag(platform.SystemLanguageTag())
	if language != LanguageAuto {
		return language
	}
	return LanguageEnglish
}

// FromTag maps a platform or BCP-47-style tag to a supported language.
func FromTag(tag string) Language {
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
