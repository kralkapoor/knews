# Knews Architecture

Knews is a terminal-based RSS news reader built in Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework. It currently supports RNZ (Radio New Zealand) feeds, with an extensible design for adding more providers.

## Table of Contents

- [High-Level Overview](#high-level-overview)
- [Project Structure](#project-structure)
- [Type Diagram](#type-diagram)
- [Application Lifecycle](#application-lifecycle)
- [RSS Subsystem](#rss-subsystem)
- [Article Parsing Subsystem](#article-parsing-subsystem)
- [UI System](#ui-system)
- [Full Data Flow](#full-data-flow)

---

## High-Level Overview

The application is organised into four internal packages, each with a single responsibility. The main entry point (`knews.go`) wires them together.

```mermaid
graph TB
    subgraph knews.go
        main["main()"]
    end

    subgraph "internal/config"
        config["Provider registry<br/>Feed URLs & ordering"]
    end

    subgraph "internal/rss"
        rss["Feed fetching & parsing<br/>Worker pool concurrency"]
    end

    subgraph "internal/article"
        article["Article fetching & parsing<br/>Parser registry pattern"]
    end

    subgraph "internal/ui"
        ui["Bubble Tea TUI<br/>Screen management"]
    end

    main --> rss
    main --> article
    main --> ui
    ui --> rss
    ui --> article
    ui --> config
    rss -.->|"RSS XML"| Internet["RSS Feed Servers"]
    article -.->|"HTML"| Internet2["News Article Pages"]
```

**Package responsibilities:**

| Package | Purpose | Key pattern |
|---------|---------|-------------|
| `config` | Defines providers, feed URLs, and display ordering | Static registry |
| `rss` | Fetches and parses RSS XML feeds concurrently | Worker pool |
| `article` | Fetches full article HTML and extracts readable content | Parser registry |
| `ui` | Terminal interface with screen-based navigation | Bubble Tea Model-Update-View |

---

## Project Structure

```
knews/
├── knews.go                    # Entry point — wires all layers together
└── internal/
    ├── config/
    │   ├── types.go            # Provider, FeedConfiguration structs
    │   ├── feeds.go            # RNZ provider definition & feed URLs
    │   └── config.go           # Test configuration helper
    ├── rss/
    │   ├── types.go            # Rss, Channel, Item, FeedResult, FeedRequest
    │   ├── client.go           # FeedClient — HTTP fetch + XML parse
    │   ├── parser.go           # ParseXmlResponse — XML unmarshalling
    │   └── fetcher.go          # FeedFetcher — concurrent worker pool
    ├── article/
    │   ├── types.go            # Article struct
    │   ├── parser.go           # ArticleParser interface, ParserRegistry, ArticleFetcher
    │   └── rnz_parser.go       # RNZParser — go-readability based extraction
    └── ui/
        ├── model.go            # Root Model — screen routing & message dispatch
        ├── commands.go         # Async tea.Cmd functions & message types
        ├── splash.go           # Provider selection screen
        ├── main_view.go        # Two-panel feed/article listing
        ├── article_view.go     # Scrollable article reader
        ├── help.go             # Keybinding help overlay
        ├── styles.go           # Colour palette & layout helpers
        └── keybindings.go      # Key definitions & descriptions
```

---

## Type Diagram

All key types and their relationships across packages:

```mermaid
classDiagram
    direction TB

    namespace rss {
        class FeedClient {
            +Fetch(ctx, url) []byte, error
            +FetchAndParse(ctx, url) Rss, error
        }

        class FeedFetcher {
            -client *FeedClient
            -maxWorkers int
            +FetchAll(ctx, requests) []FeedResult
            +FetchAllURLs(ctx, urls) []FeedResult
            +FetchAllFromMap(ctx, feeds) []FeedResult
            -worker(ctx, requests, results, ch, wg)
            -fetchSingle(ctx, request) FeedResult
        }

        class Rss {
            +Channel Channel
        }

        class Channel {
            +Title string
            +Description string
            +Link string
            +Copyright string
            +Items []Item
        }

        class Item {
            +Title string
            +Description string
            +PubDate string
            +Link string
            +GUID string
        }

        class FeedRequest {
            +URL string
            +Name string
        }

        class FeedResult {
            +Feed Rss
            +SourceURL string
            +FeedName string
            +FetchedAt time.Time
            +Error error
        }
    }

    namespace article {
        class ArticleParser {
            <<interface>>
            +Parse(htmlContent, url) *Article, error
            +CanParse(url) bool
        }

        class ParserRegistry {
            -parsers []ArticleParser
            +Register(parser)
            +GetParser(url) ArticleParser
            +Parse(htmlContent, url) *Article, error
        }

        class ArticleFetcher {
            -registry *ParserRegistry
            +Fetch(ctx, url) *Article, error
        }

        class RNZParser {
            +Parse(htmlContent, url) *Article, error
            +CanParse(url) bool
        }

        class Article {
            +URL string
            +Title string
            +Author string
            +Content string
            +PublishedAt time.Time
            +Source string
        }
    }

    namespace config {
        class Provider {
            +Name string
            +Description string
            +Feeds map~string string~
            +FeedOrder []string
        }
    }

    namespace ui {
        class Model {
            -activeScreen screen
            -splash splashModel
            -main mainModel
            -articleView articleModel
            -help helpModel
            -feedFetcher *FeedFetcher
            -articleFetcher *ArticleFetcher
            +Init() tea.Cmd
            +Update(msg) tea.Model, tea.Cmd
            +View() string
        }

        class mainModel {
            -feedList list.Model
            -articleList list.Model
            -focus focusPanel
            -feedResults map~string FeedResult~
            -currentFeed string
        }

        class articleModel {
            -viewport viewport.Model
            -currentURL string
            -title string
            +showLoading(title)
            +showArticle(article)
            +showError(title, err)
        }

        class splashModel {
            -providerList list.Model
        }

        class helpModel {
        }
    }

    Rss *-- Channel
    Channel *-- Item
    FeedFetcher --> FeedClient : uses
    FeedFetcher ..> FeedRequest : accepts
    FeedFetcher ..> FeedResult : returns
    FeedResult *-- Rss

    RNZParser ..|> ArticleParser : implements
    ParserRegistry o-- ArticleParser : contains
    ArticleFetcher --> ParserRegistry : uses
    ArticleFetcher ..> Article : returns

    Model --> FeedFetcher : uses
    Model --> ArticleFetcher : uses
    Model *-- splashModel
    Model *-- mainModel
    Model *-- articleModel
    Model *-- helpModel
    mainModel ..> FeedResult : displays
    mainModel ..> Item : displays
```

---

## Application Lifecycle

### Startup Sequence

```mermaid
sequenceDiagram
    participant main as main()
    participant rss as rss.FeedClient
    participant fetcher as rss.FeedFetcher
    participant registry as article.ParserRegistry
    participant rnz as article.RNZParser
    participant af as article.ArticleFetcher
    participant ui as ui.Model
    participant tea as bubbletea.Program

    main->>rss: NewFeedClient()
    main->>fetcher: NewFeedFetcher(client, 5)
    main->>registry: NewParserRegistry()
    main->>rnz: NewRNZParser()
    main->>registry: Register(rnzParser)
    main->>af: NewArticleFetcher(registry)
    main->>ui: NewModel(feedFetcher, articleFetcher)
    Note over ui: activeScreen = screenSplash
    main->>tea: NewProgram(model, WithAltScreen())
    main->>tea: Run()
    Note over tea: Enters event loop
```

### Provider Selection → Feed Loading

```mermaid
sequenceDiagram
    participant User
    participant Splash as splashModel
    participant Root as Model (root)
    participant Config as config
    participant Cmd as fetchFeedsCmd
    participant Fetcher as FeedFetcher
    participant Workers as Worker Pool
    participant Server as RSS Servers
    participant Main as mainModel

    User->>Splash: Enter (select provider)
    Splash->>Root: providerSelectedMsg{name: "RNZ"}
    Root->>Config: GetProvider("RNZ")
    Config-->>Root: Provider{Feeds, FeedOrder}
    Root->>Main: setFeeds(feedOrder)
    Note over Root: activeScreen = screenMain
    Root->>Cmd: fetchFeedsCmd(fetcher, provider)

    Note over Cmd: Runs asynchronously (30s timeout)
    Cmd->>Fetcher: FetchAllFromMap(ctx, feeds)
    Fetcher->>Workers: Distribute FeedRequests across workers
    par Concurrent fetching
        Workers->>Server: HTTP GET business.xml
        Workers->>Server: HTTP GET national.xml
        Workers->>Server: HTTP GET world.xml
        Note over Workers: ...up to 5 concurrent
    end
    Server-->>Workers: XML responses
    Workers-->>Fetcher: []FeedResult
    Fetcher-->>Cmd: []FeedResult

    Cmd->>Root: feedsLoadedMsg{results}
    Root->>Main: setFeedResults(results, firstFeed)
    Note over Main: Populates feed list & article list
```

### Article Selection → Content Display

```mermaid
sequenceDiagram
    participant User
    participant Main as mainModel
    participant Root as Model (root)
    participant AView as articleModel
    participant Cmd as fetchArticleCmd
    participant AF as ArticleFetcher
    participant Registry as ParserRegistry
    participant Parser as RNZParser
    participant Server as News Server

    User->>Main: Enter (select article)
    Main->>Root: openArticleMsg{item}
    Root->>AView: showLoading(item.Title)
    Note over Root: activeScreen = screenArticle
    Root->>Cmd: fetchArticleCmd(fetcher, item)

    Note over Cmd: Runs asynchronously (15s timeout)
    Cmd->>AF: Fetch(ctx, item.Link)
    AF->>Registry: GetParser(url)
    Registry->>Parser: CanParse(url)?
    Parser-->>Registry: true (contains "rnz.co.nz")
    Registry-->>AF: RNZParser

    AF->>Server: HTTP GET article page
    Server-->>AF: HTML response
    AF->>Parser: Parse(htmlBytes, url)
    Note over Parser: go-readability extracts<br/>title, author, content, date
    Parser-->>AF: *Article

    AF-->>Cmd: *Article
    Cmd->>Root: articleLoadedMsg{article, title}
    Root->>AView: showArticle(article)
    Note over AView: Renders content in viewport
```

### Feed Refresh

```mermaid
sequenceDiagram
    participant User
    participant Main as mainModel
    participant Root as Model (root)
    participant Cmd as refreshFeedsCmd
    participant Fetcher as FeedFetcher

    User->>Main: Press 'r'
    Main->>Root: requestRefreshMsg{}
    Root->>Root: statusMsg = "Refreshing feeds..."
    Root->>Cmd: refreshFeedsCmd(fetcher, provider)
    Note over Cmd: Same as fetchFeedsCmd but returns<br/>feedsRefreshedMsg instead
    Cmd->>Fetcher: FetchAllFromMap(ctx, feeds)
    Fetcher-->>Cmd: []FeedResult
    Cmd->>Root: feedsRefreshedMsg{results}
    Root->>Main: setFeedResults(results, currentFeed)
    Note over Main: Updates article list,<br/>preserves current feed selection
```

---

## RSS Subsystem

The RSS subsystem handles fetching and parsing RSS XML feeds with controlled concurrency.

### Worker Pool Architecture

```mermaid
graph LR
    subgraph Input
        R1[FeedRequest 1]
        R2[FeedRequest 2]
        R3[FeedRequest 3]
        R4[FeedRequest 4]
        R5[FeedRequest 5]
        R6[FeedRequest 6]
        R7[FeedRequest 7]
        R8[FeedRequest 8]
    end

    subgraph "Buffered Channel"
        CH["requestChannel<br/>(indices)"]
    end

    subgraph "Worker Pool (max 5)"
        W1["Worker 1"]
        W2["Worker 2"]
        W3["Worker 3"]
        W4["Worker 4"]
        W5["Worker 5"]
    end

    subgraph "Per Worker"
        direction TB
        FS["fetchSingle()"]
        FC["FeedClient.FetchAndParse()"]
        HTTP["HTTP GET"]
        XML["ParseXmlResponse()"]
        FS --> FC --> HTTP --> XML
    end

    subgraph Output
        RS["results []FeedResult<br/>(same order as input)"]
    end

    R1 & R2 & R3 & R4 & R5 & R6 & R7 & R8 --> CH
    CH --> W1 & W2 & W3 & W4 & W5
    W1 & W2 & W3 & W4 & W5 --> RS
```

**Key design decisions:**

- **Index-based channel**: The channel carries request indices, not requests. Workers read from the shared `requests` slice and write to the shared `results` slice at matching indices, preserving input order.
- **Context cancellation**: Each worker checks `ctx.Done()` before processing the next request. If the context is cancelled (e.g. timeout), remaining results are filled with the context error.
- **Worker count**: `min(maxWorkers, len(requests))` — never spawns more goroutines than there are requests.

### RSS Data Flow

```mermaid
graph TB
    URL["Feed URL string"] --> Fetch["FeedClient.Fetch()<br/>HTTP GET → []byte"]
    Fetch --> Parse["ParseXmlResponse()<br/>xml.Unmarshal"]
    Parse --> RssStruct["Rss{Channel{Items[]}}"]
    RssStruct --> Result["FeedResult{<br/>Feed, SourceURL,<br/>FeedName, FetchedAt, Error}"]

    style URL fill:#e8f4f8
    style Result fill:#e8f4f8
```

---

## Article Parsing Subsystem

The article subsystem uses the **registry pattern** to support multiple news sources. Each source gets its own parser implementation; the registry dispatches to the right one based on URL.

### Parser Registry Pattern

```mermaid
graph TB
    URL["Article URL"] --> Registry["ParserRegistry.GetParser(url)"]
    Registry --> Loop{"For each parser:<br/>parser.CanParse(url)?"}
    Loop -->|"contains rnz.co.nz"| RNZ["RNZParser"]
    Loop -->|"future: contains stuff.co.nz"| Future["StuffParser"]
    Loop -->|"no match"| Err["Error: no parser available"]

    RNZ --> Readability["go-readability<br/>(Firefox Reader View algorithm)"]
    Readability --> Article["Article{<br/>Title, Author,<br/>Content, PublishedAt, Source}"]

    style Future fill:#f0f0f0,stroke-dasharray: 5 5
```

### Article Fetch Sequence

```mermaid
graph LR
    AF["ArticleFetcher.Fetch(ctx, url)"]
    AF --> GetParser["registry.GetParser(url)"]
    GetParser --> HTTP["HTTP GET full page HTML"]
    HTTP --> Parse["parser.Parse(htmlBytes, url)"]
    Parse --> Art["*Article"]

    style AF fill:#e8f4f8
    style Art fill:#e8f4f8
```

**Adding a new news source** requires just two steps:
1. Create a struct implementing `ArticleParser` (the `Parse` and `CanParse` methods)
2. Register it in `main()`: `parserRegistry.Register(NewMyParser())`

---

## UI System

The UI uses Bubble Tea's **Model-Update-View** (MVU) architecture — a unidirectional data flow pattern where the model is the single source of truth.

### Screen State Machine

```mermaid
stateDiagram-v2
    [*] --> Splash

    Splash --> Main : providerSelectedMsg

    Main --> Article : openArticleMsg
    Main --> Help : showHelpMsg (?)

    Article --> Main : closeArticleMsg (Esc)
    Help --> Main : closeHelpMsg (Esc)

    Main --> Main : feedsLoadedMsg
    Main --> Main : feedsRefreshedMsg
    Main --> Main : requestRefreshMsg (r)

    Article --> Article : articleLoadedMsg
    Article --> Article : articleErrorMsg
```

### Bubble Tea Update Loop

Every interaction in the UI follows this cycle:

```mermaid
graph TB
    Event["User Input / Async Result<br/>(tea.Msg)"] --> Update["Model.Update(msg)"]

    Update --> Global{"Global handler?"}
    Global -->|"WindowSizeMsg"| Resize["Resize all sub-models"]
    Global -->|"feedsLoadedMsg"| FeedLoaded["main.setFeedResults()"]
    Global -->|"articleLoadedMsg"| ArtLoaded["articleView.showArticle()"]
    Global -->|"providerSelectedMsg"| Navigate["Switch screen + fetchFeedsCmd"]
    Global -->|"openArticleMsg"| NavArt["Switch screen + fetchArticleCmd"]
    Global -->|"No match"| Route{"Route by activeScreen"}

    Route -->|screenSplash| SplashUpdate["splash.Update(msg)"]
    Route -->|screenMain| MainUpdate["main.Update(msg)"]
    Route -->|screenArticle| ArticleUpdate["articleView.Update(msg)"]
    Route -->|screenHelp| HelpUpdate["help.Update(msg)"]

    Resize & FeedLoaded & ArtLoaded & Navigate & NavArt --> Return["Return (Model, tea.Cmd)"]
    SplashUpdate & MainUpdate & ArticleUpdate & HelpUpdate --> Return

    Return --> View["Model.View()"]
    View --> Render["Render to terminal"]
    Render -.->|"Next event"| Event
```

### Main View Layout

The main screen uses a two-panel layout with focus management:

```
┌──────────────────────────────────────────────────────┐
│                Knews - NZ News Reader                │  ← Header (titledBox)
├────────────────────┬─────────────────────────────────┤
│      Feeds         │         Articles                │
│  ┌──────────────┐  │  ┌───────────────────────────┐  │
│  │ ▸ Business   │  │  │ ▸ Article title here      │  │  ← feedList / articleList
│  │   Country    │  │  │     02 Apr 2026           │  │    (bubbles/list)
│  │   National   │  │  │   Another article         │  │
│  │   World      │  │  │     01 Apr 2026           │  │
│  │   ...        │  │  │   ...                     │  │
│  └──────────────┘  │  └───────────────────────────┘  │
│  (22 chars wide)   │  (remaining width)              │
├────────────────────┴─────────────────────────────────┤
│ Press ? for help │ q to quit                         │  ← Status bar
└──────────────────────────────────────────────────────┘
```

- **Tab / h / l**: Toggle focus between feed list and article list
- **j / k**: Navigate within the focused list
- **Enter**: Select feed (from feed list) or open article (from article list)
- Active panel border is highlighted in dark cyan; inactive is gray

### Message Types

| Message | Trigger | Effect |
|---------|---------|--------|
| `providerSelectedMsg` | Enter on splash | Load feeds, switch to main screen |
| `feedsLoadedMsg` | Async feed fetch completes | Populate feed results |
| `feedsRefreshedMsg` | Async refresh completes | Update feed results, preserve selection |
| `openArticleMsg` | Enter on article item | Show loading, switch to article screen |
| `articleLoadedMsg` | Async article fetch completes | Render article in viewport |
| `articleErrorMsg` | Async article fetch fails | Show error message |
| `showHelpMsg` | Press `?` | Switch to help screen |
| `closeArticleMsg` | Press `Esc` on article | Return to main screen |
| `closeHelpMsg` | Press `Esc` on help | Return to main screen |
| `requestRefreshMsg` | Press `r` | Re-fetch all feeds |

---

## Full Data Flow

End-to-end flow from startup to reading an article:

```mermaid
graph TB
    subgraph "1. Startup"
        Main["main()"] --> Wire["Wire FeedClient → FeedFetcher<br/>Wire RNZParser → Registry → ArticleFetcher"]
        Wire --> UI["Create Model, run bubbletea"]
    end

    subgraph "2. Provider Selection"
        UI --> Splash["Splash screen<br/>User picks 'RNZ'"]
        Splash --> Config["config.GetProvider('RNZ')<br/>→ 8 feed URLs + ordering"]
    end

    subgraph "3. Feed Fetching"
        Config --> FetchCmd["fetchFeedsCmd<br/>(30s timeout)"]
        FetchCmd --> Pool["FeedFetcher worker pool<br/>up to 5 concurrent HTTP GETs"]
        Pool --> XMLParse["ParseXmlResponse<br/>for each feed"]
        XMLParse --> Results["[]FeedResult<br/>with Rss → Channel → []Item"]
    end

    subgraph "4. Feed Display"
        Results --> MainView["mainModel<br/>Feed list (left) + Article list (right)"]
        MainView --> Browse["User browses feeds & articles<br/>Tab/h/l to switch panels, j/k to navigate"]
    end

    subgraph "5. Article Reading"
        Browse --> Select["User presses Enter on article"]
        Select --> ArtCmd["fetchArticleCmd<br/>(15s timeout)"]
        ArtCmd --> ArtFetch["ArticleFetcher.Fetch()<br/>HTTP GET → HTML"]
        ArtFetch --> Parse["RNZParser.Parse()<br/>go-readability extraction"]
        Parse --> ArtView["articleModel viewport<br/>Title, author, date, content"]
    end

    style Main fill:#e8f4f8
    style ArtView fill:#e8f4f8
```
