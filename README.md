# AnkiCards Generator

An AI-powered web application that automatically generates Anki flashcards from PDF documents. Supports multiple languages (German, English, French, Spanish, Italian) with automatic language detection.

## Features

- 📚 **PDF Processing**
  - Upload and process multiple PDF files
  - Automatic text extraction with OCR fallback
  - Supports various languages with automatic detection
  - Special handling for German texts and characters

- 🎯 **Card Generation**
  - AI-powered question-answer pair generation
  - Automatic language adaptation
  - Focused content chunking for detailed coverage
  - Optional topic summary cards

- 📱 **Modern Web Interface**
  - Clean, responsive Material-UI design
  - Real-time processing status
  - Card review and editing
  - Deck management

- 📤 **Export Options**
  - CSV export for flexibility
  - Anki package (.apkg) export (coming soon)

## Prerequisites

- Go (1.19 or later)
- Node.js (16 or later)
- Python 3.8+
- Tesseract OCR with language packs
- Poppler Utils
- OpenAI API key

## Quick Start

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/AnkiCards.git
   cd AnkiCards
   ```

2. **Create and configure environment file:**
   ```bash
   cp .env.example .env
   # Edit .env and add your OpenAI API key
   ```

3. **Run the start script:**
   ```bash
   ./start.sh
   ```
   This will:
   - Install all dependencies
   - Set up Python virtual environment
   - Install required language packs
   - Start both backend and frontend servers

4. **Access the application:**
   - Frontend: http://localhost:3001
   - Backend API: http://localhost:8081

## Current Limitations

⚠️ Please note the following current limitations:

- Card generation quality and quantity is being optimized
- Very large PDFs might hit token limits
- Duplicate card detection is not yet implemented
- .apkg export is under development

## Development Setup

### Backend (Go)
```bash
cd backend
go mod download
go run cmd/server/main.go
```

### Frontend (React + TypeScript)
```bash
cd frontend/react-app
npm install
npm run dev
```

### Environment Variables

Required environment variables:
```env
OPENAI_API_KEY=your_api_key_here
BACKEND_PORT=8081
FRONTEND_PORT=3001
UPLOAD_DIR=../data/uploads
CARDS_DIR=../data/cards
DECKS_DIR=../data/decks
```

## Project Structure

```
AnkiCards/
├── backend/              # Go backend server
├── frontend/react-app/   # React frontend
├── data/                # Data storage
└── start.sh            # Setup and run script
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- OpenAI for GPT API
- Tesseract for OCR capabilities
- Material-UI for the frontend framework
- Anki for inspiration and card format 