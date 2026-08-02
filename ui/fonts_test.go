package ui

import (
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

func TestParseFontCollection(t *testing.T) {
	faces, err := ParseFontCollection(goregular.TTF)
	if err != nil {
		t.Fatalf("parse regular font: %v", err)
	}
	if len(faces) != 1 {
		t.Fatalf("parsed faces = %d, want 1", len(faces))
	}
}

func TestParseFontCollectionRejectsInvalidData(t *testing.T) {
	if _, err := ParseFontCollection([]byte("not a font")); err == nil {
		t.Fatal("invalid font data parsed successfully")
	}
}
