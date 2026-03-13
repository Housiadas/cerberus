package config

type App struct {
	Name        string `validate:"required"`
	Namespace   string
	Environment string  `validate:"required"`
	Version     Version `validate:"required"`
	FrontendURL string  `validate:"required"`
}

type Version struct {
	Build       string
	Description string
}
