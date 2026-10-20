package locale

import (
	"os"
	"strings"

	giolocale "gioui.org/x/pref/locale"
	"github.com/qianniancn/flowui/i18n"
	"golang.org/x/text/language"
)

// Language identifies the language used by localized FlowUI widgets.
type Language = i18n.Language

const (
	LanguageAuto    = i18n.LanguageAuto
	LanguageEnglish = i18n.LanguageEnglish
	LanguageChinese = i18n.LanguageChinese
)

// Resolve normalizes an explicit BCP-47 tag and detects the host language for
// LanguageAuto. Invalid explicit tags fall back to English.
func Resolve(value Language) Language {
	if value == LanguageAuto {
		return Detect()
	}
	if normalized, ok := normalizeTag(string(value)); ok {
		return normalized
	}
	return LanguageEnglish
}

// IsChinese reports whether value belongs to the Chinese language family.
// Locale-sensitive widgets use the language family for their built-in domain
// defaults while preserving the complete BCP-47 tag for Catalog lookup.
func IsChinese(value Language) bool {
	normalized, ok := normalizeTag(string(value))
	if !ok {
		return false
	}
	base, _, _ := strings.Cut(string(normalized), "-")
	return base == string(LanguageChinese)
}

var localeEnvKeys = [...]string{"FLOWUI_LANG", "LC_ALL", "LC_MESSAGES", "LANGUAGE", "LANG"}

// Detect returns the normalized host language or English as a fallback. Any
// valid BCP-47 language is retained so an application Catalog can serve it.
func Detect() Language {
	// Check environment variables first
	for _, key := range localeEnvKeys {
		if value := os.Getenv(key); value != "" {
			if language, ok := normalizeTag(value); ok {
				return canonicalBuiltIn(language)
			}
		}
	}
	// Fall back to Gio's platform locale adapter.
	if tag, err := giolocale.Language(); err == nil {
		if language, ok := normalizeTag(tag.String()); ok {
			return canonicalBuiltIn(language)
		}
	}
	return LanguageEnglish
}

func canonicalBuiltIn(value Language) Language {
	base, _, _ := strings.Cut(string(value), "-")
	switch base {
	case "en":
		return LanguageEnglish
	case "zh":
		return LanguageChinese
	default:
		return value
	}
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
	parsed, err := language.Parse(tag)
	if err != nil || parsed == language.Und {
		return LanguageAuto
	}
	base, _ := parsed.Base()
	switch base.String() {
	case "en":
		return LanguageEnglish
	case "zh":
		return LanguageChinese
	default:
		return LanguageAuto
	}
}

func normalizeTag(tag string) (Language, bool) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" || tag == "c" || tag == "posix" {
		return LanguageAuto, false
	}
	if before, _, ok := strings.Cut(tag, ":"); ok {
		tag = before
	}
	if before, _, ok := strings.Cut(tag, "."); ok {
		tag = before
	}
	tag = strings.ReplaceAll(tag, "_", "-")
	parsed, err := language.Parse(tag)
	if err != nil || parsed == language.Und {
		return LanguageAuto, false
	}
	return Language(parsed.String()), true
}
