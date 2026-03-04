package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"dist-ocr/worker"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/ocr_test/main.go <image_path>")
		os.Exit(1)
	}
	imagePath := os.Args[1]

	// Dependency Validation: Check if tesseract is installed
	cmd := exec.Command("tesseract", "--version")
	if err := cmd.Run(); err != nil {
		fmt.Println("Tesseract OCR is not installed. Install it using:")
		fmt.Println("sudo apt install tesseract-ocr")
		fmt.Println("sudo apt install libtesseract-dev")
		os.Exit(1)
	}

	// Disable default log flags to match expected simple output, or keep default.
	// We'll leave it as is to allow worker logging to appear.
	log.SetFlags(0)

	engine, err := worker.NewOCREngine()
	if err != nil {
		fmt.Printf("Error: OCR failed to initialize: %v\n", err)
		os.Exit(1)
	}
	defer engine.Close()

	text, err := engine.ExtractText(imagePath)
	if err != nil {
		fmt.Println("Error: OCR failed")
		os.Exit(1)
	}

	fmt.Println("Extracted Text:")
	// Print exactly the extracted text, trimming potential trailing spaces just to be safe.
	fmt.Println(strings.TrimSpace(text))
}
