// Package main provides update checking dialog
package main

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var updateWindow fyne.Window

func showUpdateDialog(a fyne.App, message string, updateAvailable bool) {
	if updateWindow != nil && updateWindow.Content().Visible() {
		updateWindow.Show()
		updateWindow.RequestFocus()
		return
	}

	updateWindow = a.NewWindow(appName + " - Update Check")
	updateWindow.SetIcon(resourceKrankyBearHackerPng)

	icon := newBrandingDialogImage(resourceKrankyBearHackerPng)

	messageLabel := widget.NewLabel(message)
	messageLabel.Wrapping = fyne.TextWrapWord
	messageLabel.Alignment = fyne.TextAlignLeading

	var textColumn *fyne.Container
	if updateAvailable {
		releaseURL, _ := url.Parse("https://github.com/amarillier/KrankyBearDiff/releases/latest")
		releaseLink := widget.NewHyperlink("Download Latest Release", releaseURL)
		releaseLink.Alignment = fyne.TextAlignLeading

		notesURL, _ := url.Parse("https://github.com/amarillier/KrankyBearDiff/blob/main/ReleaseNotes.txt")
		notesLink := widget.NewHyperlink("View Release Notes", notesURL)
		notesLink.Alignment = fyne.TextAlignLeading

		textColumn = container.NewVBox(
			messageLabel,
			widget.NewSeparator(),
			releaseLink,
			notesLink,
		)
	} else {
		textColumn = container.NewVBox(messageLabel)
	}

	// Border gives the text column the remaining width; HBox + wrapped labels
	// collapses to ~one character wide (min width) and stacks one char per line.
	mainArea := container.NewBorder(
		nil, nil,
		container.NewPadded(icon), nil,
		container.NewPadded(textColumn),
	)

	updateWindow.SetContent(container.NewPadded(mainArea))
	updateWindow.Resize(fyne.NewSize(640, 320))

	updateWindow.SetCloseIntercept(func() {
		updateWindow.Hide()
	})

	updateWindow.Show()
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
