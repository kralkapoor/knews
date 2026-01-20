package config

// Provider represents a news provider with its feeds and metadata.
type Provider struct {
	Name        string
	Description string
	Feeds       map[string]string
	FeedOrder   []string
}

// Providers is the registry of all available news providers.
var Providers = map[string]Provider{
	"RNZ": {
		Name:        "RNZ",
		Description: "Radio New Zealand",
		Feeds:       RNZFeeds,
		FeedOrder:   RNZFeedNames(),
	},
}

// ProviderNames returns the list of available provider names in display order.
func ProviderNames() []string {
	return []string{"RNZ"}
}

// GetProvider returns the Provider for a given name, or an empty Provider if not found.
func GetProvider(name string) Provider {
	return Providers[name]
}

// RNZFeeds defines the available RSS feed categories from Radio New Zealand.
// Each key is a human-readable category name, and the value is the feed URL.
var RNZFeeds = map[string]string {
	"Business":  		"https://www.rnz.co.nz/rss/business.xml",
	"Country":   		"https://www.rnz.co.nz/rss/country.xml",
	"Environment":		"https://www.rnz.co.nz/rss/environment.xml",
	"Identity":			"https://www.rnz.co.nz/rss/identity.xml",
	"In Depth":			"https://www.rnz.co.nz/rss/in-depth.xml",
	"Indo NZ":			"https://www.rnz.co.nz/rss/indonz.xml",
	// "Local Democracy": 	"https://www.rnz.co.nz/rss/ldr.xml",
	"Media and Tech": 	"https://www.rnz.co.nz/rss/media-technology.xml",
	"National": 		"https://www.rnz.co.nz/rss/national.xml",
	"Political": 		"https://www.rnz.co.nz/rss/political.xml",
	"Sport":     		"https://www.rnz.co.nz/rss/sport.xml",
	"Te Ao Maori":		"https://www.rnz.co.nz/rss/te-manu-korihi.xml",
	"World":    		"https://www.rnz.co.nz/rss/world.xml",
}

// RNZFeedNames returns the feed category names in display order.
func RNZFeedNames() []string {
	feedNames := []string{}

	for name := range RNZFeeds {
		feedNames = append(feedNames, name)
	}
	return feedNames
}

// GetRNZFeedURL returns the URL for a given feed name, or empty string if not found.
func GetRNZFeedURL(name string) string {
	return RNZFeeds[name]
}
