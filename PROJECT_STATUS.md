# AnkiCards Generator - Project Status

## Project Overview
AnkiCards Generator is a web application that automatically generates Anki flashcards from PDF documents using AI (OpenAI's GPT models) for content extraction and question generation.

## Architecture

### Backend (Go)
- Server: Gin web framework
- Port: 8081
- Main components:
  - PDF Service: Handles PDF text extraction and card generation
  - OCR Service: Fallback for PDFs that can't be extracted directly
  - Anki Service: Manages decks and cards

### Frontend (React + TypeScript)
- Framework: Vite
- Port: 3001
- UI: Material-UI (MUI)
- State Management: React Query

## Core Features

### 1. PDF Processing
- Upload multiple PDF files
- Text extraction using:
  - Primary: `pdfcpu` for direct text extraction
  - Fallback: OCR using `pdftoppm` + Tesseract
- Chunked processing for large documents
- Progress tracking via job IDs

### 2. Card Generation
- Uses GPT-3.5-turbo (recently switched from GPT-4 due to token limits)
- Splits text into 1000-word chunks
- Generates Q&A pairs for each chunk
- Optional topic-based summary cards
- Duplicate detection (planned but not fully implemented)

### 3. Card Management
- Organize cards into decks
- View, edit, and delete cards
- CSV export for each deck
- Anki package (.apkg) export (planned but not implemented)

## Current Status

### Working Features
- PDF upload and processing
- Text extraction (both direct and OCR)
- Card generation with chunking
- Basic deck and card management
- Frontend UI for all core functions

### Recent Changes
- Switched to GPT-3.5-turbo for better rate limits
- Implemented text chunking (100 words per chunk)
- Added delay between API calls
- Fixed file path handling issues
- Improved language detection with German character recognition
- Updated card generation prompts for better output
- Added support for multiple languages (German, English, French, Spanish, Italian)

### Known Issues/TODOs
- Card generation quality needs improvement:
  - Number of generated cards is insufficient
  - Card quality and coverage could be better
  - Need to optimize prompts for more comprehensive content coverage
- Token limit management for very large PDFs
- Duplicate card detection needs implementation
- Anki package export not implemented
- Error handling could be improved
- Some TypeScript linting issues in frontend

## Project Structure
```
AnkiCards/
├── backend/
│   ├── cmd/server/           # Main server entry
│   └── internal/
│       ├── api/handlers/     # API endpoints
│       └── services/
│           ├── pdf/          # PDF processing
│           ├── ocr/          # OCR handling
│           └── anki/         # Card management
├── frontend/react-app/
│   └── src/
│       └── components/       # React components
└── data/
    ├── uploads/             # PDF storage
    ├── cards/              # Generated cards (CSV)
    └── decks/              # Deck storage (JSON)
```

## Dependencies

### System Requirements
- Go
- Node.js
- Tesseract OCR
- Poppler Utils
- pdfcpu

### Environment Variables
```env
# Server Configuration
BACKEND_HOST=localhost
BACKEND_PORT=8081
FRONTEND_HOST=localhost
FRONTEND_PORT=3001

# OpenAI Configuration
OPENAI_API_KEY=your_api_key_here

# Card Generation
DEFAULT_CARDS_PER_TOPIC=5
MIN_SIMILARITY_THRESHOLD=0.85  # For duplicate detection

# File Paths
UPLOAD_DIR=../data/uploads
CARDS_DIR=../data/cards
DECKS_DIR=../data/decks

# Development
DEBUG=true
```

## API Endpoints

### Backend API Routes
- `GET /api/health` - Health check endpoint
- `POST /api/upload` - Upload PDF files
- `POST /api/process` - Start processing uploaded files
- `GET /api/process/:jobId` - Get processing status
- `GET /api/cards/:id` - Get cards for a deck
- `PUT /api/decks/:deckId/cards/:cardId` - Update a card
- `DELETE /api/decks/:deckId/cards/:cardId` - Delete a card
- `GET /api/decks` - List all decks

## Frontend Routes
- `/` - File upload page
- `/status/:jobId` - Processing status page
- `/cards` - Card review and management page

## Next Steps
1. Improve card generation quality and quantity
   - Optimize prompts for better coverage
   - Fine-tune chunk size and processing
   - Implement quality metrics
2. Improve error handling for PDF processing
3. Implement duplicate card detection
4. Add Anki package export
5. Add batch processing optimizations
6. Implement proper test coverage

## Development Workflow
The project uses a `start.sh` script to handle:
- Dependency checks
- Environment setup
- Starting both backend and frontend services
- Monitoring service health

All changes should be tested using this script to ensure proper integration.

## Recent Issues
1. Token limit exceeded with GPT-4 for large PDFs
   - Solution: Switched to GPT-3.5-turbo and implemented chunking
2. PDF text extraction failures
   - Solution: Added OCR fallback using Tesseract
3. Frontend stability
   - Solution: Added strict port configuration and improved error handling

## Testing
- Backend: Go tests (needs more coverage)
- Frontend: React Testing Library (to be implemented)
- Integration tests needed
- Manual testing workflow in place

## Performance Considerations
- Chunk size of 100 words for more granular card generation
- 1-second delay between API calls
- Parallel processing of multiple PDFs
- Rate limit awareness for OpenAI API
- Language detection optimization with character recognition

## Security Measures
- CORS configuration in place
- Environment variable management
- File type validation
- API key protection

## Maintenance
- Regular dependency updates
- Error log monitoring
- API rate limit monitoring
- Disk space management for uploads 