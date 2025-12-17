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
		res := rdb.Client.Get(rdb.Ctx, cache_key)
		if res.Err() == nil {
			val, err := res.Bytes()
			if err != nil {
				return nil, nil, err
			}

			// Unmarshal into raw to allow tolerant conversion of older formats
			var raw struct {
				Version int                `json:"version"`
				Data    []json.RawMessage `json:"data_per_time"`
			}
			if err := json.Unmarshal(val, &raw); err != nil {
				return nil, nil, err
			}

			// Check version
			expected_version, ok := cacheVersions[date]
			if !ok || raw.Version != expected_version {
				continue
			}

			var dayData []T
			for _, itemRaw := range raw.Data {
				// Try direct unmarshal into T
				var item T
				if err := json.Unmarshal(itemRaw, &item); err == nil {
					dayData = append(dayData, item)
					continue
				}

				// Fallback: try to normalize possible old formats where
				// countries/companies could be arrays instead of maps.
				var m map[string]json.RawMessage
				if err := json.Unmarshal(itemRaw, &m); err != nil {
					// cannot parse this item, skip
					continue
				}

				// normalize stats if present
				if statsRaw, ok := m["stats"]; ok {
					var statsArr []map[string]json.RawMessage
					if err := json.Unmarshal(statsRaw, &statsArr); err == nil {
						for si, stat := range statsArr {
							// countries
							if cRaw, ok := stat["countries"]; ok {
								// if it's an array, convert to map[string]uint64
								var arr []string
								if err := json.Unmarshal(cRaw, &arr); err == nil {
									cmap := make(map[string]uint64, len(arr))
									for _, k := range arr {
										cmap[k] = cmap[k] + 1
									}
									// replace with object
									b, _ := json.Marshal(cmap)
									statsArr[si]["countries"] = b
								}
							}
							// companies
							if cRaw, ok := stat["companies"]; ok {
								var arr []string
								if err := json.Unmarshal(cRaw, &arr); err == nil {
									cmap := make(map[string]uint64, len(arr))
									for _, k := range arr {
										cmap[k] = cmap[k] + 1
									}
									b, _ := json.Marshal(cmap)
									statsArr[si]["companies"] = b
								}
							}
						}
						// put back normalized stats
						if b, err := json.Marshal(statsArr); err == nil {
							m["stats"] = b
						}
					}
				}

				// marshal modified map back to bytes and unmarshal into T
				if finalB, err := json.Marshal(m); err == nil {
					var finalItem T
					if err := json.Unmarshal(finalB, &finalItem); err == nil {
						dayData = append(dayData, finalItem)
					}
				}
			}

			if len(dayData) > 0 {
				days_caches[date] = dayData
				intervals = append(intervals, date)
			}
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






