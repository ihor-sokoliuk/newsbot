package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	consts "github.com/ihor-sokoliuk/newsbot/configs"
	"github.com/ihor-sokoliuk/newsbot/helpers"
	"github.com/ihor-sokoliuk/newsbot/logs"
	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"time"
)

var config Configs

type RssNews struct {
	Name string
	URL  string
	IsEnabled bool
}

type Configs struct {
	Token   string    `token`
	RssNews []RssNews `newslist`
}

type PieceOfNews struct {
	Title   string
	Message string
	Url     string
}

func init() {
	file, err := ioutil.ReadFile(consts.ConfigFileName)
	logs.HandlePanic(err)
	err = yaml.Unmarshal(file, &config)
	logs.HandlePanic(err)
}

func main() {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	logs.HandlePanic(err)

	//bot.Debug = true

	logs.Info(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	for _, news := range config.RssNews {
		if news.IsEnabled {
			go scanningRssNews(bot, news.URL, news.Name)
		}
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	logs.HandleError(err)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		err := helpers.WriteChannelSubscriptions(helpers.ChannelSubscription{ChannelId: update.Message.Chat.ID, HotNewsSubscriptions: true, AllNewsSubscriptions: true})

		if !logs.HandleError(err) {
			logs.Info(fmt.Sprintf("Chat ID was successfully registred: %v", update.Message.Chat.ID))
			logs.Info(fmt.Sprintf("Its message: %v", update.Message.Text))
		}
		//channelId = update.Message.Chat.ID
	}
}

func scanningRssNews(bot *tgbotapi.BotAPI, newsRssUrl, newsName string) {
	rssNews, _ := readRssNews(getLastNewsUrl(newsRssUrl), newsRssUrl)
	for rss := range rssNews {
		if rss.Message == "" {
			continue
		}
		messageToSend := fmt.Sprintf("*%v*\n\n%v", newsName, rss)
		setLastNewsUrl(newsRssUrl, rss.Url)
		hotNewsChannelsIds := readHotNewsSubscribers()
		for _, channelId := range hotNewsChannelsIds {
			msg := tgbotapi.NewMessage(channelId, messageToSend)
			msg.ParseMode = tgbotapi.ModeMarkdown
			_, err := bot.Send(msg)
			logs.HandleError(err)
		}
	}
}

func readRssNews(lastNewsRss string, newsRssUrl string) (chan PieceOfNews, error) {
	ch := make(chan PieceOfNews)

	go func() {
		for {
			logs.Info(fmt.Sprintf("Fetching news from %v...", newsRssUrl))
			feed, err := gofeed.NewParser().ParseURL(newsRssUrl)
			if logs.HandleError(err) {
				logs.Info(fmt.Sprintf("Failed to get news from %v.\nError: %v.\nRetrying in 10 seconds...", newsRssUrl, err))
			}
			if lastNewsRss != feed.Items[0].Link {
				// TODO Add news history check
				logs.Info(fmt.Sprintf("Successfully fetched piece of news\nNews title: %v\nNews date: %v\n", feed.Items[0].Title, feed.Items[0].PublishedParsed))
				lastNewsRss = feed.Items[0].Link
				ch <- PieceOfNews{feed.Items[0].Title, feed.Items[0].Description, feed.Items[0].Link}
			}
			time.Sleep(time.Minute * 1)
		}
	}()

	return ch, nil
}

func (n PieceOfNews) String() string {
	return fmt.Sprintf("*%v*\n\n%v\n\n[URL](%v)", n.Title, getMessageDescription(n.Message), n.Url)
}

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
	if len(description) > 2048 {
		description = description[:2048] + "..."
	}
	return description
}

func saveHotNewsSubscription(channelIdToSave int64) {
	err := helpers.WriteHotNewsSubscription(channelIdToSave, true)
	logs.HandleError(err)
}

func readHotNewsSubscribers() []int64 {
	channels, err := helpers.ReadAllChannelsSubscriptions()
	logs.HandleError(err)

	resultChannelIdsList := make([]int64, 0, cap(channels))

	for _, channel := range channels {
		if channel.HotNewsSubscriptions {
			resultChannelIdsList = append(resultChannelIdsList, channel.ChannelId)
		}
	}

	return resultChannelIdsList
}

func getLastNewsUrl(newsRssUrl string) string {
	lastUrl, err := helpers.GetConfigByName("Last_News_Url_" + newsRssUrl)
	if logs.HandleError(err) {
		return ""
	}
	return lastUrl
}

func setLastNewsUrl(newsRssUrl string, lastNewsUrl string) {
	err := helpers.SetConfigByName("Last_News_Url_"+newsRssUrl, lastNewsUrl)
	logs.HandleError(err)
}
