package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	consts "github.com/ihor-sokoliuk/newsbot/configs"
	_ "github.com/mattn/go-sqlite3"
)

const channelSubscriptionsTableName = "ChannelSubscriptions"
const newsHistoryTableName = "NewsHistory"

// NewsBotDatabase represents a SQL Database
type NewsBotDatabase struct {
	sql.DB
}

// NewDatabase creates a sandart database instance
func NewDatabase() (*NewsBotDatabase, error) {
	db, err := sql.Open("sqlite3", consts.DatabaseFileName)
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS ` + channelSubscriptionsTableName + ` (
		ID INTEGER PRIMARY KEY, 
		ChatID INTEGER, 
		NewsID INTEGER
	)
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS ` + newsHistoryTableName + ` (
		NewsID INTEGER PRIMARY KEY,
		LastPublish TEXT
	)
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return &NewsBotDatabase{*db}, nil
}

// GetChannelSubscriptions returns a list of news that user was subscribed on
func GetChannelSubscriptions(db *NewsBotDatabase, chatID int64) ([]int64, error) {
	rows, err := db.Query(fmt.Sprintf(`
		SELECT 
			NewsID 
		FROM %v 
		WHERE 
			ChatID = %v`,
		channelSubscriptionsTableName, chatID))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var newsID int64
	var resultList = make([]int64, 0)

	for rows.Next() {
		err = rows.Scan(&newsID)
		if err != nil {
			return nil, err
		}
		resultList = append(resultList, newsID)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

// GetNewsSubscribers returns a list of users that were subscribed on a news
func GetNewsSubscribers(db *NewsBotDatabase, newsID int64) ([]int64, error) {
	rows, err := db.Query(fmt.Sprintf(`
		SELECT 
			ChatID 
		FROM %v 
		WHERE 
			NewsID = %v`,
		channelSubscriptionsTableName, newsID))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var chatID int64
	var resultList = make([]int64, 0)

	for rows.Next() {
		err = rows.Scan(&chatID)
		if err != nil {
			return nil, err
		}
		resultList = append(resultList, chatID)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

// AddNewsSubscriber subscribes a user on a news
func AddNewsSubscriber(db *NewsBotDatabase, chatID, newsID int64) error {
	_, err := db.Exec(fmt.Sprintf(`
		INSERT 
		INTO %v 
			(ChatID, NewsID) 
		values 
			(%v, %v)`,
		channelSubscriptionsTableName, chatID, newsID))
	return err
}

// DeleteNewsSubscriber unsubscribes a user from a news
func DeleteNewsSubscriber(db *NewsBotDatabase, chatID, newsID int64) error {
	_, err := db.Exec(fmt.Sprintf(`
		DELETE 
		FROM %v 
		WHERE 
			ChatID = %v AND 
			NewsID = %v`,
		channelSubscriptionsTableName, chatID, newsID))
	return err
}

// IfUserSubscribedOnNews checks if user was subscribed on a news
func IfUserSubscribedOnNews(db *NewsBotDatabase, chatID, newsID int64) (bool, error) {
	count, err := rowsCount(db, fmt.Sprintf(`
		SELECT 
			COUNT(*) 
		FROM %v 
		WHERE 
			ChatID = %v AND 
			NewsID = %v`,
		channelSubscriptionsTableName, chatID, newsID))
	if err == nil {
		return count > 0, nil
	}
	return false, err
}

// GetLastPublishOfNews returns last publish time of a news
func GetLastPublishOfNews(db *NewsBotDatabase, newsID int64) (*time.Time, error) {
	var dt string
	rows := db.QueryRow(fmt.Sprintf(`
		SELECT 
			LastPublish 
		FROM %v 
		WHERE 
			NewsID = %v`,
		newsHistoryTableName, newsID))
	err := rows.Scan(&dt)
	if err == nil {
		lastPublish, err := time.Parse(time.RFC3339, dt)
		return &lastPublish, err
	}
	lastPublish := time.Now().Add(-time.Hour * 24)
	return &lastPublish, fmt.Errorf("No last pub date for news #%v", newsID)
}

// SaveLastPublishOfNews save last publish time of a news
func SaveLastPublishOfNews(db *NewsBotDatabase, newsID int64, lastPublish time.Time) error {
	_, err := db.Exec(fmt.Sprintf(`
		REPLACE
		INTO %v 
			(NewsID, LastPublish) 
		values 
			(%v, "%v")`,
		newsHistoryTableName, newsID, lastPublish.Format(time.RFC3339)))
	return err
}

func rowsCount(db *NewsBotDatabase, query string) (int, error) {
	var count int
	row := db.QueryRow(query)
	err := row.Scan(&count)
	if err == nil {
		return count, nil
	} else if err != nil && strings.Contains(err.Error(), "sql: no rows in result set") {
		return 0, nil
	}
	return -1, err
}
