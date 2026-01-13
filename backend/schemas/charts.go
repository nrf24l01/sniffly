package schemas

type ChartDataRangeRequest struct {
	From     int64   `query:"from" validate:"required,gt=0"`
	To       int64   `query:"to" validate:"required,gtfield=From"`
	DeviceID *string `query:"device_id" validate:"omitempty,uuid4"`
}