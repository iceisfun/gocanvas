package gocanvas

import (
	"os"
	"testing"
)

const testFontPath = "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"

func loadTestFont(t *testing.T) *Font {
	t.Helper()
	if _, err := os.Stat(testFontPath); err != nil {
		t.Skipf("test font not available: %s", testFontPath)
	}
	f, err := LoadFontFile(testFontPath)
	if err != nil {
		t.Fatalf("LoadFontFile: %v", err)
	}
	return f
}

func TestLoadFont(t *testing.T) {
	f := loadTestFont(t)
	if f == nil {
		t.Fatal("LoadFontFile returned nil")
	}
}

func TestNewFace(t *testing.T) {
	f := loadTestFont(t)
	face, err := f.NewFace(16)
	if err != nil {
		t.Fatalf("NewFace: %v", err)
	}
	if face.Size() != 16 {
		t.Errorf("Size() = %v, want 16", face.Size())
	}
	if face.Ascent() <= 0 {
		t.Error("expected positive ascent")
	}
	if face.Descent() <= 0 {
		t.Error("expected positive descent")
	}
}

func TestGlyphCache(t *testing.T) {
	f := loadTestFont(t)
	face, err := f.NewFace(16)
	if err != nil {
		t.Fatal(err)
	}

	g1 := face.glyph('A')
	g2 := face.glyph('A')
	if g1 != g2 {
		t.Error("expected same glyph entry on second call (cache hit)")
	}
}
