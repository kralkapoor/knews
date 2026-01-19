package config

func GetTestFeedConfiguration() FeedConfiguration {
	rnzProvider := Provider{
		Name: "RNZ",
		Feeds: []string{"http://rnz.co.nz/rss/political"}, 
	}

	return FeedConfiguration{Providers: []Provider{rnzProvider}}
}
