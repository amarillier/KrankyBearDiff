# KrankyBear Diff

A cross-platform **two-file diff** application written in Go with [Fyne](https://fyne.io/). It is aimed at developers who want a straightforward side-by-side comparison of source (or any text) files, with simple navigation, optional merges from one side to the other, and saves back to disk.

**Version** (packaging): see `FyneApp.toml` (e.g. 0.1.0).

## Design philosophy

The UI follows Fyne’s patterns: clarity, a small set of strong controls, and behavior that stays predictable across macOS, Windows, and Linux. Functionality and fixes take priority; **very large files** can feel heavy because the view is line-oriented—narrow comparisons stay snappier.

## Features

- **Side-by-side diff** — Lines are aligned so inserts, deletes, and unchanged regions match between left and right panes (line-level, not word-level).
- **Open & save** — Per-side paths, dirty tracking, save left / right / both from the menu or toolbar.
- **Recents** — Separate recent lists per pane; **File → Open Recent** and the **history** icon in each pane’s toolbar.
- **Navigation** — Jump to start/end of the diff, previous/next change (wraps), optional **sync scroll** on the left pane toolbar.
- **Display** — Monospace text; optional **line numbers**; optional **visible whitespace** (· for space, → for tab); light / dark / system theme (stored in preferences).
- **Editing** — Right-click a row for **take left into right**, **take right into left**, or **delete** a line on a side; changes stay in memory until you save.
- **Tooltips** — Main and per-pane toolbar icons use [fyne-tooltip](https://github.com/dweymouth/fyne-tooltip); hover for short descriptions.
- **System tray** — Show/hide windows, open files, preferences, theme, help, about, updates, quit (where supported).
- **Updates** — **Help → Check for Updates…** queries GitHub for the latest release.

## Requirements

| Platform | Notes |
|----------|--------|
| **macOS** | 10.13 (High Sierra) or later |
| **Windows** | Windows 10 or later |
| **Linux** | Desktop environments with X11 or Wayland (e.g. GNOME, KDE, XFCE, Cinnamon, MATE) |

The Fyne desktop driver expects **OpenGL** support on the machine (typical for normal desktop installs).

## Quick start

You need a [Go](https://go.dev/) toolchain (this repo targets a recent Go release; see `go.mod`).

```sh
go mod download
go build -o kbdiff .
./kbdiff
```

Or run without a named binary:

```sh
go run .
```

For packaging with Fyne’s tooling, see [Fyne installation](https://developer.fyne.io/started/) and your `FyneApp.toml`.

## Usage (summary)

For full detail, open **Help** inside the app (same content as the in-app help window).

- **Open files:** **File → Open Left/Right File…**, **Browse…** in a pane, drag-and-drop onto a pane, or **Open Recent** / history icon (per side).
- **Save:** **File → Save…** or the save icons under the app title when a side is dirty.
- **Preferences:** Theme, line numbers, whitespace, recent lists.

## Dependencies

Direct modules (see `go.mod`):

- [fyne.io/fyne/v2](https://fyne.io/) — GUI toolkit  
- [github.com/sergi/go-diff](https://github.com/sergi/go-diff) — diff engine  
- [github.com/dweymouth/fyne-tooltip](https://github.com/dweymouth/fyne-tooltip) — toolbar tooltips  

## Known limitations

- Line-oriented diff only (not syntax-highlighted, not word-level).
- Extremely large files may be slow; prefer comparing focused slices of a project.

## License

This project is provided as-is, free for personal, educational, and commercial use, under **GNU GPL-3.0** — see [`LICENSE`](LICENSE).

## Contributing

Issues and pull requests are welcome.

## Author

Allan Marillier

## Acknowledgments

- [Fyne](https://fyne.io/) — cross-platform GUI for Go  
- [go-diff](https://github.com/sergi/go-diff) — Myers diff and cleanup utilities  
- [fyne-tooltip](https://github.com/dweymouth/fyne-tooltip) — hover help for icon buttons  

Repository: [github.com/amarillier/KrankyBearDiff](https://github.com/amarillier/KrankyBearDiff)
