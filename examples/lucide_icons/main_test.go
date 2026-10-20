package main

import (
	"strings"
	"testing"
)

func TestCatalogContainsAllGeneratedIcons(t *testing.T) {
	if got := len(catalog); got != 1899 {
		t.Fatalf("catalog size = %d, want 1899", got)
	}
	for index := 1; index < len(catalog); index++ {
		if strings.ToLower(catalog[index-1].name) > strings.ToLower(catalog[index].name) {
			t.Fatalf("catalog is not sorted at %q", catalog[index].name)
		}
	}
}

func TestFilterIconsMatchesCompactNames(t *testing.T) {
	icons := filterIcons(Model{query: "arrowleft"})
	if len(icons) == 0 || icons[0].name != "ArrowLeft" {
		t.Fatalf("first compact match = %v, want ArrowLeft", iconNames(icons))
	}
}

func TestFilterIconsAcceptsSeparatedTerms(t *testing.T) {
	icons := filterIcons(Model{query: "arrow left"})
	if len(icons) == 0 || icons[0].name != "ArrowLeft" {
		t.Fatalf("first term match = %v, want ArrowLeft", iconNames(icons))
	}
}

func TestFilterIconsByLetter(t *testing.T) {
	icons := filterIcons(Model{letter: "Z"})
	if len(icons) == 0 {
		t.Fatal("letter filter returned no icons")
	}
	for _, item := range icons {
		if item.name[0] != 'Z' {
			t.Fatalf("letter filter returned %q", item.name)
		}
	}
}

func TestFilterIconsNoMatch(t *testing.T) {
	if icons := filterIcons(Model{query: "zzzz-not-an-icon"}); len(icons) != 0 {
		t.Fatalf("unexpected matches: %v", iconNames(icons))
	}
}

func TestFormatCount(t *testing.T) {
	for value, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 1899: "1,899"} {
		if got := formatCount(value); got != want {
			t.Fatalf("formatCount(%d) = %q, want %q", value, got, want)
		}
	}
}

func TestSearchClearsLetterAndSubmitSelectsFirstMatch(t *testing.T) {
	model := Model{letter: "Z", selected: "Zap"}
	Update(&model, SearchChanged{Value: "arrowleft"})
	if model.letter != "" || model.query != "arrowleft" {
		t.Fatalf("search state = %#v", model)
	}
	Update(&model, SearchSubmitted{Value: "arrowleft"})
	if model.selected != "ArrowLeft" {
		t.Fatalf("submitted selection = %q, want ArrowLeft", model.selected)
	}
}

func TestClearFiltersResetsSearchState(t *testing.T) {
	model := Model{query: "arrow", letter: "A", selected: "ArrowLeft"}
	Update(&model, ClearFilters{})
	if model.query != "" || model.letter != "" {
		t.Fatalf("cleared state = %#v", model)
	}
}

func iconNames(icons []catalogIcon) []string {
	names := make([]string, 0, len(icons))
	for _, item := range icons {
		names = append(names, item.name)
	}
	return names
}
