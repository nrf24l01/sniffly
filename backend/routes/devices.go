package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nrf24l01/sniffly/backend/handlers"
	"github.com/nrf24l01/sniffly/backend/schemas"

	echokitMW "github.com/nrf24l01/go-web-utils/echokit/middleware"
)

func RegisterDeviceRoutes(e *echo.Echo, h *handlers.Handler) {
    group := e.Group("/devices")
    group.Use(echokitMW.JWTMiddleware(*h.Config.JWTConfig))

    group.GET("", h.GetDevicesHandler)
    group.PATCH("/:id", h.UpdateDeviceLabelHandler, echokitMW.PathUuidV4Middleware("id"), echokitMW.BodyValidationMiddleware(func() interface{} {
        return &schemas.UpdateDeviceLabelRequest{}
    }))
}
