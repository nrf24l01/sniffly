package schemas

type DeviceListItem struct {
	UUID      string `json:"uuid"`
	MAC       string `json:"mac"`
	IP        string `json:"ip"`
	UserLabel string `json:"user_label"`
}

type UpdateDeviceLabelRequest struct {
	UserLabel string `json:"user_label" validate:"required,min=1"`
}

type DeviceBlockRuleItem struct {
	UUID        string `json:"uuid"`
	DeviceUUID  string `json:"device_uuid"`
	DeviceMAC   string `json:"device_mac"`
	DeviceIP    string `json:"device_ip"`
	DeviceLabel string `json:"device_label"`
	TargetType  string `json:"target_type"`
	TargetValue string `json:"target_value"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateDeviceBlockRuleRequest struct {
	TargetType  string `json:"target_type" validate:"required,oneof=ip cidr sni"`
	TargetValue string `json:"target_value" validate:"required,min=1,max=255"`
	Enabled     *bool  `json:"enabled" validate:"omitempty"`
}

type UpdateDeviceBlockRuleRequest struct {
	TargetType  *string `json:"target_type" validate:"omitempty,oneof=ip cidr sni"`
	TargetValue *string `json:"target_value" validate:"omitempty,min=1,max=255"`
	Enabled     *bool   `json:"enabled" validate:"omitempty"`
}
