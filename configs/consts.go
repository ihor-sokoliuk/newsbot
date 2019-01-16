package configs

import "os"

var ProjectName = os.Args[0]
var ConfigFileName = ProjectName + ".yml"

// Database constants
var DatabaseFileName = ProjectName + ".db"

const UsersTableName = "BotUsers"
const ConfigTableName = "Configs"
const NewsHistoryTableName = "NewsHistory"
