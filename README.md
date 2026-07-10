# Knews

A terminal news reader for New Zealand RSS feeds — read the news in a clean,
keyboard-driven TUI, without ads, trackers, popups, or cookie banners.

Knews fetches articles from RSS feeds and renders the full text in a distraction-free
reader (using the same readability algorithm as Firefox's Reader View), so you get the
story and nothing else. It's built in Go with the
[Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework.

> Currently supports [RNZ (Radio New Zealand)](https://www.rnz.co.nz). The parser
> layer is designed to make adding more providers straightforward.

## Features

- Browse RSS feed categories and headlines in a two-panel layout
- Read full article text in a clean, centred reader view
- Fully keyboard-driven (vim-style `j`/`k`/`h`/`l` navigation)
- Concurrent feed fetching via a bounded worker pool
- Open the original article in your browser with `o`

## Install

Requires [Go](https://go.dev/dl/) 1.25 or newer.

```sh
# Build from source
git clone https://github.com/kralkapoor/knews.git
cd knews
go build -o knews .
./knews

# Or run directly
go run .
```

## Usage

Launch `knews`, pick a provider, then browse. Press `?` at any time for the full
keybinding list.

| Key            | Action                              |
| -------------- | ----------------------------------- |
| `j` / `k`      | Move down / up                      |
| `h` / `l`      | Focus feed list / article list      |
| `Tab`          | Switch panels                       |
| `Enter`        | Select feed / open article          |
| `r`            | Refresh feeds                       |
| `o`            | Open current article in browser     |
| `g` / `G`      | Jump to top / bottom (reader)       |
| `?`            | Show help                           |
| `q` / `Esc`    | Back / quit                         |

## Architecture

The codebase is split into four `internal` packages — `config`, `rss`, `article`,
and `ui` — wired together in `knews.go`. For a detailed walkthrough with diagrams,
see [ARCHITECTURE.md](ARCHITECTURE.md).

## Ideas to explore

A running list of things I'd like to investigate or build out next.

### More feeds & providers

- **Stuff** (`stuff.co.nz/rss`) — appears to publish all articles on a single feed.
- **NZ Herald** — appears to require a subscription.
- Re-enable the RNZ feeds currently commented out (Sport, Te Ao Māori, Country, etc.).

### TUI

- Compare alternative frameworks (e.g. [tview](https://github.com/rivo/tview)).
- Render article images into ASCII/ANSI for inline display.

### Serving over the network

- Run the binary on a home server and serve the TUI over SSH/HTTP/WebSockets
  (e.g. [Wish](https://github.com/charmbracelet/wish)).
- Gate access behind reverse-proxy auth for simple security.
- Investigate a web-terminal frontend.

### Companion frontend

- A supplementary REST API over the same fetching/parsing core.
- An ad-free web frontend (Astro?).
