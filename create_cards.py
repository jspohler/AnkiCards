import genanki
import random
import csv
import os
import argparse

# Create an Anki Model (template for cards)
model_id = random.randrange(1 << 30, 1 << 31)
basic_model = genanki.Model(
    model_id,
    'Basic Model',
    fields=[
        {'name': 'Question'},
        {'name': 'Answer'},
    ],
    templates=[
        {
            'name': 'Card 1',
            'qfmt': '{{Question}}',
            'afmt': '{{FrontSide}}<hr id="answer">{{Answer}}',
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
    Creates an Anki deck from a CSV file.
    
    Args:
        csv_file (str): Path to the CSV file
        deck_name (str): Name of the deck
    """
    # Create a random deck ID
    deck_id = random.randrange(1 << 30, 1 << 31)
    deck = genanki.Deck(deck_id, deck_name)
    
    # Read the CSV file
    with open(csv_file, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            note = genanki.Note(
                model=basic_model,
                fields=[row['Question'], row['Answer']]
            )
            deck.add_note(note)
    
    # Create 'decks' directory if it doesn't exist
    os.makedirs('decks', exist_ok=True)
    
    # Save the deck as .apkg file in the 'decks' folder
    output_path = os.path.join('decks', f'{deck_name}.apkg')
    genanki.Package(deck).write_to_file(output_path)
    print(f"Deck created: {output_path}")

if __name__ == '__main__':
    # Create ArgumentParser
    parser = argparse.ArgumentParser(description='Creates an Anki deck from a CSV file.')
    parser.add_argument('csv_file', help='Path to CSV file with question-answer pairs')
    parser.add_argument('--deck-name', '-n', default='New_Deck',
                      help='Name of the Anki deck (default: New_Deck)')
    
    # Parse arguments
    args = parser.parse_args()
    
    # Create the deck
    create_deck_from_csv(args.csv_file, args.deck_name) 