package main

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
func quitFromMainWindow(v *diffView) {
	if v == nil {
		return
	}
	saveMainWindowGeometryIfEnabled(v.app, v.win)
	if v.mainSplit != nil {
		v.app.Preferences().SetFloat(prefSplitOffset, v.mainSplit.Offset)
	}
	hideAuxiliaryWindows()
	v.app.Quit()
}
