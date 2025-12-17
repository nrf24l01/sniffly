package handlers

import (
	"log"
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

func (h *Handler) GetChartsDomainsHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)

	data, err := aggregators.GetDomainChartData(h.DB, h.RDB, h.Config, aggregators.TimeRange{
		Start: req.From,
		End:   req.To,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetChartsProtosHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)

	data, err := aggregators.GetProtoChartData(h.DB, h.RDB, h.Config, aggregators.TimeRange{
		Start: req.From,
		End:   req.To,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetChartsCountriesHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)

	data, err := aggregators.GetCountryChartData(h.DB, h.RDB, h.Config, aggregators.TimeRange{
		Start: req.From,
		End:   req.To,
	})
	if err != nil {
		log.Printf("GetCountryChartData error: %v", err)
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}