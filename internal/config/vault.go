package config

type Vault struct {
	Address string `validate:"required"`
	Token   string `validate:"required"`
}
