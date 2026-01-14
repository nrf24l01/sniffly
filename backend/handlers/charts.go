package handlers

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
	"github.com/nrf24l01/sniffly/backend/aggregators"
	"github.com/nrf24l01/sniffly/backend/schemas"
)

func parseDeviceIDs(ids []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(ids))
	for _, s := range ids {
		uid, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		out = append(out, uid)
	}
	return out, nil
}

func (h *Handler) GetChartsTrafficHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)
	deviceIDs, err := parseDeviceIDs(req.DeviceIDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echokitSchemas.DefaultBadRequestResponse)
	}

	traffic, err := aggregators.GetTrafficChartData(h.DB, h.RDB, h.Config, aggregators.TimeRange{
		Start: req.From,
		End:   req.To,
	}, deviceIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, traffic)
}

func (h *Handler) GetChartsDomainsHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)
	deviceIDs, err := parseDeviceIDs(req.DeviceIDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echokitSchemas.DefaultBadRequestResponse)
	}

	data, err := aggregators.GetDomainChartData(h.DB, h.RDB, h.Config, aggregators.TimeRange{
		Start: req.From,
		End:   req.To,
	}, deviceIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetChartsProtosHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)
	deviceIDs, err := parseDeviceIDs(req.DeviceIDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echokitSchemas.DefaultBadRequestResponse)
	}

	data, err := aggregators.GetProtoChartData(h.DB, h.RDB, h.Config, aggregators.TimeRange{
		Start: req.From,
		End:   req.To,
	}, deviceIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}

func (h *Handler) GetChartsCountriesHandler(c echo.Context) error {
	req := c.Get("validatedQuery").(*schemas.ChartDataRangeRequest)
	deviceIDs, err := parseDeviceIDs(req.DeviceIDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echokitSchemas.DefaultBadRequestResponse)
	}

	data, err := aggregators.GetCountryChartData(h.DB, h.RDB, h.Config, aggregators.TimeRange{
		Start: req.From,
		End:   req.To,
	}, deviceIDs)
	if err != nil {
		log.Printf("GetCountryChartData error: %v", err)
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, data)
}