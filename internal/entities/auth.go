package entities

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type TokenClaims struct {
	jwt.StandardClaims
	UserId int    `json:"user_id"`
	JTI    string `json:"token_id"`
}

type AccessTokenClaims struct {
	jwt.StandardClaims
	UserId int          `json:"user_id"`
	JTI    string       `json:"token_id"`
	Role   *models.Role `json:"role"`
}

type SignUpUser struct {
	Name     string  `json:"name" binding:"required,min=3"`
	Username string  `json:"username" binding:"required,min=3"`
	Password string  `json:"password" binding:"required,password,min=6"`
	Email    *string `json:"email" binding:"omitempty,email"`
	TgId     *int    `json:"tg_id"`
}

type SignInUser struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
