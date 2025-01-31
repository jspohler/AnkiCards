import os
import csv
from pdf2image import convert_from_path
import pytesseract
import argparse
import subprocess

# Define important directories
LECTURES_DIR = "lectures"
DECKS_DIR = "decks"
os.makedirs(LECTURES_DIR, exist_ok=True)
os.makedirs(DECKS_DIR, exist_ok=True)

def extract_text_from_pdf(pdf_path):
    """
    Extracts text from a PDF using OCR.
    
    Args:
        pdf_path (str): Path to the PDF file
    
    Returns:
        str: Extracted text
    """
    print(f"  - Converting PDF to images...")
    # Convert PDF to images
    images = convert_from_path(pdf_path)
    
    print(f"  - Extracting text from {len(images)} pages...")
    # Extract text from each image
    text = ""
    for i, image in enumerate(images, 1):
        print(f"    Processing page {i}/{len(images)}...")
        text += pytesseract.image_to_string(image) + "\n"
    
    return text

def save_qa_pairs_to_csv(qa_pairs, output_file):
    """
    Saves question-answer pairs to a CSV file.
    
    Args:
        qa_pairs (list): List of tuples (question, answer)
        output_file (str): Path to output CSV file
    """
    with open(output_file, 'w', newline='', encoding='utf-8') as f:
        writer = csv.writer(f, quoting=csv.QUOTE_ALL)
        writer.writerow(['Question', 'Answer'])
        writer.writerows(qa_pairs)

def process_pdf_to_anki(pdf_path, output_csv, deck_name):
    """
    Processes a PDF file and creates Anki cards.
    
    Args:
        pdf_path (str): Path to PDF file
        output_csv (str): Path to output CSV file
        deck_name (str): Name of the Anki deck
    """
    print("\n" + "="*80)
    print(f"Processing PDF: {os.path.basename(pdf_path)}")
    print("="*80)
    
    # Extract text from PDF
    try:
        text = extract_text_from_pdf(pdf_path)
        print("  ✓ Text successfully extracted")
    except Exception as e:
        print(f"  ✗ Error extracting text: {str(e)}")
        return
    
    # Save extracted text temporarily
    temp_file = "extracted_text.txt"
    try:
        with open(temp_file, "w", encoding="utf-8") as f:
            f.write(text)
        print(f"  ✓ Text saved to {temp_file}")
    except Exception as e:
        print(f"  ✗ Error saving text: {str(e)}")
        return
    
    print(f"""
NEXT STEPS:
----------------
1. Open the file '{temp_file}'
2. Copy the content
3. Open a new chat with Cursor AI
4. Enter the following prompt:

Create question-answer pairs for Anki flashcards from the following text.
The questions should cover all important concepts and core topics.
The answers should be precise and understandable.
The number of questions should be based on the content - create as many questions as needed
to cover the important concepts (can be more or less than 20).
Format the output as CSV with two columns (Question, Answer), one pair per line.
Use quotation marks for both columns.

[Paste the copied text here]

5. Copy the generated question-answer pairs
6. Save them in the file '{output_csv}'
""")

    # Wait for user input to continue
    input("\nPress Enter when you have saved the question-answer pairs in the CSV file...")

    # Create the Anki deck
    print("\nCreating Anki deck...")
    try:
        subprocess.run(['python', 'create_cards.py', output_csv, '--deck-name', deck_name], check=True)
        print(f"  ✓ Deck '{deck_name}' successfully created!")
    except subprocess.CalledProcessError as e:
        print(f"  ✗ Error creating deck: {str(e)}")
        return

def process_all_pdfs():
    """Processes all PDFs in the lectures directory."""
    pdfs = [f for f in os.listdir(LECTURES_DIR) if f.endswith('.pdf')]
    if not pdfs:
        print("No PDF files found in 'lectures' directory.")
        print(f"Please place your lecture PDFs in the '{LECTURES_DIR}' directory.")
        return
    
    print(f"\nFound PDFs ({len(pdfs)}):")
    for i, pdf in enumerate(pdfs, 1):
        print(f"{i}. {pdf}")
    
    total_pdfs = len(pdfs)
    for i, pdf_file in enumerate(pdfs, 1):
        print(f"\nProcessing PDF {i}/{total_pdfs}")
        pdf_path = os.path.join(LECTURES_DIR, pdf_file)
        pdf_basename = os.path.splitext(pdf_file)[0]
        output_csv = f"{pdf_basename}_qa.csv"
        deck_name = pdf_basename.replace('_', ' ').title()
        
        process_pdf_to_anki(pdf_path, output_csv, deck_name)
    
    print("\n" + "="*80)
    print("All PDFs have been processed!")
    print("="*80)

def list_available_lectures():
    """Shows all available PDF files in the lectures directory."""
    pdfs = [f for f in os.listdir(LECTURES_DIR) if f.endswith('.pdf')]
    if not pdfs:
        print("No PDF files found in 'lectures' directory.")
        print(f"Please place your lecture PDFs in the '{LECTURES_DIR}' directory.")
        return
    
    print("\nAvailable lectures:")
    for i, pdf in enumerate(pdfs, 1):
        print(f"{i}. {pdf}")

if __name__ == '__main__':
    # Create ArgumentParser
    parser = argparse.ArgumentParser(description='Creates Anki cards from PDF files.')
    parser.add_argument('--all', '-a', action='store_true',
                      help='Process all PDFs in lectures directory')
    parser.add_argument('--pdf-file', 
                      help='Name of PDF file in lectures directory or path to PDF file')
    parser.add_argument('--deck-name', '-n', 
                      help='Name of the Anki deck (default: based on PDF name)')
    parser.add_argument('--output-csv', '-o', 
                      help='Name of output CSV file (default: based on PDF name)')
    parser.add_argument('--list', '-l', action='store_true',
                      help='Show all available PDF files in lectures directory')
    
    # Parse arguments
    args = parser.parse_args()
    
    # Show available PDFs if --list option is used
    if args.list:
        list_available_lectures()
        exit()
    
    # Process all PDFs if --all option is used
    if args.all:
        process_all_pdfs()
        exit()
    
    # Check if a PDF file was specified
    if not args.pdf_file:
        parser.print_help()
        exit(1)
    
    # Determine full PDF path
    if os.path.isfile(args.pdf_file):
        pdf_path = args.pdf_file
    else:
        pdf_path = os.path.join(LECTURES_DIR, args.pdf_file)
        if not os.path.isfile(pdf_path):
            print(f"Error: File '{args.pdf_file}' not found.")
            print("\nAvailable PDFs in lectures directory:")
            list_available_lectures()
            exit(1)
    
    # If no CSV name was specified, create one from the PDF name
    if not args.output_csv:
        pdf_basename = os.path.splitext(os.path.basename(pdf_path))[0]
        args.output_csv = f"{pdf_basename}_qa.csv"
    
    # If no deck name was specified, create one from the PDF name
    if not args.deck_name:
        pdf_basename = os.path.splitext(os.path.basename(pdf_path))[0]
        args.deck_name = pdf_basename.replace('_', ' ').title()
    
    # Process the PDF
    process_pdf_to_anki(pdf_path, args.output_csv, args.deck_name) 