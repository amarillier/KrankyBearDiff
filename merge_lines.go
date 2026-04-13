package main

import (
	"slices"
	"strings"
)

func normalizeSourceNewlines(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

func splitSourceLines(s string) []string {
	s = normalizeSourceNewlines(s)
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func joinSourceLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n")
}

// nextRightInsertIndex is the 0-based index in the right file before which a new
// line should be inserted, using the next row after from that has a right line.
func nextRightInsertIndex(rows []DiffRow, from int) int {
	for j := from + 1; j < len(rows); j++ {
		if rows[j].RightLineNo > 0 {
			return rows[j].RightLineNo - 1
		}
	}
	return -1
}

func nextLeftInsertIndex(rows []DiffRow, from int) int {
	for j := from + 1; j < len(rows); j++ {
		if rows[j].LeftLineNo > 0 {
			return rows[j].LeftLineNo - 1
		}
	}
	return -1
}

// applyLeftToRightAtRow updates a copy of rightLines so the aligned row matches the left file.
func applyLeftToRightAtRow(m *DiffModel, row int, leftLines, rightLines []string) ([]string, bool) {
	if m == nil || row < 0 || row >= len(m.Rows) {
		return nil, false
	}
	dr := m.Rows[row]
	rl := slices.Clone(rightLines)

	switch {
	case dr.LeftTag == LineEqual && dr.RightTag == LineEqual:
		li, ri := dr.LeftLineNo-1, dr.RightLineNo-1
		if li < 0 || ri < 0 || li >= len(leftLines) || ri >= len(rl) {
			return nil, false
		}
		rl[ri] = leftLines[li]
		return rl, true

	case dr.LeftTag == LineRemoved && dr.RightTag == LinePadding:
		li := dr.LeftLineNo - 1
		if li < 0 || li >= len(leftLines) {
			return nil, false
		}
		line := leftLines[li]
		insertAt := nextRightInsertIndex(m.Rows, row)
		if insertAt < 0 {
			return append(rl, line), true
		}
		if insertAt > len(rl) {
			return nil, false
		}
		return slices.Insert(rl, insertAt, line), true

	case dr.LeftTag == LinePadding && dr.RightTag == LineAdded:
		ri := dr.RightLineNo - 1
		if ri < 0 || ri >= len(rl) {
			return nil, false
		}
		return slices.Delete(rl, ri, ri+1), true

	default:
		return nil, false
	}
}

// applyRightToLeftAtRow updates a copy of leftLines so the aligned row matches the right file.
func applyRightToLeftAtRow(m *DiffModel, row int, leftLines, rightLines []string) ([]string, bool) {
	if m == nil || row < 0 || row >= len(m.Rows) {
		return nil, false
	}
	dr := m.Rows[row]
	ll := slices.Clone(leftLines)

	switch {
	case dr.LeftTag == LineEqual && dr.RightTag == LineEqual:
		li, ri := dr.LeftLineNo-1, dr.RightLineNo-1
		if li < 0 || ri < 0 || ri >= len(rightLines) || li >= len(ll) {
			return nil, false
		}
		ll[li] = rightLines[ri]
		return ll, true

	case dr.LeftTag == LinePadding && dr.RightTag == LineAdded:
		ri := dr.RightLineNo - 1
		if ri < 0 || ri >= len(rightLines) {
			return nil, false
		}
		line := rightLines[ri]
		insertAt := nextLeftInsertIndex(m.Rows, row)
		if insertAt < 0 {
			return append(ll, line), true
		}
		if insertAt > len(ll) {
			return nil, false
		}
		return slices.Insert(ll, insertAt, line), true

	case dr.LeftTag == LineRemoved && dr.RightTag == LinePadding:
		li := dr.LeftLineNo - 1
		if li < 0 || li >= len(ll) {
			return nil, false
		}
		return slices.Delete(ll, li, li+1), true

	default:
		return nil, false
	}
}

// deleteLeftLineAtRow removes one line from the left file at the current aligned row (if that row has a left line).
func deleteLeftLineAtRow(m *DiffModel, row int, leftLines []string) ([]string, bool) {
	if m == nil || row < 0 || row >= len(m.Rows) {
		return nil, false
	}
	dr := m.Rows[row]
	if dr.LeftLineNo <= 0 {
		return nil, false
	}
	li := dr.LeftLineNo - 1
	if li < 0 || li >= len(leftLines) {
		return nil, false
	}
	return slices.Delete(slices.Clone(leftLines), li, li+1), true
}

// deleteRightLineAtRow removes one line from the right file at the current aligned row.
func deleteRightLineAtRow(m *DiffModel, row int, rightLines []string) ([]string, bool) {
	if m == nil || row < 0 || row >= len(m.Rows) {
		return nil, false
	}
	dr := m.Rows[row]
	if dr.RightLineNo <= 0 {
		return nil, false
	}
	ri := dr.RightLineNo - 1
	if ri < 0 || ri >= len(rightLines) {
		return nil, false
	}
	return slices.Delete(slices.Clone(rightLines), ri, ri+1), true
}
