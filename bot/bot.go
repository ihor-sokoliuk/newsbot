package bot

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ihor-sokoliuk/newsbot/configs"
	"github.com/ihor-sokoliuk/newsbot/database"
	"github.com/ihor-sokoliuk/newsbot/logs"
	"strconv"
	"strings"
	"time"
)

const (
	List        = "list"
	Subscribe   = "subscribe"
	Unsubscribe = "unsubscribe"
)

var messageChan chan tgbotapi.MessageConfig

type Env struct {
	Db      *database.NewsBotDatabase
	Logger  *logs.NewsBotLogger
	Configs *configs.Configs
}

func RunBot(env Env) {
	bot, err := tgbotapi.NewBotAPI(env.Configs.Token)
	env.Logger.HandlePanic(errors.New("BOT PANIC: " + err.Error()))
	messageChan = make(chan tgbotapi.MessageConfig)

	//bot.Debug = true

	env.Logger.Info(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	for _, newsRss := range env.Configs.NewsRss {
		if newsRss.IsEnabled {
			go scanningRssNews(newsRss, env)
		}
	}
	go messageSender(bot, env.Logger)

	u := tgbotapi.NewUpdate(0)

	updates, err := bot.GetUpdatesChan(u)
	env.Logger.HandleError(err)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}
		channelId := update.Message.Chat.ID
		env.Logger.Info(fmt.Sprintf("Recieved a command %v from %v", update.Message.Command(), update.Message.Chat.ID))
		msg := tgbotapi.NewMessage(channelId, "")
		// Extract the command from the Message.
		if update.Message.Command() == List {
			msg.Text = generateNewsSubscriptionList(env, channelId)
		} else if i := strings.Index(update.Message.Command(), Subscribe); i == 0 {
			newsNumber, err := strconv.Atoi(update.Message.Command()[len(Subscribe):])
			if !env.Logger.HandleError(err) && ifNewsIsAvailable(env.Configs.NewsRss, newsNumber) {
				err = database.AddNewsSubscriber(env.Db, channelId, newsNumber)
				msg.Text = fmt.Sprintf("Subscribed on newsRss #%v", newsNumber)
			}
		} else if i := strings.Index(update.Message.Command(), Unsubscribe); i == 0 {
			newsNumber, err := strconv.Atoi(update.Message.Command()[len(Unsubscribe):])
			if !env.Logger.HandleError(err) && ifNewsIsAvailable(env.Configs.NewsRss, newsNumber) {
				err = database.DeleteNewsSubscriber(env.Db, channelId, newsNumber)
				msg.Text = fmt.Sprintf("Unsubscribed from newsRss #%v", newsNumber)
			}
		} else {
			msg.Text = "It's a " + configs.ProjectName + " bot\nType /list to view news list."
		}
		msg.ParseMode = tgbotapi.ModeMarkdown
		messageChan <- msg
	}
}

func scanningRssNews(rssNews configs.RssNews, env Env) {
	env.Logger.Info(fmt.Sprintf("Started scanning news for %v-%v(%v)", rssNews.ID, rssNews.Name, rssNews.URL))
	lastUrl := ""
	for {
		fetchedRssNews, err := readRssNews(lastUrl, rssNews.URL)
		if !env.Logger.HandleError(err) && fetchedRssNews != nil && fetchedRssNews.Message != "" {
			messageToSend := fmt.Sprintf("*%v*\n\n%v", rssNews.Name, fetchedRssNews)
			lastUrl = fetchedRssNews.Url
			newsSubscribers, err := database.NewsSubscribers(env.Db, rssNews.ID)
			if !env.Logger.HandleError(err) {
				for _, channelId := range newsSubscribers {
					msg := tgbotapi.NewMessage(channelId, messageToSend)
					msg.ParseMode = tgbotapi.ModeMarkdown
					messageChan <- msg
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func messageSender(bot *tgbotapi.BotAPI, logger *logs.NewsBotLogger) {
	period := time.Second / 25
	for r := range messageChan {
		if r.Text != "" {
			_, err := bot.Send(r)
			logger.HandleError(err)
			time.Sleep(period)
		}
	}
}

func generateNewsSubscriptionList(env Env, channelId int64) string {
	newsIDs, err := database.ChannelSubscriptions(env.Db, channelId)
	if env.Logger.HandleError(err) {
		return ""
	}
	var sb strings.Builder
m0:
	for _, rssNews := range env.Configs.NewsRss {
		for _, newsID := range newsIDs {
			if newsID == rssNews.ID {
				sb.WriteString(fmt.Sprintf("- %v (subscribed)\n  /unsubscribe%v\n", rssNews.Name, rssNews.ID))
				continue m0
			}
		}
		sb.WriteString(fmt.Sprintf("- %v (*unsubscribed*)\n  /subscribe%v\n", rssNews.Name, rssNews.ID))
	}
	return sb.String()
}

func ifNewsIsAvailable(newsRssList []configs.RssNews, newsId int) bool {
	for _, newsRss := range newsRssList {
		if newsRss.ID == newsId {
			return true
		}
	}

	return false
}

//func saveHotNewsSubscription(channelIdToSave int64) {
//	err := database.WriteHotNewsSubscription(channelIdToSave, true)
//	Logger.HandleError(err)
//}
//
//func readHotNewsSubscribers() []int64 {
//	channels, err := database.ReadAllChannelsSubscriptions()
//	Logger.HandleError(err)
//
//	resultChannelIdsList := make([]int64, 0, cap(channels))
//
//	for _, channel := range channels {
//		if channel.HotNewsSubscriptions {
//			resultChannelIdsList = append(resultChannelIdsList, channel.ChannelId)
//		}
//	}
//
//	return resultChannelIdsList
//}
