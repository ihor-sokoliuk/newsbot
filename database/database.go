package database

import (
	"database/sql"
	"fmt"
	consts "github.com/ihor-sokoliuk/newsbot/configs"
	_ "github.com/mattn/go-sqlite3"
	"strings"
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

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS ` + NewsHistoryTableName + ` (
		Id INTEGER PRIMARY KEY,
		NewsID INTEGER,
		NewsUrl TEXT
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

func IfNewsWasBefore(db *NewsBotDatabase, newsId int64, newsLink string) (bool, error) {
	count, err := RowsCount(db, fmt.Sprintf(`
		SELECT 
			COUNT(*) 
		FROM %v 
		WHERE 
			NewsURL = %v AND 
			NewsID = %v`,
		NewsHistoryTableName, newsLink, newsId))
	if err == nil {
		return count > 0, nil
	}
	return false, err
}

func SaveNewsLink(db *NewsBotDatabase, newsId int64, newsLink string) error {
	_, err := db.Exec(fmt.Sprintf(`
		INSERT 
		INTO %v 
			(NewsURL, NewsID) 
		values 
			(%v, %v)`,
		NewsHistoryTableName, newsLink, newsId))
	return err
}

func CleanNewsHistoryTable(db *NewsBotDatabase, newsId int64) error {
	count, err := RowsCount(db, fmt.Sprintf(`
		SELECT 
			COUNT(*) 
		FROM %v 
		WHERE
			NewsID = %v`,
		NewsHistoryTableName, newsId))
	if err == nil && count > 5 {
		_, err = db.Exec(fmt.Sprintf(`
			DELETE TOP(%v) 
			FROM %v`,
			count-5, NewsHistoryTableName))
	}
	return err
}
