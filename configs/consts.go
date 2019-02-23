package configs

// ProjectName variable contains project name
var ProjectName = "testnewsbot"

// ConfigFileName variable contains project config file name
var ConfigFileName = ProjectName + ".yml"

// DatabaseFileName variable contains project database file name
var DatabaseFileName = ProjectName + ".db"

// InitConfigVariables is function to init program variables
func InitConfigVariables(projectName string) {
	ProjectName = projectName
	ConfigFileName = projectName + ".yml"
	DatabaseFileName = projectName + ".db"
}
