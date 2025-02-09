package pdf

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

// Card represents a flashcard
type Card struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type ProcessingStatus struct {
	Status     string  `json:"status"` // "pending", "processing", "completed", "failed"
	Progress   float64 `json:"progress"`
	Error      string  `json:"error,omitempty"`
	TotalCards int     `json:"totalCards"`
	DeckName   string  `json:"deckName,omitempty"`
	Filename   string  `json:"filename,omitempty"`
}

// Service handles PDF-related operations
type Service struct {
	uploadDir     string
	cardsDir      string
	openAIClient  *openai.Client
	activeJobs    map[string]*ProcessingStatus
	jobsMutex     sync.RWMutex
	cardsPerTopic int
}

// NewService creates a new PDF service
func NewService(uploadDir, cardsDir string, openAIKey string, cardsPerTopic int) (*Service, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	if err := os.MkdirAll(cardsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cards directory: %w", err)
	}

	return &Service{
		uploadDir:     uploadDir,
		cardsDir:      cardsDir,
		openAIClient:  openai.NewClient(openAIKey),
		activeJobs:    make(map[string]*ProcessingStatus),
		cardsPerTopic: cardsPerTopic,
	}, nil
}

// GetUploadDir returns the upload directory path
func (s *Service) GetUploadDir() string {
	return s.uploadDir
}

// SaveUploadedFile saves an uploaded PDF file to disk
func (s *Service) SaveUploadedFile(fileData []byte, filename string) (string, error) {
	// Generate a unique filename with absolute path
	absUploadDir, err := filepath.Abs(s.uploadDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	filename = filepath.Join(absUploadDir, filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Save the file
	if err := os.WriteFile(filename, fileData, 0644); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filename, nil
}

// ExtractText extracts text from a PDF file
func (s *Service) ExtractText(filePath string) (string, error) {
	// Ensure we have absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get the project root directory
	projectRoot := filepath.Join(filepath.Dir(absPath), "..", "..")

	// Create a Python script for text extraction
	extractScript := `
import sys
from pdf2image import convert_from_path
import pytesseract

def extract_text_from_pdf(pdf_path):
    # Convert PDF to images
    images = convert_from_path(pdf_path)
    
    # Extract text from each image
    text = ""
    for image in images:
        text += pytesseract.image_to_string(image) + "\n"
    
    return text

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print("Usage: script.py <pdf_path>")
        sys.exit(1)
    
    pdf_path = sys.argv[1]
    try:
        text = extract_text_from_pdf(pdf_path)
        print(text)
    except Exception as e:
        print(f"Error: {str(e)}", file=sys.stderr)
        sys.exit(1)
`
	// Save the script
	scriptPath := filepath.Join(projectRoot, "backend", "internal", "services", "pdf", "extract_text.py")
	if err := os.WriteFile(scriptPath, []byte(extractScript), 0755); err != nil {
		return "", fmt.Errorf("failed to create extraction script: %w", err)
	}

	// Get the path to the virtual environment's Python interpreter
	venvPython := filepath.Join(projectRoot, "venv", "bin", "python3")
	if _, err := os.Stat(venvPython); err != nil {
		return "", fmt.Errorf("virtual environment Python not found at %s: %w", venvPython, err)
	}

	// Run the Python script
	cmd := exec.Command(venvPython, scriptPath, absPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to extract text: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

func (s *Service) StartProcessing(filePaths []string, includeTopicCards bool) (string, error) {
	jobID := fmt.Sprintf("job_%d", time.Now().UnixNano())

	status := &ProcessingStatus{
		Status:   "pending",
		Progress: 0,
	}

	s.jobsMutex.Lock()
	s.activeJobs[jobID] = status
	s.jobsMutex.Unlock()

	go s.processFilesInBackground(jobID, filePaths, includeTopicCards)

	return jobID, nil
}

func (s *Service) GetJobStatus(jobID string) *ProcessingStatus {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()

	if status, exists := s.activeJobs[jobID]; exists {
		return status
	}
	return nil
}

func (s *Service) processFilesInBackground(jobID string, filePaths []string, includeTopicCards bool) {
	go func() {
		s.jobsMutex.Lock()
		s.activeJobs[jobID] = &ProcessingStatus{
			Status:   "processing",
			Progress: 0,
		}
		s.jobsMutex.Unlock()

		var totalCards int
		var lastError error

		// Use the first file's name as the deck name
		filename := filepath.Base(filePaths[0])
		deckName := strings.TrimSuffix(filename, filepath.Ext(filename))

		for i, filePath := range filePaths {
			// Extract text from PDF
			text, err := s.ExtractText(filePath)
			if err != nil {
				lastError = err
				continue
			}

			// Generate cards
			cards, err := s.generateCards(text, includeTopicCards)
			if err != nil {
				lastError = err
				continue
			}

			// Save cards to CSV
			if err := s.saveCardsToCSV(cards, filePath); err != nil {
				lastError = err
				continue
			}

			totalCards += len(cards)

			// Update progress
			s.jobsMutex.Lock()
			s.activeJobs[jobID].Progress = float64(i+1) / float64(len(filePaths)) * 100
			s.activeJobs[jobID].TotalCards = totalCards
			s.jobsMutex.Unlock()
		}

		// Update final status
		s.jobsMutex.Lock()
		if lastError != nil {
			s.activeJobs[jobID].Status = "failed"
			s.activeJobs[jobID].Error = lastError.Error()
		} else {
			s.activeJobs[jobID].Status = "completed"
			s.activeJobs[jobID].Progress = 100
			s.activeJobs[jobID].DeckName = deckName
			s.activeJobs[jobID].Filename = filename
		}
		s.jobsMutex.Unlock()
	}()
}

func (s *Service) generateCards(text string, includeTopicCards bool) ([]Card, error) {
	// Save original text for debugging
	debugOrigPath := filepath.Join(s.cardsDir, "original_text.txt")
	if debugFile, err := os.Create(debugOrigPath); err == nil {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "=== Original Text ===\n%s\n", text)
	}

	// Preprocess the text
	text = preprocessText(text)

	// Save preprocessed text for review
	debugFilePath := filepath.Join(s.cardsDir, "preprocessed_chunks.txt")
	debugFile, err := os.Create(debugFilePath)
	if err != nil {
		log.Printf("Warning: Failed to create debug file: %v", err)
	} else {
		defer debugFile.Close()
		fmt.Fprintf(debugFile, "=== Original Text Length: %d ===\n\n", len(text))
		fmt.Fprintf(debugFile, "=== Full Preprocessed Text ===\n%s\n\n", text)
	}

	// Split text into chunks of roughly 1000 words each
	words := strings.Fields(text)
	var chunks []string
	chunkSize := 1000 // Approximately 1500 tokens

	for i := 0; i < len(words); i += chunkSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		chunks = append(chunks, chunk)

		if debugFile != nil {
			fmt.Fprintf(debugFile, "=== Chunk %d/%d (Length: %d) ===\n", (i/chunkSize)+1, (len(words)+chunkSize-1)/chunkSize, len(chunk))
			fmt.Fprintf(debugFile, "%s\n\n", chunk)
		}
	}

	var allCards []Card
	cardsPerChunk := s.cardsPerTopic / len(chunks)
	if cardsPerChunk < 1 {
		cardsPerChunk = 1
	}

	for i, chunk := range chunks {
		prompt := fmt.Sprintf(`Create %d high-quality Anki flashcards from this academic text about optimization. 

Requirements for the flashcards:
1. Focus ONLY on the actual content and concepts from the text
2. Each card should teach a specific concept, definition, or relationship
3. Questions should promote understanding and critical thinking
5. Use clear, academic language appropriate for the subject matter
6. Include relevant examples or applications when available
7. Ensure each card is unique and not redundant
8. Questions should be specific and unambiguous
9. Answers should be comprehensive yet concise
10. Cards should build upon each other for progressive learning

This is part %d of %d from the document.

Format each card exactly as:
Q: [Question]
A: [Answer]

Text to process:
%s`, cardsPerChunk, i+1, len(chunks), chunk)

		resp, err := s.openAIClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT3Dot5Turbo,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: "You are an expert in optimization and mathematics, creating precise and educational flashcards. Focus only on the academic content provided, not on meta-information or technical artifacts.",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: prompt,
					},
				},
				MaxTokens:   2000,
				Temperature: 0.5, // Reduced for more consistent output
			},
		)

		if err != nil {
			return nil, fmt.Errorf("OpenAI API error (chunk %d/%d): %w", i+1, len(chunks), err)
		}

		cards, err := parseCardsFromResponse(resp.Choices[0].Message.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse cards from chunk %d/%d: %w", i+1, len(chunks), err)
		}

		allCards = append(allCards, cards...)

		// Add a small delay between API calls to respect rate limits
		time.Sleep(time.Second)
	}

	// If requested, generate additional topic cards from a summary
	if includeTopicCards && len(allCards) > 0 {
		summaryPrompt := fmt.Sprintf(`Create %d high-level conceptual flashcards that connect and synthesize the main themes and concepts from this document.

Guidelines for creating summary flashcards:
1. Focus on relationships between major concepts
2. Emphasize fundamental principles and their applications
3. Include cards that compare and contrast key ideas
4. Create cards that test understanding of broader implications
5. Avoid surface-level or trivial information
6. Questions should promote critical thinking
7. Answers should provide clear, comprehensive explanations

Format each card exactly as:
Q: [Question]
A: [Answer]

Topics covered in the document:
%s`, s.cardsPerTopic, text[:min(500, len(text))])

		resp, err := s.openAIClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT3Dot5Turbo,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: "You are an expert educator specializing in creating high-level conceptual flashcards that promote deep understanding and connections between ideas.",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: summaryPrompt,
					},
				},
				MaxTokens:   2000,
				Temperature: 0.7,
			},
		)

		if err != nil {
			log.Printf("Warning: Failed to generate topic cards: %v", err)
		} else {
			topicCards, err := parseCardsFromResponse(resp.Choices[0].Message.Content)
			if err == nil {
				allCards = append(allCards, topicCards...)
			}
		}
	}

	return allCards, nil
}

// Helper function to find minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Service) checkAndRefineCards(cards []Card) ([]Card, error) {
	// Convert cards to embeddings and check similarity
	// Request new cards for similar ones
	// This is a placeholder - implement the actual logic
	return cards, nil
}

func (s *Service) saveCardsToCSV(cards []Card, originalFile string) error {
	baseFile := filepath.Base(originalFile)
	csvPath := filepath.Join(s.cardsDir, strings.TrimSuffix(baseFile, ".pdf")+".csv")

	file, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Question", "Answer"}); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write cards
	for _, card := range cards {
		if err := writer.Write([]string{card.Question, card.Answer}); err != nil {
			return fmt.Errorf("failed to write card to CSV: %w", err)
		}
	}

	return nil
}

func parseCardsFromResponse(response string) ([]Card, error) {
	// Split the response into lines
	lines := strings.Split(response, "\n")
	var cards []Card
	var currentCard Card

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line starts with Q: or A:
		if strings.HasPrefix(line, "Q:") {
			// If we have a previous card, add it
			if currentCard.Question != "" && currentCard.Answer != "" {
				cards = append(cards, currentCard)
				currentCard = Card{}
			}
			currentCard.Question = strings.TrimPrefix(line, "Q:")
		} else if strings.HasPrefix(line, "A:") {
			currentCard.Answer = strings.TrimPrefix(line, "A:")
		}
	}

	// Add the last card if it exists
	if currentCard.Question != "" && currentCard.Answer != "" {
		cards = append(cards, currentCard)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("no valid cards found in response")
	}

	return cards, nil
}

// preprocessText cleans and filters the text before sending to OpenAI
func preprocessText(text string) string {
	// Split into lines for processing
	lines := strings.Split(text, "\n")
	var processedLines []string

	// Common patterns to filter out - reduced to only clear technical artifacts
	skipPatterns := []string{
		`^\s*\d+\s*$`,         // Just numbers (like page numbers)
		`^\s*$`,               // Empty lines
		`\[\d+\s+\d+\s+\d+\]`, // Matrix values
		`(?i)TJ|Tf|Tm`,        // PDF operators
		`\d+\.\d+\s+\d+\.\d+\s+\d+\.\d+\s+\d+\.\d+\s+-?\d+\.\d+\s+-?\d+\.\d+\s+cm`, // Transform matrices
	}

	// Compile regex patterns
	patterns := make([]*regexp.Regexp, len(skipPatterns))
	for i, pattern := range skipPatterns {
		patterns[i] = regexp.MustCompile(pattern)
	}

	for _, line := range lines {
		// Skip lines matching any of the patterns
		skip := false
		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Clean the line
		line = strings.TrimSpace(line)

		// Skip if too short after cleaning
		if len(line) < 5 { // Reduced minimum length to keep more content
			continue
		}

		// Remove multiple spaces and normalize whitespace
		line = regexp.MustCompile(`\s+`).ReplaceAllString(line, " ")

		// Only remove clearly non-text characters
		line = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(line, "")

		if line != "" {
			processedLines = append(processedLines, line)
		}
	}

	// Join the lines and normalize the text
	processedText := strings.Join(processedLines, "\n")

	// Remove any remaining technical artifacts or irrelevant content
	processedText = regexp.MustCompile(`(?i)(guidelines?|flashcards?|cards?)\s+for\s+creating`).ReplaceAllString(processedText, "")

	return processedText
}

// GenerateAPKG generates an Anki package file from a CSV file
