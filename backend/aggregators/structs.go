package aggregators

type TimeRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type Device struct {
	MAC 	string  `json:"mac"`
	IP 		string  `json:"ip"`
	Label 	string  `json:"label"`
	Hostname string `json:"hostname"`
}

type Traffic struct {
	Bucket  	int64   `json:"bucket"`
	UpBytes		uint64  `json:"up_bytes"`
	DownBytes	uint64  `json:"down_bytes"`
	ReqCount	uint64  `json:"req_count"`
}

type TrafficChartData struct {
	Device  Device    `json:"device"`
	Traffic []Traffic `json:"stats"`
}