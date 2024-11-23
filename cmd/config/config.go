package config

type InitOptions struct {
	ProjectName string
	LogLevel    string
}

type JotlConfig struct {
	InitConfig InitOptions
}

func (c *JotlConfig) SetProjectName(name string) {
	c.InitConfig.ProjectName = name
}

func (c *JotlConfig) SetLogLevel(loglevel string) {
	c.InitConfig.LogLevel = loglevel
}
