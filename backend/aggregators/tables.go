package aggregators

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	analyzerModels "github.com/nrf24l01/sniffly/analyzer/postgres"
	"gorm.io/gorm"
)


func GetTrafficTableData(db *gorm.DB, timerange TimeRange, deviceIDs []uuid.UUID) (TrafficTableResponse, error) {
	type row struct {
		UpBytes uint64
	}

	q := db.Model(&analyzerModels.DeviceTraffic5s{}).
		Select("COALESCE(SUM(up_bytes), 0) as up_bytes").
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0))

	if len(deviceIDs) > 0 {
		q = q.Where("device_id IN ?", deviceIDs)
	}

	var r row
	if err := q.Scan(&r).Error; err != nil {
		return TrafficTableResponse{}, err
	}

	out := TrafficTableResponse{}
	out.Stats.UpBytes = r.UpBytes
	out.Stats.DownBytes = 0
	return out, nil
}

func GetDomainTableData(db *gorm.DB, timerange TimeRange, deviceIDs []uuid.UUID) (DomainTableResponse, error) {
	entries := []analyzerModels.DeviceDomain5s{}
	q := db.Model(&analyzerModels.DeviceDomain5s{}).
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0))

	if len(deviceIDs) > 0 {
		q = q.Where("device_id IN ?", deviceIDs)
	}

	if err := q.Find(&entries).Error; err != nil {
		return DomainTableResponse{}, err
	}

	stats := make(map[string]uint64)
	for _, e := range entries {
		var domains map[string]uint64
		if err := json.Unmarshal([]byte(e.Domain), &domains); err != nil {
			continue
		}

		for k, v := range domains {
			stats[k] += v
		}
	}

	return DomainTableResponse{Stats: stats}, nil
}

func GetCountryTableData(db *gorm.DB, timerange TimeRange, deviceIDs []uuid.UUID) (CountryTableResponse, error) {
	entries := []analyzerModels.DeviceCountry5s{}
	q := db.Model(&analyzerModels.DeviceCountry5s{}).
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0))

	if len(deviceIDs) > 0 {
		q = q.Where("device_id IN ?", deviceIDs)
	}

	if err := q.Find(&entries).Error; err != nil {
		return CountryTableResponse{}, err
	}

	stats := make(map[string]uint64)
	for _, e := range entries {
		countries := make(map[string]uint64)
		companies := make(map[string]uint64)

		if err := json.Unmarshal([]byte(e.Countries), &countries); err != nil {
			// fallback if stored as array
			var arr []string
			if err2 := json.Unmarshal([]byte(e.Countries), &arr); err2 == nil {
				for _, c := range arr {
					countries[c]++
				}
			}
		}

		if err := json.Unmarshal([]byte(e.Companies), &companies); err != nil {
			var arr []string
			if err2 := json.Unmarshal([]byte(e.Companies), &arr); err2 == nil {
				for _, c := range arr {
					companies[c]++
				}
			}
		}

		for k, v := range countries {
			stats[k] += v
		}
		for k, v := range companies {
			stats[k] += v
		}
	}

	return CountryTableResponse{Stats: stats}, nil
}

func GetCompanyTableData(db *gorm.DB, timerange TimeRange, deviceIDs []uuid.UUID) (CompanyTableResponse, error) {
	entries := []analyzerModels.DeviceCountry5s{}
	q := db.Model(&analyzerModels.DeviceCountry5s{}).
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0))

	if len(deviceIDs) > 0 {
		q = q.Where("device_id IN ?", deviceIDs)
	}

	if err := q.Find(&entries).Error; err != nil {
		return CompanyTableResponse{}, err
	}

	stats := make(map[string]uint64)
	for _, e := range entries {
		companies := make(map[string]uint64)
		if err := json.Unmarshal([]byte(e.Companies), &companies); err != nil {
			var arr []string
			if err2 := json.Unmarshal([]byte(e.Companies), &arr); err2 == nil {
				for _, c := range arr {
					companies[c]++
				}
			}
		}

		for k, v := range companies {
			stats[k] += v
		}
	}

	return CompanyTableResponse{Stats: stats}, nil
}

func GetProtoTableData(db *gorm.DB, timerange TimeRange, deviceIDs []uuid.UUID) (ProtoTableResponse, error) {
	entries := []analyzerModels.DeviceProto5s{}
	q := db.Model(&analyzerModels.DeviceProto5s{}).
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0))

	if len(deviceIDs) > 0 {
		q = q.Where("device_id IN ?", deviceIDs)
	}

	if err := q.Find(&entries).Error; err != nil {
		return ProtoTableResponse{}, err
	}

	stats := make(map[string]uint64)
	for _, e := range entries {
		protos := make(map[string]uint64)
		if err := json.Unmarshal([]byte(e.Proto), &protos); err != nil {
			var arr []string
			if err2 := json.Unmarshal([]byte(e.Proto), &arr); err2 == nil {
				for _, p := range arr {
					protos[p]++
				}
			}
		}

		for k, v := range protos {
			stats[k] += v
		}
	}

	return ProtoTableResponse{Stats: stats}, nil
}
