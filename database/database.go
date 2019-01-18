package database

import (
	"database/sql"
	"errors"
	"fmt"
	consts "github.com/ihor-sokoliuk/newsbot/configs"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"time"
)

const ChannelSubscriptionsTableName = "ChannelSubscriptions"
const NewsHistoryTableName = "NewsHistory"

type NewsBotDatabase struct {
	sql.DB
}

func NewDatabase() (*NewsBotDatabase, error) {
	db, err := sql.Open("sqlite3", consts.DatabaseFileName)
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS ` + ChannelSubscriptionsTableName + ` (
		ID INTEGER PRIMARY KEY, 
		ChatID INTEGER, 
		NewsID INTEGER
	)
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	// TODO DON'T FORGET TO REMOVE IT
	sqlStmt = `DROP TABLE IF EXISTS ` + NewsHistoryTableName
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS ` + NewsHistoryTableName + ` (
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

func GetChannelSubscriptions(db *NewsBotDatabase, chatID int64) ([]int64, error) {
	rows, err := db.Query(fmt.Sprintf(`
		SELECT 
			NewsID 
		FROM %v 
		WHERE 
			ChatID = %v`,
		ChannelSubscriptionsTableName, chatID))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var newsId int64
	var resultList = make([]int64, 0)

	for rows.Next() {
		err = rows.Scan(&newsId)
		if err != nil {
			return nil, err
		}
		resultList = append(resultList, newsId)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func GetNewsSubscribers(db *NewsBotDatabase, newsID int64) ([]int64, error) {
	rows, err := db.Query(fmt.Sprintf(`
		SELECT 
			ChatID 
		FROM %v 
		WHERE 
			NewsID = %v`,
		ChannelSubscriptionsTableName, newsID))
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

func AddNewsSubscriber(db *NewsBotDatabase, chatID, newsID int64) error {
	_, err := db.Exec(fmt.Sprintf(`
		INSERT 
		INTO %v 
			(ChatID, NewsID) 
		values 
			(%v, %v)`,
		ChannelSubscriptionsTableName, chatID, newsID))
	return err
}

func DeleteNewsSubscriber(db *NewsBotDatabase, chatID, newsID int64) error {
	_, err := db.Exec(fmt.Sprintf(`
		DELETE 
		FROM %v 
		WHERE 
			ChatID = %v AND 
			NewsID = %v`,
		ChannelSubscriptionsTableName, chatID, newsID))
	return err
}

func RowsCount(db *NewsBotDatabase, query string) (int, error) {
	var count int
	rows := db.QueryRow(query)
	err := rows.Scan(&count)
	if err == nil {
		return count, nil
	} else if err != nil && strings.Contains(err.Error(), "sql: no rows in result set") {
		return 0, nil
	}
	return -1, err
}

func IfUserSubscribedOnNews(db *NewsBotDatabase, chatID, newsID int64) (bool, error) {
	count, err := RowsCount(db, fmt.Sprintf(`
		SELECT 
			COUNT(*) 
		FROM %v 
		WHERE 
			ChatID = %v AND 
			NewsID = %v`,
		ChannelSubscriptionsTableName, chatID, newsID))
	if err == nil {
		return count > 0, nil
	}
	return false, err
}

func GetLastPublishOfNews(db *NewsBotDatabase, newsID int64) (*time.Time, error) {
	var dt string
	rows := db.QueryRow(fmt.Sprintf(`
		SELECT 
			LastPublish 
		FROM %v 
		WHERE 
			NewsID = %v`,
		NewsHistoryTableName, newsID))
	err := rows.Scan(&dt)
	if err == nil {
		lastPublish, err := time.Parse(dt, time.RFC3339)
		return &lastPublish, err
	} else {
		lastPublish := time.Now().Add(-time.Hour * 24)
		return &lastPublish, errors.New(fmt.Sprintf("No last pub date for news #%v", newsID))
	}
}

func SaveLastPublishOfNews(db *NewsBotDatabase, newsId int64, lastPublish time.Time) error {
	_, err := db.Exec(fmt.Sprintf(`
		REPLACE
		INTO %v 
			(NewsID, LastPublish) 
		values 
			(%v, "%v")`,
		NewsHistoryTableName, newsId, lastPublish.Format(time.RFC3339)))
	return err
}
