package aggregators

import (
	"encoding/json"
	"time"

	redisutil "github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/backend/core"
)

func getCacheEntriesPerInterval(config *core.Config, rdb *redisutil.RedisClient, interval TimeRange) (map[time.Time][]TrafficChartData, []time.Time, error) {
	days_caches := make(map[time.Time][]TrafficChartData)
	var intervals []time.Time

	start_date := time.Unix(interval.Start, 0).Truncate(24 * time.Hour)
	end_date := time.Unix(interval.End, 0).Truncate(24 * time.Hour)

	for date := start_date; !date.After(end_date); date = date.Add(24 * time.Hour) {
		cache_key := config.BackendConfig.CacheDayAggPrefix + date.Format("2006_01_02")
		var day_cache []TrafficChartData
		res := rdb.Client.Get(rdb.Ctx, cache_key)
		if res.Err() == nil {
			val, err := res.Bytes()
			if err != nil {
				return nil, nil, err
			}
			err = json.Unmarshal(val, &day_cache)
			if err != nil {
				return nil, nil, err
			}

			days_caches[date] = day_cache
			intervals = append(intervals, date)
		}
	}

	return days_caches, intervals, nil
}