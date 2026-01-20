package config

// GetTestFeedConfiguration returns a test configuration for development purposes.
func GetTestFeedConfiguration() FeedConfiguration {
	return FeedConfiguration{Providers: []Provider{Providers["RNZ"]}}
}
