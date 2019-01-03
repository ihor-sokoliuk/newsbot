package main

import (
	"Practicing/news_rss_telegram_bot/helper"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
	"log"
	"strings"
	"time"
)

const token = "627296139:AAHfLC6biwiuLn5cIGNmbaRi5987bn8yCY4"
const hotNewsRssUrl = "https://24tv.ua/rss/tag/1792.xml"
const allNewsRssUrl = "https://24tv.ua/rss/all.xml"

//var channelId int64 = 684861449

//const channelName = "@tv_ua_rss_bot"

type PieceOfNews struct {
	Title   string
	Message string
	Url     string
}

func main() {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Print(err)
	}

	//bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	go scanningHotRssNews(bot)
	//go scanningAllRssNews(bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		err := helper.WriteChannelSubscriptions(helper.ChannelSubscription{ChannelId: update.Message.Chat.ID, HotNewsSubscriptions: true, AllNewsSubscriptions: true})
		if err != nil {
			log.Println(errors.Wrap(err, 1))
		} else {
			log.Printf("Chat ID was successfully registred: %v", update.Message.Chat.ID)
			log.Printf("Its message: %v", update.Message.Text)
		}
		//channelId = update.Message.Chat.ID
	}
}

func scanningHotRssNews(bot *tgbotapi.BotAPI) {
	rssNews, _ := readRssNews(getLastNewsUrl(hotNewsRssUrl), hotNewsRssUrl)
	for rss := range rssNews {
		if rss.Message == "" {
			continue
		}
		setLastNewsUrl(hotNewsRssUrl, rss.Url)
		hotNewsChannelsIds := readHotNewsSubscribers()
		for _, channelId := range hotNewsChannelsIds {
			msg := tgbotapi.NewMessage(channelId, fmt.Sprintf("Рубрика: горячие новости\n%v", rss))
			msg.ParseMode = tgbotapi.ModeMarkdown
			if _, err := bot.Send(msg); err != nil {
				log.Println(errors.Wrap(err, 1))
			}
		}
	}
}

func scanningAllRssNews(bot *tgbotapi.BotAPI) {

	rssNews, _ := readRssNews(getLastNewsUrl(allNewsRssUrl), allNewsRssUrl)
	for rss := range rssNews {
		if rss.Message == "" {
			continue
		}
		setLastNewsUrl(allNewsRssUrl, rss.Url)
		allNewsChannelsIds := readAllNewsSubscribers()
		for _, channelId := range allNewsChannelsIds {
			msg := tgbotapi.NewMessage(channelId, fmt.Sprintf("Рубрика: все новости\n%v", rss))
			msg.ParseMode = tgbotapi.ModeMarkdown
			if _, err := bot.Send(msg); err != nil {
				log.Println(errors.Wrap(err, 1))
			}
		}
	}
}

func readRssNews(lastNewsRss string, newsRssUrl string) (chan PieceOfNews, error) {
	ch := make(chan PieceOfNews)

	go func() {
		for {
			feed, err := gofeed.NewParser().ParseURL(newsRssUrl)
			log.Printf("Fetching news from %v...", newsRssUrl)
			if err != nil {
				log.Printf("Failed to get news from %v.\nError: %v.\nRetrying in 10 seconds...", newsRssUrl, err)
			}
			if lastNewsRss != feed.Items[0].Link {
				// TODO Add news history check
				log.Printf("Successfully fetched piece of news\nNews title: %v\nNews date: %v\n", feed.Items[0].Title, feed.Items[0].PublishedParsed)
				lastNewsRss = feed.Items[0].Link
				ch <- PieceOfNews{feed.Items[0].Title, feed.Items[0].Description, feed.Items[0].Link}
			}
			time.Sleep(time.Minute * 1)
		}
	}()

	return ch, nil
}

func (n PieceOfNews) String() string {
	return fmt.Sprintf("*%v*\n\n%v\n\n[Читати детальнiше новину...](%v)", n.Title, getMessageDescription(n.Message), n.Url)
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
	err := helper.WriteHotNewsSubscription(channelIdToSave, true)
	if err != nil {
		log.Println(errors.Wrap(err, 1))
	}
}

func saveAllNewsSubscription(channelIdToSave int64) {
	err := helper.WriteAllNewsSubscription(channelIdToSave, true)
	if err != nil {
		log.Println(errors.Wrap(err, 1))
	}
}

func readHotNewsSubscribers() []int64 {
	channels, err := helper.ReadAllChannelsSubscriptions()
	if err != nil {
		log.Println(errors.Wrap(err, 1))
		return nil
	}

	resultChannelIdsList := make([]int64, 0, cap(channels))

	for _, channel := range channels {
		if channel.HotNewsSubscriptions {
			resultChannelIdsList = append(resultChannelIdsList, channel.ChannelId)
		}
	}

	return resultChannelIdsList
}

func readAllNewsSubscribers() []int64 {
	channels, err := helper.ReadAllChannelsSubscriptions()
	if err != nil {
		log.Println(errors.Wrap(err, 1))
		return nil
	}

	resultChannelIdsList := make([]int64, 0, cap(channels))

	for _, channel := range channels {
		if channel.AllNewsSubscriptions {
			resultChannelIdsList = append(resultChannelIdsList, channel.ChannelId)
		}
	}

	return resultChannelIdsList
}

func getLastNewsUrl(newsRssUrl string) string {
	//lastUrl, err := helper.GetConfigByName("Last_News_Url_" + strings.Replace(newsRssUrl, "/", "//", -1))
	lastUrl, err := helper.GetConfigByName("Last_News_Url_" + newsRssUrl)
	if err != nil {
		log.Println(errors.Wrap(err, 1))
		return ""
	}
	return lastUrl
}

func setLastNewsUrl(newsRssUrl string, lastNewsUrl string) {
	err := helper.SetConfigByName("Last_News_Url_"+newsRssUrl, lastNewsUrl)
	if err != nil {
		log.Println(errors.Wrap(err, 1))
	}
}
