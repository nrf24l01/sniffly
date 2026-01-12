package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
	"github.com/nrf24l01/sniffly/backend/core"
	"github.com/nrf24l01/sniffly/backend/schemas"
	"github.com/nrf24l01/sniffly/capture_receiver/postgres"
	"gorm.io/gorm"
)

func (h *Handler) GetCapturersHandler(c echo.Context) error {
	var capturers []postgres.Capturer
	if err := h.DB.Find(&capturers).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	var resp []schemas.Capturer
	for _, capturer := range capturers {
		resp = append(resp, schemas.Capturer{
			UUID:    capturer.ID.String(),
			Name:    capturer.Name,
			ApiKey:  capturer.ApiKey,
			Enabled: capturer.Enabled,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateCapturerHandler(c echo.Context) error {
	req := c.Get("validatedBody").(*schemas.CapturerCreateRequest)

	// Check if capture with required name exists
	if err := h.DB.Where("name = ?", req.Name).First(&postgres.Capturer{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusConflict, echokitSchemas.ErrorResponse{
				Message: "Capturer with this name already exists",
				Code:    http.StatusConflict,
			})
		} else {
			return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
		}
	}

	capturer := postgres.Capturer{
		Name:    req.Name,
		ApiKey:  core.GenerateApiKey(h.RandomGenerator),
		Enabled: req.Enabled,
	}

	if err := h.DB.Create(&capturer).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	resp := schemas.CapturerCreateResponse{
		UUID:   capturer.ID.String(),
		ApiKey: capturer.ApiKey,
	}

	return c.JSON(http.StatusCreated, resp)
}