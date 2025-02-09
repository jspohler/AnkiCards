#!/bin/bash

# Set the project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for backend to be ready
wait_for_backend() {
    local retries=0
    local max_retries=30
    local backend_url="http://localhost:${BACKEND_PORT}/api/health"
    
    echo "Waiting for backend to be ready..."
    while [ $retries -lt $max_retries ]; do
        if curl -s "$backend_url" >/dev/null; then
            echo "Backend is ready!"
            return 0
        fi
        retries=$((retries + 1))
        sleep 1
    done
    echo "Backend failed to start within timeout"
    return 1
}

# Check and install dependencies
echo "Checking dependencies..."

# Check for tesseract-ocr
if ! command_exists tesseract; then
    echo "Installing tesseract-ocr..."
    sudo apt-get update && sudo apt-get install -y tesseract-ocr
fi

# Check for poppler-utils (pdftoppm)
if ! command_exists pdftoppm; then
    echo "Installing poppler-utils..."
    sudo apt-get update && sudo apt-get install -y poppler-utils
fi

# Check for Python3 and pip
if ! command_exists python3; then
    echo "Installing python3..."
    sudo apt-get update && sudo apt-get install -y python3 python3-pip
fi

# Check for python3-venv
if ! dpkg -l | grep -q "python3.*-venv"; then
    echo "Installing python3-venv..."
    sudo apt-get update && sudo apt-get install -y python3-venv
fi

# Check for pip3
if ! command_exists pip3; then
    echo "Installing python3-pip..."
    sudo apt-get update && sudo apt-get install -y python3-pip
fi

# Create and activate Python virtual environment
echo "Setting up Python virtual environment..."
python3 -m venv "$PROJECT_ROOT/venv" || {
    echo "Failed to create virtual environment. Please check Python installation."
    exit 1
}

# Activate virtual environment with full path
source "$PROJECT_ROOT/venv/bin/activate" || {
    echo "Failed to activate virtual environment"
    exit 1
}

# Install Python dependencies in the virtual environment
echo "Installing Python dependencies..."
"$PROJECT_ROOT/venv/bin/pip" install --upgrade pip || {
    echo "Failed to upgrade pip"
    exit 1
}
"$PROJECT_ROOT/venv/bin/pip" install genanki==0.13.0 || {
    echo "Failed to install genanki"
    exit 1
}
"$PROJECT_ROOT/venv/bin/pip" install -r "$PROJECT_ROOT/requirements.txt" || {
    echo "Failed to install requirements"
    exit 1
}

# Verify genanki installation
if ! "$PROJECT_ROOT/venv/bin/python3" -c "import genanki" 2>/dev/null; then
    echo "Failed to install genanki in virtual environment"
    exit 1
fi

# Check and install pdfcpu
if ! command_exists pdfcpu; then
    echo "Installing pdfcpu..."
    export PATH=$PATH:$(go env GOPATH)/bin
    go install github.com/pdfcpu/pdfcpu/cmd/pdfcpu@latest
fi

# Add Go bin to PATH if not already there
export PATH=$PATH:$(go env GOPATH)/bin

# Load environment variables
set -a
source .env
set +a

# Create required directories if they don't exist
mkdir -p ./data/uploads ./data/cards ./data/decks

# Start backend
echo "Starting backend server..."
cd backend

# Ensure environment variables are properly set
export UPLOAD_DIR="../data/uploads"
export CARDS_DIR="../data/cards"
export DECKS_DIR="../data/decks"
export BACKEND_PORT="8081"
export GIN_MODE=debug

# Create required directories with absolute paths
mkdir -p "$PROJECT_ROOT/data/uploads" "$PROJECT_ROOT/data/cards" "$PROJECT_ROOT/data/decks"

# Run backend with output to a log file and keep it running in the background
go run cmd/server/main.go 2>&1 | tee ../backend.log &
BACKEND_PID=$!

# Verify the process is running
if ! ps -p $BACKEND_PID > /dev/null; then
    echo "Error: Backend process failed to start"
    exit 1
fi

# Wait for backend to be ready
if ! wait_for_backend; then
    echo "Error: Backend failed to start. Here are the last few lines of the log:"
    tail -n 10 ../backend.log
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

# Monitor backend process
(
    while true; do
        if ! ps -p $BACKEND_PID > /dev/null; then
            echo "Backend process died unexpectedly. Check backend.log for details."
            exit 1
        fi
        sleep 5
    done
) &
MONITOR_PID=$!

# Start frontend
echo "Starting frontend server..."
cd ../frontend/react-app
npm run dev &
FRONTEND_PID=$!

# Handle shutdown
cleanup() {
    echo "Shutting down servers..."
    kill $BACKEND_PID $FRONTEND_PID
    exit 0
}

trap cleanup SIGINT SIGTERM

# Wait for both processes
wait 