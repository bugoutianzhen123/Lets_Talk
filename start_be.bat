@echo off
echo Starting Backend Server...
cd /d "%~dp0be"

:: Ensure dependencies are downloaded
echo Downloading dependencies...
go mod tidy

:: Start the application
echo Running main.go...
go run main.go

pause
