package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jspohler/AnkiCards/backend/internal/api/handlers"
	"github.com/jspohler/AnkiCards/backend/internal/services/anki"
	"github.com/jspohler/AnkiCards/backend/internal/services/ocr"
	"github.com/jspohler/AnkiCards/backend/internal/services/pdf"
)

func main() {
	// Load environment variables
	uploadDir := os.Getenv("UPLOAD_DIR")
	cardsDir := os.Getenv("CARDS_DIR")
	decksDir := os.Getenv("DECKS_DIR")
	openAIKey := os.Getenv("OPENAI_API_KEY")
	cardsPerTopic := 5 // TODO: Load from env

	if uploadDir == "" || cardsDir == "" || decksDir == "" {
		log.Fatal("UPLOAD_DIR, CARDS_DIR, and DECKS_DIR environment variables must be set")
	}

	// Convert relative paths to absolute paths
	if !filepath.IsAbs(uploadDir) {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}
		uploadDir = filepath.Join(dir, uploadDir)
		cardsDir = filepath.Join(dir, cardsDir)
		decksDir = filepath.Join(dir, decksDir)
	}

	// Initialize services
	pdfService, err := pdf.NewService(uploadDir, cardsDir, openAIKey, cardsPerTopic)
	if err != nil {
		log.Fatalf("Failed to create PDF service: %v", err)
	}

	ocrService := ocr.NewService("")
	ankiService, err := anki.NewService(decksDir, cardsDir)
	if err != nil {
		log.Fatalf("Failed to create Anki service: %v", err)
	}

	// Initialize handler
	handler := handlers.NewHandler(pdfService, ocrService, ankiService)

	// Initialize Gin
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3001"} // React dev server
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"}
	r.Use(cors.New(config))

	// Health check endpoint
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API Routes
	api := r.Group("/api")
	{
		// PDF Upload and Processing
		api.POST("/upload", handler.HandlePDFUpload)
		api.POST("/process", handler.StartProcessing)
		api.GET("/process/:jobId", handler.GetProcessingStatus)

		// Card Management
		api.GET("/cards/:id", handler.GetCards)
		api.PUT("/decks/:deckId/cards/:cardId", handler.UpdateCard)
		api.DELETE("/decks/:deckId/cards/:cardId", handler.DeleteCard)
		api.GET("/decks", handler.GetDecks)

		// CSV and APKG endpoints
		api.GET("/cards/list", handler.ListCardsFromCSV)
		api.GET("/cards/csv/:deckName", handler.GetCardsFromCSV)
		api.PUT("/cards/csv/:deckName", handler.UpdateCardCSV)
		api.GET("/cards/apkg/:deckName", handler.GenerateAnkiDeck)
	}

	// Start server
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server starting on :%s", port)
	log.Fatal(r.Run(":" + port))
}

func handlePDFUpload(c *gin.Context) {
	// TODO: Implement PDF upload and processing
	c.JSON(200, gin.H{
		"message": "Upload endpoint - To be implemented",
	})
}

func getCards(c *gin.Context) {
	// TODO: Implement card retrieval
	c.JSON(200, gin.H{
		"message": "Get cards endpoint - To be implemented",
	})
}

func saveCards(c *gin.Context) {
	// TODO: Implement card saving
	c.JSON(200, gin.H{
		"message": "Save cards endpoint - To be implemented",
	})
}

func getDecks(c *gin.Context) {
	// TODO: Implement deck listing
	c.JSON(200, gin.H{
		"message": "Get decks endpoint - To be implemented",
	})
}
