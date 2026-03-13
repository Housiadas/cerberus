package auth_usecase

// LoginReq defines the data needed to authenticate a user.
type LoginReq struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenReq defines the data needed to refresh an access token.
type RefreshTokenReq struct {
	Token string `json:"refreshToken" validate:"required"`
}

// LogoutReq defines the data needed to logout a user.
type LogoutReq struct {
	Token string `json:"refreshToken" validate:"required"`
}

// Token represents the user token when requested.
type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// ForgotPasswordReq defines the data needed to initiate a password reset.
type ForgotPasswordReq struct {
	Email string `json:"email" validate:"required"`
}

// ResetPasswordReq defines the data needed to complete a password reset.
type ResetPasswordReq struct {
	Token           string `json:"token"           validate:"required"`
	OldPassword     string `json:"oldPassword"     validate:"required"`
	Password        string `json:"password"        validate:"required"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required"`
}
