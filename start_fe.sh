#!/bin/bash
echo "Starting Frontend Server..."

# Get the directory where the script is located
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR/fe"

# Ensure dependencies are installed
echo "Installing dependencies..."
npm install

# Start the Vite dev server
echo "Starting Vite dev server..."
npm run dev
