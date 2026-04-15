package main

import "fyne.io/fyne/v2"

// hideAuxiliaryWindows hides About, Help, and Update dialogs so they do not
// outlive the main window during shutdown.
func hideAuxiliaryWindows() {
	if aboutWindow != nil {
		aboutWindow.Hide()
	}
	if helpWindow != nil {
		helpWindow.Hide()
	}
	if updateWindow != nil {
		updateWindow.Hide()
	}
}

// quitFromMainWindow performs a full application exit: auxiliary windows,
// GLFW windows, and the system tray (via driver Quit) are torn down.
// mainWin is the primary window; when non-nil, its size may be persisted (see preferences).
func quitFromMainWindow(a fyne.App, mainWin fyne.Window) {
	saveMainWindowGeometryIfEnabled(a, mainWin)
	hideAuxiliaryWindows()
	a.Quit()
}
