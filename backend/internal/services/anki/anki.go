package anki

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Card represents an Anki flashcard
type Card struct {
	ID       string    `json:"id"`
	Question string    `json:"question"`
	Answer   string    `json:"answer"`
	Created  time.Time `json:"created"`
}

// Deck represents an Anki deck
type Deck struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Cards   []Card    `json:"cards"`
	Created time.Time `json:"created"`
}

// Service handles Anki card and deck operations
type Service struct {
	decksDir string
	cardsDir string
}

// NewService creates a new Anki service
func NewService(decksDir string, cardsDir string) (*Service, error) {
	if err := os.MkdirAll(decksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create decks directory: %w", err)
	}
	if err := os.MkdirAll(cardsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cards directory: %w", err)
	}
	return &Service{
		decksDir: decksDir,
		cardsDir: cardsDir,
	}, nil
}

// CreateDeck creates a new Anki deck
func (s *Service) CreateDeck(name string, cards []Card) (*Deck, error) {
	deck := &Deck{
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:    name,
		Cards:   cards,
		Created: time.Now(),
	}

	if err := s.saveDeck(deck); err != nil {
		return nil, err
	}

	return deck, nil
}

// GetDeck retrieves a deck by ID
func (s *Service) GetDeck(id string) (*Deck, error) {
	filename := filepath.Join(s.decksDir, id+".json")
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("deck not found: %w", err)
	}
	defer file.Close()

	var deck Deck
	if err := json.NewDecoder(file).Decode(&deck); err != nil {
		return nil, fmt.Errorf("failed to read deck: %w", err)
	}

	return &deck, nil
}

// ListDecks returns all available decks
func (s *Service) ListDecks() ([]Deck, error) {
	files, err := os.ReadDir(s.decksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read decks directory: %w", err)
	}

	var decks []Deck
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		deck, err := s.GetDeck(strings.TrimSuffix(file.Name(), ".json"))
		if err != nil {
			continue
		}
		decks = append(decks, *deck)
	}

	return decks, nil
}

// UpdateCard updates a card in a deck
func (s *Service) UpdateCard(deckID string, cardID string, updatedCard Card) error {
	deck, err := s.GetDeck(deckID)
	if err != nil {
		return err
	}

	found := false
	for i, card := range deck.Cards {
		if card.ID == cardID {
			updatedCard.ID = cardID            // Preserve the original ID
			updatedCard.Created = card.Created // Preserve creation time
			deck.Cards[i] = updatedCard
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("card not found")
	}

	return s.saveDeck(deck)
}

// DeleteCard removes a card from a deck
func (s *Service) DeleteCard(deckID string, cardID string) error {
	deck, err := s.GetDeck(deckID)
	if err != nil {
		return err
	}

	for i, card := range deck.Cards {
		if card.ID == cardID {
			// Remove the card by replacing it with the last element and truncating
			deck.Cards[i] = deck.Cards[len(deck.Cards)-1]
			deck.Cards = deck.Cards[:len(deck.Cards)-1]
			return s.saveDeck(deck)
		}
	}

	return fmt.Errorf("card not found")
}

// saveDeck saves a deck to disk
func (s *Service) saveDeck(deck *Deck) error {
	filename := filepath.Join(s.decksDir, deck.ID+".json")
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create deck file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(deck); err != nil {
		return fmt.Errorf("failed to save deck: %w", err)
	}

	return nil
}

// ExportDeck exports a deck to the Anki .apkg format
func (s *Service) ExportDeck(id string) (string, error) {
	// TODO: Implement .apkg export
	// We'll need to use a library or implement the Anki package format
	return "", fmt.Errorf("not implemented yet")
}

// GetCardsDir returns the cards directory path
func (s *Service) GetCardsDir() string {
	return s.cardsDir
}

// GenerateAPKG generates an Anki package file from a CSV file
func (s *Service) GenerateAPKG(csvPath string, deckName string) (string, error) {
	// Create a temporary directory for the APKG files
	tmpDir := filepath.Join(s.decksDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate the APKG file path
	apkgPath := filepath.Join(tmpDir, deckName+".apkg")

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Find the project root by looking for the venv directory
	projectRoot := ""
	currentDir := cwd
	for i := 0; i < 5; i++ { // Limit the search to 5 levels up
		if _, err := os.Stat(filepath.Join(currentDir, "venv")); err == nil {
			projectRoot = currentDir
			break
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}

	if projectRoot == "" {
		// Try to find project root relative to the executable path
		if execPath, err := os.Executable(); err == nil {
			execDir := filepath.Dir(execPath)
			for i := 0; i < 5; i++ {
				if _, err := os.Stat(filepath.Join(execDir, "venv")); err == nil {
					projectRoot = execDir
					break
				}
				parentDir := filepath.Dir(execDir)
				if parentDir == execDir {
					break
				}
				execDir = parentDir
			}
		}
	}

	if projectRoot == "" {
		return "", fmt.Errorf("could not find project root directory containing venv (searched from %s)", cwd)
	}

	// Get the path to the virtual environment's Python interpreter
	venvPython := filepath.Join(projectRoot, "venv", "bin", "python3")
	if _, err := os.Stat(venvPython); err != nil {
		return "", fmt.Errorf("virtual environment Python not found at %s (project root: %s): %w", venvPython, projectRoot, err)
	}

	// Get the path of the Python script
	scriptPath := filepath.Join(projectRoot, "backend", "internal", "services", "anki", "generate_deck.py")
	if _, err := os.Stat(scriptPath); err != nil {
		return "", fmt.Errorf("Python script not found at %s (project root: %s): %w", scriptPath, projectRoot, err)
	}

	// Check if the CSV file exists
	if _, err := os.Stat(csvPath); err != nil {
		return "", fmt.Errorf("CSV file not found at %s: %w", csvPath, err)
	}

	// Make the script executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make script executable at %s: %w", scriptPath, err)
	}

	// Run the Python script with error logging using the virtual environment's Python
	cmd := exec.Command(venvPython, scriptPath, csvPath, apkgPath)
	cmd.Dir = projectRoot // Set working directory to project root
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to generate Anki deck: %w\nCommand output: %s\nWorking dir: %s\nScript path: %s\nCSV path: %s\nAPKG path: %s\nPython path: %s",
			err, string(output), projectRoot, scriptPath, csvPath, apkgPath, venvPython)
	}

	// Verify the file was created
	if _, err := os.Stat(apkgPath); err != nil {
		return "", fmt.Errorf("APKG file was not created at %s: %w\nScript output: %s", apkgPath, err, string(output))
	}

	return apkgPath, nil
}
