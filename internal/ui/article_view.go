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

type articleModel struct {
	viewport viewport.Model
	title    string
	url      string           // page to open with "o"; set even while loading or after an error
	article  *article.Article // nil while loading or after an error
	message  string           // shown instead of an article: loading text or an error
	notice   string           // transient status message, cleared on the next key
	width    int
	height   int
}

func newArticleModel() articleModel {
	return articleModel{viewport: viewport.New(articleMaxWidth, 20)}
}

// contentWidth calculates the readable content width: capped at 80 columns
// and no more than 3/4 of the terminal width, for comfortable reading.
func (a articleModel) contentWidth() int {
	w := min(a.width*3/4, articleMaxWidth)
	if w < 40 {
		w = a.width - 4
	}
	return max(w, 1)
}

func (a *articleModel) setSize(width, height int) {
	a.width = width
	a.height = height

	// Reserve 1 line for the status bar, 2 for the box borders.
	a.viewport.Width = a.contentWidth()
	a.viewport.Height = max(height-3, 1)

	// The text is wrapped to the old width; rewrap it.
	a.render()
}

func (a *articleModel) showLoading(item rss.Item) {
	a.title = item.Title
	a.url = item.Link
	a.article = nil
	a.message = keyStyle.Render("\n\nLoading article...")
	a.render()
	a.viewport.GotoTop()
}

func (a *articleModel) showError(err error) {
	a.article = nil
	a.message = "\n\n" + errorStyle.Render("Error loading article:") + "\n\n" + err.Error() +
		"\n\n" + statusText("Press ", "o", " to open it in your browser.")
	a.render()
	a.viewport.GotoTop()
}

func (a *articleModel) showArticle(art *article.Article) {
	a.title = art.Title
	a.url = art.URL
	a.article = art
	a.render()
	a.viewport.GotoTop()
}

// render wraps the current article or message to the content width and loads
// it into the viewport. The viewport clips long lines rather than wrapping.
func (a *articleModel) render() {
	text := a.message
	if a.article != nil {
		text = formatArticle(a.article, a.contentWidth())
	}
	a.viewport.SetContent(lipgloss.NewStyle().Width(a.contentWidth()).Render(text))
}

// formatArticle lays out an article's title, metadata and body as text.
func formatArticle(art *article.Article, width int) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(art.Title))
	b.WriteString("\n\n")

	if art.Author != "" {
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

	b.WriteString("\n")
	b.WriteString(metadataStyle.Render(strings.Repeat("─", min(width, dividerMaxWidth))))
	b.WriteString("\n\n")

	// go-readability separates paragraphs with "\n\n"; widen the gap so
	// paragraphs stand apart.
	b.WriteString(strings.ReplaceAll(art.Content, "\n\n", "\n\n\n"))

	// Trailing space so the end of the article can be scrolled clear.
	b.WriteString("\n\n")

	return b.String()
}

func (a articleModel) Update(msg tea.Msg) (articleModel, tea.Cmd) {
	switch msg := msg.(type) {
	case openURLMsg:
		if msg.err != nil {
			a.notice = errorStyle.Render("Couldn't open browser: " + msg.err.Error())
		}
		return a, nil

	case tea.KeyMsg:
		a.notice = ""
		switch msg.String() {
		case "q", "esc":
			return a, func() tea.Msg { return closeArticleMsg{} }
		case "?":
			return a, func() tea.Msg { return showHelpMsg{} }
		case "o":
			if a.url == "" {
				return a, nil
			}
			return a, openURLCmd(a.url)
		case "g", "home":
			a.viewport.GotoTop()
			return a, nil
		case "G", "end":
			a.viewport.GotoBottom()
			return a, nil
		}
	}

	// The viewport's own keymap handles scrolling: j/k, d/u, space/f/b, pgup/pgdn.
	var cmd tea.Cmd
	a.viewport, cmd = a.viewport.Update(msg)
	return a, cmd
}

func (a articleModel) View() string {
	// Center the viewport content horizontally within the full-width box.
	vpView := lipgloss.PlaceHorizontal(a.width-2, lipgloss.Center, a.viewport.View())
	box := titledBox(vpView, a.title, a.width, a.height-statusHeight, colorDarkCyan)

	status := statusText(" ", "q", " Back │ ", "o", " Open in browser │ ", "j/k", " Scroll")
	if a.notice != "" {
		status = " " + a.notice
	}
	status = lipgloss.NewStyle().MaxWidth(a.width).Render(status)

	return lipgloss.JoinVertical(lipgloss.Left, box, status)
}
