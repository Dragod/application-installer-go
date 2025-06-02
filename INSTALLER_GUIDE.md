# Windows Installer Guide for PfCode - Application Installer

## Overview

This guide covers how to build and use the Windows installer for PfCode - Application Installer. The installer is built using Inno Setup and provides a professional installation experience for Windows users.

## Prerequisites

### Required Software

1. **Inno Setup 6.2.0 or later**

   - Download from: https://jrsoftware.org/isdl.php
   - Install to default location: `C:\Program Files (x86)\Inno Setup 6\`

2. **Built Application**

   - Execute: `build_no_gpu.cmd`
   - Ensures: `pfcode-installer.exe` exists in root directory

3. **Documentation Files**
   - `README.md` - Application documentation
   - `LICENSE.txt` - License information

## Building the Installer

### Automatic Build (Recommended)

```batch
# Execute the automated build script
.\build_inno_installer.cmd
```

This script will:

1. Build the application (`pfcode-installer.exe`)
2. Compile the installer using Inno Setup
3. Create `PfCode-ApplicationInstaller-Setup.exe`

### Manual Build Steps

1. **Build the Application First**
   ```batch
   .\build_no_gpu.cmd
   ```
2. **Compile the Installer**
   ```batch
   "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" installer.iss
   ```

### Build Output

- **Output File**: `PfCode-ApplicationInstaller-Setup.exe`
- **Size**: ~19MB (contains the ~49MB application plus installer overhead)
- **Location**: Project root directory

## Installer Features

### User Experience

- **Modern UI**: Clean, professional installation wizard
- **User-Mode Installation**: No administrator privileges required
- **Optional Desktop Icon**: Unchecked by default
- **Start Menu Integration**: Creates program group
- **File Associations**: Associates CSV files for import functionality

### Installation Details

#### Default Installation Path

```
%USERPROFILE%\AppData\Local\Programs\PfCode - Application Installer\
```

#### Application Data Directory

```
%APPDATA%\PfCode - Application Installer\
```

#### Files Installed

- `pfcode-installer.exe` - Main application executable
- `README.md` - Application documentation

#### Registry Entries (User Mode)

- File association for CSV import (`HKCU\SOFTWARE\Classes\.csv`)
- Application registration (`HKCU\SOFTWARE\Classes\PfCodeInstaller.csv`)

### Start Menu Items

- "PfCode - Application Installer" - Launches the application
- "Uninstall PfCode - Application Installer" - Uninstalls the program

### Desktop Icon (Optional)

- Created only if user selects the option during installation
- Links to the main application executable

## Installation Process

### Running the Installer

1. **Download**: Get `PfCode-ApplicationInstaller-Setup.exe`
2. **Execute**: Double-click to run (no admin rights needed)
3. **Follow Wizard**: Complete the installation steps
4. **Launch**: Optionally start the application immediately

### Installation Wizard Steps

1. **Welcome Screen**: Introduction and setup information
2. **License Agreement**: Review and accept MIT license
3. **Information Screen**: Shows README.md content
4. **Destination Folder**: Choose installation directory (user mode default)
5. **Start Menu Folder**: Configure program group name
6. **Additional Tasks**: Select desktop icon creation
7. **Ready to Install**: Summary before installation
8. **Installing**: Progress indicator during file copying
9. **Finished**: Option to launch application immediately

## Uninstallation

### Automatic Uninstaller

The installer creates a standard Windows uninstaller:

- **Location**: `%USERPROFILE%\AppData\Local\Programs\PfCode - Application Installer\unins000.exe`
- **Access**: Via "Add or Remove Programs" in Windows Settings
- **Start Menu**: "Uninstall PfCode - Application Installer"

### What Gets Removed

- All installed application files
- Start Menu shortcuts and program group
- Desktop icon (if created)
- Registry entries for file associations
- **Note**: User data in `%APPDATA%\PfCode - Application Installer\` is preserved

### Manual Cleanup (if needed)

```batch
# Remove user data directory
rmdir /s "%APPDATA%\PfCode - Application Installer"

# Remove any remaining program files
rmdir /s "%USERPROFILE%\AppData\Local\Programs\PfCode - Application Installer"
```

## Installer Configuration

### Key Settings in `installer.iss`

```ini
[Setup]
AppName=PfCode - Application Installer
AppVersion=1.0.0
DefaultDirName={userpf}\PfCode - Application Installer
PrivilegesRequired=lowest
OutputBaseFilename=PfCode-ApplicationInstaller-Setup
```

### User-Mode Installation Benefits

- **No Admin Required**: Users can install without administrator privileges
- **User Directory**: Installs to user's local programs folder
- **Safer**: Reduces security risks and UAC prompts
- **Modern Standard**: Follows Windows 10/11 best practices

## Troubleshooting

### Build Issues

#### "ISCC.exe not found"

```batch
# Verify Inno Setup installation
dir "C:\Program Files (x86)\Inno Setup 6\ISCC.exe"

# If not found, download and install Inno Setup from:
# https://jrsoftware.org/isdl.php
```

#### "File not found: pfcode-installer.exe"

```batch
# Build the application first
.\build_no_gpu.cmd

# Verify the executable exists
dir pfcode-installer.exe
```

#### "Permission denied" during build

```batch
# Close any running instances of the application
taskkill /f /im pfcode-installer.exe

# Try building again
.\build_inno_installer.cmd
```

### Installation Issues

#### "Installation failed" or "Access denied"

- **Cause**: Insufficient permissions or locked files
- **Solution**: Close all instances of PfCode - Application Installer and retry

#### "Cannot create shortcut"

- **Cause**: Start Menu folder permissions
- **Solution**: Try installing to a different Start Menu folder

#### File association not working

- **Cause**: Registry permissions or conflicting associations
- **Solution**: Manually associate .csv files with PfCode - Application Installer

### Runtime Issues

#### Application won't start after installation

1. **Check executable**:

   ```batch
   dir "%USERPROFILE%\AppData\Local\Programs\PfCode - Application Installer\pfcode-installer.exe"
   ```

2. **Test direct execution**:

   ```batch
   cd "%USERPROFILE%\AppData\Local\Programs\PfCode - Application Installer"
   .\pfcode-installer.exe
   ```

3. **Check dependencies**:
   - Ensure Microsoft Visual C++ Redistributable is installed
   - Verify Windows version compatibility (Windows 10/11)

#### Database creation errors

- **Cause**: Application data directory permissions
- **Solution**: Manually create directory:
  ```batch
  mkdir "%APPDATA%\PfCode - Application Installer"
  ```

## Advanced Configuration

### Customizing the Installer

#### Changing Installation Directory

```ini
# In installer.iss [Setup] section
DefaultDirName={pf}\PfCode - Application Installer  ; Program Files (requires admin)
DefaultDirName={userpf}\PfCode - Application Installer  ; User Programs (current)
DefaultDirName={localappdata}\PfCode - Application Installer  ; Local AppData
```

#### Adding More File Associations

```ini
# In installer.iss [Registry] section
Root: HKCU; Subkey: "SOFTWARE\Classes\.json\OpenWithProgids"; ValueType: string; ValueName: "PfCodeInstaller.json"; ValueData: ""; Flags: uninsdeletevalue
```

#### Custom Icons and Graphics

- Replace default icons with custom ones
- Add installer graphics (164x314 pixels recommended)
- Customize installer colors and fonts

### Building Portable Version

For users who prefer portable applications:

```batch
# Build standalone version
.\build_no_gpu.cmd

# Create portable package
mkdir PfCode-Portable
copy pfcode-installer.exe PfCode-Portable\
copy README.md PfCode-Portable\
copy LICENSE.txt PfCode-Portable\

# Create archive
powershell Compress-Archive PfCode-Portable PfCode-ApplicationInstaller-Portable.zip
```

## Distribution

### Release Checklist

- [ ] Application builds without errors (`build_no_gpu.cmd`)
- [ ] Installer compiles successfully (`build_inno_installer.cmd`)
- [ ] Test installation on clean Windows system
- [ ] Verify application launches after installation
- [ ] Test uninstallation process
- [ ] Check file associations work correctly
- [ ] Validate all shortcuts function properly

### Digital Signing (Optional)

For distribution, consider code signing:

```batch
# Sign the main executable
signtool sign /f certificate.pfx /p password pfcode-installer.exe

# Sign the installer
signtool sign /f certificate.pfx /p password PfCode-ApplicationInstaller-Setup.exe
```

### Publishing Locations

- **GitHub Releases**: Attach installer to release tags
- **Microsoft Store**: Consider MSIX packaging for Store distribution
- **Chocolatey**: Create Chocolatey package for command-line installation
- **Winget**: Submit to Windows Package Manager community repository

---

**Note**: This installer provides a professional installation experience while maintaining simplicity and user-mode operation for maximum compatibility and ease of use.
