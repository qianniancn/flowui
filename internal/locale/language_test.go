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

func TestDetectHonorsFlowUIOverride(t *testing.T) {
	t.Setenv("FLOWUI_LANG", "zh-CN")
	t.Setenv("LC_ALL", "en-US")
	if got := Detect(); got != LanguageChinese {
		t.Fatalf("Detect() = %q, want %q", got, LanguageChinese)
	}
}
