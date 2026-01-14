package schemas

type ChartDataRangeRequest struct {
	From     int64   `query:"from" validate:"required,gt=0"`
	To       int64   `query:"to" validate:"required,gtfield=From"`
	DeviceIDs []string `query:"device_id" validate:"omitempty,dive,uuid4"`
}