@echo off
setlocal enabledelayedexpansion

REM Build script for athens-prefill Windows amd64 executable
REM This script builds the Windows binary

echo Building athens-prefill for Windows amd64...

REM Set environment variables for Windows build
set GOOS=windows
set GOARCH=amd64

REM Build the executable with optimizations
go build -v ^
  -ldflags="-s -w" ^
  -o athens-prefill.exe ^
  ./cmd/athens-prefill

if errorlevel 1 (
  echo Build failed!
  exit /b 1
)

echo Build successful! Created: athens-prefill.exe
echo.
echo File info:
dir athens-prefill.exe

endlocal
