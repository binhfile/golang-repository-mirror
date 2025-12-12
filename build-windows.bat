@echo off
setlocal enabledelayedexpansion

REM Build script for go-mod-clone Windows amd64 executable
REM This script builds the Windows binary

echo Building go-mod-clone for Windows amd64...

REM Set environment variables for Windows build
set GOOS=windows
set GOARCH=amd64

REM Build the executable with optimizations
go build -v ^
  -ldflags="-s -w" ^
  -o go-mod-clone.exe ^
  ./cmd/go-mod-clone

if errorlevel 1 (
  echo Build failed!
  exit /b 1
)

echo Build successful! Created: go-mod-clone.exe
echo.
echo File info:
dir go-mod-clone.exe

endlocal
