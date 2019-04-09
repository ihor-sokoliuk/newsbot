package bot

import (
	"fmt"
	"strings"
	"time"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/mmcdole/gofeed"
)

type pieceOfNews struct {
	Title       string
	Message     string
	URL         string
	PublishDate *time.Time
}

func readRssNews(newsRssURL string) (*pieceOfNews, error) {
	feed, err := gofeed.NewParser().ParseURL(newsRssURL)
	if err != nil {
		return nil, err
	}

	if len(feed.Items) > 0 {
		lastNews := feed.Items[0]
		if lastNews.PublishedParsed != nil {
			return &pieceOfNews{lastNews.Title, lastNews.Description, lastNews.Link, lastNews.PublishedParsed}, nil
		} else if feed.UpdatedParsed != nil {
			return &pieceOfNews{lastNews.Title, lastNews.Description, lastNews.Link, feed.UpdatedParsed}, nil
		}
	}

	return nil, nil
}

// TODO: Improve formatting description for all types of news rss
func getMessageDescription(description string) string {
	i := strings.Index(description, "\r\n\r\n")
	for i >= 0 {
		description = strings.Replace(description, "\r\n\r\n", "\r\n", -1)
		i = strings.Index(description, "\r\n\r\n")
	}
	i = strings.Index(description, "\r\r")
	for i >= 0 {
		description = strings.Replace(description, "\r\r", "\r", -1)
		i = strings.Index(description, "\r\r")
	}
	i = strings.Index(description, "\n\n")
	for i >= 0 {
		description = strings.Replace(description, "\n\n", "\n", -1)
		i = strings.Index(description, "\n\n")
	}

	if i := strings.Index(description, "'>"); i != -1 {
		description = description[i+2:]
	}

	if i := strings.Index(description, "<body>"); i != -1 {
		if j := strings.Index(description, "<p>"); j != -1 {
			description = description[i+6 : j-1]
		}
	}

	if i := strings.Index(description, "<![CDATA["); i != -1 {
		if j := strings.Index(description, "]]>"); j != -1 {
			description = description[i+9 : j-1]
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

func (n pieceOfNews) String() string {
	// Message with news
	return fmt.Sprintf("[*%v*](%v)\n\n%v\n", n.Title, n.URL, getMessageDescription(n.Message))
}
