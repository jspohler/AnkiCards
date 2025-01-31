# AnkiCards Generator

Ein Python-Tool zur automatischen Generierung von Anki-Karteikarten aus PDF-Dokumenten.

## Features

- Extraktion von Text aus PDF-Dateien mittels OCR
- Unterstützung für deutsche Sprache
- Generierung von Frage-Antwort-Paaren
- Automatische Erstellung von Anki-Decks
- Batch-Verarbeitung mehrerer PDFs
- Benutzerfreundliche Kommandozeilen-Schnittstelle

## Installation

1. Klone das Repository:
```bash
git clone https://github.com/[dein-username]/AnkiCards.git
cd AnkiCards
```

2. Erstelle eine virtuelle Umgebung und aktiviere sie:
```bash
python -m venv venv
source venv/bin/activate  # Unter Windows: venv\Scripts\activate
```

3. Installiere die Abhängigkeiten:
```bash
pip install -r requirements.txt
```

4. Installiere Tesseract OCR:
- Unter macOS: `brew install tesseract`
- Unter Ubuntu: `sudo apt-get install tesseract-ocr`
- Unter Windows: [Tesseract Download](https://github.com/UB-Mannheim/tesseract/wiki)

## Verwendung

### Einzelne PDF verarbeiten:
```bash
python generate_questions.py --pdf-file "vorlesung.pdf"
```

### Alle PDFs im lectures-Verzeichnis verarbeiten:
```bash
python generate_questions.py --all
```

### Verfügbare PDFs anzeigen:
```bash
python generate_questions.py --list
```

### Anki-Deck aus CSV erstellen:
```bash
python create_cards.py "fragen_antworten.csv" --deck-name "Mein Deck"
```

## Projektstruktur

```
AnkiCards/
├── README.md
├── requirements.txt
├── generate_questions.py
├── create_cards.py
├── lectures/          # Verzeichnis für PDF-Dateien
├── decks/            # Verzeichnis für generierte Anki-Decks
└── venv/             # Virtuelle Python-Umgebung
```

## Lizenz

MIT License 