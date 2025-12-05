package aggregators

import (
	"time"

	redisutil "github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/backend/core"
	"gorm.io/gorm"
)

func GetTrafficChartData(db *gorm.DB, rdb *redisutil.RedisClient, config *core.Config, timerange TimeRange) ([]TrafficChartData, error) {
	traffic_per_day_cache, days_from_cache, err := getCacheEntriesPerInterval(config, rdb, timerange)
	if err != nil {
		return nil, err
	}
	
	non_cached_days := []time.Time{}
	for date := time.Unix(timerange.Start, 0).Truncate(24 * time.Hour); !date.After(time.Unix(timerange.End, 0).Truncate(24 * time.Hour)); date = date.Add(24 * time.Hour) {
		found := false
		for _, cached_date := range days_from_cache {
			if date.Equal(cached_date) {
				found = true
				break
			}
		}
		if !found {
			non_cached_days = append(non_cached_days, date)
		}
	}

	traffic_per_day_uncache, err := loadFromPostgres(db, non_cached_days)
	if err != nil {
		return nil, err
	}

	// Merge cached and uncached data
	final_traffic_data := make(map[string]TrafficChartData)

	// Process cached data
	for _, day_cache := range traffic_per_day_cache {
		for _, entry := range day_cache {
			device_key := entry.Device.MAC
			if _, exists := final_traffic_data[device_key]; !exists {
				final_traffic_data[device_key] = TrafficChartData{
					Device:  entry.Device,
					Traffic: []Traffic{},
				}
			}
			temp := final_traffic_data[device_key]
			temp.Traffic = append(temp.Traffic, entry.Traffic...)
			final_traffic_data[device_key] = temp
		}
	}

	// Process uncached data
	for _, entry := range traffic_per_day_uncache {
		device := Device{
			MAC:      entry.Device.MAC,
			IP:       entry.Device.IP,
			Label:    entry.Device.Label,
			Hostname: entry.Device.Hostname,
		}
		traffic := Traffic{
			Bucket:    entry.Bucket.Unix(),
			UpBytes:   entry.UpBytes,
			DownBytes: 0,
			ReqCount:  entry.ReqCount,
		}
		device_key := device.MAC
		if _, exists := final_traffic_data[device_key]; !exists {
			final_traffic_data[device_key] = TrafficChartData{
				Device:  device,
				Traffic: []Traffic{},
			}
		}
		temp := final_traffic_data[device_key]
		temp.Traffic = append(temp.Traffic, traffic)
		final_traffic_data[device_key] = temp
	}

	// Convert map to slice
	result := make([]TrafficChartData, 0, len(final_traffic_data))
	for _, data := range final_traffic_data {
		result = append(result, data)
	}

	return result, nil
}