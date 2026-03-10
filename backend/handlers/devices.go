package handlers

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echokitSchemas "github.com/nrf24l01/go-web-utils/echokit/schemas"
	analyzerModels "github.com/nrf24l01/sniffly/analyzer/postgres"
	backendModels "github.com/nrf24l01/sniffly/backend/postgres"
	"github.com/nrf24l01/sniffly/backend/schemas"
	"gorm.io/gorm"
)

type deviceBlockRuleRow struct {
	ID          uuid.UUID
	DeviceID    uuid.UUID
	DeviceMAC   string
	DeviceIP    string
	DeviceLabel string
	TargetType  string
	TargetValue string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func normalizeBlockRule(target_type string, target_value string) (string, string, error) {
	target_type = strings.ToLower(strings.TrimSpace(target_type))
	target_value = strings.TrimSpace(target_value)
	if target_value == "" {
		return "", "", fmt.Errorf("target_value is required")
	}

	switch target_type {
	case backendModels.DeviceBlockRuleTargetIP:
		if strings.Contains(target_value, "/") || net.ParseIP(target_value) == nil {
			return "", "", fmt.Errorf("target_value must be a valid IP address")
		}
		return target_type, target_value, nil
	case backendModels.DeviceBlockRuleTargetCIDR:
		if _, _, err := net.ParseCIDR(target_value); err != nil {
			return "", "", fmt.Errorf("target_value must be a valid subnet in CIDR notation")
		}
		return target_type, target_value, nil
	case backendModels.DeviceBlockRuleTargetSNI:
		return target_type, strings.ToLower(target_value), nil
	default:
		return "", "", fmt.Errorf("target_type must be one of: ip, cidr, sni")
	}
}

func deviceBlockRuleToSchema(row deviceBlockRuleRow) schemas.DeviceBlockRuleItem {
	return schemas.DeviceBlockRuleItem{
		UUID:        row.ID.String(),
		DeviceUUID:  row.DeviceID.String(),
		DeviceMAC:   row.DeviceMAC,
		DeviceIP:    row.DeviceIP,
		DeviceLabel: row.DeviceLabel,
		TargetType:  row.TargetType,
		TargetValue: row.TargetValue,
		Enabled:     row.Enabled,
		CreatedAt:   row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   row.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (h *Handler) getBlockRuleRow(device_id string, rule_id string) (*deviceBlockRuleRow, error) {
	var row deviceBlockRuleRow
	err := h.DB.Raw(`
		SELECT
			dbr.id,
			dbr.device_id,
			di.mac AS device_mac,
			di.ip AS device_ip,
			di.label AS device_label,
			dbr.target_type,
			dbr.target_value,
			dbr.enabled,
			dbr.created_at,
			dbr.updated_at
		FROM device_block_rules dbr
		JOIN device_info di ON di.id = dbr.device_id
		WHERE dbr.device_id = ? AND dbr.id = ?
	`, device_id, rule_id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}

	return &row, nil
}

func (h *Handler) listDeviceBlockRules(device_id *string) ([]schemas.DeviceBlockRuleItem, error) {
	query := `
		SELECT
			dbr.id,
			dbr.device_id,
			di.mac AS device_mac,
			di.ip AS device_ip,
			di.label AS device_label,
			dbr.target_type,
			dbr.target_value,
			dbr.enabled,
			dbr.created_at,
			dbr.updated_at
		FROM device_block_rules dbr
		JOIN device_info di ON di.id = dbr.device_id
	`
	args := make([]interface{}, 0, 1)
	if device_id != nil {
		query += " WHERE dbr.device_id = ?"
		args = append(args, *device_id)
	}
	query += " ORDER BY di.label ASC, di.mac ASC, dbr.created_at DESC"

	rows := make([]deviceBlockRuleRow, 0)
	if err := h.DB.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	rules := make([]schemas.DeviceBlockRuleItem, 0, len(rows))
	for _, row := range rows {
		rules = append(rules, deviceBlockRuleToSchema(row))
	}

	return rules, nil
}

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

func (h *Handler) ListDeviceBlockRulesHandler(c echo.Context) error {
	rules, err := h.listDeviceBlockRules(nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, rules)
}

func (h *Handler) ListDeviceBlockRulesForDeviceHandler(c echo.Context) error {
	device_id := c.Param("id")

	var device analyzerModels.DeviceInfo
	if err := h.DB.Where("id = ?", device_id).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{Message: "Device not found", Code: http.StatusNotFound})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	rules, err := h.listDeviceBlockRules(&device_id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, rules)
}

func (h *Handler) CreateDeviceBlockRuleHandler(c echo.Context) error {
	device_id := c.Param("id")
	req := c.Get("validatedBody").(*schemas.CreateDeviceBlockRuleRequest)

	var device analyzerModels.DeviceInfo
	if err := h.DB.Where("id = ?", device_id).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{Message: "Device not found", Code: http.StatusNotFound})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	target_type, target_value, err := normalizeBlockRule(req.TargetType, req.TargetValue)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echokitSchemas.ErrorResponse{Message: err.Error(), Code: http.StatusBadRequest})
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	rule := backendModels.DeviceBlockRule{
		DeviceID:    device_id,
		TargetType:  target_type,
		TargetValue: target_value,
		Enabled:     enabled,
	}

	if err := h.DB.Create(&rule).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return c.JSON(http.StatusConflict, echokitSchemas.ErrorResponse{Message: "Such block rule already exists for the device", Code: http.StatusConflict})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	row, err := h.getBlockRuleRow(device_id, rule.ID.String())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusCreated, deviceBlockRuleToSchema(*row))
}

func (h *Handler) UpdateDeviceBlockRuleHandler(c echo.Context) error {
	device_id := c.Param("id")
	rule_id := c.Param("rule_id")
	req := c.Get("validatedBody").(*schemas.UpdateDeviceBlockRuleRequest)

	var device analyzerModels.DeviceInfo
	if err := h.DB.Where("id = ?", device_id).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{Message: "Device not found", Code: http.StatusNotFound})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	var rule backendModels.DeviceBlockRule
	if err := h.DB.Where("id = ? AND device_id = ?", rule_id, device_id).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{Message: "Block rule not found", Code: http.StatusNotFound})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	next_target_type := rule.TargetType
	next_target_value := rule.TargetValue
	if req.TargetType != nil {
		next_target_type = *req.TargetType
	}
	if req.TargetValue != nil {
		next_target_value = *req.TargetValue
	}

	normalized_type, normalized_value, err := normalizeBlockRule(next_target_type, next_target_value)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echokitSchemas.ErrorResponse{Message: err.Error(), Code: http.StatusBadRequest})
	}
	rule.TargetType = normalized_type
	rule.TargetValue = normalized_value
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}

	if err := h.DB.Save(&rule).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return c.JSON(http.StatusConflict, echokitSchemas.ErrorResponse{Message: "Such block rule already exists for the device", Code: http.StatusConflict})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	row, err := h.getBlockRuleRow(device_id, rule_id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	return c.JSON(http.StatusOK, deviceBlockRuleToSchema(*row))
}

func (h *Handler) DeleteDeviceBlockRuleHandler(c echo.Context) error {
	device_id := c.Param("id")
	rule_id := c.Param("rule_id")

	var device analyzerModels.DeviceInfo
	if err := h.DB.Where("id = ?", device_id).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{Message: "Device not found", Code: http.StatusNotFound})
		}
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}

	result := h.DB.Where("id = ? AND device_id = ?", rule_id, device_id).Delete(&backendModels.DeviceBlockRule{})
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, echokitSchemas.DefaultInternalErrorResponse)
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, echokitSchemas.ErrorResponse{Message: "Block rule not found", Code: http.StatusNotFound})
	}

	return c.NoContent(http.StatusNoContent)
}
