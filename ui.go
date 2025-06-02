//go:build !console
// +build !console

package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func createSearchPanel(appManager *AppManager) *fyne.Container {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in createSearchPanel: %v", r)
			log.Printf("Stack trace: %s", debug.Stack())
			panic(r) // Re-panic to maintain the original behavior
		}
	}()

	// Search entry
	log.Println("Creating search entry...")
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search for applications...")
	log.Println("Search entry created successfully")

	log.Println("Creating search button...")
	searchButton := widget.NewButtonWithIcon("Search", theme.SearchIcon(), nil)
	searchButton.Importance = widget.HighImportance // Prominent blue styling
	searchButton.OnTapped = func() {
		query := searchEntry.Text

		// Check for empty query and show error
		if strings.TrimSpace(query) == "" {
			// Get the main window for dialogs
			windows := fyne.CurrentApp().Driver().AllWindows()
			if len(windows) > 0 {
				mainWindow := windows[0]
				dialog.ShowError(fmt.Errorf("Please enter a search term"), mainWindow)
			}
			return
		}

		searchButton.SetText("Searching...")
		searchButton.Disable()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
				}
				searchButton.SetText("Search")
				searchButton.Enable()
			}()
			appManager.SearchApps(query)
		}()
	}
	log.Println("Search button created successfully")

	log.Println("Creating clear button...")
	clearButton := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), nil)
	clearButton.Importance = widget.MediumImportance // Medium blue styling
	clearButton.OnTapped = func() {
		searchEntry.SetText("")
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
				}
			}()
			appManager.ClearSearch()
		}()
	}
	log.Println("Clear button created successfully")

	log.Println("Creating filter radio group...")
	// Filter radio buttons
	var filterGroup *widget.RadioGroup
	filterGroup = widget.NewRadioGroup(
		[]string{"All Results", "Installed Only", "Saved Apps"},
		func(value string) {
			// Use unified filtering system
			appManager.SetViewFilter(value)
		},
	)
	filterGroup.SetSelected("Installed Only") // Default to installed since that's what we load on startup
	log.Println("Filter radio group created successfully")

	log.Println("Creating source radio group...")
	// Source filter
	sourceGroup := widget.NewRadioGroup(
		[]string{"All Sources", "Winget", "Chocolatey"},
		func(value string) {
			// Filter current apps by source
			switch value {
			case "All Sources":
				// Show all current apps
				appManager.FilterBySource("All Sources")
			case "Winget":
				// Show only winget apps
				appManager.FilterBySource("winget")
			case "Chocolatey":
				// Show only chocolatey apps
				appManager.FilterBySource("chocolatey")
			}
		},
	)
	sourceGroup.SetSelected("All Sources")
	log.Println("Source radio group created successfully")

	log.Println("Creating list selector...")
	// List selector
	listSelect := widget.NewSelect([]string{}, func(selected string) {
		// Handle list selection
		lists := appManager.GetLists()
		for _, list := range lists {
			if list.Name == selected {
				appManager.SetCurrentList(list)

				// Auto-switch to "Saved Apps" view to show the selected list's contents
				// This makes the list selection immediately visible and intuitive
				appManager.SetViewFilter("Saved Apps")

				// Update the filter radio group to reflect the change
				if filterGroup != nil {
					filterGroup.SetSelected("Saved Apps")
				}
				break
			}
		}
	})
	listSelect.PlaceHolder = "Select a list..."

	// Update list selector with current lists
	updateListSelector := func() {
		lists := appManager.GetLists()
		options := make([]string, len(lists))
		for i, list := range lists {
			options[i] = list.Name
		}
		listSelect.Options = options

		// Set current selection
		currentList := appManager.GetCurrentList()
		if currentList != nil {
			listSelect.SetSelected(currentList.Name)
		}
	}

	// Initial load of lists
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Handle panic gracefully
			}
		}()
		updateListSelector()
	}()

	log.Println("Creating manage lists button...")
	manageListsButton := widget.NewButtonWithIcon("Manage Lists", theme.FolderOpenIcon(), func() {
		// Get the main window for dialogs
		windows := fyne.CurrentApp().Driver().AllWindows()
		if len(windows) > 0 {
			mainWindow := windows[0]
			showListManagementDialog(mainWindow, appManager, updateListSelector)
		}
	})
	manageListsButton.Importance = widget.MediumImportance
	log.Println("Manage lists button created successfully")

	log.Println("Creating export button...")
	exportButton := widget.NewButtonWithIcon("Export to CSV", theme.DocumentSaveIcon(), nil)
	exportButton.Importance = widget.MediumImportance
	exportButton.OnTapped = func() {
		currentList := appManager.GetCurrentList()
		if currentList == nil {
			// Get the main window for dialogs
			windows := fyne.CurrentApp().Driver().AllWindows()
			if len(windows) > 0 {
				mainWindow := windows[0]
				dialog.ShowError(fmt.Errorf("No list selected"), mainWindow)
			}
			return
		}

		exportButton.SetText("Exporting...")
		exportButton.Disable()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
				}
				exportButton.SetText("Export to CSV")
				exportButton.Enable()
			}()

			err := appManager.ExportCurrentListToCSV()

			// Get the main window for dialogs
			windows := fyne.CurrentApp().Driver().AllWindows()
			if len(windows) == 0 {
				return
			}
			mainWindow := windows[0]

			if err != nil {
				dialog.ShowError(err, mainWindow)
			} else {
				dialog.ShowInformation("Export Complete",
					fmt.Sprintf("List '%s' has been exported to CSV in the exports folder.", currentList.Name),
					mainWindow)
			}
		}()
	}
	log.Println("Export button created successfully")

	log.Println("Creating import button...")
	importButton := widget.NewButtonWithIcon("Import from CSV", theme.FolderOpenIcon(), nil)
	importButton.Importance = widget.MediumImportance
	importButton.OnTapped = func() {
		// Get the main window for dialogs
		windows := fyne.CurrentApp().Driver().AllWindows()
		if len(windows) > 0 {
			mainWindow := windows[0]
			showImportCSVDialog(mainWindow, appManager, updateListSelector)
		}
	}
	log.Println("Import button created successfully")

	log.Println("Creating refresh button...")
	// Quick actions
	refreshButton := widget.NewButtonWithIcon("Refresh Installed", theme.ViewRefreshIcon(), nil)
	refreshButton.Importance = widget.MediumImportance // Blue styling
	refreshButton.OnTapped = func() {
		refreshButton.SetText("Refreshing...")
		refreshButton.Disable()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
				}
				refreshButton.SetText("Refresh Installed")
				refreshButton.Enable()
			}()
			appManager.RefreshInstalledApps()
		}()
	}
	log.Println("Refresh button created successfully")

	log.Println("Creating install all button...")
	installAllButton := widget.NewButtonWithIcon("Install All in List", theme.DownloadIcon(), nil)
	installAllButton.Importance = widget.HighImportance // Prominent styling
	installAllButton.OnTapped = func() {
		currentList := appManager.GetCurrentList()
		if currentList == nil {
			// Get the main window for dialogs
			windows := fyne.CurrentApp().Driver().AllWindows()
			if len(windows) > 0 {
				mainWindow := windows[0]
				dialog.ShowError(fmt.Errorf("No list selected"), mainWindow)
			}
			return
		}

		installAllButton.SetText("Installing...")
		installAllButton.Disable()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
				}
				installAllButton.SetText("Install All in List")
				installAllButton.Enable()
			}()

			err := appManager.InstallAllAppsInCurrentList()

			// Get the main window for dialogs
			windows := fyne.CurrentApp().Driver().AllWindows()
			if len(windows) == 0 {
				return
			}
			mainWindow := windows[0]

			if err != nil {
				dialog.ShowError(err, mainWindow)
			} else {
				listName := "list"
				if currentList := appManager.GetCurrentList(); currentList != nil {
					listName = fmt.Sprintf("'%s'", currentList.Name)
				}
				dialog.ShowInformation("Success", fmt.Sprintf("All applications in %s have been installed.", listName), mainWindow)
			}
		}()
	}
	log.Println("Install all button created successfully")

	log.Println("Creating container layout...")
	container := container.NewVBox(
		widget.NewCard("Search", "", container.NewVBox(
			searchEntry,
			container.NewHBox(searchButton, clearButton),
		)),
		widget.NewCard("Filter", "", container.NewVBox(
			filterGroup,
		)),
		widget.NewCard("Source", "", container.NewVBox(
			sourceGroup,
		)),
		widget.NewCard("Lists", "", container.NewVBox(
			listSelect,
			manageListsButton,
			exportButton,
			importButton,
		)),
		widget.NewCard("Actions", "", container.NewVBox(
			refreshButton,
			installAllButton,
		)),
	)
	log.Println("Container layout created successfully")

	return container
}

func createAppListPanel(appManager *AppManager) *fyne.Container {
	log.Println("Creating app list widget...")
	// Create the list
	appList := widget.NewList(
		func() int {
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
				}
			}()
			apps := appManager.GetCurrentApps()
			return len(apps)
		},
		func() fyne.CanvasObject {
			return createAppListItem()
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
				}
			}()
			apps := appManager.GetCurrentApps()
			if id >= 0 && id < len(apps) {
				updateAppListItem(obj, apps[id], appManager)
			}
		},
	)
	log.Println("App list widget created successfully")

	log.Println("Creating empty state labels...")
	// Empty state message
	emptyMessageLabel := widget.NewLabel("Loading applications...")
	emptyMessageLabel.Alignment = fyne.TextAlignCenter
	emptyMessageLabel.TextStyle = fyne.TextStyle{Italic: true}

	emptyIconLabel := widget.NewLabel("📦")
	emptyIconLabel.Alignment = fyne.TextAlignCenter
	emptyIconLabel.TextStyle = fyne.TextStyle{Bold: true}
	log.Println("Empty state labels created successfully")

	log.Println("Creating empty container...")
	emptyContainer := container.NewVBox(
		widget.NewSeparator(),
		emptyIconLabel,
		emptyMessageLabel,
		widget.NewLabel(""), // spacer
		widget.NewLabel("Try:"),
		widget.NewLabel("• Search for applications"),
		widget.NewLabel("• Refresh installed apps"),
		widget.NewLabel("• Change filter settings"),
		widget.NewSeparator(),
	)
	emptyContainer.Hide() // Initially hidden
	log.Println("Empty container created successfully")

	log.Println("Creating content stack...")
	// Content stack to switch between list and empty state
	contentStack := container.NewStack(appList, emptyContainer)
	log.Println("Content stack created successfully")

	log.Println("Adding callback to app manager...")
	// Add callback to refresh list and update empty state when data changes
	appManager.AddCallback(func() {
		defer func() {
			if r := recover(); r != nil {
				// Handle panic gracefully
			}
		}()

		// Check if we're currently loading
		if appManager.IsLoading() {
			// Show loading state with context-aware message
			emptyIconLabel.SetText("⏳")
			currentViewFilter := appManager.GetCurrentViewFilter()

			var loadingMessage string
			if currentViewFilter == "All Results" {
				loadingMessage = "Searching for applications..."
			} else if currentViewFilter == "Installed Only" {
				loadingMessage = "Refreshing installed applications..."
			} else {
				loadingMessage = "Loading applications..."
			}

			emptyMessageLabel.SetText(loadingMessage)
			emptyContainer.Show()
			appList.Hide()
			return
		}

		// Reset icon when not loading
		emptyIconLabel.SetText("📦")

		apps := appManager.GetCurrentApps()
		if len(apps) == 0 {
			// Show empty state with appropriate message
			updateEmptyStateMessage(emptyMessageLabel, appManager)
			emptyContainer.Show()
			appList.Hide()
		} else {
			// Show list
			emptyContainer.Hide()
			appList.Show()
		}

		appList.Refresh()
	})
	log.Println("Callback added successfully")

	log.Println("Creating status and header labels...")
	// Applications header
	headerLabel := widget.NewLabel("Applications")
	headerLabel.TextStyle = fyne.TextStyle{Bold: true}
	log.Println("Status and header labels created successfully")

	log.Println("Creating border container...")
	// Use border container to give maximum space to the content
	borderContainer := container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator()), // top: header
		nil,          // bottom: no status label
		nil,          // left
		nil,          // right
		contentStack, // center: stack of list or empty state
	)
	log.Println("Border container created successfully")

	return borderContainer
}

func createAppListItem() fyne.CanvasObject {
	nameLabel := widget.NewLabel("")
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}

	versionLabel := widget.NewLabel("")
	packageIDLabel := widget.NewLabel("")
	packageIDLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// New label specifically for showing which lists contain this app
	listsLabel := widget.NewLabel("")
	listsLabel.TextStyle = fyne.TextStyle{Italic: true}
	listsLabel.Importance = widget.MediumImportance

	sourceLabel := widget.NewLabel("")
	sourceLabel.TextStyle = fyne.TextStyle{Italic: true}

	installButton := widget.NewButtonWithIcon("Install", theme.DownloadIcon(), func() {})
	saveButton := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {})
	removeButton := widget.NewButtonWithIcon("Remove", theme.DeleteIcon(), func() {})

	statusIcon := widget.NewIcon(theme.InfoIcon())

	topRow := container.NewHBox(
		statusIcon,
		container.NewVBox(nameLabel, versionLabel, packageIDLabel, listsLabel),
		widget.NewSeparator(),
		sourceLabel,
	)

	// Add padding to align buttons with text content (same width as status icon)
	buttonPadding := widget.NewIcon(theme.InfoIcon())
	buttonPadding.Hide() // Hidden spacer to match icon width

	buttonRow := container.NewHBox(
		buttonPadding, // Spacer to align with text content
		installButton,
		saveButton,
		removeButton,
	)

	return container.NewVBox(
		topRow,
		buttonRow,
		widget.NewSeparator(),
	)
}

func updateAppListItem(obj fyne.CanvasObject, app *AppInfo, appManager *AppManager) {
	if obj == nil || app == nil {
		return
	}

	cont, ok := obj.(*fyne.Container)
	if !ok || len(cont.Objects) < 2 {
		return
	}

	topRow, ok := cont.Objects[0].(*fyne.Container)
	if !ok || len(topRow.Objects) < 4 {
		return
	}

	buttonRow, ok := cont.Objects[1].(*fyne.Container)
	if !ok || len(buttonRow.Objects) < 4 { // Now we have 4 elements (spacer + 3 buttons)
		return
	}

	// Update status icon
	if statusIcon, ok := topRow.Objects[0].(*widget.Icon); ok {
		if app.IsInstalled {
			statusIcon.SetResource(theme.ConfirmIcon())
		} else {
			statusIcon.SetResource(theme.InfoIcon())
		}
	}

	// Update labels
	if labelContainer, ok := topRow.Objects[1].(*fyne.Container); ok && len(labelContainer.Objects) >= 4 {
		if nameLabel, ok := labelContainer.Objects[0].(*widget.Label); ok {
			nameLabel.SetText(app.Name)
		}
		if versionLabel, ok := labelContainer.Objects[1].(*widget.Label); ok {
			versionLabel.SetText(fmt.Sprintf("Version: %s", app.Version))
		}
		if packageIDLabel, ok := labelContainer.Objects[2].(*widget.Label); ok {
			// Clean package ID display without list information
			packageIDLabel.SetText(fmt.Sprintf("Package ID: %s", app.PackageID))
		}
		if listsLabel, ok := labelContainer.Objects[3].(*widget.Label); ok {
			// Show which lists contain this app
			listsContaining, err := appManager.GetAppListsContaining(app.PackageID)
			listsText := "Not in any lists"

			if err == nil && len(listsContaining) > 0 {
				listNames := make([]string, len(listsContaining))
				for i, list := range listsContaining {
					listNames[i] = list.Name
				}
				listsText = fmt.Sprintf("In lists: %s", strings.Join(listNames, ", "))
			}

			listsLabel.SetText(listsText)
		}
	}

	if sourceLabel, ok := topRow.Objects[3].(*widget.Label); ok {
		sourceLabel.SetText(fmt.Sprintf("Source: %s", app.Source))
	}

	// Get current view filter to determine which buttons to show
	currentViewFilter := appManager.GetCurrentViewFilter()

	// Update buttons
	if installButton, ok := buttonRow.Objects[1].(*widget.Button); ok {
		if app.IsInstalled {
			installButton.SetText("Installed")
			installButton.Disable()
		} else {
			installButton.SetText("Install")
			installButton.Enable()
			installButton.OnTapped = func() {
				installButton.SetText("Installing...")
				installButton.Disable()

				go func() {
					defer func() {
						if r := recover(); r != nil {
							installButton.SetText("Install")
							installButton.Enable()
						}
					}()

					err := appManager.InstallApp(app)

					// Get the main window for dialogs
					windows := fyne.CurrentApp().Driver().AllWindows()
					if len(windows) == 0 {
						return
					}
					mainWindow := windows[0]

					if err != nil {
						dialog.ShowError(err, mainWindow)
						installButton.SetText("Install")
						installButton.Enable()
					} else {
						installButton.SetText("Installed")
						app.IsInstalled = true
					}
				}()
			}
		}
	}

	// Save/Remove button (button index 2)
	if actionButton, ok := buttonRow.Objects[2].(*widget.Button); ok {
		if currentViewFilter == "Saved Apps" {
			// Show as Remove button when viewing saved apps
			currentList := appManager.GetCurrentList()
			if currentList != nil {
				actionButton.SetText(fmt.Sprintf("Remove from %s", currentList.Name))
			} else {
				actionButton.SetText("Remove")
			}
			actionButton.SetIcon(theme.DeleteIcon())
			actionButton.Enable()
			actionButton.OnTapped = func() {
				actionButton.SetText("Removing...")
				actionButton.Disable()

				go func() {
					defer func() {
						if r := recover(); r != nil {
							actionButton.SetText("Remove")
							actionButton.Enable()
						}
					}()

					err := appManager.RemoveSavedApp(app.PackageID)

					// Get the main window for dialogs
					windows := fyne.CurrentApp().Driver().AllWindows()
					if len(windows) == 0 {
						return
					}
					mainWindow := windows[0]

					actionButton.SetText("Remove")
					actionButton.Enable()

					if err != nil {
						dialog.ShowError(err, mainWindow)
					} else {
						// If we're viewing saved apps, refresh the view to show the updated list
						if appManager.GetCurrentViewFilter() == "Saved Apps" {
							appManager.SetViewFilter("Saved Apps") // This will refresh the current view
						}
						currentList := appManager.GetCurrentList()
						listName := "list"
						if currentList != nil {
							listName = fmt.Sprintf("'%s'", currentList.Name)
						}
						dialog.ShowInformation("Removed",
							fmt.Sprintf("Application '%s' has been removed from %s.", app.Name, listName),
							mainWindow)
					}
				}()
			}
		} else {
			// Show as Save button when not viewing saved apps
			if app.IsSaved {
				// Check which lists contain this app
				listsContaining, err := appManager.GetAppListsContaining(app.PackageID)
				if err == nil && len(listsContaining) > 0 {
					if len(listsContaining) == 1 {
						actionButton.SetText(fmt.Sprintf("Saved in %s", listsContaining[0].Name))
					} else {
						actionButton.SetText(fmt.Sprintf("Saved in %d lists", len(listsContaining)))
					}
				} else {
					actionButton.SetText("Saved")
				}
				actionButton.SetIcon(theme.ConfirmIcon())
				// Make saved button clickable to show list management
				actionButton.Enable()
				actionButton.OnTapped = func() {
					// Get the main window for dialogs
					windows := fyne.CurrentApp().Driver().AllWindows()
					if len(windows) == 0 {
						return
					}
					mainWindow := windows[0]

					// Show comprehensive list management dialog
					showAppListManagementDialog(mainWindow, appManager, app, func() {
						updateAppListItem(obj, app, appManager)
					})
				}
			} else {
				actionButton.SetText("Save to List")
				actionButton.SetIcon(theme.DocumentSaveIcon())
				actionButton.Enable()
				actionButton.OnTapped = func() {
					// Get the main window for dialogs
					windows := fyne.CurrentApp().Driver().AllWindows()
					if len(windows) == 0 {
						return
					}
					mainWindow := windows[0]

					// Show list selection dialog
					showSaveToListDialog(mainWindow, appManager, app, func() {
						// Refresh the app list item to reflect saved status
						updateAppListItem(obj, app, appManager)
					})
				}
			}
		}
	}

	// Third button for advanced list management
	if manageButton, ok := buttonRow.Objects[3].(*widget.Button); ok {
		if currentViewFilter != "Saved Apps" && app.IsSaved {
			// Show "Manage in Lists" button for saved apps in search/all results
			manageButton.SetText("Manage Lists")
			manageButton.SetIcon(theme.FolderOpenIcon())
			manageButton.Show()
			manageButton.Enable()
			manageButton.OnTapped = func() {
				// Get the main window for dialogs
				windows := fyne.CurrentApp().Driver().AllWindows()
				if len(windows) == 0 {
					return
				}
				mainWindow := windows[0]

				// Show comprehensive list management dialog
				showAppListManagementDialog(mainWindow, appManager, app, func() {
					updateAppListItem(obj, app, appManager)
				})
			}
		} else {
			manageButton.Hide()
		}
	}
}

func updateEmptyStateMessage(messageLabel *widget.Label, appManager *AppManager) {
	// Get current filter states to show appropriate message
	currentSourceFilter := appManager.GetCurrentSourceFilter()
	currentViewFilter := appManager.GetCurrentViewFilter()
	currentSearchQuery := appManager.GetCurrentSearchQuery()

	var message string

	// Check if we're in search mode and have a search query
	isSearching := appManager.IsSearchMode() && currentSearchQuery != ""

	if currentViewFilter == "Installed Only" {
		if isSearching {
			if currentSourceFilter == "All Sources" {
				message = fmt.Sprintf("No application found: \"%s\"\nNo installed apps match your search", currentSearchQuery)
			} else {
				message = fmt.Sprintf("No application found: \"%s\"\nNo installed %s apps match your search", currentSearchQuery, currentSourceFilter)
			}
		} else {
			if currentSourceFilter == "All Sources" {
				message = "No installed applications found\nTry refreshing or searching with different terms"
			} else {
				message = fmt.Sprintf("No %s applications found in your installed apps\nTry searching with different terms", currentSourceFilter)
			}
		}
	} else if currentViewFilter == "Saved Apps" {
		// Get current list name for better context
		currentList := appManager.GetCurrentList()
		listName := "Unknown List"
		if currentList != nil {
			listName = currentList.Name
		}

		if isSearching {
			if currentSourceFilter == "All Sources" {
				message = fmt.Sprintf("No application found: \"%s\"\nNo saved apps match your search in list: %s", currentSearchQuery, listName)
			} else {
				message = fmt.Sprintf("No application found: \"%s\"\nNo saved %s apps match your search in list: %s", currentSearchQuery, currentSourceFilter, listName)
			}
		} else {
			if currentSourceFilter == "All Sources" {
				message = fmt.Sprintf("No saved applications found in list: %s\nTry searching with different terms or save some apps to this list", listName)
			} else {
				message = fmt.Sprintf("No %s applications found in list: %s\nTry searching with different terms or save some %s apps to this list", currentSourceFilter, listName, currentSourceFilter)
			}
		}
	} else {
		// All Results
		if isSearching {
			if currentSourceFilter == "All Sources" {
				message = fmt.Sprintf("No application found: \"%s\"\nTry a different search term", currentSearchQuery)
			} else {
				message = fmt.Sprintf("No application found: \"%s\"\nNo %s apps found with that name", currentSearchQuery, currentSourceFilter)
			}
		} else {
			if currentSourceFilter == "All Sources" {
				message = "No applications found\nTry installing some apps or searching for new ones"
			} else {
				message = fmt.Sprintf("No %s applications found\nTry searching for %s applications or install some", currentSourceFilter, currentSourceFilter)
			}
		}
	}

	messageLabel.SetText(message)
}

func showListManagementDialog(parent fyne.Window, appManager *AppManager, updateCallback func()) {
	listWindow := fyne.CurrentApp().NewWindow("Manage Lists")
	listWindow.Resize(fyne.NewSize(700, 600))
	listWindow.CenterOnScreen()

	// Declare listsList variable first
	var listsList *widget.List

	// Lists list widget
	listsList = widget.NewList(
		func() int {
			lists := appManager.GetLists()
			return len(lists)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),             // Name
				widget.NewLabel(""),             // Description
				widget.NewButton("Edit", nil),   // Edit button
				widget.NewButton("Delete", nil), // Delete button
				widget.NewButton("Export", nil), // Export button
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			lists := appManager.GetLists()
			if id >= 0 && id < len(lists) {
				list := lists[id]

				cont := obj.(*fyne.Container)
				nameLabel := cont.Objects[0].(*widget.Label)
				descLabel := cont.Objects[1].(*widget.Label)
				editBtn := cont.Objects[2].(*widget.Button)
				deleteBtn := cont.Objects[3].(*widget.Button)
				exportBtn := cont.Objects[4].(*widget.Button)

				nameLabel.SetText(list.Name)
				descLabel.SetText(list.Description)

				editBtn.OnTapped = func() {
					showEditListDialog(listWindow, appManager, list, func() {
						listsList.Refresh()
						updateCallback()
					})
				}

				deleteBtn.OnTapped = func() {
					if list.ID == 1 { // Default list
						dialog.ShowError(fmt.Errorf("Cannot delete the default list"), listWindow)
						return
					}

					dialog.ShowConfirm("Delete List",
						fmt.Sprintf("Are you sure you want to delete '%s'? All applications in this list will be removed.", list.Name),
						func(confirmed bool) {
							if confirmed {
								err := appManager.DeleteList(list.ID)
								if err != nil {
									dialog.ShowError(err, listWindow)
								} else {
									listsList.Refresh()
									updateCallback()
									dialog.ShowInformation("Deleted", fmt.Sprintf("List '%s' has been deleted.", list.Name), listWindow)
								}
							}
						}, listWindow)
				}

				exportBtn.OnTapped = func() {
					exportBtn.SetText("Exporting...")
					exportBtn.Disable()

					go func() {
						defer func() {
							if r := recover(); r != nil {
								// Handle panic gracefully
							}
							exportBtn.SetText("Export")
							exportBtn.Enable()
						}()

						err := appManager.ExportListToCSV(list.ID)
						if err != nil {
							dialog.ShowError(err, listWindow)
						} else {
							dialog.ShowInformation("Export Complete",
								fmt.Sprintf("List '%s' has been exported to CSV in the exports folder.", list.Name),
								listWindow)
						}
					}()
				}

				// Disable delete button for default list
				if list.ID == 1 {
					deleteBtn.Disable()
				} else {
					deleteBtn.Enable()
				}
			}
		},
	)

	// Create new list button
	createButton := widget.NewButtonWithIcon("Create New List", theme.ContentAddIcon(), func() {
		showCreateListDialog(listWindow, appManager, func() {
			listsList.Refresh()
			updateCallback()
		})
	})
	createButton.Importance = widget.HighImportance

	// Export all lists button
	exportAllButton := widget.NewButtonWithIcon("Export All Lists", theme.DocumentSaveIcon(), nil)
	exportAllButton.Importance = widget.MediumImportance
	exportAllButton.OnTapped = func() {
		exportAllButton.SetText("Exporting...")
		exportAllButton.Disable()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
				}
				exportAllButton.SetText("Export All Lists")
				exportAllButton.Enable()
			}()

			err := appManager.ExportAllListsToCSV()
			if err != nil {
				dialog.ShowError(err, listWindow)
			} else {
				dialog.ShowInformation("Export Complete",
					"All lists have been exported to CSV files in the exports folder.",
					listWindow)
			}
		}()
	}

	// Import button
	importButton := widget.NewButtonWithIcon("Import Lists", theme.FolderOpenIcon(), func() {
		showImportCSVDialog(listWindow, appManager, func() {
			listsList.Refresh()
			updateCallback()
		})
	})
	importButton.Importance = widget.MediumImportance

	// Close button
	closeButton := widget.NewButton("Close", func() {
		listWindow.Close()
	})

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Manage Application Lists"),
			widget.NewSeparator(),
		), // top
		container.NewHBox(createButton, exportAllButton, importButton, closeButton), // bottom
		nil,       // left
		nil,       // right
		listsList, // center
	)

	listWindow.SetContent(content)
	listWindow.Show()
}

func showCreateListDialog(parent fyne.Window, appManager *AppManager, updateCallback func()) {
	createWindow := fyne.CurrentApp().NewWindow("Create New List")
	createWindow.Resize(fyne.NewSize(400, 300))
	createWindow.CenterOnScreen()

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter list name...")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Enter description (optional)...")
	descEntry.Resize(fyne.NewSize(350, 100))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: nameEntry},
			{Text: "Description", Widget: descEntry},
		},
		OnSubmit: func() {
			name := strings.TrimSpace(nameEntry.Text)
			description := strings.TrimSpace(descEntry.Text)

			if name == "" {
				dialog.ShowError(fmt.Errorf("List name cannot be empty"), createWindow)
				return
			}

			_, err := appManager.CreateList(name, description)
			if err != nil {
				dialog.ShowError(err, createWindow)
			} else {
				updateCallback()
				dialog.ShowInformation("Created", fmt.Sprintf("List '%s' has been created.", name), createWindow)
				createWindow.Close()
			}
		},
		OnCancel: func() {
			createWindow.Close()
		},
	}

	createWindow.SetContent(form)
	createWindow.Show()
}

func showEditListDialog(parent fyne.Window, appManager *AppManager, list *AppList, updateCallback func()) {
	editWindow := fyne.CurrentApp().NewWindow("Edit List")
	editWindow.Resize(fyne.NewSize(400, 300))
	editWindow.CenterOnScreen()

	nameEntry := widget.NewEntry()
	nameEntry.SetText(list.Name)

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetText(list.Description)
	descEntry.Resize(fyne.NewSize(350, 100))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: nameEntry},
			{Text: "Description", Widget: descEntry},
		},
		OnSubmit: func() {
			name := strings.TrimSpace(nameEntry.Text)
			description := strings.TrimSpace(descEntry.Text)

			if name == "" {
				dialog.ShowError(fmt.Errorf("List name cannot be empty"), editWindow)
				return
			}

			err := appManager.UpdateList(list.ID, name, description)
			if err != nil {
				dialog.ShowError(err, editWindow)
			} else {
				updateCallback()
				dialog.ShowInformation("Updated", fmt.Sprintf("List '%s' has been updated.", name), editWindow)
				editWindow.Close()
			}
		},
		OnCancel: func() {
			editWindow.Close()
		},
	}

	// Disable name editing for default list
	if list.ID == 1 {
		nameEntry.Disable()
	}

	editWindow.SetContent(form)
	editWindow.Show()
}

func showSaveToListDialog(parent fyne.Window, appManager *AppManager, app *AppInfo, updateCallback func()) {
	dialogWindow := fyne.CurrentApp().NewWindow("Save Application to List")
	dialogWindow.Resize(fyne.NewSize(400, 300))
	dialogWindow.CenterOnScreen()

	// Create a list of available lists
	lists := appManager.GetLists()
	listOptions := make([]string, len(lists))
	for i, list := range lists {
		listOptions[i] = list.Name
	}

	var selectedListID int64 = 0
	currentList := appManager.GetCurrentList()
	if currentList != nil {
		selectedListID = currentList.ID
	}

	// Create a list widget for selecting the list
	listSelect := widget.NewSelect(listOptions, func(selected string) {
		// Handle list selection
		for _, list := range lists {
			if list.Name == selected {
				selectedListID = list.ID
				break
			}
		}
	})
	listSelect.PlaceHolder = "Select a list..."

	// Set current list as default selection
	if currentList != nil {
		listSelect.SetSelected(currentList.Name)
	}

	// Create save button with submit logic
	saveButton := widget.NewButton("Save", func() {
		if selectedListID == 0 {
			dialog.ShowError(fmt.Errorf("Please select a list"), dialogWindow)
			return
		}

		err := appManager.SaveAppToSpecificList(app, selectedListID)
		if err != nil {
			dialog.ShowError(err, dialogWindow)
		} else {
			updateCallback()
			// Find the list name for the confirmation
			listName := "Unknown"
			for _, list := range lists {
				if list.ID == selectedListID {
					listName = list.Name
					break
				}
			}
			dialog.ShowInformation("Saved", fmt.Sprintf("Application '%s' has been saved to '%s'.", app.Name, listName), dialogWindow)
			dialogWindow.Close()
		}
	})
	saveButton.Importance = widget.HighImportance

	// Create cancel button
	cancelButton := widget.NewButton("Cancel", func() {
		dialogWindow.Close()
	})

	// Create content for the dialog
	content := container.NewVBox(
		widget.NewLabel("Save Application to List"),
		widget.NewSeparator(),
		widget.NewLabel("Select a list to save this application to:"),
		listSelect,
	)

	// Create layout for the dialog
	layout := container.NewBorder(
		nil, // top
		container.NewHBox(saveButton, cancelButton), // bottom
		nil,     // left
		nil,     // right
		content, // center
	)

	dialogWindow.SetContent(layout)
	dialogWindow.Show()
}

func showAppListManagementDialog(parent fyne.Window, appManager *AppManager, app *AppInfo, updateCallback func()) {
	dialogWindow := fyne.CurrentApp().NewWindow(fmt.Sprintf("Manage '%s' in Lists", app.Name))
	dialogWindow.Resize(fyne.NewSize(500, 400))
	dialogWindow.CenterOnScreen()

	// Get all lists and which ones contain this app
	allLists := appManager.GetLists()
	listsContaining, _ := appManager.GetAppListsContaining(app.PackageID)

	// Create a map for quick lookup
	containingMap := make(map[int64]bool)
	for _, list := range listsContaining {
		containingMap[list.ID] = true
	}

	// Create checkboxes for each list
	var checkboxes []*widget.Check

	listContainer := container.NewVBox()

	for _, list := range allLists {
		isInList := containingMap[list.ID]
		checkbox := widget.NewCheck(list.Name, nil)
		checkbox.SetChecked(isInList)

		// Add description if available
		if list.Description != "" {
			listItem := container.NewVBox(
				checkbox,
				widget.NewLabelWithStyle(fmt.Sprintf("    %s", list.Description),
					fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
				widget.NewSeparator(),
			)
			listContainer.Add(listItem)
		} else {
			listContainer.Add(container.NewVBox(checkbox, widget.NewSeparator()))
		}

		checkboxes = append(checkboxes, checkbox)
	}

	// Scroll container for the list
	scrollableList := container.NewScroll(listContainer)
	scrollableList.SetMinSize(fyne.NewSize(450, 250))

	// Apply changes button
	applyButton := widget.NewButtonWithIcon("Apply Changes", theme.DocumentSaveIcon(), func() {
		// Process changes
		var added, removed []string

		for i, checkbox := range checkboxes {
			list := allLists[i]
			wasInList := containingMap[list.ID]
			isChecked := checkbox.Checked

			if !wasInList && isChecked {
				// Add to list
				err := appManager.SaveAppToSpecificList(app, list.ID)
				if err == nil {
					added = append(added, list.Name)
				}
			} else if wasInList && !isChecked {
				// Remove from list
				err := appManager.RemoveAppFromList(app.PackageID, list.ID)
				if err == nil {
					removed = append(removed, list.Name)
				}
			}
		}

		// Show confirmation message
		var message string
		if len(added) > 0 && len(removed) > 0 {
			message = fmt.Sprintf("Added to: %s\nRemoved from: %s",
				strings.Join(added, ", "), strings.Join(removed, ", "))
		} else if len(added) > 0 {
			message = fmt.Sprintf("Added to: %s", strings.Join(added, ", "))
		} else if len(removed) > 0 {
			message = fmt.Sprintf("Removed from: %s", strings.Join(removed, ", "))
		} else {
			message = "No changes made"
		}

		updateCallback()
		dialog.ShowInformation("Lists Updated", message, dialogWindow)
		dialogWindow.Close()
	})
	applyButton.Importance = widget.HighImportance

	// Close button
	closeButton := widget.NewButton("Close", func() {
		dialogWindow.Close()
	})

	// Quick action buttons
	selectAllButton := widget.NewButton("Select All", func() {
		for _, checkbox := range checkboxes {
			checkbox.SetChecked(true)
		}
	})

	selectNoneButton := widget.NewButton("Select None", func() {
		for _, checkbox := range checkboxes {
			checkbox.SetChecked(false)
		}
	})

	// Header with app info
	headerLabel := widget.NewLabelWithStyle(fmt.Sprintf("Managing: %s", app.Name),
		fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	packageLabel := widget.NewLabelWithStyle(fmt.Sprintf("Package ID: %s", app.PackageID),
		fyne.TextAlignCenter, fyne.TextStyle{Monospace: true})

	// Layout
	content := container.NewBorder(
		container.NewVBox(
			headerLabel,
			packageLabel,
			widget.NewSeparator(),
			widget.NewLabel("Select which lists should contain this application:"),
		), // top
		container.NewVBox(
			widget.NewSeparator(),
			container.NewHBox(selectAllButton, selectNoneButton),
			container.NewHBox(applyButton, closeButton),
		), // bottom
		nil,            // left
		nil,            // right
		scrollableList, // center
	)

	dialogWindow.SetContent(content)
	dialogWindow.Show()
}

func showImportCSVDialog(parent fyne.Window, appManager *AppManager, updateCallback func()) {
	importWindow := fyne.CurrentApp().NewWindow("Import Lists from CSV")
	importWindow.Resize(fyne.NewSize(600, 400))
	importWindow.CenterOnScreen()

	// Instructions
	instructions := widget.NewLabel("Select CSV files to import as lists. Each file will create or update a list.")
	instructions.Wrapping = fyne.TextWrapWord

	// File selection area
	selectedFiles := make([]string, 0)
	fileList := widget.NewList(
		func() int { return len(selectedFiles) },
		func() fyne.CanvasObject {
			return widget.NewLabel("File path")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= 0 && id < len(selectedFiles) {
				if label, ok := obj.(*widget.Label); ok {
					// Show just the filename, not full path
					filename := selectedFiles[id]
					if lastSlash := strings.LastIndex(filename, "\\"); lastSlash != -1 {
						filename = filename[lastSlash+1:]
					}
					label.SetText(filename)
				}
			}
		},
	)

	// Browse button
	browseButton := widget.NewButtonWithIcon("Browse Files", theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, importWindow)
				return
			}
			if reader != nil {
				filepath := reader.URI().Path()
				reader.Close()

				// Check if file already selected
				for _, existing := range selectedFiles {
					if existing == filepath {
						return // Already selected
					}
				}

				// Add to selected files
				selectedFiles = append(selectedFiles, filepath)
				fileList.Refresh()
			}
		}, importWindow)
	})

	// Clear selection button
	clearButton := widget.NewButton("Clear Selection", func() {
		selectedFiles = make([]string, 0)
		fileList.Refresh()
	})

	// Import button
	importButton := widget.NewButtonWithIcon("Import Selected Files", theme.DocumentIcon(), nil)
	importButton.Importance = widget.HighImportance
	importButton.OnTapped = func() {
		if len(selectedFiles) == 0 {
			dialog.ShowError(fmt.Errorf("No files selected"), importWindow)
			return
		}

		importButton.SetText("Importing...")
		importButton.Disable()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Handle panic gracefully
				}
				importButton.SetText("Import Selected Files")
				importButton.Enable()
			}()

			results, err := appManager.ImportMultipleListsFromCSV(selectedFiles)

			if err != nil {
				dialog.ShowError(err, importWindow)
				return
			}

			// Show results
			showImportResultsDialog(importWindow, results, updateCallback)
			importWindow.Close()
		}()
	}

	// Cancel button
	cancelButton := widget.NewButton("Cancel", func() {
		importWindow.Close()
	})

	// Layout
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Import Lists from CSV"),
			widget.NewSeparator(),
			instructions,
			widget.NewSeparator(),
		), // top
		container.NewHBox(browseButton, clearButton, importButton, cancelButton), // bottom
		nil,      // left
		nil,      // right
		fileList, // center
	)

	importWindow.SetContent(content)
	importWindow.Show()
}

func showImportResultsDialog(parent fyne.Window, results []ImportResult, updateCallback func()) {
	resultsWindow := fyne.CurrentApp().NewWindow("Import Results")
	resultsWindow.Resize(fyne.NewSize(700, 500))
	resultsWindow.CenterOnScreen()

	// Count successes and failures
	successCount := 0
	failureCount := 0
	totalImported := 0

	for _, result := range results {
		if result.Error == nil {
			successCount++
			totalImported += result.ImportedCount
		} else {
			failureCount++
		}
	}

	// Summary
	summary := widget.NewLabel(fmt.Sprintf("Import completed: %d successful, %d failed, %d applications imported",
		successCount, failureCount, totalImported))
	summary.TextStyle = fyne.TextStyle{Bold: true}

	// Results list
	resultsList := widget.NewList(
		func() int { return len(results) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabel("Filename"),
				widget.NewLabel("Status"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= 0 && id < len(results) {
				result := results[id]
				cont := obj.(*fyne.Container)

				// Filename
				filename := result.Filepath
				if lastSlash := strings.LastIndex(filename, "\\"); lastSlash != -1 {
					filename = filename[lastSlash+1:]
				}
				if filenameLabel, ok := cont.Objects[0].(*widget.Label); ok {
					filenameLabel.SetText(filename)
					filenameLabel.TextStyle = fyne.TextStyle{Bold: true}
				}

				// Status
				var status string
				if result.Error == nil {
					status = fmt.Sprintf("✅ Success: Imported %d apps to list '%s'", result.ImportedCount, result.ListName)
				} else {
					status = fmt.Sprintf("❌ Error: %v", result.Error)
				}

				if statusLabel, ok := cont.Objects[1].(*widget.Label); ok {
					statusLabel.SetText(status)
					statusLabel.Wrapping = fyne.TextWrapWord
				}
			}
		},
	)

	// Close button
	closeButton := widget.NewButton("Close", func() {
		resultsWindow.Close()
		updateCallback() // Refresh the UI
	})

	// Layout
	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Import Results"),
			widget.NewSeparator(),
			summary,
			widget.NewSeparator(),
		), // top
		closeButton, // bottom
		nil,         // left
		nil,         // right
		resultsList, // center
	)

	resultsWindow.SetContent(content)
	resultsWindow.Show()
}
