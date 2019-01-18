package database

import (
	"database/sql"
	"fmt"
	consts "github.com/ihor-sokoliuk/newsbot/configs"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

type NewsBotDatabase struct {
	sql.DB
}

func NewDatabase() (*NewsBotDatabase, error) {
	db, err := sql.Open("sqlite3", consts.DatabaseFileName)
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS ` + consts.ChannelSubscriptionsTableName + ` (
		ID INTEGER PRIMARY KEY, 
		ChatID INTEGER, 
		NewsID INTEGER
	)
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	//sqlStmt = `
	//CREATE TABLE IF NOT EXISTS ` + consts.ConfigTableName + ` (
	//	Setting TEXT PRIMARY KEY,
	//	Value TEXT
	//)
	//`
	//_, err = db.Exec(sqlStmt)
	//if err != nil {
	//	return nil, err
	//}

	//sqlStmt = `
	//CREATE TABLE IF NOT EXISTS ` + consts.NewsHistoryTableName + ` (
	//	Id INTEGER PRIMARY KEY,
	//	NewsRssUrl TEXT,
	//	NewsUrl TEXT
	//)
	//`
	//_, err = db.Exec(sqlStmt)
	//if err != nil {
	//	return nil, err
	//}

	return &NewsBotDatabase{*db}, nil
}

func GetChannelSubscriptions(db *NewsBotDatabase, chatID int64) ([]int64, error) {
	rows, err := db.Query(fmt.Sprintf(`
	SELECT 
		NewsID 
	FROM %v 
	WHERE 
		ChatID = %v`,
		consts.ChannelSubscriptionsTableName, chatID))
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
		consts.ChannelSubscriptionsTableName, newsID))
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
	_, err := db.Exec(fmt.Sprintf("INSERT INTO %v (ChatID, NewsID) values (%v, %v)", consts.ChannelSubscriptionsTableName, chatID, newsID))
	return err
}

func DeleteNewsSubscriber(db *NewsBotDatabase, chatID, newsID int64) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %v WHERE ChatID = %v AND NewsID = %v", consts.ChannelSubscriptionsTableName, chatID, newsID))
	return err
}

func IfUserSubscribedOnNews(db *NewsBotDatabase, chatID, newsID int64) (bool, error) {
	var count int
	rows := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE ChatID = %v AND NewsID = %v", consts.ChannelSubscriptionsTableName, chatID, newsID))
	err := rows.Scan(&count)
	if err == nil {
		return count > 0, nil
	} else if err != nil && strings.Contains(err.Error(), "sql: no rows in result set") {
		return false, nil
	}
	return false, err
}

//func IfNewsExistsInHistory(db *NewsBotDatabase, rssUrl, newsUrl string) (bool, error) {
//	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE NewsRssUrl = '%v' and NewsUrl = '%v'", consts.NewsHistoryTableName, rssUrl, newsUrl))
//	count := 0
//	err := row.Scan(&count)
//	if err != nil {
//		return false, err
//	}
//
//	if count != 0 {
//		return true, nil
//	}
//
//	return false, nil
//}
//
//func (db *NewsBotDatabase) AddNewsToNewsHistory(rssUrl, newsUrl string) error {
//	_, err := db.Exec(fmt.Sprintf("INSERT INTO %v (NewsRssUrl, NewsUrl) values ('%v', '%v')", consts.NewsHistoryTableName, rssUrl, newsUrl))
//	if err != nil {
//		return err
//	}
//
//	return db.CleanNewsHistoryTable(rssUrl)
//}
//
//func (db *NewsBotDatabase) CleanNewsHistoryTable(rssUrl string) error {
//	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE NewsRssUrl = '%v'", consts.NewsHistoryTableName, rssUrl))
//	count := 0
//	err := row.Scan(&count)
//	if err != nil {
//		return err
//	}
//
//	if count > 10 {
//		_, err = db.Exec(fmt.Sprintf("delete top(%v) from %v", count-10, consts.NewsHistoryTableName))
//	}
//	return err
//}
