package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nrf24l01/sniffly/backend/handlers"
)

func RegisterRoutes(e *echo.Echo, h *handlers.Handler) {
	RegisterAuthRoutes(e, h)
	RegisterChartsRoutes(e, h)
	RegisterTablesRoutes(e, h)
	RegisterDeviceRoutes(e, h)
	RegisterCapturerRoutes(e, h)
}
