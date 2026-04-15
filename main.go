package main

import (
	"fmt"
	"time"
)

const (
	// appName    = "KrankyBear Diff"
	appVersion = "0.2.2" // see FyneApp.toml
	appAuthor  = "Allan Marillier"
)

var appName = "KrankyBear Diff"
var appCopyright = buildCopyrightNotice()

func buildCopyrightNotice() string {
	const startYear = 2026
	currentYear := time.Now().Year()
	if currentYear <= startYear {
		return "Copyright (c) Allan Marillier, 2026"
	}
	return fmt.Sprintf("Copyright (c) Allan Marillier, 2026-%d", currentYear)
}

func main() {
	runApp()
}

// "Now this is not the end. It is not even the beginning of the end. But it is, perhaps, the end of the beginning." Winston Churchill, November 10, 1942
