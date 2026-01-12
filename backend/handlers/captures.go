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
	err := h.DB.Where("name = ?", req.Name).First(&postgres.Capturer{}).Error
	if err == nil {
		return c.JSON(http.StatusConflict, echokitSchemas.ErrorResponse{
			Message: "Capturer with this name already exists",
			Code:    http.StatusConflict,
		})
	} else if err != gorm.ErrRecordNotFound {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
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

func (h *Handler) GetCapturerHandler(c echo.Context) error {
	uuid := c.Param("uuid")

	var capturer postgres.Capturer

	// Check if capturer exists
	if err := h.DB.Where("id = ?", uuid).First(&capturer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{
				Message: "Capturer not found",
				Code:    http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	resp := schemas.Capturer{
		UUID:    capturer.ID.String(),
		Name:    capturer.Name,
		ApiKey:  capturer.ApiKey,
		Enabled: capturer.Enabled,
	}

	return c.JSON(http.StatusOK, resp)
}

	func (h *Handler) UpdateCapturerHandler(c echo.Context) error {
		uuid := c.Param("uuid")

	req := c.Get("validatedBody").(*schemas.CapturerUpdateRequest)
	var capturer postgres.Capturer
	if err := h.DB.Where("id = ?", uuid).First(&capturer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{
				Message: "Capturer not found",
				Code:    http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	// Update fields if provided
	if req.Name != nil {
		// Ensure unique name (exclude current record)
		var conflict postgres.Capturer
		if err := h.DB.Where("name = ? AND id <> ?", *req.Name, capturer.ID).First(&conflict).Error; err == nil {
			return c.JSON(http.StatusConflict, echokitSchemas.ErrorResponse{
				Message: "Capturer with this name already exists",
				Code:    http.StatusConflict,
			})
		} else if err != gorm.ErrRecordNotFound {
			return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
		}

		capturer.Name = *req.Name
	}

	if req.Enabled != nil {
		capturer.Enabled = *req.Enabled
	}

	if err := h.DB.Save(&capturer).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	resp := schemas.Capturer{
		UUID:    capturer.ID.String(),
		Name:    capturer.Name,
		ApiKey:  capturer.ApiKey,
		Enabled: capturer.Enabled,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) DeleteCapturerHandler(c echo.Context) error {
	uuid := c.Param("uuid")

	// Check if capturer exists
	if err := h.DB.Where("id = ?", uuid).First(&postgres.Capturer{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{
				Message: "Capturer not found",
				Code:    http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	if err := h.DB.Where("id = ?", uuid).Delete(&postgres.Capturer{}).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) RegenerateCapturerApiKeyHandler(c echo.Context) error {
	uuid := c.Param("uuid")

	var capturer postgres.Capturer
	if err := h.DB.Where("id = ?", uuid).First(&capturer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{
				Message: "Capturer not found",
				Code:    http.StatusNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	capturer.ApiKey = core.GenerateApiKey(h.RandomGenerator)

	if err := h.DB.Save(&capturer).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	resp := schemas.Capturer{
		UUID:    capturer.ID.String(),
		Name:    capturer.Name,
		ApiKey:  capturer.ApiKey,
		Enabled: capturer.Enabled,
	}

	return c.JSON(http.StatusOK, resp)
}
