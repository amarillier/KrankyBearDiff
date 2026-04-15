package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

func undoMenuItem(v *diffView) *fyne.MenuItem {
	it := fyne.NewMenuItem("Undo", func() { v.undoEdit() })
	it.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: fyne.KeyModifierShortcutDefault}
	it.Disabled = len(v.undoStack) == 0
	return it
}

func redoMenuItem(v *diffView) *fyne.MenuItem {
	it := fyne.NewMenuItem("Redo", func() { v.redoEdit() })
	it.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift}
	it.Disabled = len(v.redoStack) == 0
	return it
}

func lineNumsMenuItem(v *diffView) *fyne.MenuItem {
	it := fyne.NewMenuItem("Line Numbers", func() {
		v.showLineNumbers = !v.showLineNumbers
		v.app.Preferences().SetBool(prefShowLineNumbers, v.showLineNumbers)
		v.refreshDiffLists()
		v.refreshMainToolbar()
		v.refreshMainMenu()
	})
	it.Checked = v.showLineNumbers
	return it
}

func showWhitespaceMenuItem(v *diffView) *fyne.MenuItem {
	it := fyne.NewMenuItem("Show Whitespace", func() {
		v.showWhitespace = !v.showWhitespace
		v.app.Preferences().SetBool(prefShowWhitespace, v.showWhitespace)
		v.refreshDiffLists()
		v.refreshMainToolbar()
		v.refreshMainMenu()
	})
	it.Checked = v.showWhitespace
	return it
}

func syncScrollMenuItem(v *diffView) *fyne.MenuItem {
	it := fyne.NewMenuItem("Sync scroll", func() {
		v.syncScrollOn = !v.syncScrollOn
		if v.syncScrollOn && v.leftList != nil && v.rightList != nil {
			v.syncScrollPrevL = v.leftList.GetScrollOffset()
			v.syncScrollPrevR = v.rightList.GetScrollOffset()
		}
		v.app.Preferences().SetBool(prefSyncScroll, v.syncScrollOn)
		v.refreshMainToolbar()
		v.refreshMainMenu()
	})
	it.Checked = v.syncScrollOn
	return it
}

func bringAllAppWindowsToFront(a fyne.App, mainW fyne.Window) {
	for _, win := range a.Driver().AllWindows() {
		if win == nil {
			continue
		}
		win.Show()
	}
	if mainW != nil {
		mainW.RequestFocus()
	}
}

func hideAllAppWindows(a fyne.App) {
	for _, win := range a.Driver().AllWindows() {
		if win == nil {
			continue
		}
		win.Hide()
	}
}

func (v *diffView) buildRecentSubmenu(side int) *fyne.Menu {
	paths := loadRecentList(v.app, recentKey(side))
	if len(paths) == 0 {
		empty := fyne.NewMenuItem("(no recent files)", nil)
		empty.Disabled = true
		return fyne.NewMenu("", empty)
	}
	items := make([]*fyne.MenuItem, 0, len(paths))
	for _, p := range paths {
		path := p
		items = append(items, fyne.NewMenuItem(recentMenuLabel(path), func() {
			v.loadPathFromRecent(side, path)
		}))
	}
	return fyne.NewMenu("", items...)
}

func (v *diffView) buildMainMenu() *fyne.MainMenu {
	recentLeft := fyne.NewMenuItem("Left file", nil)
	recentLeft.ChildMenu = v.buildRecentSubmenu(0)
	recentRight := fyne.NewMenuItem("Right file", nil)
	recentRight.ChildMenu = v.buildRecentSubmenu(1)
	openRecent := fyne.NewMenuItem("Open Recent", nil)
	openRecent.ChildMenu = fyne.NewMenu("", recentLeft, recentRight)

	saveLeft := fyne.NewMenuItem("Save Left File", func() { v.saveSideAttempt(0) })
	saveLeft.Disabled = v.leftP == "" || !v.leftDirty
	saveRight := fyne.NewMenuItem("Save Right File", func() { v.saveSideAttempt(1) })
	saveRight.Disabled = v.rightP == "" || !v.rightDirty
	saveBoth := fyne.NewMenuItem("Save Both Files", func() { v.saveBothAttempt() })
	saveBoth.Disabled = (v.leftP == "" || !v.leftDirty) && (v.rightP == "" || !v.rightDirty)

	file := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Left File…", func() { v.openFileDialog(0) }),
		fyne.NewMenuItem("Open Right File…", func() { v.openFileDialog(1) }),
		openRecent,
		fyne.NewMenuItemSeparator(),
		saveLeft,
		saveRight,
		saveBoth,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Preferences…", func() { showPreferences(v.app, v) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() { quitFromMainWindow(v.app, v.win) }),
	)

	copyAligned := fyne.NewMenuItem("Copy aligned row", func() { v.copySelectedRowToClipboard() })
	copyAligned.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyC, Modifier: fyne.KeyModifierShortcutDefault}
	selOK := v.hasDiffSelection && v.model != nil && v.selectedDiffRow >= 0 && v.selectedDiffRow < len(v.model.Rows)
	rowCopyOK := false
	if selOK {
		dr := v.model.Rows[v.selectedDiffRow]
		rowCopyOK = dr.LeftLineNo > 0 || dr.RightLineNo > 0
	}
	copyAligned.Disabled = !selOK || !rowCopyOK

	edit := fyne.NewMenu("Edit",
		undoMenuItem(v),
		redoMenuItem(v),
		fyne.NewMenuItemSeparator(),
		copyAligned,
	)

	noChanges := v.model == nil || len(v.model.ChangeIndices) == 0
	emptyDiff := v.model == nil || len(v.model.Rows) == 0
	prevChange := fyne.NewMenuItem("Previous change", func() { v.jumpDiff(-1) })
	prevChange.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: fyne.KeyModifierAlt}
	prevChange.Disabled = noChanges
	nextChange := fyne.NewMenuItem("Next change", func() { v.jumpDiff(1) })
	nextChange.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyPeriod, Modifier: fyne.KeyModifierAlt}
	nextChange.Disabled = noChanges
	jumpStart := fyne.NewMenuItem("Jump to start of diff", func() { v.jumpToFileStart() })
	jumpStart.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyHome, Modifier: fyne.KeyModifierAlt}
	jumpStart.Disabled = emptyDiff
	jumpEnd := fyne.NewMenuItem("Jump to end of diff", func() { v.jumpToFileEnd() })
	jumpEnd.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyEnd, Modifier: fyne.KeyModifierAlt}
	jumpEnd.Disabled = emptyDiff
	swapSides := fyne.NewMenuItem("Swap left and right", func() { v.swapSides() })
	swapSides.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyX, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift}

	view := fyne.NewMenu("View",
		fyne.NewMenuItem("Show All Windows", func() { bringAllAppWindowsToFront(v.app, v.win) }),
		fyne.NewMenuItem("Hide All Windows", func() { hideAllAppWindows(v.app) }),
		fyne.NewMenuItemSeparator(),
		prevChange,
		nextChange,
		jumpStart,
		jumpEnd,
		swapSides,
		fyne.NewMenuItemSeparator(),
		lineNumsMenuItem(v),
		showWhitespaceMenuItem(v),
		syncScrollMenuItem(v),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Light Theme", func() {
			setLightTheme(v.app)
			v.refreshDiffLists()
		}),
		fyne.NewMenuItem("Dark Theme", func() {
			setDarkTheme(v.app)
			v.refreshDiffLists()
		}),
		fyne.NewMenuItem("System Theme", func() {
			setSystemTheme(v.app)
			v.refreshDiffLists()
		}),
	)

	help := fyne.NewMenu("Help",
		fyne.NewMenuItem("Help", func() { showHelp(v.app) }),
		fyne.NewMenuItem("About", func() { showAbout(v.app) }),
		fyne.NewMenuItem("Check for Updates…", func() { checkForUpdates(v.app) }),
	)

	return fyne.NewMainMenu(file, edit, view, help)
}

// buildTrayMenu returns the system tray menu. It must stay shallow (no nested
// submenus that change after launch): each SetSystemTrayMenu call on the GLFW
// driver spawns new goroutines per item without tearing down old listeners,
// so rebuilding the tray repeatedly leads to leaks and native crashes on macOS.
func (v *diffView) buildTrayMenu() *fyne.Menu {
	trayRecentHint := fyne.NewMenuItem("Recent files: use menu bar → File → Open Recent", nil)
	trayRecentHint.Disabled = true

	return fyne.NewMenu(appName,
		fyne.NewMenuItem("Show All Windows", func() { bringAllAppWindowsToFront(v.app, v.win) }),
		fyne.NewMenuItem("Hide All Windows", func() { hideAllAppWindows(v.app) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Open Left File…", func() { v.openFileDialog(0) }),
		fyne.NewMenuItem("Open Right File…", func() { v.openFileDialog(1) }),
		trayRecentHint,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Preferences…", func() { showPreferences(v.app, v) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Light Theme", func() {
			setLightTheme(v.app)
			v.refreshDiffLists()
		}),
		fyne.NewMenuItem("Dark Theme", func() {
			setDarkTheme(v.app)
			v.refreshDiffLists()
		}),
		fyne.NewMenuItem("System Theme", func() {
			setSystemTheme(v.app)
			v.refreshDiffLists()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Help", func() { showHelp(v.app) }),
		fyne.NewMenuItem("About", func() { showAbout(v.app) }),
		fyne.NewMenuItem("Check for Updates…", func() { checkForUpdates(v.app) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() { quitFromMainWindow(v.app, v.win) }),
	)
}

// refreshMainMenu rebuilds only the window menu bar (safe to call often).
func (v *diffView) refreshMainMenu() {
	fyne.Do(func() {
		if v.win == nil {
			return
		}
		v.win.SetMainMenu(v.buildMainMenu())
	})
}

// setupMenus installs the menu bar and system tray once after the main window is shown.
func (v *diffView) setupMenus() {
	fyne.Do(func() {
		if v.win == nil {
			return
		}
		v.win.SetMainMenu(v.buildMainMenu())
		if desk, ok := v.app.(desktop.App); ok {
			desk.SetSystemTrayMenu(v.buildTrayMenu())
			desk.SetSystemTrayIcon(resourceKrankyBearHackerPng)
		}
	})
}
