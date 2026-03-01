package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// Version is set during build via ldflags
var version = "dev"

func main() {
	// Setup file logging
	setupLogging()
	
	// Create an instance of the app structure
	app := NewApp()

	// Create menu
	AppMenu := menu.NewMenu()

	// File menu
	FileMenu := AppMenu.AddSubmenu("File")
	FileMenu.AddSeparator()
	FileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(app.ctx)
	})

	// View menu
	ViewMenu := AppMenu.AddSubmenu("View")
	ViewMenu.AddText("Dashboard", keys.CmdOrCtrl("1"), func(_ *menu.CallbackData) {
		// Emit event to frontend to navigate
		runtime.EventsEmit(app.ctx, "navigate", "dashboard")
	})
	ViewMenu.AddText("Tasks", keys.CmdOrCtrl("2"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "navigate", "tasks")
	})
	ViewMenu.AddText("Reports", keys.CmdOrCtrl("3"), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "navigate", "reports")
	})
	ViewMenu.AddText("Settings", keys.CmdOrCtrl(","), func(_ *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "navigate", "settings")
	})

	// Window menu (macOS)
	WindowMenu := AppMenu.AddSubmenu("Window")
	WindowMenu.AddText("Minimize", keys.CmdOrCtrl("m"), func(_ *menu.CallbackData) {
		runtime.WindowMinimise(app.ctx)
	})

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "JITRACK",
		Width:     1200,
		Height:    800,
		MinWidth:  900,
		MinHeight: 600,
		Menu:      AppMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 14, B: 26, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
		},
		// IMPORTANT: Set to false so the window closes properly when clicking X
		HideWindowOnClose: false,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// setupLogging redirects stdout and stderr to a log file
func setupLogging() {
	// Get logs directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	
	logsDir := filepath.Join(homeDir, "Library", "Logs", "JITRACK")
	
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return
	}
	
	// Open log file
	logFile := filepath.Join(logsDir, "jitrack.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	
	// Redirect stdout and stderr to the log file
	os.Stdout = f
	os.Stderr = f
}
