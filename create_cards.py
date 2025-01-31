import genanki
import random
import csv
import os
import argparse

# Erstelle ein Anki Model (Template für die Karten)
model_id = random.randrange(1 << 30, 1 << 31)
basic_model = genanki.Model(
    model_id,
    'Basic Model',
    fields=[
        {'name': 'Fragen'},
        {'name': 'Antworten'},
    ],
    templates=[
        {
            'name': 'Card 1',
            'qfmt': '{{Fragen}}',
            'afmt': '{{FrontSide}}<hr id="answer">{{Antworten}}',
        },
    ],
    css='''
        .card {
            font-family: arial;
            font-size: 20px;
            text-align: center;
            color: black;
            background-color: white;
        }
    '''
)

def create_deck_from_csv(csv_file, deck_name):
    """
    Erstellt ein Anki-Deck aus einer CSV-Datei.
    
    Args:
        csv_file (str): Pfad zur CSV-Datei
        deck_name (str): Name des Decks
    """
    # Erstelle eine zufällige Deck-ID
    deck_id = random.randrange(1 << 30, 1 << 31)
    deck = genanki.Deck(deck_id, deck_name)
    
    # Lese die CSV-Datei
    with open(csv_file, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            note = genanki.Note(
                model=basic_model,
                fields=[row['Fragen'], row['Antworten']]
            )
            deck.add_note(note)
    
    # Erstelle den Ordner 'decks', falls er nicht existiert
    os.makedirs('decks', exist_ok=True)
    
    # Speichere das Deck als .apkg Datei im 'decks' Ordner
    output_path = os.path.join('decks', f'{deck_name}.apkg')
    genanki.Package(deck).write_to_file(output_path)
    print(f"Deck wurde erstellt: {output_path}")

if __name__ == '__main__':
    # Erstelle den ArgumentParser
    parser = argparse.ArgumentParser(description='Erstellt ein Anki-Deck aus einer CSV-Datei.')
    parser.add_argument('csv_file', help='Pfad zur CSV-Datei mit den Frage-Antwort-Paaren')
    parser.add_argument('--deck-name', '-n', default='Neues_Deck',
                      help='Name des Anki-Decks (Standard: Neues_Deck)')
    
    # Parse die Argumente
    args = parser.parse_args()
    
    # Erstelle das Deck
    create_deck_from_csv(args.csv_file, args.deck_name) 