@echo off
echo Starting Frontend Server...
cd /d "%~dp0fe"

:: Ensure dependencies are installed
echo Installing dependencies...
call npm install

:: Start the Vite dev server
echo Starting Vite dev server...
call npm run dev

pause
