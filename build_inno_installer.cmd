@echo off
echo === BUILDING INNO SETUP INSTALLER ===

echo [1/3] Building application...
call .\build_no_gpu.cmd
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Failed to build application
    goto end
)

echo.
echo [2/3] Checking for Inno Setup...
set INNO_PATH=
if exist "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" (
    set INNO_PATH="C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
) else if exist "C:\Program Files\Inno Setup 6\ISCC.exe" (
    set INNO_PATH="C:\Program Files\Inno Setup 6\ISCC.exe"
) else (
    echo ERROR: Inno Setup not found!
    echo Please download and install Inno Setup from:
    echo https://jrsoftware.org/isdl.php
    echo.
    echo Direct download: innosetup-6.4.3.exe
    goto end
)

echo ✅ Inno Setup found at %INNO_PATH%

echo.
echo [3/3] Building installer with Inno Setup...
%INNO_PATH% installer.iss

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ SUCCESS: Inno Setup installer created!
    echo.
    echo Output file: PFInstaller-Setup-InnoSetup.exe
    echo.
    echo This installer should work better than NSIS!
) else (
    echo ERROR: Inno Setup build failed
)

:end
echo.
pause 