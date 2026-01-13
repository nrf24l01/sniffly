package aggregators

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	analyzerModels "github.com/nrf24l01/sniffly/analyzer/postgres"
	"gorm.io/gorm"
)

// TrafficTableRow matches swagger DeviceTrafficSummary
// DownBytes is not available in the current schema, so we expose 0 for compatibility.
type TrafficTableRow struct {
	Device Device `json:"device"`
	Stats  struct {
		UpBytes   uint64 `json:"up_bytes"`
		DownBytes uint64 `json:"down_bytes"`
	} `json:"stats"`
}

type DomainTableRow struct {
	Device Device            `json:"device"`
	Stats  map[string]uint64 `json:"stats"`
}

type CountryTableRow struct {
	Device Device            `json:"device"`
	Stats  map[string]uint64 `json:"stats"`
}

type ProtoTableRow struct {
	Device Device            `json:"device"`
	Stats  map[string]uint64 `json:"stats"`
}

type CompanyTableRow struct {
	Device Device            `json:"device"`
	Stats  map[string]uint64 `json:"stats"`
}

func GetTrafficTableData(db *gorm.DB, timerange TimeRange, deviceID *string) ([]TrafficTableRow, error) {
	type row struct {
		MAC      string
		IP       string
		Label    string
		Hostname string
		UpBytes  uint64
	}

	q := db.Model(&analyzerModels.DeviceTraffic5s{}).
		Select("device_info.mac as mac, device_info.ip as ip, device_info.label as label, device_info.hostname as hostname, SUM(devices_traffics_5s.up_bytes) as up_bytes").
		Joins("JOIN device_info ON device_info.id = devices_traffics_5s.device_id").
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0)).
		Group("device_info.mac, device_info.ip, device_info.label, device_info.hostname")

	if deviceID != nil {
		uid, err := uuid.Parse(*deviceID)
		if err != nil {
			return nil, err
		}
		q = q.Where("devices_traffics_5s.device_id = ?", uid)
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]TrafficTableRow, 0, len(rows))
	for _, r := range rows {
		item := TrafficTableRow{
			Device: Device{
				MAC:       r.MAC,
				IP:        r.IP,
				Label:     r.Label,
				UserLabel: r.Label,
				Hostname:  r.Hostname,
			},
		}
		item.Stats.UpBytes = r.UpBytes
		item.Stats.DownBytes = 0
		out = append(out, item)
	}

	return out, nil
}

func GetDomainTableData(db *gorm.DB, timerange TimeRange, deviceID *string) ([]DomainTableRow, error) {
	entries := []analyzerModels.DeviceDomain5s{}
	q := db.Model(&analyzerModels.DeviceDomain5s{}).
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0)).
		Preload("Device")

	if deviceID != nil {
		uid, err := uuid.Parse(*deviceID)
		if err != nil {
			return nil, err
		}
		q = q.Where("device_id = ?", uid)
	}

	if err := q.Find(&entries).Error; err != nil {
		return nil, err
	}

	acc := make(map[string]DomainTableRow)
	for _, e := range entries {
		var domains map[string]uint64
		if err := json.Unmarshal([]byte(e.Domain), &domains); err != nil {
			continue
		}

		key := e.Device.ID.String()
		row, ok := acc[key]
		if !ok {
			row = DomainTableRow{
				Device: Device{
					MAC:       e.Device.MAC,
					IP:        e.Device.IP,
					Label:     e.Device.Label,
					UserLabel: e.Device.Label,
					Hostname:  e.Device.Hostname,
				},
				Stats: make(map[string]uint64),
			}
		}

		for k, v := range domains {
			row.Stats[k] += v
		}

		acc[key] = row
	}

	out := make([]DomainTableRow, 0, len(acc))
	for _, r := range acc {
		out = append(out, r)
	}
	return out, nil
}

func GetCountryTableData(db *gorm.DB, timerange TimeRange, deviceID *string) ([]CountryTableRow, error) {
	entries := []analyzerModels.DeviceCountry5s{}
	q := db.Model(&analyzerModels.DeviceCountry5s{}).
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0)).
		Preload("Device")

	if deviceID != nil {
		uid, err := uuid.Parse(*deviceID)
		if err != nil {
			return nil, err
		}
		q = q.Where("device_id = ?", uid)
	}

	if err := q.Find(&entries).Error; err != nil {
		return nil, err
	}

	acc := make(map[string]CountryTableRow)
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

		key := e.Device.ID.String()
		row, ok := acc[key]
		if !ok {
			row = CountryTableRow{
				Device: Device{
					MAC:       e.Device.MAC,
					IP:        e.Device.IP,
					Label:     e.Device.Label,
					UserLabel: e.Device.Label,
					Hostname:  e.Device.Hostname,
				},
				Stats: make(map[string]uint64),
			}
		}

		for k, v := range countries {
			row.Stats[k] += v
		}
		for k, v := range companies {
			row.Stats[k] += v
		}

		acc[key] = row
	}

	out := make([]CountryTableRow, 0, len(acc))
	for _, r := range acc {
		out = append(out, r)
	}
	return out, nil
}

func GetCompanyTableData(db *gorm.DB, timerange TimeRange, deviceID *string) ([]CompanyTableRow, error) {
	entries := []analyzerModels.DeviceCountry5s{}
	q := db.Model(&analyzerModels.DeviceCountry5s{}).
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0)).
		Preload("Device")

	if deviceID != nil {
		uid, err := uuid.Parse(*deviceID)
		if err != nil {
			return nil, err
		}
		q = q.Where("device_id = ?", uid)
	}

	if err := q.Find(&entries).Error; err != nil {
		return nil, err
	}

	acc := make(map[string]CompanyTableRow)
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

		key := e.Device.ID.String()
		row, ok := acc[key]
		if !ok {
			row = CompanyTableRow{
				Device: Device{
					MAC:       e.Device.MAC,
					IP:        e.Device.IP,
					Label:     e.Device.Label,
					UserLabel: e.Device.Label,
					Hostname:  e.Device.Hostname,
				},
				Stats: make(map[string]uint64),
			}
		}

		for k, v := range companies {
			row.Stats[k] += v
		}

		acc[key] = row
	}

	out := make([]CompanyTableRow, 0, len(acc))
	for _, r := range acc {
		out = append(out, r)
	}
	return out, nil
}

func GetProtoTableData(db *gorm.DB, timerange TimeRange, deviceID *string) ([]ProtoTableRow, error) {
	entries := []analyzerModels.DeviceProto5s{}
	q := db.Model(&analyzerModels.DeviceProto5s{}).
		Where("bucket >= ? AND bucket <= ?", time.Unix(timerange.Start, 0), time.Unix(timerange.End, 0)).
		Preload("Device")

	if deviceID != nil {
		uid, err := uuid.Parse(*deviceID)
		if err != nil {
			return nil, err
		}
		q = q.Where("device_id = ?", uid)
	}

	if err := q.Find(&entries).Error; err != nil {
		return nil, err
	}

	acc := make(map[string]ProtoTableRow)
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

		key := e.Device.ID.String()
		row, ok := acc[key]
		if !ok {
			row = ProtoTableRow{
				Device: Device{
					MAC:       e.Device.MAC,
					IP:        e.Device.IP,
					Label:     e.Device.Label,
					UserLabel: e.Device.Label,
					Hostname:  e.Device.Hostname,
				},
				Stats: make(map[string]uint64),
			}
		}

		for k, v := range protos {
			row.Stats[k] += v
		}
		acc[key] = row
	}

	out := make([]ProtoTableRow, 0, len(acc))
	for _, r := range acc {
		out = append(out, r)
	}
	return out, nil
}
