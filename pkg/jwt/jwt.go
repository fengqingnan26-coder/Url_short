package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jekyulll/url_shortener/config"
)

type JWT struct {
	secret   []byte
	duration time.Duration
}

func NewJWT(cfg config.JWTConfig) *JWT {
	return &JWT{
		secret:   []byte(cfg.Secret),
		duration: cfg.Duration,
	}
}

type UserClaims struct {
	Email  string `json:"email"`
	UserID int    `json:"user_id"`
	jwt.RegisteredClaims
}

func (j *JWT) Generate(email string, userId int) (string, error) {
	claims := UserClaims{
		Email:  email,
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	// 两行重点
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWT) ParseToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok {
		return claims, nil
	}
	return nil, fmt.Errorf("failed to parseToken: %s", tokenString)
}
