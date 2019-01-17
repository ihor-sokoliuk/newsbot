package configs

var ProjectName = "testnewsbot"
var ConfigFileName = ProjectName + ".yml"

// Database constants
var DatabaseFileName = ProjectName + ".db"

const ChannelSubscriptionsTableName = "ChannelSubscriptions"
const ConfigTableName = "Configs"
const NewsHistoryTableName = "NewsHistory"

func InitConfigVariables(projectName string) {
	ProjectName = projectName
	ConfigFileName = projectName + ".yml"
	DatabaseFileName = projectName + ".db"
}
