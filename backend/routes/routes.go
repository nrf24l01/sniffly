package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nrf24l01/sniffly/backend/handlers"
	"github.com/nrf24l01/sniffly/backend/schemas"

	echokitMW "github.com/nrf24l01/go-web-utils/echokit/middleware"
)


func RegisterCapturerRoutes(e *echo.Echo, h *handlers.Handler) {
	group := e.Group("/captures")

	group.GET("", h.GetCapturersHandler)
	group.POST("", h.CreateCapturerHandler, echokitMW.BodyValidationMiddleware(func() interface{} {
		return &schemas.CapturerCreateRequest{}
	}))
}