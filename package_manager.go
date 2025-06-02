package main

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type PackageManager interface {
	Search(query string) ([]*AppInfo, error)
	Install(packageID string) error
	GetInstalledApps() ([]*AppInfo, error)
	IsAvailable() bool
}

type WingetManager struct{}
type ChocolateyManager struct{}

const commandTimeout = 30 * time.Second

// Windows constants for hiding console windows
const (
	CREATE_NO_WINDOW = 0x08000000
)

// hideConsoleWindow configures the command to not show console windows
func hideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NO_WINDOW,
	}
}

func (w *WingetManager) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "winget", "--version")
	hideConsoleWindow(cmd)
	return cmd.Run() == nil
}

func (w *WingetManager) Search(query string) ([]*AppInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "winget", "search", query, "--accept-source-agreements")
	hideConsoleWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("winget search failed: %v", err)
	}

	return parseWingetSearchOutput(string(output))
}

func (w *WingetManager) Install(packageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "winget", "install", packageID, "--accept-source-agreements", "--accept-package-agreements")
	hideConsoleWindow(cmd)
	return cmd.Run()
}

func (w *WingetManager) GetInstalledApps() ([]*AppInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "winget", "list")
	hideConsoleWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("winget list failed: %v", err)
	}

	return parseWingetListOutput(string(output))
}

func (c *ChocolateyManager) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "choco", "--version")
	hideConsoleWindow(cmd)
	return cmd.Run() == nil
}

func (c *ChocolateyManager) Search(query string) ([]*AppInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "choco", "search", query, "--limit-output")
	hideConsoleWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("chocolatey search failed: %v", err)
	}

	return parseChocoSearchOutput(string(output))
}

func (c *ChocolateyManager) Install(packageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "choco", "install", packageID, "-y")
	hideConsoleWindow(cmd)
	return cmd.Run()
}

func (c *ChocolateyManager) GetInstalledApps() ([]*AppInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "choco", "list", "--local-only", "--limit-output")
	hideConsoleWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("chocolatey list failed: %v", err)
	}

	return parseChocoListOutput(string(output))
}

func parseWingetSearchOutput(output string) ([]*AppInfo, error) {
	lines := strings.Split(output, "\n")
	var apps []*AppInfo

	// Skip header lines
	dataStarted := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for separator line that indicates start of data
		if strings.Contains(line, "---") {
			dataStarted = true
			continue
		}

		if !dataStarted || strings.HasPrefix(line, "Name") {
			continue
		}

		// Parse each line with regex to handle varying spacing
		parts := regexp.MustCompile(`\s{2,}`).Split(line, -1)
		if len(parts) >= 2 {
			app := &AppInfo{
				Name:      strings.TrimSpace(parts[0]),
				PackageID: strings.TrimSpace(parts[1]),
				Source:    "winget",
			}
			if len(parts) >= 3 {
				app.Version = strings.TrimSpace(parts[2])
			}
			apps = append(apps, app)
		}
	}

	return apps, nil
}

func parseWingetListOutput(output string) ([]*AppInfo, error) {
	apps, err := parseWingetSearchOutput(output)
	if err != nil {
		return nil, err
	}

	// Mark all as installed
	for _, app := range apps {
		app.IsInstalled = true
	}

	return apps, nil
}

func parseChocoSearchOutput(output string) ([]*AppInfo, error) {
	lines := strings.Split(output, "\n")
	var apps []*AppInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			app := &AppInfo{
				Name:      strings.TrimSpace(parts[0]),
				PackageID: strings.TrimSpace(parts[0]),
				Version:   strings.TrimSpace(parts[1]),
				Source:    "chocolatey",
			}
			apps = append(apps, app)
		}
	}

	return apps, nil
}

func parseChocoListOutput(output string) ([]*AppInfo, error) {
	apps, err := parseChocoSearchOutput(output)
	if err != nil {
		return nil, err
	}

	// Mark all as installed
	for _, app := range apps {
		app.IsInstalled = true
	}

	return apps, nil
}
