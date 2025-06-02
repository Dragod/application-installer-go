package main

import "time"

type AppInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PackageID   string `json:"package_id"`
	Version     string `json:"version"`
	Source      string `json:"source"` // "winget" or "chocolatey"
	Description string `json:"description"`
	IsInstalled bool   `json:"is_installed"`
	IsSaved     bool   `json:"is_saved"`
	ListID      int64  `json:"list_id"` // Which list this app is saved to
}

type AppList struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type ImportResult struct {
	Filepath      string `json:"filepath"`
	ListName      string `json:"list_name"`
	ImportedCount int    `json:"imported_count"`
	Error         error  `json:"error"`
}
