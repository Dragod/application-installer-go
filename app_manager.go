//go:build !console
// +build !console

package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Package manager settings accessors
var wingetEnabled = true
var chocoEnabled = true

func getWingetEnabled() bool {
	return wingetEnabled
}

func getChocoEnabled() bool {
	return chocoEnabled
}

func setWingetEnabled(enabled bool) {
	wingetEnabled = enabled
}

func setChocoEnabled(enabled bool) {
	chocoEnabled = enabled
}

type AppManager struct {
	db                  *sql.DB
	wingetManager       *WingetManager
	chocoManager        *ChocolateyManager
	currentApps         []*AppInfo
	allApps             []*AppInfo // Store all unfiltered apps
	installedApps       []*AppInfo
	savedApps           []*AppInfo
	allLists            []*AppList // Store all available lists
	currentList         *AppList   // Currently selected list
	currentSourceFilter string     // Track current source filter
	currentViewFilter   string     // Track current view filter (All Results, Installed Only, Saved Apps)
	currentSearchQuery  string     // Track current search query
	isSearchMode        bool       // Track if we're showing search results
	isLoading           bool       // Track if we're currently loading/searching
	mutex               sync.RWMutex
	callbacks           []func()
}

func NewAppManager(db *sql.DB) *AppManager {
	am := &AppManager{
		db:                  db,
		wingetManager:       &WingetManager{},
		chocoManager:        &ChocolateyManager{},
		currentApps:         make([]*AppInfo, 0),
		allApps:             make([]*AppInfo, 0),
		installedApps:       make([]*AppInfo, 0),
		savedApps:           make([]*AppInfo, 0),
		allLists:            make([]*AppList, 0),
		currentSourceFilter: "All Sources",    // Default filter
		currentViewFilter:   "Installed Only", // Default to installed
		isSearchMode:        false,
		isLoading:           false,
		callbacks:           make([]func(), 0),
	}

	// Load lists and set default list on startup
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Handle panic gracefully
			}
		}()
		am.LoadLists()
		// Set default list as current
		if len(am.allLists) > 0 {
			am.SetCurrentList(am.allLists[0]) // Default list should be first
		}
	}()

	// REMOVED AUTO-LOADING TO PREVENT UI DEADLOCK
	// Auto-loading will be triggered by user action (refresh button) instead

	return am
}

func (am *AppManager) AddCallback(callback func()) {
	if callback != nil {
		am.callbacks = append(am.callbacks, callback)
	}
}

func (am *AppManager) notifyCallbacks() {
	for _, callback := range am.callbacks {
		if callback != nil {
			go func(cb func()) {
				defer func() {
					if r := recover(); r != nil {
						// Handle panic gracefully
					}
				}()
				cb()
			}(callback)
		}
	}
}

func (am *AppManager) SearchApps(query string) error {
	// Handle empty query by clearing search mode
	if strings.TrimSpace(query) == "" {
		am.ClearSearch()
		return nil
	}

	// Set loading state BEFORE acquiring mutex to avoid deadlock
	am.SetLoading(true)

	am.mutex.Lock()

	// Store the search query
	am.currentSearchQuery = query

	var searchApps []*AppInfo

	// Set search mode
	am.isSearchMode = true

	// Determine which apps to search based on current view filter
	switch am.currentViewFilter {
	case "Saved Apps":
		// Search within saved apps only
		for _, app := range am.savedApps {
			if strings.Contains(strings.ToLower(app.Name), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(app.PackageID), strings.ToLower(query)) {
				searchApps = append(searchApps, app)
			}
		}
		am.allApps = searchApps // Set search results as all apps for filtering
	case "Installed Only":
		// Search within installed apps only
		for _, app := range am.installedApps {
			if strings.Contains(strings.ToLower(app.Name), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(app.PackageID), strings.ToLower(query)) {
				searchApps = append(searchApps, app)
			}
		}
		am.allApps = searchApps // Set search results as all apps for filtering
	default: // "All Results"
		// Search with winget if available AND enabled
		if am.wingetManager.IsAvailable() && getWingetEnabled() {
			apps, err := am.wingetManager.Search(query)
			if err == nil {
				searchApps = append(searchApps, apps...)
			}
		}

		// Search with chocolatey if available AND enabled
		if am.chocoManager.IsAvailable() && getChocoEnabled() {
			apps, err := am.chocoManager.Search(query)
			if err == nil {
				searchApps = append(searchApps, apps...)
			}
		}

		// Mark saved status
		am.markSavedStatus(searchApps)
		am.allApps = searchApps // Store all apps
	}

	// Clear loading state BEFORE applying filters so UI can display results
	am.isLoading = false

	// Apply current source filter to search results
	am.applyAllFilters()

	am.mutex.Unlock()

	// Clear loading state after everything is done
	am.SetLoading(false)

	return nil
}

func (am *AppManager) RefreshInstalledApps() error {
	// Set loading state BEFORE acquiring mutex to avoid deadlock
	am.SetLoading(true)

	am.mutex.Lock()

	var allApps []*AppInfo

	// Exit search mode when refreshing
	am.isSearchMode = false

	// Get installed apps from winget if available AND enabled
	if am.wingetManager.IsAvailable() && getWingetEnabled() {
		apps, err := am.wingetManager.GetInstalledApps()
		if err == nil {
			allApps = append(allApps, apps...)
		}
	}

	// Get installed apps from chocolatey if available AND enabled
	if am.chocoManager.IsAvailable() && getChocoEnabled() {
		apps, err := am.chocoManager.GetInstalledApps()
		if err == nil {
			allApps = append(allApps, apps...)
		}
	}

	// Mark saved status
	am.markSavedStatus(allApps)

	am.installedApps = allApps
	am.allApps = allApps // Store all apps for filtering

	// Clear loading state BEFORE applying filters so UI can display results
	am.isLoading = false

	// Apply current source filter to refresh results
	am.applyAllFilters()

	am.mutex.Unlock()

	// Clear loading state after everything is done
	am.SetLoading(false)

	return nil
}

func (am *AppManager) InstallApp(app *AppInfo) error {
	var err error

	switch app.Source {
	case "winget":
		if am.wingetManager.IsAvailable() {
			err = am.wingetManager.Install(app.PackageID)
		} else {
			err = fmt.Errorf("winget is not available")
		}
	case "chocolatey":
		if am.chocoManager.IsAvailable() {
			err = am.chocoManager.Install(app.PackageID)
		} else {
			err = fmt.Errorf("chocolatey is not available")
		}
	default:
		err = fmt.Errorf("unknown package source: %s", app.Source)
	}

	if err == nil {
		// Refresh installed apps after successful installation
		go am.RefreshInstalledApps()
	}

	return err
}

// List management methods
func (am *AppManager) LoadLists() error {
	lists, err := GetLists(am.db)
	if err != nil {
		return err
	}

	am.mutex.Lock()
	am.allLists = lists
	am.mutex.Unlock()

	return nil
}

func (am *AppManager) GetLists() []*AppList {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	result := make([]*AppList, len(am.allLists))
	copy(result, am.allLists)
	return result
}

func (am *AppManager) GetCurrentList() *AppList {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	if am.currentList != nil {
		// Return a copy
		listCopy := *am.currentList
		return &listCopy
	}
	return nil
}

func (am *AppManager) SetCurrentList(list *AppList) error {
	if list == nil {
		return fmt.Errorf("list cannot be nil")
	}

	am.mutex.Lock()
	am.currentList = list
	am.mutex.Unlock()

	// Reload saved apps for the new list
	return am.LoadSavedApps()
}

func (am *AppManager) CreateList(name, description string) (*AppList, error) {
	listID, err := CreateList(am.db, name, description)
	if err != nil {
		return nil, err
	}

	// Reload lists
	am.LoadLists()

	// Find and return the newly created list
	for _, list := range am.allLists {
		if list.ID == listID {
			return list, nil
		}
	}

	return nil, fmt.Errorf("failed to find newly created list")
}

func (am *AppManager) UpdateList(listID int64, name, description string) error {
	err := UpdateList(am.db, listID, name, description)
	if err != nil {
		return err
	}

	// Reload lists
	am.LoadLists()

	// Update current list if it was the one being modified
	if am.currentList != nil && am.currentList.ID == listID {
		for _, list := range am.allLists {
			if list.ID == listID {
				am.mutex.Lock()
				am.currentList = list
				am.mutex.Unlock()
				break
			}
		}
	}

	am.notifyCallbacks()
	return nil
}

func (am *AppManager) DeleteList(listID int64) error {
	err := DeleteList(am.db, listID)
	if err != nil {
		return err
	}

	// Reload lists
	am.LoadLists()

	// If the deleted list was current, switch to default
	if am.currentList != nil && am.currentList.ID == listID {
		if len(am.allLists) > 0 {
			am.SetCurrentList(am.allLists[0]) // Switch to default list
		} else {
			am.mutex.Lock()
			am.currentList = nil
			am.savedApps = make([]*AppInfo, 0)
			am.mutex.Unlock()
		}
	}

	am.notifyCallbacks()
	return nil
}

// Enhanced app management methods
func (am *AppManager) SaveAppToCurrentList(app *AppInfo) error {
	if am.currentList == nil {
		return fmt.Errorf("no list selected")
	}

	err := SaveAppToList(am.db, am.currentList.ID, app)
	if err == nil {
		am.LoadSavedApps()

		// Update current apps to reflect saved status
		am.mutex.Lock()
		am.markSavedStatus(am.currentApps)
		am.mutex.Unlock()

		am.notifyCallbacks()
	}
	return err
}

func (am *AppManager) SaveAppToSpecificList(app *AppInfo, listID int64) error {
	err := SaveAppToList(am.db, listID, app)
	if err == nil {
		// If it was saved to current list, reload saved apps
		if am.currentList != nil && am.currentList.ID == listID {
			am.LoadSavedApps()

			// Update current apps to reflect saved status
			am.mutex.Lock()
			am.markSavedStatus(am.currentApps)
			am.mutex.Unlock()
		}

		am.notifyCallbacks()
	}
	return err
}

func (am *AppManager) RemoveAppFromCurrentList(packageID string) error {
	if am.currentList == nil {
		return fmt.Errorf("no list selected")
	}

	err := RemoveAppFromList(am.db, am.currentList.ID, packageID)
	if err == nil {
		am.LoadSavedApps()

		// Update current apps to reflect saved status
		am.mutex.Lock()
		am.markSavedStatus(am.currentApps)
		am.mutex.Unlock()

		am.notifyCallbacks()
	}
	return err
}

func (am *AppManager) RemoveAppFromList(packageID string, listID int64) error {
	err := RemoveAppFromList(am.db, listID, packageID)
	if err == nil {
		// If we removed from the current list, reload saved apps
		if am.currentList != nil && am.currentList.ID == listID {
			am.LoadSavedApps()
		}

		// Update current apps to reflect saved status
		am.mutex.Lock()
		am.markSavedStatus(am.currentApps)
		am.mutex.Unlock()

		am.notifyCallbacks()
	}
	return err
}

func (am *AppManager) GetAppListsContaining(packageID string) ([]*AppList, error) {
	return GetAppListsContaining(am.db, packageID)
}

func (am *AppManager) InstallAllAppsInList(listID int64) error {
	apps, err := GetAppsInList(am.db, listID)
	if err != nil {
		return err
	}

	for _, app := range apps {
		err := am.InstallApp(app)
		if err != nil {
			return fmt.Errorf("failed to install %s: %v", app.Name, err)
		}
	}

	return nil
}

func (am *AppManager) InstallAllAppsInCurrentList() error {
	if am.currentList == nil {
		return fmt.Errorf("no list selected")
	}
	return am.InstallAllAppsInList(am.currentList.ID)
}

// Modified existing methods to work with current list
func (am *AppManager) LoadSavedApps() error {
	if am.currentList == nil {
		am.mutex.Lock()
		am.savedApps = make([]*AppInfo, 0)
		am.mutex.Unlock()
		return nil
	}

	apps, err := GetAppsInList(am.db, am.currentList.ID)
	if err != nil {
		return err
	}

	am.mutex.Lock()
	am.savedApps = apps
	am.mutex.Unlock()

	return nil
}

func (am *AppManager) markSavedStatus(apps []*AppInfo) {
	if am.currentList == nil {
		return
	}

	savedMap := make(map[string]bool)
	for _, saved := range am.savedApps {
		savedMap[saved.PackageID] = true
	}

	for _, app := range apps {
		app.IsSaved = savedMap[app.PackageID]
		if app.IsSaved {
			app.ListID = am.currentList.ID
		}
	}
}

// Legacy methods for backward compatibility
func (am *AppManager) SaveApp(app *AppInfo) error {
	return am.SaveAppToCurrentList(app)
}

func (am *AppManager) RemoveSavedApp(packageID string) error {
	return am.RemoveAppFromCurrentList(packageID)
}

func (am *AppManager) InstallAllSavedApps() error {
	return am.InstallAllAppsInCurrentList()
}

func (am *AppManager) GetCurrentApps() []*AppInfo {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	result := make([]*AppInfo, len(am.currentApps))
	copy(result, am.currentApps)
	return result
}

func (am *AppManager) SetCurrentApps(apps []*AppInfo) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.currentApps = make([]*AppInfo, len(apps))
	copy(am.currentApps, apps)
	am.markSavedStatus(am.currentApps)
	am.notifyCallbacks()
}

func (am *AppManager) GetSavedApps() []*AppInfo {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	result := make([]*AppInfo, len(am.savedApps))
	copy(result, am.savedApps)
	return result
}

func (am *AppManager) GetInstalledApps() []*AppInfo {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	result := make([]*AppInfo, len(am.installedApps))
	copy(result, am.installedApps)
	return result
}

func (am *AppManager) SetViewFilter(viewFilter string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Exit search mode when changing view filter
	am.isSearchMode = false

	am.currentViewFilter = viewFilter
	am.applyAllFilters()
}

func (am *AppManager) FilterBySource(source string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Store the current filter state
	am.currentSourceFilter = source
	am.applyAllFilters()
}

func (am *AppManager) applyAllFilters() {
	// This method assumes the mutex is already locked
	var baseApps []*AppInfo

	// If we're in search mode, always use allApps regardless of view filter
	if am.isSearchMode {
		baseApps = make([]*AppInfo, len(am.allApps))
		copy(baseApps, am.allApps)
	} else {
		// Normal browsing mode - determine which base set of apps to use based on view filter
		switch am.currentViewFilter {
		case "Installed Only":
			baseApps = make([]*AppInfo, len(am.installedApps))
			copy(baseApps, am.installedApps)
		case "Saved Apps":
			baseApps = make([]*AppInfo, len(am.savedApps))
			copy(baseApps, am.savedApps)
		default: // "All Results"
			// In "All Results" view without search, show combined installed + saved apps
			combinedApps := make([]*AppInfo, 0)

			// Add all installed apps
			combinedApps = append(combinedApps, am.installedApps...)

			// Add saved apps that aren't already installed (avoid duplicates)
			installedMap := make(map[string]bool)
			for _, installed := range am.installedApps {
				installedMap[installed.PackageID] = true
			}

			for _, saved := range am.savedApps {
				if !installedMap[saved.PackageID] {
					combinedApps = append(combinedApps, saved)
				}
			}

			baseApps = combinedApps
		}
	}

	// Then apply source filter to the base set
	if am.currentSourceFilter == "All Sources" {
		am.currentApps = baseApps
	} else {
		var filteredApps []*AppInfo
		for _, app := range baseApps {
			if strings.ToLower(app.Source) == strings.ToLower(am.currentSourceFilter) {
				filteredApps = append(filteredApps, app)
			}
		}
		am.currentApps = filteredApps
	}

	am.notifyCallbacks()
}

func (am *AppManager) applyCurrentSourceFilter() {
	// This method is now deprecated in favor of applyAllFilters
	// But keeping for compatibility - just call the unified filter
	am.applyAllFilters()
}

func (am *AppManager) GetCurrentSourceFilter() string {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.currentSourceFilter
}

func (am *AppManager) GetCurrentViewFilter() string {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.currentViewFilter
}

func (am *AppManager) GetCurrentSearchQuery() string {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.currentSearchQuery
}

func (am *AppManager) IsSearchMode() bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.isSearchMode
}

func (am *AppManager) IsLoading() bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.isLoading
}

func (am *AppManager) SetLoading(loading bool) {
	am.mutex.Lock()
	am.isLoading = loading
	am.mutex.Unlock()
	am.notifyCallbacks()
}

func (am *AppManager) ClearSearch() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Exit search mode and clear search query
	am.isSearchMode = false
	am.currentSearchQuery = ""

	// Reapply filters to show normal view
	am.applyAllFilters()
}

func (am *AppManager) ExportListToCSV(listID int64) error {
	list, err := GetListByID(am.db, listID)
	if err != nil {
		return err
	}

	// Get user data directory for exports
	userDataDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to user home directory
		userDataDir, err = os.UserHomeDir()
		if err != nil {
			// Last resort fallback
			userDataDir = "."
		}
	}

	// Create exports directory in user data
	exportsDir := filepath.Join(userDataDir, "PF Installer", "exports")
	err = os.MkdirAll(exportsDir, 0755)
	if err != nil {
		return err
	}

	// Clean filename by replacing spaces and special characters
	cleanName := strings.ReplaceAll(list.Name, " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "/", "_")
	cleanName = strings.ReplaceAll(cleanName, "\\", "_")

	filename := fmt.Sprintf("%s_%s.csv", cleanName, time.Now().Format("2006-01-02_15-04-05"))
	filepath := filepath.Join(exportsDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	err = writer.Write([]string{
		"Name",
		"Package ID",
		"Version",
		"Source",
		"Description",
		"Is Installed",
		"Is Saved",
		"List ID",
	})
	if err != nil {
		return err
	}

	apps, err := GetAppsInList(am.db, listID)
	if err != nil {
		return err
	}

	for _, app := range apps {
		err := writer.Write([]string{
			app.Name,
			app.PackageID,
			app.Version,
			app.Source,
			app.Description,
			strconv.FormatBool(app.IsInstalled),
			strconv.FormatBool(app.IsSaved),
			strconv.FormatInt(app.ListID, 10),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (am *AppManager) ExportCurrentListToCSV() error {
	if am.currentList == nil {
		return fmt.Errorf("no list selected")
	}
	return am.ExportListToCSV(am.currentList.ID)
}

func (am *AppManager) ExportAllListsToCSV() error {
	lists := am.GetLists()

	for _, list := range lists {
		err := am.ExportListToCSV(list.ID)
		if err != nil {
			return fmt.Errorf("failed to export list '%s': %v", list.Name, err)
		}
	}

	return nil
}

func (am *AppManager) ImportListFromCSV(filepath string) (*AppList, int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read CSV header: %v", err)
	}

	// Validate header format
	if len(header) < 5 { // At minimum we need Name, Package ID, Version, Source, Description
		return nil, 0, fmt.Errorf("invalid CSV format: expected at least 5 columns, got %d", len(header))
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read CSV records: %v", err)
	}

	if len(records) == 0 {
		return nil, 0, fmt.Errorf("CSV file is empty")
	}

	// Extract list name from filename or use first app's data
	var listName string
	var listDescription string

	// Try to extract list name from filename
	filename := filepath[strings.LastIndex(filepath, "\\")+1:]
	filename = filename[:strings.LastIndex(filename, ".")]
	if strings.Contains(filename, "_") {
		// Remove timestamp if present (format: ListName_YYYY-MM-DD_HH-MM-SS)
		parts := strings.Split(filename, "_")
		if len(parts) >= 2 {
			// Check if last parts look like a timestamp
			if len(parts) >= 4 && len(parts[len(parts)-3]) == 10 { // Date part
				listName = strings.Join(parts[:len(parts)-3], "_")
			} else {
				listName = strings.Join(parts[:len(parts)-1], "_")
			}
		} else {
			listName = filename
		}
	} else {
		listName = filename
	}

	// Clean list name
	listName = strings.ReplaceAll(listName, "_", " ")
	if listName == "" {
		listName = "Imported List"
	}
	listDescription = fmt.Sprintf("Imported from %s", filename)

	// Check if list already exists
	var targetList *AppList
	existingLists := am.GetLists()
	for _, list := range existingLists {
		if strings.EqualFold(list.Name, listName) {
			targetList = list
			break
		}
	}

	// Create list if it doesn't exist
	if targetList == nil {
		newList, err := am.CreateList(listName, listDescription)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create list '%s': %v", listName, err)
		}
		targetList = newList
	}

	// Import applications
	importedCount := 0
	for i, record := range records {
		if len(record) < 5 {
			continue // Skip invalid records
		}

		app := &AppInfo{
			Name:        strings.TrimSpace(record[0]),
			PackageID:   strings.TrimSpace(record[1]),
			Version:     strings.TrimSpace(record[2]),
			Source:      strings.TrimSpace(record[3]),
			Description: strings.TrimSpace(record[4]),
			IsSaved:     true,
			ListID:      targetList.ID,
		}

		// Skip empty records
		if app.Name == "" || app.PackageID == "" {
			continue
		}

		// Check if app already exists in this list
		exists, err := IsAppInList(am.db, targetList.ID, app.PackageID)
		if err != nil {
			return targetList, importedCount, fmt.Errorf("error checking if app exists (row %d): %v", i+2, err)
		}

		if !exists {
			err = SaveAppToList(am.db, targetList.ID, app)
			if err != nil {
				return targetList, importedCount, fmt.Errorf("failed to save app '%s' (row %d): %v", app.Name, i+2, err)
			}
			importedCount++
		}
	}

	// Refresh saved apps if this is the current list
	if am.currentList != nil && am.currentList.ID == targetList.ID {
		am.LoadSavedApps()
		am.notifyCallbacks()
	}

	return targetList, importedCount, nil
}

func (am *AppManager) ImportMultipleListsFromCSV(filepaths []string) ([]ImportResult, error) {
	results := make([]ImportResult, 0, len(filepaths))

	for _, filepath := range filepaths {
		list, count, err := am.ImportListFromCSV(filepath)
		result := ImportResult{
			Filepath:      filepath,
			ListName:      "",
			ImportedCount: count,
			Error:         err,
		}

		if list != nil {
			result.ListName = list.Name
		}

		results = append(results, result)
	}

	// Reload lists to reflect any new lists
	am.LoadLists()
	am.notifyCallbacks()

	return results, nil
}
