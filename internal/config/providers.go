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

var providers = []Provider{
	{
		Name:        "RNZ",
		Description: "Radio New Zealand",
		Hosts:       []string{"rnz.co.nz"},
		FullText:    true,
		Feeds: []Feed{
			{"Business", "https://www.rnz.co.nz/rss/business.xml"},
			// {"Country", "https://www.rnz.co.nz/rss/country.xml"},
			{"Environment", "https://www.rnz.co.nz/rss/environment.xml"},
			{"Identity", "https://www.rnz.co.nz/rss/identity.xml"},
			{"In Depth", "https://www.rnz.co.nz/rss/in-depth.xml"},
			// {"Indo NZ", "https://www.rnz.co.nz/rss/indonz.xml"},
			{"Media and Tech", "https://www.rnz.co.nz/rss/media-technology.xml"},
			{"National", "https://www.rnz.co.nz/rss/national.xml"},
			{"Political", "https://www.rnz.co.nz/rss/political.xml"},
			// {"Sport", "https://www.rnz.co.nz/rss/sport.xml"},
			// {"Te Ao Māori", "https://www.rnz.co.nz/rss/te-manu-korihi.xml"},
			{"World", "https://www.rnz.co.nz/rss/world.xml"},
		},
	},
	{
		// Atom feeds. Article pages are rendered client-side, so there's no
		// HTML for readability to extract.
		Name:        "Stuff",
		Description: "Stuff (summaries only)",
		Feeds: []Feed{
			{"NZ News", "https://www.stuff.co.nz/rss/nz-news"},
			{"Politics", "https://www.stuff.co.nz/rss/politics"},
			{"Business", "https://www.stuff.co.nz/rss/business"},
			{"World", "https://www.stuff.co.nz/rss/world-news"},
			{"Climate Change", "https://www.stuff.co.nz/rss/climate-change"},
			{"Sport", "https://www.stuff.co.nz/rss/sport"},
			{"Travel", "https://www.stuff.co.nz/rss/travel"},
			{"Opinion", "https://www.stuff.co.nz/rss/opinion"},
		},
	},
	{
		// Article bodies are hidden until a script runs a subscription check,
		// so readability only sees the first paragraph. Unknown section slugs
		// return an empty feed rather than a 404.
		Name:        "NZ Herald",
		Description: "The New Zealand Herald (summaries only)",
		Feeds: []Feed{
			{"New Zealand", "https://www.nzherald.co.nz/arc/outboundfeeds/rss/section/nz/?outputType=xml"},
			{"Politics", "https://www.nzherald.co.nz/arc/outboundfeeds/rss/section/nz/politics/?outputType=xml"},
			{"Business", "https://www.nzherald.co.nz/arc/outboundfeeds/rss/section/business/?outputType=xml"},
			{"World", "https://www.nzherald.co.nz/arc/outboundfeeds/rss/section/world/?outputType=xml"},
			{"Technology", "https://www.nzherald.co.nz/arc/outboundfeeds/rss/section/technology/?outputType=xml"},
			{"Sport", "https://www.nzherald.co.nz/arc/outboundfeeds/rss/section/sport/?outputType=xml"},
			{"Entertainment", "https://www.nzherald.co.nz/arc/outboundfeeds/rss/section/entertainment/?outputType=xml"},
			{"Lifestyle", "https://www.nzherald.co.nz/arc/outboundfeeds/rss/section/lifestyle/?outputType=xml"},
		},
	},
}

// Providers returns the available providers in display order.
func Providers() []Provider {
	return slices.Clone(providers)
}
