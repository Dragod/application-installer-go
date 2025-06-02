//go:build !console
// +build !console

package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	// Get user data directory
	userDataDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to user home directory
		userDataDir, err = os.UserHomeDir()
		if err != nil {
			// Last resort fallback
			userDataDir = "."
		}
	}

	// Create PF Installer directory in user data
	appDataDir := filepath.Join(userDataDir, "PF Installer")
	err = os.MkdirAll(appDataDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create app data directory: %v", err)
	}

	// Database path in user data directory
	dbPath := filepath.Join(appDataDir, "applications.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS lists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS saved_apps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		list_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		package_id TEXT NOT NULL,
		version TEXT,
		source TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE,
		UNIQUE(list_id, package_id)
	);
	
	CREATE INDEX IF NOT EXISTS idx_package_id ON saved_apps(package_id);
	CREATE INDEX IF NOT EXISTS idx_source ON saved_apps(source);
	CREATE INDEX IF NOT EXISTS idx_list_id ON saved_apps(list_id);
	CREATE INDEX IF NOT EXISTS idx_list_name ON lists(name);
	
	-- Create default list if it doesn't exist
	INSERT OR IGNORE INTO lists (name, description) VALUES ('Default', 'Default saved applications list');
	`

	_, err := db.Exec(query)
	return err
}

// List management functions
func CreateList(db *sql.DB, name, description string) (int64, error) {
	query := `INSERT INTO lists (name, description) VALUES (?, ?)`
	result, err := db.Exec(query, name, description)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetLists(db *sql.DB) ([]*AppList, error) {
	query := `SELECT id, name, description, created_at FROM lists ORDER BY name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []*AppList
	for rows.Next() {
		list := &AppList{}
		err := rows.Scan(&list.ID, &list.Name, &list.Description, &list.CreatedAt)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}

	return lists, rows.Err()
}

func GetListByName(db *sql.DB, name string) (*AppList, error) {
	query := `SELECT id, name, description, created_at FROM lists WHERE name = ?`

	list := &AppList{}
	err := db.QueryRow(query, name).Scan(&list.ID, &list.Name, &list.Description, &list.CreatedAt)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func GetListByID(db *sql.DB, listID int64) (*AppList, error) {
	query := `SELECT id, name, description, created_at FROM lists WHERE id = ?`

	list := &AppList{}
	err := db.QueryRow(query, listID).Scan(&list.ID, &list.Name, &list.Description, &list.CreatedAt)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func UpdateList(db *sql.DB, listID int64, name, description string) error {
	query := `UPDATE lists SET name = ?, description = ? WHERE id = ?`
	_, err := db.Exec(query, name, description, listID)
	return err
}

func DeleteList(db *sql.DB, listID int64) error {
	// Check if this is the default list (shouldn't be deleted)
	if listID == 1 {
		return fmt.Errorf("cannot delete the default list")
	}

	query := `DELETE FROM lists WHERE id = ?`
	_, err := db.Exec(query, listID)
	return err
}

// App management functions (updated for lists)
func SaveAppToList(db *sql.DB, listID int64, app *AppInfo) error {
	query := `
	INSERT OR REPLACE INTO saved_apps (list_id, name, package_id, version, source, description)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, listID, app.Name, app.PackageID, app.Version, app.Source, app.Description)
	return err
}

func GetAppsInList(db *sql.DB, listID int64) ([]*AppInfo, error) {
	query := `
	SELECT id, name, package_id, version, source, description
	FROM saved_apps
	WHERE list_id = ?
	ORDER BY name
	`

	rows, err := db.Query(query, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*AppInfo
	for rows.Next() {
		app := &AppInfo{IsSaved: true, ListID: listID}
		err := rows.Scan(&app.ID, &app.Name, &app.PackageID, &app.Version, &app.Source, &app.Description)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, rows.Err()
}

func RemoveAppFromList(db *sql.DB, listID int64, packageID string) error {
	query := `DELETE FROM saved_apps WHERE list_id = ? AND package_id = ?`
	_, err := db.Exec(query, listID, packageID)
	return err
}

func IsAppInList(db *sql.DB, listID int64, packageID string) (bool, error) {
	query := `SELECT COUNT(*) FROM saved_apps WHERE list_id = ? AND package_id = ?`
	var count int
	err := db.QueryRow(query, listID, packageID).Scan(&count)
	return count > 0, err
}

func GetAppListsContaining(db *sql.DB, packageID string) ([]*AppList, error) {
	query := `
	SELECT l.id, l.name, l.description, l.created_at
	FROM lists l
	INNER JOIN saved_apps sa ON l.id = sa.list_id
	WHERE sa.package_id = ?
	ORDER BY l.name
	`

	rows, err := db.Query(query, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []*AppList
	for rows.Next() {
		list := &AppList{}
		err := rows.Scan(&list.ID, &list.Name, &list.Description, &list.CreatedAt)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}

	return lists, rows.Err()
}

// Legacy functions for backward compatibility (use default list)
func SaveApp(db *sql.DB, app *AppInfo) error {
	return SaveAppToList(db, 1, app) // Use default list (ID = 1)
}

func GetSavedApps(db *sql.DB) ([]*AppInfo, error) {
	return GetAppsInList(db, 1) // Use default list (ID = 1)
}

func RemoveSavedApp(db *sql.DB, packageID string) error {
	return RemoveAppFromList(db, 1, packageID) // Use default list (ID = 1)
}

func IsAppSaved(db *sql.DB, packageID string) (bool, error) {
	return IsAppInList(db, 1, packageID) // Use default list (ID = 1)
}
