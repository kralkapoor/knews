// Package config defines the news providers and feeds knews offers.
package config

import "slices"

// Feed is a named RSS feed.
type Feed struct {
	Name string
	URL  string
}

// Provider is a news organisation and the feeds knews offers for it.
type Provider struct {
	Name        string
	Description string

	// FullText reports whether article pages can be fetched and extracted
	// with readability. When false, knews shows the feed's summary and offers
	// to open the page in a browser instead.
	FullText bool

	// Hosts are the domains a FullText provider's articles are served from.
	Hosts []string

	// Feeds are listed in display order.
	Feeds []Feed
}

const (
	rnzBase    = "https://www.rnz.co.nz/rss/"
	stuffBase  = "https://www.stuff.co.nz/rss/"
	heraldBase = "https://www.nzherald.co.nz/arc/outboundfeeds/rss/section/"
)

func rnzFeed(name, slug string) Feed   { return Feed{name, rnzBase + slug + ".xml"} }
func stuffFeed(name, slug string) Feed { return Feed{name, stuffBase + slug} }
func heraldFeed(name, section string) Feed {
	return Feed{name, heraldBase + section + "/?outputType=xml"}
}

var providers = []Provider{
	{
		Name:        "RNZ",
		Description: "Radio New Zealand",
		Hosts:       []string{"rnz.co.nz"},
		FullText:    true,
		Feeds: []Feed{
			rnzFeed("Business", "business"),
			rnzFeed("Country", "country"),
			rnzFeed("Education", "education"),
			rnzFeed("Environment", "environment"),
			rnzFeed("Media and Tech", "media-technology"),
			rnzFeed("New Zealand", "national"),
			rnzFeed("Political", "political"),
			rnzFeed("Sport", "sport"),
			rnzFeed("Weather", "weather"),
			rnzFeed("World", "world"),
		},
	},
	{
		// Atom feeds. Article pages are rendered client-side, so there's no
		// HTML for readability to extract.
		Name:        "Stuff",
		Description: "Stuff (summaries only)",
		Feeds: []Feed{
			stuffFeed("NZ News", "nz-news"),
			stuffFeed("Politics", "politics"),
			stuffFeed("Business", "business"),
			stuffFeed("World", "world-news"),
			stuffFeed("Climate Change", "climate-change"),
			stuffFeed("Sport", "sport"),
			stuffFeed("Travel", "travel"),
			stuffFeed("Opinion", "opinion"),
		},
	},
	{
		// Article bodies are hidden until a script runs a subscription check,
		// so readability only sees the first paragraph. Unknown section slugs
		// return an empty feed rather than a 404.
		Name:        "NZ Herald",
		Description: "The New Zealand Herald (summaries only)",
		Feeds: []Feed{
			heraldFeed("New Zealand", "nz"),
			heraldFeed("Politics", "nz/politics"),
			heraldFeed("Business", "business"),
			heraldFeed("World", "world"),
			heraldFeed("Technology", "technology"),
			heraldFeed("Sport", "sport"),
			heraldFeed("Entertainment", "entertainment"),
			heraldFeed("Lifestyle", "lifestyle"),
		},
	},
}

// Providers returns the available providers in display order.
func Providers() []Provider {
	return slices.Clone(providers)
}
