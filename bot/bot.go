package bot

import (
	"fmt"
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
	Help        = "help"
	Subscribe   = "subscribe"
	Unsubscribe = "unsubscribe"
)

const MessageSenderPeriod = time.Second / 25

var messageChan = make(chan tgbotapi.MessageConfig)

var BotEnv *Env

type Env struct {
	Db      *database.NewsBotDatabase
	Logger  *logs.NewsBotLogger
	Configs *configs.Configs
}

func RunBot(env *Env) {
	validateEnvironmentVariable(env)
	BotEnv = env

	// Create Telegram Bot
	bot, err := tgbotapi.NewBotAPI(BotEnv.Configs.Token)
	BotEnv.Logger.HandlePanic(err)
	BotEnv.Logger.Info(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	// Run message sender thread
	go messageSender(bot)

	// Scan news list and run RSS fetching for each enabled news
	for _, newsRss := range BotEnv.Configs.NewsRss {
		if newsRss.IsEnabled {
			go scanningRssNews(newsRss)
		}
	}

	// Read channel updates, users' messages
	u := tgbotapi.NewUpdate(0)
	updates, err := bot.GetUpdatesChan(u)
	BotEnv.Logger.HandleError(err)
	for update := range updates {
		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		chatId := update.Message.Chat.ID
		command := update.Message.Command()
		BotEnv.Logger.Info(fmt.Sprintf("Recieved a command %v from %v", command, chatId))
		msg := tgbotapi.NewMessage(chatId, "")

		// Read a command
		if command == List {
			msg.Text = generateNewsSubscriptionList(chatId)
		} else if command == Help {
			msg.Text = "It's a " + configs.ProjectName + " bot\nType /list to view news list to subscribe on."
		} else if newsId, err := validateCommand(command, Subscribe); !BotEnv.Logger.HandleError(err) {
			BotEnv.Logger.Info(fmt.Sprintf("\nCommand: %v\nErr: %v\nNewsId: %v", command, err, newsId))
			msg.Text = subscribe(chatId, newsId)
		} else if newsId, err := validateCommand(command, Unsubscribe); !BotEnv.Logger.HandleError(err) {
			BotEnv.Logger.Info(fmt.Sprintf("\nCommand: %v\nErr: %v\nNewsId: %v", command, err, newsId))
			msg.Text = unsubscribe(chatId, newsId)
		} else {
			continue
		}

		// Send an answer
		msg.ParseMode = tgbotapi.ModeMarkdown
		messageChan <- msg
	}
}

func validateEnvironmentVariable(env *Env) {
	if env == nil || env.Db == nil || env.Configs == nil || env.Logger == nil {
		panic("Bot environment parameters validation failed")
	}
}

func messageSender(bot *tgbotapi.BotAPI) {
	for r := range messageChan {
		if r.Text != "" {
			_, err := bot.Send(r)
			BotEnv.Logger.HandleError(err)
			time.Sleep(MessageSenderPeriod)
		}
	}
}

func scanningRssNews(rssNews configs.RssNews) {
	BotEnv.Logger.Info(fmt.Sprintf("Started scanning news for %v-%v(%v)", rssNews.ID, rssNews.Name, rssNews.URL))
	lastUrl := ""
	for {
		fetchedRssNews, err := readRssNews(lastUrl, rssNews.URL)
		if !BotEnv.Logger.HandleError(err, rssNews.ID, rssNews.Name, rssNews.URL) && fetchedRssNews != nil && fetchedRssNews.Message != "" {
			messageToSend := fmt.Sprintf("*%v*\n\n%v", rssNews.Name, fetchedRssNews)
			lastUrl = fetchedRssNews.Url
			newsSubscribers, err := database.GetNewsSubscribers(BotEnv.Db, rssNews.ID)
			if !BotEnv.Logger.HandleError(err) {
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

func generateNewsSubscriptionList(channelId int64) string {
	newsIDs, err := database.GetChannelSubscriptions(BotEnv.Db, channelId)
	if BotEnv.Logger.HandleError(err) {
		return ""
	}
	var sb strings.Builder
m0:
	for _, rssNews := range BotEnv.Configs.NewsRss {
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

func validateCommand(command, botCommand string) (int64, error) {
	if i := strings.Index(command, botCommand); i == 0 {
		BotEnv.Logger.Info(fmt.Sprintf("\nCommand: %v\nErr: %v\nNewsId: %v", command, nil, i))
		newsId, err := strconv.ParseInt(command[len(botCommand):], 10, 64)
		BotEnv.Logger.Info(fmt.Sprintf("\nCommand: %v\nErr: %v\nNewsId: %v", command, err, newsId))
		if err != nil {
			BotEnv.Logger.Info(fmt.Sprintf("\nCommand: %v\nErr: %v\nNewsId: %v", command, err, -1))
			return -1, err
		} else {
			BotEnv.Logger.Info(fmt.Sprintf("\nCommand: %v\nErr: %v\nNewsId: %v", command, err, newsId))
			if ifNewsIsAvailable(newsId) {
				BotEnv.Logger.Info(fmt.Sprintf("\nCommand: %v\nErr: %v\nNewsId: %v", command, err, newsId))
				return newsId, nil
			}
		}
	}
	BotEnv.Logger.Info(fmt.Sprintf("\nCommand: %v\nErr: %v\nNewsId: %v", command, nil, -1))
	return -1, nil
}

func ifNewsIsAvailable(newsId int64) bool {
	for _, newsRss := range BotEnv.Configs.NewsRss {
		if newsRss.ID == newsId {
			return true
		}
	}

	return false
}

func subscribe(chatId, newsId int64) string {
	BotEnv.Logger.Info("Subscribe #1")
	ifSubscribed, err := database.IfUserSubscribedOnNews(BotEnv.Db, chatId, newsId)
	BotEnv.Logger.Info("Subscribe #2")
	if !BotEnv.Logger.HandleError(err) && ifSubscribed {
		BotEnv.Logger.Info("Subscribe #3")
		return fmt.Sprintf("You are already subsribed on newsRss #%v", newsId)
	} else if !ifSubscribed {
		BotEnv.Logger.Info("Subscribe #4")
		err = database.AddNewsSubscriber(BotEnv.Db, chatId, newsId)
		if !BotEnv.Logger.HandleError(err) {
			BotEnv.Logger.Info("Subscribe #5")
			return fmt.Sprintf("Subscribed on newsRss #%v", newsId)
		}
	}
	BotEnv.Logger.Info("Subscribe #6")
	return ""
}

func unsubscribe(chatId, newsId int64) string {
	BotEnv.Logger.Info("Unsubscribe #1")
	ifSubscribed, err := database.IfUserSubscribedOnNews(BotEnv.Db, chatId, newsId)
	BotEnv.Logger.Info("Unsubscribe #2")
	if !BotEnv.Logger.HandleError(err) && ifSubscribed {
		BotEnv.Logger.Info("Unsubscribe #3")
		err = database.DeleteNewsSubscriber(BotEnv.Db, chatId, newsId)
		if !BotEnv.Logger.HandleError(err) {
			BotEnv.Logger.Info("Unsubscribe #4")
			return fmt.Sprintf("Unsubscribed from newsRss #%v", newsId)
		}
	} else if !ifSubscribed {
		BotEnv.Logger.Info("Unsubscribe #5")
		return fmt.Sprintf("You are already unsubsribed from newsRss #%v", newsId)
	}
	BotEnv.Logger.Info("Unsubscribe #6")
	return ""
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
