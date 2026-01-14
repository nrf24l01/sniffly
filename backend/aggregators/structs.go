package aggregators

type TimeRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type Device struct {
	MAC       string `json:"mac"`
	IP        string `json:"ip"`
	Label     string `json:"label,omitempty"`
	UserLabel string `json:"user_label"`
	Hostname  string `json:"hostname"`
}

type Traffic struct {
	Bucket    int64  `json:"bucket"`
	UpBytes   uint64 `json:"up_bytes"`
	DownBytes uint64 `json:"down_bytes"`
	ReqCount  uint64 `json:"req_count"`
}

type TrafficChartData struct {
	Device Device    `json:"device"`
	Stats  []Traffic `json:"stats"`
}

type DomainStat struct {
	Bucket   int64             `json:"bucket"`
	Domains  map[string]uint64 `json:"domains"`
	ReqCount uint64            `json:"req_count"`
}

type DomainChartData struct {
	Device Device       `json:"device"`
	Stats  []DomainStat `json:"stats"`
}

type ProtoStat struct {
	Bucket   int64             `json:"bucket"`
	Protos   map[string]uint64 `json:"protos"`
	ReqCount uint64            `json:"req_count"`
}

type ProtoChartData struct {
	Device Device      `json:"device"`
	Stats  []ProtoStat `json:"stats"`
}

type CountryStat struct {
	Bucket    int64             `json:"bucket"`
	Countries map[string]uint64 `json:"countries"`
	Companies map[string]uint64 `json:"companies"`
	ReqCount  uint64            `json:"req_count"`
}

type CountryChartData struct {
	Device Device        `json:"device"`
	Stats  []CountryStat `json:"stats"`
}

// Aggregated (device-less) API responses

type TrafficChartResponse struct {
	Stats []Traffic `json:"stats"`
}

type DomainChartResponse struct {
	Stats []DomainStat `json:"stats"`
}

type ProtoChartResponse struct {
	Stats []ProtoStat `json:"stats"`
}

type CountryChartResponse struct {
	Stats []CountryStat `json:"stats"`
}

type TrafficTableResponse struct {
	Stats struct {
		UpBytes   uint64 `json:"up_bytes"`
		DownBytes uint64 `json:"down_bytes"`
	} `json:"stats"`
}

type DomainTableResponse struct {
	Stats map[string]uint64 `json:"stats"`
}

type CountryTableResponse struct {
	Stats map[string]uint64 `json:"stats"`
}

type ProtoTableResponse struct {
	Stats map[string]uint64 `json:"stats"`
}

type CompanyTableResponse struct {
	Stats map[string]uint64 `json:"stats"`
}
