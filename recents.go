package main

import (
	"encoding/json"
	"strings"

	"fyne.io/fyne/v2"
)

const (
	prefRecentLeft      = "recentPathsLeft"
	prefRecentRight     = "recentPathsRight"
	prefShowLineNumbers = "showLineNumbers"
	prefShowWhitespace  = "showWhitespace"
	maxRecentFiles      = 10
)

func recentKey(side int) string {
	if side == 0 {
		return prefRecentLeft
	}
	return prefRecentRight
}

func loadRecentList(a fyne.App, key string) []string {
	raw := a.Preferences().String(key)
	if raw == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

func saveRecentList(a fyne.App, key string, paths []string) {
	if len(paths) == 0 {
		a.Preferences().SetString(key, "")
		return
	}
	b, err := json.Marshal(paths)
	if err != nil {
		return
	}
	a.Preferences().SetString(key, string(b))
}

func addRecent(a fyne.App, side int, path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	key := recentKey(side)
	list := loadRecentList(a, key)
	out := []string{path}
	for _, p := range list {
		if p == path {
			continue
		}
		out = append(out, p)
		if len(out) >= maxRecentFiles {
			break
		}
	}
	saveRecentList(a, key, out)
}

func removeRecent(a fyne.App, side int, path string) {
	key := recentKey(side)
	list := loadRecentList(a, key)
	if len(list) == 0 {
		return
	}
	out := list[:0]
	for _, p := range list {
		if p != path {
			out = append(out, p)
		}
	}
	saveRecentList(a, key, out)
}

func clearRecent(a fyne.App, side int) {
	saveRecentList(a, recentKey(side), nil)
}

func clearAllRecent(a fyne.App) {
	clearRecent(a, 0)
	clearRecent(a, 1)
}

func recentMenuLabel(path string) string {
	if path == "" {
		return ""
	}
	const max = 72
	if len(path) <= max {
		return path
	}
	return "…" + path[len(path)-(max-1):]
}
