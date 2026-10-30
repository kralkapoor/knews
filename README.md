# Knews

A terminal news reader for New Zealand RSS feeds — read the news in a clean,
keyboard-driven TUI, without ads, trackers, popups, or cookie banners.

Knews fetches articles from RSS feeds and renders the full text in a distraction-free
reader (using the same readability algorithm as Firefox's Reader View), so you get the
story and nothing else. It's built in Go with the
[Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework.

> Currently supports [RNZ (Radio New Zealand)](https://www.rnz.co.nz). Providers are
> plain configuration in `internal/config/providers.go`, so adding a server-rendered
> news site needs no new code.

## Features

- Browse RSS feed categories and headlines in a two-panel layout
- Read full article text in a clean, centred reader view
- Fully keyboard-driven (vim-style `j`/`k`/`h`/`l` navigation)
- Concurrent feed fetching via a bounded worker pool
- Open the original article in your browser with `o`

## Install

Requires [Go](https://go.dev/dl/) 1.26 or newer.

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
| `d` / `u`      | Half page down / up (reader)        |
| `Space` / `b`  | Page down / up (reader)             |
| `g` / `G`      | Jump to top / bottom (reader)       |
| `?`            | Show help                           |
| `q` / `Esc`    | Back to list (reader)               |
| `q`            | Quit (lists)                        |
| `Ctrl+C`       | Quit from anywhere                  |

## Architecture

The codebase is split into five `internal` packages — `config`, `httpx`, `rss`,
`article`, and `ui` — wired together in `knews.go`. For a detailed walkthrough with
diagrams, see [ARCHITECTURE.md](ARCHITECTURE.md).

Run the tests with `go test -race ./...`.

## Ideas to explore

A running list of things I'd like to investigate or build out next.

### More feeds & providers

- **Stuff** — publishes per-section Atom feeds at `stuff.co.nz/rss/<section>`
  (`nz-news`, `world-news`, `business`, `politics`, ...), but article pages are
  rendered client-side, so readability can't extract them. Show the feed summary
  plus "open in browser" instead.
- **NZ Herald** — publishes RSS 2.0 at
  `nzherald.co.nz/arc/outboundfeeds/rss/section/<section>/?outputType=xml`;
  Premium articles are paywalled.
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
