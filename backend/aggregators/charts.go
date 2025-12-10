package aggregators

import (
	"encoding/json"
	"log"
	"time"

	redisutil "github.com/nrf24l01/go-web-utils/redis"
	analyzerModels "github.com/nrf24l01/sniffly/analyzer/postgres"
	"github.com/nrf24l01/sniffly/backend/core"
	"gorm.io/gorm"
)

func GetGenericChartData[TChartData any, TPostgresModel any](
	db *gorm.DB,
	rdb *redisutil.RedisClient,
	config *core.Config,
	timerange TimeRange,
	cachePrefix string,
	loadPostgres func(*gorm.DB, []time.Time) ([]TPostgresModel, error),
	convert func([]TPostgresModel) []TChartData,
	getTimestamp func(TChartData) int64,
	getDeviceMAC func(TChartData) string,
	mergeStats func(TChartData, TChartData) TChartData,
) ([]TChartData, error) {
	cacheVersions, err := loadCacheVersionsFromPostgres(db, timerange)
	if err != nil {
		return nil, err
	}

	data_per_day_cache, days_from_cache, err := getCacheEntriesPerInterval[TChartData](config, rdb, timerange, cacheVersions, cachePrefix)
	if err != nil {
		return nil, err
	}

	non_cached_days := getNonCachedDays(timerange, days_from_cache)

	data_per_day_uncache, err := loadPostgres(db, non_cached_days)
	if err != nil {
		return nil, err
	}

	uncached_converted := convert(data_per_day_uncache)

	if len(uncached_converted) > 0 {
		err = cacheUncachedData(config, rdb, cacheVersions, uncached_converted, getTimestamp, cachePrefix)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("Cache stats: loaded from cache %d days, loaded from postgres %d days", len(days_from_cache), len(non_cached_days))

	return mergeData(data_per_day_cache, uncached_converted, getDeviceMAC, mergeStats), nil
}

func getNonCachedDays(timerange TimeRange, days_from_cache []time.Time) []time.Time {
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
	return non_cached_days
}

func mergeData[T any](
	cached map[time.Time][]T,
	uncached []T,
	getDeviceKey func(T) string,
	mergeStats func(T, T) T,
) []T {
	final_data := make(map[string]T)

	// Process cached data
	for _, day_cache := range cached {
		for _, entry := range day_cache {
			key := getDeviceKey(entry)
			if val, exists := final_data[key]; !exists {
				final_data[key] = entry
			} else {
				final_data[key] = mergeStats(val, entry)
			}
		}
	}

	// Process uncached data
	for _, entry := range uncached {
		key := getDeviceKey(entry)
		if val, exists := final_data[key]; !exists {
			final_data[key] = entry
		} else {
			final_data[key] = mergeStats(val, entry)
		}
	}

	result := make([]T, 0, len(final_data))
	for _, data := range final_data {
		result = append(result, data)
	}
	return result
}

func cacheUncachedData[T any](
	config *core.Config,
	rdb *redisutil.RedisClient,
	cacheVersions map[time.Time]int,
	uncached []T,
	getTimestamp func(T) int64,
	prefix string,
) error {
	// Group by day
	byDay := make(map[time.Time][]T)
	for _, item := range uncached {
		ts := getTimestamp(item)
		date := time.Unix(ts, 0).UTC().Truncate(24 * time.Hour)
		byDay[date] = append(byDay[date], item)
	}

	for date, items := range byDay {
		start_of_day := date.Unix()
		end_of_day := date.Add(24 * time.Hour).Unix() - 1
		
		if version, ok := cacheVersions[date]; ok {
			err := setCacheEntryForInterval(config, rdb, TimeRange{Start: start_of_day, End: end_of_day}, items, version, prefix)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetTrafficChartData(db *gorm.DB, rdb *redisutil.RedisClient, config *core.Config, timerange TimeRange) ([]TrafficChartData, error) {
	return GetGenericChartData(
		db, rdb, config, timerange, "",
		loadFromPostgres,
		func(models []analyzerModels.DeviceTraffic5s) []TrafficChartData {
			var result []TrafficChartData
			for _, entry := range models {
				result = append(result, TrafficChartData{
					Device: Device{
						MAC:      entry.Device.MAC,
						IP:       entry.Device.IP,
						Label:    entry.Device.Label,
						Hostname: entry.Device.Hostname,
					},
					Stats: []Traffic{{
						Bucket:    entry.Bucket.Unix(),
						UpBytes:   entry.UpBytes,
						DownBytes: 0,
						ReqCount:  entry.ReqCount,
					}},
				})
			}
			return result
		},
		func(t TrafficChartData) int64 { return t.Stats[0].Bucket },
		func(t TrafficChartData) string { return t.Device.MAC },
		func(a, b TrafficChartData) TrafficChartData {
			a.Stats = append(a.Stats, b.Stats...)
			return a
		},
	)
}

func GetDomainChartData(db *gorm.DB, rdb *redisutil.RedisClient, config *core.Config, timerange TimeRange) ([]DomainChartData, error) {
	return GetGenericChartData(
		db, rdb, config, timerange, "domain_",
		loadDomainsFromPostgres,
		func(models []analyzerModels.DeviceDomain5s) []DomainChartData {
			var result []DomainChartData
			for _, entry := range models {
				var domains map[string]uint64
				if err := json.Unmarshal([]byte(entry.Domain), &domains); err != nil {
					continue
				}
				result = append(result, DomainChartData{
					Device: Device{
						MAC:      entry.Device.MAC,
						IP:       entry.Device.IP,
						Label:    entry.Device.Label,
						Hostname: entry.Device.Hostname,
					},
					Stats: []DomainStat{{
						Bucket:   entry.Bucket.Unix(),
						Domains:  domains,
						ReqCount: entry.Requests,
					}},
				})
			}
			return result
		},
		func(t DomainChartData) int64 { return t.Stats[0].Bucket },
		func(t DomainChartData) string { return t.Device.MAC },
		func(a, b DomainChartData) DomainChartData {
			a.Stats = append(a.Stats, b.Stats...)
			return a
		},
	)
}

func GetProtoChartData(db *gorm.DB, rdb *redisutil.RedisClient, config *core.Config, timerange TimeRange) ([]ProtoChartData, error) {
	return GetGenericChartData(
		db, rdb, config, timerange, "proto_",
		loadProtosFromPostgres,
		func(models []analyzerModels.DeviceProto5s) []ProtoChartData {
			var result []ProtoChartData
			for _, entry := range models {
				var protos map[string]uint64
				if err := json.Unmarshal([]byte(entry.Proto), &protos); err != nil {
					continue
				}
				result = append(result, ProtoChartData{
					Device: Device{
						MAC:      entry.Device.MAC,
						IP:       entry.Device.IP,
						Label:    entry.Device.Label,
						Hostname: entry.Device.Hostname,
					},
					Stats: []ProtoStat{{
						Bucket:   entry.Bucket.Unix(),
						Protos:   protos,
						ReqCount: entry.Requests,
					}},
				})
			}
			return result
		},
		func(t ProtoChartData) int64 { return t.Stats[0].Bucket },
		func(t ProtoChartData) string { return t.Device.MAC },
		func(a, b ProtoChartData) ProtoChartData {
			a.Stats = append(a.Stats, b.Stats...)
			return a
		},
	)
}

func GetCountryChartData(db *gorm.DB, rdb *redisutil.RedisClient, config *core.Config, timerange TimeRange) ([]CountryChartData, error) {
	return GetGenericChartData(
		db, rdb, config, timerange, "country_",
		loadCountriesFromPostgres,
		func(models []analyzerModels.DeviceCountry5s) []CountryChartData {
			var result []CountryChartData
			for _, entry := range models {
				result = append(result, CountryChartData{
					Device: Device{
						MAC:      entry.Device.MAC,
						IP:       entry.Device.IP,
						Label:    entry.Device.Label,
						Hostname: entry.Device.Hostname,
					},
					Stats: []CountryStat{{
						Bucket:    entry.Bucket.Unix(),
						Countries: entry.Countries,
						Companies: entry.Companies,
						ReqCount:  entry.Requests,
					}},
				})
			}
			return result
		},
		func(t CountryChartData) int64 { return t.Stats[0].Bucket },
		func(t CountryChartData) string { return t.Device.MAC },
		func(a, b CountryChartData) CountryChartData {
			a.Stats = append(a.Stats, b.Stats...)
			return a
		},
	)
}