package ui

import "github.com/qianniancn/flowui/internal/locale"

// Language identifies the language used by localized FlowUI widgets.
type Language = locale.Language

const (
	// LanguageAuto detects the language from the host system.
	LanguageAuto = locale.LanguageAuto
	// LanguageEnglish selects English UI strings.
	LanguageEnglish = locale.LanguageEnglish
	// LanguageChinese selects Chinese UI strings.
	LanguageChinese = locale.LanguageChinese
)
