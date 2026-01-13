package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
	"github.com/nrf24l01/sniffly/backend/aggregators"
	"github.com/nrf24l01/sniffly/backend/schemas"
)

func (h *Handler) GetTablesTrafficHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)

	data, err := aggregators.GetTrafficTableData(h.DB, aggregators.TimeRange{Start: req.From, End: req.To}, req.DeviceID)
	if err != nil {
		log.Printf("GetTablesTrafficHandler error: %v", err)
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetTablesDomainsHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)

	data, err := aggregators.GetDomainTableData(h.DB, aggregators.TimeRange{Start: req.From, End: req.To}, req.DeviceID)
	if err != nil {
		log.Printf("GetTablesDomainsHandler error: %v", err)
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetTablesCountriesHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)

	data, err := aggregators.GetCountryTableData(h.DB, aggregators.TimeRange{Start: req.From, End: req.To}, req.DeviceID)
	if err != nil {
		log.Printf("GetTablesCountriesHandler error: %v", err)
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetTablesProtosHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)

	data, err := aggregators.GetProtoTableData(h.DB, aggregators.TimeRange{Start: req.From, End: req.To}, req.DeviceID)
	if err != nil {
		log.Printf("GetTablesProtosHandler error: %v", err)
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetTablesCompaniesHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)

	data, err := aggregators.GetCompanyTableData(h.DB, aggregators.TimeRange{Start: req.From, End: req.To}, req.DeviceID)
	if err != nil {
		log.Printf("GetTablesCompaniesHandler error: %v", err)
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}
