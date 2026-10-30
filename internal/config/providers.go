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

	// Hosts are the domains the provider's articles are served from.
	Hosts []string

	// FullText reports whether article pages can be fetched and extracted
	// with readability. Sites that render articles with JavaScript cannot.
	FullText bool

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
}

// Providers returns the available providers in display order.
func Providers() []Provider {
	return slices.Clone(providers)
}
