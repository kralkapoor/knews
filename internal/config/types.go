package config

type FeedConfiguration struct {
	Providers []Provider 
}

type Provider struct {
	Name string
	Feeds []string
}