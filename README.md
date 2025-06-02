# PfCode - Application Installer

A modern GUI application for Windows that allows you to search, install, and organize applications using both Winget and Chocolatey package managers. Features comprehensive list management for organizing applications into custom categories.

![PfCode Application Installer](https://img.shields.io/badge/Platform-Windows-blue)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8)
![License](https://img.shields.io/badge/License-MIT-green)

## 🚀 Features

### 📦 **Package Management**

- **Search Applications**: Search across both Winget and Chocolatey repositories
- **Install Applications**: One-click installation with progress tracking
- **Source Filtering**: Filter by Winget, Chocolatey, or show all sources
- **Batch Installation**: Install multiple applications from lists at once

### 📋 **Advanced List Management**

- **Multiple Named Lists**: Create unlimited custom lists (Work Apps, Gaming, Development Tools, etc.)
- **Cross-List Support**: Save applications to multiple lists simultaneously
- **List Operations**: Create, edit, delete, and organize lists with descriptions
- **Visual Indicators**: See which lists contain each application
- **Smart Navigation**: Auto-switch to "Saved Apps" when selecting a list

### 🎯 **Filtering & Views**

- **All Results**: Combined search results and saved applications
- **Installed Only**: View only currently installed applications
- **Saved Apps**: Browse applications in the selected list
- **Source Filtering**: Filter by package manager (Winget/Chocolatey)
- **Real-time Updates**: Instant filtering and list switching

### 💾 **Data Management**

- **SQLite Database**: Robust local storage for lists and applications
- **Persistent Storage**: All lists and saved apps survive application restarts
- **Data Integrity**: Foreign key constraints and proper relationships
- **Backup Friendly**: Simple database file for easy backup/restore
- **CSV Export**: Export lists to CSV format for external use and backup

### 🎨 **Modern UI**

- **Dark/Light Themes**: Customizable appearance
- **Responsive Layout**: Adaptive interface for different screen sizes
- **Contextual Buttons**: Smart button text based on current view
- **Professional Design**: Clean, intuitive interface built with Fyne
- **Comprehensive Help**: Built-in user guide accessible via toolbar

## 📋 Prerequisites

### **System Requirements**

- **OS**: Windows 10/11 (64-bit)
- **Go**: Version 1.21 or later (for building from source)
- **CGO**: Required for SQLite integration
- **MinGW64**: Required for building (via MSYS2)

### **Package Managers**

- **Winget**: Usually pre-installed on Windows 11, [install guide](https://docs.microsoft.com/en-us/windows/package-manager/winget/)
- **Chocolatey**: Optional but recommended, [install guide](https://chocolatey.org/install)

### **Build Environment Setup**

1. **Install MSYS2**: Download from [msys2.org](https://www.msys2.org/)
2. **Install MinGW64 toolchain**:
   ```bash
   pacman -S mingw-w64-x86_64-gcc
   pacman -S mingw-w64-x86_64-sqlite3
   ```
3. **Add to PATH**: `C:\msys64\mingw64\bin`

## 🛠️ Building & Installation

### **Quick Build**

```bash
# Clone the repository
git clone <repository-url>
cd pf-installer

# Install Go dependencies
go mod tidy

# Build GUI application (Windows)
.\build_no_gpu.cmd
```

### **Manual Build Steps**

```bash
# Set build environment
set PATH=C:\msys64\mingw64\bin;%PATH%
set CGO_ENABLED=1
set CC=gcc

# Build without console window
go build -ldflags "-H windowsgui" -tags "software" -o pfcode-installer.exe .

# Build with console (for debugging)
go build -o pfcode-installer-debug.exe .
```

### **Build Scripts**

- `build_no_gpu.cmd`: Main build script for GUI application
- `build_inno_installer.cmd`: Build complete installer package
- Creates `pfcode-installer.exe` (~49MB)

## 📱 Usage Guide

### **Getting Started**

1. **Launch**: Run `pfcode-installer.exe`
2. **Search**: Enter application names in the search box
3. **Install**: Click install buttons for any application
4. **Organize**: Create custom lists to organize your applications
5. **Export**: Save your lists as CSV files for backup

### **Package Search & Installation**

1. **Search Applications**:

   - Type application names (e.g., "Visual Studio Code", "Discord")
   - Results show from both Winget and Chocolatey sources
   - Filter by source using the dropdown menu

2. **Install Applications**:

   - Click "Install" next to any application
   - Windows will prompt for admin privileges when needed
   - Installation progress shows in the background

3. **View Installed Apps**:
   - Use "Installed Only" filter to see current installations
   - Refresh the list to check for newly installed applications

### **List Management System**

#### **Creating & Managing Lists**

1. **Create New Lists**:

   - Click "Manage Lists" button
   - Choose "Create New List"
   - Enter name and optional description
   - Examples: "Work Apps", "Gaming Tools", "Development Environment"

2. **Switch Between Lists**:

   - Use the dropdown menu to select different lists
   - View automatically switches to "Saved Apps" when selecting a list
   - Each list maintains its own collection of applications

3. **Edit List Details**:
   - Click "Manage Lists" → "Edit Selected List"
   - Update name, description, or delete lists
   - Default list cannot be deleted but can be renamed

#### **Organizing Applications**

1. **Save to Lists**:

   - Find any application (search results or installed apps)
   - Click "Save to List" button
   - Select target list from the dropdown
   - Apps can be saved to multiple lists simultaneously

2. **Remove from Lists**:

   - Method 1: Select target list → "Saved Apps" view → "Remove from [ListName]"
   - Method 2: Click "Manage Lists" on any saved app → uncheck lists → "Apply Changes"

3. **Batch Operations**:
   - "Install All in List": Install all applications from the current list
   - Great for setting up new systems or environments
   - Windows will handle UAC prompts for each installation

### **Advanced Features**

#### **Filtering & Views**

- **All Results**: Shows search results plus saved applications
- **Installed Only**: Displays only currently installed applications
- **Saved Apps**: Shows applications in the currently selected list
- **Source Filters**: Winget only, Chocolatey only, or All Sources
- **Search within Views**: Search functionality works within any view

#### **Data Export & Backup**

1. **CSV Export**:

   - Select any list from the dropdown
   - Click "Export to CSV" (when available)
   - Files saved to `exports/` folder
   - Includes app names, versions, sources, and descriptions

2. **Database Backup**:
   - Database stored in: `%APPDATA%\PfCode - Application Installer\applications.db`
   - Copy this file to backup all lists and saved applications
   - Restore by replacing the file (while application is closed)

### **Application Settings**

Access via toolbar Settings button:

- **Package Managers**: Enable/disable Winget or Chocolatey
- **Theme**: Switch between Dark and Light themes
- **Validation**: Prevents disabling both package managers

### **File Locations**

- **Executable**: `pfcode-installer.exe`
- **Database**: `%APPDATA%\PfCode - Application Installer\applications.db`
- **Logs**: `app.log` (for debugging)
- **Exports**: `exports\` folder (CSV files)
- **Configuration**: Stored in application settings

## 📚 Help & Documentation

### **Built-in Help**

- Click the **Help icon** (❓) in the toolbar
- Comprehensive user guide with all features
- Searchable content and examples

### **Troubleshooting**

#### **Build Issues**

```bash
# Missing GCC
pacman -S mingw-w64-x86_64-gcc

# CGO disabled
set CGO_ENABLED=1

# Wrong path
set PATH=C:\msys64\mingw64\bin;%PATH%
```

#### **Runtime Issues**

- **No search results**: Check package manager installation
- **Install failures**: Run as administrator when prompted
- **Database errors**: Delete applications.db to recreate
- **UI freezing**: Check for antivirus interference

#### **Package Manager Verification**

```powershell
# Test Winget
winget --version
winget search "notepad++"

# Test Chocolatey
choco --version
choco search "notepad++"
```

## 🚀 Installation

### **Pre-built Installer**

1. Download `PfCode-ApplicationInstaller-Setup.exe`
2. Run the installer (no admin privileges required)
3. Launch "PfCode - Application Installer" from Start Menu

### **Portable Usage**

1. Build with `.\build_no_gpu.cmd`
2. Run `pfcode-installer.exe` directly
3. Database created in same directory

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**PfCode - Application Installer**: Making Windows package management simple and organized! 🚀
