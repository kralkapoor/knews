package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kralkapoor/knews/internal/article"
)

type articleModel struct {
	viewport   viewport.Model
	currentURL string
	title      string
	width      int
	height     int
}

func newArticleModel() articleModel {
	vp := viewport.New(80, 20)
	return articleModel{
		viewport: vp,
	}
}

// contentWidth calculates the readable content width: capped at 80 columns
// and no more than 3/4 of the terminal width, for comfortable reading.
func (a *articleModel) contentWidth() int {
	maxWidth := a.width * 3 / 4
	if maxWidth > 80 {
		maxWidth = 80
	}
	if maxWidth < 40 {
		maxWidth = a.width - 4
	}
	if maxWidth < 1 {
		maxWidth = 1
	}
	return maxWidth
}

func (a *articleModel) setSize(width, height int) {
	a.width = width
	a.height = height

	// Reserve 1 line for the status bar, 2 for the box borders
	vpHeight := height - 3
	if vpHeight < 1 {
		vpHeight = 1
	}
	a.viewport.Width = a.contentWidth()
	a.viewport.Height = vpHeight
}

func (a *articleModel) showLoading(title string) {
	a.title = truncateTitle(title, 60)
	a.currentURL = ""
	a.viewport.SetContent(keyStyle.Render("\n\nLoading article..."))
	a.viewport.GotoTop()
}

func (a *articleModel) showError(title string, err error) {
	a.title = truncateTitle(title, 60)
	a.currentURL = ""
	content := fmt.Sprintf("\n\n%s\n\n%s",
		errorStyle.Render("Error loading article:"),
		err.Error(),
	)
	a.viewport.SetContent(content)
	a.viewport.GotoTop()
}

func (a *articleModel) showArticle(art *article.Article) {
	a.title = truncateTitle(art.Title, 60)
	a.currentURL = art.URL

	contentWidth := a.contentWidth()
	dividerWidth := contentWidth
	if dividerWidth > 60 {
		dividerWidth = 60
	}

	var b strings.Builder

	// Leading whitespace for breathing room
	b.WriteString("\n")

	// Title
	b.WriteString(titleStyle.Render(art.Title))
	b.WriteString("\n\n")

	// Metadata
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
	b.WriteString(metadataStyle.Render(strings.Repeat("─", dividerWidth)))
	b.WriteString("\n\n")

	// Content — add extra blank lines between paragraphs for readability.
	// go-readability's RenderText uses \n\n for paragraph breaks; we widen
	// those to \n\n\n so paragraphs have visible separation.
	content := strings.ReplaceAll(art.Content, "\n\n", "\n\n\n")
	b.WriteString(content)

	// Trailing whitespace so user can scroll past the end comfortably
	b.WriteString("\n\n")

	// Word-wrap the entire content to the viewport width so lines don't get
	// clipped. The viewport itself doesn't wrap text — it just clips.
	wrapped := lipgloss.NewStyle().Width(contentWidth).Render(b.String())
	a.viewport.SetContent(wrapped)
	a.viewport.GotoTop()
}

func (a articleModel) Update(msg tea.Msg) (articleModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return a, func() tea.Msg { return closeArticleMsg{} }
		case "o":
			a.openInBrowser()
			return a, nil
		case "j", "down":
			a.viewport.LineDown(1)
			return a, nil
		case "k", "up":
			a.viewport.LineUp(1)
			return a, nil
		case "g", "home":
			a.viewport.GotoTop()
			return a, nil
		case "G", "end":
			a.viewport.GotoBottom()
			return a, nil
		case "pgdown":
			a.viewport.HalfViewDown()
			return a, nil
		case "pgup":
			a.viewport.HalfViewUp()
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.viewport, cmd = a.viewport.Update(msg)
	return a, cmd
}

func (a articleModel) View() string {
	// Center the viewport content horizontally within the full-width box
	vpView := lipgloss.PlaceHorizontal(a.width-2, lipgloss.Center, a.viewport.View())
	box := titledBox(vpView, a.title, a.width, a.height-1, colorDarkCyan)

	// Status bar
	status := statusText(" ", "q", " Back │ ", "o", " Open in browser │ ", "j/k", " Scroll")

	return lipgloss.JoinVertical(lipgloss.Left, box, status)
}

func (a *articleModel) openInBrowser() {
	if a.currentURL == "" {
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", a.currentURL)
	case "darwin":
		cmd = exec.Command("open", a.currentURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", a.currentURL)
	default:
		return
	}

	cmd.Start()
}

// truncateTitle shortens a title string for display in the border.
func truncateTitle(s string, maxLength int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}
