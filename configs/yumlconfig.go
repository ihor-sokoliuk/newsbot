package configs

type RssNews struct {
	Name string
	URL  string
	IsEnabled bool
}

type Configs struct {
	Token   string    `token`
	RssNews []RssNews `newslist`
}


