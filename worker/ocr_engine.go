package worker

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/otiai10/gosseract/v2"
)

// OCREngine wraps the Tesseract OCR client to extract text from images
type OCREngine struct {
	client *gosseract.Client
}

// NewOCREngine initializes and configures a reusable Tesseract OCR client.
// It sets the default language to English and the page segmentation mode to automatic.
func NewOCREngine() (*OCREngine, error) {
	client := gosseract.NewClient()

	if err := client.SetLanguage("eng"); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to set language to eng: %w", err)
	}

	if err := client.SetPageSegMode(gosseract.PSM_AUTO); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to set page segmentation mode: %w", err)
	}

	log.Println("[OCR] Initialized OCR Engine successfully")
	return &OCREngine{
		client: client,
	}, nil
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

	if err := o.client.SetImage(imagePath); err != nil {
		log.Printf("[OCR] Error setting image %s: %v\n", imagePath, err)
		return "", fmt.Errorf("failed to set image: %w", err)
	}

	text, err := o.client.Text()
	if err != nil {
		log.Printf("[OCR] Error processing image %s: %v\n", imagePath, err)
		return "", fmt.Errorf("failed to extract text: %w", err)
	}

	// Normalize and clean up the extracted text
	result := strings.TrimSpace(text)
	result = strings.ReplaceAll(result, "\r\n", "\n")

	log.Println("[OCR] OCR completed successfully")
	return result, nil
}

// Close releases the resources held by the underlying gosseract client.
func (o *OCREngine) Close() {
	if o.client != nil {
		o.client.Close()
		log.Println("[OCR] Closed OCR Engine client")
	}
}
