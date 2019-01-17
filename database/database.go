package database

import (
	"database/sql"
	"fmt"
	consts "github.com/ihor-sokoliuk/newsbot/configs"
	_ "github.com/mattn/go-sqlite3"
)

type ChannelSubscription struct {
	id        int64
	ChannelId int64
	NewsId    int64
}

type NewsBotDatabase struct {
	sql.DB
}

func NewDatabase() (*NewsBotDatabase, error) {
	db, err := sql.Open("sqlite3", consts.DatabaseFileName)
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS ` + consts.ChannelSubscriptionsTableName + ` (
		id INTEGER PRIMARY KEY, 
		ChannelId INTEGER, 
		NewsId INTEGER
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

func ChannelSubscriptions(db *NewsBotDatabase, channelID int64) ([]int, error) {

	rows, err := db.Query(fmt.Sprintf("SELECT NewsId FROM %v where ChannelId = %v", consts.ChannelSubscriptionsTableName, channelID))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var newsId int
	var resultList = make([]int, 0)

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

func NewsSubscribers(db *NewsBotDatabase, newsID int) ([]int64, error) {
	rows, err := db.Query(fmt.Sprintf(`
	SELECT 
		ChannelId 
	FROM %v 
	WHERE 
		NewsId = %v`,
		consts.ChannelSubscriptionsTableName, newsID))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var channelId int64
	var resultList = make([]int64, 0, 1)

	for rows.Next() {
		err = rows.Scan(&channelId)
		if err != nil {
			return nil, err
		}
		resultList = append(resultList, channelId)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func AddNewsSubscriber(db *NewsBotDatabase, channelID int64, newsID int) error {
	_, err := db.Exec(fmt.Sprintf("INSERT INTO %v (ChannelId, NewsId) values (%v, %v)", consts.ChannelSubscriptionsTableName, channelID, newsID))
	return err
}

func DeleteNewsSubscriber(db *NewsBotDatabase, channelID int64, newsID int) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM %v WHERE ChannelId = %v AND NewsId = %v", consts.ChannelSubscriptionsTableName, channelID, newsID))
	return err
}

//
//func ReadChannelSubscription(db *NewsBotDatabase, channelID, newsID int) ([]ChannelSubscription, error) {
//	rows, err := db.Query(fmt.Sprintf(`
//	SELECT
//		ChannelId,
//		NewsId
//	FROM %v
//	WHERE
//		ChannelId = %v AND
//		NewsId = %v`,
//		consts.ChannelSubscriptionsTableName,
//		channelID,
//		newsID))
//	if err != nil {
//		return nil, err
//	}
//
//	defer rows.Close()
//
//	var channelSub ChannelSubscription
//	var resultList = make([]ChannelSubscription, 0)
//
//	for rows.Next() {
//		err = rows.Scan(&channelSub.ChannelId, &channelSub.NewsId)
//		if err != nil {
//			return nil, err
//		}
//		resultList = append(resultList, channelSub)
//	}
//	err = rows.Err()
//	if err != nil {
//		return nil, err
//	}
//
//	return resultList, nil
//}

//
//func WriteNewsSubscription(channelId, newsId int) error {
//
//}
//
//func WriteCurrencySubscription(channelId int, currencySubscription bool) error {
//	channelSub, err := ChannelSubscriptions(channelId)
//	if logs.HandleError(err) {
//		channelSub = ChannelSubscription{ChannelId: channelId}
//	}
//
//	if channelSub.CurrencySubscriptions == currencySubscription {
//		return nil
//	}
//
//	channelSub.CurrencySubscriptions = currencySubscription
//
//	return WriteChannelSubscriptions(channelSub)
//}

//func GetConfigByName(db *NewsBotDatabase, setting string) (value string, err error) {
//	row := db.QueryRow(fmt.Sprintf("SELECT Value FROM %v WHERE Setting = '%v'", consts.ConfigTableName, setting))
//
//	err = row.Scan(&value)
//	if err != nil {
//		err = SetConfigByName(db, setting, "")
//		if err != nil {
//			return value, err
//		}
//		return GetConfigByName(db, setting)
//	}
//
//	return value, nil
//}
//
//func SetConfigByName(db *NewsBotDatabase, setting string, value string) error {
//	_, err := db.Exec(fmt.Sprintf("INSERT OR REPLACE INTO %v (Setting, Value) values ('%v', '%v')", consts.ConfigTableName, setting, value))
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
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
