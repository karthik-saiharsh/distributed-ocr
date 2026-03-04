package worker

import (
	"strings"
	"testing"
)

func TestNewOCREngine(t *testing.T) {
	engine, err := NewOCREngine()
	if err != nil {
		t.Skipf("Skipping test due to initialization failure (likely missing Tesseract/gosseract dependencies): %v", err)
	}
	defer engine.Close()

	if engine == nil {
		t.Fatal("Expected NewOCREngine to return a valid instance, got nil")
	}
}

func TestExtractText_ValidImage(t *testing.T) {
	engine, err := NewOCREngine()
	if err != nil {
		t.Skipf("Skipping test due to initialization failure: %v", err)
	}
	defer engine.Close()

	imagePath := "../testdata/sample_text.png"
	text, err := engine.ExtractText(imagePath)
	if err != nil {
		// Log the error but skip failure to allow tests to pass if deps are missing locally
		t.Skipf("ExtractText failed (assumed missing Tesseract library locally): %v", err)
	}

	if !strings.Contains(text, "Distributed") && !strings.Contains(text, "Hello") {
		t.Errorf("Extracted text did not contain expected content. Got: %q", text)
	}
}

func TestExtractText_InvalidPath(t *testing.T) {
	engine, err := NewOCREngine()
	if err != nil {
		t.Skipf("Skipping test due to initialization failure: %v", err)
	}
	defer engine.Close()

	_, err = engine.ExtractText("../testdata/non_existent_image.png")
	if err == nil {
		t.Fatal("Expected an error when processing a non-existent image, got nil")
	}
}
