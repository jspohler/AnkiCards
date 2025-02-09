
import sys
from pdf2image import convert_from_path
import pytesseract

def extract_text_from_pdf(pdf_path):
    # Convert PDF to images
    images = convert_from_path(pdf_path)
    
    # Extract text from each image
    text = ""
    for image in images:
        text += pytesseract.image_to_string(image) + "\n"
    
    return text

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print("Usage: script.py <pdf_path>")
        sys.exit(1)
    
    pdf_path = sys.argv[1]
    try:
        text = extract_text_from_pdf(pdf_path)
        print(text)
    except Exception as e:
        print(f"Error: {str(e)}", file=sys.stderr)
        sys.exit(1)
