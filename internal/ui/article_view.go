package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kralkapoor/knews/internal/article"
	"github.com/kralkapoor/knews/internal/rss"
)

const (
	articleMaxWidth = 80 // readable line length cap
	dividerMaxWidth = 60
)

// articleState is everything the article reader screen remembers.
type articleState struct {
	viewport viewport.Model
	title    string
	url      string           // page to open with "o"; set even while loading or after an error
	article  *article.Article // nil while loading
	loading  bool             // show "Loading article..." instead of an article
	note     string           // shown after a summary: why it's not the full text
	notice   string           // transient status message, cleared on the next key
}

func newArticleState() articleState {
	return articleState{viewport: viewport.New(articleMaxWidth, 20)}
}

// articleContentWidth calculates the readable content width: capped at 80
// columns and no more than 3/4 of the terminal width, for comfortable reading.
func articleContentWidth(termWidth int) int {
	w := min(termWidth*3/4, articleMaxWidth)
	if w < 40 {
		w = termWidth - 4
	}
	return max(w, 1)
}

// resizeArticleViewport fits the viewport to the terminal and rewraps the
// text, which was wrapped to the old width.
func (m *Model) resizeArticleViewport() {
	// Reserve 1 line for the status bar, 2 for the box borders.
	m.article.viewport.Width = articleContentWidth(m.width)
	m.article.viewport.Height = max(m.height-3, 1)
	m.renderArticleContent()
}

func (m *Model) showArticleLoading(item rss.Item) {
	m.article.title = item.Title
	m.article.url = item.Link
	m.article.article = nil
	m.article.loading = true
	m.article.note = ""
	m.renderArticleContent()
	m.article.viewport.GotoTop()
}

// showArticle shows art, followed by note if it isn't empty.
func (m *Model) showArticle(art *article.Article, note string) {
	m.article.title = art.Title
	m.article.url = art.URL
	m.article.article = art
	m.article.loading = false
	m.article.note = note
	m.renderArticleContent()
	m.article.viewport.GotoTop()
}

// showArticleSummary shows an article built from its feed entry rather than
// the full page, with a note saying why: err is the reason the full text
// couldn't be loaded, or nil if the provider only offers summaries.
func (m *Model) showArticleSummary(art *article.Article, err error) {
	note := "This is the summary from the feed. Press " + keyHint("o", "to read the full article in your browser.")
	if err != nil {
		note = errorStyle.Render("Couldn't load the full article: ") + err.Error() + "\n\n" + note
	}
	m.showArticle(art, note)
}

// renderArticleContent wraps the current article (or loading message) to the
// content width and loads it into the viewport. The viewport clips long lines
// rather than wrapping them.
func (m *Model) renderArticleContent() {
	width := articleContentWidth(m.width)

	text := ""
	if m.article.loading {
		text = keyStyle.Render("\n\nLoading article...")
	}
	if m.article.article != nil {
		text = formatArticle(m.article.article, m.article.note, width)
	}
	m.article.viewport.SetContent(lipgloss.NewStyle().Width(width).Render(text))
}

// formatArticle lays out an article's title, metadata and body as text,
// followed by note if it isn't empty.
func formatArticle(art *article.Article, note string, width int) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(art.Title))
	b.WriteString("\n\n")

	// Some feeds credit every article to the outlet itself; skip that.
	if art.Author != "" && art.Author != art.Source {
		b.WriteString(metadataStyle.Render("By " + art.Author))
		b.WriteString("\n")
	}
	if !art.PublishedAt.IsZero() {
		b.WriteString(metadataStyle.Render(art.PublishedAt.Format("Monday, 02 January 2006")))
		b.WriteString("\n")
	}
	if art.Source != "" {
		b.WriteString(metadataStyle.Render("Source: " + art.Source))
		b.WriteString("\n")
	}

	divider := metadataStyle.Render(strings.Repeat("─", min(width, dividerMaxWidth)))
	b.WriteString("\n")
	b.WriteString(divider)
	b.WriteString("\n\n")

	// go-readability separates paragraphs with "\n\n"; widen the gap so
	// paragraphs stand apart.
	b.WriteString(strings.ReplaceAll(art.Content, "\n\n", "\n\n\n"))

	if note != "" {
		b.WriteString("\n\n")
		b.WriteString(divider)
		b.WriteString("\n\n")
		b.WriteString(note)
	}

	// Trailing space so the end of the article can be scrolled clear.
	b.WriteString("\n\n")

	return b.String()
}

// updateArticle handles the browser-launch result and key presses on the
// article reader screen.
func (m *Model) updateArticle(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case openURLMsg:
		if msg.err != nil {
			m.article.notice = errorStyle.Render("Couldn't open browser: " + msg.err.Error())
		}
		return nil

	case tea.KeyMsg:
		m.article.notice = ""
		switch msg.String() {
		case "q", "esc":
			// Back to the lists, abandoning any fetch still in flight.
			m.stopArticleFetch()
			m.activeScreen = screenMain
			return nil
		case "o":
			if m.article.url == "" {
				return nil
			}
			return openURLCmd(m.article.url)
		case "g", "home":
			m.article.viewport.GotoTop()
			return nil
		case "G", "end":
			m.article.viewport.GotoBottom()
			return nil
		}
	}

	// The viewport's own keymap handles scrolling: j/k, d/u, space/f/b, pgup/pgdn.
	var cmd tea.Cmd
	m.article.viewport, cmd = m.article.viewport.Update(msg)
	return cmd
}

func (m Model) viewArticle() string {
	// Center the viewport content horizontally within the full-width box.
	vpView := lipgloss.PlaceHorizontal(m.width-2, lipgloss.Center, m.article.viewport.View())
	box := titledBox(vpView, m.article.title, m.width, m.height-statusHeight, colorDarkCyan)

	status := " " + keyHint("q", "Back") + " │ " + keyHint("o", "Open in browser") + " │ " + keyHint("j/k", "Scroll")
	if m.article.notice != "" {
		status = " " + m.article.notice
	}
	status = lipgloss.NewStyle().MaxWidth(m.width).Render(status)

	return lipgloss.JoinVertical(lipgloss.Left, box, status)
}
