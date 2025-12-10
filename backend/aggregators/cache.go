package aggregators

import (
	"encoding/json"
	"time"

	redisutil "github.com/nrf24l01/go-web-utils/redis"
	"github.com/nrf24l01/sniffly/backend/core"
)

type CacheValue[T any] struct {
	Version  int `json:"version"`
	Data     []T `json:"data_per_time"`
}

func getCacheEntriesPerInterval[T any](config *core.Config, rdb *redisutil.RedisClient, interval TimeRange, cacheVersions map[time.Time]int, prefix string) (map[time.Time][]T, []time.Time, error) {
	days_caches := make(map[time.Time][]T)
	var intervals []time.Time

	start_date := time.Unix(interval.Start, 0).UTC().Truncate(24 * time.Hour)
	end_date := time.Unix(interval.End, 0).UTC().Truncate(24 * time.Hour)

	for date := start_date; !date.After(end_date); date = date.Add(24 * time.Hour) {
		cache_key := config.BackendConfig.CacheDayAggPrefix + prefix + date.Format("2006_01_02")
		var day_cache CacheValue[T]
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
			// Check version
			expected_version, ok := cacheVersions[date]
			if !ok || day_cache.Version != expected_version {
				continue
			}

			days_caches[date] = day_cache.Data
			intervals = append(intervals, date)
		}
	}

	return days_caches, intervals, nil
}

func setCacheEntryForInterval[T any](config *core.Config, rdb *redisutil.RedisClient, interval TimeRange, data []T, version int, prefix string) error {
	cache_key := config.BackendConfig.CacheDayAggPrefix + prefix + time.Unix(interval.Start, 0).UTC().Format("2006_01_02")
	day_cache := CacheValue[T]{
		Version: version,
		Data:    data,
	}
	val, err := json.Marshal(day_cache)
	if err != nil {
		return err
	}
	err = rdb.Client.Set(rdb.Ctx, cache_key, val, 0).Err()
	if err != nil {
		return err
	}
	return nil
}






