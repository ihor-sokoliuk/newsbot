package bot

import (
	"fmt"
	"github.com/grokify/html-strip-tags-go"
	"github.com/mmcdole/gofeed"
	"strings"
	"time"
)

type PieceOfNews struct {
	Title       string
	Message     string
	Url         string
	PublishDate *time.Time
}

func readRssNews(newsRssUrl string) (*PieceOfNews, error) {
	feed, err := gofeed.NewParser().ParseURL(newsRssUrl)
	if err != nil {
		return nil, err
	}

	if len(feed.Items) > 0 {
		lastNews := feed.Items[0]
		return &PieceOfNews{lastNews.Title, lastNews.Description, lastNews.Link, lastNews.PublishedParsed}, nil
	}

	return nil, nil
}

// TODO: Improve formatting description for all types of news rss
func getMessageDescription(description string) string {
	description = strings.Replace(description, "\r\n", "", -1)
	if i := strings.Index(description, "'>"); i != -1 {
		description = description[i+2:]
	}
	if i := strings.Index(description, "<body>"); i != -1 {
		if j := strings.Index(description, "<p>"); j != -1 {
			description = description[i+6 : j-1]
		}
	}

	description = strip.StripTags(description)

	if len(description) > 2048 {
		description = description[:2048] + "..."
	}
	return description
}

func (n PieceOfNews) String() string {
	// Message with news
	return fmt.Sprintf("*%v*\n\n%v\n\n[URL](%v)", n.Title, getMessageDescription(n.Message), n.Url)
}
