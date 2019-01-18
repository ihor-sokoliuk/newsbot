package configs

type RssNews struct {
	Name      string `Name`
	URL       string `URL`
	IsEnabled bool   `IsEnabled`
	ID        int64  `ID`
}

type Configs struct {
	Token   string    `Token`
	NewsRss []RssNews `NewsList`
}
