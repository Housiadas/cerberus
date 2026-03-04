package config

type App struct {
	Name        string
	Namespace   string
	Environment string
	Version     Version
}

type Version struct {
	Build       string
	Description string
}
