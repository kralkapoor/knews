package config

// FeedConfiguration holds the overall feed configuration.
type FeedConfiguration struct {
	Providers []Provider
}

// Provider represents a news provider with its feeds and metadata.
type Provider struct {
	Name        string
	Description string
	Feeds       map[string]string
	FeedOrder   []string
}
