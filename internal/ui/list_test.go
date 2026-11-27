package ui

import "testing"

func TestClampCursor(t *testing.T) {
	tests := []struct{ cursor, count, want int }{
		{2, 3, 2},
		{3, 3, 2},
		{-1, 3, 0},
		{0, 0, 0},
		{5, 0, 0},
	}
	for _, tt := range tests {
		if got := clampCursor(tt.cursor, tt.count); got != tt.want {
			t.Errorf("clampCursor(%d, %d) = %d, want %d", tt.cursor, tt.count, got, tt.want)
		}
	}
}

func TestAdjustScrollOffset(t *testing.T) {
	tests := []struct{ cursor, offset, visible, total, want int }{
		{0, 0, 5, 10, 0}, // already visible
		{4, 0, 5, 10, 0}, // last visible row
		{5, 0, 5, 10, 1}, // one past the bottom: scroll down by one
		{9, 1, 5, 10, 5}, // jump to the end
		{2, 5, 5, 10, 2}, // above the top: scroll up
		{0, 4, 5, 3, 0},  // list shorter than the panel: no scrolling
		{7, 0, 0, 10, 7}, // zero visible rows counts as one
		{0, 0, 5, 0, 0},  // empty list
	}
	for _, tt := range tests {
		got := adjustScrollOffset(tt.cursor, tt.offset, tt.visible, tt.total)
		if got != tt.want {
			t.Errorf("adjustScrollOffset(%d, %d, %d, %d) = %d, want %d", tt.cursor, tt.offset, tt.visible, tt.total, got, tt.want)
		}
	}
}

func TestListCursorHandleKey(t *testing.T) {
	const count, pageSize = 20, 5
	tests := []struct {
		key        string
		start      int
		wantCursor int
		wantOK     bool
	}{
		{"j", 3, 4, true},
		{"down", 3, 4, true},
		{"k", 3, 2, true},
		{"up", 3, 2, true},
		{"k", 0, 0, true},   // stops at the top
		{"j", 19, 19, true}, // stops at the bottom
		{"d", 3, 8, true},
		{"f", 3, 8, true},
		{"pgdown", 18, 19, true},
		{"u", 8, 3, true},
		{"b", 8, 3, true},
		{"pgup", 2, 0, true},
		{"g", 12, 0, true},
		{"home", 12, 0, true},
		{"G", 2, 19, true},
		{"end", 2, 19, true},
		{"x", 7, 7, false}, // not a navigation key: nothing moves
	}
	for _, tt := range tests {
		l := listCursor{cursor: tt.start}
		ok := l.handleKey(tt.key, count, pageSize)
		if ok != tt.wantOK || l.cursor != tt.wantCursor {
			t.Errorf("%q from %d: cursor = %d, ok = %v; want %d, %v", tt.key, tt.start, l.cursor, ok, tt.wantCursor, tt.wantOK)
		}
		if ok && (l.cursor < l.offset || l.cursor >= l.offset+pageSize) {
			t.Errorf("%q from %d: cursor %d is off screen (offset %d)", tt.key, tt.start, l.cursor, l.offset)
		}
	}
}

func TestListCursorEmptyList(t *testing.T) {
	for _, key := range []string{"j", "k", "d", "u", "g", "G"} {
		var l listCursor
		l.handleKey(key, 0, 5)
		if l.cursor != 0 || l.offset != 0 {
			t.Errorf("%q on an empty list: cursor = %d, offset = %d; want 0, 0", key, l.cursor, l.offset)
		}
	}
}
