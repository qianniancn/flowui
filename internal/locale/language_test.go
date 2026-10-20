package locale

import "testing"

func TestFromTag(t *testing.T) {
	tests := map[string]Language{
		"zh-CN":       LanguageChinese,
		"zh_CN.UTF-8": LanguageChinese,
		"en-US":       LanguageEnglish,
		"zh-CN:en-US": LanguageChinese,
		"fr-FR":       LanguageAuto,
		"C":           LanguageAuto,
	}
	for tag, want := range tests {
		if got := FromTag(tag); got != want {
			t.Errorf("FromTag(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestResolvePreservesExplicitBCP47Language(t *testing.T) {
	if got := Resolve(Language("fr-ca")); got != Language("fr-CA") {
		t.Fatalf("Resolve(fr-ca) = %q, want fr-CA", got)
	}
	if got := Resolve(Language("not a language")); got != LanguageEnglish {
		t.Fatalf("Resolve(invalid) = %q, want %q", got, LanguageEnglish)
	}
}

func TestIsChineseMatchesLanguageFamily(t *testing.T) {
	tests := map[Language]bool{
		LanguageChinese:       true,
		Language("zh-CN"):     true,
		Language("zh-Hant"):   true,
		LanguageEnglish:       false,
		Language("en-US"):     false,
		Language("fr-CA"):     false,
		LanguageAuto:          false,
		Language("not-valid"): false,
	}
	for value, want := range tests {
		if got := IsChinese(value); got != want {
			t.Errorf("IsChinese(%q) = %v, want %v", value, got, want)
		}
	}
}

func TestDetectHonorsFlowUIOverride(t *testing.T) {
	t.Setenv("FLOWUI_LANG", "zh-CN")
	t.Setenv("LC_ALL", "en-US")
	if got := Detect(); got != LanguageChinese {
		t.Fatalf("Detect() = %q, want %q", got, LanguageChinese)
	}
}

func TestDetectPreservesApplicationLanguage(t *testing.T) {
	t.Setenv("FLOWUI_LANG", "fr-ca")
	t.Setenv("LC_ALL", "en-US")
	if got := Detect(); got != Language("fr-CA") {
		t.Fatalf("Detect() = %q, want fr-CA", got)
	}
}
