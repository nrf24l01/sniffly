package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nrf24l01/sniffly/backend/handlers"
	"github.com/nrf24l01/sniffly/backend/schemas"

	echokitMW "github.com/nrf24l01/go-web-utils/echokit/middleware"
)


func RegisterCapturerRoutes(e *echo.Echo, h *handlers.Handler) {
	group := e.Group("/captures")
	group.Use(echokitMW.JWTMiddleware(*h.Config.JWTConfig))

	group.GET("", h.GetCapturersHandler)
	group.POST("", h.CreateCapturerHandler, echokitMW.BodyValidationMiddleware(func() interface{} {
		return &schemas.CapturerCreateRequest{}
	}))
	group.GET("/:uuid", h.GetCapturerHandler, echokitMW.PathUuidV4Middleware("uuid"))
	group.PATCH("/:uuid", h.UpdateCapturerHandler, echokitMW.BodyValidationMiddleware(func() interface{} {
		return &schemas.CapturerUpdateRequest{}
	}), echokitMW.PathUuidV4Middleware("uuid"))
	group.DELETE("/:uuid", h.DeleteCapturerHandler, echokitMW.PathUuidV4Middleware("uuid"))
	group.POST("/:uuid/regenerate", h.RegenerateCapturerApiKeyHandler, echokitMW.PathUuidV4Middleware("uuid"))
}