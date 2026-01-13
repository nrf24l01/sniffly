package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
	analyzerModels "github.com/nrf24l01/sniffly/analyzer/postgres"
	"github.com/nrf24l01/sniffly/backend/schemas"
	"gorm.io/gorm"
)

func (h *Handler) GetDevicesHandler(c echo.Context) error {
    var devices []analyzerModels.DeviceInfo
    if err := h.DB.Find(&devices).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
    }

    resp := make([]schemas.DeviceListItem, 0, len(devices))
    for _, d := range devices {
        resp = append(resp, schemas.DeviceListItem{
            UUID:      d.ID.String(),
            MAC:       d.MAC,
            IP:        d.IP,
            UserLabel: d.Label,
        })
    }

    return c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdateDeviceLabelHandler(c echo.Context) error {
    id := c.Param("id")
    req := c.Get("validatedBody").(*schemas.UpdateDeviceLabelRequest)

    var device analyzerModels.DeviceInfo
    if err := h.DB.Where("id = ?", id).First(&device).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{Message: "Device not found", Code: http.StatusNotFound})
        }
        return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
    }

    device.Label = req.UserLabel
    if err := h.DB.Save(&device).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
    }

    resp := schemas.DeviceListItem{
        UUID:      device.ID.String(),
        MAC:       device.MAC,
        IP:        device.IP,
        UserLabel: device.Label,
    }

    return c.JSON(http.StatusOK, resp)
}
