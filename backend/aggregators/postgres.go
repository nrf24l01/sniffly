package aggregators

import (
	"time"

	analyzerModels "github.com/nrf24l01/sniffly/analyzer/postgres"
	"gorm.io/gorm"
)

func loadFromPostgres(db *gorm.DB, times_to_load []time.Time) ([]analyzerModels.DeviceTraffic5s, error) {
	results := make([]analyzerModels.DeviceTraffic5s, 0)

	batchSize := 100
	for i := 0; i < len(times_to_load); i += batchSize {
		end := i + batchSize
		if end > len(times_to_load) {
			end = len(times_to_load)
		}
		
		batch := times_to_load[i:end]
		dayStart := batch[0].Truncate(24 * time.Hour)
		dayEnd := batch[len(batch)-1].Truncate(24 * time.Hour).Add(24 * time.Hour)
		
		batchResults := make([]analyzerModels.DeviceTraffic5s, 0)
		if err := db.Where("bucket >= ? AND bucket < ?", dayStart, dayEnd).Find(&batchResults).Error; err != nil {
			return nil, err
		}
		
		results = append(results, batchResults...)
	}
	
	return results, nil
}

func loadCacheVersionsFromPostgres(db *gorm.DB, time_range TimeRange) (map[time.Time]int, error) {
	cacheVersions := make(map[time.Time]int)

	times_to_load := []time.Time{}
	for date := time.Unix(time_range.Start, 0).UTC().Truncate(24 * time.Hour); !date.After(time.Unix(time_range.End, 0).UTC().Truncate(24 * time.Hour)); date = date.Add(24 * time.Hour) {
		times_to_load = append(times_to_load, date)
	}

	batchSize := 100
	for i := 0; i < len(times_to_load); i += batchSize {
		end := i + batchSize
		if end > len(times_to_load) {
			end = len(times_to_load)
		}
		
		batch := times_to_load[i:end]
		dayStart := batch[0].Truncate(24 * time.Hour)
		dayEnd := batch[len(batch)-1].Truncate(24 * time.Hour).Add(24 * time.Hour)
		
		var batchResults []analyzerModels.DayCacheVersion
		if err := db.Where("day >= ? AND day < ?", dayStart, dayEnd).Find(&batchResults).Error; err != nil {
			return nil, err
		}
		
		for _, entry := range batchResults {
			cacheVersions[entry.Day] = entry.Version
		}
	}
	
	return cacheVersions, nil
}

func loadDomainsFromPostgres(db *gorm.DB, times_to_load []time.Time) ([]analyzerModels.DeviceDomain5s, error) {
	results := make([]analyzerModels.DeviceDomain5s, 0)

	batchSize := 100
	for i := 0; i < len(times_to_load); i += batchSize {
		end := i + batchSize
		if end > len(times_to_load) {
			end = len(times_to_load)
		}
		
		batch := times_to_load[i:end]
		dayStart := batch[0].Truncate(24 * time.Hour)
		dayEnd := batch[len(batch)-1].Truncate(24 * time.Hour).Add(24 * time.Hour)
		
		batchResults := make([]analyzerModels.DeviceDomain5s, 0)
		if err := db.Where("bucket >= ? AND bucket < ?", dayStart, dayEnd).Find(&batchResults).Error; err != nil {
			return nil, err
		}
		
		results = append(results, batchResults...)
	}
	
	return results, nil
}

func loadCountriesFromPostgres(db *gorm.DB, times_to_load []time.Time) ([]analyzerModels.DeviceCountry5s, error) {
	results := make([]analyzerModels.DeviceCountry5s, 0)

	batchSize := 100
	for i := 0; i < len(times_to_load); i += batchSize {
		end := i + batchSize
		if end > len(times_to_load) {
			end = len(times_to_load)
		}
		
		batch := times_to_load[i:end]
		dayStart := batch[0].Truncate(24 * time.Hour)
		dayEnd := batch[len(batch)-1].Truncate(24 * time.Hour).Add(24 * time.Hour)
		
		batchResults := make([]analyzerModels.DeviceCountry5s, 0)
		if err := db.Where("bucket >= ? AND bucket < ?", dayStart, dayEnd).Find(&batchResults).Error; err != nil {
			return nil, err
		}
		
		results = append(results, batchResults...)
	}
	
	return results, nil
}

func loadProtosFromPostgres(db *gorm.DB, times_to_load []time.Time) ([]analyzerModels.DeviceProto5s, error) {
	results := make([]analyzerModels.DeviceProto5s, 0)

	batchSize := 100
	for i := 0; i < len(times_to_load); i += batchSize {
		end := i + batchSize
		if end > len(times_to_load) {
			end = len(times_to_load)
		}
		
		batch := times_to_load[i:end]
		dayStart := batch[0].Truncate(24 * time.Hour)
		dayEnd := batch[len(batch)-1].Truncate(24 * time.Hour).Add(24 * time.Hour)
		
		batchResults := make([]analyzerModels.DeviceProto5s, 0)
		if err := db.Where("bucket >= ? AND bucket < ?", dayStart, dayEnd).Find(&batchResults).Error; err != nil {
			return nil, err
		}
		
		results = append(results, batchResults...)
	}
	
	return results, nil
}