package main

import (
	"github.com/ihor-sokoliuk/newsbot/bot"
	"github.com/ihor-sokoliuk/newsbot/configs"
	"github.com/ihor-sokoliuk/newsbot/database"
	"github.com/ihor-sokoliuk/newsbot/logs"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	projectName := os.Args[0][strings.LastIndex(os.Args[0], "/")+1:]
	configs.InitConfigVariables(projectName)

	// Init logger
	newsBotLogger := &logs.NewsBotLogger{Logger: *logs.NewLogger("")}

	// Init configs
	var config configs.Configs
	file, err := ioutil.ReadFile(configs.ConfigFileName)
	newsBotLogger.HandlePanic(err)
	err = yaml.Unmarshal(file, &config)
	newsBotLogger.HandlePanic(err)

	// Init database
	newsBotDatabase, err := database.NewDatabase()
	newsBotLogger.HandlePanic(err)

	// Run bot
	env := bot.Env{
		Db:      newsBotDatabase,
		Logger:  newsBotLogger,
		Configs: &config,
	}

	newsBotLogger.Info("Project: " + configs.ProjectName)
	newsBotLogger.Info("Database: " + configs.DatabaseFileName)
	newsBotLogger.Info("Config file: " + configs.ConfigFileName)

	bot.RunBot(&env)
}
