@echo off
echo Setting up build environment (Software Rendering)...
set PATH=C:\msys64\mingw64\bin;%PATH%
set CGO_ENABLED=1
set CC=gcc
set FYNE_THEME=dark

echo Building GUI application (no GPU acceleration, no console window, dark theme)...
go build -ldflags "-H windowsgui" -tags "software" -o pfcode-installer.exe

if %ERRORLEVEL% EQU 0 (
    echo Build successful!
    echo This version should not trigger NVIDIA overlay and uses dark theme
    echo Note: This version does NOT require admin privileges
    dir pfcode-installer.exe
) else (
    echo Build failed with error code %ERRORLEVEL%
)