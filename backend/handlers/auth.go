package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nrf24l01/sniffly/backend/postgres"
	"github.com/nrf24l01/sniffly/backend/schemas"

	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
)

func (h *Handler) LoginHandler(c echo.Context) error {
	req := c.Get("validatedBody").(*schemas.LoginRequest)

	var user postgres.User
	
	if err := h.DB.Select("id", "username", "password").Where("username = ?", req.Username).First(&user).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, echokitSchemas.DefaultUnauthorizedResponse)
	}

	passwordMatch, err := user.CheckPassword(req.Password, h.Config.Argon2idConfig)
	if err != nil || !passwordMatch {
		return echo.NewHTTPError(http.StatusUnauthorized, echokitSchemas.DefaultUnauthorizedResponse)
	}

	token, refreshToken, err := user.GenerateJWTpair(h.Config.JWTConfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate JWT token")
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