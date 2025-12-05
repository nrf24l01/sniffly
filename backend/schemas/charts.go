package schemas

type ChartDataRangeRequest struct {
	From     int64  `json:"from" validate:"required,gt=0"`
	To       int64  `json:"to" validate:"required,gtfield=From"`
}