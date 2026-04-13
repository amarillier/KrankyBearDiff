# KrankyBear Diff

A cross-platform **two-file diff** application written in Go with [Fyne](https://fyne.io/). It is aimed at developers who want a straightforward side-by-side comparison of source (or any text) files, with simple navigation, optional merges from one side to the other, and saves back to disk.

**Version** (packaging): see `FyneApp.toml` (e.g. 0.1.0).

## Design philosophy

The UI follows Fyne’s patterns: clarity, a small set of strong controls, and behavior that stays predictable across macOS, Windows, and Linux. Functionality and fixes take priority; **very large files** can feel heavy because the view is line-oriented—narrow comparisons stay snappier.

## Features

- **Side-by-side diff** — Lines are aligned so inserts, deletes, and unchanged regions match between left and right panes (line-level, not word-level).
- **Open & save** — Per-side paths, dirty tracking, save left / right / both from the menu or toolbar.
- **Recents** — Separate recent lists per pane; **File → Open Recent** and the **history** icon in each pane’s toolbar.
- **Find in pane** — Each pane has a wide **Find…** field (comfortable for long literals or regex), optional **regex** and **match case**, plus previous/next match (wraps). Matches jump the **same aligned row** in both lists; with **sync scroll** on, scroll offsets are re-synced after the jump.
- **Navigation** — Per-pane icons: jump to start/end of the diff, previous/next change (wraps). Main toolbar: **sync scroll** (linked panes while scrolling) and **swap sides** (exchange paths and buffers; clears undo history). **View** menu mirrors jump/swap with keyboard shortcuts.
- **Copy** — Right-click a row: **copy left line**, **copy right line**, or **copy aligned row** (one or both sides). With a row selected, **Cmd/Ctrl+C** (or **Edit → Copy aligned row**) copies the aligned row.
- **Undo / redo** — Up to **five** merge-edit steps (take line across, delete line); main toolbar, **Edit** menu, **Cmd/Ctrl+Z** and **Shift+Cmd/Ctrl+Z**. Opening or swapping files clears the history.
- **Display** — Monospace text; optional **line numbers**; optional **visible whitespace** (· for space, → for tab); light / dark / system theme (stored in preferences). The selected row is highlighted so merge actions are obvious.
- **Editing** — Right-click a row for copy actions, then merge actions labeled for the row type (**replace** line on one side with the other, **insert** a line from the other file, or **remove** a line). The pane you click lists its most relevant action first. **Delete** removes a line on that side when present; changes stay in memory until you save.
- **Tooltips** — Main and per-pane toolbar icons use [fyne-tooltip](https://github.com/dweymouth/fyne-tooltip) (vendored with a small overlay fallback; see **Dependencies**); hover for short descriptions even when a dialog or other overlay is open.
- **System tray** — Show/hide windows, open files, preferences, theme, help, about, updates, quit (where supported).
- **Updates** — **Help → Check for Updates…** compares your version to GitHub’s latest release (semver-style); optional **Nerd** bear artwork when your build is ahead of the published tag.

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

For full detail, open **Help** inside the app (in-app help includes a **KEYBOARD** section).

- **Open files:** **File → Open Left/Right File…**, **Browse…** in a pane, drag-and-drop onto a pane, or **Open Recent** / history icon (per side).
- **Save:** **File → Save…** or the save icons under the app title when a side is dirty.
- **Preferences:** Theme, line numbers, whitespace, recent lists.
- **Keyboard (highlights):** **Cmd/Ctrl+Z** / **Shift+Cmd/Ctrl+Z** — undo / redo merge edits. **Cmd/Ctrl+C** — copy aligned row. **Shift+Cmd/Ctrl+X** — swap left and right. `Alt+,` / `Alt+.` — previous / next change. **Alt+Home** / **Alt+End** — jump to start / end of diff. **Cmd/Ctrl+Q** — quit.

## Dependencies

Direct modules (see `go.mod`):

- [fyne.io/fyne/v2](https://fyne.io/) — GUI toolkit  
- [github.com/sergi/go-diff](https://github.com/sergi/go-diff) — diff engine  
- [github.com/dweymouth/fyne-tooltip](https://github.com/dweymouth/fyne-tooltip) — toolbar tooltips (vendored under `third_party/fyne-tooltip`; see `go.mod` `replace`)

### Why we vendor and patch fyne-tooltip

The library attaches tooltip rendering to the **window canvas** and to **sub-layers** it creates for `*widget.PopUp` menus (for example merge and recent-file flyouts). When you hover a control, it looks up the tooltip layer for whatever overlay is **on top** of the canvas.

**Problem:** Fyne’s **dialogs** (and some other full-window overlays) are not those PopUps. Their top overlay therefore has **no** registered tooltip sub-layer. In the upstream package, that case was treated as an error: it logged *“no tool tip layer for current overlay”* and **did not show** the tooltip.

**Change:** In `third_party/fyne-tooltip/internal/tooltip_layer.go`, when the top overlay is not a PopUp with a sub-layer, we **fall back to the root window tooltip layer** instead of failing. Tooltips keep working after a dialog opens or closes, and the log is not flooded. The rest of the module matches the vendored upstream version; revisit this if upstream adopts the same behavior—then the `replace` can be dropped.

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
