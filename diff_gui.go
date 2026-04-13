package main

import (
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fynetooltip "github.com/dweymouth/fyne-tooltip"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"
)

const appID = "com.github.amarillier.KrankyBearDiff"

// diffLineRow wraps a list cell so we can handle primary tap (selection) and
// secondary tap (merge menu) without blocking the list's selection behavior.
type diffLineRow struct {
	widget.BaseWidget
	inner fyne.CanvasObject
	view  *diffView
	side  int
	rowID widget.ListItemID
}

func newDiffLineRow(view *diffView, side int, inner fyne.CanvasObject) *diffLineRow {
	r := &diffLineRow{view: view, side: side, inner: inner}
	r.ExtendBaseWidget(r)
	return r
}

func (r *diffLineRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.inner)
}

func (r *diffLineRow) Tapped(ev *fyne.PointEvent) {
	if r.view == nil {
		return
	}
	var lst *widget.List
	if r.side == 0 {
		lst = r.view.leftList
	} else {
		lst = r.view.rightList
	}
	if lst == nil {
		return
	}
	if c := fyne.CurrentApp().Driver().CanvasForObject(lst); c != nil {
		if f, ok := any(lst).(fyne.Focusable); ok {
			c.Focus(f)
		}
	}
	lst.Select(r.rowID)
}

func (r *diffLineRow) TappedSecondary(ev *fyne.PointEvent) {
	if r.view != nil {
		r.view.showMergeMenu(r.rowID, ev.AbsolutePosition)
	}
}

type diffView struct {
	app    fyne.App
	win    fyne.Window
	model  *DiffModel
	leftP  string
	rightP string
	leftT  string
	rightT string

	leftTitle  *widget.Label
	rightTitle *widget.Label
	leftList   *widget.List
	rightList  *widget.List
	leftCol    fyne.CanvasObject
	rightCol   fyne.CanvasObject

	syncSel    bool
	changeSlot int

	syncScrollOn       bool
	syncScrollApplying bool
	syncScrollPrevL    float32
	syncScrollPrevR    float32

	showLineNumbers bool
	showWhitespace  bool

	leftDirty  bool
	rightDirty bool

	btnSaveLeft, btnSaveRight, btnSaveBoth *ttwidget.Button
	btnLineNums, btnWhitespace             *ttwidget.Button
}

func rgbaAlpha(c color.Color, alpha uint8) color.NRGBA {
	r16, g16, b16, _ := c.RGBA()
	return color.NRGBA{R: uint8(r16 >> 8), G: uint8(g16 >> 8), B: uint8(b16 >> 8), A: alpha}
}

func (v *diffView) tagBackground(tag LineTag) color.Color {
	switch tag {
	case LineAdded:
		return rgbaAlpha(theme.Color(theme.ColorNameSuccess), 0x44)
	case LineRemoved:
		return rgbaAlpha(theme.Color(theme.ColorNameError), 0x44)
	case LinePadding:
		return rgbaAlpha(theme.Color(theme.ColorNameDisabled), 0x22)
	default:
		return color.Transparent
	}
}

func (v *diffView) recompute() {
	v.model = BuildDiffModel(v.leftT, v.rightT)
	if v.leftList != nil {
		v.leftList.Refresh()
		v.leftList.UnselectAll()
	}
	if v.rightList != nil {
		v.rightList.Refresh()
		v.rightList.UnselectAll()
	}
	v.changeSlot = 0
	v.refreshTitles()
	if v.syncScrollOn && v.leftList != nil && v.rightList != nil {
		v.syncScrollPrevL = v.leftList.GetScrollOffset()
		v.syncScrollPrevR = v.rightList.GetScrollOffset()
	}
}

func (v *diffView) refreshTitles() {
	lp, rp := "(none)", "(none)"
	if v.leftP != "" {
		lp = v.leftP
	}
	if v.rightP != "" {
		rp = v.rightP
	}
	v.leftTitle.SetText(lp)
	v.rightTitle.SetText(rp)
}

// diffLineCellMinHeight is the list row height: glyph box for monospace text plus padding.
func diffLineCellMinHeight() float32 {
	probe := canvas.NewText("Mg", color.Black)
	probe.TextStyle = fyne.TextStyle{Monospace: true}
	probe.TextSize = theme.TextSize()
	return probe.MinSize().Height + theme.InnerPadding()*2 + theme.LineSpacing()
}

func (v *diffView) makeLineCell(side int) func() fyne.CanvasObject {
	return func() fyne.CanvasObject {
		bg := canvas.NewRectangle(color.Transparent)
		bg.SetMinSize(fyne.NewSize(32, diffLineCellMinHeight()))
		num := widget.NewLabel("")
		num.TextStyle = fyne.TextStyle{Monospace: true}
		num.Alignment = fyne.TextAlignTrailing
		num.Truncation = fyne.TextTruncateOff
		num.Wrapping = fyne.TextWrapOff
		txt := widget.NewLabel("")
		txt.TextStyle = fyne.TextStyle{Monospace: true}
		txt.Truncation = fyne.TextTruncateEllipsis
		txt.Wrapping = fyne.TextWrapOff
		// Border: fixed-width gutter on the left, text fills the rest (HBox shared space and truncated labels).
		row := container.NewBorder(nil, nil, num, nil, txt)
		inner := container.NewMax(bg, row)
		return newDiffLineRow(v, side, inner)
	}
}

func (v *diffView) updateLineCell(side int) func(id widget.ListItemID, o fyne.CanvasObject) {
	return func(id widget.ListItemID, o fyne.CanvasObject) {
		if v.model == nil || id < 0 || id >= len(v.model.Rows) {
			return
		}
		wrap := o.(*diffLineRow)
		wrap.rowID = id
		dr := v.model.Rows[id]
		max := wrap.inner.(*fyne.Container)
		bg := max.Objects[0].(*canvas.Rectangle)
		line := max.Objects[1].(*fyne.Container)
		// NewBorder(nil,nil,num,nil,txt) stores objects as [txt, num].
		txt := line.Objects[0].(*widget.Label)
		num := line.Objects[1].(*widget.Label)

		if v.showLineNumbers {
			num.Show()
			if side == 0 {
				if dr.LeftLineNo > 0 {
					num.SetText(fmt.Sprintf("%6d ", dr.LeftLineNo))
				} else {
					num.SetText(fmt.Sprintf("%6s ", "·"))
				}
			} else {
				if dr.RightLineNo > 0 {
					num.SetText(fmt.Sprintf("%6d ", dr.RightLineNo))
				} else {
					num.SetText(fmt.Sprintf("%6s ", "·"))
				}
			}
		} else {
			num.Hide()
		}

		var text string
		var tag LineTag
		if side == 0 {
			text = dr.Left
			tag = dr.LeftTag
		} else {
			text = dr.Right
			tag = dr.RightTag
		}
		txt.SetText(formatLineForDisplay(text, v.showWhitespace))
		if tag == LineEqual {
			bg.FillColor = theme.Color(theme.ColorNameBackground)
		} else {
			bg.FillColor = v.tagBackground(tag)
		}
		bg.Refresh()
	}
}

func (v *diffView) syncScrollAfterEdit() {
	if v.syncScrollOn && v.leftList != nil && v.rightList != nil {
		v.syncScrollPrevL = v.leftList.GetScrollOffset()
		v.syncScrollPrevR = v.rightList.GetScrollOffset()
	}
}

func (v *diffView) refreshMainToolbar() {
	if v.btnSaveLeft == nil {
		return
	}
	fyne.Do(func() {
		if v.leftP != "" && v.leftDirty {
			v.btnSaveLeft.Enable()
		} else {
			v.btnSaveLeft.Disable()
		}
		if v.rightP != "" && v.rightDirty {
			v.btnSaveRight.Enable()
		} else {
			v.btnSaveRight.Disable()
		}
		if (v.leftP != "" && v.leftDirty) || (v.rightP != "" && v.rightDirty) {
			v.btnSaveBoth.Enable()
		} else {
			v.btnSaveBoth.Disable()
		}
		if v.showLineNumbers {
			v.btnLineNums.Importance = widget.MediumImportance
		} else {
			v.btnLineNums.Importance = widget.LowImportance
		}
		v.btnLineNums.Refresh()
		if v.showWhitespace {
			v.btnWhitespace.Importance = widget.MediumImportance
		} else {
			v.btnWhitespace.Importance = widget.LowImportance
		}
		v.btnWhitespace.Refresh()
	})
}

func (v *diffView) buildMainChromeToolbar() fyne.CanvasObject {
	v.btnSaveLeft = ttwidget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() { v.saveSideAttempt(0) })
	v.btnSaveLeft.SetToolTip("Save left file (only when it has unsaved changes)")
	v.btnSaveRight = ttwidget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() { v.saveSideAttempt(1) })
	v.btnSaveRight.SetToolTip("Save right file (only when it has unsaved changes)")
	v.btnSaveBoth = ttwidget.NewButtonWithIcon("", theme.ConfirmIcon(), func() { v.saveBothAttempt() })
	v.btnSaveBoth.SetToolTip("Save both files (only sides that are dirty)")
	v.btnLineNums = ttwidget.NewButtonWithIcon("", theme.ListIcon(), func() {
		v.showLineNumbers = !v.showLineNumbers
		v.app.Preferences().SetBool(prefShowLineNumbers, v.showLineNumbers)
		v.refreshDiffLists()
		v.refreshMainToolbar()
		v.refreshMainMenu()
	})
	v.btnLineNums.SetToolTip("Toggle source line numbers in the gutter")
	v.btnWhitespace = ttwidget.NewButtonWithIcon("", theme.VisibilityIcon(), func() {
		v.showWhitespace = !v.showWhitespace
		v.app.Preferences().SetBool(prefShowWhitespace, v.showWhitespace)
		v.refreshDiffLists()
		v.refreshMainToolbar()
		v.refreshMainMenu()
	})
	v.btnWhitespace.SetToolTip("Toggle visible whitespace (· space, → tab)")
	for _, b := range []*ttwidget.Button{v.btnSaveLeft, v.btnSaveRight, v.btnSaveBoth, v.btnLineNums, v.btnWhitespace} {
		b.Importance = widget.LowImportance
	}
	// Icon-only: order is save left, save right, save both | line numbers | show whitespace.
	return container.NewHBox(
		v.btnSaveLeft,
		v.btnSaveRight,
		v.btnSaveBoth,
		widget.NewSeparator(),
		v.btnLineNums,
		v.btnWhitespace,
	)
}

func (v *diffView) showMergeMenu(rowID widget.ListItemID, abs fyne.Position) {
	if v.model == nil || rowID < 0 || rowID >= len(v.model.Rows) || v.win == nil {
		return
	}
	rid := rowID
	dr := v.model.Rows[rid]
	takeLeft := fyne.NewMenuItem("Take left into right file", func() {
		ll := splitSourceLines(v.leftT)
		rr := splitSourceLines(v.rightT)
		if v.model == nil {
			return
		}
		newR, ok := applyLeftToRightAtRow(v.model, rid, ll, rr)
		if !ok {
			dialog.ShowError(fmt.Errorf("cannot apply left to right for this row"), v.win)
			return
		}
		v.rightT = joinSourceLines(newR)
		v.rightDirty = true
		v.recompute()
		v.syncScrollAfterEdit()
		v.refreshMainToolbar()
		v.refreshMainMenu()
	})
	takeRight := fyne.NewMenuItem("Take right into left file", func() {
		ll := splitSourceLines(v.leftT)
		rr := splitSourceLines(v.rightT)
		if v.model == nil {
			return
		}
		newL, ok := applyRightToLeftAtRow(v.model, rid, ll, rr)
		if !ok {
			dialog.ShowError(fmt.Errorf("cannot apply right to left for this row"), v.win)
			return
		}
		v.leftT = joinSourceLines(newL)
		v.leftDirty = true
		v.recompute()
		v.syncScrollAfterEdit()
		v.refreshMainToolbar()
		v.refreshMainMenu()
	})
	delLeft := fyne.NewMenuItem("Delete line from left file", func() {
		ll := splitSourceLines(v.leftT)
		if v.model == nil {
			return
		}
		newL, ok := deleteLeftLineAtRow(v.model, rid, ll)
		if !ok {
			dialog.ShowError(fmt.Errorf("no line to delete on the left for this row"), v.win)
			return
		}
		v.leftT = joinSourceLines(newL)
		v.leftDirty = true
		v.recompute()
		v.syncScrollAfterEdit()
		v.refreshMainToolbar()
		v.refreshMainMenu()
	})
	delRight := fyne.NewMenuItem("Delete line from right file", func() {
		rr := splitSourceLines(v.rightT)
		if v.model == nil {
			return
		}
		newR, ok := deleteRightLineAtRow(v.model, rid, rr)
		if !ok {
			dialog.ShowError(fmt.Errorf("no line to delete on the right for this row"), v.win)
			return
		}
		v.rightT = joinSourceLines(newR)
		v.rightDirty = true
		v.recompute()
		v.syncScrollAfterEdit()
		v.refreshMainToolbar()
		v.refreshMainMenu()
	})
	delLeft.Disabled = dr.LeftLineNo <= 0
	delRight.Disabled = dr.RightLineNo <= 0
	widget.ShowPopUpMenuAtPosition(fyne.NewMenu("",
		takeLeft,
		takeRight,
		fyne.NewMenuItemSeparator(),
		delLeft,
		delRight,
	), v.win.Canvas(), abs)
}

func (v *diffView) syncFrom(_ *widget.List, dst *widget.List, id widget.ListItemID) {
	if v.syncSel {
		return
	}
	v.syncSel = true
	dst.Select(id)
	dst.ScrollTo(id)
	v.syncSel = false
}

func (v *diffView) jumpToFileStart() {
	if v.model == nil || len(v.model.Rows) == 0 {
		return
	}
	v.changeSlot = 0
	v.syncSel = true
	v.leftList.Select(0)
	v.leftList.ScrollToTop()
	v.rightList.Select(0)
	v.rightList.ScrollToTop()
	v.syncSel = false
	if v.syncScrollOn {
		v.syncScrollPrevL = v.leftList.GetScrollOffset()
		v.syncScrollPrevR = v.rightList.GetScrollOffset()
	}
}

func (v *diffView) jumpToFileEnd() {
	if v.model == nil || len(v.model.Rows) == 0 {
		return
	}
	last := widget.ListItemID(len(v.model.Rows) - 1)
	if len(v.model.ChangeIndices) > 0 {
		v.changeSlot = len(v.model.ChangeIndices) - 1
	}
	v.syncSel = true
	v.leftList.Select(last)
	v.leftList.ScrollToBottom()
	v.rightList.Select(last)
	v.rightList.ScrollToBottom()
	v.syncSel = false
	if v.syncScrollOn {
		v.syncScrollPrevL = v.leftList.GetScrollOffset()
		v.syncScrollPrevR = v.rightList.GetScrollOffset()
	}
}

func (v *diffView) jumpDiff(delta int) {
	if v.model == nil || len(v.model.ChangeIndices) == 0 {
		return
	}
	v.changeSlot += delta
	if v.changeSlot < 0 {
		v.changeSlot = len(v.model.ChangeIndices) - 1
	}
	if v.changeSlot >= len(v.model.ChangeIndices) {
		v.changeSlot = 0
	}
	id := v.model.ChangeIndices[v.changeSlot]
	v.syncSel = true
	v.leftList.Select(id)
	v.leftList.ScrollTo(id)
	v.rightList.Select(id)
	v.rightList.ScrollTo(id)
	v.syncSel = false
	if v.syncScrollOn {
		v.syncScrollPrevL = v.leftList.GetScrollOffset()
		v.syncScrollPrevR = v.rightList.GetScrollOffset()
	}
}

func abs32(x float32) float32 {
	return float32(math.Abs(float64(x)))
}

func (v *diffView) syncScrollTick() {
	if v.leftList == nil || v.rightList == nil || !v.syncScrollOn {
		return
	}
	const eps = float32(2)
	lo := v.leftList.GetScrollOffset()
	ro := v.rightList.GetScrollOffset()
	if v.syncScrollApplying {
		v.syncScrollPrevL, v.syncScrollPrevR = lo, ro
		return
	}
	leftMoved := abs32(lo-v.syncScrollPrevL) > eps
	rightMoved := abs32(ro-v.syncScrollPrevR) > eps
	switch {
	case leftMoved && !rightMoved:
		v.syncScrollApplying = true
		v.rightList.ScrollToOffset(lo)
		v.syncScrollApplying = false
	case rightMoved && !leftMoved:
		v.syncScrollApplying = true
		v.leftList.ScrollToOffset(ro)
		v.syncScrollApplying = false
	case leftMoved && rightMoved && abs32(lo-ro) > eps:
		v.syncScrollApplying = true
		v.rightList.ScrollToOffset(lo)
		v.syncScrollApplying = false
	}
	v.syncScrollPrevL = v.leftList.GetScrollOffset()
	v.syncScrollPrevR = v.rightList.GetScrollOffset()
}

func (v *diffView) syncScrollPollLoop() {
	tick := time.NewTicker(33 * time.Millisecond)
	defer tick.Stop()
	for range tick.C {
		fyne.Do(v.syncScrollTick)
	}
}

func (v *diffView) readURI(u fyne.URI) (string, []byte, error) {
	rc, err := storage.Reader(u)
	if err != nil {
		return "", nil, err
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return "", nil, err
	}
	path := u.Path()
	if path == "" {
		path = u.String()
	}
	return path, b, nil
}

func (v *diffView) completeLoad(side int, path, text string) {
	fyne.Do(func() {
		if side == 0 {
			v.leftP, v.leftT = path, text
			v.leftDirty = false
		} else {
			v.rightP, v.rightT = path, text
			v.rightDirty = false
		}
		if path != "" {
			addRecent(v.app, side, path)
		}
		v.recompute()
		v.refreshMainMenu()
		v.refreshMainToolbar()
	})
}

func (v *diffView) loadPathFromRecent(side int, path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		fyne.Do(func() {
			removeRecent(v.app, side, path)
			v.refreshMainMenu()
			dialog.ShowError(fmt.Errorf("could not open recent file: %w", err), v.win)
		})
		return
	}
	v.completeLoad(side, path, string(b))
}

func (v *diffView) loadSide(side int, uri fyne.URI) {
	path, b, err := v.readURI(uri)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("read file: %w", err), v.win)
		})
		return
	}
	v.completeLoad(side, path, string(b))
}

func (v *diffView) showRecentMenuForSide(side int, anchor fyne.CanvasObject) {
	if v.win == nil || anchor == nil {
		return
	}
	ap := v.app.Driver().AbsolutePositionForObject(anchor)
	sz := anchor.Size()
	widget.ShowPopUpMenuAtPosition(v.buildRecentSubmenu(side), v.win.Canvas(), fyne.NewPos(ap.X, ap.Y+sz.Height))
}

func (v *diffView) openFileDialog(side int) {
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			fyne.Do(func() { dialog.ShowError(err, v.win) })
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		u := reader.URI()
		b, err := io.ReadAll(reader)
		if err != nil {
			fyne.Do(func() { dialog.ShowError(err, v.win) })
			return
		}
		path := u.Path()
		if path == "" {
			path = u.String()
		}
		v.completeLoad(side, path, string(b))
	}, v.win)
	d.Show()
}

func (v *diffView) dropTarget(pos fyne.Position, uris []fyne.URI) {
	if len(uris) == 0 {
		return
	}
	d := v.app.Driver()
	apL := d.AbsolutePositionForObject(v.leftCol)
	szL := v.leftCol.Size()
	apR := d.AbsolutePositionForObject(v.rightCol)
	szR := v.rightCol.Size()

	inLeft := pos.X >= apL.X && pos.X < apL.X+szL.Width && pos.Y >= apL.Y && pos.Y < apL.Y+szL.Height
	inRight := pos.X >= apR.X && pos.X < apR.X+szR.Width && pos.Y >= apR.Y && pos.Y < apR.Y+szR.Height

	side := -1
	switch {
	case inLeft && !inRight:
		side = 0
	case inRight && !inLeft:
		side = 1
	case inLeft && inRight:
		if pos.X < apL.X+szL.Width/2 {
			side = 0
		} else {
			side = 1
		}
	default:
		if v.leftP == "" {
			side = 0
		} else {
			side = 1
		}
	}
	if side >= 0 {
		v.loadSide(side, uris[0])
	}
}

func (v *diffView) buildToolbar(side int) *fyne.Container {
	first := ttwidget.NewButtonWithIcon("", theme.MediaFastRewindIcon(), func() { v.jumpToFileStart() })
	first.SetToolTip("Jump to start of diff")
	prev := ttwidget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { v.jumpDiff(-1) })
	prev.SetToolTip("Previous change (wraps)")
	next := ttwidget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { v.jumpDiff(1) })
	next.SetToolTip("Next change (wraps)")
	last := ttwidget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), func() { v.jumpToFileEnd() })
	last.SetToolTip("Jump to end of diff")
	for _, b := range []*ttwidget.Button{first, prev, next, last} {
		b.Importance = widget.LowImportance
	}
	var recent *ttwidget.Button
	recent = ttwidget.NewButtonWithIcon("", theme.HistoryIcon(), func() {
		v.showRecentMenuForSide(side, recent)
	})
	recent.SetToolTip("Open a recently used file on this side")
	recent.Importance = widget.LowImportance
	open := ttwidget.NewButton("Browse…", func() { v.openFileDialog(side) })
	open.SetToolTip("Choose a file from disk for this side")
	open.Importance = widget.MediumImportance

	row := container.NewHBox(first, prev, next, last, widget.NewSeparator(), recent, widget.NewSeparator(), open)
	if side == 0 {
		sync := ttwidget.NewCheck("Sync scroll", func(on bool) {
			v.syncScrollOn = on
			if on && v.leftList != nil && v.rightList != nil {
				v.syncScrollPrevL = v.leftList.GetScrollOffset()
				v.syncScrollPrevR = v.rightList.GetScrollOffset()
			}
		})
		sync.SetToolTip("Keep left and right scroll position aligned")
		row = container.NewHBox(first, prev, next, last, widget.NewSeparator(), sync, widget.NewSeparator(), recent, widget.NewSeparator(), open)
	}
	return row
}

func (v *diffView) buildUI() fyne.CanvasObject {
	v.leftTitle = widget.NewLabel("(none)")
	v.rightTitle = widget.NewLabel("(none)")
	v.leftTitle.TextStyle = fyne.TextStyle{Bold: true}
	v.rightTitle.TextStyle = fyne.TextStyle{Bold: true}
	v.leftTitle.Truncation = fyne.TextTruncateEllipsis
	v.rightTitle.Truncation = fyne.TextTruncateEllipsis

	hintL := widget.NewLabel("Drop a file here or use Browse / File menu")
	hintR := widget.NewLabel("Drop a file here or use Browse / File menu")
	hintL.TextStyle = fyne.TextStyle{Italic: true}
	hintR.TextStyle = fyne.TextStyle{Italic: true}

	v.leftList = widget.NewList(
		func() int {
			if v.model == nil {
				return 0
			}
			return len(v.model.Rows)
		},
		v.makeLineCell(0),
		v.updateLineCell(0),
	)
	v.rightList = widget.NewList(
		func() int {
			if v.model == nil {
				return 0
			}
			return len(v.model.Rows)
		},
		v.makeLineCell(1),
		v.updateLineCell(1),
	)

	v.leftList.OnSelected = func(id widget.ListItemID) { v.syncFrom(v.leftList, v.rightList, id) }
	v.rightList.OnSelected = func(id widget.ListItemID) { v.syncFrom(v.rightList, v.leftList, id) }

	leftHead := container.NewVBox(v.buildToolbar(0), v.leftTitle, hintL)
	rightHead := container.NewVBox(v.buildToolbar(1), v.rightTitle, hintR)

	leftScroll := container.NewScroll(v.leftList)
	rightScroll := container.NewScroll(v.rightList)
	leftScroll.SetMinSize(fyne.NewSize(200, 200))
	rightScroll.SetMinSize(fyne.NewSize(200, 200))

	v.leftCol = container.NewBorder(leftHead, nil, nil, nil, leftScroll)
	v.rightCol = container.NewBorder(rightHead, nil, nil, nil, rightScroll)

	split := container.NewHSplit(v.leftCol, v.rightCol)
	split.Offset = 0.5

	pad := layout.NewCustomPaddedLayout(3, 0, 3, 0)
	titleLbl := widget.NewLabel(appName)
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}
	verLbl := widget.NewLabel("v" + appVersion)
	headerRow := container.NewHBox(
		container.New(pad, newBrandingHeaderImage(resourceKrankyBearHackerPng)),
		container.NewVBox(titleLbl, verLbl),
		layout.NewSpacer(),
	)

	toolRow := v.buildMainChromeToolbar()
	v.recompute()
	v.refreshMainToolbar()
	// VBox would size the split to its minimum height only; Border lets the split
	// consume all space below the header (full-height scroll panes).
	topBar := container.NewVBox(headerRow, toolRow, widget.NewSeparator())
	return container.NewBorder(topBar, nil, nil, nil, split)
}

func (v *diffView) refreshDiffLists() {
	if v.leftList != nil {
		v.leftList.Refresh()
	}
	if v.rightList != nil {
		v.rightList.Refresh()
	}
}

func runApp() {
	a := app.NewWithID(appID)
	loadTheme(a)

	v := &diffView{
		app:             a,
		showLineNumbers: a.Preferences().BoolWithFallback(prefShowLineNumbers, false),
		showWhitespace:  a.Preferences().BoolWithFallback(prefShowWhitespace, false),
	}
	w := a.NewWindow(appName)
	v.win = w
	w.SetIcon(resourceKrankyBearHackerPng)
	w.SetContent(fynetooltip.AddWindowToolTipLayer(container.NewPadded(v.buildUI()), w.Canvas()))
	w.Resize(fyne.NewSize(1100, 700))
	w.SetOnDropped(v.dropTarget)

	// Closing the main window must exit the app and tear down the system tray
	// (driver Quit), not leave helper windows or tray running.
	w.SetMaster()
	w.SetCloseIntercept(func() {
		fynetooltip.DestroyWindowToolTipLayer(w.Canvas())
		quitFromMainWindow(a)
	})

	w.Show()
	go v.syncScrollPollLoop()
	v.setupMenus()
	a.Run()
}
