package configs

// RssNews is a struct to contain an info about News RSS page
type RssNews struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	IsEnabled bool   `yaml:"isEnabled"`
	ID        int64  `yaml:"id"`
}

// Configs struct contains all the program configuration
type Configs struct {
	Token       string    `yaml:"token"`
	RssNewsList []RssNews `yaml:"rssNewsList"`
}
