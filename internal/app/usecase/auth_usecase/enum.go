package auth_usecase

type TokenType int

const (
	AccessToken TokenType = iota
	RefreshToken
)

var tokenName = map[TokenType]string{
	AccessToken:  "accessToken",
	RefreshToken: "refreshToken",
}

func (tt TokenType) String() string {
	return tokenName[tt]
}
