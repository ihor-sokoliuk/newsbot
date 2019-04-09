package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ihor-sokoliuk/newsbot/configs"
	"github.com/ihor-sokoliuk/newsbot/database"
	"github.com/ihor-sokoliuk/newsbot/logs"
)

const (
	listCmd        = "list"
	helpCmd        = "help"
	startCmd       = "start"
	subscribeCmd   = "subscribe"
	unsubscribeCmd = "unsubscribe"
)

const messageSenderPeriod = time.Second / 25

var messageChan = make(chan tgbotapi.MessageConfig)

var botEnv *Env

// Env bot environment structure
type Env struct {
	Db      *database.NewsBotDatabase
	Logger  *logs.NewsBotLogger
	Configs *configs.Configs
}

//RunBot function to run a bot
func RunBot(env *Env) {
	validateEnvironmentVariable(env)
	botEnv = env

	botEnv.Logger.Info(fmt.Sprintf("Tocken: %v", botEnv.Configs.Token))
	// Create Telegram Bot
	bot, err := tgbotapi.NewBotAPI(botEnv.Configs.Token)
	botEnv.Logger.HandlePanic(err)
	botEnv.Logger.Info(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	// Run message sender thread
	go messageSender(bot)

	// Scan news list and run RSS fetching for each enabled news
	for _, newsRss := range botEnv.Configs.RssNewsList {
		if newsRss.IsEnabled {
			go scanningRssNews(newsRss)
		}
	}

	// Read channel updates, users' messages
	u := tgbotapi.NewUpdate(0)
	updates, err := bot.GetUpdatesChan(u)
	botEnv.Logger.HandleError(err)
	for update := range updates {
		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		chatID := update.Message.Chat.ID
		command := update.Message.Command()
		msg := tgbotapi.NewMessage(chatID, "")

		// Read a command
		if command == listCmd {
			msg.Text = generateNewsSubscriptionList(chatID)
		} else if command == helpCmd || command == startCmd {
			msg.Text = "It's " + configs.ProjectName + ".\nType /list to view news list to subscribe on."
		} else if newsID, err := validateCommand(command, subscribeCmd); !botEnv.Logger.HandleError(err) && newsID > 0 {
			msg.Text = subscribeUser(chatID, newsID) + "\n\n" + generateNewsSubscriptionList(chatID)
		} else if newsID, err := validateCommand(command, unsubscribeCmd); !botEnv.Logger.HandleError(err) && newsID > 0 {
			msg.Text = unsubscribeUser(chatID, newsID) + "\n\n" + generateNewsSubscriptionList(chatID)
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
			botEnv.Logger.HandleError(err)
			time.Sleep(messageSenderPeriod)
		}
	}
}

func scanningRssNews(rssNews configs.RssNews) {
	botEnv.Logger.Info(fmt.Sprintf("Started scanning news for %v-%v(%v)", rssNews.ID, rssNews.Name, rssNews.URL))
	lastPublishDate, err := database.GetLastPublishOfNews(botEnv.Db, rssNews.ID)
	botEnv.Logger.HandleError(err)
	lastNewsURL := ""
	for {
		fetchedRssNews, err := readRssNews(rssNews.URL)
		if !botEnv.Logger.HandleError(err, rssNews.ID, rssNews.Name, rssNews.URL) && fetchedRssNews != nil && fetchedRssNews.Message != "" && fetchedRssNews.PublishDate.After(*lastPublishDate) && lastNewsURL != fetchedRssNews.URL {
			messageToSend := fmt.Sprintf("*%v*: %v", rssNews.Name, fetchedRssNews)
			messageToSend += fmt.Sprintf("\n\nDon't like this news site? /unsubscribe%v", rssNews.ID)
			newsSubscribers, err := database.GetNewsSubscribers(botEnv.Db, rssNews.ID)
			if !botEnv.Logger.HandleError(err) {
				for _, channelID := range newsSubscribers {
					msg := tgbotapi.NewMessage(channelID, messageToSend)
					msg.ParseMode = tgbotapi.ModeMarkdown
					messageChan <- msg
				}
			}
			lastNewsURL = fetchedRssNews.URL
			lastPublishDate = fetchedRssNews.PublishDate
			err = database.SaveLastPublishOfNews(botEnv.Db, rssNews.ID, *lastPublishDate)
			botEnv.Logger.HandleError(err)
		}
		time.Sleep(time.Minute)
	}
}

func generateNewsSubscriptionList(channelID int64) string {
	newsIDs, err := database.GetChannelSubscriptions(botEnv.Db, channelID)
	if botEnv.Logger.HandleError(err) {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("List of subscriptions:\n")
m0:
	for _, rssNews := range botEnv.Configs.RssNewsList {
		if rssNews.IsEnabled {
			for _, newsID := range newsIDs {
				if newsID == rssNews.ID {
					sb.WriteString(fmt.Sprintf("- %v (subscribed)\n  /unsubscribe%v\n", rssNews.Name, rssNews.ID))
					continue m0
				}
			}
			sb.WriteString(fmt.Sprintf("- %v (*unsubscribed*)\n  /subscribe%v\n", rssNews.Name, rssNews.ID))
		}
	}
	return sb.String()
}

func validateCommand(command, botCommand string) (int64, error) {
	if i := strings.Index(command, botCommand); i == 0 {
		newsID, err := strconv.ParseInt(command[len(botCommand):], 10, 64)
		if err == nil && ifNewsIsAvailable(newsID) {
			return newsID, nil
		}
	}
	return -1, nil
}

func ifNewsIsAvailable(newsID int64) bool {
	for _, newsRss := range botEnv.Configs.RssNewsList {
		if newsRss.ID == newsID && newsRss.IsEnabled {
			return true
		}
	}
	return false
}

func subscribeUser(chatID, newsID int64) string {
	ifSubscribed, err := database.IfUserSubscribedOnNews(botEnv.Db, chatID, newsID)
	if !botEnv.Logger.HandleError(err) {
		if ifSubscribed {
			return fmt.Sprintf("You are already subsribed on %v", getNewsNameByID(newsID))
		}
		err = database.AddNewsSubscriber(botEnv.Db, chatID, newsID)
		if !botEnv.Logger.HandleError(err) {
			return fmt.Sprintf("Subscribed on %v", getNewsNameByID(newsID))
		}
	}
	return ""
}

func unsubscribeUser(chatID, newsID int64) string {
	ifSubscribed, err := database.IfUserSubscribedOnNews(botEnv.Db, chatID, newsID)
	if !botEnv.Logger.HandleError(err) {
		if ifSubscribed {
			err = database.DeleteNewsSubscriber(botEnv.Db, chatID, newsID)
			if !botEnv.Logger.HandleError(err) {
				return fmt.Sprintf("Unsubscribed from %v", getNewsNameByID(newsID))
			}
		} else {
			return fmt.Sprintf("You are already unsubsribed from #%v", getNewsNameByID(newsID))
		}
	}
	return ""
}

func getNewsNameByID(id int64) string {
	for _, news := range botEnv.Configs.RssNewsList {
		if news.ID == id {
			return news.Name
		}
	}
	return fmt.Sprintf("news %v", id)
}
