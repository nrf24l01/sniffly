package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nrf24l01/sniffly/backend/handlers"
	"github.com/nrf24l01/sniffly/backend/schemas"

	echokitMW "github.com/nrf24l01/go-web-utils/echokit/middleware"
)

func RegisterTablesRoutes(e *echo.Echo, h *handlers.Handler) {
	group := e.Group("/tables")
	group.Use(echokitMW.JWTMiddleware(*h.Config.JWTConfig))

	validator := func() interface{} { return &schemas.ChartDataRangeRequest{} }

	group.GET("/traffic", h.GetTablesTrafficHandler, echokitMW.QueryValidationMiddleware(validator))
	group.GET("/domains", h.GetTablesDomainsHandler, echokitMW.QueryValidationMiddleware(validator))
	group.GET("/countries", h.GetTablesCountriesHandler, echokitMW.QueryValidationMiddleware(validator))
	group.GET("/protos", h.GetTablesProtosHandler, echokitMW.QueryValidationMiddleware(validator))
	group.GET("/companies", h.GetTablesCompaniesHandler, echokitMW.QueryValidationMiddleware(validator))
}
