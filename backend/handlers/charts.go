package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
	"github.com/nrf24l01/sniffly/backend/aggregators"
	"github.com/nrf24l01/sniffly/backend/schemas"
)

func (h *Handler) GetChartsTrafficHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)

	traffic, err := aggregators.GetTrafficChartData(h.DB, h.RDB, h.Config, aggregators.TimeRange{
		Start: req.From,
		End:   req.To,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, traffic)
}