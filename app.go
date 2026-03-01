package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wailsapp/wails/v2/pkg/menu"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"jitrack/internal/db"
	"jitrack/internal/models"
	"jitrack/internal/server"
	"jitrack/internal/timer"
)

// App struct
type App struct {
	ctx      context.Context
	db       *db.Database
	server   *server.JitrackServer
	timer    *timer.Timer
	settings Settings
}

// Settings holds app settings
type Settings struct {
	Theme string `json:"theme"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		settings: Settings{Theme: "dark"},
	}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize database
	database, err := db.NewDatabase()
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed to initialize database: %v", err)
		return
	}
	a.db = database

	// Initialize timer
	a.timer = timer.NewTimer(database)

	// Start HTTP server for Chrome extension
	srv := server.NewServer(database, func(task models.Task) {
		// Callback when task is imported from Chrome
		wailsRuntime.EventsEmit(a.ctx, "task:imported", task)
	})
	
	port, err := srv.Start()
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed to start server: %v", err)
	} else {
		wailsRuntime.LogInfof(a.ctx, "Server started on port %d", port)
	}
	a.server = srv

	// Check for running timer from previous session
	if running := a.timer.GetRunningTimer(); running != nil {
		wailsRuntime.EventsEmit(a.ctx, "timer:started", running)
	}

	// Start timer ticker
	go a.timerTicker()
}

// Shutdown is called when the app is closing
func (a *App) Shutdown(ctx context.Context) {
	// Stop running timer if any
	if a.timer != nil {
		a.timer.StopTimer()
	}
	
	// Stop HTTP server
	if a.server != nil {
		a.server.Stop()
	}
}

// GetTrayMenu returns the tray/menu bar menu
func (a *App) GetTrayMenu() *menu.Menu {
	// Create menu
	trayMenu := menu.NewMenu()
	
	// Title item (disabled)
	titleItem := trayMenu.AddText("⏱️ JITRACK", nil, nil)
	titleItem.Disabled = true
	
	trayMenu.AddSeparator()
	
	// Show JITRACK
	trayMenu.AddText("Show JITRACK", nil, func(_ *menu.CallbackData) {
		wailsRuntime.WindowShow(a.ctx)
		wailsRuntime.WindowUnminimise(a.ctx)
	})
	
	// Stop Timer
	trayMenu.AddText("⏹ Stop Timer", nil, func(_ *menu.CallbackData) {
		a.StopTimer()
	})
	
	trayMenu.AddSeparator()
	
	// Quit
	trayMenu.AddText("Quit", nil, func(_ *menu.CallbackData) {
		wailsRuntime.Quit(a.ctx)
	})
	
	return trayMenu
}

// timerTicker emits timer updates every second
func (a *App) timerTicker() {
	// Create our own ticker for UI updates
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if running := a.timer.GetRunningTimer(); running != nil {
			elapsed := running.ElapsedSeconds()
			wailsRuntime.EventsEmit(a.ctx, "timer:tick", map[string]int{
				"elapsedSeconds": elapsed,
			})
		}
	}
}

// formatDuration formats seconds to HH:MM:SS
func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

// GetTasks returns all tasks
func (a *App) GetTasks() []models.Task {
	tasks, err := a.db.GetAllTasks()
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed to get tasks: %v", err)
		return nil
	}
	return tasks
}

// GetTask returns a single task
func (a *App) GetTask(issueKey string) *models.Task {
	task, err := a.db.GetTask(issueKey)
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed to get task: %v", err)
		return nil
	}
	return task
}

// GetWorklogs returns worklogs for a task
func (a *App) GetWorklogs(issueKey string) []models.Worklog {
	worklogs, err := a.db.GetWorklogsForTask(issueKey)
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed to get worklogs: %v", err)
		return nil
	}
	return worklogs
}

// GetDailyReport returns daily report
func (a *App) GetDailyReport(date string) []models.ReportRow {
	report, err := a.db.GetDailyReport(date)
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed to get report: %v", err)
		return nil
	}
	return report
}

// StartTimer starts timer for a task
func (a *App) StartTimer(issueKey string) error {
	worklog, err := a.timer.StartTimer(issueKey)
	if err != nil {
		return err
	}
	
	wailsRuntime.EventsEmit(a.ctx, "timer:started", worklog)
	return nil
}

// StopTimer stops current timer
func (a *App) StopTimer() (*models.Worklog, error) {
	worklog, err := a.timer.StopTimer()
	if err != nil {
		return nil, err
	}
	
	wailsRuntime.EventsEmit(a.ctx, "timer:stopped", worklog)
	return worklog, nil
}

// GetRunningTimer returns currently running timer
func (a *App) GetRunningTimer() *models.Worklog {
	return a.timer.GetRunningTimer()
}

// CompleteTask marks task as completed
func (a *App) CompleteTask(issueKey string) error {
	return a.db.CompleteTask(issueKey)
}

// UncompleteTask marks task as not completed
func (a *App) UncompleteTask(issueKey string) error {
	return a.db.UncompleteTask(issueKey)
}

// DeleteTask deletes a task
func (a *App) DeleteTask(issueKey string) error {
	return a.db.DeleteTask(issueKey)
}

// GetServerPort returns HTTP server port
func (a *App) GetServerPort() int {
	if a.server == nil {
		return 0
	}
	return a.server.GetPort()
}

// GetAuthToken returns auth token for Chrome extension
func (a *App) GetAuthToken() string {
	if a.server == nil {
		return ""
	}
	return a.server.GetToken()
}

// GetTheme returns current theme
func (a *App) GetTheme() string {
	return a.settings.Theme
}

// SetTheme sets theme
func (a *App) SetTheme(theme string) {
	a.settings.Theme = theme
}

// DeveloperInfo holds developer information
type DeveloperInfo struct {
	DatabasePath string `json:"database_path"`
	LogsPath     string `json:"logs_path"`
	GoVersion    string `json:"go_version"`
	WailsVersion string `json:"wails_version"`
	AppVersion   string `json:"app_version"`
	Platform     string `json:"platform"`
}

// GetDeveloperInfo returns developer information
// GetVersion returns the application version
func (a *App) GetVersion() string {
	return version
}

func (a *App) GetDeveloperInfo() DeveloperInfo {
	// Try to get paths, use fallback if error
	dbPath, _ := db.GetDatabasePath()
	if dbPath == "" {
		dbPath = "~/Library/Application Support/JITRACK/jitrack.db"
	}
	logsPath, _ := getLogsPath()
	
	return DeveloperInfo{
		DatabasePath: dbPath,
		LogsPath:     logsPath,
		GoVersion:    "1.24",
		WailsVersion: "2.11.0",
		AppVersion:   "2.0.0",
		Platform:     wailsRuntime.Environment(a.ctx).Platform,
	}
}

// getLogsPath returns the platform-specific logs path
func getLogsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	// macOS
	return filepath.Join(homeDir, "Library", "Logs", "JITRACK"), nil
}

// OpenFolder opens the given folder in the file manager
func (a *App) OpenFolder(path string) {
	// Check if path is a file or directory
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		// It's a file, get the directory
		path = filepath.Dir(path)
	}
	
	// Ensure the folder exists
	if err := os.MkdirAll(path, 0755); err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed to create folder: %v", err)
		return
	}
	
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed to get absolute path: %v", err)
		return
	}
	
	// Encode the path for URL (handle spaces and special chars)
	// On macOS, we can use 'open' command directly
	wailsRuntime.LogInfof(a.ctx, "Opening folder: %s", absPath)
	
	// Use native command to open folder (works better than BrowserOpenURL with file://)
	var cmd *exec.Cmd
	goos := runtime.GOOS
	if goos == "darwin" {
		cmd = exec.Command("open", absPath)
	} else if goos == "windows" {
		cmd = exec.Command("explorer", absPath)
	} else {
		cmd = exec.Command("xdg-open", absPath)
	}
	
	if err := cmd.Start(); err != nil {
		wailsRuntime.LogErrorf(a.ctx, "Failed to open folder: %v", err)
	}
}
