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
