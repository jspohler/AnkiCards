# AnkiCards Generator

A Python tool for automatically generating Anki flashcards from PDF documents.

## Features

- Text extraction from PDF files using OCR
- Support for multiple languages
- Generation of question-answer pairs
- Automatic creation of Anki decks
- Batch processing of multiple PDFs
- User-friendly command-line interface

## Installation

1. Clone the repository:
```bash
git clone https://github.com/jspohler/AnkiCards.git
cd AnkiCards
```

2. Create and activate a virtual environment:
```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

3. Install dependencies:
```bash
pip install -r requirements.txt
```

4. Install Tesseract OCR:
- On macOS: `brew install tesseract`
- On Ubuntu: `sudo apt-get install tesseract-ocr`
- On Windows: [Tesseract Download](https://github.com/UB-Mannheim/tesseract/wiki)

## Usage

### Process a single PDF:
```bash
python generate_questions.py --pdf-file "lecture.pdf"
```

### Process all PDFs in the lectures directory:
```bash
python generate_questions.py --all
```

### List available PDFs:
```bash
python generate_questions.py --list
```

### Create Anki deck from CSV:
```bash
python create_cards.py "questions_answers.csv" --deck-name "My Deck"
```

## Project Structure

```
AnkiCards/
├── README.md
├── requirements.txt
├── generate_questions.py
├── create_cards.py
├── lectures/          # Directory for PDF files
├── decks/            # Directory for generated Anki decks
└── venv/             # Virtual Python environment
```

## License

MIT License 