package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nrf24l01/sniffly/backend/handlers"
	"github.com/nrf24l01/sniffly/backend/schemas"

	echokitMW "github.com/nrf24l01/go-web-utils/echokit/middleware"
)


func RegisterChartsRoutes(e *echo.Echo, h *handlers.Handler) {
	group := e.Group("/charts")

	group.POST("/traffic", h.GetChartsTrafficHandler, echokitMW.QueryValidationMiddleware(func() interface{} {
		return &schemas.ChartDataRangeRequest{}
	}))
}

