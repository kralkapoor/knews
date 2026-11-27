package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/kralkapoor/knews/internal/article"
	"github.com/kralkapoor/knews/internal/rss"
	"github.com/muesli/termenv"
)

func TestMain(m *testing.M) {
	// Render without colour codes so output is stable regardless of terminal.
	lipgloss.SetColorProfile(termenv.Ascii)
	os.Exit(m.Run())
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		in   string
		max  int
		want string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long", 10, "this is..."},
		{"  padded  ", 6, "padded"},
		{"Whangaparāoa", 8, "Whang..."},
		{"Māori", 4, "M..."},
		{"abc", 2, "ab"},
		{"abc", 0, ""},
		{"abc", -5, ""},
	}
	for _, tt := range tests {
		if got := truncateString(tt.in, tt.max); got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
		}
	}
}

func TestKeyHint(t *testing.T) {
	if got, want := keyHint("q", "to quit"), "q to quit"; got != want {
		t.Errorf("keyHint() = %q, want %q", got, want)
	}
}

func TestTitledBox(t *testing.T) {
	content := "line one\n" + strings.Repeat("x", 100) + "\nline three\nline four\nline five"
	for _, width := range []int{8, 22, 40} {
		for _, height := range []int{2, 3, 6} {
			for _, title := range []string{"", "Feeds", strings.Repeat("Long title ", 10)} {
				box := titledBox(content, title, width, height, colorGray)
				lines := strings.Split(box, "\n")
				if len(lines) != height {
					t.Errorf("titledBox(w=%d, h=%d, title=%q) has %d lines, want %d", width, height, title, len(lines), height)
				}
				for i, line := range lines {
					if w := lipgloss.Width(line); w != width {
						t.Errorf("titledBox(w=%d, h=%d, title=%q) line %d is %d wide: %q", width, height, title, i, w, line)
					}
				}
			}
		}
	}
}

func TestContentWidth(t *testing.T) {
	tests := []struct{ width, want int }{
		{0, 1},
		{3, 1},
		{50, 46},
		{100, 75},
		{200, 80},
	}
	for _, tt := range tests {
		if got := articleContentWidth(tt.width); got != tt.want {
			t.Errorf("articleContentWidth() at width %d = %d, want %d", tt.width, got, tt.want)
		}
	}
}

func TestFormatArticleByline(t *testing.T) {
	tests := []struct {
		author, source string
		wantByline     bool
	}{
		{"Jane Smith", "RNZ", true},
		{"Stuff", "Stuff", false},
		{"", "RNZ", false},
	}
	for _, tt := range tests {
		out := formatArticle(&article.Article{Title: "T", Author: tt.author, Source: tt.source}, "", 60)
		if got := strings.Contains(out, "By "); got != tt.wantByline {
			t.Errorf("author %q, source %q: byline shown = %v, want %v", tt.author, tt.source, got, tt.wantByline)
		}
	}
}

func TestFormatPubDate(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Sun, 18 Jan 2026 19:25:50 +1300", "18 Jan 2026"},
		{"Fri, 2 Jan 2026 08:00:00 GMT", "02 Jan 2026"},
		{"2026-10-30T00:21:36Z", "30 Oct 2026"},
		{"sometime last week, probably", "sometime last..."},
		{"", ""},
	}
	for _, tt := range tests {
		if got := formatPubDate(rss.Item{PubDate: tt.in}); got != tt.want {
			t.Errorf("formatPubDate(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
