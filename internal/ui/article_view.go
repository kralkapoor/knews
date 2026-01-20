package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/kralkapoor/knews/internal/article"
	"github.com/rivo/tview"
)

// ArticleView represents the full article display view.
type ArticleView struct {
	layout     *tview.Flex
	textView   *tview.TextView
	app        *App
	currentURL string
}

// NewArticleView creates a new ArticleView component.
func NewArticleView(app *App) *ArticleView {
	av := &ArticleView{
		textView: tview.NewTextView(),
		app:      app,
	}

	av.textView.SetDynamicColors(true).
		SetWordWrap(true).
		SetScrollable(true)
	av.textView.SetBorder(true).
		SetBorderColor(tcell.ColorDarkCyan)

	// Create layout with status bar
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText(" [yellow]q[white] Back | [yellow]o[white] Open in browser | [yellow]j/k[white] Scroll")

	av.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(av.textView, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	return av
}

// HandleInput processes keyboard input for the article view.
func (av *ArticleView) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			av.app.closeArticle()
			return nil
		case 'o':
			av.openInBrowser()
			return nil
		case 'j':
			row, col := av.textView.GetScrollOffset()
			av.textView.ScrollTo(row+1, col)
			return nil
		case 'k':
			row, col := av.textView.GetScrollOffset()
			if row > 0 {
				av.textView.ScrollTo(row-1, col)
			}
			return nil
		case 'g':
			av.textView.ScrollToBeginning()
			return nil
		case 'G':
			av.textView.ScrollToEnd()
			return nil
		}
	case tcell.KeyEscape:
		av.app.closeArticle()
		return nil
	case tcell.KeyDown:
		row, col := av.textView.GetScrollOffset()
		av.textView.ScrollTo(row+1, col)
		return nil
	case tcell.KeyUp:
		row, col := av.textView.GetScrollOffset()
		if row > 0 {
			av.textView.ScrollTo(row-1, col)
		}
		return nil
	case tcell.KeyPgDn:
		row, col := av.textView.GetScrollOffset()
		av.textView.ScrollTo(row+10, col)
		return nil
	case tcell.KeyPgUp:
		row, col := av.textView.GetScrollOffset()
		if row > 10 {
			av.textView.ScrollTo(row-10, col)
		} else {
			av.textView.ScrollTo(0, col)
		}
		return nil
	case tcell.KeyHome:
		av.textView.ScrollToBeginning()
		return nil
	case tcell.KeyEnd:
		av.textView.ScrollToEnd()
		return nil
	}
	return event
}

// ShowLoading displays a loading message while fetching the article.
func (av *ArticleView) ShowLoading(title string) {
	av.textView.SetTitle(" " + truncateTitle(title, 60) + " ")
	av.textView.SetText("\n\n[yellow]Loading article...[white]")
	av.textView.ScrollToBeginning()
}

// ShowError displays an error message when article fetching fails.
func (av *ArticleView) ShowError(title string, err error) {
	av.textView.SetTitle(" " + truncateTitle(title, 60) + " ")
	av.textView.SetText(fmt.Sprintf("\n\n[red]Error loading article:[white]\n\n%s", err.Error()))
}

// ShowArticle displays the full article content.
func (av *ArticleView) ShowArticle(art *article.Article) {
	av.currentURL = art.URL
	av.textView.SetTitle(" " + truncateTitle(art.Title, 60) + " ")

	var builder strings.Builder

	// Title
	builder.WriteString(fmt.Sprintf("[::b]%s[::-]\n\n", art.Title))

	// Metadata
	if art.Author.Name != "" {
		builder.WriteString(fmt.Sprintf("[gray]By %s[white]\n", art.Author.Name))
	}
	if !art.PublishedAt.IsZero() {
		builder.WriteString(fmt.Sprintf("[gray]%s[white]\n", art.PublishedAt.Format("Monday, 02 January 2006")))
	}
	if art.Source != "" {
		builder.WriteString(fmt.Sprintf("[gray]Source: %s[white]\n", art.Source))
	}

	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("─", 60))
	builder.WriteString("\n\n")

	// Content blocks
	for _, block := range art.Content {
		switch block.Type {
		case article.BlockTypeHeading:
			builder.WriteString(fmt.Sprintf("[::b]%s[::-]\n\n", block.Content))
		case article.BlockTypeQuote:
			builder.WriteString(fmt.Sprintf("[gray]> %s[white]\n\n", block.Content))
		case article.BlockTypeParagraph:
			builder.WriteString(block.Content)
			builder.WriteString("\n\n")
		default:
			builder.WriteString(block.Content)
			builder.WriteString("\n\n")
		}
	}

	// Tags
	if len(art.Tags) > 0 {
		builder.WriteString(strings.Repeat("─", 60))
		builder.WriteString("\n\n")
		builder.WriteString("[gray]Tags: ")
		builder.WriteString(strings.Join(art.Tags, ", "))
		builder.WriteString("[white]\n")
	}

	av.textView.SetText(builder.String())
	av.textView.ScrollToBeginning()
}

// openInBrowser opens the current article URL in the default browser.
func (av *ArticleView) openInBrowser() {
	if av.currentURL == "" {
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", av.currentURL)
	case "darwin":
		cmd = exec.Command("open", av.currentURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", av.currentURL)
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
