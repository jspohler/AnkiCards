package ocr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Service handles OCR operations
type Service struct {
	tesseractPath string
}

// NewService creates a new OCR service
func NewService(tesseractPath string) *Service {
	if tesseractPath == "" {
		tesseractPath = "tesseract" // Use system tesseract
	}
	return &Service{tesseractPath: tesseractPath}
}

// ExtractText performs OCR on an image file
func (s *Service) ExtractText(imagePath string) (string, error) {
	// Create a temporary output file
	outputBase := strings.TrimSuffix(imagePath, ".pdf") + "_ocr"

	// Run tesseract with detailed error output
	cmd := exec.Command(s.tesseractPath, imagePath, outputBase)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run tesseract: %w\nOutput: %s", err, string(output))
	}

	// Read the output file
	outputFile := outputBase + ".txt"
	outputText, err := os.ReadFile(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to read OCR output file %s: %w", outputFile, err)
	}

	// Clean up the output file
	os.Remove(outputFile)

	if len(strings.TrimSpace(string(outputText))) == 0 {
		return "", fmt.Errorf("OCR produced no text output")
	}

	return string(outputText), nil
}

// CheckTesseract verifies that tesseract is installed and working
func (s *Service) CheckTesseract() error {
	cmd := exec.Command(s.tesseractPath, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tesseract not found or not working: %w", err)
	}
	return nil
}
