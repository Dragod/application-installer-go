//go:build !console
// +build !console

package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"runtime/debug"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	APP_TITLE = "PfCode - Application Installer"
	APP_ID    = "com.pfcode.application-installer"
)

// Global variable to track current theme state
var isDarkTheme = true // Default to dark theme

func main() {
	// Set up logging to file
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Set log output to file
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("=== APPLICATION STARTING ===")

	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC CAUGHT IN MAIN: %v", r)
			log.Printf("Stack trace: %s", debug.Stack())
			fmt.Printf("Application crashed: %v\n", r)
			fmt.Printf("Check app.log for detailed stack trace\n")
		}
	}()

	log.Println("Creating Fyne application...")
	myApp := app.NewWithID(APP_ID)

	// Set dark theme as default
	myApp.Settings().SetTheme(&darkTheme{})

	log.Println("Fyne application created successfully")

	log.Println("Creating main window...")
	myWindow := myApp.NewWindow(APP_TITLE)
	myWindow.Resize(fyne.NewSize(1920, 1080)) // Back to original 4K size
	myWindow.CenterOnScreen()

	// Set window icon
	log.Println("Setting window icon...")
	myWindow.SetIcon(resourceFaviconPng)
	log.Println("Window icon set successfully (embedded resource)")

	// Force window to be visible and on top
	myWindow.SetOnClosed(func() {
		log.Println("Window closed by user")
	})

	log.Println("Main window created successfully")

	log.Println("Initializing database...")
	db, err := InitDB()
	if err != nil {
		log.Printf("Database initialization failed: %v", err)
		fmt.Printf("Database error: %v\n", err)
		return
	}
	defer db.Close()
	log.Println("Database initialized successfully")

	log.Println("Creating AppManager...")
	appManager := NewAppManager(db)
	log.Println("AppManager created successfully")

	log.Println("Creating main UI...")
	content := createMainUI(myWindow, appManager)
	log.Println("Main UI created successfully")

	log.Println("Setting window content...")
	myWindow.SetContent(content)
	log.Println("Window content set successfully")

	log.Println("Showing window...")
	myWindow.Show()
	log.Println("Window shown successfully")

	log.Println("Running application...")

	// Auto-load installed apps after UI is fully created (safe from deadlocks)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Error auto-loading installed apps: %v", r)
			}
		}()
		appManager.RefreshInstalledApps()
	}()

	myApp.Run()
	log.Println("Application closed normally")
}

func createMainUI(window fyne.Window, appManager *AppManager) *fyne.Container {
	log.Println("Creating toolbar...")
	// Create toolbar
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			appManager.RefreshInstalledApps()
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			showHelp(window)
		}),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			showAbout(window)
		}),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			showSettings(window)
		}),
	)
	log.Println("Toolbar created successfully")

	log.Println("Creating main content...")
	// Create main content
	mainContent := createMainContent(appManager)
	log.Println("Main content created successfully")

	log.Println("Creating border container...")
	borderContainer := container.NewBorder(
		toolbar,     // top
		nil,         // bottom
		nil,         // left
		nil,         // right
		mainContent, // center
	)
	log.Println("Border container created successfully")

	return borderContainer
}

func showSettings(parent fyne.Window) {
	settingsWindow := fyne.CurrentApp().NewWindow("Settings")
	settingsWindow.Resize(fyne.NewSize(450, 350))
	settingsWindow.CenterOnScreen()

	// Package Manager Settings
	wingetCheck := widget.NewCheck("Enable Winget", nil)
	chocoCheck := widget.NewCheck("Enable Chocolatey", nil)

	// Set up validation callbacks after creating both checkboxes
	wingetCheck.OnChanged = func(checked bool) {
		// Prevent disabling both package managers
		if !checked && !getChocoEnabled() {
			// Show warning and revert the change
			dialog.ShowError(fmt.Errorf("At least one package manager must be enabled.\nWinget will remain enabled."), settingsWindow)
			wingetCheck.SetChecked(true) // Revert the change
			return
		}
		setWingetEnabled(checked)
		log.Printf("Winget %s", map[bool]string{true: "enabled", false: "disabled"}[checked])
	}

	chocoCheck.OnChanged = func(checked bool) {
		// Prevent disabling both package managers
		if !checked && !getWingetEnabled() {
			// Show warning and revert the change
			dialog.ShowError(fmt.Errorf("At least one package manager must be enabled.\nChocolatey will remain enabled."), settingsWindow)
			chocoCheck.SetChecked(true) // Revert the change
			return
		}
		setChocoEnabled(checked)
		log.Printf("Chocolatey %s", map[bool]string{true: "enabled", false: "disabled"}[checked])
	}

	wingetCheck.SetChecked(getWingetEnabled())
	chocoCheck.SetChecked(getChocoEnabled())

	// Add informational note
	infoNote := widget.NewLabel("* At least one package manager must be enabled")
	infoNote.TextStyle = fyne.TextStyle{Italic: true}

	// Theme Settings
	themeLabel := widget.NewLabel("Theme:")
	themeLabel.TextStyle = fyne.TextStyle{Bold: true}

	themeRadio := widget.NewRadioGroup(
		[]string{"Dark Theme", "Light Theme"},
		func(value string) {
			if value == "Dark Theme" && !isDarkTheme {
				// Switch to dark theme
				fyne.CurrentApp().Settings().SetTheme(&darkTheme{})
				isDarkTheme = true
				log.Println("Switched to dark theme")
			} else if value == "Light Theme" && isDarkTheme {
				// Switch to light theme
				fyne.CurrentApp().Settings().SetTheme(&lightTheme{})
				isDarkTheme = false
				log.Println("Switched to light theme")
			}
		},
	)

	// Set current theme selection
	if isDarkTheme {
		themeRadio.SetSelected("Dark Theme")
	} else {
		themeRadio.SetSelected("Light Theme")
	}

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Package Managers", Widget: container.NewVBox(wingetCheck, chocoCheck, infoNote)},
			{Text: "", Widget: widget.NewSeparator()}, // Visual separator
			{Text: "Appearance", Widget: container.NewVBox(themeLabel, themeRadio)},
		},
		OnSubmit: func() {
			settingsWindow.Close()
		},
		OnCancel: func() {
			settingsWindow.Close()
		},
	}

	content := container.NewVBox(
		widget.NewLabel("Application Settings"),
		widget.NewSeparator(),
		form,
	)

	settingsWindow.SetContent(content)
	settingsWindow.Show()
}

func createMainContent(appManager *AppManager) *container.Split {
	log.Println("Creating left panel (search)...")
	// Left panel - Search and filters
	leftPanel := createSearchPanel(appManager)
	log.Println("Left panel created successfully")

	log.Println("Creating right panel (app list)...")
	// Right panel - App list and details
	rightPanel := createAppListPanel(appManager)
	log.Println("Right panel created successfully")

	log.Println("Creating horizontal split...")
	// Create horizontal split with much more space for the right panel (applications list)
	split := container.NewHSplit(leftPanel, rightPanel)
	log.Println("Horizontal split created successfully")

	log.Println("Setting split offset...")
	split.SetOffset(0.2) // Give 20% to left panel, 80% to right panel
	log.Println("Split offset set successfully")

	return split
}

func showHelp(parent fyne.Window) {
	helpWindow := fyne.CurrentApp().NewWindow("PfCode - Application Installer - How to Use")
	helpWindow.Resize(fyne.NewSize(1000, 800))
	helpWindow.CenterOnScreen()

	helpContent := `🚀 PFCODE - APPLICATION INSTALLER - COMPLETE USER GUIDE

📋 OVERVIEW

This application installer helps you discover, install, and organize applications from multiple sources (Winget, Chocolatey). You can create custom lists to organize your applications and manage them efficiently.


🔍 SEARCHING & INSTALLING

• Enter search terms in the search box (e.g., "Visual Studio Code", "Discord")
• Click "Search" to find applications from all enabled sources
• Click "Install" next to any app to install it on your system
• Use "Clear" to reset your search and return to browsing mode


📊 FILTERING OPTIONS

View Filters:
• All Results: Shows search results or combined installed + saved apps
• Installed Only: Shows only applications currently installed on your system
• Saved Apps: Shows apps saved in the currently selected list

Source Filters:
• All Sources: Shows apps from both Winget and Chocolatey
• Winget: Shows only Windows Package Manager apps
• Chocolatey: Shows only Chocolatey package manager apps


📋 LIST MANAGEMENT

• Use the dropdown to switch between your application lists
• Click "Manage Lists" to create, edit, or delete lists
• Each list can have a name and optional description
• Default list cannot be deleted (but can be renamed)
• Selecting a list automatically switches to "Saved Apps" view


💾 SAVING APPLICATIONS

• Click "Save to List" to add an app to a specific list
• Apps can be saved to multiple lists simultaneously
• Saved apps show "In lists: [List1, List2, ...]" for easy reference
• Use "Manage Lists" button on saved apps for advanced list management


🔄 MANAGING APPS IN LISTS

Basic Method:
1. Select the target list from dropdown
2. Switch to "Saved Apps" view (or it auto-switches)
3. Click "Remove from [ListName]" on any app

Advanced Method:
1. Find any saved app (in any view)
2. Click "Manage Lists" or "Saved in X lists" button
3. Use checkboxes to add/remove from multiple lists at once
4. Click "Apply Changes" to save modifications


⚡ QUICK ACTIONS

• "Refresh Installed": Updates the list of installed applications
• "Install All in List": Installs all apps from the currently selected list
• List dropdown: Instantly switch between your organized lists
• Auto-switch: Selecting a list automatically shows its contents


🎯 VISUAL INDICATORS

• ✅ Green check: Application is installed
• ℹ️ Blue info: Application is available but not installed
• 💾 Save icons: Shows which lists contain each app
• List context: Buttons show relevant list names (e.g., "Remove from Games")
• Empty states: Clear messages showing which specific list is empty


💡 PRO TIPS

• Create themed lists: "Work Apps", "Gaming", "Development Tools"
• Use search within saved apps to find specific items in large lists
• Apps like browsers or chat tools can be in multiple lists
• List descriptions help you remember what each list is for
• Source filtering helps when you prefer specific package managers


🛠️ TROUBLESHOOTING

• If no results appear, check that Winget/Chocolatey are properly installed
• Empty lists show helpful messages with the specific list name
• Use "Refresh Installed" if recently installed apps don't appear
• List changes update immediately when switching between lists


📱 KEYBOARD SHORTCUTS

• Enter in search box: Triggers search
• Tab: Navigate between UI elements
• Escape: Close dialogs and popups


⚙️ SYSTEM REQUIREMENTS

• Windows 10/11 with Winget installed (usually included)
• Chocolatey (optional, but recommended for more software options)
• Administrator privileges may be required for some installations
• Internet connection for searching and downloading applications


🎉 GETTING STARTED

This tool is designed to make package management simple and organized. Start by searching for your favorite applications, save them to custom lists, and use the "Install All in List" feature for quick system setup!`

	helpLabel := widget.NewRichTextFromMarkdown(helpContent)
	helpLabel.Wrapping = fyne.TextWrapWord

	scrollContainer := container.NewScroll(helpLabel)
	scrollContainer.SetMinSize(fyne.NewSize(950, 700))

	closeButton := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		helpWindow.Close()
	})
	closeButton.Importance = widget.HighImportance

	content := container.NewBorder(
		widget.NewLabelWithStyle("PfCode - Application Installer - User Guide",
			fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), // top
		container.NewHBox(closeButton), // bottom
		nil,                            // left
		nil,                            // right
		scrollContainer,                // center
	)

	helpWindow.SetContent(content)
	helpWindow.Show()
}

func showAbout(parent fyne.Window) {
	aboutWindow := fyne.CurrentApp().NewWindow("About PfCode - Application Installer")
	aboutWindow.Resize(fyne.NewSize(500, 400))
	aboutWindow.CenterOnScreen()

	appTitle := widget.NewLabel("PfCode - Application Installer")
	appTitle.TextStyle = fyne.TextStyle{Bold: true}
	appTitle.Alignment = fyne.TextAlignCenter

	appVersion := widget.NewLabel("Version 1.0.0")
	appVersion.Alignment = fyne.TextAlignCenter

	aboutContent := `📦 PfCode - Application Installer

A modern GUI application for managing Windows applications using package managers.

📅 Build Date: December 2024
🏗️ Version: 1.0.0
👨‍💻 Author: PfCode
📧 Email: support@pfcode.dev
🌐 Website: https://github.com/pfcode/application-installer

🔧 Technical Details:
• Built with Go and Fyne GUI framework
• Uses Winget and Chocolatey package managers
• SQLite3 database for saved applications
• Cross-platform compatible (Windows focus)

💻 System Requirements:
• Windows 10/11
• Winget (included with Windows)
• Chocolatey (optional)

📄 License: MIT License
© 2025 PfCode. All rights reserved

Special thanks to the open-source community and the developers of Fyne, Go, Winget, and Chocolatey.`

	aboutLabel := widget.NewRichTextFromMarkdown(aboutContent)
	aboutLabel.Wrapping = fyne.TextWrapWord

	scrollContainer := container.NewScroll(aboutLabel)

	headerContainer := container.NewVBox(
		widget.NewSeparator(),
		appTitle,
		appVersion,
		widget.NewSeparator(),
	)

	closeButton := widget.NewButton("Close", func() {
		aboutWindow.Close()
	})

	content := container.NewBorder(
		headerContainer, // top
		closeButton,     // bottom
		nil,             // left
		nil,             // right
		scrollContainer, // center
	)

	aboutWindow.SetContent(content)
	aboutWindow.Show()
}

// Custom dark theme with colorful buttons
type darkTheme struct{}

func (t *darkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameButton:
		return color.RGBA{R: 70, G: 130, B: 180, A: 255} // Steel blue
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0, G: 150, B: 255, A: 255} // Bright blue
	case theme.ColorNameSuccess:
		return color.RGBA{R: 40, G: 200, B: 40, A: 255} // Green
	case theme.ColorNameWarning:
		return color.RGBA{R: 255, G: 165, B: 0, A: 255} // Orange
	case theme.ColorNameError:
		return color.RGBA{R: 255, G: 60, B: 60, A: 255} // Red
	case theme.ColorNameHover:
		return color.RGBA{R: 100, G: 160, B: 210, A: 255} // Light blue hover
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (t *darkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *darkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *darkTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Custom light theme with colorful buttons
type lightTheme struct{}

func (t *lightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameButton:
		return color.RGBA{R: 100, G: 149, B: 237, A: 255} // Cornflower blue
	case theme.ColorNamePrimary:
		return color.RGBA{R: 30, G: 144, B: 255, A: 255} // Dodger blue
	case theme.ColorNameSuccess:
		return color.RGBA{R: 50, G: 205, B: 50, A: 255} // Lime green
	case theme.ColorNameWarning:
		return color.RGBA{R: 255, G: 140, B: 0, A: 255} // Dark orange
	case theme.ColorNameError:
		return color.RGBA{R: 220, G: 20, B: 60, A: 255} // Crimson
	case theme.ColorNameHover:
		return color.RGBA{R: 135, G: 206, B: 250, A: 255} // Light sky blue hover
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

func (t *lightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *lightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *lightTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
