package configs

var ProjectName = "testnewsbot"
var ConfigFileName = ProjectName + ".yml"
var DatabaseFileName = ProjectName + ".db"

func InitConfigVariables(projectName string) {
	ProjectName = projectName
	ConfigFileName = projectName + ".yml"
	DatabaseFileName = projectName + ".db"
}
