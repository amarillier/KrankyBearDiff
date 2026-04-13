package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

const githubReleasesAPI = "https://api.github.com/repos/amarillier/KrankyBearDiff/releases/latest"

type ghRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// checkForUpdates queries GitHub and opens the update dialog on the UI thread.
func checkForUpdates(a fyne.App) {
	go func() {
		client := &http.Client{Timeout: 12 * time.Second}
		req, err := http.NewRequest(http.MethodGet, githubReleasesAPI, nil)
		if err != nil {
			fyne.Do(func() { showUpdateDialog(a, "Could not check for updates.", false) })
			return
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", appName+"/"+appVersion)

		resp, err := client.Do(req)
		if err != nil {
			fyne.Do(func() {
				showUpdateDialog(a, "Could not reach GitHub to check for updates.\n\n"+err.Error(), false)
			})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if resp.StatusCode != http.StatusOK {
			fyne.Do(func() {
				showUpdateDialog(a, fmt.Sprintf("Update check failed (%s).", resp.Status), false)
			})
			return
		}

		var rel ghRelease
		if err := json.Unmarshal(body, &rel); err != nil {
			fyne.Do(func() { showUpdateDialog(a, "Could not read release information.", false) })
			return
		}

		remote := strings.TrimPrefix(strings.TrimSpace(rel.TagName), "v")
		local := strings.TrimSpace(appVersion)
		available := remote != "" && remote != local

		var msg string
		if available {
			msg = fmt.Sprintf("A newer release is available.\n\nYou have: %s\nLatest: %s (%s).", local, rel.TagName, rel.Name)
		} else {
			msg = fmt.Sprintf("You are up to date.\n\nCurrent version: %s\nLatest release: %s", local, rel.TagName)
		}

		fyne.Do(func() { showUpdateDialog(a, msg, available) })
	}()
}
