package ui

// emptyListText is drawn in place of the rows of an empty list.
var emptyListText = metadataStyle.Render("No items.")

// listCursor remembers the highlighted row of a list and the first row shown,
// so lists longer than their panel can scroll.
type listCursor struct {
	cursor int // highlighted row
	offset int // first row drawn
}

// handleKey moves the cursor for a list navigation key, in a list of count
// rows with pageSize of them on screen. It reports whether key was one.
func (l *listCursor) handleKey(key string, count, pageSize int) bool {
	switch key {
	case "j", "down":
		l.cursor++
	case "k", "up":
		l.cursor--
	case "d", "f", "pgdown":
		l.cursor += pageSize
	case "u", "b", "pgup":
		l.cursor -= pageSize
	case "g", "home":
		l.cursor = 0
	case "G", "end":
		l.cursor = count - 1
	default:
		return false
	}
	l.scroll(count, pageSize)
	return true
}

// scroll keeps the cursor inside a list of count rows and scrolls just
// enough to keep it among the pageSize rows on screen. Call it whenever the
// list or the screen size changes.
func (l *listCursor) scroll(count, pageSize int) {
	l.cursor = clampCursor(l.cursor, count)
	l.offset = adjustScrollOffset(l.cursor, l.offset, pageSize, count)
}

// clampCursor keeps a list cursor inside a list of count items: between 0 and
// count-1, or 0 if the list is empty.
func clampCursor(cursor, count int) int {
	if cursor > count-1 {
		cursor = count - 1
	}
	if cursor < 0 {
		cursor = 0
	}
	return cursor
}

// adjustScrollOffset returns the index of the first visible item of a list
// that shows visibleCount of its totalItems items at a time. It scrolls the
// minimum needed to keep the cursor on screen, and never past the end.
func adjustScrollOffset(cursor, offset, visibleCount, totalItems int) int {
	if visibleCount < 1 {
		visibleCount = 1
	}

	// Scroll up if the cursor is above the first visible item.
	if cursor < offset {
		offset = cursor
	}
	// Scroll down if the cursor is below the last visible item.
	if cursor >= offset+visibleCount {
		offset = cursor - visibleCount + 1
	}

	// Don't leave blank space at the bottom of the list.
	maxOffset := totalItems - visibleCount
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}
	return offset
}
