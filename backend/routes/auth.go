package routes

import (
	"github.com/labstack/echo/v4"
	echokitMW "github.com/nrf24l01/go-web-utils/echokit/middleware"
	"github.com/nrf24l01/sniffly/backend/handlers"
	"github.com/nrf24l01/sniffly/backend/schemas"
)

func RegisterAuthRoutes(e *echo.Echo, h *handlers.Handler) {
	group := e.Group("/auth")

	group.POST("/login", h.LoginHandler, echokitMW.BodyValidationMiddleware(func() interface{} {
		return &schemas.LoginRequest{}
	}))
}

