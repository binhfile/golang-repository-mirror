@echo off
setlocal enabledelayedexpansion

REM Build script for athens-prefill Linux amd64 static executable
REM This script cross-compiles a static Linux binary from Windows

echo Building athens-prefill for Linux amd64...

REM Set environment variables for Linux cross-compilation
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

REM Build the static executable with optimizations
go build -v ^
  -ldflags="-s -w" ^
  -o athens-prefill ^
  ./cmd/athens-prefill

if errorlevel 1 (
  echo Build failed!
  exit /b 1
)

echo Build successful! Created: athens-prefill
echo.
echo File info:
dir athens-prefill

endlocal
