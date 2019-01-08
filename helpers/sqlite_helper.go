package helpers

import (
	"database/sql"
	"fmt"
	consts "github.com/ihor-sokoliuk/newsbot/configs"
	"github.com/ihor-sokoliuk/newsbot/logs"
	_ "github.com/mattn/go-sqlite3"
)

type ChannelSubscription struct {
	ChannelId             int64
	AllNewsSubscriptions  bool
	HotNewsSubscriptions  bool
	WeatherSubscriptions  bool
	CurrencySubscriptions bool
}

func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", consts.DatabaseFileName)
	if err != nil {
		return nil, err
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS ` + consts.UsersTableName + ` (
		ChannelId INTEGER PRIMARY KEY, 
		AllNewsSubscription INTEGER, 
		HotNewsSubscription INTEGER, 
		WeatherSubscription INTEGER, 
		CurrencySubscription INTEGER
	)
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS ` + consts.ConfigTableName + ` (
		Setting TEXT PRIMARY KEY,  
		Value TEXT
	)
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	sqlStmt = `
	CREATE TABLE IF NOT EXISTS ` + consts.NewsHistoryTableName + ` (
		Id INTEGER PRIMARY KEY,
		NewsRssUrl TEXT,  
		NewsUrl TEXT
	)
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ReadAllChannelIds() ([]int64, error) {
	db, err := initDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT ChannelId FROM %v", consts.UsersTableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channelId int64
	var resultIdsList = make([]int64, 0)

	for rows.Next() {
		err = rows.Scan(&channelId)
		if err != nil {
			return nil, err
		}
		resultIdsList = append(resultIdsList, channelId)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resultIdsList, nil
}

func ReadAllChannelsSubscriptions() ([]ChannelSubscription, error) {
	db, err := initDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT ChannelId, AllNewsSubscription, HotNewsSubscription, WeatherSubscription, CurrencySubscription FROM %v", consts.UsersTableName))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var channelSub ChannelSubscription
	var resultList = make([]ChannelSubscription, 0)

	for rows.Next() {
		err = rows.Scan(&channelSub.ChannelId, &channelSub.AllNewsSubscriptions, &channelSub.HotNewsSubscriptions, &channelSub.WeatherSubscriptions, &channelSub.CurrencySubscriptions)
		if err != nil {
			return nil, err
		}
		resultList = append(resultList, channelSub)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func ReadChannelSubscriptions(channelId int64) (channelSub ChannelSubscription, err error) {
	db, err := initDatabase()
	if err != nil {
		return channelSub, err
	}
	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("SELECT ChannelId, AllNewsSubscription, HotNewsSubscription, WeatherSubscription, CurrencySubscription FROM %v WHERE ChannelId = %v", consts.UsersTableName, channelId))

	err = row.Scan(&channelSub.ChannelId, &channelSub.AllNewsSubscriptions, &channelSub.HotNewsSubscriptions, &channelSub.WeatherSubscriptions, &channelSub.CurrencySubscriptions)
	if err != nil {
		return channelSub, err
	}

	return channelSub, nil
}

func WriteAllNewsSubscription(channelId int64, allNewsSubscription bool) error {
	channelSub, err := ReadChannelSubscriptions(channelId)
	if logs.HandleError(err) {
		channelSub = ChannelSubscription{ChannelId: channelId}
	}

	if channelSub.AllNewsSubscriptions == allNewsSubscription {
		return nil
	}

	channelSub.AllNewsSubscriptions = allNewsSubscription

	return WriteChannelSubscriptions(channelSub)
}

func WriteHotNewsSubscription(channelId int64, hotNewsSubscription bool) error {
	channelSub, err := ReadChannelSubscriptions(channelId)
	if logs.HandleError(err) {
		channelSub = ChannelSubscription{ChannelId: channelId}
	}

	if channelSub.HotNewsSubscriptions == hotNewsSubscription {
		return nil
	}

	channelSub.HotNewsSubscriptions = hotNewsSubscription

	return WriteChannelSubscriptions(channelSub)
}

func WriteWeatherSubscription(channelId int64, weatherSubscription bool) error {
	channelSub, err := ReadChannelSubscriptions(channelId)
	if logs.HandleError(err) {
		channelSub = ChannelSubscription{ChannelId: channelId}
	}

	if channelSub.WeatherSubscriptions == weatherSubscription {
		return nil
	}

	channelSub.WeatherSubscriptions = weatherSubscription

	return WriteChannelSubscriptions(channelSub)
}

func WriteCurrencySubscription(channelId int64, currencySubscription bool) error {
	channelSub, err := ReadChannelSubscriptions(channelId)
	if logs.HandleError(err) {
		channelSub = ChannelSubscription{ChannelId: channelId}
	}

	if channelSub.CurrencySubscriptions == currencySubscription {
		return nil
	}

	channelSub.CurrencySubscriptions = currencySubscription

	return WriteChannelSubscriptions(channelSub)
}

func WriteChannelSubscriptions(channelSub ChannelSubscription) error {
	db, err := initDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("INSERT OR REPLACE INTO %v (ChannelId, AllNewsSubscription, HotNewsSubscription, WeatherSubscription, CurrencySubscription) values (%v, %v, %v, %v, %v)", consts.UsersTableName, channelSub.ChannelId, channelSub.AllNewsSubscriptions, channelSub.HotNewsSubscriptions, channelSub.WeatherSubscriptions, channelSub.CurrencySubscriptions))
	if err != nil {
		return err
	}

	return nil
}

func GetConfigByName(setting string) (value string, err error) {
	db, err := initDatabase()
	if err != nil {
		return value, err
	}
	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("SELECT Value FROM %v WHERE Setting = '%v'", consts.ConfigTableName, setting))

	err = row.Scan(&value)
	if err != nil {
		err = SetConfigByName(setting, "")
		if err != nil {
			return value, err
		}
		return GetConfigByName(setting)
	}

	return value, nil
}

func SetConfigByName(setting string, value string) error {
	db, err := initDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("INSERT OR REPLACE INTO %v (Setting, Value) values ('%v', '%v')", consts.ConfigTableName, setting, value))
	if err != nil {
		return err
	}

	return nil
}

func IfNewsExistsInHistory(rssUrl, newsUrl string) (bool, error) {
	db, err := initDatabase()
	if err != nil {
		return false, err
	}
	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE NewsRssUrl = '%v' and NewsUrl = '%v'", consts.NewsHistoryTableName, rssUrl, newsUrl))
	count := 0
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}

	if count != 0 {
		return true, nil
	}

	return false, nil
}

func AddNewsToNewsHistory(rssUrl, newsUrl string) error {
	db, err := initDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("INSERT INTO %v (NewsRssUrl, NewsUrl) values ('%v', '%v')", consts.NewsHistoryTableName, rssUrl, newsUrl))
	if err != nil {
		return err
	}

	return CleanNewsHistoryTable(rssUrl)
}

func CleanNewsHistoryTable(rssUrl string) error {
	db, err := initDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE NewsRssUrl = '%v'", consts.NewsHistoryTableName, rssUrl))
	count := 0
	err = row.Scan(&count)
	if err != nil {
		return err
	}

	if count > 1 {
		_, err = db.Exec(fmt.Sprintf("delete top(%v) from %v", count-100, consts.NewsHistoryTableName))
		logs.HandleError(err)
	}
	return nil
}
