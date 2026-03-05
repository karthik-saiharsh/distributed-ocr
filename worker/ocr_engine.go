package worker

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// OCREngine uses the local Tesseract CLI to extract text from images
type OCREngine struct {
	// No persistent client needed for CLI invocation
}

// NewOCREngine initializes the OCR engine.
func NewOCREngine() (*OCREngine, error) {
	// Verify tesseract is installed
	_, err := exec.LookPath("tesseract")
	if err != nil {
		return nil, fmt.Errorf("tesseract CLI not found. Please install it using 'brew install tesseract'")
	}

	log.Println("[OCR] Initialized OCR Engine successfully via CLI")
	return &OCREngine{}, nil
}

// ExtractText loads an image from the provided path, runs OCR, and returns the
// extracted text. It handles errors and normalizes the resulting text.
func (o *OCREngine) ExtractText(imagePath string) (string, error) {
	if imagePath == "" {
		err := errors.New("empty image path provided")
		log.Println("[OCR] Error processing image:", err)
		return "", err
	}

	log.Printf("[OCR] Processing image: %s\n", imagePath)

	// Invoke the tesseract CLI: `tesseract <imagePath> stdout -l eng --psm 3`
	cmd := exec.Command("tesseract", imagePath, "stdout", "-l", "eng", "--psm", "3")

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		log.Printf("[OCR] Error processing image %s: %v. Stderr: %s\n", imagePath, err, errBuf.String())
		return "", fmt.Errorf("failed to extract text: %w", err)
	}

	// Normalize and clean up the extracted text
	result := strings.TrimSpace(outBuf.String())
	result = strings.ReplaceAll(result, "\r\n", "\n")

	log.Println("[OCR] OCR completed successfully")
	return result, nil
}

// Close is a no-op for the CLI engine.
func (o *OCREngine) Close() {
	log.Println("[OCR] Closed OCR Engine client")
}
