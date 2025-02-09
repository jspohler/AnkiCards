#!/usr/bin/env python3
import genanki
import csv
import sys
import os
import random

def generate_deck_id():
    # Generate a random deck ID (needs to be a 32-bit unsigned integer)
    return random.randrange(1 << 30, 1 << 31)

def generate_model_id():
    # Generate a random model ID (needs to be a 32-bit unsigned integer)
    return random.randrange(1 << 30, 1 << 31)

def create_anki_deck(csv_path, output_path):
    # Create a unique model for the cards
    model = genanki.Model(
        generate_model_id(),
        'Simple Model',
        fields=[
            {'name': 'Question'},
            {'name': 'Answer'},
        ],
        templates=[{
            'name': 'Card 1',
            'qfmt': '{{Question}}',
            'afmt': '{{FrontSide}}<hr id="answer">{{Answer}}',
        }]
    )

    # Create the deck
    deck_name = os.path.splitext(os.path.basename(csv_path))[0]
    deck = genanki.Deck(generate_deck_id(), deck_name)

    # Read the CSV file and create cards
    with open(csv_path, 'r', encoding='utf-8') as file:
        reader = csv.DictReader(file)
        for row in reader:
            note = genanki.Note(
                model=model,
                fields=[row['Question'], row['Answer']]
            )
            deck.add_note(note)

    # Create the package
    package = genanki.Package(deck)
    package.write_to_file(output_path)

if __name__ == '__main__':
    if len(sys.argv) != 3:
        print("Usage: generate_deck.py <input_csv_path> <output_apkg_path>")
        sys.exit(1)

    csv_path = sys.argv[1]
    output_path = sys.argv[2]

    try:
        create_anki_deck(csv_path, output_path)
        print(f"Successfully created Anki deck: {output_path}")
    except Exception as e:
        print(f"Error creating Anki deck: {str(e)}", file=sys.stderr)
        sys.exit(1) 