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
		if lastNews.PublishedParsed != nil {
			return &PieceOfNews{lastNews.Title, lastNews.Description, lastNews.Link, lastNews.PublishedParsed}, nil
		} else if feed.UpdatedParsed != nil {
			return &PieceOfNews{lastNews.Title, lastNews.Description, lastNews.Link, feed.UpdatedParsed}, nil
		}
	}

	return nil, nil
}

// TODO: Improve formatting description for all types of news rss
func getMessageDescription(description string) string {

	description = strings.Replace(description, "\r\n\r\n", "\r\n", -1)
	description = strings.Replace(description, "\r\r", "\r", -1)
	description = strings.Replace(description, "\n\n", "\n", -1)

	if i := strings.Index(description, "'>"); i != -1 {
		description = description[i+2:]
	}

	if i := strings.Index(description, "<body>"); i != -1 {
		if j := strings.Index(description, "<p>"); j != -1 {
			description = description[i+6 : j-1]
		}
	}

	if i := strings.Index(description, "<!CDATA["); i != -1 {
		if j := strings.Index(description, "]>"); j != -1 {
			description = description[i+8 : j-1]
		}
	}

	if strings.HasPrefix(description, "\r\n") {
		description = description[2:]
	}
	if strings.HasPrefix(description, "\r") {
		description = description[1:]
	}
	if strings.HasPrefix(description, "\n") {
		description = description[1:]
	}
	if strings.HasSuffix(description, "\r\n") {
		description = description[:len(description)-2]
	}
	if strings.HasSuffix(description, "\r") {
		description = description[:len(description)-1]
	}
	if strings.HasSuffix(description, "\n") {
		description = description[:len(description)-1]
	}

	description = strip.StripTags(description)

	if len(description) > 2048 {
		description = description[:2048] + "..."
	}
	return description
}

func (n PieceOfNews) String() string {
	// Message with news
	return fmt.Sprintf("*%v*\n\n%v\n[Read more...](%v)", n.Title, getMessageDescription(n.Message), n.Url)
}
