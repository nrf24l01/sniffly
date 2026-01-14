package aggregators

import (
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
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
	loadPostgres func(*gorm.DB, []time.Time, []uuid.UUID) ([]TPostgresModel, error),
	convert func([]TPostgresModel) []TChartData,
	getTimestamp func(TChartData) int64,
	getDeviceMAC func(TChartData) string,
	mergeStats func(TChartData, TChartData) TChartData,
	filter func([]TChartData, TimeRange) []TChartData,
	deviceIDs []uuid.UUID,
) ([]TChartData, error) {
	cacheVersions, err := loadCacheVersionsFromPostgres(db, timerange)
	if err != nil {
		return nil, err
	}

	data_per_day_cache := make(map[time.Time][]TChartData)
	days_from_cache := []time.Time{}

	// Device-scoped requests must bypass day cache (cache is not device-aware).
	if len(deviceIDs) == 0 {
		data_per_day_cache, days_from_cache, err = getCacheEntriesPerInterval[TChartData](config, rdb, timerange, cacheVersions, cachePrefix)
		if err != nil {
			return nil, err
		}
	}

	non_cached_days := getNonCachedDays(timerange, days_from_cache)

	data_per_day_uncache, err := loadPostgres(db, non_cached_days, deviceIDs)
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

	merged := mergeData(data_per_day_cache, uncached_converted, getDeviceMAC, mergeStats)
	if filter != nil {
		merged = filter(merged, timerange)
	}

	return merged, nil
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
		end_of_day := date.Add(24*time.Hour).Unix() - 1

		if version, ok := cacheVersions[date]; ok {
			err := setCacheEntryForInterval(config, rdb, TimeRange{Start: start_of_day, End: end_of_day}, items, version, prefix)
			if err != nil {
				return err
			}
		}
	}
	return nil
}


func compressTrafficBuckets(stats []Traffic) []Traffic {
	if len(stats) == 0 {
		return nil
	}
	acc := make(map[int64]Traffic, len(stats))
	for _, s := range stats {
		x := acc[s.Bucket]
		x.Bucket = s.Bucket
		x.UpBytes += s.UpBytes
		x.DownBytes += s.DownBytes
		x.ReqCount += s.ReqCount
		acc[s.Bucket] = x
	}
	out := make([]Traffic, 0, len(acc))
	for _, v := range acc {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Bucket < out[j].Bucket })
	return out
}

func compressDomainBuckets(stats []DomainStat) []DomainStat {
	if len(stats) == 0 {
		return nil
	}
	acc := make(map[int64]DomainStat, len(stats))
	for _, s := range stats {
		x := acc[s.Bucket]
		x.Bucket = s.Bucket
		if x.Domains == nil {
			x.Domains = make(map[string]uint64)
		}
		for k, v := range s.Domains {
			x.Domains[k] += v
		}
		x.ReqCount += s.ReqCount
		acc[s.Bucket] = x
	}
	out := make([]DomainStat, 0, len(acc))
	for _, v := range acc {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Bucket < out[j].Bucket })
	return out
}

func compressProtoBuckets(stats []ProtoStat) []ProtoStat {
	if len(stats) == 0 {
		return nil
	}
	acc := make(map[int64]ProtoStat, len(stats))
	for _, s := range stats {
		x := acc[s.Bucket]
		x.Bucket = s.Bucket
		if x.Protos == nil {
			x.Protos = make(map[string]uint64)
		}
		for k, v := range s.Protos {
			x.Protos[k] += v
		}
		x.ReqCount += s.ReqCount
		acc[s.Bucket] = x
	}
	out := make([]ProtoStat, 0, len(acc))
	for _, v := range acc {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Bucket < out[j].Bucket })
	return out
}

func compressCountryBuckets(stats []CountryStat) []CountryStat {
	if len(stats) == 0 {
		return nil
	}
	acc := make(map[int64]CountryStat, len(stats))
	for _, s := range stats {
		x := acc[s.Bucket]
		x.Bucket = s.Bucket
		if x.Countries == nil {
			x.Countries = make(map[string]uint64)
		}
		if x.Companies == nil {
			x.Companies = make(map[string]uint64)
		}
		for k, v := range s.Countries {
			x.Countries[k] += v
		}
		for k, v := range s.Companies {
			x.Companies[k] += v
		}
		x.ReqCount += s.ReqCount
		acc[s.Bucket] = x
	}
	out := make([]CountryStat, 0, len(acc))
	for _, v := range acc {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Bucket < out[j].Bucket })
	return out
}

func GetTrafficChartData(db *gorm.DB, rdb *redisutil.RedisClient, config *core.Config, timerange TimeRange, deviceIDs []uuid.UUID) (TrafficChartResponse, error) {
	merged, err := GetGenericChartData(
		db, rdb, config, timerange, "v2_traffic_",
		loadFromPostgres,
		func(models []analyzerModels.DeviceTraffic5s) []TrafficChartData {
			var result []TrafficChartData
			for _, entry := range models {
				result = append(result, TrafficChartData{
					Device: Device{MAC: "__all__"},
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
		filterTrafficByRange,
		deviceIDs,
	)
	if err != nil {
		return TrafficChartResponse{}, err
	}

	if len(merged) == 0 {
		return TrafficChartResponse{Stats: []Traffic{}}, nil
	}

	stats := compressTrafficBuckets(merged[0].Stats)
	return TrafficChartResponse{Stats: stats}, nil
}

func GetDomainChartData(db *gorm.DB, rdb *redisutil.RedisClient, config *core.Config, timerange TimeRange, deviceIDs []uuid.UUID) (DomainChartResponse, error) {
	merged, err := GetGenericChartData(
		db, rdb, config, timerange, "v2_domain_",
		loadDomainsFromPostgres,
		func(models []analyzerModels.DeviceDomain5s) []DomainChartData {
			var result []DomainChartData
			for _, entry := range models {
				var domains map[string]uint64
				if err := json.Unmarshal([]byte(entry.Domain), &domains); err != nil {
					continue
				}
				result = append(result, DomainChartData{
					Device: Device{MAC: "__all__"},
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
		filterDomainsByRange,
		deviceIDs,
	)
	if err != nil {
		return DomainChartResponse{}, err
	}
	if len(merged) == 0 {
		return DomainChartResponse{Stats: []DomainStat{}}, nil
	}
	stats := compressDomainBuckets(merged[0].Stats)
	return DomainChartResponse{Stats: stats}, nil
}

func GetProtoChartData(db *gorm.DB, rdb *redisutil.RedisClient, config *core.Config, timerange TimeRange, deviceIDs []uuid.UUID) (ProtoChartResponse, error) {
	merged, err := GetGenericChartData(
		db, rdb, config, timerange, "v2_proto_",
		loadProtosFromPostgres,
		func(models []analyzerModels.DeviceProto5s) []ProtoChartData {
			var result []ProtoChartData
			for _, entry := range models {
				var protos map[string]uint64
				if err := json.Unmarshal([]byte(entry.Proto), &protos); err != nil {
					continue
				}
				result = append(result, ProtoChartData{
					Device: Device{MAC: "__all__"},
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
		filterProtosByRange,
		deviceIDs,
	)
	if err != nil {
		return ProtoChartResponse{}, err
	}
	if len(merged) == 0 {
		return ProtoChartResponse{Stats: []ProtoStat{}}, nil
	}
	stats := compressProtoBuckets(merged[0].Stats)
	return ProtoChartResponse{Stats: stats}, nil
}

func GetCountryChartData(db *gorm.DB, rdb *redisutil.RedisClient, config *core.Config, timerange TimeRange, deviceIDs []uuid.UUID) (CountryChartResponse, error) {
	merged, err := GetGenericChartData(
		db, rdb, config, timerange, "v2_country_",
		loadCountriesFromPostgres,
		func(models []analyzerModels.DeviceCountry5s) []CountryChartData {
			var result []CountryChartData
			for _, entry := range models {
				var countriesMap map[string]uint64
				var companiesMap map[string]uint64

				if err := json.Unmarshal([]byte(entry.Countries), &countriesMap); err != nil {
					// try fallback: maybe it's an array of strings -> convert to map with count 1
					var arr []string
					if err2 := json.Unmarshal([]byte(entry.Countries), &arr); err2 != nil {
						log.Printf("Error unmarshalling countries: %v", err)
						continue
					}
					countriesMap = make(map[string]uint64, len(arr))
					for _, k := range arr {
						countriesMap[k] = 1
					}
				}

				if err := json.Unmarshal([]byte(entry.Companies), &companiesMap); err != nil {
					var arr []string
					if err2 := json.Unmarshal([]byte(entry.Companies), &arr); err2 != nil {
						log.Printf("Error unmarshalling companies: %v", err)
						continue
					}
					companiesMap = make(map[string]uint64, len(arr))
					for _, k := range arr {
						companiesMap[k] = 1
					}
				}

				result = append(result, CountryChartData{
					Device: Device{MAC: "__all__"},
					Stats: []CountryStat{{
						Bucket:    entry.Bucket.Unix(),
						Countries: countriesMap,
						Companies: companiesMap,
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
		filterCountriesByRange,
		deviceIDs,
	)
	if err != nil {
		return CountryChartResponse{}, err
	}
	if len(merged) == 0 {
		return CountryChartResponse{Stats: []CountryStat{}}, nil
	}
	stats := compressCountryBuckets(merged[0].Stats)
	return CountryChartResponse{Stats: stats}, nil
}

func filterTrafficByRange(data []TrafficChartData, tr TimeRange) []TrafficChartData {
	return filterByRange(data, tr, func(t TrafficChartData) []int64 {
		out := make([]int64, 0, len(t.Stats))
		for _, s := range t.Stats {
			out = append(out, s.Bucket)
		}
		return out
	}, func(t TrafficChartData, keep []bool) TrafficChartData {
		stats := t.Stats[:0]
		for i, s := range t.Stats {
			if keep[i] {
				stats = append(stats, s)
			}
		}
		t.Stats = stats
		return t
	})
}

func filterDomainsByRange(data []DomainChartData, tr TimeRange) []DomainChartData {
	return filterByRange(data, tr, func(t DomainChartData) []int64 {
		out := make([]int64, 0, len(t.Stats))
		for _, s := range t.Stats {
			out = append(out, s.Bucket)
		}
		return out
	}, func(t DomainChartData, keep []bool) DomainChartData {
		stats := t.Stats[:0]
		for i, s := range t.Stats {
			if keep[i] {
				stats = append(stats, s)
			}
		}
		t.Stats = stats
		return t
	})
}

func filterProtosByRange(data []ProtoChartData, tr TimeRange) []ProtoChartData {
	return filterByRange(data, tr, func(t ProtoChartData) []int64 {
		out := make([]int64, 0, len(t.Stats))
		for _, s := range t.Stats {
			out = append(out, s.Bucket)
		}
		return out
	}, func(t ProtoChartData, keep []bool) ProtoChartData {
		stats := t.Stats[:0]
		for i, s := range t.Stats {
			if keep[i] {
				stats = append(stats, s)
			}
		}
		t.Stats = stats
		return t
	})
}

func filterCountriesByRange(data []CountryChartData, tr TimeRange) []CountryChartData {
	return filterByRange(data, tr, func(t CountryChartData) []int64 {
		out := make([]int64, 0, len(t.Stats))
		for _, s := range t.Stats {
			out = append(out, s.Bucket)
		}
		return out
	}, func(t CountryChartData, keep []bool) CountryChartData {
		stats := t.Stats[:0]
		for i, s := range t.Stats {
			if keep[i] {
				stats = append(stats, s)
			}
		}
		t.Stats = stats
		return t
	})
}

// filterByRange removes bucket entries outside the requested timerange.
// It keeps devices that still have at least one bucket after filtering.
func filterByRange[T any](data []T, tr TimeRange, buckets func(T) []int64, apply func(T, []bool) T) []T {
	out := make([]T, 0, len(data))
	for _, item := range data {
		bs := buckets(item)
		keep := make([]bool, len(bs))
		kept := 0
		for i, b := range bs {
			if b >= tr.Start && b <= tr.End {
				keep[i] = true
				kept++
			}
		}
		if kept == 0 {
			continue
		}
		out = append(out, apply(item, keep))
	}
	return out
}
