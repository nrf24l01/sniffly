package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nrf24l01/sniffly/backend/postgres"
	"github.com/nrf24l01/sniffly/backend/schemas"

	"github.com/nrf24l01/go-web-utils/auth"
	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
)

func (h *Handler) LoginHandler(c echo.Context) error {
	req := c.Get("validatedBody").(*schemas.LoginRequest)

	var user postgres.User
	
	if err := h.DB.Select("id", "username", "password").Where("username = ?", req.Username).First(&user).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, echokitSchemas.DefaultUnauthorizedResponse)
	}

	passwordMatch, err := user.CheckPassword(req.Password, h.Config.Argon2idConfig)
	if err != nil || !passwordMatch {
		return c.JSON(http.StatusUnauthorized, echokitSchemas.DefaultUnauthorizedResponse)
	}

	token, refreshToken, err := user.GenerateJWTpair(h.Config.JWTConfig)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = refreshToken
	cookie.HttpOnly = true
	cookie.Path = "/"
	c.SetCookie(cookie)

	resp := &schemas.JwtResponse{
		Token: token,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) TokenRefreshHandler(c echo.Context) error {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, schemas.DefaultNoTokenResponse)
	}

	tokenClaims, err := auth.ValidateToken(refreshToken.Value, []byte(h.Config.JWTConfig.RefreshJWTSecret))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, schemas.DefaultInvalidTokenResponse)
	}

	userID, ok := tokenClaims["user_id"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, schemas.DefaultInvalidTokenResponse)
	}

	var user postgres.User
	if err := h.DB.Select("id", "username").Where("id = ?", userID).First(&user).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, echokitSchemas.DefaultUnauthorizedResponse)
	}

	newAccessToken, err := user.GenerateAccessToken(h.Config.JWTConfig)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	resp := &schemas.JwtResponse{
		Token: newAccessToken,
	}

	return c.JSON(http.StatusOK, resp)
}