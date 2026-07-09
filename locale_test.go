package flowui

import "testing"

func TestLanguageFromTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want Language
	}{
		{name: "Chinese", tag: "zh-CN", want: LanguageChinese},
		{name: "ChineseWithEncoding", tag: "zh_CN.UTF-8", want: LanguageChinese},
		{name: "English", tag: "en-US", want: LanguageEnglish},
		{name: "LanguageList", tag: "zh-CN:en-US", want: LanguageChinese},
		{name: "Unknown", tag: "fr-FR", want: LanguageAuto},
		{name: "POSIX", tag: "C", want: LanguageAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := languageFromTag(tt.tag); got != tt.want {
				t.Fatalf("language = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDatePickerLocaleForLanguage(t *testing.T) {
	if got := datePickerLocaleForLanguage(LanguageChinese).WeekStart; got != 1 {
		t.Fatalf("Chinese week start = %v, want Monday", got)
	}
	if got := datePickerLocaleForLanguage(LanguageEnglish).WeekStart; got != 0 {
		t.Fatalf("English week start = %v, want Sunday", got)
	}
}
