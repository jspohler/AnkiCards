package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jspohler/AnkiCards/backend/internal/services/anki"
	"github.com/jspohler/AnkiCards/backend/internal/services/ocr"
	"github.com/jspohler/AnkiCards/backend/internal/services/pdf"
)

type Handler struct {
	pdfService  *pdf.Service
	ocrService  *ocr.Service
	ankiService *anki.Service
}

type ProcessRequest struct {
	Files             []string `json:"files"`
	IncludeTopicCards bool     `json:"includeTopicCards"`
	CardsPerTopic     int      `json:"cardsPerTopic"`
}

func NewHandler(pdfService *pdf.Service, ocrService *ocr.Service, ankiService *anki.Service) *Handler {
	return &Handler{
		pdfService:  pdfService,
		ocrService:  ocrService,
		ankiService: ankiService,
	}
}

func (h *Handler) HandlePDFUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	uploadedFiles := make([]string, 0, len(files))
	for _, file := range files {
		if filepath.Ext(file.Filename) != ".pdf" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File %s is not a PDF", file.Filename)})
			return
		}

		// Read the file data
		fileContent, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read %s: %v", file.Filename, err)})
			return
		}
		defer fileContent.Close()

		// Read all bytes
		fileBytes, err := io.ReadAll(fileContent)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read %s: %v", file.Filename, err)})
			return
		}

		// Save the file using the PDF service
		filePath, err := h.pdfService.SaveUploadedFile(fileBytes, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save %s: %v", file.Filename, err)})
			return
		}

		uploadedFiles = append(uploadedFiles, filePath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Successfully uploaded %d files", len(uploadedFiles)),
		"files":   uploadedFiles,
	})
}

func (h *Handler) StartProcessing(c *gin.Context) {
	var req ProcessRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Convert file names to full paths in the uploads directory
	uploadDir := h.pdfService.GetUploadDir()
	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get absolute upload directory path: %v", err)})
		return
	}

	fullPaths := make([]string, len(req.Files))
	for i, filename := range req.Files {
		fullPaths[i] = filepath.Join(absUploadDir, filepath.Base(filename))
		// Verify file exists
		if _, err := os.Stat(fullPaths[i]); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File not found: %s", filename)})
			return
		}
	}

	jobID, err := h.pdfService.StartProcessing(fullPaths, req.IncludeTopicCards)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to start processing: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Processing started",
		"jobId":   jobID,
	})
}

func (h *Handler) GetProcessingStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	status := h.pdfService.GetJobStatus(jobID)

	if status == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *Handler) GetCards(c *gin.Context) {
	deckID := c.Param("id")
	deck, err := h.ankiService.GetDeck(deckID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deck not found"})
		return
	}

	c.JSON(http.StatusOK, deck.Cards)
}

func (h *Handler) UpdateCard(c *gin.Context) {
	deckID := c.Param("deckId")
	cardID := c.Param("cardId")

	var card anki.Card
	if err := c.BindJSON(&card); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card data"})
		return
	}

	if err := h.ankiService.UpdateCard(deckID, cardID, card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update card"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card updated successfully"})
}

func (h *Handler) DeleteCard(c *gin.Context) {
	deckID := c.Param("deckId")
	cardID := c.Param("cardId")

	if err := h.ankiService.DeleteCard(deckID, cardID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete card"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card deleted successfully"})
}

func (h *Handler) GetDecks(c *gin.Context) {
	decks, err := h.ankiService.ListDecks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list decks"})
		return
	}

	c.JSON(http.StatusOK, decks)
}

// UpdateCardCSV updates the CSV file with reviewed cards
func (h *Handler) UpdateCardCSV(c *gin.Context) {
	deckName := c.Param("deckName")
	// Ensure we only use the base name
	deckName = filepath.Base(deckName)

	var cards []struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}

	if err := c.BindJSON(&cards); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create CSV content
	csvContent := "Question,Answer\n"
	for _, card := range cards {
		// Escape quotes and newlines in question and answer
		q := strings.ReplaceAll(card.Question, "\"", "\"\"")
		a := strings.ReplaceAll(card.Answer, "\"", "\"\"")
		csvContent += fmt.Sprintf("\"%s\",\"%s\"\n", q, a)
	}

	// Write to CSV file
	csvPath := filepath.Join(h.ankiService.GetCardsDir(), deckName+".csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update CSV file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cards updated successfully"})
}

// GenerateAnkiDeck generates an .apkg file from the reviewed CSV
func (h *Handler) GenerateAnkiDeck(c *gin.Context) {
	deckName := c.Param("deckName")
	// Ensure we only use the base name
	deckName = filepath.Base(deckName)
	csvPath := filepath.Join(h.ankiService.GetCardsDir(), deckName+".csv")

	// Generate APKG file using genanki-like functionality
	apkgPath, err := h.ankiService.GenerateAPKG(csvPath, deckName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"details": map[string]string{
				"deckName": deckName,
				"csvPath":  csvPath,
				"message":  "Failed to generate Anki deck",
			},
		})
		return
	}

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.apkg", deckName))
	c.Header("Content-Type", "application/octet-stream")
	c.File(apkgPath)
}

// GetCardsFromCSV retrieves cards from a CSV file
func (h *Handler) GetCardsFromCSV(c *gin.Context) {
	deckName := c.Param("deckName")
	// Ensure we only use the base name
	deckName = filepath.Base(deckName)
	csvPath := filepath.Join(h.ankiService.GetCardsDir(), deckName+".csv")

	// Read CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CSV file not found"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read CSV file"})
		return
	}

	// Skip header row and convert to JSON
	var cards []map[string]string
	for i, record := range records {
		if i == 0 { // Skip header row
			continue
		}
		if len(record) >= 2 {
			cards = append(cards, map[string]string{
				"question": record[0],
				"answer":   record[1],
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"cards": cards})
}

// ListCardsFromCSV lists all available CSV files and their card counts
func (h *Handler) ListCardsFromCSV(c *gin.Context) {
	files, err := os.ReadDir(h.ankiService.GetCardsDir())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read cards directory"})
		return
	}

	var decks []map[string]interface{}
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".csv" {
			continue
		}

		// Read CSV file to count cards
		csvPath := filepath.Join(h.ankiService.GetCardsDir(), file.Name())
		file, err := os.Open(csvPath)
		if err != nil {
			continue
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			continue
		}

		// Subtract 1 for header row
		cardCount := len(records) - 1
		if cardCount < 0 {
			cardCount = 0
		}

		// Use only the base name without extension
		deckName := strings.TrimSuffix(file.Name(), ".csv")
		decks = append(decks, map[string]interface{}{
			"name":       deckName,
			"totalCards": cardCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"decks": decks})
}
