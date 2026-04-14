package main

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var helpWindow fyne.Window

// showHelp displays help for KrankyBear Diff.
func showHelp(a fyne.App) {
	if helpWindow != nil && helpWindow.Content().Visible() {
		helpWindow.Show()
		helpWindow.RequestFocus()
		return
	}

	helpWindow = a.NewWindow(appName + " - Help")
	helpWindow.SetIcon(resourceKrankyBearHackerPng)

	icon := newBrandingDialogImage(resourceKrankyBearHackerPng)

	helpText := `KrankyBear Diff — two-file source diff

OVERVIEW
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Compare two text files side by side. Lines are aligned so inserts,
deletes, and unchanged regions line up between the left and right
panes (similar in spirit to graphical diff tools, with only two files).

OPENING FILES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• File → Open Left File… / Open Right File… / Open Recent (left or right lists)
• File → Save Left File / Save Right File / Save Both Files — write the in-memory buffer back to the
  path shown above each pane (only enabled when that side has unsaved edits and a path).
• File → Preferences… — theme, line numbers & whitespace, and clearing recent-file lists (applied when you click Save)
• Browse… in each pane’s toolbar
• Drag and drop a file onto the left or right pane (drop near the
  pane you want to load; if the pointer is between panes, the drop
  targets the side under the cursor).

VIEW
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Monospace text. By default tabs are expanded to four spaces for alignment. View → Show Whitespace
  (and Preferences): each space is shown as · and each tab character as → so you can compare spacing
  without an editor normalizing tabs.
• Highlighting: removed lines (left), added lines (right), padding
  rows keep alignment
• Line numbers (off by default): View → Line Numbers (and Preferences) shows 1-based source line numbers per side. Rows with no line on that side show “·” in the gutter — those are not blank lines in
  the file; they pad the other pane when only one side has new or deleted lines.
• View → Show All Windows / Hide All Windows (same entries on the system tray menu)
• View → Light / Dark / System theme (also in Preferences and tray; stored in preferences)
• System tray: show/hide all windows, open left/right file…, preferences, theme, help, about, check updates, quit.
  Recent paths are not on the tray menu—use File → Open Recent or the history icon in a pane’s toolbar.

NAVIGATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
- Main toolbar (under the app title): icon buttons left-to-right — save left file, save right file,
  save both (if either side is dirty); undo and redo; swap left/right files; sync-scroll toggle;
  line numbers toggle; show-whitespace toggle. Medium emphasis indicates an option is on. Hover those
  icons (and the per-pane icons below) for short descriptions (tooltips via fyne-tooltip on the main window;
  a patched vendored copy avoids spurious errors when dialogs or other overlays are on screen).
- Per-pane toolbar: double-arrow icons jump to the start or end of the diff; single arrows jump to the
  previous or next changed block (wraps at the ends). The clock/history icon opens recent files for
  that pane (same entries as File → Open Recent → Left file or Right file). Browse… opens the file picker.
  Below that: a wide Find field (press Enter for next match), “.*” for regex (Go regexp), “Aa” for match case,
  and skip-previous / skip-next icons to search only that pane’s file (wraps). A match selects the aligned
  row in both lists and scrolls both to it; when Sync scroll is on, scroll anchors update so linked scrolling stays coherent.
- With sync scroll on, scrolling either list (wheel, trackpad, or scrollbar) keeps both panes at the
  same vertical offset.
- Click a line in one list: the other pane selects and scrolls to the same aligned row (one-click sync).
- Right-click (secondary click) a line: context menu — copy left line, copy right line, or copy the
  aligned row (one or both sides, newline between them); then merge actions with row-specific labels
  (replace one side with the other on matching lines, insert a missing line, or remove an extra line).
  The first merge action is the one for the pane you clicked (left vs right). Delete removes a line on
  that side when this row has a real line there. Edits stay in memory until you use File → Save or the main save icons.

UPDATES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Help → Check for Updates… queries GitHub for the latest release (and compares versions).
On startup the app may also check in the background about once every seven days; if a newer
release exists, the same style of update window appears. Failed background checks are silent.

LIMITATIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Line-oriented diff (not word-level). Very large files may feel
  heavy; prefer focused comparisons.
• Display is text with diff highlighting, not full syntax
  highlighting for every language.

KEYBOARD
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Cmd/Ctrl Q — Quit (also File → Quit)
• Cmd/Ctrl Z / Shift Cmd/Ctrl Z — Undo / Redo merge edits
• Cmd/Ctrl C — Copy aligned row (selected line; see Edit menu)
• Shift Cmd/Ctrl X — Swap left and right files (paths and contents; clears undo history)
• Alt , / Alt . — Previous / next change (wraps; also View menu)
• Alt Home / Alt End — Jump to start / end of diff (also View menu)
• Standard window shortcuts for minimize / close

MORE INFORMATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• GitHub: https://github.com/amarillier/KrankyBearDiff
• License: https://github.com/amarillier/KrankyBearDiff/blob/allanm/LICENSE
`

	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	githubURL, _ := url.Parse("https://github.com/amarillier/KrankyBearDiff")
	githubLink := widget.NewHyperlink("Visit GitHub Repository", githubURL)
	githubLink.Alignment = fyne.TextAlignCenter

	licenseURL, _ := url.Parse("https://github.com/amarillier/KrankyBearDiff/blob/allanm/LICENSE")
	licenseLink := widget.NewHyperlink("View License", licenseURL)
	licenseLink.Alignment = fyne.TextAlignCenter

	scrollContent := container.NewScroll(helpLabel)
	scrollContent.SetMinSize(fyne.NewSize(560, 480))

	header := widget.NewLabelWithStyle(appName+" — Help", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewCenter(container.NewHBox(githubLink, licenseLink)),
	)

	readingArea := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		nil,
		nil,
		nil,
		scrollContent,
	)

	mainArea := container.NewHBox(
		container.NewPadded(icon),
		readingArea,
	)

	content := container.NewBorder(
		nil,
		footer,
		nil,
		nil,
		mainArea,
	)

	helpWindow.SetContent(container.NewPadded(content))
	helpWindow.Resize(fyne.NewSize(900, 650))

	helpWindow.SetCloseIntercept(func() {
		helpWindow.Hide()
	})

	helpWindow.Show()
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
