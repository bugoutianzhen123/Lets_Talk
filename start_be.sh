#!/bin/bash
echo "Starting Backend Server..."

# Get the directory where the script is located
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR/be"

# Ensure dependencies are downloaded
echo "Downloading dependencies..."
go mod tidy

# Start the application
echo "Running main.go..."
go run main.go
