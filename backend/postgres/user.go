package postgres

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/nrf24l01/go-web-utils/auth"
	utilsConfig "github.com/nrf24l01/go-web-utils/config"
	"github.com/nrf24l01/go-web-utils/pg_kit"
)

type User struct {
	pg_kit.BaseModel

	Username string `gorm:"unique;size:50;not null"`
	Password string `gorm:"size:100;not null"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) SetPassword(password string, config *utilsConfig.Argon2idConfig) error {
	hash, err := auth.HashPassword(password, config)
	if err != nil {
		return err
	}
	u.Password = hash
	return nil
}

func (u *User) CheckPassword(password string, config *utilsConfig.Argon2idConfig) (bool, error) {
	return auth.CheckPassword(password, u.Password)
}

func (u *User) GenerateJWTpair(config *utilsConfig.JWTConfig) (string, string, error) {
	access_claims := jwt.MapClaims{
		"user_id": u.ID.String(),
		"username": u.Username,
	}
	refresh_claims := jwt.MapClaims{
		"user_id": u.ID.String(),
	}

	return auth.GenerateTokenPair(access_claims, refresh_claims, config)
}

func (u *User) GenerateAccessToken(config *utilsConfig.JWTConfig) (string, error) {
	claims := jwt.MapClaims{
		"user_id": u.ID.String(),
		"username": u.Username,
	}
	return auth.GenerateAccessToken(claims, config)
}