@echo off
setlocal enabledelayedexpansion

REM Build script for go-mod-clone Linux amd64 static executable
REM This script cross-compiles a static Linux binary from Windows

echo Building go-mod-clone for Linux amd64...

REM Set environment variables for Linux cross-compilation
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

REM Build the static executable with optimizations
go build -v ^
  -ldflags="-s -w" ^
  -o go-mod-clone ^
  ./cmd/go-mod-clone

if errorlevel 1 (
  echo Build failed!
  exit /b 1
)

echo Build successful! Created: go-mod-clone
echo.
echo File info:
dir go-mod-clone

endlocal
