import os
import csv
from pdf2image import convert_from_path
import pytesseract
import argparse
import subprocess

# Definiere wichtige Verzeichnisse
LECTURES_DIR = "lectures"
DECKS_DIR = "decks"
os.makedirs(LECTURES_DIR, exist_ok=True)
os.makedirs(DECKS_DIR, exist_ok=True)

def extract_text_from_pdf(pdf_path):
    """
    Extrahiert Text aus einem PDF mittels OCR.
    
    Args:
        pdf_path (str): Pfad zur PDF-Datei
    
    Returns:
        str: Extrahierter Text
    """
    print(f"  - Konvertiere PDF zu Bildern...")
    # Konvertiere PDF zu Bildern
    images = convert_from_path(pdf_path)
    
    print(f"  - Extrahiere Text aus {len(images)} Seiten...")
    # Extrahiere Text aus jedem Bild
    text = ""
    for i, image in enumerate(images, 1):
        print(f"    Verarbeite Seite {i}/{len(images)}...")
        text += pytesseract.image_to_string(image, lang='deu') + "\n"
    
    return text

def save_qa_pairs_to_csv(qa_pairs, output_file):
    """
    Speichert Frage-Antwort-Paare in einer CSV-Datei.
    
    Args:
        qa_pairs (list): Liste von Tupeln (Frage, Antwort)
        output_file (str): Pfad zur Ausgabe-CSV-Datei
    """
    with open(output_file, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f, quoting=csv.QUOTE_ALL)
        writer.writerow(['Frage', 'Antwort'])
        writer.writerows(qa_pairs)

def process_pdf_to_anki(pdf_path, output_csv, deck_name):
    """
    Verarbeitet eine PDF-Datei und erstellt Anki-Karten.
    
    Args:
        pdf_path (str): Pfad zur PDF-Datei
        output_csv (str): Pfad zur Ausgabe-CSV-Datei
        deck_name (str): Name des Anki-Decks
    """
    print("\n" + "="*80)
    print(f"Verarbeite PDF: {os.path.basename(pdf_path)}")
    print("="*80)
    
    # Extrahiere Text aus PDF
    try:
        text = extract_text_from_pdf(pdf_path)
        print("  ✓ Text erfolgreich extrahiert")
    except Exception as e:
        print(f"  ✗ Fehler beim Extrahieren des Texts: {str(e)}")
        return
    
    # Speichere den extrahierten Text temporär
    temp_file = "extracted_text.txt"
    try:
        with open(temp_file, "w", encoding="utf-8") as f:
            f.write(text)
        print(f"  ✓ Text in {temp_file} gespeichert")
    except Exception as e:
        print(f"  ✗ Fehler beim Speichern des Texts: {str(e)}")
        return
    
    print(f"""
NÄCHSTE SCHRITTE:
----------------
1. Öffne die Datei '{temp_file}'
2. Kopiere den Inhalt
3. Öffne einen neuen Chat mit der Cursor AI
4. Gib folgenden Prompt ein:

Erstelle Frage-Antwort-Paare für Anki-Karteikarten aus dem folgenden Text.
Die Fragen sollten alle wichtigen Konzepte und Kernthemen abdecken.
Die Antworten sollten präzise und verständlich sein.
Die Anzahl der Fragen sollte sich am Inhalt orientieren - erstelle so viele Fragen wie nötig, 
um die wichtigen Konzepte abzudecken (es können mehr oder weniger als 20 sein).
Formatiere die Ausgabe als CSV mit zwei Spalten (Frage, Antwort), eine Zeile pro Paar.
Verwende Anführungszeichen für beide Spalten.

[Füge hier den kopierten Text ein]

5. Kopiere die generierten Frage-Antwort-Paare
6. Speichere sie in der Datei '{output_csv}'
""")

    # Warte auf Benutzereingabe, um fortzufahren
    input("\nDrücke Enter, wenn du die Frage-Antwort-Paare in der CSV-Datei gespeichert hast...")

    # Erstelle das Anki-Deck
    print("\nErstelle Anki-Deck...")
    try:
        subprocess.run(['python', 'create_cards.py', output_csv, '--deck-name', deck_name], check=True)
        print(f"  ✓ Deck '{deck_name}' wurde erfolgreich erstellt!")
    except subprocess.CalledProcessError as e:
        print(f"  ✗ Fehler beim Erstellen des Decks: {str(e)}")
        return

def process_all_pdfs():
    """Verarbeitet alle PDFs im lectures Verzeichnis."""
    pdfs = [f for f in os.listdir(LECTURES_DIR) if f.endswith('.pdf')]
    if not pdfs:
        print("Keine PDF-Dateien im 'lectures' Verzeichnis gefunden.")
        print(f"Bitte lege deine Vorlesungs-PDFs im Verzeichnis '{LECTURES_DIR}' ab.")
        return
    
    print(f"\nGefundene PDFs ({len(pdfs)}):")
    for i, pdf in enumerate(pdfs, 1):
        print(f"{i}. {pdf}")
    
    total_pdfs = len(pdfs)
    for i, pdf_file in enumerate(pdfs, 1):
        print(f"\nVerarbeite PDF {i}/{total_pdfs}")
        pdf_path = os.path.join(LECTURES_DIR, pdf_file)
        pdf_basename = os.path.splitext(pdf_file)[0]
        output_csv = f"{pdf_basename}_qa.csv"
        deck_name = pdf_basename.replace('_', ' ').title()
        
        process_pdf_to_anki(pdf_path, output_csv, deck_name)
    
    print("\n" + "="*80)
    print("Alle PDFs wurden verarbeitet!")
    print("="*80)

def list_available_lectures():
    """Zeigt alle verfügbaren PDF-Dateien im lectures Verzeichnis an."""
    pdfs = [f for f in os.listdir(LECTURES_DIR) if f.endswith('.pdf')]
    if not pdfs:
        print("Keine PDF-Dateien im 'lectures' Verzeichnis gefunden.")
        print(f"Bitte lege deine Vorlesungs-PDFs im Verzeichnis '{LECTURES_DIR}' ab.")
        return
    
    print("\nVerfügbare Vorlesungen:")
    for i, pdf in enumerate(pdfs, 1):
        print(f"{i}. {pdf}")

if __name__ == '__main__':
    # Erstelle den ArgumentParser
    parser = argparse.ArgumentParser(description='Erstellt Anki-Karten aus PDF-Dateien.')
    parser.add_argument('--all', '-a', action='store_true',
                      help='Verarbeitet alle PDFs im lectures Verzeichnis')
    parser.add_argument('--pdf-file', 
                      help='Name der PDF-Datei im lectures Verzeichnis oder Pfad zur PDF-Datei')
    parser.add_argument('--deck-name', '-n', 
                      help='Name des Anki-Decks (Standard: basierend auf PDF-Namen)')
    parser.add_argument('--output-csv', '-o', 
                      help='Name der Ausgabe-CSV-Datei (Standard: basierend auf PDF-Namen)')
    parser.add_argument('--list', '-l', action='store_true',
                      help='Zeigt alle verfügbaren PDF-Dateien im lectures Verzeichnis an')
    
    # Parse die Argumente
    args = parser.parse_args()
    
    # Zeige verfügbare PDFs an, wenn --list Option verwendet wird
    if args.list:
        list_available_lectures()
        exit()
    
    # Verarbeite alle PDFs wenn --all Option verwendet wird
    if args.all:
        process_all_pdfs()
        exit()
    
    # Prüfe ob eine PDF-Datei angegeben wurde
    if not args.pdf_file:
        parser.print_help()
        exit(1)
    
    # Bestimme den vollständigen PDF-Pfad
    if os.path.isfile(args.pdf_file):
        pdf_path = args.pdf_file
    else:
        pdf_path = os.path.join(LECTURES_DIR, args.pdf_file)
        if not os.path.isfile(pdf_path):
            print(f"Fehler: Die Datei '{args.pdf_file}' wurde nicht gefunden.")
            print("\nVerfügbare PDFs im lectures Verzeichnis:")
            list_available_lectures()
            exit(1)
    
    # Wenn kein CSV-Name angegeben wurde, erstelle einen aus dem PDF-Namen
    if not args.output_csv:
        pdf_basename = os.path.splitext(os.path.basename(pdf_path))[0]
        args.output_csv = f"{pdf_basename}_qa.csv"
    
    # Wenn kein Deck-Name angegeben wurde, erstelle einen aus dem PDF-Namen
    if not args.deck_name:
        pdf_basename = os.path.splitext(os.path.basename(pdf_path))[0]
        args.deck_name = pdf_basename.replace('_', ' ').title()
    
    # Verarbeite die PDF
    process_pdf_to_anki(pdf_path, args.output_csv, args.deck_name) 